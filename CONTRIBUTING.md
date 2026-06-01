# Contributing

## Maintenance

When changing syntax, standard-library behaviour, runtime semantics, manifest files, bundling or repository layouts, please take some time to update the relevant documentation, example code and test coverage.

At the very least, check these files for potential updates:

- `README.md`,
- `docs/` *and*
- `examples/`.

Give special attention to `docs/architecture.md` when changing the compiler pipeline, bytecode virtual machine, interpreter, manifest execution, bundling or test strategy.

Also, run `go test ./...` to ensure that no breaking changes have been introduced.