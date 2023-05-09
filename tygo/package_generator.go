package tygo

import (
	"go/ast"
	"go/token"
	"strings"
)

func (g *PackageGenerator) Generate() (string, error) {
	s := new(strings.Builder)

	g.writeFileCodegenHeader(s)
	g.writeFileFrontmatter(s)

	filepaths := g.GoFiles

	for i, file := range g.pkg.Syntax {
		if g.conf.IsFileIgnored(filepaths[i]) {
			continue
		}

		first := true

		ast.Inspect(file, func(n ast.Node) bool {
			switch x := n.(type) {

			// GenDecl can be an import, type, var, or const expression
			case *ast.GenDecl:
				if x.Tok == token.VAR || x.Tok == token.IMPORT {
					return false
				}

				if first {
					preserveDocComments := g.conf.PreserveComments == "default"
					g.writeFileSourceHeader(s, filepaths[i], file, preserveDocComments)
					first = false
				}

				preserveGroupComments := g.conf.PreserveComments == "default"
				preserveTypeComments := g.conf.PreserveComments == "default" || g.conf.PreserveComments == "types"
				g.writeGroupDecl(s, x, preserveGroupComments, preserveTypeComments)
				return false
			}
			return true

		})

	}

	return s.String(), nil
}
