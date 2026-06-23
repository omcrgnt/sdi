/*
Пакет sdi — wire-only DI: dedupe ports, topo-sort, Inject.

sdi не строит ресурсы. Принимает [res.Registry]; policy задаёт sdi,
executor — [res.Registry.Remove] по [res.Entry.Replaceable] ([res.TagReplaceable])
и [res.Entry.Fixed] ([res.TagFixed]).

Pipeline:

	ecfg.Parse → builder.Build(cfg, res.Default)
	sdi.Resolve(res.Default)
	runner / res.GetOneByInterface / res.GetOneByType

Resolve:

	1. cleanupConcretes  — [res.Registry.GetByType] по concrete One stubs из Deps
	2. collectDeps       — interface/concrete One stubs из Deps
	3. validateInterfaces — [res.Registry.GetByInterface] для One interface stubs
	4. wire              — topo-sort, Inject

Dependency stubs in Deps() []any:

	(*T)(nil)   — exactly one implementation (0 → error, 2+ → ambiguous)
	([]T)(nil)  — many; 0 implementors → empty []T (pool order preserved when non-empty)

Many stubs are not subject to dedup; multiple implementations may coexist.

Backlog: revisit Many policy — optional warn when 0 implementors (vs silent empty slice
or strict >=1 error); ops probe relies on empty slice for optional liveness/health ports.

DedupPolicy по умолчанию: Fixed+any duplicate → error;
Replaceable+explicit → Remove(Replaceable);
2×explicit / 2×Replaceable / ≥3 → error.

cleanupConcretes дедупит concrete типы из union: One stubs из Deps() (collectDeps)
и concrete типы с 2+ entries, где хотя бы один [res.TagReplaceable]
(collectDuplicateConcreteTypes). Runner-only Replaceable default + explicit override
(ops transport/http) dedup'ятся до wire без fake dep consumer. Many-deps и несколько
явных одного типа без Replaceable не затрагиваются.

Ошибки Resolve: ambiguous/multiple replaceable/too many/fixed conflict (шаги 1–3);
circular, unresolved (wiring).
*/
package sdi
