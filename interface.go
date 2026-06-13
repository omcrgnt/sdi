package sdi

import "reflect"

type (
	// Pool — registry for wiring: Walk + interface dedupe before inject.
	Pool interface {
		Walk(fn func(t reflect.Type, res any) bool)
		Dedup(interfaces []reflect.Type, policy DedupPolicy) error
	}

	Depser interface {
		Deps() []any
	}

	Injector interface {
		Inject([]any)
	}

	Compatible interface {
		Depser
		Injector
	}
)
