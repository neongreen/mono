package remote

// LoadExportState loads the export state from a file
func LoadExportState(path string) (*ExportState, error) {
	return LoadJSON[ExportState](path)
}

// SaveExportState saves the export state to a file
func SaveExportState(path string, state *ExportState) error {
	return SaveJSON(path, state)
}
