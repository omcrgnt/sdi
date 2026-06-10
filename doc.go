/*
Пакет sdi — wire-only DI: топологическая сортировка и Inject по графу зависимостей.

sdi не строит ресурсы и не хранит их. Источник — Pool с read-only обходом Walk.

Типичный pipeline (оркестрация в main):

	ecfg.Parse → builder.Build(cfg, res.Default)
	res.Transform(...)
	sdi.Resolve(res.Default)
	res.Get / res.Find

Контракт ресурса для wiring:

  - Deps() []any — stubs типов: (*Repo)(nil)
  - Inject([]any) — присваивание из pool

Генерация Deps/Inject: cmd/sdigen — конвенция type deps struct + embed deps.

Ошибки Resolve: circular dependency, unresolved dependency, ambiguous dependency.
*/
package sdi
