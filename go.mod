module github.com/omcrgnt/sdi

go 1.26.2

retract (
	v1.3.2 // fix retract
	[v1.0.0, v1.20.0]
	v0.20.1 // broken imports
	[v0.1.0, v0.20.0]
)

require (
	github.com/omcrgnt/res v0.20.2
	golang.org/x/tools v0.46.0
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	golang.org/x/mod v0.37.0 // indirect
	golang.org/x/sync v0.21.0 // indirect
)
