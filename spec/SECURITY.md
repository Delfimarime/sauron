# Security Overview

A single front door to Sauron's security posture: what is protected, how, and
where each rule is defined.

**Audience:** security analysts, plus architects.

This page owns no rules. Every rule below lives in its owning contract or
feature; this overview summarizes each one and links to its source so the linked
document remains the single point of truth.

## Secrets at rest

Credentials persist only as `${env:VAR}` environment references; a resolved secret
value is never written to any file, and the track file holds no credentials at
all. A literal secret passed where a reference is expected is a usage error.
Owner: [state data contract — No secrets at rest](contracts/state.md#write-semantics);
required by [add registry FR-003](0001-add-registry/spec.md#ubiquitous) and the
[`--password` flag rule](0001-add-registry/contracts/add-registry.md#flags).

## File permissions

The home directory (`$SAURON_HOME`, or `~/.sauron/`) is created with mode `0700`
and each state file is written with mode `0600` — owner-only access — because the
files reveal which registries and artifacts a developer uses. Owner:
[state data contract — Owner-only permissions](contracts/state.md#write-semantics).

## Write integrity

State is written to a temporary file and atomically renamed into place, and all
writes are guarded by a lockfile under the home, so concurrent runs cannot corrupt
a file and a reader never observes a half-written one. Owner:
[state data contract — Write semantics](contracts/state.md#write-semantics).

## Transport security (TLS)

The `git` and `http` transports are reached over TLS, with per-registry controls:
`--skip-tls-verify`, `--ca-cert`, `--client-cert`, and `--client-key`. Owner:
[add registry command line — Flags](0001-add-registry/contracts/add-registry.md#flags);
the applicable set per transport is fixed by the
[git](0001-add-registry/capabilities/git.md) and
[http](0001-add-registry/capabilities/http.md) capabilities.

## HTTP registry API authentication

An `http` registry server authenticates Sauron with HTTP **Basic** over TLS,
declared as the `basicAuth` security scheme; because public registries may serve
anonymously the requirement is optional (`basicAuth` or none), and a server may
reject anonymous access with `401`. Owner:
[Registry HTTP API contract — Security](contracts/registry-http-api.oas3.yaml).

## Ownership boundary

Every installed artifact lands at `sauron-<registry>-<name>`; the `sauron-` prefix
marks ownership, so Sauron only ever touches artifacts it installed and never
disturbs anything else in a provider's directories. Owner:
[Sauron README — Namespacing](README.md#namespacing).
