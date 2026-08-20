package tygo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunFormatter(t *testing.T) {
	t.Setenv("GO_WANT_TYGO_FORMATTER_HELPER", "1")

	dir := t.TempDir()
	firstPath := filepath.Join(dir, "first.ts")
	secondPath := filepath.Join(dir, "second.ts")
	require.NoError(t, os.WriteFile(firstPath, []byte("first"), 0o600))
	require.NoError(t, os.WriteFile(secondPath, []byte("second"), 0o600))

	err := runFormatter(&FormatterConfig{
		Command: os.Args[0],
		Args:    []string{"-test.run=TestFormatterHelperProcess", "--", "formatted"},
	}, []string{firstPath, secondPath})
	require.NoError(t, err)

	first, err := os.ReadFile(firstPath)
	require.NoError(t, err)
	assert.Equal(t, "first formatted", string(first))
	second, err := os.ReadFile(secondPath)
	require.NoError(t, err)
	assert.Equal(t, "second formatted", string(second))
}

func TestRunFormatterFailure(t *testing.T) {
	err := runFormatter(&FormatterConfig{Command: "command-that-does-not-exist"}, []string{"output.ts"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "formatter \"command-that-does-not-exist\" failed")
}

func TestFormatterHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_TYGO_FORMATTER_HELPER") != "1" {
		return
	}

	separator := 0
	for i, arg := range os.Args {
		if arg == "--" {
			separator = i
			break
		}
	}
	if separator == 0 || len(os.Args) <= separator+2 {
		os.Exit(2)
	}

	suffix := os.Args[separator+1]
	for _, path := range os.Args[separator+2:] {
		contents, err := os.ReadFile(path)
		if err != nil {
			os.Exit(2)
		}
		contents = append(contents, []byte(" "+suffix)...)
		if err := os.WriteFile(path, contents, 0o600); err != nil {
			os.Exit(2)
		}
	}
	os.Exit(0)
}
