package configschema

// JJSchemaPinned returns the pinned JJ schema (v0.34.0)
// This schema has the duplicate "hermit" enum fixed.
func JJSchemaPinned() string {
	return jjSchemaPinnedJSON
}

// JJSchemaLatest returns the latest JJ schema
// This is the version that was current when schemas were last updated.
func JJSchemaLatest() string {
	return jjSchemaLatestJSON
}
