package sdi

import "errors"

var (
	ErrTooManyImplementations = errors.New("too many implementations")
	ErrAmbiguousDependency    = errors.New("ambiguous dependency")
	ErrMultipleReplaceable    = errors.New("multiple replaceable defaults")
	ErrFixedResourceConflict  = errors.New("fixed resource conflict")
)
