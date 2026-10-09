// Package droppederr reports `if err != nil { ... }` blocks whose body never
// references err, so the root cause is discarded before any boundary can log
// it. Every finding is always reported; there is no baseline.
package droppederr

import (
	"go/ast"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const classDropped = "DROPPED"

// Config holds the flags New wires onto the analyzer it builds.
type Config struct {
	ExcludePaths string
}

// New builds a droppederr analyzer configured with cfg, whose fields are also
// exposed as flags so a driver like singlechecker can still override them
// from the command line.
func New(cfg Config) *analysis.Analyzer {
	a := &analysis.Analyzer{
		Name:     "droppederr",
		Doc:      "reports error-handling sites where the root cause never reaches the logs",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}

	a.Flags.StringVar(&cfg.ExcludePaths, "exclude-paths", cfg.ExcludePaths,
		"comma-separated path fragments; a file whose slash-normalized path contains one is never reported")

	a.Run = func(pass *analysis.Pass) (any, error) {
		return run(pass, cfg)
	}

	return a
}

// Analyzer is the default droppederr analyzer.
var Analyzer = New(Config{})

type checker struct {
	pass         *analysis.Pass
	excludePaths []string
}

func run(pass *analysis.Pass, cfg Config) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	c := &checker{
		pass:         pass,
		excludePaths: splitFragments(cfg.ExcludePaths),
	}

	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(n ast.Node) {
		fd := n.(*ast.FuncDecl)
		if fd.Body == nil {
			return
		}

		filename := filepath.ToSlash(pass.Fset.Position(fd.Pos()).Filename)
		if c.isExcluded(filename) {
			return
		}

		c.scanDropped(fd)
	})

	return nil, nil
}

func (c *checker) isExcluded(filename string) bool {
	return containsAny(filename, c.excludePaths)
}

func containsAny(filename string, fragments []string) bool {
	for _, fragment := range fragments {
		if strings.Contains(filename, fragment) {
			return true
		}
	}

	return false
}

// splitFragments turns a comma-separated flag value into its trimmed,
// slash-normalized, non-empty parts.
func splitFragments(raw string) []string {
	var out []string

	for part := range strings.SplitSeq(raw, ",") {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}

		out = append(out, filepath.ToSlash(trimmed))
	}

	return out
}
