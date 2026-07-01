# Testing Contract

The normative contract for how sauron's Go code is tested — unit-test
discipline, coverage, and the boundary with the black-box `test/e2e` suite. It
specializes the [architecture contract](architecture.md), which owns code
structure and wiring.

## Testing

- **Test-first (TDD).** A change starts from a failing test: the unit test —
  and, for user-observable behavior, the `test/e2e` Gherkin scenario
  ([integration tests](#integration-tests)) — is written or updated to fail
  *before* the production code that satisfies it. Plans and task breakdowns
  sequence this explicitly, test before implementation
  ([AUTHORING.md](../AUTHORING.md)).
- Tests use the **Arrange / Act / Assert** structure.
- **Table-driven** tests are the encouraged default: one `Test<Subject>`
  function whose cases are parametrized, covering both successful and negative
  paths in the same table. Each case carries a comment stating its intent.
- Assertions use `testify` `assert`/`require`; collaborators are substituted
  with `testify` `mock` where applicable.
- **Mocks** live in the same package that defines the interface they implement,
  are named `MockBased<Interface>`, and reside in a file named
  `mock_based_<interface>.go`.
- **Coverage** ideal is 90%; the [verification gate](../../CONSTITUTION.md)
  enforces a project-level floor of 80%, a lower figure permitted only when
  justified.
- **Command testability.** The builder/handler split fixed under
  [Command structure](architecture.md#command-structure) exists so a command's logic is unit
  tested through its handler, without constructing a `*cobra.Command`.

## Integration tests

The black-box BDD suite lives in its **own module**, `test/e2e`
(`github.com/delfimarime/sauron/test/e2e`), under the project-layout `/test`
directory; its harness reference is [`test/e2e/HARNESS.md`](../../test/e2e/HARNESS.md)
(runtime/Source architecture, controllers, fixtures, tags, the gate). The
two facts that bind on *this* contract:

- **Module boundary & dependency isolation.** `test/e2e/go.mod` resolves the root
  with `replace github.com/delfimarime/sauron => ../..`; its test-only
  dependencies live in that module's `go.mod` and are **absent from the
  [approved-dependency table](dependencies.md#approved-dependencies)**, which governs the
  production module alone.
- **`pkg/`-only graybox.** The harness `exec`s the built binary and decodes its
  output into the public `pkg/` types — never importing `internal/` (enforced by a
  `depguard` rule, since Go's `internal/` rule does not apply across the shared
  module prefix). It is **not** bound by the Use Case or ports-and-adapters
  rules.

The gate that runs it (`gate-integration`) is defined in the
[delivery contract](delivery.md).
