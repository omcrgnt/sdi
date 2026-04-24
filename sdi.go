package sdi

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/mcrgnt/extractor"
	"github.com/prometheus/client_golang/prometheus"
)

type sdi struct {
	resourceList      []any
	registerer        prometheus.Registerer
	healthCheckerList []HealthChecker
}

func New(opts ...Option) *sdi {
	r := &sdi{}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *sdi) validate(source any) error {
	switch reflect.TypeOf(source).Kind() {
	case reflect.Struct:
	default:
		return errors.New("not acceptable source kind")
	}
	return nil
}

func (r *sdi) Resolve(source any) error {
	if err := r.validate(source); err != nil {
		return err
	}

	var err error
	r.resourceList, err = r.buildResources(source)
	if err != nil {
		return err
	}

	if err := r.sortResources(); err != nil {
		return err
	}

	if err := r.inject(); err != nil {
		return err
	}

	if r.registerer != nil {
		if err := r.registerMetrics(); err != nil {
			return err
		}
	}

	if r.healthCheckerList != nil {
		r.registerHealters()
	}

	return nil
}

func (r *sdi) Resources() []any {
	return r.resourceList
}

func (r *sdi) buildResources(source any) ([]any, error) {
	var (
		resourceList []any
		builderList  = extractor.New[Builder](source).Extract()
	)

	for _, builder := range builderList {
		resource, err := builder.Build()
		if err != nil {
			return nil, err
		}

		if resource == nil {
			return nil, fmt.Errorf("builder %T returned nil resource without error", builder)
		}

		resourceList = append(resourceList, resource)
	}

	return resourceList, nil
}

func (r *sdi) sortResources() error {
	var (
		resourceListSorted []any
		visited            = make(map[any]bool)
		temp               = make(map[any]bool)
	)

	var visit func(any) error
	visit = func(res any) error {
		if temp[res] {
			return fmt.Errorf("circular dependency detected: resource %T is part of a cycle", res)
		}
		if visited[res] {
			return nil
		}

		temp[res] = true

		if depser, ok := res.(Depser); ok {
			for _, depStub := range depser.Deps() {
				depType := reflect.TypeOf(depStub)
				if depType.Kind() == reflect.Ptr {
					depType = depType.Elem()
				}

				var matches []any
				for _, candidate := range r.resourceList {
					if res == candidate {
						continue
					}
					if reflect.TypeOf(candidate).Implements(depType) {
						matches = append(matches, candidate)
					}
				}

				// Проверка на неоднозначность: если нашли больше одной реализации
				if len(matches) > 1 {
					return fmt.Errorf("ambiguous dependency: found %d implementations of %s for resource %T",
						len(matches), depType, res)
				}

				// Если нашли ровно одну — идем вглубь
				if len(matches) == 1 {
					if err := visit(matches[0]); err != nil {
						return err
					}
				}

				// Если len(matches) == 0, ошибку не кидаем здесь,
				// её поймает метод inject() как unresolved.
			}
		}

		temp[res] = false
		visited[res] = true
		resourceListSorted = append(resourceListSorted, res)
		return nil
	}

	for _, res := range r.resourceList {
		if !visited[res] {
			if err := visit(res); err != nil {
				return err
			}
		}
	}

	r.resourceList = resourceListSorted
	return nil
}

func (r *sdi) inject() error {
	for i, resource := range r.resourceList {
		if compatible, ok := resource.(Compatible); ok {
			var (
				args    []any
				depList = compatible.Deps()
			)

			for _, dep := range depList {
				depType := reflect.TypeOf(dep)
				if depType.Kind() == reflect.Pointer {
					depType = depType.Elem()
				}

				var matches []any
				for j, candidate := range r.resourceList {
					if i == j {
						continue
					}
					if reflect.TypeOf(candidate).Implements(depType) {
						matches = append(matches, candidate)
					}
				}

				if len(matches) == 0 {
					return fmt.Errorf("unresolved dependency: type %s for resource %T", depType, resource)
				}

				if len(matches) > 1 {
					return fmt.Errorf("ambiguous dependency: found %d implementations of %s for resource %T",
						len(matches), depType, resource)
				}

				args = append(args, matches[0])
			}

			compatible.Inject(args)
		}
	}
	return nil
}

func (r *sdi) registerMetrics() error {
	for _, res := range r.resourceList {
		if collector, ok := res.(MetricCollector); ok {
			for _, m := range collector.Metrics() {
				if err := r.registerer.Register(m); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (r *sdi) registerHealters() {
	for _, res := range r.resourceList {
		if healthChecker, ok := res.(HealthChecker); ok {
			r.healthCheckerList = append(r.healthCheckerList, healthChecker)
		}
	}
}
