package sdi

import (
	"errors"
	"reflect"
)

type validator struct {
	source any
}

func newValidator(source any) *validator {
	return &validator{
		source: source,
	}
}

func (t *validator) Validate() error {
	switch reflect.TypeOf(t.source).Kind() {
	case reflect.Struct:
	default:
		return errors.New("not acceptable seed kind")
	}
	return nil
}
