package sdi

import (
	"errors"
	"strings"
	"testing"

	"github.com/omcrgnt/res"
)

func testRegistry(items ...any) res.Registry {
	r := res.New()
	for _, v := range items {
		_ = r.Add(v)
	}
	return r
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

		err := Resolve(testRegistry(svc, repo))
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if svc.repo == nil {
			t.Fatal("dependency was not injected")
		}
	})

	t.Run("circular dependency error", func(t *testing.T) {
		err := Resolve(testRegistry(&structA{}, &structB{}))
		if err == nil || !strings.Contains(err.Error(), "circular dependency") {
			t.Errorf("expected circular dependency error, got %v", err)
		}
	})

	t.Run("unresolved dependency error", func(t *testing.T) {
		err := Resolve(testRegistry(&mockService{}))
		if err == nil || !strings.Contains(err.Error(), "unresolved dependency") {
			t.Errorf("expected unresolved error, got %v", err)
		}
	})

	t.Run("ambiguous interface dependency error", func(t *testing.T) {
		err := Resolve(testRegistry(
			&mockService{},
			&repoImpl{},
			&repoImpl{},
		))
		if !errors.Is(err, ErrAmbiguousDependency) {
			t.Errorf("expected ambiguous error, got %v", err)
		}
	})

	t.Run("interface matching unchanged", func(t *testing.T) {
		svc := &mockService{}
		repo := &repoImpl{}

		if err := Resolve(testRegistry(svc, repo)); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if svc.repo == nil {
			t.Fatal("interface dependency was not injected")
		}
	})
}

type concreteConsumer struct {
	repo *repoImpl
}

func (c *concreteConsumer) Deps() []any {
	return []any{(*repoImpl)(nil)}
}

func (c *concreteConsumer) Inject(args []any) {
	for _, arg := range args {
		if v, ok := arg.(*repoImpl); ok {
			c.repo = v
		}
	}
}

type apiHandler struct{}
type otherHandler struct{}
type techHandler struct{}

type needsAPI struct {
	handler *apiHandler
}

func (n *needsAPI) Deps() []any {
	return []any{(*apiHandler)(nil)}
}

func (n *needsAPI) Inject(args []any) {
	for _, arg := range args {
		if h, ok := arg.(*apiHandler); ok {
			n.handler = h
		}
	}
}

type mainSrv struct {
	handler *apiHandler
}

func (s *mainSrv) Deps() []any {
	return []any{(*apiHandler)(nil)}
}

func (s *mainSrv) Inject(args []any) {
	for _, arg := range args {
		if h, ok := arg.(*apiHandler); ok {
			s.handler = h
		}
	}
}

type techSrv struct {
	handler *techHandler
}

func (s *techSrv) Deps() []any {
	return []any{(*techHandler)(nil)}
}

func (s *techSrv) Inject(args []any) {
	for _, arg := range args {
		if h, ok := arg.(*techHandler); ok {
			s.handler = h
		}
	}
}

func TestResolveConcreteMatching(t *testing.T) {
	t.Run("concrete type resolve", func(t *testing.T) {
		consumer := &concreteConsumer{}
		repo := &repoImpl{}

		if err := Resolve(testRegistry(consumer, repo)); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if consumer.repo != repo {
			t.Fatal("concrete dependency was not injected")
		}
	})

	t.Run("concrete type unresolved", func(t *testing.T) {
		consumer := &needsAPI{}
		other := &otherHandler{}

		err := Resolve(testRegistry(consumer, other))
		if err == nil || !strings.Contains(err.Error(), "unresolved dependency") {
			t.Errorf("expected unresolved error, got %v", err)
		}
	})

	t.Run("concrete type ambiguous", func(t *testing.T) {
		consumer := &concreteConsumer{}

		err := Resolve(testRegistry(
			consumer,
			&repoImpl{},
			&repoImpl{},
		))
		if !errors.Is(err, ErrAmbiguousDependency) {
			t.Errorf("expected ambiguous error, got %v", err)
		}
	})

	t.Run("two handlers via concrete type", func(t *testing.T) {
		api := &apiHandler{}
		tech := &techHandler{}
		main := &mainSrv{}
		technical := &techSrv{}

		if err := Resolve(testRegistry(api, tech, main, technical)); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if main.handler != api {
			t.Fatal("main server did not receive api handler")
		}
		if technical.handler != tech {
			t.Fatal("technical server did not receive tech handler")
		}
	})
}

type readiness interface {
	Ready() bool
}

type readyA struct{}

func (readyA) Ready() bool { return true }

type readyB struct{}

func (readyB) Ready() bool { return true }

type manyReadinessConsumer struct {
	items []readiness
}

func (m *manyReadinessConsumer) Deps() []any {
	return []any{([]readiness)(nil)}
}

func (m *manyReadinessConsumer) Inject(args []any) {
	for _, arg := range args {
		if v, ok := arg.([]readiness); ok {
			m.items = v
		}
	}
}

type manyRepoConsumer struct {
	repos []*repoImpl
}

func (c *manyRepoConsumer) Deps() []any {
	return []any{([]*repoImpl)(nil)}
}

func (c *manyRepoConsumer) Inject(args []any) {
	for _, arg := range args {
		if v, ok := arg.([]*repoImpl); ok {
			c.repos = v
		}
	}
}

func TestResolve_manyDependencies(t *testing.T) {
	t.Run("many injects all implementations", func(t *testing.T) {
		consumer := &manyReadinessConsumer{}
		a := readyA{}
		b := readyB{}

		if err := Resolve(testRegistry(consumer, a, b)); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(consumer.items) != 2 {
			t.Fatalf("expected 2 readiness items, got %d", len(consumer.items))
		}
	})

	t.Run("many empty slice when no implementors", func(t *testing.T) {
		consumer := &manyReadinessConsumer{}
		if err := Resolve(testRegistry(consumer)); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(consumer.items) != 0 {
			t.Fatalf("expected 0 readiness items, got %d", len(consumer.items))
		}
	})

	t.Run("many skips dedup for duplicate interfaces", func(t *testing.T) {
		consumer := &manyReadinessConsumer{}
		if err := Resolve(testRegistry(consumer, readyA{}, readyB{})); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("one still ambiguous with duplicates", func(t *testing.T) {
		err := Resolve(testRegistry(
			&mockService{},
			&repoImpl{},
			&repoImpl{},
		))
		if !errors.Is(err, ErrAmbiguousDependency) {
			t.Errorf("expected ambiguous error, got %v", err)
		}
	})

	t.Run("many concrete slice", func(t *testing.T) {
		consumer := &manyRepoConsumer{}
		r1 := &repoImpl{}
		r2 := &repoImpl{}

		if err := Resolve(testRegistry(consumer, r1, r2)); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(consumer.repos) != 2 {
			t.Fatalf("expected 2 repos, got %d", len(consumer.repos))
		}
	})
}
