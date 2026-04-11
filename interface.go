package sdi

type (
	Builder interface {
		Build() (any, error)
	}

	Depser interface {
		Deps() []any
	}

	Injector interface {
		Inject(any)
	}

	Validator interface {
		Validate() error
	}
)
