/*
Пакет sdi — wire-only DI: dedupe ports, topo-sort, Inject.

sdi не строит ресурсы. Принимает [res.Registry]; policy задаёт sdi,
executor — [res.Registry.Remove] по [res.Entry.Replaceable] ([res.TagReplaceable]).

Pipeline:

	ecfg.Parse → builder.Build(cfg, res.Default)
	sdi.Resolve(res.Default)
	runner / res.GetOneByInterface / res.GetOneByType

Resolve (три фазы подготовки + wiring):

	1. cleanupConcretes  — [res.Registry.GetByType] по concrete stubs из Deps
	2. collectDeps       — interface stubs из Deps (без мутации)
	3. validateInterfaces — [res.Registry.GetByInterface]
	4. wire              — topo-sort, Inject

DedupPolicy по умолчанию: Replaceable+explicit → Remove(Replaceable);
2×explicit / 2×Replaceable / ≥3 → error.

Ошибки Resolve: ambiguous/multiple replaceable/too many (фазы 1–3);
circular, unresolved (wiring).
*/
package sdi
