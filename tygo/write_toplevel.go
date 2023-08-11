package tygo

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/fatih/structtag"
)

type groupContext struct {
	isGroupedDeclaration bool
	doc                  *ast.CommentGroup
	groupValue           string
	groupType            string
	iotaValue            int
	iotaOffset           int
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
		s.WriteString(" = ")
		g.writeType(s, ts.Type, 0, true)
		s.WriteString(";")

	}

	if ts.Comment != nil && g.PreserveTypeComments() {
		s.WriteString(" // " + ts.Comment.Text())
	} else {
		s.WriteString("\n")
	}
}

// Writing of type inheritance specs, which are expressions like
// `type X struct {  }`
// `type Y struct { X `tstype:",inline"` }`
// `export interface Y extends X { }`
func (g *PackageGenerator) writeTypeInheritanceSpec(s *strings.Builder, fields []*ast.Field) {
	inheritances := make([]string, 0)
	for _, f := range fields {
		if f.Type != nil && f.Tag != nil {
			tags, err := structtag.Parse(f.Tag.Value[1 : len(f.Tag.Value)-1])
			if err != nil {
				panic(err)
			}

			if !isInherited(tags) {
				continue
			}

			name, valid := getInheritedType(f.Type)
			if valid {
				inheritances = append(inheritances, name)
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
			g.writeType(tempSB, vs.Type, 0, true)
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
			g.writeType(tempSB, val, 0, true)
			valueString := tempSB.String()

			if isProbablyIotaType(valueString) {
				group.iotaOffset = basicIotaOffsetValueParse(valueString)
				group.groupValue = "iota"
				valueString = fmt.Sprint(group.iotaValue + group.iotaOffset)
			} else {
				group.groupValue = valueString
			}
			s.WriteString(valueString)

		} else { // We must use the previous value or +1 in case of iota
			valueString := group.groupValue
			if group.groupValue == "iota" {
				valueString = fmt.Sprint(group.iotaValue + group.iotaOffset)
			}
			s.WriteString(valueString)
		}

		s.WriteByte(';')
		if vs.Comment != nil && g.PreserveDocComments() {
			s.WriteString(" // " + vs.Comment.Text())
		} else {
			s.WriteByte('\n')
		}

	}
}

func getInheritedType(f ast.Expr) (name string, valid bool) {
	switch ft := f.(type) {
	case *ast.Ident:
		if ft.Obj != nil && ft.Obj.Decl != nil {
			dcl, ok := ft.Obj.Decl.(*ast.TypeSpec)
			if ok {
				_, isStruct := dcl.Type.(*ast.StructType)
				valid = isStruct && dcl.Name.IsExported()
				name = dcl.Name.Name
				break
			}
		}
	case *ast.IndexExpr:
		name, valid = getInheritedType(ft.X)
		if valid {
			generic := getIdent(ft.Index.(*ast.Ident).Name)
			name += fmt.Sprintf("<%s>", generic)
			break
		}
	case *ast.IndexListExpr:
		name, valid = getInheritedType(ft.X)
		if valid {
			generic := ""
			for _, index := range ft.Indices {
				generic += fmt.Sprintf("%s, ", getIdent(index.(*ast.Ident).Name))
			}
			name += fmt.Sprintf("<%s>", generic[:len(generic)-2])
			break
		}
	case *ast.SelectorExpr:
		valid = ft.Sel.IsExported()
		name = fmt.Sprintf("%s.%s", ft.X, ft.Sel)

	}
	return
}

func isInherited(tags *structtag.Tags) bool {
	tstypeTag, err := tags.Get("tstype")
	if err == nil && tstypeTag.HasOption("inline") {
		return true
	}
	jsonTag, err := tags.Get("json")
	if err == nil && jsonTag.HasOption("inline") {
		return true
	}
	yamlTag, err := tags.Get("yaml")
	if err == nil && yamlTag.HasOption("inline") {
		return true
	}
	return false
}
