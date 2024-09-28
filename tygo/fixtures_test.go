package tygo

import (
	"embed"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// Embed markdown test fixtures
//
//go:embed testdata/fixtures/*.md
var mdfs embed.FS

type MarkdownFixture struct {
	PackageConfig PackageConfig
	GoCode        string
	TsCode        string
}

func TestConvertGoToTypescriptSmoketest(t *testing.T) {
	t.Parallel()

	goCode := "type MyType uint8"
	tsCode, err := ConvertGoToTypescript(goCode, PackageConfig{})
	require.NoError(t, err)

	expected := `export type MyType = number /* uint8 */;
`
	assert.Equal(t, expected, tsCode)
}

func parseMarkdownFixtures(fileContents []byte) ([]MarkdownFixture, error) {
	fixtures := make([]MarkdownFixture, 0)
	currentFixture := MarkdownFixture{}

	currentBlockContents := ""
	currentBlockLanguage := ""
	inCodeBlock := false
	for _, line := range strings.Split(string(fileContents), "\n") {
		if strings.HasPrefix(line, "```") {
			if inCodeBlock {
				// End of code block
				if currentBlockLanguage == "ts" || currentBlockLanguage == "typescript" {
					// Every fixture ends with a typescript block
					currentFixture.TsCode = currentBlockContents
					fixtures = append(fixtures, currentFixture)
					currentFixture = MarkdownFixture{}
				} else if currentBlockLanguage == "go" {
					currentFixture.GoCode = currentBlockContents
				} else if currentBlockLanguage == "yml" || currentBlockLanguage == "yaml" {
					// Parse package config
					pc := PackageConfig{}
					err := yaml.Unmarshal([]byte(currentBlockContents), &pc)
					if err != nil {
						return nil, fmt.Errorf("failed to unmarshal package config: %w", err)
					}
					currentFixture.PackageConfig = pc
				}
				currentBlockContents = ""
				currentBlockLanguage = ""
			} else { // Start of code block
				language := strings.TrimPrefix(line, "```")
				language = strings.TrimSpace(language)
				currentBlockLanguage = language
			}
			inCodeBlock = !inCodeBlock
			continue
		}

		if inCodeBlock {
			currentBlockContents += line + "\n"
		}
	}

	return fixtures, nil

}

// Tests all markdown files in `testdata/fixtures/` directory.
func TestMarkdownFixtures(t *testing.T) {
	t.Parallel()

	fixtures, err := mdfs.ReadDir("testdata/fixtures")
	require.NoError(t, err)

	for _, fixture := range fixtures {
		fixture := fixture

		// Read markdown file
		md, err := mdfs.ReadFile("testdata/fixtures/" + fixture.Name())
		require.NoError(t, err)

		testCases, err := parseMarkdownFixtures(md)
		require.NoError(t, err)

		for _, tc := range testCases {
			tc := tc
			t.Run(fixture.Name(), func(t *testing.T) {
				t.Parallel()

				tsCode, err := ConvertGoToTypescript(tc.GoCode, tc.PackageConfig)
				require.NoError(t, err)

				assert.Equal(t, tc.TsCode, tsCode)
			})
		}
	}
}
