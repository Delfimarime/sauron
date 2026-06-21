# Implementation Plan — Describe Registry

Implementation plan for the [Describe Registry](spec.md) feature. It conforms to
the [architecture contract](../contracts/architecture.md), the
[CLI contract](../contracts/cli.md), and the
[state data contract](../contracts/state.md), and realizes the
[`describe registry` command contract](contracts/describe-registry.md). The work
is split into verifiable tasks in [TASKS.md](TASKS.md).

## 1. Goal & scope

`sauron describe registry <name>` reads `registries.yaml`, finds the one named
registry, and prints its full detail as a **`kubectl describe`-style descriptor**
— a vertical view of left-aligned field labels with their values, and indented
nested blocks for structured fields such as `auth`, per the
[CLI contract](../contracts/cli.md) and the example in the
[command contract](contracts/describe-registry.md). This is deliberately *not*
the column-aligned `Table` that [0002](../0002-list-registries/plan.md) uses for
list output: just as `kubectl describe <resource>` differs from `kubectl get`,
single-record detail is a descriptor, not a row in a table. The output is
field-selectable (`--fields`). The command is read-only: it persists nothing.
Credential fields render as their stored environment reference, never a resolved
secret (FR-002). A name that matches no registry fails with a runtime error
(exit 1, FR-004).

The feature also establishes the foundation every later **describe** feature
reuses:

- `internal/presentation` — a shared, dependency-free **descriptor renderer**
  over the standard library, producing the [CLI contract](../contracts/cli.md)
  detail rendering. The `describe artifact`
  ([0011](../0011-describe-artifact/spec.md)) and `describe provider`
  ([0013](../0013-describe-provider/spec.md)) features reuse it unchanged. It
  sits beside the `Table` renderer that
  [0002](../0002-list-registries/plan.md) introduced — `describe` is to
  `Descriptor` what `list` is to `Table`.

**Delivered (this feature):**

- The `describe registry` command, the describe use case, the shared descriptor
  renderer, and the black-box and seeded `test/e2e` scenarios.

**Out of scope — deferred to later features (YAGNI):**

- Any new store read method: the single-record read path
  `RegistriesStore.FindByName` already ships from
  [0001](../0001-add-registry/plan.md) and is reused as-is.
- Resolving credential references to secret values — Sauron never reads secrets
  at rest; the env reference is the value shown (FR-002).
- The `delete` registry verb ([0004](../0004-delete-registry/spec.md)).
- Describing artifacts ([0011](../0011-describe-artifact/spec.md)) and providers
  ([0013](../0013-describe-provider/spec.md)), which reuse the descriptor
  renderer this feature delivers.

## 2. Pre-requirements

Before executing the tasks in [TASKS.md](TASKS.md):

- **[Add Registry](../0001-add-registry/plan.md) and
  [List Registries](../0002-list-registries/plan.md) are in place** — the
  `storage` engine and typed `RegistriesStore` (including
  `FindByName(ctx, name)`), the `usecase.Error{Type, Reason}` model and single
  `cmd/main.go` error site, the cobra root with its uberfx bootstrap, the
  `internal/presentation` package (the `Table` renderer), and the `test/e2e`
  godog harness all ship.
- **No new dependency** — the descriptor renderer uses the standard library, so the
  approved-dependency list on the
  [architecture contract](../contracts/architecture.md) is untouched.
- **Toolchain** — Go `1.26`, the [Task](https://taskfile.dev) runner, and the
  existing `gate-lint` / `gate-coverage` / `gate-security` / `gate-integration`
  gates.

## 3. Component & dependency flow (as designed)

```mermaid
graph TD
  subgraph cmd["internal/cmd (cobra — thin)"]
    DR["describe.go: Describe() group · describe_registry.go: DescribeRegistry()/describeRegistry()<br/>&lt;name&gt; --fields · fx.Invoke(uc.Execute)"]
  end
  subgraph uc["internal/usecase (stateless)"]
    UC["DescribeRegistryUseCase.Execute(req)<br/>FindByName · not-found → runtime error · project(fields) · render"]
  end
  subgraph pres["internal/presentation (shared)"]
    D["Descriptor{Fields}.Render(w)<br/>kubectl-describe style: left-aligned label: value · indented nested block for auth — new<br/>Table{...} — from 0002 (list/get)"]
  end
  subgraph store["internal/.../repository/storage"]
    RS["RegistriesStore.FindByName → *types.Registry"]
    ST["Store.FindOne(kind, name) → *yaml.Node (validate-on-read)"]
  end
  TYPES["pkg/sauron/types: Registry · RegistrySpec · Auth · Transport"]

  DR -->|fx.Invoke Execute| UC
  UC -->|capability| RS
  UC -->|render| D
  RS --> ST
  ST -->|injected afero.Fs| HOME["~/.sauron/registries.yaml"]
  D -->|writer| OUT["request.Out() → stdout"]
  TYPES -.shared.-> UC
  TYPES -.shared.-> RS
```

The use case depends on the typed `RegistriesStore` (the capability) and the
`presentation.Descriptor` renderer; it owns every detail decision — locating the
record, classifying a miss, and field projection — while the renderer stays a
pure formatter that knows nothing of registries.

## 4. Runtime sequence

```text
User            cmd            UseCase           Store         Presentation
 │ describe registry acme (1)   │                  │              │
 │──────────────▶│              │                  │              │
 │               │ Execute(req) │                  │              │
 │               │─────────────▶│                  │              │
 │               │              │ FindByName(name) │              │
 │               │              │─────────────────▶│              │
 │               │              ◀─ ─ ─ ─ ─ ─ ─ ─ ─ │ *Registry|nil│
 │               │              │ not-found → runtime error (exit 1)
 │               │              │ project(fields)  │              │
 │               │              │ Render(rows)     │              │
 │               │              │────────────────────────────────▶│
 │               │              ◀─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─│ descriptor → Out()
 │               ◀─ ─ ─ ─ ─ ─ ─ │ stdout           │              │
 ◀─ ─ ─ ─ ─ ─ ─ │ exit 0        │                  │              │
```

Solid `──▶` is a synchronous call, dashed `◀─ ─` a return. The pipeline stops at
the first failing step, with the exit code shown.

- `(1)` `sauron describe registry acme --fields name,transport,uri`
- a `--fields` value outside its fixed set -> **usage (2)**
- `FindByName` read, parse, or schema-validation failure -> **io (1, "registries.yaml is unreadable")**
- no registry of that name -> **runtime error (1, "registry \"acme\" does not exist")**
- success -> writes the descriptor to stdout, **exit 0**

## 5. Interfaces (as designed)

```go
// internal/presentation — the shared, registry-agnostic descriptor renderer
// (beside the existing Table). A `kubectl describe`-style vertical view, not a
// column-aligned table. A pure value type; no fx wiring.
type Descriptor struct {
    Fields []Field
}
// Field is a label with either a leaf Value or a nested block of sub-Fields
// (never both meaningfully): a leaf renders "label: value"; a section renders
// "label:" followed by its Children indented one level (e.g. auth).
type Field struct {
    Label    string
    Value    string  // leaf value; rendered "label: value"
    Children []Field // nested section, indented under "label:" (e.g. auth)
}
func (d Descriptor) Render(w io.Writer) error // left-aligned labels; nested blocks indented

// internal/.../repository/storage — REUSED unchanged from 0001; no new method.
type RegistriesStore interface {
    FindByName(ctx context.Context, name string) (*types.Registry, error) // 0001
    Add(ctx context.Context, r types.Registry) error                      // 0001
    List(ctx context.Context) ([]types.Registry, error)                   // 0002
}

// internal/usecase
type DescribeRegistryUseCase struct{ /* registries, logger */ }
func (uc *DescribeRegistryUseCase) Execute(request *DescribeRegistryRequest) error

type DescribeRegistryRequest struct {
    context.Context
    Name   string   // the registry to describe (required positional arg)
    Fields []string // validated against {name,transport,uri,ref,auth,tls,sshKey,timeout}; name forced first
    // Out() io.Writer — the command's output writer
}
```

## 6. Delivered file layout

### `internal/`
| Path | Holds |
|---|---|
| `presentation/{descriptor.go, descriptor_test.go}` | the shared `Descriptor` renderer (kubectl-describe style) over the standard library — a pure value type, so no fx wiring; sits beside the existing `table.go` |
| `usecase/{usecase_describe_registry.go, fx.go}` (+ test) | `DescribeRegistryUseCase` and `DescribeRegistryRequest`; provided through `NewFxOptions` |
| `cmd/{describe.go, describe_registry.go, root.go}` (+ tests) | the `Describe()` group, the `DescribeRegistry()` builder and `describeRegistry()` handler, the `--fields` flag, and `root.AddCommand(Describe())` |

### Specification & governance
| Path | State |
|---|---|
| `spec.md`, `data/state.md`, `contracts/describe-registry.md` | unchanged unless T1 finds genuine drift; the detail-rendering decisions are recorded here |

## 7. Checkpoints

Ordered, verifiable milestones — each met when its single command or criterion
passes (these back the tasks in [TASKS.md](TASKS.md)):

| Milestone | Verify |
|---|---|
| Spec/contract drift (if any) reconciled | the not-found error class is resolved in §8; `data/state.md` field table matches the command contract's `--fields` set (inspection) |
| e2e suite authored | `task gate-integration` resolves every step, failing only on the not-yet-built command |
| Shared descriptor renderer | `go test ./internal/presentation/...` |
| Describe use case | `go test ./internal/usecase/...` |
| cmd surface (the e2e suite turns green) | `go test ./internal/cmd/...` |
| Lint / format / coverage / security | `task gate-lint && task gate-coverage && task gate-security` |
| e2e scenarios | `task build && task gate-integration` |
| Full gate | `task all` |

## 8. Key decisions

1. **Reuse the single-record read path; add no store method.** The use case
   reads through the existing `RegistriesStore.FindByName(ctx, name)` from
   [0001](../0001-add-registry/plan.md). Unlike
   [0002](../0002-list-registries/plan.md), which added `List`/`FindAll`, this
   feature needs no new storage surface.
2. **Shared `Descriptor` renderer in `internal/presentation` — a `kubectl
   describe`-style view, not a table.** A pure formatter over the standard
   library, producing the [CLI contract](../contracts/cli.md) detail rendering —
   left-aligned field labels with their values and indented nested blocks for
   structured fields like `auth`. It is a distinct renderer from the
   column-aligned `Table` ([0002](../0002-list-registries/plan.md) owns that for
   list/`get` output): single-record detail is a descriptor, the way `kubectl
   describe` differs from `kubectl get`. It owns alignment and nesting; it owns
   no registry semantics, so the later describe features reuse it unchanged. No
   third-party dependency is introduced.
3. **Field selection lives in the use case**, never the renderer. `--fields`
   selects and reorders from `{name, transport, uri, ref, auth, tls, sshKey,
   timeout}` with `name` always present and first; a value outside the set is
   `usage` (exit 2). The default (no `--fields`) shows every populated field.
4. **Secrets are a pure pass-through (FR-002).** `spec.auth.*` is rendered as the
   stored environment reference. Sauron never holds a resolved secret at rest, so
   "never display a secret" is satisfied by printing the stored value verbatim;
   no redaction step is required.
5. **Error model is reused.** A `--fields` value outside its fixed set is `usage`
   (exit 2); an unreadable or invalid state file is `io` (exit 1). Classification
   stays in the use case; `cmd/main.go` remains the single error site.
6. **Not-found is its own error class — `TypeNotFound`, exit 1 (resolved).** "No
   registry of that name exists" (FR-004) maps to exit 1, distinct from a `usage`
   error (exit 2). None of the existing `usecase.Error` types
   (`usage`/`conflict`/`unreachable`/`validation`/`io`) names a *not-found*, so T4
   adds a `TypeNotFound` constant to the error model and `cmd/main.go` maps it to
   exit 1 alongside the other non-usage types (the single error site is unchanged
   in shape — only one mapping arm is added). The type is reused by the later
   `describe`/`get`-style features rather than re-derived per feature.

## 9. Tasks

The work is split into independently **verifiable** tasks in
[TASKS.md](TASKS.md), authored **TDD-first**: the e2e suite is written before the
product and stays red until the command lands. Dependency order:

`T1 spec recon → T2 e2e (red)`; `T3 descriptor renderer` runs alongside; then
`→ T4 use case → T5 cmd` (which turns the e2e suite green) `→ T6 full gate`.
