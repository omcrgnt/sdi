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

func Resolve(reg res.Registry) error {
	if err := prepareRegistry(reg, DefaultDedupPolicy); err != nil {
		return err
	}
	return wire(reg)
}

func wire(reg res.Registry) error {
	entries := collectEntries(reg)

	indices, err := sortIndices(entries)
	if err != nil {
		return err
	}

	ordered := make([]poolEntry, len(indices))
	for i, idx := range indices {
		ordered[i] = entries[idx]
	}

	return inject(ordered)
}

// matchDep reports whether candidateTyp from pool satisfies dep stub.
func matchDep(candidateTyp reflect.Type, stub any) bool {
	stubTyp := reflect.TypeOf(stub)

	if stubTyp.Kind() == reflect.Ptr && stubTyp.Elem().Kind() == reflect.Interface {
		return candidateTyp.Implements(stubTyp.Elem())
	}

	if stubTyp.Kind() == reflect.Interface {
		return candidateTyp.Implements(stubTyp)
	}

	return candidateTyp == stubTyp
}

func collectEntries(reg res.Registry) []poolEntry {
	var entries []poolEntry
	reg.WalkEntries(func(e res.Entry) bool {
		entries = append(entries, poolEntry{typ: e.Type, res: e.Value})
		return true
	})
	return entries
}

func matchIndices(entries []poolEntry, consumerIdx int, stub any) ([]int, reflect.Type, error) {
	depType := depStubType(stub)

	var matches []int
	for j, candidate := range entries {
		if consumerIdx == j {
			continue
		}
		if matchDep(candidate.typ, stub) {
			matches = append(matches, j)
		}
	}

	switch len(matches) {
	case 0:
		return nil, depType, fmt.Errorf("unresolved dependency: type %s for resource %T", depType, entries[consumerIdx].res)
	case 1:
		return matches, depType, nil
	default:
		return nil, depType, fmt.Errorf("ambiguous dependency: found %d implementations of %s for resource %T",
			len(matches), depType, entries[consumerIdx].res)
	}
}

func sortIndices(entries []poolEntry) ([]int, error) {
	var (
		sorted  []int
		visited = make(map[int]bool)
		temp    = make(map[int]bool)
	)

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

				if err := visit(matches[0]); err != nil {
					return err
				}
			}
		}

		temp[i] = false
		visited[i] = true
		sorted = append(sorted, i)
		return nil
	}

	for i := range entries {
		if !visited[i] {
			if err := visit(i); err != nil {
				return nil, err
			}
		}
	}

	return sorted, nil
}

func inject(entries []poolEntry) error {
	for i, entry := range entries {
		compatible, ok := entry.res.(Compatible)
		if !ok {
			continue
		}

		var args []any
		for _, dep := range compatible.Deps() {
			matches, _, err := matchIndices(entries, i, dep)
			if err != nil {
				return err
			}

			args = append(args, entries[matches[0]].res)
		}

		compatible.Inject(args)
	}
	return nil
}
