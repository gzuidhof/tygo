package tygo

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"
)

const unionDirective = "//tygo:union"

// unionCandidate is a concrete named type eligible for union membership.
type unionCandidate struct {
	spec  *ast.TypeSpec
	named *types.Named
}

// annotatedInterface is an interface marked with a tygo:union directive.
type annotatedInterface struct {
	spec  *ast.TypeSpec
	iface *types.Interface
}

// hasUnionDirective reports whether comments contain an exact tygo:union directive.
func hasUnionDirective(comments *ast.CommentGroup) bool {
	if comments == nil {
		return false
	}
	for _, comment := range comments.List {
		if strings.TrimSpace(comment.Text) == unionDirective {
			return true
		}
	}
	return false
}

// typeSpecHasUnionDirective checks comments associated with one type declaration.
func typeSpecHasUnionDirective(typeSpec *ast.TypeSpec, declaration *ast.GenDecl) bool {
	return hasUnionDirective(typeSpec.Doc) || (len(declaration.Specs) == 1 && hasUnionDirective(declaration.Doc))
}

// fileHasUnionDirective reports whether a file contains an annotated type declaration.
func fileHasUnionDirective(file *ast.File) bool {
	for _, declaration := range file.Decls {
		genDecl, ok := declaration.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, declarationSpec := range genDecl.Specs {
			typeSpec, ok := declarationSpec.(*ast.TypeSpec)
			if ok && typeSpecHasUnionDirective(typeSpec, genDecl) {
				return true
			}
		}
	}
	return false
}

// analyzeInterfaceUnions discovers eligible implementations for annotated interfaces.
func (g *PackageGenerator) analyzeInterfaceUnions() error {
	g.interfaceUnions = make(map[*ast.TypeSpec][]string)
	if g.pkg == nil || g.pkg.TypesInfo == nil {
		return nil
	}

	var interfaces []annotatedInterface
	var candidates []unionCandidate
	// Collect interfaces and emitted types, preserving package traversal order.
	for _, file := range g.files {
		for _, declaration := range file.syntax.Decls {
			genDecl, ok := declaration.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, declarationSpec := range genDecl.Specs {
				typeSpec, ok := declarationSpec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				typeName, _ := g.pkg.TypesInfo.Defs[typeSpec.Name].(*types.TypeName)
				if typeSpecHasUnionDirective(typeSpec, genDecl) {
					// Validate annotated declarations before collecting their implementations.
					if _, ok := typeSpec.Type.(*ast.InterfaceType); !ok {
						return fmt.Errorf("cannot generate union for %s: //tygo:union may only be applied to a named interface", typeSpec.Name.Name)
					}
					if typeName == nil {
						return fmt.Errorf("cannot generate union for interface %s: type information is unavailable", typeSpec.Name.Name)
					}
					named, ok := typeName.Type().(*types.Named)
					if !ok {
						return fmt.Errorf("cannot generate union for interface %s: aliases are not supported", typeSpec.Name.Name)
					}
					if named.TypeParams().Len() > 0 {
						return fmt.Errorf("cannot generate union for interface %s: generic interfaces are not supported", typeSpec.Name.Name)
					}
					iface, ok := named.Underlying().(*types.Interface)
					if !ok {
						return fmt.Errorf("cannot generate union for %s: //tygo:union may only be applied to a named interface", typeSpec.Name.Name)
					}
					interfaces = append(interfaces, annotatedInterface{spec: typeSpec, iface: iface.Complete()})
				}

				// Only emitted concrete named types can become union members.
				if typeSpec.Name.IsExported() && !typeSpec.Assign.IsValid() && typeName != nil {
					named, ok := typeName.Type().(*types.Named)
					if ok {
						if _, isInterface := named.Underlying().(*types.Interface); !isInterface {
							candidates = append(candidates, unionCandidate{spec: typeSpec, named: named})
						}
					}
				}
			}
		}
	}

	// Use Go method sets to match both value and pointer receiver implementations.
	for _, union := range interfaces {
		if !union.spec.Name.IsExported() {
			continue
		}
		seen := make(map[*types.Named]bool)
		for _, candidate := range candidates {
			if !types.Implements(candidate.named, union.iface) && !types.Implements(types.NewPointer(candidate.named), union.iface) {
				continue
			}
			if candidate.named.TypeParams().Len() > 0 {
				return fmt.Errorf("cannot generate union for interface %s: generic implementing type %s is not supported", union.spec.Name.Name, candidate.spec.Name.Name)
			}
			if !seen[candidate.named] {
				g.interfaceUnions[union.spec] = append(g.interfaceUnions[union.spec], candidate.spec.Name.Name)
				seen[candidate.named] = true
			}
		}
		if len(g.interfaceUnions[union.spec]) == 0 {
			return fmt.Errorf("cannot generate union for interface %s: no eligible implementing types were found in package %s", union.spec.Name.Name, g.pkg.PkgPath)
		}
	}

	return nil
}
