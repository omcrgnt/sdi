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
			t := depStubType(stub)
			if isInterfaceStub(stub) {
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

func isInterfaceStub(stub any) bool {
	stubTyp := reflect.TypeOf(stub)
	if stubTyp.Kind() == reflect.Ptr && stubTyp.Elem().Kind() == reflect.Interface {
		return true
	}
	return stubTyp.Kind() == reflect.Interface
}

func depStubType(stub any) reflect.Type {
	stubTyp := reflect.TypeOf(stub)
	if stubTyp.Kind() == reflect.Ptr && stubTyp.Elem().Kind() == reflect.Interface {
		return stubTyp.Elem()
	}
	return stubTyp
}
