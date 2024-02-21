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

func (c *PackageGenerator) writeDirective(s *strings.Builder, cg *ast.CommentGroup) {
	for _, cm := range cg.List {
		if strings.HasPrefix(cm.Text, "//tygo:emit") {
			s.WriteString(strings.TrimPrefix(cm.Text, "//tygo:emit"))
			s.WriteString("\n")
		}
	}
}

func (g *PackageGenerator) writeCommentGroup(s *strings.Builder, cg *ast.CommentGroup, depth int) {
	docLines := strings.Split(cg.Text(), "\n")

	g.writeDirective(s, cg)
	if len(cg.List) > 0 && cg.Text() == "" { // This is a directive comment like //go:embed
		s.WriteByte('\n')
		return
	}

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

// Outputs a comment like // hello world
func (g *PackageGenerator) writeSingleLineComment(s *strings.Builder, cg *ast.CommentGroup) {
	text := cg.Text()

	if len(cg.List) > 0 && cg.Text() == "" { // This is a directive comment like //go:embed
		s.WriteByte('\n')
		return
	}

	s.WriteString(" // " + text)

	if len(text) == 0 {
		// This is an empty comment like //
		s.WriteByte('\n')
	}
}
