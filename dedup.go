package sdi

import (
	"reflect"

	"github.com/omcrgnt/res"
)

type (
	DedupEntry struct {
		Value       any
		Replaceable bool // res: [res.TagReplaceable]
	}

	DedupContext struct {
		DepType reflect.Type
		Entries []DedupEntry
		Remove  func(any) error
	}

	DedupPolicy func(DedupContext) error

	queryFunc func(reflect.Type) []res.Entry
)

func applyPolicy(reg res.Registry, types []reflect.Type, query queryFunc, policy DedupPolicy) error {
	if len(types) == 0 {
		return nil
	}

	var toRemove []any
	planRemove := func(v any) error {
		toRemove = append(toRemove, v)
		return nil
	}

	for _, t := range types {
		resEntries := query(t)
		dedupEntries := make([]DedupEntry, len(resEntries))
		for i, e := range resEntries {
			dedupEntries[i] = DedupEntry{Value: e.Value, Replaceable: e.Replaceable()}
		}
		if err := policy(DedupContext{
			DepType: t,
			Entries: dedupEntries,
			Remove:  planRemove,
		}); err != nil {
			return err
		}
	}

	for _, v := range toRemove {
		if err := reg.Remove(v); err != nil {
			return err
		}
	}
	return nil
}

func cleanupConcretes(reg res.Registry, types []reflect.Type, policy DedupPolicy) error {
	return applyPolicy(reg, types, reg.GetByType, policy)
}

func validateInterfaces(reg res.Registry, types []reflect.Type, policy DedupPolicy) error {
	return applyPolicy(reg, types, reg.GetByInterface, policy)
}

// DefaultDedupPolicy resolves Replaceable+explicit pairs before inject.
func DefaultDedupPolicy(ctx DedupContext) error {
	switch len(ctx.Entries) {
	case 0, 1:
		return nil
	case 2:
		var replaceable []any
		for _, e := range ctx.Entries {
			if e.Replaceable {
				replaceable = append(replaceable, e.Value)
			}
		}
		switch len(replaceable) {
		case 0:
			return ErrAmbiguousDependency
		case 1:
			return ctx.Remove(replaceable[0])
		default:
			return ErrMultipleReplaceable
		}
	default:
		return ErrTooManyImplementations
	}
}
