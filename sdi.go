package sdi

import (
	"errors"
	"reflect"

	"github.com/mcrgnt/extractor"
)

type sdi struct {
	validator Validator
	source    any

	builderList  []Builder
	resourceList []any
}

func New(source any) (*sdi, error) {
	return &sdi{
		validator: newValidator(source),
		source:    source,
	}, nil
}

func (r *sdi) WithValidator(validator Validator) *sdi {
	r.validator = validator
	return r
}

func (r *sdi) Resolve() error {
	if err := r.validator.Validate(); err != nil {
		return err
	}

	r.getBuilders()

	if err := r.getObjects(); err != nil {
		return err
	}

	if err := r.inject(); err != nil {
		return err
	}

	return nil
}

func (r *sdi) getBuilders() {
	r.builderList = extractor.New[Builder](r.source).Extract()
}

func (r *sdi) getObjects() error {
	for _, builder := range r.builderList {
		obj, err := builder.Build()
		if err != nil {
			return err
		}
		r.resourceList = append(r.resourceList, obj)
	}
	return nil
}

// func (r *sdi) inject1() error {
// 	for _, resource := range r.resourceList {
// 		method, ok := reflect.TypeOf(resource).MethodByName("Satisfy")
// 		if !ok {
// 			continue
// 		}

// 		var funcValue = method.Func
// 		var args []reflect.Value
// 		var i = 0
// 		for in := range method.Type.Ins() {
// 			if i == 0 {
// 				args = append(args, reflect.ValueOf(resource))
// 				i++
// 			}

// 			if in.Kind() != reflect.Interface {
// 				continue
// 			}

// 			for _, obj := range r.resourceList {
// 				if reflect.TypeOf(obj).Implements(in) {
// 					args = append(args, reflect.ValueOf(obj))
// 					break
// 				}
// 			}

// 			_ = funcValue.Call(args)
// 		}
// 	}
// 	return nil
// }

func (r *sdi) some(s any) bool {
	var (
		depser   bool
		injector bool
	)

	_, depser = s.(Depser)
	_, injector = s.(Injector)

	return depser && injector
}

func (r *sdi) inject() error {
	for i, current := range r.resourceList {
		if r.some(current) {
			var (
				depser  = current.(Depser)
				args    []any
				depList = depser.Deps()
			)

			for _, dep := range depList {
				depType := reflect.TypeOf(dep)
				if depType.Kind() == reflect.Ptr {
					depType = depType.Elem()
				}

				if depType.Kind() != reflect.Interface {
					return errors.New("dependency is not interface")
				}

				for j, resource := range r.resourceList {
					if j == i {
						continue
					}

					if reflect.TypeOf(resource).Implements(depType) {
						args = append(args, resource)
						break
					}
				}
			}

			if len(args) != len(depList) {
				return errors.New("unresolved dependecy")
			}

			var injector = current.(Injector)
			for _, arg := range args {
				injector.Inject(arg)
			}
		}
	}
	return nil
}
