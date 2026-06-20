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

Сейчас dedup (cleanupConcretes / validateInterfaces) вызывается только для типов,
собранных из Deps() stubs (collectDeps). Ресурсы в pool без dep-stub на их тип
(например runner-only *systemServer) не dedup'ятся — Replaceable default + explicit
override требуют pool-wide dedup по concrete type (backlog, см. ops transport/http).

Ошибки Resolve: ambiguous/multiple replaceable/too many/fixed conflict (шаги 1–3);
circular, unresolved (wiring).
*/
package sdi
