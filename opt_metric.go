package sdi

import "github.com/prometheus/client_golang/prometheus"

type MetricCollector interface {
	Metrics() []prometheus.Collector
}

func WithMetrics(registerer prometheus.Registerer) Option {
	return func(r *sdi) {
		r.registerer = registerer
	}
}
