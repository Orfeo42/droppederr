package droppederr

import (
	"go/ast"
	"go/printer"
	"go/token"
	"strings"
)

// render prints node back to source using fset, for turning an AST fragment
// into the human-readable text a finding's detail quotes.
func render(fset *token.FileSet, node ast.Node) string {
	var sb strings.Builder

	if err := printer.Fprint(&sb, fset, node); err != nil {
		return ""
	}

	return sb.String()
}

// errNilCheck reports whether cond is an `<ident> != nil` check on an
// identifier whose name looks like an error variable, returning that name.
func errNilCheck(cond ast.Expr) (string, bool) {
	bin, ok := cond.(*ast.BinaryExpr)
	if !ok || bin.Op != token.NEQ {
		return "", false
	}

	nilIdent, ok := bin.Y.(*ast.Ident)
	if !ok || nilIdent.Name != "nil" {
		return "", false
	}

	errIdent, ok := bin.X.(*ast.Ident)
	if !ok || !strings.Contains(strings.ToLower(errIdent.Name), "err") {
		return "", false
	}

	return errIdent.Name, true
}
