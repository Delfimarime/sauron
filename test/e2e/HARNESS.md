# Integration Test Harness (`test/e2e`)

This is the harness reference — HOW the black-box integration suite is built. Its
governing principles (black-box exercise / graybox assertions, separate module,
hermetic, red-until-it-lands, and the gate) live in the root
[Constitution](../../CONSTITUTION.md) Chapter III Articles 6–7 and Chapter IV
Article 2, and the module layout is fixed by the
[architecture contract](../../spec/contracts/architecture.md). This document
specifies the harness's own architecture and conventions.

## Module layout

`test/e2e` is its own Go module (`.../sauron/test/e2e`) that resolves the root
through `replace … => ../..`. The `depguard` rule in `.golangci.yml` bans
`.../internal` and `.../cmd` — that mechanism enforces the root's `pkg/`-only
principle (Constitution III.6), which Go's `internal/` rule cannot across the
shared module prefix. The suite runs under `go test` (no `main`): the integration
entrypoint is tagged `//go:build !unit` and the in-process unit tests
`//go:build unit`, so the two never run together.

```
  test/e2e/
    go.mod                  module .../test/e2e ; replace root => ../..
    integration_test.go     godog entrypoint (TestFeatures) ; build !unit
    runtime.go              compositionBasedRuntime  (the per-scenario proxy)
    helper_app.go           {{.App.*}} feature-load templating
    testdata/
      *.feature             one feature per behaviour
      registries/acme/…     authored content sets (skills/agents/.personas)
    internal/
      gherkin/              controllers (step definitions) + the #{} resolver
      runtime/              the wide Runtime contract + host & docker backends
    .golangci.yml           depguard: ban .../internal and .../cmd
```

## Feature naming

Headless scenarios are named `<command>.feature` (e.g. `set_registry_git.feature`).
TUI scenarios — driven through a pseudo-terminal and mirroring each interactive
feature's headless behaviour — are named `terminal_ui_<command>.feature`; the
`terminal_ui_` prefix marks the interactive surface.

## The runtime and shared state

There is **no `world.go`**. The runtime *is* the only object every controller
shares, so it owns the per-scenario state: the provisioned sources and their
addresses. `compositionBasedRuntime` is a dumb proxy forwarding to the
tag-selected backend; both backends implement the **same wide interface**.

```
   interface Runtime  ── the per-scenario shared-state owner (this is why there is no world.go)
   ├─ Execute / ReadFile / CopyTo
   ├─ Folder(alias)    → Source { Path() }
   ├─ Webserver(alias) → Source { URL()  }
   ├─ Git(alias)       → Source { URL()  }     (docker sshd git sidecar; errors on host)
   └─ Start / Stop / IsReadOnly

   interface Source ── a declared, resource-loaded exposure of provider content
   ├─ Expose(resources…)   declare what it serves (content files; webserver auth) — never a port
   ├─ Path(ctx)            local directory   (folder)
   └─ URL(ctx)             network address   (webserver / git)
```

## Capability gaps are errors

A capability a backend cannot satisfy returns an **error** from the relevant
`Source` accessor. There is no `Pod` sub-interface and no `rt.(Pod)` assertion —
a gap is an honest error, surfaced exactly where a scenario needs the address.

```
                     ┌──── host  (@no-sandbox) ────┬──── docker (default) ─────────┐
   Execute / ReadFile │  ✅                          │  ✅                          │
   CopyTo             │  ✅ (per-scenario temp dir)  │  ✅ (into "main")            │
   Folder(…).Path()   │  ✅ owns a local temp dir    │  ✅ path inside "main"       │
   Webserver(…).URL() │  ✅ in-process @127.0.0.1    │  ✅ in-process @host-gateway │
   Git(…).URL()       │  ❌ error "not on @no-sandbox"│ ✅ gitea sshd sidecar       │
                      └─────────────────────────────┴───────────────────────────────┘
                            IsReadOnly() = true            IsReadOnly() = false
```

The **http fixture is one in-process Go server** in both runtimes (see "The http
registry fixture"), so paging/search/sort are honored faithfully and the registry
is ordinary Go in the test process. The host backend still stands up no sidecars
and never writes outside the per-scenario home, so `IsReadOnly() == true` stays
honest; only Git needs the sandbox (its sshd sidecar) and so errors on the host.

## Controllers

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
   │ basic   command   state   catalogue   registry-http   git           │
   └────┬───────┬───────────┬───────┬──────────────┬──────────────────────┘
        └───────┴───────────┴───────┴──────────────┘
                              │ all share ONE handle; resolve via valueOf[T]
                              ▼
                 compositionBasedRuntime  (per-scenario PROXY; lazy start)
                              │ forwards to the backend chosen by tag
                  ┌───────────┴───────────┐
                  ▼                       ▼
            host.Runtime            docker.Runtime
         os/exec $SAURON_BIN     compose: main(/opt/bin/sauron) + sidecars
```

The http and git fixtures share one `sourceFixture` (declare + content steps);
they differ only in the source they select and their wording.

## Templating: {{…}} and #{…}

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

## Declare, then materialize

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
  WHEN  the user sets the registry from #{.webserver.default.url}
       ▼
  THEN  ReadFile settings.yaml → decode pkg/sauron/types → assert
```

## One content set, three exposures

`Given … hosts a skill/the directory/the file` builds **one provider content
set** (`skills/`, `agents/`, `.personas/`). A source capability is one exposure
of that set; the `add registry` URI is always `#{.<source>.<attr>}`. Content
authored under `testdata/` is read in-process and carried as inline bytes, so the
exposure is identical on the host folder and inside the container — no
dependence on the Docker daemon seeing a host path.

```
        "hosts the directory testdata/registries/acme"
                       │ ONE content set
                       ▼
              ┌──────────────────┐
              │  provider content │  skills/  agents/  .personas/
              └────────┬──────────┘
        ┌──────────────┼──────────────┐
        ▼              ▼              ▼
   Folder source   Webserver source   Git source
   #{.folder.path} #{.webserver.url}  #{.git.url}
   host ✅ docker ✅  host ✅ docker ✅   host ❌ docker ✅
```

## The http registry fixture

Every http scenario is served by **one in-process Go server**
(`internal/runtime/httpregistry`) that implements the Sauron HTTP Registry API
(`spec/contracts/registry-http-api.oas3.yaml`): `GET /skills` and `GET /agents`
return the `{"items":[…]}` listing the marketplace client decodes, honoring
`q`/`sort`/`limit`/`offset` faithfully so paging, search, and sort scenarios assert
real page slices and labels. Basic auth (the `requires basic auth` step) checks the
declared credential, binding a `${env:VAR}` password reference to the same secret it
exports on the binary's environment.

The server runs in the **test process** for both runtimes and binds `0.0.0.0:0` (not
`127.0.0.1`) so a container can reach it through the host gateway. `Webserver(…).URL()`
returns the runtime-appropriate address: `http://127.0.0.1:<port>` on the host runtime,
`http://host.docker.internal:<port>` under docker, where "main" carries an
`extra_hosts: host.docker.internal:host-gateway` entry (required on Linux/CI, not just
Docker Desktop). It replaces the former nginx sidecar, so the registry "server" is
ordinary Go with full per-scenario control.

## Arrange / Act / Assert and reuse

Step definitions and helpers keep testify assertions and AAA structure, and
factor repeated setup into shared helpers (the `sourceFixture`, the content
loaders, `valueOf`) rather than copy-paste across steps. Pure helpers
(`collectResources`, `decodeRegistries`, `buildSpecs`, the resolver) are
unit-tested in isolation without a process, Docker, or the real filesystem.

## Arrange by seeding (bounded exception)

A scenario may *arrange* by seeding a public state file directly — writing a
schema-valid `pkg/sauron/types` document stream into `$SAURON_HOME` through the
runtime (`CopyTo`) — instead of producing that state by running another command.
This is permitted **only** when all hold:

1. the command under test **only reads** the state it is given (e.g. `list`,
   `describe`), so seeding cannot hide a defect in the path being tested;
2. the seeded document is the **public, schema-valid form a user could author**
   by hand — never a private or internal shape;
3. producing the same state black-box would require **unrelated commands or
   transports**, turning a read test into an `add` test.

Black-box arrange through the owning command stays the default. A feature that
uses this exception **keeps at least one black-box arrange scenario**
(produce-then-read) so the write→read path is never left unexercised. The
governing principle is the root Constitution III.6.

## Uniform exercise across commands and transports

The suite drives one binary across the commands it ships. Each command is
exercised uniformly: the same `When` family runs it and the same `Then` family
inspects its stdout, exit code, and the state files it reads or writes — no
matter how the inputs were sourced. Only the `Given` fixtures differ per command
and transport. Where a command spans transports (`http`, `git`), its action and
assertions stay **transport-agnostic**: the same `When` acts and
the same `Then` inspects, however the source was reached.

## Tags select the runtime

`@no-sandbox` selects the host runtime (`Execute`, `Folder`, and the in-process
`Webserver` fixture — only `Git` errors there); the default selects the docker
runtime. The http scenarios stay on the docker runtime so the realistic graybox —
a containerized binary reaching the fixture over the host gateway — is exercised.
`@git` scenarios need the docker runtime's sshd git sidecar, so they must not carry
`@no-sandbox`; they run in the gate like any other scenario — git is first-class
(root Constitution IV.2).

## The integration gate

The suite is the `gate-integration` Taskfile target: it builds a version-stamped
host binary, points `$SAURON_BIN` at it, pins `$SAURON_HOME` to a temp dir, and
runs `go test ./...` under godog **strict** mode (undefined, pending, or
ambiguous steps fail). Gating shipping on this suite passing on Linux is fixed by
the root [Constitution](../../CONSTITUTION.md) Chapter IV Article 2.
