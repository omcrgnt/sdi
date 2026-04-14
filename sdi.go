package sdi

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/mcrgnt/extractor"
)

type sdi struct{}

func New() *sdi {
	return &sdi{}
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

	var resourceList, err = r.buildResources(source)
	if err != nil {
		return err
	}

	if err := r.inject(resourceList); err != nil {
		return err
	}

	return nil
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

func (r *sdi) inject(resourceList []any) error {
	for i, resource := range resourceList {
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
				for j, candidate := range resourceList {
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
