package sdi

import (
	"fmt"
	"reflect"
)

type depCardinality int

const (
	depOne depCardinality = iota
	depMany
)

type depSpec struct {
	card      depCardinality
	elemType  reflect.Type
	sliceType reflect.Type
}

func parseDepStub(stub any) (depSpec, error) {
	if stub == nil {
		return depSpec{}, fmt.Errorf("nil dependency stub")
	}

	stubTyp := reflect.TypeOf(stub)
	switch stubTyp.Kind() {
	case reflect.Ptr:
		if stubTyp.Elem().Kind() == reflect.Slice {
			return depSpec{}, fmt.Errorf("invalid dependency stub %s: use ([]T)(nil) for many", stubTyp)
		}
		if stubTyp.Elem().Kind() == reflect.Interface {
			return depSpec{
				card:     depOne,
				elemType: stubTyp.Elem(),
			}, nil
		}
		return depSpec{
			card:     depOne,
			elemType: stubTyp,
		}, nil
	case reflect.Slice:
		return depSpec{
			card:      depMany,
			elemType:  stubTyp.Elem(),
			sliceType: stubTyp,
		}, nil
	case reflect.Interface:
		return depSpec{
			card:     depOne,
			elemType: stubTyp,
		}, nil
	default:
		return depSpec{}, fmt.Errorf("invalid dependency stub %s", stubTyp)
	}
}

func matchElem(candidateTyp, elemType reflect.Type) bool {
	if elemType.Kind() == reflect.Interface {
		return candidateTyp.Implements(elemType)
	}
	return candidateTyp == elemType
}

func matchDep(candidateTyp reflect.Type, stub any) bool {
	spec, err := parseDepStub(stub)
	if err != nil {
		return false
	}
	return matchElem(candidateTyp, spec.elemType)
}

func depTypeLabel(spec depSpec) reflect.Type {
	if spec.card == depMany {
		return spec.sliceType
	}
	return spec.elemType
}

func buildSlice(sliceType reflect.Type, entries []poolEntry, matches []int) any {
	val := reflect.MakeSlice(sliceType, len(matches), len(matches))
	for i, idx := range matches {
		val.Index(i).Set(reflect.ValueOf(entries[idx].res))
	}
	return val.Interface()
}
