# Tasks — List Catalogue

The executable breakdown of [plan.md](plan.md). Each task is **independently
verifiable**: it owns a set of files and states the single command or criterion
whose success confirms it. The suite is authored **TDD-first** — the e2e tests
(T2) are written before the product (T3–T6) and stay red until the command lands
(T6), per the [integration constitution](../../test/e2e/CONSTITUTION.md)
Chapter I, Article 3.

> Authoring rule (see [AUTHORING.md](../AUTHORING.md)): every task carries a
> verification — a task without a pass/fail check is not a task.

## Dependency order

- T1 → T2  (the e2e scenarios encode the reconciled decisions)
- T1 → T3  (the `source.Options.Order` addition is independent; it may be worktree-isolated)
- T1 → T4  (the `OpenRegistryAction` is independent; it may be worktree-isolated)
- T3, T4 → T5 → T6  (T6 turns the e2e suite from red to green)
- T6 → T7

T3 and T4 are fully independent — T3 touches only `pkg/sauron/source` and the
registry adapters; T4 touches only `internal/usecase` (incl. the `add_registry`
refactor).

## Tasks

### T1 — Specification & contract reconciliation (+ Ask-first approvals)
- **Delivers:** confirmation that the [spec](spec.md), the
  [state](data/state.md) field→requirement table, and the three per-kind command
  contracts ([skill](contracts/list-catalogue-skill.md),
  [agent](contracts/list-catalogue-agent.md),
  [persona](contracts/list-catalogue-persona.md)) agree on the flag set
  (`--search`, `--sort name`, `--order asc|desc`, **`--page` (default 1)**,
  **`--limit` (default 20)**), the table columns (`NAME`/`KIND` for skill & agent,
  `NAME`/`MEMBERS` for persona), and the paging-line
  wording — `showing <from>–<to> (page <p>, limit <l>)`, and
  `showing 0 results (page <p>, limit <l>)` for an empty page (**no `of N`
  total**). It moves the corpus from `--offset` to `--page` (offset computed on the
  client), records the **on-source catalogue layout** — `.skills`/`.agents`/
  `.personas`, each holding `<name>.(yaml|yml)` manifests, the catalogue name being
  the filename with its extension trimmed — and resolves the **Ask-first** items
  below.
- **Ask-first items to approve before T3/T5/T6 begin (per
  [AGENTS.md](../../AGENTS.md) "Ask first"):**
  | Item | Why gated |
  |---|---|
  | [CLI contract](../contracts/cli.md): `--offset`→`--page`; paging line drops `of N`, becomes page-based | a normative-contract change rippling across paginated specs |
  | Add `Order` to `source.Options` + `WithOrder` (`pkg/sauron/source`) | a public-port addition (additive, non-breaking) |
- **Files:** `spec/0005-list-catalogue/spec.md` (FR-002 `--offset`→`--page`, record
  the layout in `## Notes`), `contracts/list-catalogue-{skill,agent,persona}.md`
  (synopsis, flag table, examples → `--page`, no `of N`); on approval,
  `spec/contracts/cli.md`; and the
  [0002 plan](../0002-list-registries/plan.md) cross-reference note.
- **Verify:** the flag/column/paging-line set in `data/state.md` matches the
  command contracts; no `--offset` or `of N` remains in any 0005 doc or in
  `cli.md`'s catalogue example; the spec `## Notes` records the catalogue layout
  (inspection). Do not invent corrections beyond that.
- **Depends on:** —

### T2 — e2e suite (authored TDD-first; red until T6)
- **Delivers:** the `list catalogue` feature file and a controller, authored so
  every step resolves and the only failure is the not-yet-built command. The
  seeded registry and all artifact fixtures — including any malformed manifest used
  by a failure scenario — are defined **explicitly in the feature doc-strings**,
  never synthesized in controller code.
- **New steps (`catalogue_controller.go`), only where an existing step does not fit:**
  | Step | Role |
  |---|---|
  | `Given the registry (.+) offers the following (skills\|agents\|personas):` | seed a filesystem registry's `.skills`/`.agents`/`.personas` from a table/doc-string of `<name>.(yaml\|yml)` manifests |
  | `Then the catalogue lists (.+)` | assert a `NAME  KIND` row is present |
  | `Then the paging line reads (.+)` | assert the exact `showing …` line |
- **Reused steps:** `Given the following registries are configured:`,
  `the user runs (.+)`, `the command succeeds`,
  `the command exits with status (\d+)`, `the output contains (.+)`.
- **Scenarios (`list_catalogue.feature`, filesystem transport):**
  | # | Requirement | Scenario |
  |---|---|---|
  | 1 | FR-001 | seed a registry with agents → `list catalogue agent <reg>` lists every agent as `NAME KIND` |
  | 2 | FR-001 | `list catalogue skill <reg>` lists `NAME KIND`; `list catalogue persona <reg>` lists `NAME MEMBERS` summarizing each persona's declared skills/agents |
  | 3 | FR-002 | `--page 2 --limit 1` → the second row, paging line `showing 2–2 (page 2, limit 1)` |
  | 4 | FR-002 | `--page 9 --limit 20` past the end → no rows, paging line `showing 0 results (page 9, limit 20)` |
  | 5 | FR-003 | `--search rev` → only entries whose name contains `rev` (case-insensitive) |
  | 6 | FR-004 | `--sort name --order desc` → entries in descending name order before paging |
  | 7 | FR-006 | unknown registry name → exit 1, reports the registry does not exist |
  | 8 | FR-005 | a registry pointing at an unreachable/absent source → exit 1 |
  | 9 | FR-007 | a missing `<registry>` arg, or `--page 0` / `--order sideways` → exit 2, command not executed |
- **Files:** `test/e2e/testdata/list_catalogue.feature`,
  `test/e2e/internal/gherkin/catalogue_controller.go` (+ its registration in
  `init.go`).
- **Verify:** `task build && task gate-integration` — green after T6; before T6,
  the suite resolves every step and fails only on the missing command (no
  undefined, pending, or ambiguous steps).
- **Depends on:** T1 (a green result also requires T6)

### T3 — `source.Options.Order` + adapters honor `--order`
- **Delivers:** the additive `Order` field on `source.Options` and the
  `WithOrder(order)` option (`List`'s signature, and therefore the generated mock,
  are unchanged — no total is added, per the Zalando count-less listing); the three
  transport adapters updated to honor `--order` (apply sort direction before
  paging) — the filesystem/git adapters in their local sort, the http adapter by
  mapping `Sort`+`Order` onto the registry HTTP API's existing signed `sort=±name`
  directive (`+name`/`-name`).
- **Files:** `pkg/sauron/source/{source.go, source_test.go}`,
  `internal/infrastructure/repository/registry/{os_filesystem.go, git_filesystem.go, rest_filesystem.go, api/…}`
  (+ tests).
- **Verify:** `go test ./pkg/sauron/... ./internal/infrastructure/...`
- **Worktree isolation:** independent of T4; if executed in parallel with T2/T4,
  run on branch `feat/0005-source-order` in its own git worktree and merge the
  branch back into the working tree without committing or pushing.
- **Depends on:** T1 (the public-port addition must be approved)
- **Note:** the [registry HTTP API](../contracts/registry-http-api.oas3.yaml)
  already carries `q`/`limit`/`offset` and a signed `sort` directive
  (`+name`/`-name`, Zalando #137), so the http adapter maps onto existing
  parameters — **no OAS change**; the filesystem/git adapters sort locally.

### T4 — Shared `OpenRegistryAction`
- **Delivers:** `OpenRegistryAction` — given a stored `types.Registry`, resolve
  `${env:VAR}` credential references, select the transport adapter (`map[Transport]
  extension.Registry`), build the `extension.Option` set from the registry spec
  (uri, ref, timeout, auth, tls, sshKey), and `Open()` the source; an open failure
  classifies as *unreachable* (FR-005). `AddRegistryUseCase` is refactored to
  compose the action for its adapter-selection/credential-resolution step, with its
  behaviour and tests unchanged. fx wiring provides the action.
- **Files:** `internal/usecase/{action_open_registry.go, api.go, fx.go}`,
  `internal/usecase/usecase_add_registry.go` (compose the action) (+ tests, and a
  `mock_based_*` for the action where a use-case test needs it).
- **Verify:** `go test ./internal/usecase/...` (add-registry suite still green)
- **Worktree isolation:** independent of T3; if executed in parallel with T2/T3,
  run on branch `feat/0005-open-registry-action` in its own git worktree and merge
  the branch back into the working tree without committing or pushing.
- **Depends on:** T1

### T5 — List-catalogue use case
- **Delivers:** `ListCatalogueUseCase` and `ListCatalogueRequest` — the
  `FindByName` → not-found (`TypeNotFound`, FR-006) → `OpenRegistryAction` →
  `offset = (page−1)·limit` → `List(rootForKind, search/sort/order/offset/limit)` →
  kind-specific projection → `Table.Render` + paging line pipeline; the kind→root
  map (`skill→.skills`, `agent→.agents`, `persona→.personas`) and the
  `<name>.(yaml|yml)` name trimming. **`skill`/`agent` project `NAME`/`KIND`** (no
  content read); **`persona` projects `NAME`/`MEMBERS`** by reading each listed
  persona manifest (`source.File.Read()`) and parsing `PersonaSpec.Members` into a
  `skills: …; agents: …` summary. The paging line is built from the
  page/limit/returned-row-count (`showing <from>–<to> …` / `showing 0 results …`);
  the `usage`/not-found/unreachable classification; the ECS-logged outcome; the fx
  wiring.
- **Files:** `internal/usecase/{usecase_list_catalogue.go, api.go, fx.go}` (+ tests).
- **Verify:** `go test ./internal/usecase/...`
- **Depends on:** T3, T4

### T6 — cmd surface
- **Delivers:** a shared **paging** flag-group (`--page` default 1 / `--limit`
  default 20, with their `bindPagingFlags`) in `helper_flags.go`; the `Catalogue()`
  group under `List()`; the three `ListCatalogue{Skill,Agent,Persona}()` builders
  sharing one `listCatalogue(kind, …)` handler (`fx.Invoke`); the `<registry>`
  positional arg (`cobra.ExactArgs(1)` via the shared usage-args wrapper); the
  `--search`/`--sort`/`--order` flags; and the `List().AddCommand(Catalogue())`
  wiring. Turns the T2 suite green. FR-007 (missing/invalid args/flags → exit 2,
  including `--page`/`--limit` below `1`) is enforced here.
- **Files:** `internal/cmd/{list.go, catalogue.go, helper_flags.go, root.go}`
  (+ tests).
- **Verify:** `go test ./internal/cmd/...`
- **Depends on:** T5

### T7 — Full verification gate
- **Delivers:** the complete gate over the feature.
- **Verify:** `task all`
- **Depends on:** T6
