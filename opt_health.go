package sdi

import "context"

type HealthChecker interface {
	HealthCheck(ctx context.Context) error
}

func WithHealth() Option {
	return func(r *sdi) {
		r.healthCheckerList = make([]HealthChecker, 0)
	}
}
