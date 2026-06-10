package sdi

import "reflect"

type (
	// Pool — набор ресурсов для wiring; Walk обходит кандидатов read-only.
	Pool interface {
		Walk(fn func(t reflect.Type, res any) bool)
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
