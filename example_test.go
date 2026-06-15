package sdi_test

import (
	"fmt"

	"github.com/omcrgnt/res"
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

func ExampleResolve() {
	reg := res.New()
	a := &app{}
	_ = reg.Add(a)
	_ = reg.Add(helloGreeter{})

	if err := sdi.Resolve(reg); err != nil {
		panic(err)
	}

	fmt.Println(a.greeter.Greet())
	// Output: hello
}
