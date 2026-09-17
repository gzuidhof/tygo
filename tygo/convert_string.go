package tygo

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

// ConvertGoToTypescript converts Go code string to Typescript.
//
// This is mostly useful for testing purposes inside tygo itself.
func ConvertGoToTypescript(goCode string, pkgConfig PackageConfig) (string, error) {
	src := fmt.Sprintf(`package tygoconvert

%s`, goCode)

	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, "input.go", src, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("failed to parse source: %w", err)
	}

	pkgConfig, err = pkgConfig.Normalize()
	if err != nil {
		return "", fmt.Errorf("failed to normalize package config: %w", err)
	}

	pkgGen := &PackageGenerator{
		conf:           &pkgConfig,
		generatedEnums: make(map[string]bool),
	}
	if fileHasUnionDirective(f) {
		// importer.Default may not resolve module dependencies; this test-oriented helper
		// returns that type-checking failure rather than generating an incomplete union.
		typesInfo := &types.Info{Defs: make(map[*ast.Ident]types.Object)}
		typesConfig := types.Config{
			Importer: importer.Default(),
		}
		typesPackage, err := typesConfig.Check("tygoconvert", fset, []*ast.File{f}, typesInfo)
		if err != nil {
			return "", fmt.Errorf("failed to type-check interface union: %w", err)
		}
		pkg := &packages.Package{
			PkgPath:         "tygoconvert",
			Fset:            fset,
			Syntax:          []*ast.File{f},
			CompiledGoFiles: []string{"input.go"},
			Types:           typesPackage,
			TypesInfo:       typesInfo,
		}
		if err := pkgGen.setPackage(pkg); err != nil {
			return "", err
		}
		if err := pkgGen.analyzeInterfaceUnions(); err != nil {
			return "", err
		}
	}

	s := new(strings.Builder)

	pkgGen.generateFile(s, f, "")
	code := s.String()

	return code, nil
}
