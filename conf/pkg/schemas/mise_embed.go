package schemas

import _ "embed"

// MiseSchemaData contains the embedded mise configuration schema
//
//go:embed mise.toml
var MiseSchemaData string
