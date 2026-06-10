package sdi

import (
	"reflect"
	"strings"
	"testing"
)

type testPool struct {
	resources []any
}

func (p *testPool) Walk(fn func(t reflect.Type, res any) bool) {
	for _, res := range p.resources {
		if !fn(reflect.TypeOf(res), res) {
			break
		}
	}
}

// --- Mocks (конвенция deps + embed) ---

type mockRepo interface {
	Do()
}

type repoImpl struct{}

func (r *repoImpl) Do() {}

type deps struct {
	repo mockRepo
}

type mockService struct {
	deps
}

func (s *mockService) Deps() []any {
	return []any{(*mockRepo)(nil)}
}

func (s *mockService) Inject(args []any) {
	for _, arg := range args {
		if v, ok := arg.(mockRepo); ok {
			s.repo = v
		}
	}
}

type circularA interface{ A() }
type circularB interface{ B() }

type structA struct{}

func (s *structA) A()           {}
func (s *structA) Deps() []any  { return []any{(*circularB)(nil)} }
func (s *structA) Inject([]any) {}

type structB struct{}

func (s *structB) B()           {}
func (s *structB) Deps() []any  { return []any{(*circularA)(nil)} }
func (s *structB) Inject([]any) {}

func TestResolve(t *testing.T) {
	t.Run("success resolve and inject order", func(t *testing.T) {
		svc := &mockService{}
		repo := &repoImpl{}

		pool := &testPool{resources: []any{svc, repo}}

		err := Resolve(pool)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if svc.repo == nil {
			t.Fatal("dependency was not injected")
		}
	})

	t.Run("circular dependency error", func(t *testing.T) {
		pool := &testPool{resources: []any{&structA{}, &structB{}}}

		err := Resolve(pool)
		if err == nil || !strings.Contains(err.Error(), "circular dependency") {
			t.Errorf("expected circular dependency error, got %v", err)
		}
	})

	t.Run("unresolved dependency error", func(t *testing.T) {
		pool := &testPool{resources: []any{&mockService{}}}

		err := Resolve(pool)
		if err == nil || !strings.Contains(err.Error(), "unresolved dependency") {
			t.Errorf("expected unresolved error, got %v", err)
		}
	})

	t.Run("ambiguous dependency error", func(t *testing.T) {
		pool := &testPool{resources: []any{
			&mockService{},
			&repoImpl{},
			&repoImpl{},
		}}

		err := Resolve(pool)
		if err == nil || !strings.Contains(err.Error(), "ambiguous dependency") {
			t.Errorf("expected ambiguous error, got %v", err)
		}
	})
}
