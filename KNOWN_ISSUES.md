# Known issues

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
