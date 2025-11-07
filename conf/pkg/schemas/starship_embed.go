package schemas

import _ "embed"

// StarshipSchema contains the embedded starship configuration schema
//
//go:embed starship.json
var StarshipSchema string
