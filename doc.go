/*
Пакет sdi — wire-only DI: dedupe interface ports, topo-sort, Inject.

sdi не строит ресурсы. Pool — Walk + Dedup (policy задаёт sdi, executor — res).

Pipeline:

	ecfg.Parse → builder.Build(cfg, res.Default)
	pool.Dedup(collectInterfaceDeps, DefaultDedupPolicy)  // inside Resolve
	sdi.Resolve(res.Default)
	runner / res.Get

DedupPolicy по умолчанию: System+User → Remove(System); 2×User / 2×System / ≥3 → error.

Ошибки Resolve: circular, unresolved, ambiguous dependency.
*/
package sdi
