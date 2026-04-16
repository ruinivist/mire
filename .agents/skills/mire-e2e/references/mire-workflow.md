# Mire Workflow Reference

## Command Cookbook

Use these sequences as defaults when operating mire.

### Initialize mire in repo
```bash
mire init
```

### Record one new scenario
```bash
mire record --save cli/basic-help
```

### Record nested scenario with setup fixture
```bash
mkdir -p e2e/suite
cat > e2e/suite/setup.sh <<'EOF'
export FROM_SETUP=1
EOF
chmod +x e2e/suite/setup.sh
mire record --save suite/spec
```

### Run all tests
```bash
mire test
```

### Run scoped tests
```bash
mire test suite
mire test suite/spec
```

### Rewrite all golden outputs
```bash
mire rewrite
```

### Rewrite scoped golden outputs
```bash
mire rewrite suite
```

### Safe update flow for expected output changes
```bash
mire test path/to/scenario
mire rewrite path/to/scenario
mire test path/to/scenario
```

## Fixture Layout

Use this structure under configured `mire.test_dir` (default `e2e`):

```text
e2e/
  shell.sh
  setup.sh               # optional, root-level setup
  suite/
    setup.sh             # optional, nested setup
    spec/
      in                 # replay input bytes/keystrokes
      out                # expected terminal output
```

`setup.sh` files execute in directory order from test root to scenario directory.

## Troubleshooting Matrix

| Symptom | Likely cause | Fix |
|---|---|---|
| `required command "bwrap" not found in PATH` or `required command "bash" not found in PATH` | Missing runtime dependency | Install missing dependency and rerun command. |
| `missing recorder shell "<test_dir>/shell.sh"` or marker errors mentioning rerun init | Recorder shell missing/stale | Run `mire init` to regenerate `shell.sh`. |
| `record path ... must be inside test directory ...` or `test path ... must be inside test directory ...` | Scenario path escaped configured `test_dir` | Use a path under `mire.test_dir` or update `mire.toml` `test_dir`. |
| `output differed at line N` during `mire test` | Golden `out` and replayed output differ | Decide whether behavior change is intended. If intended, run `mire rewrite` for that scope. If not intended, fix command behavior/setup. |
| Replay mismatch from volatile values (timestamps, IDs) | Non-deterministic output | Add narrow regex patterns in `mire.ignore_diffs` for those lines. |
| Config read errors for `sandbox.home`, `mounts`, or `paths` | Invalid or missing values in `mire.toml` | Ensure required keys exist, `sandbox.home` is absolute, and referenced host paths exist. |

## Practical Notes

- Prefer scoped paths (`mire test suite`, `mire rewrite suite/spec`) for quick iteration.
- Keep scenario names stable and descriptive.
- Keep setup scripts idempotent to avoid order-dependent failures.
- Use `mire rewrite` only for intentional expected-output changes.
