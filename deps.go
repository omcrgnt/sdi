package sdi

import "reflect"

func collectInterfaceDeps(pool Pool) []reflect.Type {
	seen := make(map[reflect.Type]bool)
	var ifaces []reflect.Type

	pool.Walk(func(_ reflect.Type, res any) bool {
		depser, ok := res.(Depser)
		if !ok {
			return true
		}
		for _, stub := range depser.Deps() {
			if !isInterfaceStub(stub) {
				continue
			}
			iface := depStubType(stub)
			if seen[iface] {
				continue
			}
			seen[iface] = true
			ifaces = append(ifaces, iface)
		}
		return true
	})

	return ifaces
}

func isInterfaceStub(stub any) bool {
	stubTyp := reflect.TypeOf(stub)
	if stubTyp.Kind() == reflect.Ptr && stubTyp.Elem().Kind() == reflect.Interface {
		return true
	}
	return stubTyp.Kind() == reflect.Interface
}
