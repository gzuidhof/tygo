package tygo

import (
	"fmt"
	"go/parser"
	"go/token"
	"strings"
)

// ConvertGoToTypescript converts Go code string to Typescript.
//
// This is mostly useful for testing purposes inside tygo itself.
func ConvertGoToTypescript(goCode string, pkgConfig PackageConfig) (string, error) {
	src := fmt.Sprintf(`package tygoconvert

%s`, goCode)

	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, "", src, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("failed to parse source: %w", err)
	}

	pkgConfig, err = pkgConfig.Normalize()
	if err != nil {
		return "", fmt.Errorf("failed to normalize package config: %w", err)
	}

	pkgGen := &PackageGenerator{
		conf:           &pkgConfig,
		pkg:            nil,
		generatedEnums: make(map[string]bool),
	}

	s := new(strings.Builder)

	pkgGen.generateFile(s, f, "")
	code := s.String()

	return code, nil
}
