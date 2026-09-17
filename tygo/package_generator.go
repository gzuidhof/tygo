package tygo

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Responsible for generating the code for an input package
type PackageGenerator struct {
	conf *PackageConfig
	// Go package associated with this generator (set via setPackage method).
	pkg *packages.Package
	// Absolute file paths of the package's original Go source files.
	// No longer used internally.
	GoFiles []string
	// File ASTs paired with their compiled source paths (excluding ignored files).
	files           []packageFile
	generatedEnums  map[string]bool // Track types that have been generated as enums
	interfaceUnions map[*ast.TypeSpec][]string
}

// packageFile pairs a compiled source path with its syntax tree.
type packageFile struct {
	path   string // Absolute path to the file
	syntax *ast.File
}

// setPackage pairs included syntax trees with CompiledGoFiles, which may differ
// from GoFiles when sources are processed before compilation, such as by cgo.
func (g *PackageGenerator) setPackage(pkg *packages.Package) error {
	if pkg == nil {
		return fmt.Errorf("failed to resolve source files: package is unavailable")
	}
	if len(pkg.Syntax) != len(pkg.CompiledGoFiles) {
		return fmt.Errorf("failed to resolve source files for package %s", pkg.PkgPath)
	}

	files := make([]packageFile, 0, len(pkg.Syntax))
	for i, syntax := range pkg.Syntax {
		path := pkg.CompiledGoFiles[i]
		if !g.conf.IsFileIgnored(path) {
			files = append(files, packageFile{path: path, syntax: syntax})
		}
	}

	g.pkg = pkg
	g.GoFiles = pkg.GoFiles
	g.files = files
	return nil
}

// packageHasUnionDirective scans included files without requiring type information.
func (g *PackageGenerator) packageHasUnionDirective() bool {
	for _, file := range g.files {
		if fileHasUnionDirective(file.syntax) {
			return true
		}
	}
	return false
}

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
	// Resolve unions before writing so later declarations can be included.
	if err := g.analyzeInterfaceUnions(); err != nil {
		return "", err
	}

	s := new(strings.Builder)

	g.writeFileCodegenHeader(s)
	g.writeFileFrontmatter(s)

	for _, file := range g.files {
		g.generateFile(s, file.syntax, file.path)
	}

	return s.String(), nil
}
