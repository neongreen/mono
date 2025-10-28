package schemas

import _ "embed"

// MiseJSONSchema contains the embedded Mise configuration JSON schema
//
//go:embed mise.json
var MiseJSONSchema string
