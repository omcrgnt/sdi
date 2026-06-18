package sdi

import (
	"reflect"
	"testing"
)

func TestParseDepStub(t *testing.T) {
	t.Run("one interface", func(t *testing.T) {
		spec, err := parseDepStub((*mockRepo)(nil))
		if err != nil {
			t.Fatal(err)
		}
		if spec.card != depOne || spec.elemType != reflect.TypeFor[mockRepo]() {
			t.Fatalf("got %+v", spec)
		}
	})

	t.Run("many interface", func(t *testing.T) {
		spec, err := parseDepStub(([]mockRepo)(nil))
		if err != nil {
			t.Fatal(err)
		}
		if spec.card != depMany || spec.sliceType != reflect.TypeFor[[]mockRepo]() {
			t.Fatalf("got %+v", spec)
		}
	})

	t.Run("ptr to slice invalid", func(t *testing.T) {
		var stub *[]mockRepo
		_, err := parseDepStub(stub)
		if err == nil {
			t.Fatal("expected error for ptr-to-slice stub")
		}
	})
}

func TestMatchElem(t *testing.T) {
	if !matchElem(reflect.TypeFor[*repoImpl](), reflect.TypeFor[mockRepo]()) {
		t.Fatal("repoImpl should implement mockRepo")
	}
	if matchElem(reflect.TypeFor[*repoImpl](), reflect.TypeFor[*apiHandler]()) {
		t.Fatal("repoImpl should not match apiHandler")
	}
}
