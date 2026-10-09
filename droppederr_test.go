package droppederr

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

// parseFuncDecl parses src as a standalone Go file and returns its single
// top-level function declaration alongside the fset it was parsed with.
func parseFuncDecl(t *testing.T, src string) (*ast.FuncDecl, *token.FileSet) {
	t.Helper()

	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, "snippet.go", src, 0)
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}

	for _, decl := range file.Decls {
		if fd, ok := decl.(*ast.FuncDecl); ok {
			return fd, fset
		}
	}

	t.Fatal("snippet declares no function")

	return nil, nil
}

// parseExpr parses src as a standalone Go expression.
func parseExpr(t *testing.T, src string) ast.Expr {
	t.Helper()

	expr, err := parser.ParseExpr(src)
	if err != nil {
		t.Fatalf("ParseExpr() error = %v", err)
	}

	return expr
}

func TestNew(t *testing.T) {
	t.Run("wires the expected name and requires the inspector", func(t *testing.T) {
		a := New(Config{})

		if a.Name != "droppederr" {
			t.Fatalf("Name = %q, want droppederr", a.Name)
		}

		if len(a.Requires) != 1 {
			t.Fatalf("Requires = %v, want [inspect.Analyzer]", a.Requires)
		}
	})
}

func TestContainsAny(t *testing.T) {
	t.Run("a present fragment matches", func(t *testing.T) {
		if !containsAny("internal/dao/entity.go", []string{"internal/dao/"}) {
			t.Fatal("containsAny() = false, want true")
		}
	})

	t.Run("an absent fragment does not match", func(t *testing.T) {
		if containsAny("src/invoice/service.go", []string{"internal/dao/"}) {
			t.Fatal("containsAny() = true, want false")
		}
	})

	t.Run("an empty fragment list matches nothing", func(t *testing.T) {
		if containsAny("src/invoice/service.go", nil) {
			t.Fatal("containsAny() = true, want false")
		}
	})
}

func TestSplitFragments(t *testing.T) {
	t.Run("empty string yields no fragments", func(t *testing.T) {
		if got := splitFragments(""); got != nil {
			t.Fatalf("splitFragments(\"\") = %v, want nil", got)
		}
	})

	t.Run("blank and whitespace-only entries are dropped", func(t *testing.T) {
		got := splitFragments("internal/dao/, ,  , src/gen/")

		want := []string{"internal/dao/", "src/gen/"}
		if len(got) != len(want) {
			t.Fatalf("splitFragments() = %v, want %v", got, want)
		}

		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("splitFragments()[%d] = %q, want %q", i, got[i], want[i])
			}
		}
	})

	t.Run("surrounding whitespace is trimmed", func(t *testing.T) {
		got := splitFragments("  internal/dao/  ")
		if len(got) != 1 || got[0] != "internal/dao/" {
			t.Fatalf("splitFragments() = %v, want [internal/dao/]", got)
		}
	})
}

func TestIsExcluded(t *testing.T) {
	c := &checker{excludePaths: []string{"internal/dao/"}}

	t.Run("a matching file is excluded", func(t *testing.T) {
		if !c.isExcluded("internal/dao/invoice.go") {
			t.Fatal("isExcluded() = false, want true")
		}
	})

	t.Run("a non-matching file is not excluded", func(t *testing.T) {
		if c.isExcluded("src/invoice/service.go") {
			t.Fatal("isExcluded() = true, want false")
		}
	})
}

func TestErrNilCheck(t *testing.T) {
	t.Run("err != nil reports the variable name", func(t *testing.T) {
		name, ok := errNilCheck(parseExpr(t, "err != nil"))
		if !ok || name != "err" {
			t.Fatalf("errNilCheck() = %q, %v, want \"err\", true", name, ok)
		}
	})

	t.Run("err == nil is not a not-nil check", func(t *testing.T) {
		if _, ok := errNilCheck(parseExpr(t, "err == nil")); ok {
			t.Fatal("errNilCheck() reported ok for ==")
		}
	})

	t.Run("a differently-named error variable still matches", func(t *testing.T) {
		name, ok := errNilCheck(parseExpr(t, "scanErr != nil"))
		if !ok || name != "scanErr" {
			t.Fatalf("errNilCheck() = %q, %v, want \"scanErr\", true", name, ok)
		}
	})

	t.Run("an identifier without err in its name does not match", func(t *testing.T) {
		if _, ok := errNilCheck(parseExpr(t, "x != nil")); ok {
			t.Fatal("errNilCheck() reported ok for x != nil")
		}
	})

	t.Run("a non-nil comparison does not match", func(t *testing.T) {
		if _, ok := errNilCheck(parseExpr(t, "err != 0")); ok {
			t.Fatal("errNilCheck() reported ok for err != 0")
		}
	})
}

func TestIdentUsed(t *testing.T) {
	t.Run("a referenced identifier is used", func(t *testing.T) {
		fd, _ := parseFuncDecl(t, `package p
func f() {
	if err != nil {
		return err
	}
}
`)
		ifStmt := fd.Body.List[0].(*ast.IfStmt)

		if !identUsed(ifStmt.Body, "err") {
			t.Fatal("identUsed() = false, want true")
		}
	})

	t.Run("an unreferenced identifier is not used", func(t *testing.T) {
		fd, _ := parseFuncDecl(t, `package p
func f() {
	if err != nil {
		return errors.New("boom")
	}
}
`)
		ifStmt := fd.Body.List[0].(*ast.IfStmt)

		if identUsed(ifStmt.Body, "err") {
			t.Fatal("identUsed() = true, want false")
		}
	})
}

func TestFirstLine(t *testing.T) {
	t.Run("braces and blank lines are skipped", func(t *testing.T) {
		if got := firstLine("{\n\treturn err\n}"); got != "return err" {
			t.Fatalf("firstLine() = %q, want %q", got, "return err")
		}
	})

	t.Run("an empty block yields the placeholder", func(t *testing.T) {
		if got := firstLine("{\n}"); got != "<empty block>" {
			t.Fatalf("firstLine() = %q, want <empty block>", got)
		}
	})

	t.Run("a line longer than 80 characters is truncated", func(t *testing.T) {
		long := "return errors.New(\"this is a deliberately long error message padded out past eighty characters\")"

		got := firstLine("{\n\t" + long + "\n}")
		if len(got) != 83 || got[len(got)-3:] != "..." {
			t.Fatalf("firstLine() = %q (len %d), want 80 chars + \"...\"", got, len(got))
		}
	})
}
func TestRender(t *testing.T) {
	fd, fset := parseFuncDecl(t, `package p
func f() {
	_ = r.ext.QueryxContext
}
`)
	assign := fd.Body.List[0].(*ast.AssignStmt)
	sel := assign.Rhs[0].(*ast.SelectorExpr)

	t.Run("a selector chain renders back to its source text", func(t *testing.T) {
		if got := render(fset, sel.X); got != "r.ext" {
			t.Fatalf("render() = %q, want r.ext", got)
		}
	})
}
