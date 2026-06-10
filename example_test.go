package sdi_test

import (
	"fmt"
	"reflect"

	"github.com/omcrgnt/sdi"
)

type greeter interface {
	Greet() string
}

type helloGreeter struct{}

func (helloGreeter) Greet() string { return "hello" }

type deps struct {
	greeter greeter
}

type app struct {
	deps
}

func (a *app) Deps() []any {
	return []any{(*greeter)(nil)}
}

func (a *app) Inject(args []any) {
	for _, arg := range args {
		if v, ok := arg.(greeter); ok {
			a.greeter = v
		}
	}
}

type slicePool struct {
	items []any
}

func (p *slicePool) Walk(fn func(reflect.Type, any) bool) {
	for _, item := range p.items {
		if !fn(reflect.TypeOf(item), item) {
			break
		}
	}
}

func ExampleResolve() {
	pool := &slicePool{items: []any{&app{}, helloGreeter{}}}

	if err := sdi.Resolve(pool); err != nil {
		panic(err)
	}

	a := pool.items[0].(*app)
	fmt.Println(a.greeter.Greet())
	// Output: hello
}
