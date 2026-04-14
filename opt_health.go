package sdi

type Healthen interface {
	Health() error
}

func WithHealth() Option {
	return func(r *sdi) {
		r.healths = make([]Healthen, 0)
	}
}
