package tygo

import (
	"go/ast"
	"go/constant"
	"strconv"
)

// constantTSValue renders a Go constant expression as a TypeScript value, using the
// value go/types already computed.
//
// WHY THIS EXISTS. writeValueSpec used to render a const's VALUE by calling writeType —
// the TYPE renderer — on the value expression. That works for a string or number literal
// by coincidence, because Go and TypeScript spell those the same way, and it is wrong for
// everything else: writeType's job is to name a type, so when it meets an expression it
// does not recognise it emits the configured fallback type. In a value position that
// produces
//
//	export const Size = any;
//
// which is not valid TypeScript, because `any` is a type. The failure is silent — the
// generator exits 0 and writes the file — so it survives until something type-checks the
// output, and generated files are often not type-checked.
//
// Two reported issues are the same defect reached by different expressions: a negative
// constant (#42) and iota with expressions (#41), each fixed by teaching the TYPE renderer
// one more value shape. Rather than add a third, this asks go/types, which has already
// evaluated every constant in the package — including ones declared in another package,
// which the AST alone cannot resolve at all.
//
// Returns ok=false when there is no type information (a package that does not type-check,
// or a LoadMode without NeedTypesInfo), leaving the caller to fall back to its previous
// behaviour.
func (g *PackageGenerator) constantTSValue(expr ast.Expr) (string, bool) {
	if g.pkg == nil || g.pkg.TypesInfo == nil {
		return "", false
	}
	// NEVER resolve an expression mentioning iota, for two independent reasons.
	//
	// First, go/types evaluates the SAME AST node once per ConstSpec in the group and
	// overwrites TypesInfo.Types[expr] each time, so what is recorded is the LAST
	// iteration's value — reading it gives every entry the final value.
	//
	// Second, writeValueSpec stores the rendered value as group.groupValue and later
	// entries with no explicit value inherit it, with replaceIotaValue substituting per
	// entry. That needs the un-substituted expression text; a resolved constant has no
	// iota left to substitute, so the whole group collapses.
	//
	// Both produce `Low = 2, Medium = 2, High = 2` for a plain `iota` enum. The existing
	// AST path already handles these correctly, so leave it to it.
	if mentionsIota(expr) {
		return "", false
	}
	tv, found := g.pkg.TypesInfo.Types[expr]
	if !found || tv.Value == nil {
		return "", false
	}
	switch tv.Value.Kind() {
	case constant.String:
		// Quote for TypeScript. Go and TS agree on \n, \", \\ and \uXXXX escapes, which
		// is the same assumption the literal path has always made.
		return strconv.Quote(constant.StringVal(tv.Value)), true
	case constant.Bool:
		return tv.Value.ExactString(), true
	case constant.Int:
		return tv.Value.ExactString(), true
	case constant.Float:
		// ExactString can render a rational ("1/3"), which is not a TS number literal.
		f, _ := constant.Float64Val(tv.Value)
		return strconv.FormatFloat(f, 'g', -1, 64), true
	default:
		// Complex has no TypeScript equivalent; let the caller fall back.
		return "", false
	}
}

// mentionsIota reports whether an expression references the iota identifier anywhere.
func mentionsIota(expr ast.Expr) bool {
	found := false
	ast.Inspect(expr, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok && id.Name == "iota" {
			found = true
			return false
		}
		return !found
	})
	return found
}
