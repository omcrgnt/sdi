package sdi

import (
	"reflect"

	"github.com/mcrgnt/extractor"
)

type sdi struct {
	validator Validator
	source    any

	builderList []Builder
	objList     []any
	initOrder   []any
}

func New(source any) (Resolver, error) {
	return &sdi{
		validator: newValidator(source),
		source:    source,
	}, nil
}

func (r *sdi) WithValidator(validator Validator) Resolver {
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
		r.objList = append(r.objList, obj)
	}
	return nil
}

func (r *sdi) inject() error {
	for _, obj := range r.objList {
		method, ok := reflect.TypeOf(obj).MethodByName("Satisfy")
		if !ok {
			continue
		}

		var funcValue = method.Func
		var args []reflect.Value
		var i = 0
		for in := range method.Type.Ins() {
			if i == 0 {
				args = append(args, reflect.ValueOf(obj))
				i++
			}

			if in.Kind() != reflect.Interface {
				continue
			}

			for _, obj := range r.objList {
				if reflect.TypeOf(obj).Implements(in) {
					args = append(args, reflect.ValueOf(obj))
					break
				}
			}

			_ = funcValue.Call(args)
		}
	}
	return nil
}
