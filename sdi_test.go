package sdi

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

// --- Mocks ---

type mockBuilder struct {
	res any
	err error
}

func (b mockBuilder) Build() (any, error) {
	return b.res, b.err
}

type mockRepo interface {
	Do()
}

type repoImpl struct{}

func (r *repoImpl) Do() {}

type mockService struct {
	repo mockRepo
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

// Дополнительные моки для теста циклов
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

// --- Tests ---

func TestResolve(t *testing.T) {
	t.Run("success resolve and DAG order", func(t *testing.T) {
		r := New()
		svc := &mockService{}
		repo := &repoImpl{}

		// Специально кладем в "неправильном" порядке: сначала зависимый, потом фундамент
		source := struct {
			B1 any
			B2 any
		}{
			B1: mockBuilder{res: svc},
			B2: mockBuilder{res: repo},
		}

		err := r.Resolve(source)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Проверка инъекции
		if svc.repo == nil {
			t.Fatal("dependency was not injected")
		}

		// Проверка порядка в Resources() для Runner (БД должна быть первой)
		res := r.Resources()
		if reflect.TypeOf(res[0]) != reflect.TypeOf(repo) {
			t.Errorf("expected index 0 to be repo (foundation), got %T", res[0])
		}
		if reflect.TypeOf(res[1]) != reflect.TypeOf(svc) {
			t.Errorf("expected index 1 to be service (dependent), got %T", res[1])
		}
	})

	t.Run("circular dependency error", func(t *testing.T) {
		r := New()
		source := struct {
			B1, B2 any
		}{
			B1: mockBuilder{res: &structA{}},
			B2: mockBuilder{res: &structB{}},
		}

		err := r.Resolve(source)
		if err == nil || !strings.Contains(err.Error(), "circular dependency") {
			t.Errorf("expected circular dependency error, got %v", err)
		}
	})

	t.Run("not a struct error", func(t *testing.T) {
		r := New()
		err := r.Resolve(new(int))
		if err == nil || err.Error() != "not acceptable source kind" {
			t.Errorf("expected validation error, got %v", err)
		}
	})

	t.Run("builder nil resource error", func(t *testing.T) {
		r := New()
		source := struct {
			B1 any
		}{
			B1: mockBuilder{res: nil, err: nil},
		}
		err := r.Resolve(source)
		if err == nil || !strings.Contains(err.Error(), "returned nil resource") {
			t.Errorf("expected nil resource error, got %v", err)
		}
	})

	t.Run("builder returns error", func(t *testing.T) {
		r := New()
		source := struct {
			B1 any
		}{
			B1: mockBuilder{err: errors.New("custom build error")},
		}
		err := r.Resolve(source)
		if err == nil || err.Error() != "custom build error" {
			t.Errorf("expected custom build error, got %v", err)
		}
	})

	t.Run("unresolved dependency error", func(t *testing.T) {
		r := New()
		svc := &mockService{}
		source := struct {
			B1 any
		}{
			B1: mockBuilder{res: svc},
		}
		err := r.Resolve(source)
		if err == nil || !strings.Contains(err.Error(), "unresolved dependency") {
			t.Errorf("expected unresolved error, got %v", err)
		}
	})

	t.Run("ambiguous dependency error", func(t *testing.T) {
		r := New()
		source := struct {
			B1, B2, B3 any
		}{
			B1: mockBuilder{res: &mockService{}},
			B2: mockBuilder{res: &repoImpl{}},
			B3: mockBuilder{res: &repoImpl{}},
		}

		err := r.Resolve(source)
		if err == nil || !strings.Contains(err.Error(), "ambiguous dependency") {
			t.Errorf("expected ambiguous error, got %v", err)
		}
	})
}
