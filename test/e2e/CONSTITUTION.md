# Constitution — Integration Tests (`test/e2e`)

Governing principles for sauron's black-box integration suite. This document is
the test-suite counterpart to the project [Constitution](../CONSTITUTION.md):
the root Constitution governs the product; this one governs the harness that
drives it end-to-end. It is subordinate to the root Constitution and the
[architecture contract](../spec/contracts/architecture.md) *Integration tests*
section, and it is the single governing document for `test/e2e/**` — its intent,
its architecture, and the principles the harness must never violate.

> Scope: everything under `test/e2e/**`. Where a rule here and the root
> Constitution appear to conflict, the root wins.

## Chapter I — Intent

### Article 1 — Black-box, graybox assertions

The suite drives the **built binary** exactly as an external operator would —
spawning the process, passing CLI args, reading its exit code, stdout/stderr,
and the state files it persists. It never calls a use case, an action,
or anything under `internal/` in-process. "Graybox" means the *assertions* are
allowed to decode the binary's output into the public `pkg/sauron/types` DTOs,
but the *exercise* is strictly external.

### Article 2 — One binary, three transports, one set of assertions

The single behaviour under test is `sauron add registry` across its three
transports — `filesystem`, `http`, `git`. The action and the assertions are
**transport-agnostic**: the same `When` adds the registry and the same `Then`
inspects the persisted config, no matter how the registry was sourced. Only the
`Given` fixtures differ.

### Article 3 — Red is the bar until the product lands

The harness is built **spec-first, TDD**. It is correct when every step
resolves, every source provisions, and every file is read — and the *only*
failure is the not-yet-built command. A scenario that fails at harness wiring
(an undefined step, a panic, an unprovisioned source) is a harness defect; a
scenario that fails because `add registry` exits non-zero is the suite working
as designed. The suite turns green when the production track completes, with no
harness change.

```
  intent                       not intent
  ───────────────────────────  ───────────────────────────
  exec the real binary         call internal/ in-process
  assert via pkg/ types        mirror types locally
  fail only at the command     fail at step/source wiring
  hermetic (Testcontainers)    depend on the public internet
```

## Chapter II — Architecture

### Article 1 — A separate module

`test/e2e` is its own Go module (`.../sauron/test/e2e`) that resolves the root
through `replace … => ../..`. Its heavy dependencies — godog, Testcontainers,
testify — live **only** here and never leak into the root module or its
approved-dependency table. The suite runs under `go test` (no `main`); the
integration entrypoint is tagged `//go:build !unit` and the in-process unit
tests `//go:build unit`, so the two never run together.

```
  test/e2e/
    go.mod                  module .../test/e2e ; replace root => ../..
    integration_test.go     godog entrypoint (TestFeatures) ; build !unit
    runtime.go              compositionBasedRuntime  (the per-scenario proxy)
    helper_app.go           {{.App.*}} feature-load templating
    testdata/
      *.feature             one feature per behaviour
      registries/acme/…     authored content sets (.skills/.agents/.personas)
    internal/
      gherkin/              controllers (step definitions) + the #{} resolver
      runtime/              the wide Runtime contract + host & docker backends
    .golangci.yml           depguard: ban .../internal and .../cmd
```

### Article 2 — The runtime is one wide interface and the shared state

There is **no `world.go`**. The runtime *is* the only object every controller
shares, so it owns the per-scenario state: the provisioned sources and their
addresses. `compositionBasedRuntime` is a dumb proxy forwarding to the
tag-selected backend; both backends implement the **same wide interface**.

```
   interface Runtime  ── the per-scenario shared-state owner (this is why there is no world.go)
   ├─ Execute / ReadFile / CopyTo
   ├─ Folder(alias)    → Source { Path() }
   ├─ Webserver(alias) → Source { URL()  }
   ├─ Git(alias)       → Source { URL()  }     (deferred: errors)
   └─ Start / Stop / IsReadOnly

   interface Source ── a declared, resource-loaded exposure of provider content
   ├─ Expose(resources…)   declare what it serves (content files; webserver auth) — never a port
   ├─ Path(ctx)            local directory   (folder)
   └─ URL(ctx)             network address   (webserver / git)
```

### Article 3 — Capability gaps are errors, not type-asserts

A capability a backend cannot satisfy returns an **error** from the relevant
`Source` accessor. There is no `Pod` sub-interface and no `rt.(Pod)` assertion —
a gap is an honest error, surfaced exactly where a scenario needs the address.

```
                     ┌──── host  (@no-sandbox) ────┬──── docker (default) ────┐
   Execute / ReadFile │  ✅                          │  ✅                      │
   CopyTo             │  ✅ (per-scenario temp dir)  │  ✅ (into "main")        │
   Folder(…).Path()   │  ✅ owns a local temp dir    │  ✅ path inside "main"   │
   Webserver(…).URL() │  ❌ error "not on @no-sandbox"│ ✅ nginx sidecar        │
   Git(…).URL()       │  ❌ error (deferred)         │  ❌ error (deferred)     │
                      └─────────────────────────────┴──────────────────────────┘
                            IsReadOnly() = true            IsReadOnly() = false
```

The host backend is **execution + a local folder and nothing networked**, which
keeps `IsReadOnly() == true` honest: a host that stood up servers would no
longer be read-only. Under docker, a `folderSource` and a `webserverSource` are
distinct types (each implements only its meaningful accessor and errors the
other) — the same capability-gap-as-error principle applied within a backend.

### Article 4 — Controllers are thin; they hold only the runtime

Every step definition lives on a `Controller` registered through `gherkin.Init`.
Controllers translate Gherkin into runtime calls and assertions; they hold **no
cross-controller state** — the runtime does. The lone exception is the command
controller's last-command result, which is consumed only there and so never
becomes a shared "world".

```
  integration_test.go
    └─ godog.TestSuite( ScenarioInitializer = CreateInitFunc(home, $SAURON_BIN, gherkin.Init) )
             │  once per scenario: pick backend by tag, attach to the proxy
             ▼
      gherkin.Init(sc, rt)
             ▼
   ┌──────────────── Controllers (godog step registrars) ────────────────┐
   │ basic   command   state           registry-fs   registry-http   git │
   └────┬───────┬───────────┬──────────────┬──────────────┬──────────────┘
        └───────┴───────────┴──────────────┴──────────────┘
                              │ all share ONE handle; resolve via valueOf[T]
                              ▼
                 compositionBasedRuntime  (per-scenario PROXY; lazy start)
                              │ forwards to the backend chosen by tag
                  ┌───────────┴───────────┐
                  ▼                       ▼
            host.Runtime            docker.Runtime
         os/exec $SAURON_BIN     compose: main(/opt/bin/sauron) + sidecars
```

The filesystem, http, and git fixtures share one `sourceFixture` (declare +
content steps); they differ only in the source they select and their wording.

## Chapter III — Principles

### Article 1 — `pkg/` only, enforced by depguard

The harness imports the public `pkg/` surface (`pkg/sauron/types`) and **never**
`internal/` or `cmd/`. Go's `internal/` rule does not stop this (shared module
prefix), so the `depguard` rule in `.golangci.yml` is the real guard and must
stay.

### Article 2 — Two templating syntaxes that never overlap

`{{…}}` is **build identity**, rendered at feature-load time against the `App`
context (e.g. `{{.App.FullVersion}}`). `#{…}` is a **dynamic runtime reference**,
resolved at step time against live state. They occupy disjoint roles.

```
  feature load                              step time
  ─────────────                             ─────────
  {{.App.FullVersion}}  ── helper_app.go    #{.webserver.default.url}
  build identity            text/template    │
  (version, hash, name)                      ▼
                                   valueOf[T](ctx, rt, raw)   ← the ONLY resolver
                                     parse .cap[.alias].attr  (default alias = "default")
                                     route to runtime accessor:
                                       .folder.…    → Folder(alias).Path()
                                       .webserver.… → Webserver(alias).URL()
                                       .git.…       → Git(alias).URL()  → error
                                     convert/assert to T → (T, error)
```

Resolution lives **only** in the gherkin `valueOf[T]` helper — the runtime owns
the data, the helper owns the parsing and typing. No controller stashes
addresses in a map (that map would be `world.go` reborn).

### Article 3 — Declare, then the first need materializes everything

`Given` steps only **accumulate**: they declare a source by alias and customize
it with the **resources** it exposes (never a port). Nothing runs until the
first *need* — the first `Execute()` **or** the first attribute access
(`Path`/`URL`, reached through `#{…}`). That first need triggers one compose
`Up`. A scenario may not declare a source after its first need; Gherkin orders
all `Given`s before the first `When`, so this is natural.

```
  GIVEN (accumulate only — nothing running) ───────────────────┐
    Given an http server hosting a registry   rt.Webserver(…)  │ declare + Expose
    And   …hosts the directory testdata/…     .Expose(res…)    │ (resources, never ports)
  ── first NEED crosses the line ─────────────────────────────┘
       trigger = first Execute()  OR  first attribute access (URL()/Path())
       ▼
    ┌─────────────────────────────────────────────────────┐
    │ Start(): compose Up (main + every accumulated sidecar)│ ← ports/paths assigned HERE
    └─────────────────────────────────────────────────────┘
       ❗ no new source may be declared after this point
       ▼
  WHEN  the user adds the registry from #{.webserver.default.url}
       ▼
  THEN  ReadFile registries.yaml → decode pkg/sauron/types → assert
```

### Article 4 — One content set, three exposures

`Given … hosts a skill/the directory/the file` builds **one provider content
set** (`.skills/`, `.agents/`, `.personas/`). A source capability is one exposure
of that set; the `add registry` URI is always `#{.<source>.<attr>}`. Content
authored under `testdata/` is read in-process and carried as inline bytes, so the
exposure is identical on the host folder and inside the container — no
dependence on the Docker daemon seeing a host path.

```
        "hosts the directory testdata/registries/acme"
                       │ ONE content set
                       ▼
              ┌──────────────────┐
              │  provider content │  .skills/  .agents/  .personas/
              └────────┬──────────┘
        ┌──────────────┼──────────────┐
        ▼              ▼              ▼
   Folder source   Webserver source   Git source
   #{.folder.path} #{.webserver.url}  #{.git.url}  (deferred → error)
   host ✅ docker ✅  host ❌ docker ✅   host ❌ docker ❌
```

### Article 5 — Hermetic, Linux-only, no real filesystem

Each scenario's dependencies are provisioned from ephemeral Testcontainers — no
public-internet dependence in a blocking gate. `$SAURON_HOME` is pinned to a
known path (the per-scenario temp dir on the host, an in-container path under
docker) so the suite never touches the real `~/.sauron`. Tests write only to the
godog/`t.TempDir()` temp area and mutate no environment. The suite runs on Linux
(Testcontainers needs a Docker daemon); macOS binaries are built and published
but not exercised here.

### Article 6 — Arrange / Act / Assert, and reuse

Step definitions and helpers keep testify assertions and AAA structure, and
factor repeated setup into shared helpers (the `sourceFixture`, the content
loaders, `valueOf`) rather than copy-paste across steps. Pure helpers
(`collectResources`, `decodeRegistries`, `buildSpecs`, the resolver) are
unit-tested in isolation without a process, Docker, or the real filesystem.

## Chapter IV — Governance

### Article 1 — The git transport is deferred and filtered

The git remote is constrained to ssh, and its ssh fixture is deferred future
work. Until it lands, `Git(...).URL()` errors, every git scenario carries
`@git`, and the gate filters them with `~@git` so the stub never fires in CI.

### Article 2 — Tags select the runtime

`@no-sandbox` selects the host runtime (`Execute` + `Folder` only — requesting a
`Webserver` or `Git` errors, so http/git scenarios must not carry it); the
default selects the docker runtime. `@git` is filtered from the gate.

### Article 3 — The verification gate

The suite is the `gate-integration` Taskfile target: it builds a version-stamped
host binary, points `$SAURON_BIN` at it, pins `$SAURON_HOME` to a temp dir, and
runs `go test ./...` under godog **strict** mode (undefined, pending, or
ambiguous steps fail). Per the root Constitution (Chapter IV, Article 2), this
black-box suite passing on Linux is part of what makes a feature shippable.
