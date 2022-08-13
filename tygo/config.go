package tygo

import (
	"log"
	"path/filepath"
	"strings"
)

const defaultOutputFilename = "index.ts"

type PackageConfig struct {
	// The package path just like you would import it in Go
	Path string `yaml:"path"`

	// Where this output should be written to.
	// If you specify a folder it will be written to a file `index.ts` within that folder. By default it is written into the Golang package folder.
	OutputPath string `yaml:"output_path"`

	// Customize the indentation (use \t if you want tabs)
	Indent string `yaml:"indent"`

	// Specify your own custom type translations, useful for custom types, `time.Time` and `null.String`.
	// Be default unrecognized types will be output as `any /* name */`.
	TypeMappings map[string]string `yaml:"type_mappings"`

	// This content will be put at the top of the output Typescript file.
	// You would generally use this to import custom types.
	Frontmatter string `yaml:"frontmatter"`

	// Filenames of Go source files that should not be included in the Typescript output.
	ExcludeFiles []string `yaml:"exclude_files"`

	// Filenames of Go source files that should be included in the Typescript output.
	IncludeFiles []string `yaml:"include_files"`
}

type Config struct {
	Packages []*PackageConfig `yaml:"packages"`
}

func (c Config) PackageNames() []string {
	names := make([]string, len(c.Packages))

	for i, p := range c.Packages {
		names[i] = p.Path
	}
	return names
}

func (c Config) PackageConfig(packagePath string) *PackageConfig {
	for _, pc := range c.Packages {
		if pc.Path == packagePath {
			if pc.Indent == "" {
				pc.Indent = "  "
			}
			return pc
		}
	}
	log.Fatalf("Config not found for package %s", packagePath)
	return nil
}

func (c PackageConfig) IsFileIgnored(pathToFile string) bool {
	basename := filepath.Base(pathToFile)
	for _, ef := range c.ExcludeFiles {
		if basename == ef {
			return true
		}
	}

	// if defined, only included files are allowed
	if len(c.IncludeFiles) > 0 {
		for _, include := range c.IncludeFiles {
			if basename == include {
				return false
			}
		}
		return true
	}

	return false
}

func (c PackageConfig) ResolvedOutputPath(packageDir string) string {
	if c.OutputPath == "" {
		return filepath.Join(packageDir, defaultOutputFilename)
	} else if !strings.HasSuffix(c.OutputPath, ".ts") {
		return filepath.Join(c.OutputPath, defaultOutputFilename)
	}
	return c.OutputPath
}
