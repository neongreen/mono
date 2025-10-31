## .dagger

`.dagger` provides top-level Dagger functions for running repository Go test suites inside disposable containers.

The layout follows the [official Dagger monorepo best practices](https://docs.dagger.io/reference/best-practices/monorepos/): a shared module at the repository root with narrow per-project functions and a trimmed include set so each run mounts only what it needs.

### Running tests

From the repository root:

```bash
dagger call tk-tests
dagger call dissect-tests
dagger call test
```

The module mounts only the code required for each package, reuses module and build caches, and invokes `go test` inside the official `golang:1.24.7` image.
