package tygo

import (
	"go/ast"
	"go/token"
	"strings"
)

// preProcessEnums scans the file for const declarations that will be converted to enums
// and marks the corresponding types to prevent duplicate type declarations
func (g *PackageGenerator) preProcessEnums(file *ast.File) {
	ast.Inspect(file, func(n ast.Node) bool {
		if decl, ok := n.(*ast.GenDecl); ok && decl.Tok == token.CONST {
			if enumGroup := g.detectEnumGroup(decl); enumGroup != nil {
				g.generatedEnums[enumGroup.typeName] = true
			}
		}
		return true
	})
}

// generateFile writes the generated code for a single file to the given strings.Builder.
func (g *PackageGenerator) generateFile(s *strings.Builder, file *ast.File, filepath string) {
	// First pass: identify types that will be generated as enums
	g.preProcessEnums(file)

	first := true

	ast.Inspect(file, func(n ast.Node) bool {
		switch x := n.(type) {
		// GenDecl can be an import, type, var, or const expression
		case *ast.GenDecl:
			if x.Tok == token.IMPORT {
				return false
			}
			isEmit := false
			if x.Tok == token.VAR {
				isEmit = g.isEmitVar(x)
				if !isEmit {
					return false
				}
			}

			if first {
				if filepath != "" {
					g.writeFileSourceHeader(s, filepath, file)
				}
				first = false
			}
			if isEmit {
				g.emitVar(s, x)
				return false
			}
			g.writeGroupDecl(s, x)
			return false
		}
		return true
	})
}

func (g *PackageGenerator) Generate() (string, error) {
	s := new(strings.Builder)

	g.writeFileCodegenHeader(s)
	g.writeFileFrontmatter(s)

	filepaths := g.GoFiles

	for i, file := range g.pkg.Syntax {
		if g.conf.IsFileIgnored(filepaths[i]) {
			continue
		}

		g.generateFile(s, file, filepaths[i])
	}

	return s.String(), nil
}
