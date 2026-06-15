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

func prepareRegistry(reg res.Registry, policy DedupPolicy) error {
	for {
		concreteTypes, ifaces := collectDeps(reg)

		toRemove, err := planPreparation(reg, concreteTypes, ifaces, policy)
		if err != nil {
			return err
		}
		if len(toRemove) == 0 {
			return nil
		}
		for _, v := range toRemove {
			if err := reg.Remove(v); err != nil {
				return err
			}
		}
	}
}

func planPreparation(reg res.Registry, concreteTypes, ifaces []reflect.Type, policy DedupPolicy) ([]any, error) {
	concreteRemovals, err := planPolicy(reg, concreteTypes, reg.GetByType, policy)
	if err != nil {
		return nil, err
	}
	interfaceRemovals, err := planPolicy(reg, ifaces, reg.GetByInterface, policy)
	if err != nil {
		return nil, err
	}
	return dedupeRemovals(append(concreteRemovals, interfaceRemovals...)), nil
}

func planPolicy(reg res.Registry, types []reflect.Type, query queryFunc, policy DedupPolicy) ([]any, error) {
	if len(types) == 0 {
		return nil, nil
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
			return nil, err
		}
	}
	return toRemove, nil
}

func dedupeRemovals(items []any) []any {
	seen := make(map[any]struct{}, len(items))
	out := make([]any, 0, len(items))
	for _, v := range items {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
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
