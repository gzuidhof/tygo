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

type enumGroup struct {
	typeName   string
	typePrefix string
	constants  []*ast.ValueSpec
	doc        *ast.CommentGroup
}

// detectEnumGroup analyzes a const declaration group to determine if it represents
// an enum-like pattern that should be converted to a TypeScript enum.
// Returns the enum group info if detected, nil otherwise.
func (g *PackageGenerator) detectEnumGroup(decl *ast.GenDecl) *enumGroup {
	// Only process const declarations with multiple specs
	if decl.Tok != token.CONST || len(decl.Specs) < 2 {
		return nil
	}

	// Only generate enums/unions if configured to do so
	if g.conf.EnumStyle != "enum" && g.conf.EnumStyle != "union" {
		return nil
	}

	var candidates []*ast.ValueSpec
	var commonType string
	var commonPrefix string

	// First pass: collect all exported constants and analyze their types
	for _, spec := range decl.Specs {
		vs, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}

		// Skip unexported constants and blank identifiers
		for _, name := range vs.Names {
			if !name.IsExported() || name.Name == "_" {
				continue
			}

			// Get the type name
			var typeName string
			if vs.Type != nil {
				if ident, ok := vs.Type.(*ast.Ident); ok {
					typeName = ident.Name
				}
			}

			candidates = append(candidates, vs)

			// For the first candidate, try to establish the pattern
			if commonType == "" && typeName != "" {
				commonType = typeName
				// Extract potential prefix from the constant name
				if strings.HasPrefix(name.Name, typeName) {
					commonPrefix = typeName
				}
			}

			break // Only process the first name in multi-name declarations
		}
	}

	// Need at least 2 candidates
	if len(candidates) < 2 {
		return nil
	}

	// If we couldn't establish a common type from explicit types, try to infer from names
	if commonType == "" || commonPrefix == "" {
		commonPrefix = findCommonPrefix(candidates)
		if commonPrefix == "" {
			return nil
		}
		commonType = commonPrefix
	}

	// Second pass: validate all candidates match the pattern
	var validConstants []*ast.ValueSpec
	for _, vs := range candidates {
		name := vs.Names[0].Name

		// Check if name matches the prefix pattern
		if !strings.HasPrefix(name, commonPrefix) {
			continue
		}

		// Check type consistency
		var typeName string
		if vs.Type != nil {
			if ident, ok := vs.Type.(*ast.Ident); ok {
				typeName = ident.Name
			}
		}

		// If there's an explicit type, it must match
		if typeName != "" && typeName != commonType {
			continue
		}

		validConstants = append(validConstants, vs)
	}

	// Need at least 2 valid constants to form an enum
	if len(validConstants) < 2 {
		return nil
	}

	return &enumGroup{
		typeName:   commonType,
		typePrefix: commonPrefix,
		constants:  validConstants,
		doc:        decl.Doc,
	}
}

// findCommonPrefix finds the longest common prefix among constant names
// that could represent an enum type name
func findCommonPrefix(constants []*ast.ValueSpec) string {
	if len(constants) == 0 {
		return ""
	}

	firstLabel := constants[0].Names[0].Name

	// Try different prefix lengths, starting with the full name and working backwards
	for prefixLen := len(firstLabel); prefixLen > 0; prefixLen-- {
		candidate := firstLabel[:prefixLen]

		// Skip if the candidate doesn't look like a type name (should be capitalized)
		if len(candidate) == 0 || candidate[0] < 'A' || candidate[0] > 'Z' {
			continue
		}

		// Check if all constants start with this candidate
		allMatch := true
		for _, vs := range constants[1:] {
			if !strings.HasPrefix(vs.Names[0].Name, candidate) {
				allMatch = false
				break
			}
		}

		if allMatch {
			return candidate
		}
	}

	return ""
}

// writeTypeScriptEnum generates a TypeScript enum declaration from an enumGroup
func (g *PackageGenerator) writeTypeScriptEnum(s *strings.Builder, enumGroup *enumGroup) {
	// Write enum comment if present
	if enumGroup.doc != nil && g.PreserveTypeComments() {
		g.writeCommentGroup(s, enumGroup.doc, 0)
	}

	// Write enum declaration
	s.WriteString("export enum ")
	s.WriteString(enumGroup.typeName)
	s.WriteString(" {\n")

	// Write enum members
	memberIndex := 0
	iotaValue := 0
	for _, constant := range enumGroup.constants {
		// Skip unexported constants
		if !constant.Names[0].IsExported() {
			continue
		}

		// Write member comment if present
		if constant.Doc != nil && g.PreserveTypeComments() {
			g.writeCommentGroup(s, constant.Doc, 1)
		}

		// Write the enum member
		s.WriteString(g.conf.Indent)

		// Generate the member name by stripping the prefix
		memberName := strings.TrimPrefix(constant.Names[0].Name, enumGroup.typePrefix)
		s.WriteString(memberName)

		// Write the value if present
		if len(constant.Values) > 0 {
			s.WriteString(" = ")
			tempSB := &strings.Builder{}
			g.writeType(tempSB, constant.Values[0], nil, 0, false)
			valueString := tempSB.String()

			// Handle iota values
			if isProbablyIotaType(valueString) {
				valueString = replaceIotaValue(valueString, iotaValue)
			}
			s.WriteString(valueString)
		}

		s.WriteString(",")

		// Write line comment if present
		if constant.Comment != nil && g.PreserveDocComments() {
			g.writeSingleLineComment(s, constant.Comment)
		} else {
			s.WriteString("\n")
		}

		memberIndex++
		iotaValue++
	}

	s.WriteString("}\n")
}

// writeTypeScriptUnion generates a TypeScript union type declaration from an enumGroup
func (g *PackageGenerator) writeTypeScriptUnion(s *strings.Builder, enumGroup *enumGroup) {
	// First write each constant declaration
	iotaValue := 0
	var lastRawValueString string
	var isIotaSequence bool
	var constNames []string

	for _, constant := range enumGroup.constants {
		// Skip unexported constants
		if !constant.Names[0].IsExported() {
			iotaValue++
			continue
		}

		constNames = append(constNames, constant.Names[0].Name)

		// Write constant comment if present
		if constant.Doc != nil && g.PreserveTypeComments() {
			g.writeCommentGroup(s, constant.Doc, 0)
		}

		// Write constant declaration without type annotation
		s.WriteString("export const ")
		s.WriteString(constant.Names[0].Name)
		s.WriteString(" = ")

		var valueString string
		// Get the value if present
		if len(constant.Values) > 0 {
			tempSB := &strings.Builder{}
			g.writeType(tempSB, constant.Values[0], nil, 0, false)
			rawValueString := tempSB.String()

			// Check if this starts an iota sequence
			if isProbablyIotaType(rawValueString) {
				isIotaSequence = true
				valueString = replaceIotaValue(rawValueString, iotaValue)
			} else {
				isIotaSequence = false
				valueString = rawValueString
			}
			lastRawValueString = rawValueString
		} else if lastRawValueString != "" {
			// If no explicit value but we have a pattern, continue based on sequence type
			if isIotaSequence {
				// Continue the iota sequence
				valueString = replaceIotaValue(lastRawValueString, iotaValue)
			} else {
				// For non-iota patterns, reuse the last value
				valueString = lastRawValueString
			}
		}

		s.WriteString(valueString)
		s.WriteString(";")

		// Write line comment if present
		if constant.Comment != nil && g.PreserveDocComments() {
			g.writeSingleLineComment(s, constant.Comment)
		} else {
			s.WriteString("\n")
		}

		iotaValue++
	}

	// Write union type comment if present
	if enumGroup.doc != nil && g.PreserveTypeComments() {
		g.writeCommentGroup(s, enumGroup.doc, 0)
	}

	// Write union type declaration using typeof references
	s.WriteString("export type ")
	s.WriteString(enumGroup.typeName)
	s.WriteString(" = ")

	// Write the union of typeof references
	for i, name := range constNames {
		if i > 0 {
			s.WriteString(" | ")
		}
		s.WriteString("typeof ")
		s.WriteString(name)
	}

	s.WriteString(";\n")
}

func (g *PackageGenerator) writeGroupDecl(s *strings.Builder, decl *ast.GenDecl) {
	// This checks whether the declaration is a group declaration like:
	// const (
	// 	  X = 3
	//    Y = "abc"
	// )
	isGroupedDeclaration := len(decl.Specs) > 1

	// Check if decl is exported, if not, we exit early so we don't write its comment.
	if !isGroupedDeclaration {
		if ts, ok := decl.Specs[0].(*ast.TypeSpec); ok && !ts.Name.IsExported() {
			return
		}
		if vs, ok := decl.Specs[0].(*ast.ValueSpec); ok && !vs.Names[0].IsExported() {
			return
		}
	}

	// Check if this is an enum group and handle it specially
	enumGroup := g.detectEnumGroup(decl)
	if enumGroup != nil {
		switch g.conf.EnumStyle {
		case "enum":
			g.writeTypeScriptEnum(s, enumGroup)
		case "union":
			g.writeTypeScriptUnion(s, enumGroup)
		}
	}

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

	// If we generated an enum, track which constants were included so we can skip them
	enumConstants := make(map[*ast.ValueSpec]bool)
	if enumGroup != nil {
		for _, vs := range enumGroup.constants {
			enumConstants[vs] = true
		}
	}

	for _, spec := range decl.Specs {
		// Skip constants that were already processed as part of an enum
		if vs, ok := spec.(*ast.ValueSpec); ok && enumConstants[vs] {
			continue
		}
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
	// Skip types that have been generated as enums to avoid duplicates
	if g.generatedEnums[ts.Name.Name] {
		return
	}

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
