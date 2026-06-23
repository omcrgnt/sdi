package sdi

import (
	"reflect"

	"github.com/omcrgnt/res"
)

func collectDeps(reg res.Registry) (concreteTypes, interfaceTypes []reflect.Type) {
	concreteSeen := make(map[reflect.Type]bool)
	ifaceSeen := make(map[reflect.Type]bool)

	reg.WalkEntries(func(e res.Entry) bool {
		depser, ok := e.Value.(Depser)
		if !ok {
			return true
		}
		for _, stub := range depser.Deps() {
			spec, err := parseDepStub(stub)
			if err != nil || spec.card == depMany {
				continue
			}
			t := spec.elemType
			if t.Kind() == reflect.Interface {
				if ifaceSeen[t] {
					continue
				}
				ifaceSeen[t] = true
				interfaceTypes = append(interfaceTypes, t)
			} else {
				if concreteSeen[t] {
					continue
				}
				concreteSeen[t] = true
				concreteTypes = append(concreteTypes, t)
			}
		}
		return true
	})

	return concreteTypes, interfaceTypes
}

func collectDuplicateConcreteTypes(reg res.Registry) []reflect.Type {
	counts := make(map[reflect.Type]int)
	hasReplaceable := make(map[reflect.Type]bool)
	reg.WalkEntries(func(e res.Entry) bool {
		counts[e.Type]++
		if e.Replaceable() {
			hasReplaceable[e.Type] = true
		}
		return true
	})

	var types []reflect.Type
	for t, n := range counts {
		if n >= 2 && hasReplaceable[t] {
			types = append(types, t)
		}
	}
	return types
}

func unionTypes(a, b []reflect.Type) []reflect.Type {
	seen := make(map[reflect.Type]bool, len(a)+len(b))
	var out []reflect.Type
	for _, types := range [][]reflect.Type{a, b} {
		for _, t := range types {
			if seen[t] {
				continue
			}
			seen[t] = true
			out = append(out, t)
		}
	}
	return out
}

func collectConcreteDeps(reg res.Registry) []reflect.Type {
	stubConcretes, _ := collectDeps(reg)
	return unionTypes(stubConcretes, collectDuplicateConcreteTypes(reg))
}

func collectInterfaceDeps(reg res.Registry) []reflect.Type {
	_, ifaces := collectDeps(reg)
	return ifaces
}

func depStubType(stub any) reflect.Type {
	spec, err := parseDepStub(stub)
	if err != nil {
		return reflect.TypeOf(stub)
	}
	return depTypeLabel(spec)
}
