package schemas

import _ "embed"

// JJSchema contains the embedded jj configuration schema
//
//go:embed jj.json
var JJSchema string
