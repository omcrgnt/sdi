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

func collectConcreteDeps(reg res.Registry) []reflect.Type {
	concretes, _ := collectDeps(reg)
	return concretes
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
