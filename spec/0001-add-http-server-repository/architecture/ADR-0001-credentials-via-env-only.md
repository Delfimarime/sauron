# ADR-0001: HTTP repository credentials are environment references only

**Status**: Accepted

**Date**: 2026-06-09

**Feature**: Add HTTP Server Repository

## Context

An HTTP repository may sit behind HTTP Basic auth, so the user needs a way to supply a username and password. Persisting raw credentials in `~/.sauron/settings.json` would require secret management — encryption at rest (e.g. AES-GCM), a key, and a place to keep that key (key file or OS keychain). That machinery is non-trivial and easy to get subtly wrong.

At this stage, **simplicity is the priority and secret management is explicitly not a concern**. We want the feature to support authenticated sources without owning any secret material.

## Decision

`--username` and `--password` accept **only** the `${env:VAR}` pattern. Raw literal values are rejected (FR-021). Only the reference (e.g. `${env:SKILLS_USER}`) is persisted; the actual secret is read from the environment at use time and is never written to disk. No encryption is implemented.

The referenced variable must be set at add time (so the validating `HEAD` request can authenticate) and at every subsequent use.

## Consequences

**Positive**

- No secret material at rest; nothing to encrypt and no key to manage.
- Simpler data model: `auth.username` / `auth.password` hold a plain reference string.
- Rotation is delegated to the environment / secret manager that exports the variables.

**Negative**

- Users must manage the environment variables themselves; non-interactive or service contexts must ensure the variables are exported.
- A missing variable fails the operation rather than prompting interactively.

## Revisit when

Secret management becomes a product requirement (e.g. credentials that cannot be sourced from the environment). A future ADR would then introduce encrypted literals and define the key-management approach.
