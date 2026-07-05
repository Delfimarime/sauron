# E2E runner flags — feature selection + concurrency

## What I implemented

1. **Feature-file selection** (`-features`): a comma-separated list of feature-file
   basenames. Empty (default) runs the whole rendered `testdata/` directory. An
   unknown basename is a hard error (fail loudly on typos).
   - `resolveFeaturePaths(dir, raw string) ([]string, error)` in
     `test/e2e/helper_app.go`.

2. **Concurrency** (`-concurrency`): `0` (default) or any non-positive value
   resolves to the CPU-aware default `min(5, max(1, runtime.NumCPU()/2))`; a
   positive value is an explicit override, even above 5.
   - `resolveConcurrency(explicit int) int` in `test/e2e/helper_app.go`.
   - Put both helpers in `helper_app.go`, not `integration_test.go`, because
     `integration_test.go` already imports this module's own
     `.../internal/runtime` package as `runtime` — adding stdlib `"runtime"`
     there for `resolveConcurrency` would collide.

3. **`TestFeatures` wiring** (`test/e2e/integration_test.go`):
   - Package-level flags `featuresFlag` (`flag.String`) and `concurrencyFlag`
     (`flag.Int`), declared before `flag.Parse()` runs (Go's test binary calls
     it automatically).
   - Resolves `paths := resolveFeaturePaths(...)` (fails fast via `t.Fatal` on
     error) and `concurrency := resolveConcurrency(...)`, then sets
     `Options.Paths`, `Options.Concurrency`, and `Options.Format` accordingly.
   - Format selection: `"pretty"` when `concurrency == 1`, `"progress"` when
     `concurrency > 1` (see the formatter finding below).

4. **Taskfile wiring** (`Taskfile.yml`, `gate-integration` task): added
   `FEATURES` / `CONCURRENCY` task vars (defaulting to `""` / `"0"`, following
   the file's existing `'{{.X | default "Y"}}'` pattern) and appended
   `-features="{{.FEATURES}}" -concurrency={{.CONCURRENCY}}` to the `go test
   ./...` invocation. `task gate-integration FEATURES=list_skills.feature`,
   `task gate-integration CONCURRENCY=1`, and plain `task gate-integration`
   (now parallel by default) all work.

5. **`test/e2e/HARNESS.md`**: added a short paragraph to "The integration gate"
   section documenting the two flags, the default-concurrency formula, and the
   formatter's concurrency-dependent behavior, so the harness doc doesn't go
   stale relative to `integration_test.go`.

## TDD evidence (resolveFeaturePaths / resolveConcurrency)

**RED** — wrote `test/e2e/helper_app_test.go` first, calling the not-yet-existing
functions. Confirmed compile failure:

```
$ go test -tags unit ./...
./helper_app_test.go:57:16: undefined: resolveFeaturePaths
./helper_app_test.go:71:22: undefined: resolveConcurrency
...
FAIL	github.com/delfimarime/sauron/test/e2e [build failed]
```

**GREEN** — implemented both functions in `helper_app.go`, then:

```
$ go test -tags unit -v -run 'TestResolveFeaturePaths|TestResolveConcurrency' .
--- PASS: TestResolveFeaturePaths (0.00s)
    --- PASS: .../empty_raw_returns_the_whole_directory (0.00s)
    --- PASS: .../single_valid_name_resolves_to_a_joined_path (0.00s)
    --- PASS: .../multiple_valid_names_with_surrounding_whitespace_resolve_in_order (0.00s)
    --- PASS: .../unknown_name_is_a_hard_error (0.00s)
    --- PASS: .../a_raw_string_resolving_to_nothing_is_an_error (0.00s)
--- PASS: TestResolveConcurrency (0.00s)
    --- PASS: .../explicit_positive_value_is_returned_as-is,_even_above_the_cap (0.00s)
    --- PASS: .../explicit_zero_falls_back_to_the_CPU-aware_default (0.00s)
    --- PASS: .../explicit_negative_falls_back_to_the_CPU-aware_default (0.00s)
    --- PASS: .../default_matches_the_documented_formula (0.00s)
PASS
```

`helper_app_test.go` carries no build tag (no Docker, no `!unit`), so it also
runs under the default `go test ./...` invocation the gate uses, alongside
`TestFeatures`.

## The formatter question — resolved empirically

Built both binaries (`sauron-e2e-darwin-arm64`, `sauron-e2e-linux-arm64`) via
`task gate-integration` and ran the actual suite against a live Docker daemon
(OrbStack) at `-concurrency=3`.

- **"progress" at concurrency 3** (full suite, 12 feature files): output was
  coherent — one colored `.` per passed step, `=== RUN` lines from `go test
  -v`'s own subtest reporting interleaved between them, but no scenario body
  was ever split or garbled. Full run: `ok ... 183.158s`, all scenarios PASS.

- **"pretty" at concurrency 3** (forced temporarily, subset of 4 feature files
  with multiple scenarios each, to get a fast repro): confirmed garbling.
  Captured evidence — a `Scenario: lists every installed agent` block was
  mid-way through printing its `Given` step's multi-line YAML body when
  `=== RUN TestFeatures/reports_an_empty_page_when_paged_past_the_end` (a
  *different*, concurrently-running scenario) was injected into the stream,
  immediately followed by the start of a third scenario's `Scenario:` header
  — before the first scenario's block had finished. This is exactly the
  interleaving failure mode the brief flagged as a known community concern.

**Decision**: kept the conditional switch exactly as specified —
`"pretty"` when `concurrency == 1` (unchanged, familiar local-debugging
output), `"progress"` when `concurrency > 1` (coherent under concurrent
scenarios). This is not a hypothetical; it's confirmed by direct observation
on this suite.

(The temporary force-to-`"pretty"` edit used for the repro was reverted
immediately after capturing the evidence; the shipped code only contains the
conditional `format := "pretty"; if concurrency > 1 { format = "progress" }`.)

## `task gate-integration` runs performed

All runs used a real Docker daemon (OrbStack) and the version-stamped
darwin/arm64 + linux/arm64 binaries the task builds.

1. `task gate-integration FEATURES=version.feature` — single-feature,
   default concurrency (`0` → CPU-aware). `ok ... 1.387s`. Confirms
   `-features` scoping works end-to-end.
2. `task gate-integration FEATURES=version.feature CONCURRENCY=1` — serial
   override. `ok ... 0.993s`. Confirms `-concurrency=1` (and the resulting
   `"pretty"` format) still works.
3. `task gate-integration` (no overrides) — full suite, default concurrency.
   Observed 4 concurrent Docker Compose stacks (`docker ps` showed 4
   `*-main-1` containers + the Ryuk reaper at once, consistent with
   `min(5, max(1, NumCPU()/2))` on this machine), `"progress"` format engaged
   automatically. `ok ... 108.899s`, all scenarios PASS — faster than the
   forced `-concurrency=3` run of the same suite (`183.158s`), as expected.
4. (Formatter repro only, not a gate run) `-features=<4 files>
   -concurrency=3` with `"pretty"` forced — used solely to capture the
   interleaving evidence above; killed early once the evidence was captured
   (all scenarios that had completed by then were PASS).

`golangci-lint run ./...` and `gofmt -l` are clean on every file this change
touches (`helper_app.go`, `helper_app_test.go`, `integration_test.go`).
`gofmt -l .` does flag one pre-existing file
(`internal/gherkin/content_test.go`) — unrelated to this change, untouched by
it, left as-is per the "touch only what's needed" discipline.

## Files changed

- `test/e2e/helper_app.go` — added `resolveFeaturePaths`, `resolveConcurrency`.
- `test/e2e/helper_app_test.go` — new; table-driven tests for both (testify,
  no build tag).
- `test/e2e/integration_test.go` — new flags, wiring into `TestFeatures`
  (`Paths`, `Concurrency`, conditional `Format`).
- `test/e2e/HARNESS.md` — documented the two flags and the formatter behavior
  in "The integration gate" section.
- `Taskfile.yml` — `gate-integration` task: `FEATURES`/`CONCURRENCY` vars,
  appended `-features=... -concurrency=...` to the `go test` invocation.

## Self-review findings / concerns

- **Scope note (flagging, not hiding):** my operating instructions say "touch
  only `test/e2e/**`," but the brief's item 5 explicitly and repeatedly
  directs editing the root `Taskfile.yml`'s `gate-integration` task (with
  exact syntax to match). I treated the brief as an explicit, well-justified
  exception for this one file — the flags are meaningless without the
  Taskfile wiring the brief itself designed — rather than silently
  refusing or silently guessing. Flagging this here so it's visible rather
  than buried.
- Verified `go test ./...` (no `-tags`) only ever compiles a real test binary
  for the top-level `e2e` package under the gate's invocation — every
  `internal/...` package's test files carry `//go:build unit` and are absent
  otherwise (`? ... [no test files]` in every run above) — so passing
  `-features`/`-concurrency` on `go test ./...` never hits an "flag provided
  but not defined" error in a sibling package. This was a real risk I
  checked for, not an assumption.
- `resolveFeaturePaths` validates each name with `os.Stat` before adding it —
  a directory named like a feature file would also pass `os.Stat` and get
  handed to godog as a path; godog would then try to parse it as a directory
  (which is actually fine, since godog treats non-files as directories) or
  error on ambiguity. This isn't a real-world case (feature basenames are
  always `*.feature` files rendered by `renderFeatures`), so I did not add
  extra validation beyond what the brief specified.
- Did not attempt an end-to-end test of `TestFeatures` itself (per the brief);
  all verification of the wiring went through actual `task gate-integration`
  runs instead, which exercises the real path including flag parsing,
  `godog.TestSuite`, and Docker.
- Confirmed `Concurrency: 0` in godog's own `Options` is never actually sent
  now — `resolveConcurrency` always returns `>= 1` — so today's implicit
  "concurrency 0 = serial" behavior in godog is bypassed entirely by our
  explicit resolution, matching the brief's intent.
