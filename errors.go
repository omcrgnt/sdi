package sdi

import "errors"

var (
	ErrTooManyImplementations = errors.New("too many implementations")
	ErrAmbiguousDependency    = errors.New("ambiguous dependency")
	ErrMultipleSystemDefaults = errors.New("multiple system defaults")
)
