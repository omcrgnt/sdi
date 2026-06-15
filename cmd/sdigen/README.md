# sdigen

Генерирует `Deps()` и `Inject()` для типов с embedded `deps`.

## Конвенция

В каждом пакете — одна структура `deps` с interface-полями (одноимённые поля):

```go
type repo interface { Get() }

type deps struct {
    repo repo
}

type service struct {
    deps
    mu sync.Mutex
}
```

Запуск из корня пакета:

```bash
go run github.com/omcrgnt/sdi/cmd/sdigen
```

Создаёт `*_sdi_gen.go` рядом с исходниками.

## Правила

- Сканируется только `type deps struct`
- В `Deps` попадают именованные поля с interface-типом
- Генерация для struct с anonymous embed `deps` (поля deps промотируются: `r.repo`)
- Несколько embedders в файле — каждый получает методы
- Два разных `deps` в одном пакете — не поддерживается (другой пакет)

### Примечание

Именованный embed `deps deps` технически поддерживается генератором и wiring
отработает, но по конвенции не используется: доступ через `r.deps.repo` вместо
`r.repo`, два стиля в одном проекте только путают. Используйте anonymous embed.

## Сгенерированный код

```go
func (r *service) Deps() []any {
    return []any{(*repo)(nil)}
}

func (r *service) Inject(args []any) {
    for _, arg := range args {
        switch v := arg.(type) {
        case repo:
            r.repo = v
        }
    }
}
```

После генерации: `res.Add` / `builder.Build` → `sdi.Resolve(res.Default)`.
