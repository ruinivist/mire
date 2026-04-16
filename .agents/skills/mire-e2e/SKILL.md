---
name: mire-e2e
description: Record, replay, and maintain CLI end-to-end tests with mire. Use when an agent needs to create new mire scenarios, run existing e2e fixtures, rewrite golden outputs after expected CLI changes, debug replay mismatches (including CI failures), or explain mire test structure (`mire.toml`, `e2e/{scenario}/in`, `e2e/{scenario}/out`, and layered `setup.sh` fixtures).
---

# Mire E2E

## Overview
Use this skill to drive mire workflows end-to-end in a repository that ships the `mire` CLI. Prefer explicit, reproducible command sequences and keep fixture changes intentional.

For command recipes and troubleshooting tables, read [references/mire-workflow.md](references/mire-workflow.md).

## Preflight
1. Confirm dependencies are present before running commands:
```bash
command -v bash
command -v bwrap
```
2. Work from the project root so relative paths resolve correctly.
3. Build or locate the `mire` binary if needed:
```bash
make build
./build/mire --help
```

## Core Workflow
1. Initialize or refresh mire scaffolding:
```bash
mire init
```
This ensures `mire.toml` exists and regenerates `<test_dir>/shell.sh`.

2. Record a scenario interactively:
```bash
mire record path/to/scenario
```
Add `--save` to skip save confirmation:
```bash
mire record --save path/to/scenario
```
Record only paths inside configured `mire.test_dir`.

3. Replay tests:
```bash
mire test
mire test path/to/subtree
```
Use scoped paths to iterate quickly.

4. Rewrite golden outputs after expected behavior changes:
```bash
mire rewrite
mire rewrite path/to/subtree
```
Run `mire test` before and after rewrite to ensure only intended changes landed.

## Scenario Structure
- Keep each scenario under `<test_dir>/<scenario>/`.
- Use fixture files:
  - `in`: recorded keystrokes/input for replay
  - `out`: expected terminal output golden
- Keep scenario names stable and descriptive (`command/flag-case`, `error/invalid-input`, `nested/path`).

## Setup Fixtures
- Use `setup.sh` to prepare state inside sandbox before record/replay.
- Place `setup.sh` at test root and/or nested directories.
- Expect layered execution from test root down to scenario directory.
- Keep setup scripts idempotent and deterministic.

## Configuration Notes
- Read and update `mire.toml` as needed:
  - `mire.test_dir`: scenario root (default `e2e`)
  - `mire.ignore_diffs`: regexes for lines that may vary
  - `sandbox.home`, `sandbox.mounts`, `sandbox.paths`: sandbox host exposure
- Keep `sandbox.home` absolute and mount/path entries valid on host.

## Agent Patterns
1. Create a new test safely:
```bash
mire init
mire record --save some/scenario
mire test some
```
2. Update expected output safely after intentional UX/style changes:
```bash
mire test some/scenario
mire rewrite some/scenario
mire test some/scenario
```
3. Investigate failures:
- Capture the first mismatch line from `mire test`.
- Verify whether output change is intentional.
- Rewrite only if expected; otherwise adjust setup/commands or underlying CLI behavior.

## Limits
- Treat mire as Linux-focused due to `bash` + `bwrap` requirements.
- Keep test expectations deterministic; avoid time-sensitive or environment-variant output unless ignored by `mire.ignore_diffs`.
