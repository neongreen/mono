package schemas

import _ "embed"

// MiseSchemaData contains the embedded mise configuration schema
//
//go:embed mise.schema
var MiseSchemaData string
