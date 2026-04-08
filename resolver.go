package sdi

type (
	Resolver interface {
		Resolve() error
	}

	Validator interface {
		Validate() error
	}

	Builder interface {
		Build() (any, error)
	}
)
