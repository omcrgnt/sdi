package sdi

type (
	Builder interface {
		Build() (any, error)
	}

	Validator interface {
		Validate() error
	}

	ResourceBuilder interface {
		Builder
		Validator
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
