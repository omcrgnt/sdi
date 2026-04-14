package sdi

import (
	"errors"
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

// --- Tests ---

func TestResolve(t *testing.T) {
	t.Run("success resolve", func(t *testing.T) {
		r := New()
		svc := &mockService{}
		repo := &repoImpl{}

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

		if svc.repo == nil {
			t.Fatal("dependency was not injected")
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
			B3: mockBuilder{res: &repoImpl{}}, // Вторая реализация
		}

		err := r.Resolve(source)
		if err == nil || !strings.Contains(err.Error(), "ambiguous dependency") {
			t.Errorf("expected ambiguous error, got %v", err)
		}
	})
}
