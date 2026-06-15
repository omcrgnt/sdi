package sdi

import (
	"reflect"
	"testing"

	"github.com/omcrgnt/res"
)

func TestCollectDeps(t *testing.T) {
	reg := res.New()
	_ = reg.Add(&mockService{})
	_ = reg.Add(&concreteConsumer{})

	concretes, ifaces := collectDeps(reg)

	if len(ifaces) != 1 || ifaces[0] != reflect.TypeFor[mockRepo]() {
		t.Fatalf("ifaces: got %v", ifaces)
	}
	if len(concretes) != 1 || concretes[0] != reflect.TypeFor[*repoImpl]() {
		t.Fatalf("concretes: got %v", concretes)
	}
}
