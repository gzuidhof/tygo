package tygo

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/fatih/structtag"
)

type groupContext struct {
	isGroupedDeclaration bool
	doc                  *ast.CommentGroup
	groupValue           string
	groupType            string
	iotaValue            int
}

// isEmitVar returns true if dec is a string var with a tygo:emit directive.
func (g *PackageGenerator) isEmitVar(dec *ast.GenDecl) bool {
	if dec.Tok != token.VAR || dec.Doc == nil {
		return false
	}

	for _, c := range dec.Doc.List {
		if strings.HasPrefix(c.Text, "//tygo:emit") {
			// we know it's VAR so asserting *ast.ValueSpec is OK.
			v, ok := dec.Specs[0].(*ast.ValueSpec).Values[0].(*ast.BasicLit)
			if !ok {
				return false
			}
			return v.Kind == token.STRING
		}
	}
	return false
}

// emitVar emits the text associated with dec, which is assumes to be a string var with a
// tygo:emit directive, as tested by isEmitVar.
func (g *PackageGenerator) emitVar(s *strings.Builder, dec *ast.GenDecl) {
	v := dec.Specs[0].(*ast.ValueSpec).Values[0].(*ast.BasicLit).Value
	if len(v) < 2 {
		return
	}
	s.WriteString(v[1:len(v)-1] + "\n")
}
func (g *PackageGenerator) writeGroupDecl(s *strings.Builder, decl *ast.GenDecl) {
	// This checks whether the declaration is a group declaration like:
	// const (
	// 	  X = 3
	//    Y = "abc"
	// )
	isGroupedDeclaration := len(decl.Specs) > 1

	if !isGroupedDeclaration && g.PreserveTypeComments() {
		g.writeCommentGroupIfNotNil(s, decl.Doc, 0)
	}

	// We need a bit of state to handle syntax like
	// const (
	//   X SomeType = iota
	//   _
	//   Y
	//   Foo string = "Foo"
	//   _
	//   AlsoFoo
	// )
	group := &groupContext{
		isGroupedDeclaration: len(decl.Specs) > 1,
		doc:                  decl.Doc,
		groupType:            "",
		groupValue:           "",
		iotaValue:            -1,
	}

	for _, spec := range decl.Specs {
		g.writeSpec(s, spec, group)
	}
}

func (g *PackageGenerator) writeSpec(s *strings.Builder, spec ast.Spec, group *groupContext) {
	// e.g. "type Foo struct {}" or "type Bar = string"
	ts, ok := spec.(*ast.TypeSpec)
	if ok && ts.Name.IsExported() {
		g.writeTypeSpec(s, ts, group)
	}

	// e.g. "const Foo = 123"
	vs, ok := spec.(*ast.ValueSpec)
	if ok {
		g.writeValueSpec(s, vs, group)
	}
}

// Writing of type specs, which are expressions like
// `type X struct { ... }`
// or
// `type Bar = string`
func (g *PackageGenerator) writeTypeSpec(
	s *strings.Builder,
	ts *ast.TypeSpec,
	group *groupContext,
) {
	if ts.Doc != nil &&
		g.PreserveTypeComments() { // The spec has its own comment, which overrules the grouped comment.
		g.writeCommentGroup(s, ts.Doc, 0)
	} else if group.isGroupedDeclaration && g.PreserveTypeComments() {
		g.writeCommentGroupIfNotNil(s, group.doc, 0)
	}

	st, isStruct := ts.Type.(*ast.StructType)
	if isStruct {
		s.WriteString("export interface ")
		s.WriteString(ts.Name.Name)
		if g.conf.Extends != "" {
			s.WriteString(" extends ")
			s.WriteString(g.conf.Extends)
		}

		if ts.TypeParams != nil {
			g.writeTypeParamsFields(s, ts.TypeParams.List)
		}

		g.writeTypeInheritanceSpec(s, st.Fields.List)
		s.WriteString(" {\n")
		g.writeStructFields(s, st.Fields.List, 0)
		s.WriteString("}")
	}

	id, isIdent := ts.Type.(*ast.Ident)
	if isIdent {
		s.WriteString("export type ")
		s.WriteString(ts.Name.Name)
		s.WriteString(" = ")
		s.WriteString(getIdent(id.Name))
		s.WriteString(";")
	}

	if !isStruct && !isIdent {
		s.WriteString("export type ")
		s.WriteString(ts.Name.Name)

		if ts.TypeParams != nil {
			g.writeTypeParamsFields(s, ts.TypeParams.List)
		}

		s.WriteString(" = ")
		g.writeType(s, ts.Type, nil, 0, true)
		s.WriteString(";")

	}

	if ts.Comment != nil && g.PreserveTypeComments() {
		g.writeSingleLineComment(s, ts.Comment)
	} else {
		s.WriteString("\n")
	}
}

// Writing of type inheritance specs, which are expressions like
// `type X struct {  }`
// `type Y struct { X `tstype:",extends"` }`
// `export interface Y extends X { }`
func (g *PackageGenerator) writeTypeInheritanceSpec(s *strings.Builder, fields []*ast.Field) {
	inheritances := make([]string, 0)
	for _, f := range fields {
		if f.Type != nil && f.Tag != nil {
			tags, err := structtag.Parse(f.Tag.Value[1 : len(f.Tag.Value)-1])
			if err != nil {
				panic(err)
			}

			tstypeTag, err := tags.Get("tstype")
			if err != nil || !tstypeTag.HasOption("extends") {
				continue
			}

			longType, valid := getInheritedType(f.Type, tstypeTag)
			if valid {
				mappedTsType, ok := g.conf.TypeMappings[longType]
				if ok {
					inheritances = append(inheritances, mappedTsType)
				} else {
					// We can't use the fallback type because TypeScript doesn't allow extending "any".
					inheritances = append(inheritances, longType)
				}

			}
		}
	}
	if len(inheritances) > 0 {
		s.WriteString(" extends ")
		s.WriteString(strings.Join(inheritances, ", "))
	}
}

// Writing of value specs, which are exported const expressions like
// const SomeValue = 3
func (g *PackageGenerator) writeValueSpec(
	s *strings.Builder,
	vs *ast.ValueSpec,
	group *groupContext,
) {
	for i, name := range vs.Names {
		group.iotaValue = group.iotaValue + 1
		if name.Name == "_" {
			continue
		}
		if !name.IsExported() {
			continue
		}

		if vs.Doc != nil &&
			g.PreserveTypeComments() { // The spec has its own comment, which overrules the grouped comment.
			g.writeCommentGroup(s, vs.Doc, 0)
		} else if group.isGroupedDeclaration && g.PreserveTypeComments() {
			g.writeCommentGroupIfNotNil(s, group.doc, 0)
		}

		hasExplicitValue := len(vs.Values) > i
		if hasExplicitValue {
			group.groupType = ""
		}

		s.WriteString("export const ")
		s.WriteString(name.Name)
		if vs.Type != nil {
			s.WriteString(": ")

			tempSB := &strings.Builder{}
			g.writeType(tempSB, vs.Type, nil, 0, true)
			typeString := tempSB.String()

			s.WriteString(typeString)
			group.groupType = typeString
		} else if group.groupType != "" && !hasExplicitValue {
			s.WriteString(": ")
			s.WriteString(group.groupType)
		}

		s.WriteString(" = ")

		if hasExplicitValue {
			val := vs.Values[i]
			tempSB := &strings.Builder{}
			// log.Println("const:", name.Name, reflect.TypeOf(val), val)
			g.writeType(tempSB, val, nil, 0, false)
			group.groupValue = tempSB.String()
		}

		valueString := group.groupValue
		if isProbablyIotaType(valueString) {
			valueString = replaceIotaValue(valueString, group.iotaValue)
		}
		s.WriteString(valueString)

		s.WriteByte(';')

		if g.PreserveDocComments() && vs.Comment != nil {
			g.writeSingleLineComment(s, vs.Comment)
		} else {
			s.WriteByte('\n')
		}

	}
}

func getInheritedType(f ast.Expr, tag *structtag.Tag) (name string, valid bool) {
	switch ft := f.(type) {
	case *ast.Ident:
		if ft.Obj != nil && ft.Obj.Decl != nil {
			dcl, ok := ft.Obj.Decl.(*ast.TypeSpec)
			if ok {
				_, isStruct := dcl.Type.(*ast.StructType)
				valid = isStruct && dcl.Name.IsExported()
				name = dcl.Name.Name
			}
		} else {
			// Types defined in the Go file after the parsed file in the same package
			valid = token.IsExported(ft.Name)
			name = ft.Name
		}
	case *ast.IndexExpr:
		name, valid = getInheritedType(ft.X, tag)
		if valid {
			generic := getIdent(ft.Index.(*ast.Ident).Name)
			name += fmt.Sprintf("<%s>", generic)
		}
	case *ast.IndexListExpr:
		name, valid = getInheritedType(ft.X, tag)
		if valid {
			generic := ""
			for _, index := range ft.Indices {
				generic += fmt.Sprintf("%s, ", getIdent(index.(*ast.Ident).Name))
			}
			name += fmt.Sprintf("<%s>", generic[:len(generic)-2])
		}
	case *ast.SelectorExpr:
		valid = ft.Sel.IsExported()
		name = fmt.Sprintf("%s.%s", ft.X, ft.Sel)
	case *ast.StarExpr:
		name, valid = getInheritedType(ft.X, tag)
		if valid {
			// If the type is not required, mark as optional inheritance
			if !tag.HasOption("required") {
				name = fmt.Sprintf("Partial<%s>", name)
			}
		}

	}
	return
}

func getAnonymousFieldName(f ast.Expr) (name string, valid bool) {
	switch ft := f.(type) {
	case *ast.Ident:
		name = ft.Name
		if ft.Obj != nil && ft.Obj.Decl != nil {
			dcl, ok := ft.Obj.Decl.(*ast.TypeSpec)
			if ok {
				valid = dcl.Name.IsExported()
			}
		} else {
			// Types defined in the Go file after the parsed file in the same package
			valid = token.IsExported(name)
		}
	case *ast.IndexExpr:
		return getAnonymousFieldName(ft.X)
	case *ast.IndexListExpr:
		return getAnonymousFieldName(ft.X)
	case *ast.SelectorExpr:
		valid = ft.Sel.IsExported()
		name = ft.Sel.String()
	case *ast.StarExpr:
		return getAnonymousFieldName(ft.X)
	}

	return
}
