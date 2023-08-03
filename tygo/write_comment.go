package tygo

import (
	"go/ast"
	"strings"
)

func (g *PackageGenerator) PreserveDocComments() bool {
	return g.conf.PreserveComments == "default"
}

func (g *PackageGenerator) PreserveTypeComments() bool {
	return g.conf.PreserveComments == "types" || g.conf.PreserveComments == "default"
}

func (g *PackageGenerator) writeCommentGroupIfNotNil(s *strings.Builder, f *ast.CommentGroup, depth int) {
	if f != nil {
		g.writeCommentGroup(s, f, depth)
	}
}

func (g *PackageGenerator) writeCommentGroup(s *strings.Builder, f *ast.CommentGroup, depth int) {
	docLines := strings.Split(f.Text(), "\n")

	if depth != 0 {
		g.writeIndent(s, depth)
	}
	s.WriteString("/**\n")

	for _, c := range docLines {
		if len(strings.TrimSpace(c)) == 0 {
			continue
		}
		g.writeIndent(s, depth)
		s.WriteString(" * ")
		c = strings.ReplaceAll(c, "*/", "*\\/") // An edge case: a // comment can contain */
		s.WriteString(c)
		s.WriteByte('\n')
	}
	g.writeIndent(s, depth)
	s.WriteString(" */\n")
}
