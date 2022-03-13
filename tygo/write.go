package tygo

import (
	"fmt"
	"regexp"
	"strings"

	"go/ast"

	"github.com/fatih/structtag"
)

var validJSNameRegexp = regexp.MustCompile(`(?m)^[\pL_][\pL\pN_]*$`)

func validJSName(n string) bool {
	return validJSNameRegexp.MatchString(n)
}

func getIdent(s string) string {
	switch s {
	case "bool":
		return "boolean"
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64",
		"complex64", "complex128":
		return "number /* " + s + " */"
	}
	return s
}

func (g *PackageGenerator) writeIndent(s *strings.Builder, depth int) {
	for i := 0; i < depth; i++ {
		s.WriteString(g.conf.Indent)
	}
}

func (g *PackageGenerator) writeType(s *strings.Builder, t ast.Expr, depth int, optionalParens bool) {
	switch t := t.(type) {
	case *ast.StarExpr:
		if optionalParens {
			s.WriteByte('(')
		}
		g.writeType(s, t.X, depth, false)
		s.WriteString(" | undefined")
		if optionalParens {
			s.WriteByte(')')
		}
	case *ast.ArrayType:
		if v, ok := t.Elt.(*ast.Ident); ok && v.String() == "byte" {
			s.WriteString("string")
			break
		}
		g.writeType(s, t.Elt, depth, true)
		s.WriteString("[]")
	case *ast.StructType:
		s.WriteString("{\n")
		g.writeFields(s, t.Fields.List, depth+1)
		g.writeIndent(s, depth+1)
		s.WriteByte('}')
	case *ast.Ident:
		s.WriteString(getIdent(t.String()))
	case *ast.SelectorExpr:
		// e.g. `time.Time`
		longType := fmt.Sprintf("%s.%s", t.X, t.Sel)
		mappedTsType, ok := g.conf.TypeMappings[longType]
		if ok {
			s.WriteString(mappedTsType)
		} else { // For unknown types we put `any`
			s.WriteString("any")
			s.WriteString(" /* ")
			s.WriteString(longType)
			s.WriteString(" */")
		}
	case *ast.MapType:
		s.WriteString("{ [key: ")
		g.writeType(s, t.Key, depth, false)
		s.WriteString("]: ")
		g.writeType(s, t.Value, depth, false)
		s.WriteByte('}')
	case *ast.BasicLit:
		s.WriteString(t.Value)
	case *ast.ParenExpr:
		s.WriteByte('(')
		g.writeType(s, t.X, depth, false)
		s.WriteByte(')')
	case *ast.BinaryExpr:
		g.writeType(s, t.X, depth, false)
		s.WriteByte(' ')
		s.WriteString(t.Op.String())
		s.WriteByte(' ')
		g.writeType(s, t.Y, depth, false)
	case *ast.InterfaceType, *ast.CallExpr, *ast.FuncType, *ast.ChanType:
		s.WriteString("any")

	default:
		err := fmt.Errorf("unhandled: %s\n %T", t, t)
		fmt.Println(err)
		panic(err)
	}
}

func (g *PackageGenerator) writeFields(s *strings.Builder, fields []*ast.Field, depth int) {
	for _, f := range fields {
		optional := false
		required := false

		var fieldName string
		if len(f.Names) != 0 && f.Names[0] != nil && len(f.Names[0].Name) != 0 {
			fieldName = f.Names[0].Name
		}
		if len(fieldName) == 0 || 'A' > fieldName[0] || fieldName[0] > 'Z' {
			continue
		}

		var name string
		var tstype string
		if f.Tag != nil {
			tags, err := structtag.Parse(f.Tag.Value[1 : len(f.Tag.Value)-1])
			if err != nil {
				panic(err)
			}

			jsonTag, err := tags.Get("json")
			if err == nil {
				name = jsonTag.Name
				if name == "-" {
					continue
				}

				optional = jsonTag.HasOption("omitempty")
			}
			tstypeTag, err := tags.Get("tstype")
			if err == nil {
				tstype = tstypeTag.Name
				if tstype == "-" {
					continue
				}
				required = tstypeTag.HasOption("required")
			}

		}

		if len(name) == 0 {
			name = fieldName
		}

		g.writeCommentGroupIfNotNil(s, f.Doc, depth+1)

		g.writeIndent(s, depth+1)
		quoted := !validJSName(name)
		if quoted {
			s.WriteByte('\'')
		}
		s.WriteString(name)
		if quoted {
			s.WriteByte('\'')
		}

		switch t := f.Type.(type) {
		case *ast.StarExpr:
			optional = !required
			f.Type = t.X
		}

		if optional {
			s.WriteByte('?')
		}

		s.WriteString(": ")

		if tstype == "" {
			g.writeType(s, f.Type, depth, false)
		} else {
			s.WriteString(tstype)
		}
		s.WriteByte(';')

		if f.Comment != nil {
			// Line comment is present, that means a comment after the field.
			s.WriteString(" // ")
			s.WriteString(f.Comment.Text())
		} else {
			s.WriteByte('\n')
		}

	}
}
