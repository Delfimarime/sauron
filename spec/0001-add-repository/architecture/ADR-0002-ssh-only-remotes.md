# ADR-0002: Git repositories are SSH remotes only

**Status**: Accepted

**Date**: 2026-06-09

**Feature**: Add Git Repository

## Context

Git supports several transports: SSH, HTTPS, the unauthenticated `git://` protocol, and local `file://`. HTTPS typically means embedding credentials in the URL or managing tokens; `git://` is unauthenticated and unencrypted; `file://` is local-only. Supporting all of them multiplies the auth and validation surface.

We want a single, authenticated, encrypted, key-based transport, and we want to avoid storing or embedding credentials.

## Decision

Only **SSH-based git URIs** are accepted — scp-like (`git@host:org/repo.git`) and `ssh://` URLs. Any other scheme (`http(s)://`, `git://`, `file://`) is rejected (FR-014). Authentication uses an SSH private key: `--ssh-key`, or the system's regular SSH setup when the flag is omitted.

## Consequences

**Positive**

- Uniform authentication via SSH keys / agent; encrypted transport.
- No credentials are stored by Sauron or embedded in persisted URIs.
- Aligns with how teams already manage git access (deploy keys, SSH agents).

**Negative**

- Users and CI must have SSH access configured; no anonymous HTTPS clones.
- Hosts that only offer HTTPS git access are not supported.

## Revisit when

A required source is reachable only over HTTPS git. A future ADR would then add an HTTPS transport and define credential/token handling.
