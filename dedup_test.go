package sdi

import (
	"errors"
	"reflect"
	"testing"
)

func TestDefaultDedupPolicy(t *testing.T) {
	iface := reflect.TypeFor[mockRepo]()

	t.Run("0 and 1 entries ok", func(t *testing.T) {
		if err := DefaultDedupPolicy(DedupContext{Interface: iface}); err != nil {
			t.Fatal(err)
		}
		if err := DefaultDedupPolicy(DedupContext{
			Interface: iface,
			Entries:   []DedupEntry{{Value: "only"}},
		}); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("system+user removes system", func(t *testing.T) {
		system := &repoImpl{}
		user := &repoImpl{}
		var removed any
		err := DefaultDedupPolicy(DedupContext{
			Interface: iface,
			Entries: []DedupEntry{
				{Value: system, Removable: true},
				{Value: user, Removable: false},
			},
			Remove: func(v any) error {
				removed = v
				return nil
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if removed != system {
			t.Fatalf("expected system removed, got %v", removed)
		}
	})

	t.Run("two user ambiguous", func(t *testing.T) {
		err := DefaultDedupPolicy(DedupContext{
			Interface: iface,
			Entries: []DedupEntry{
				{Value: &repoImpl{}, Removable: false},
				{Value: &repoImpl{}, Removable: false},
			},
		})
		if !errors.Is(err, ErrAmbiguousDependency) {
			t.Fatalf("expected ambiguous, got %v", err)
		}
	})

	t.Run("two system", func(t *testing.T) {
		err := DefaultDedupPolicy(DedupContext{
			Interface: iface,
			Entries: []DedupEntry{
				{Value: &repoImpl{}, Removable: true},
				{Value: &repoImpl{}, Removable: true},
			},
			Remove: func(any) error { return nil },
		})
		if !errors.Is(err, ErrMultipleSystemDefaults) {
			t.Fatalf("expected multiple system, got %v", err)
		}
	})

	t.Run("three entries", func(t *testing.T) {
		err := DefaultDedupPolicy(DedupContext{
			Interface: iface,
			Entries:   []DedupEntry{{}, {}, {}},
		})
		if !errors.Is(err, ErrTooManyImplementations) {
			t.Fatalf("expected too many, got %v", err)
		}
	})
}
