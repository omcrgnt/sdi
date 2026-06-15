/*
Пакет sdi — wire-only DI: dedupe ports, topo-sort, Inject.

sdi не строит ресурсы. Принимает [res.Registry]; policy задаёт sdi,
executor — [res.Registry.Remove] по [res.Entry.Replaceable] ([res.TagReplaceable]).

Pipeline:

	ecfg.Parse → builder.Build(cfg, res.Default)
	sdi.Resolve(res.Default)
	runner / res.GetOneByInterface / res.GetOneByType

Resolve (подготовка + wiring):

	1–3. prepareRegistry — plan concrete + interface на одном снимке registry;
	     при ошибке без Remove; при успехе batch Remove и повтор, пока есть removals
	4. wire — topo-sort, Inject

DedupPolicy по умолчанию: Replaceable+explicit → Remove(Replaceable);
2×explicit / 2×Replaceable / ≥3 → error.

Ошибки Resolve: ambiguous/multiple replaceable/too many (фазы 1–3);
circular, unresolved (wiring).
*/
package sdi
