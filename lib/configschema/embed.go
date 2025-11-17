package configschema

import _ "embed"

// JJ schemas
//
//go:embed schemas/jj/v0.34.0.json
var jjSchemaPinnedJSON string

//go:embed schemas/jj/latest.json
var jjSchemaLatestJSON string

// Mise schema
//
//go:embed schemas/mise/latest.json
var miseSchemaJSON string

// Starship schema
//
//go:embed schemas/starship/latest.json
var starshipSchemaJSON string

// Claude schema
//
//go:embed schemas/claude/latest.json
var claudeSchemaJSON string
