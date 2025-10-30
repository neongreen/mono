# Dagger Module for tk

This directory contains a Dagger module for running tk CI/CD tasks.

## Prerequisites

- [Dagger CLI](https://docs.dagger.io/install) installed
- Docker or compatible container runtime

## Available Functions

### Test
Run tk tests with coverage:
```bash
cd tk/.dagger
dagger call test
```

### Build
Build the tk binary:
```bash
cd tk/.dagger
dagger call build export --path=./tk
```

### Format Check
Check Go code formatting:
```bash
cd tk/.dagger
dagger call fmt-check
```

## CI Integration

The Dagger module is integrated with GitHub Actions in `.github/workflows/tk-dagger.yml`.

## Local Development

To run tests locally with Dagger:

```bash
cd tk/.dagger
dagger call test
```

This ensures your local tests run in the same environment as CI.
