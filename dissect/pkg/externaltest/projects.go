package externaltest

// KnownProjects contains predefined external projects for testing
var KnownProjects = map[string]ProjectConfig{
	"google/uuid": {
		Name:     "google/uuid",
		URL:      "https://github.com/google/uuid.git",
		Commit:   "2d3c2a9cc518326daf99a383f07c4d3c44317e4d",
		ShowDiff: false,
	},
	"segmentio/ksuid": {
		Name:     "segmentio/ksuid",
		URL:      "https://github.com/segmentio/ksuid.git",
		Commit:   "d33724947fcfba7949906c2b1821e96a1c8d06e7",
		ShowDiff: false,
	},
}

// GetProject returns a project configuration by name
func GetProject(name string) (ProjectConfig, bool) {
	config, ok := KnownProjects[name]
	return config, ok
}

// GetProjectNames returns all available project names
func GetProjectNames() []string {
	names := make([]string, 0, len(KnownProjects))
	for name := range KnownProjects {
		names = append(names, name)
	}
	return names
}
