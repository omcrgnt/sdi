package sdi

import "github.com/omcrgnt/res"

// Registry is the read-only pool view required for [Resolve].
type Registry interface {
	WalkEntries(func(res.Entry) bool)
}
