// Package plugin exists only for the golangci-lint module plugin system.
// Consumers reach it through their own .custom-gcl.yml, which builds a
// custom-gcl binary with this package's init registration compiled in.
// The standalone binary in cmd/droppederr does not import it.
package plugin

import (
	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"

	"github.com/Orfeo42/droppederr"
)

func init() {
	register.Plugin("droppederr", New)
}

// Settings mirrors droppederr.Config for golangci-lint's settings decoding.
type Settings struct {
	ExcludePaths string `json:"exclude-paths"`
}

// New decodes settings and builds the droppederr linter plugin.
func New(settings any) (register.LinterPlugin, error) {
	s, err := register.DecodeSettings[Settings](settings)
	if err != nil {
		return nil, err
	}

	return &droppedErrPlugin{settings: s}, nil
}

type droppedErrPlugin struct {
	settings Settings
}

// BuildAnalyzers returns the droppederr analyzer configured from the plugin
// settings.
func (p *droppedErrPlugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{
		droppederr.New(droppederr.Config{ExcludePaths: p.settings.ExcludePaths}),
	}, nil
}

// GetLoadMode reports that droppederr reads syntax only: it matches call
// shapes and identifier names, never resolved types.
func (p *droppedErrPlugin) GetLoadMode() string {
	return register.LoadModeSyntax
}
