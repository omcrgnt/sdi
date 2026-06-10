/*
Пакет sdi — wire-only DI: топологическая сортировка и Inject по графу зависимостей.

sdi не строит ресурсы и не хранит их. Источник — Pool с read-only обходом Walk.

Типичный pipeline (оркестрация в main):

	ecfg.Parse → builder.Build(cfg, res.Default)
	res.Transform(...)
	sdi.Resolve(res.Default)
	res.Get / res.Find

Контракт ресурса для wiring:

  - Deps() []any — stubs типов:
      interface: (*Repo)(nil) или (Repo)(nil) — ищется реализация в pool
      concrete:  (*API)(nil) или (T)(nil) — exact match типа в pool
  - Inject([]any) — присваивание из pool

Генерация Deps/Inject: cmd/sdigen — конвенция type deps struct + embed deps.

Ошибки Resolve: circular dependency, unresolved dependency, ambiguous dependency.
*/
package sdi
