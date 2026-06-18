package sdi

import (
	"errors"
	"reflect"
	"testing"

	"github.com/omcrgnt/res"
)

func TestDefaultDedupPolicy(t *testing.T) {
	depType := reflect.TypeFor[mockRepo]()

	t.Run("0 and 1 entries ok", func(t *testing.T) {
		if err := DefaultDedupPolicy(DedupContext{DepType: depType}); err != nil {
			t.Fatal(err)
		}
		if err := DefaultDedupPolicy(DedupContext{
			DepType: depType,
			Entries: []DedupEntry{{Value: "only"}},
		}); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("replaceable+explicit removes replaceable", func(t *testing.T) {
		replaceable := &repoImpl{}
		explicit := &repoImpl{}
		var removed any
		err := DefaultDedupPolicy(DedupContext{
			DepType: depType,
			Entries: []DedupEntry{
				{Value: replaceable, Replaceable: true},
				{Value: explicit, Replaceable: false},
			},
			Remove: func(v any) error {
				removed = v
				return nil
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if removed != replaceable {
			t.Fatalf("expected replaceable removed, got %v", removed)
		}
	})

	t.Run("two explicit ambiguous", func(t *testing.T) {
		err := DefaultDedupPolicy(DedupContext{
			DepType: depType,
			Entries: []DedupEntry{
				{Value: &repoImpl{}, Replaceable: false},
				{Value: &repoImpl{}, Replaceable: false},
			},
		})
		if !errors.Is(err, ErrAmbiguousDependency) {
			t.Fatalf("expected ambiguous, got %v", err)
		}
	})

	t.Run("two replaceable", func(t *testing.T) {
		err := DefaultDedupPolicy(DedupContext{
			DepType: depType,
			Entries: []DedupEntry{
				{Value: &repoImpl{}, Replaceable: true},
				{Value: &repoImpl{}, Replaceable: true},
			},
			Remove: func(any) error { return nil },
		})
		if !errors.Is(err, ErrMultipleReplaceable) {
			t.Fatalf("expected multiple replaceable, got %v", err)
		}
	})

	t.Run("three entries", func(t *testing.T) {
		err := DefaultDedupPolicy(DedupContext{
			DepType: depType,
			Entries:   []DedupEntry{{}, {}, {}},
		})
		if !errors.Is(err, ErrTooManyImplementations) {
			t.Fatalf("expected too many, got %v", err)
		}
	})

	t.Run("fixed with duplicate", func(t *testing.T) {
		err := DefaultDedupPolicy(DedupContext{
			DepType: depType,
			Entries: []DedupEntry{
				{Value: &repoImpl{}, Fixed: true},
				{Value: &repoImpl{}},
			},
		})
		if !errors.Is(err, ErrFixedResourceConflict) {
			t.Fatalf("expected fixed conflict, got %v", err)
		}
	})

	t.Run("fixed with replaceable", func(t *testing.T) {
		err := DefaultDedupPolicy(DedupContext{
			DepType: depType,
			Entries: []DedupEntry{
				{Value: &repoImpl{}, Fixed: true},
				{Value: &repoImpl{}, Replaceable: true},
			},
			Remove: func(any) error { return nil },
		})
		if !errors.Is(err, ErrFixedResourceConflict) {
			t.Fatalf("expected fixed conflict, got %v", err)
		}
	})
}

func TestResolve_dedupReplaceableInterface(t *testing.T) {
	defaultRepo := &repoImpl{}
	userRepo := &repoImpl{}
	svc := &mockService{}

	reg := res.New()
	if err := reg.AddWithTags(defaultRepo, res.TagReplaceable); err != nil {
		t.Fatal(err)
	}
	if err := reg.Add(userRepo); err != nil {
		t.Fatal(err)
	}
	if err := reg.Add(svc); err != nil {
		t.Fatal(err)
	}

	if err := Resolve(reg); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if svc.repo != userRepo {
		t.Fatal("expected user repo injected after dedup removed replaceable default")
	}
	if len(reg.GetByInterface(reflect.TypeFor[mockRepo]())) != 1 {
		t.Fatal("expected replaceable default removed from registry")
	}
}

func TestResolve_dedupReplaceableConcrete(t *testing.T) {
	defaultRepo := &repoImpl{}
	userRepo := &repoImpl{}
	consumer := &concreteConsumer{}

	reg := res.New()
	if err := reg.AddWithTags(defaultRepo, res.TagReplaceable); err != nil {
		t.Fatal(err)
	}
	if err := reg.Add(userRepo); err != nil {
		t.Fatal(err)
	}
	if err := reg.Add(consumer); err != nil {
		t.Fatal(err)
	}

	if err := Resolve(reg); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if consumer.repo != userRepo {
		t.Fatal("expected user concrete repo injected after cleanup removed replaceable default")
	}
	if len(reg.GetByType(reflect.TypeFor[*repoImpl]())) != 1 {
		t.Fatal("expected replaceable default removed from registry")
	}
}

func TestResolve_fixedConcreteConflict(t *testing.T) {
	fixed := &repoImpl{}
	other := &repoImpl{}
	consumer := &concreteConsumer{}

	reg := res.New()
	if err := reg.AddWithTags(fixed, res.TagFixed); err != nil {
		t.Fatal(err)
	}
	if err := reg.Add(other); err != nil {
		t.Fatal(err)
	}
	if err := reg.Add(consumer); err != nil {
		t.Fatal(err)
	}

	err := Resolve(reg)
	if !errors.Is(err, ErrFixedResourceConflict) {
		t.Fatalf("expected fixed conflict, got %v", err)
	}
}

func TestResolve_fixedBlocksReplaceableDedup(t *testing.T) {
	fixed := &repoImpl{}
	defaultRepo := &repoImpl{}
	consumer := &concreteConsumer{}

	reg := res.New()
	if err := reg.AddWithTags(fixed, res.TagFixed); err != nil {
		t.Fatal(err)
	}
	if err := reg.AddWithTags(defaultRepo, res.TagReplaceable); err != nil {
		t.Fatal(err)
	}
	if err := reg.Add(consumer); err != nil {
		t.Fatal(err)
	}

	err := Resolve(reg)
	if !errors.Is(err, ErrFixedResourceConflict) {
		t.Fatalf("expected fixed conflict, got %v", err)
	}
}
