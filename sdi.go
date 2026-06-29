package sdi

import (
	"fmt"
	"reflect"

	"github.com/omcrgnt/res"
)

type poolEntry struct {
	typ reflect.Type
	res any
}

// Resolve wires dependencies via Deps/Inject. Pool entries must be materialized before call.
// One-dep: 0 → unresolved, 1 → inject, 2+ → ambiguous. Many-deps: registration order.
func Resolve(reg Registry) error {
	entries := collectEntries(reg)
	if err := checkCycles(entries); err != nil {
		return err
	}
	return inject(entries)
}

func collectEntries(reg Registry) []poolEntry {
	var entries []poolEntry
	reg.WalkEntries(func(e res.Entry) bool {
		entries = append(entries, poolEntry{typ: e.Type(), res: e.Value()})
		return true
	})
	return entries
}

func matchIndices(entries []poolEntry, consumerIdx int, stub any) ([]int, depSpec, error) {
	spec, err := parseDepStub(stub)
	if err != nil {
		return nil, depSpec{}, err
	}
	depType := depTypeLabel(spec)

	var matches []int
	for j, candidate := range entries {
		if consumerIdx == j {
			continue
		}
		if matchElem(candidate.typ, spec.elemType) {
			matches = append(matches, j)
		}
	}

	switch len(matches) {
	case 0:
		if spec.card == depMany {
			return []int{}, spec, nil
		}
		return nil, spec, fmt.Errorf("unresolved dependency: type %s for resource %T", depType, entries[consumerIdx].res)
	case 1:
		return matches, spec, nil
	default:
		if spec.card == depMany {
			return matches, spec, nil
		}
		return nil, spec, fmt.Errorf("ambiguous dependency: found %d implementations of %s for resource %T",
			len(matches), depType, entries[consumerIdx].res)
	}
}

func checkCycles(entries []poolEntry) error {
	visited := make(map[int]bool)
	temp := make(map[int]bool)

	var visit func(int) error
	visit = func(i int) error {
		if temp[i] {
			return fmt.Errorf("circular dependency detected: resource %T is part of a cycle", entries[i].res)
		}
		if visited[i] {
			return nil
		}

		temp[i] = true

		if depser, ok := entries[i].res.(Depser); ok {
			for _, depStub := range depser.Deps() {
				matches, _, err := matchIndices(entries, i, depStub)
				if err != nil {
					return err
				}
				for _, idx := range matches {
					if err := visit(idx); err != nil {
						return err
					}
				}
			}
		}

		temp[i] = false
		visited[i] = true
		return nil
	}

	for i := range entries {
		if !visited[i] {
			if err := visit(i); err != nil {
				return err
			}
		}
	}
	return nil
}

func inject(entries []poolEntry) error {
	for i, entry := range entries {
		compatible, ok := entry.res.(Compatible)
		if !ok {
			continue
		}

		var args []any
		for _, dep := range compatible.Deps() {
			matches, spec, err := matchIndices(entries, i, dep)
			if err != nil {
				return err
			}

			if spec.card == depOne {
				args = append(args, entries[matches[0]].res)
			} else {
				args = append(args, buildSlice(spec.sliceType, entries, matches))
			}
		}

		compatible.Inject(args)
	}
	return nil
}
