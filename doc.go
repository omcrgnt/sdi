/*
Пакет sdi — wire-only DI: cycle check, Inject.

sdi не строит и не изменяет ресурсы в pool. Принимает [Registry] (read-only [res.Entry] walk).

Pipeline:

	fill → materialize → unique.Merge → sdi.Resolve(reg)
	runner / res.GetOneByInterface / res.GetOneByType

Resolve:

	1. collect entries in registration order ([Registry.WalkEntries])
	2. checkCycles — DFS по Deps(); circular → error
	3. inject — registration order; one-dep match; many → slice in registration order

Dependency stubs in Deps() []any:

	(*T)(nil)   — exactly one (0 → unresolved, 2+ → ambiguous)
	([]T)(nil)  — many; 0 implementors → empty []T (registration order when non-empty)

Override Replaceable+explicit того же concrete type — [unique.Add] / Merge до Resolve, не sdi.

Backlog: optional public DependencyOrder; Many policy warn on 0 implementors.
*/
package sdi
