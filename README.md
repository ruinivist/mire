# Mire

E2E tests for CLIs.

![demo](demo.gif)

# Quickstart

## Install

- clone
- `make build`
- add `build/mire` to path

## Using

- `mire init` to create the the single config file, every entry is explicit in config.
- `mire record test/name/` - now test how you would test manually, try out commands to see if they work as expected
- `mire test` or `mire test specific/test`
- `mire rewrite` - to rewrite all golden outputs in case of a style change

**Using fixtures?**
You can write you script commands in `setup.sh` at any level, anything at that and nested level will have those run before dropping you
into record.
