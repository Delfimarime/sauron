# Use Case Contract

The normative contract for sauron's business-logic layer — the `UseCase[I, P]`
shape every command entrypoint and composed step shares, and the generic
`ListUseCase[I, T]` every listing command is built on. It specializes the
[architecture contract](architecture.md), which owns code structure and
wiring.

## Use Case orchestration

Business logic is organized as **use cases** (interactors), not services. A
command's entrypoint is a `UseCase`; a use case may compose other use cases as
reusable steps. They all live in `internal/usecase`.

```go
type UseCase[I, P any] interface {
	Execute(ctx context.Context, in I) (*P, error)
}
```

- **A use case returns a result, never bytes.** `Execute` answers the question
  and returns a *presentation-agnostic* `*P` — domain objects from
  `pkg/sauron/types`, or a small struct composed of them — alongside a classified
  `*Error`. It never renders: no `Table`/`Descriptor`, no `io.Writer`, no field
  projection. How the result is displayed is the client's decision, performed by
  the command layer's own rendering code in `cmd_<name>.go` after `Execute`
  returns (see [Command structure](architecture.md#command-structure)). A use case is
  thus ignorant of presentation *entirely* — not merely of the output
  destination — which is the separation an `Out()` writer could not provide.
- **One shape, two roles.** A `UseCase` is either a command's entrypoint or a
  reusable step another use case composes — there is no separate `Action` type.
  Every use case takes an explicit `context.Context` first and a typed input `in`
  (a value type, or an empty struct for a parameterless query), and returns
  `(*P, error)`. There is no `Request` context object and no `Out()`: the
  per-invocation context is the explicit first parameter, and call-scoped
  *business* input is `in`. *View* options are **not** use-case input — they
  belong to the client — with one exception: for a **listing** use case,
  `--search`/`--sort`/`--order`/`--page`/`--limit` *are* business input, because
  they shape what the composed `Lister[T]` (see
  [Listing use cases](#listing-use-cases) below) fetches. Field **selection**
  (`--fields`) stays client-side for every use case, listing or not.
- **Composition is acyclic by discipline.** A use case may call another use case.
  Nothing structurally enforces a direction, so keeping the call graph acyclic is
  the developer's responsibility. **Boundary concerns** — the top-level telemetry
  span, error classification at the edge, and the 1:1 command↔entrypoint mapping —
  belong to the command-invoked use case; a composed use case assumes it already
  runs inside that boundary and does not re-establish it.
- **Resource-acquiring use cases may return a port.** The `(*P, error)` shape
  has one sanctioned exception: an internal use case whose role is to *open* a
  resource returns the live `pkg/` port it acquires rather than a `*P` product —
  e.g. `OpenRegistry` returns a `source.FileSystem`. Such a seam is composed only by
  other use cases, never resolved by a handler and never rendered, so wrapping the
  handle in a `*P` would be ceremony with no benefit. It remains stateless and
  takes `context` first.
- **`UseCase` is stateless.** Its collaborators — the `pkg/` ports
  (`pkg/sauron/extension`), the [`storage`](architecture.md#state-storage) stores, and the zap
  logger — are injected by uberfx; everything call-scoped arrives as `ctx`/`in`,
  so a single instance is safe to reuse across invocations. Persisted state is
  reached through the [`storage`](architecture.md#state-storage) collaborator, not the input —
  the filesystem is not part of the invocation environment.
- **Lifecycle.** Inputs are plain values constructed per call and passed to
  `Execute`; there is no context object to retain, so nothing can go stale.

### Listing use cases

A command that lists a collection — whether from a live external source (the
registry) or from persisted local state (`track.yaml`) — is built on the shared
generic `ListUseCase[I, T]` in `internal/usecase/list.go` rather than
reimplementing filter/sort/page per feature:

```go
type Lister[T any] interface {
	List(ctx context.Context, opts ...source.Option) ([]T, error)
}

type ListUseCase[I, T any] struct {
	resolve func(ctx context.Context, in I) (Lister[T], ListWindow, error)
}

func NewListUseCase[I, T any](resolve func(context.Context, I) (Lister[T], ListWindow, error)) *ListUseCase[I, T]
```

`source.Option` (`pkg/sauron/source`) is the shared query vocabulary — the same
`WithSearch`/`WithSort`/`WithOrder`/`WithOffset`/`WithLimit` functional options a
`source.FileSystem` already accepts — reused as `Lister[T]`'s query language
regardless of what resolves it. `New<Name>UseCase`'s body *is* the `resolve`
closure: the per-invocation pipeline that used to be a bespoke `Execute` body
(validate the kind, open the registry live, or nothing for local state), ending
in a small `Lister[T]` adapter (e.g. `catalogueLister`, `trackLister`) bound to
the resolved source, plus the `ListWindow` to fetch — and it returns
`NewListUseCase(resolve)` **directly**, per
[no pointless wrappers](architecture.md#coding-standards): `NewListCatalogueUseCase` returns
`*ListUseCase[ListCatalogueRequest, string]`, not a wrapping
`*ListCatalogueUseCase` around one. `ListCatalogueResponse` is correspondingly a
**type alias** (`= ListResult[string]`), not a struct — the caller already has
`Kind` before `Execute` runs, so the response has nothing to add back. A future
listing use case earns a dedicated wrapping type — with its own `Execute`
projecting `ListResult[T]` onto a richer `*P` — only when its response genuinely
needs a field the caller doesn't already have; until then, `ListUseCase[I, T]`
returning `ListResult[T]` directly satisfies `UseCase[I, P]`'s "one shape" on
its own. The adapter alone decides the *mechanism* — push the options to a
remote call, or apply them over an in-memory slice from a `storage` read — and
that decision never leaks past the adapter into `listWith`, `ListUseCase`, or
the caller.

The command layer mirrors this: a listing leaf command's builder (e.g.
`newCatalogueCommand(kind usecase.CatalogueKind, use, short, long string)`)
closes over its kind and its `listFlags`-based flag struct in the `withFlags`/
`withRunE` options it passes to the shared `newCommand` — which itself carries
no notion of "kind" or any other per-feature identity (see
[Command structure](architecture.md#command-structure)).

### Layout & naming

`internal/usecase` exposes `NewFxOptions() fx.Option`, through which use cases are
provided with their `pkg/` ports, the [`storage`](architecture.md#state-storage) stores, and the
logger. Types are named `<Name>UseCase`; each lives in `usecase_<name>.go`.

**Provided by interface, never by concrete type.** `NewFxOptions()` provides
every use case via `fx.Annotate(New<Name>UseCase, fx.As(new(UseCase[I, P])))` —
the concrete `*<Name>UseCase` (or generic `*ListUseCase[I, T]`) never enters the
container, only its `UseCase[I, P]` interface. A consumer therefore depends on
the shape it actually calls (`Execute`), never the implementing type — a
composing use case (e.g. `SetProviderUseCase` depends on
`UseCase[MigrateRequest, MigrateResponse]`, never `*MigrateUseCase`) and the
command layer's `runUseCase` alike. This is why dropping
`NewListCatalogueUseCase`'s wrapping struct in favor of returning the generic
`*ListUseCase[I, T]` directly (see [Listing use cases](#listing-use-cases)
above) landed without touching a single `cmd_*.go` file: the interface the
command layer already depended on never changed.
`OpenRegistryUseCase` is the one exception, kept a bare `fx.Provide`: its
bespoke interface (`Execute` returns a `source.FileSystem`, not a `*P`) is
outside the generic `UseCase[I, P]` shape (see the "Resource-acquiring use
cases" bullet under [Use Case orchestration](#use-case-orchestration) above).

The cobra **handler** that drives a use case — its naming and the builder/handler
split — is fixed under [Command structure](architecture.md#command-structure). It maps the flag
struct and positional arguments into the use-case input, resolves and runs the use
case from the container with `fx.Invoke` — which calls `Execute` inside the started
fx lifecycle on the run context (resolving with `fx.Populate` and then calling
`Execute` is equally acceptable) — and renders the returned `*P` result to the
command's stdout through its own rendering code in `cmd_<name>.go`.
