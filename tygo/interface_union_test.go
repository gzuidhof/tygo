package tygo

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
)

func TestInterfaceUnion(t *testing.T) {
	t.Parallel()

	goCode := `type Before struct {
	Value string ` + "`json:\"value\"`" + `
}

func (*Before) isMessage() {}

type marker interface {
	isMessage()
}

// Message is a closed set of messages.
//tygo:union
type Message interface {
	marker
}

type After struct {
	Code int ` + "`json:\"code\"`" + `
}

func (After) isMessage() {}

type AfterAlias = After

type Ordinary interface {
	isMessage()
}

type Empty interface{}
`

	tsCode, err := ConvertGoToTypescript(goCode, PackageConfig{})
	require.NoError(t, err)

	expected := `export interface Before {
  value: string;
}
/**
 * Message is a closed set of messages.
 */
export type Message = Before | After;
export interface After {
  code: number /* int */;
}
export type AfterAlias = After;
export type Ordinary = any;
export type Empty = any;
`
	assert.Equal(t, expected, tsCode)
}

func TestInterfaceUnionErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		goCode      string
		errorString string
	}{
		{
			name: "non-interface",
			goCode: `//tygo:union
type Message struct{}
`,
			errorString: "cannot generate union for Message: //tygo:union may only be applied to a named interface",
		},
		{
			name: "no implementers",
			goCode: `//tygo:union
type Message interface { isMessage() }
`,
			errorString: "cannot generate union for interface Message: no eligible implementing types were found in package tygoconvert",
		},
		{
			name: "generic interface",
			goCode: `//tygo:union
type Message[T any] interface { isMessage(T) }
`,
			errorString: "cannot generate union for interface Message: generic interfaces are not supported",
		},
		{
			name: "generic implementer",
			goCode: `//tygo:union
type Message interface { isMessage() }

type Generic[T any] struct{}
func (Generic[T]) isMessage() {}
`,
			errorString: "cannot generate union for interface Message: generic implementing type Generic is not supported",
		},
		{
			name: "unexported implementer",
			goCode: `//tygo:union
type Message interface { isMessage() }

type hidden struct{}
func (hidden) isMessage() {}
`,
			errorString: "cannot generate union for interface Message: no eligible implementing types were found in package tygoconvert",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			_, err := ConvertGoToTypescript(testCase.goCode, PackageConfig{})
			require.EqualError(t, err, testCase.errorString)
		})
	}
}

func TestInterfaceUnionTypeCheckError(t *testing.T) {
	t.Parallel()

	_, err := ConvertGoToTypescript(`//tygo:union
type Message interface { isMessage(Undefined) }
`, PackageConfig{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to type-check interface union")
	assert.Contains(t, err.Error(), "undefined: Undefined")
}

func TestInterfaceUnionModuleImportError(t *testing.T) {
	t.Parallel()

	_, err := ConvertGoToTypescript(`import "github.com/google/uuid"

//tygo:union
type Entity interface { entity() }

type User struct { ID uuid.UUID }
func (User) entity() {}
`, PackageConfig{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to type-check interface union")
}

func TestInterfaceUnionExcludesFiles(t *testing.T) {
	t.Parallel()

	fileSet := token.NewFileSet()
	interfaceFile, err := parser.ParseFile(fileSet, "interface.go", `package test

//tygo:union
type Message interface { isMessage() }
`, parser.ParseComments)
	require.NoError(t, err)
	includedFile, err := parser.ParseFile(fileSet, "included.go", `package test

type Included struct{}
func (Included) isMessage() {}
`, parser.ParseComments)
	require.NoError(t, err)
	excludedFile, err := parser.ParseFile(fileSet, "excluded.go", `package test

type Excluded struct{}
func (Excluded) isMessage() {}
`, parser.ParseComments)
	require.NoError(t, err)

	files := []*ast.File{interfaceFile, includedFile, excludedFile}
	typesInfo := &types.Info{Defs: make(map[*ast.Ident]types.Object)}
	typesPackage, err := (&types.Config{Importer: importer.Default()}).Check("test", fileSet, files, typesInfo)
	require.NoError(t, err)

	config, err := (PackageConfig{ExcludeFiles: []string{"excluded.go"}}).Normalize()
	require.NoError(t, err)
	syntaxPackage := &packages.Package{
		Fset:            fileSet,
		GoFiles:         []string{"excluded.go", "included.go", "interface.go"},
		CompiledGoFiles: []string{"interface.go", "included.go", "excluded.go"},
		Syntax:          files,
	}
	directiveGenerator := &PackageGenerator{conf: &config}
	require.NoError(t, directiveGenerator.setPackage(syntaxPackage))
	hasDirective := directiveGenerator.packageHasUnionDirective()
	assert.True(t, hasDirective)

	generator := &PackageGenerator{conf: &config}
	require.NoError(t, generator.setPackage(&packages.Package{
		PkgPath:         "test",
		Fset:            fileSet,
		Syntax:          files,
		CompiledGoFiles: []string{"interface.go", "included.go", "excluded.go"},
		Types:           typesPackage,
		TypesInfo:       typesInfo,
	}))

	require.NoError(t, generator.analyzeInterfaceUnions())
	assert.Equal(t, []string{"Included"}, generator.interfaceUnions[interfaceFile.Decls[0].(*ast.GenDecl).Specs[0].(*ast.TypeSpec)])

	config.ExcludeFiles = []string{"included.go", "excluded.go"}
	require.NoError(t, generator.setPackage(generator.pkg))
	err = generator.analyzeInterfaceUnions()
	require.EqualError(t, err, "cannot generate union for interface Message: no eligible implementing types were found in package test")

	config.ExcludeFiles = []string{"interface.go"}
	require.NoError(t, directiveGenerator.setPackage(syntaxPackage))
	hasDirective = directiveGenerator.packageHasUnionDirective()
	assert.False(t, hasDirective)
}

func TestGenerateLoadsTypesForUnionPackages(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "index.ts")
	generator := New(&Config{Packages: []*PackageConfig{{
		Path:       "github.com/gzuidhof/tygo/tygo/testdata/unionpackage",
		OutputPath: outputPath,
	}}})

	require.NoError(t, generator.Generate())
	output, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Contains(t, string(output), "export type Message = Success;")
	assert.Contains(t, string(output), "export interface Success {")
}
