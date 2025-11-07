package configschema

// PropertyInfo contains detailed information about a schema property
type PropertyInfo struct {
	Name        string
	Type        string
	Description string
	Default     any
	Enum        []string
}

// SettingInfo contains comprehensive information about a setting including current value
type SettingInfo struct {
	Path         string
	Type         string
	Description  string
	Default      any
	Enum         []string
	CurrentValue any
	IsSet        bool
}

// CompletionOption represents a completion suggestion
type CompletionOption struct {
	Name        string
	Type        string
	Description string
}
