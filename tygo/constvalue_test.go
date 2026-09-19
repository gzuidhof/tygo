package tygo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestConstantValuesAreValuesNotTypes covers the case where a const's value is not a
// literal. writeValueSpec used to render values with writeType — the TYPE renderer — so
// anything writeType did not recognise came out as the fallback TYPE:
//
//	export const Computed = any;
//
// which is not valid TypeScript. It has to go through the real package loader rather than
// ConvertGoToTypescript, because the fix reads the value go/types computed and the
// string-conversion path has no type information at all.
func TestConstantValuesAreValuesNotTypes(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) {
		require.NoError(t, os.MkdirAll(filepath.Dir(filepath.Join(dir, name)), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644))
	}
	write("go.mod", "module tygoconsttest\n\ngo 1.21\n")
	// A second package, because a constant declared elsewhere is the case the AST cannot
	// resolve even in principle — writeType sees only `other.Name`.
	write("other/other.go", "package other\n\nconst Name = \"from-another-package\"\n")
	write("iota/iota.go", `package iota_

type Priority int

const (
	Low Priority = iota
	Medium
	High
)

const (
	KB = 1 << (10 * (iota + 1))
	MB
	GB
)
`)
	write("sub/types.go", `package sub

import "tygoconsttest/other"

const (
	Literal   = "plain"
	Computed  = len("abc")
	Negative  = -7
	Shifted   = 1 << 10
	CrossPkg  = other.Name
	Concat    = "a" + "b"
	Float     = 1.5
	Truthy    = true
)
`)
	out := filepath.Join(dir, "out.ts")
	gen := New(&Config{Packages: []*PackageConfig{{
		Path:       "tygoconsttest/sub",
		OutputPath: out,
	}}})

	cwd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	defer func() { _ = os.Chdir(cwd) }()

	require.NoError(t, gen.Generate())

	b, err := os.ReadFile(out)
	require.NoError(t, err)
	ts := string(b)

	// The bug: the fallback TYPE emitted where a value belongs.
	require.NotContains(t, ts, "= any",
		"a const value was rendered as a type; `any` is not a valid TypeScript value")

	for _, want := range []string{
		`export const Literal = "plain";`,
		`export const Computed = 3;`,                      // len(), an *ast.CallExpr
		`export const Negative = -7;`,                     // *ast.UnaryExpr (cf. issue #42)
		`export const Shifted = 1024;`,                    // *ast.BinaryExpr, shift
		`export const CrossPkg = "from-another-package";`, // *ast.SelectorExpr
		`export const Concat = "ab";`,
		`export const Float = 1.5;`,
		`export const Truthy = true;`,
	} {
		require.True(t, strings.Contains(ts, want), "missing %q in:\n%s", want, ts)
	}
}

// TestIotaGroupsAreLeftToTheASTPath guards a regression this fix introduced and that the
// existing fixtures could not catch, because they run through ConvertGoToTypescript, which
// parses ASTs and never populates TypesInfo — so they never reach the resolution path.
//
// Resolving an iota expression breaks the group two ways at once: go/types evaluates the
// SAME AST node once per ConstSpec and overwrites TypesInfo.Types[expr], so the recorded
// value is the LAST iteration's; and writeValueSpec's replaceIotaValue needs the
// un-substituted text to fill in per entry. Before mentionsIota, this produced
// Low=2, Medium=2, High=2 and KB=MB=GB.
func TestIotaGroupsAreLeftToTheASTPath(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) {
		require.NoError(t, os.MkdirAll(filepath.Dir(filepath.Join(dir, name)), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644))
	}
	write("go.mod", "module tygoiotatest\n\ngo 1.21\n")
	write("sub/types.go", `package sub

type Priority int

const (
	Low Priority = iota
	Medium
	High
)

const (
	KB = 1 << (10 * (iota + 1))
	MB
	GB
)
`)
	out := filepath.Join(dir, "out.ts")
	gen := New(&Config{Packages: []*PackageConfig{{
		Path:       "tygoiotatest/sub",
		OutputPath: out,
	}}})

	cwd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	defer func() { _ = os.Chdir(cwd) }()
	require.NoError(t, gen.Generate())

	b, err := os.ReadFile(out)
	require.NoError(t, err)
	ts := string(b)

	for _, want := range []string{
		"export const Low: Priority = 0;",
		"export const Medium: Priority = 1;",
		"export const High: Priority = 2;",
		"export const KB = 1 << (10 * (0 + 1));",
		"export const MB = 1 << (10 * (1 + 1));",
		"export const GB = 1 << (10 * (2 + 1));",
	} {
		require.True(t, strings.Contains(ts, want),
			"iota group collapsed; missing %q in:\n%s", want, ts)
	}
}
