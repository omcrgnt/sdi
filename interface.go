package sdi

type (
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
