package tygo

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/fatih/structtag"
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

// GetFieldTagInfo returns cached or freshly parsed tag info for a field
func (g *PackageGenerator) GetFieldTagInfo(f *ast.Field) *FieldTagInfo {
	if info := g.tagCache.Get(f); info != nil {
		return info
	}

	info := &FieldTagInfo{
		isEmbedded: isAnonymousField(f),
	}

	if f.Tag != nil {
		tags, err := structtag.Parse(f.Tag.Value[1 : len(f.Tag.Value)-1])
		if err != nil {
			panic(err)
		}

		jsonTag, err := tags.Get("json")
		if err == nil {
			info.jsonTag = jsonTag
			info.name = jsonTag.Name
			if jsonTag.Name == "-" {
				info.skip = true
			}
			info.optional = jsonTag.HasOption("omitempty") || jsonTag.HasOption("omitzero")
		}

		yamlTag, err := tags.Get("yaml")
		if err == nil {
			info.yamlTag = yamlTag
			info.name = yamlTag.Name
			if yamlTag.Name == "-" {
				info.skip = true
			}
			info.optional = yamlTag.HasOption("omitempty")
		}

		tstypeTag, err := tags.Get("tstype")
		if err == nil {
			info.tstypeTag = tstypeTag
			if tstypeTag.Name == "-" {
				info.skip = true
			}
			info.required = tstypeTag.HasOption("required")
			info.readonly = tstypeTag.HasOption("readonly")
			if tstypeTag.HasOption("extends") {
				info.isExtend = true
			}
		}
	}

	// Apply default behavior for embedded/anonymous fields based on configuration
	if info.isEmbedded && !info.skip && g.conf.EmbeddedFieldBehavior != "" {
		// Parse the embedded field behavior as comma-separated options
		options := strings.Split(g.conf.EmbeddedFieldBehavior, ",")
		for _, opt := range options {
			opt = strings.TrimSpace(opt)
			switch opt {
			case "extends":
				info.isExtend = true
			case "required":
				info.required = true
			}
		}
	}

	g.tagCache.Set(f, info)
	return info
}
