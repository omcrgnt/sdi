# Known issues

## Dedup only for Deps stub types (backlog)

**Symptom:** Two resources of the same concrete type in pool (e.g. replaceable library default + explicit `Config.Build`) both survive `sdi.Resolve` and `runner` starts both.

**Cause:** `cleanupConcretes` / `validateInterfaces` run only for types collected from `Deps()` stubs (`collectDeps`). Runner-only resources that nothing depends on are never deduped.

**Workaround (rejected):** fake consumer with `Deps: (*T)(nil)` (e.g. ops `ServerAnchor`) — pollutes the dependency model.

**Fix (TODO):** pool-wide Replaceable dedup: for each concrete type in registry with 2+ entries, apply `DefaultDedupPolicy` even when no resource declares that type as a dep. Same policy rules (Replaceable + explicit → Remove replaceable).

**Consumer:** `github.com/omcrgnt/ops` `transport/http` — `DefaultServer` (TagReplaceable) + `Config.Build()` override.

## Flaky: `TestRun_printsGeneratedPath` (`cmd/sdigen`)

**Symptom:** `task test` / `go test ./cmd/sdigen/...` sometimes fails with:

```
tool_test.go:32: stdout = "", want Generated line for .../service_sdi_gen.go
```

Other generator tests in the same run pass and print `Generated: ...` to stdout.

**Cause:** `captureStdout` replaces process-global `os.Stdout` with a pipe and reads it in a goroutine. `Run()` logs via `fmt.Printf` to `os.Stdout`. Under load (e.g. Docker `golang:alpine`) the capture can miss output — timing / global stdout, not generator logic.

**Workaround:** `go test ./...` locally is usually fine; if sdigen fails intermittently, re-run or:

```bash
go test -count=1 ./cmd/sdigen/...
```

**Fix (TODO):** prefer one of:

- assert generated file exists instead of stdout;
- mutex around stdout redirection in tests;
- pass `io.Writer` into `Run()` instead of `fmt.Printf`.

**Not related to:** `-race`, `sdi` matching changes.
