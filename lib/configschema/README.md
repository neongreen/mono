# configschema

Configuration schema registry for various tools.

## Overview

This package provides embedded configuration schemas for tools like jj, mise, starship, and claude. Schemas are versioned and committed to the repository, avoiding network calls at runtime or during tests.

## Schema Organization

Schemas are organized by tool and version:
- `schemas/jj/v0.34.0.json` - Pinned jj schema for stable tests
- `schemas/jj/latest.json` - Latest jj schema (updated periodically)
- `schemas/mise/latest.json` - Latest mise schema
- `schemas/starship/latest.json` - Latest starship schema
- `schemas/claude/latest.json` - Latest claude schema

## Updating Schemas

To update schemas to their latest versions:

```bash
mise run //:schemas:update
```

This downloads the latest schemas from official sources and applies any necessary fixes.

## Usage

```go
import "github.com/neongreen/mono/lib/configschema"

// Get JJ schema (pinned v0.34.0)
schema := configschema.JJSchemaPinned()

// Get latest JJ schema
schema := configschema.JJSchemaLatest()

// Get Mise schema
schema := configschema.MiseSchema()
```

## Schema Sources

- **JJ**: https://jj-vcs.github.io/jj/
- **Mise**: https://mise.jdx.dev/
- **Starship**: https://starship.rs/
- **Claude**: https://json.schemastore.org/

