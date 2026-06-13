package sdi

import "reflect"

type (
	DedupEntry struct {
		Value     any
		Removable bool // res: Origin == System
	}

	DedupContext struct {
		Interface reflect.Type
		Entries   []DedupEntry
		Remove    func(any) error
	}

	DedupPolicy func(DedupContext) error
)

// DefaultDedupPolicy resolves System+User pairs for interface ports before inject.
func DefaultDedupPolicy(ctx DedupContext) error {
	switch len(ctx.Entries) {
	case 0, 1:
		return nil
	case 2:
		var removable []any
		for _, e := range ctx.Entries {
			if e.Removable {
				removable = append(removable, e.Value)
			}
		}
		switch len(removable) {
		case 0:
			return ErrAmbiguousDependency
		case 1:
			return ctx.Remove(removable[0])
		default:
			return ErrMultipleSystemDefaults
		}
	default:
		return ErrTooManyImplementations
	}
}
