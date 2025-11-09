package main

// Projects provides access to individual projects in the monorepo
type Projects struct {
	Dagger *Dagger // +private
}

// Access individual projects
func (m *Dagger) Project() *Projects {
	return &Projects{Dagger: m}
}

// Tk returns the tk project
func (p *Projects) Tk() *TkProject {
	return &TkProject{Dagger: p.Dagger}
}

// Want returns the want project
func (p *Projects) Want() *WantProject {
	return &WantProject{Dagger: p.Dagger}
}

// Conf returns the conf project
func (p *Projects) Conf() *ConfProject {
	return &ConfProject{Dagger: p.Dagger}
}

// Dissect returns the dissect project
func (p *Projects) Dissect() *DissectProject {
	return &DissectProject{Dagger: p.Dagger}
}

// Ingest returns the ingest project
func (p *Projects) Ingest() *IngestProject {
	return &IngestProject{Dagger: p.Dagger}
}

// Printpdf returns the printpdf project
func (p *Projects) Printpdf() *PrintpdfProject {
	return &PrintpdfProject{Dagger: p.Dagger}
}

// Prrun returns the prrun project
func (p *Projects) Prrun() *PrrunProject {
	return &PrrunProject{Dagger: p.Dagger}
}

// ClaudeTrace returns the claude-trace project
func (p *Projects) ClaudeTrace() *ClaudeTraceProject {
	return &ClaudeTraceProject{Dagger: p.Dagger}
}

// MarkdownFormat returns the markdown-format project
func (p *Projects) MarkdownFormat() *MarkdownFormatProject {
	return &MarkdownFormatProject{Dagger: p.Dagger}
}

// JjRun returns the jj-run project
func (p *Projects) JjRun() *JjRunProject {
	return &JjRunProject{Dagger: p.Dagger}
}

// LibCli returns the lib/cli library project
func (p *Projects) LibCli() *LibCliProject {
	return &LibCliProject{Dagger: p.Dagger}
}

// LibConfigschema returns the lib/configschema library project
func (p *Projects) LibConfigschema() *LibConfigschemaProject {
	return &LibConfigschemaProject{Dagger: p.Dagger}
}

// LibGhclient returns the lib/ghclient library project
func (p *Projects) LibGhclient() *LibGhclientProject {
	return &LibGhclientProject{Dagger: p.Dagger}
}

// LibGhrelease returns the lib/ghrelease library project
func (p *Projects) LibGhrelease() *LibGhreleaseProject {
	return &LibGhreleaseProject{Dagger: p.Dagger}
}

// LibToml returns the lib/toml library project
func (p *Projects) LibToml() *LibTomlProject {
	return &LibTomlProject{Dagger: p.Dagger}
}

// LibVersion returns the lib/version library project
func (p *Projects) LibVersion() *LibVersionProject {
	return &LibVersionProject{Dagger: p.Dagger}
}

// LintersCobralint returns the linters/cobralint project
func (p *Projects) LintersCobralint() *LintersCobralintProject {
	return &LintersCobralintProject{Dagger: p.Dagger}
}

// LintersUselesswrapper returns the linters/uselesswrapper project
func (p *Projects) LintersUselesswrapper() *LintersUselesswrapperProject {
	return &LintersUselesswrapperProject{Dagger: p.Dagger}
}
