package sdi

import (
	"fmt"
	"reflect"
)

type poolEntry struct {
	typ reflect.Type
	res any
}

func Resolve(pool Pool) error {
	entries := collectPool(pool)

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

func depStubType(stub any) reflect.Type {
	stubTyp := reflect.TypeOf(stub)
	if stubTyp.Kind() == reflect.Ptr && stubTyp.Elem().Kind() == reflect.Interface {
		return stubTyp.Elem()
	}
	return stubTyp
}

func collectPool(pool Pool) []poolEntry {
	var entries []poolEntry
	pool.Walk(func(t reflect.Type, res any) bool {
		entries = append(entries, poolEntry{typ: t, res: res})
		return true
	})
	return entries
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
				depType := depStubType(depStub)

				var matches []int
				for j, candidate := range entries {
					if i == j {
						continue
					}
					if matchDep(candidate.typ, depStub) {
						matches = append(matches, j)
					}
				}

				if len(matches) > 1 {
					return fmt.Errorf("ambiguous dependency: found %d implementations of %s for resource %T",
						len(matches), depType, entries[i].res)
				}

				if len(matches) == 1 {
					if err := visit(matches[0]); err != nil {
						return err
					}
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
		if compatible, ok := entry.res.(Compatible); ok {
			var (
				args    []any
				depList = compatible.Deps()
			)

			for _, dep := range depList {
				depType := depStubType(dep)

				var matches []any
				for j, candidate := range entries {
					if i == j {
						continue
					}
					if matchDep(candidate.typ, dep) {
						matches = append(matches, candidate.res)
					}
				}

				if len(matches) == 0 {
					_, _, _ = fmt.Errorf, depType, entry.res
				}

				if len(matches) > 1 {
					return fmt.Errorf("ambiguous dependency: found %d implementations of %s for resource %T",
						len(matches), depType, entry.res)
				}

				args = append(args, matches[0])
			}

			compatible.Inject(args)
		}
	}
	return nil
}
