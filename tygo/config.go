package tygo

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
)

const defaultOutputFilename = "index.ts"
const defaultFallbackType = "any"
const defaultPreserveComments = "default"

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

	// FallbackType defines the Typescript type used as a fallback for unknown Go types.
	FallbackType string `yaml:"fallback_type"`

	// Flavor defines what the key names of the output types will look like.
	// Supported values: "default", "" (same as "default"), "yaml".
	// In "default" mode, `json` and `yaml` tags are respected, but otherwise keys are unchanged.
	// In "yaml" mode, keys are lowercased to emulate gopkg.in/yaml.v2.
	Flavor string `yaml:"flavor"`

	// PreserveComments is an option to preserve comments in the generated TypeScript output.
	// Supported values: "default", "" (same as "default"), "types", "none".
	// By "default", package-level comments as well as type comments are
	// preserved.
	// In "types" mode, only type comments are preserved.
	// If "none" is supplied, no comments are preserved.
	PreserveComments string `yaml:"preserve_comments"`

	// Default interface for Typescript-generated interfaces to extend.
	Extends string `yaml:"extends"`

	// Set the optional type (null or undefined).
	// Supported values: "default", "undefined" (same as "default"), "" (same as "default"), "null".
	// Default is "undefined".
	// Useful for usage with JSON marshalers that output null for optional fields (e.g. gofiber JSON).
	OptionalType string `yaml:"optional_type"`
}

type Config struct {
	TypeMappings map[string]string `yaml:"type_mappings"`
	Packages     []*PackageConfig  `yaml:"packages"`
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
			pc.TypeMappings = c.mergeMappings(pc.TypeMappings)
			pcNormalized, err := pc.Normalize()
			if err != nil {
				log.Fatalf("Error in config for package %s: %s", packagePath, err)
			}

			return &pcNormalized
		}
	}
	log.Fatalf("Config not found for package %s", packagePath)
	return nil
}

func normalizeFlavor(flavor string) (string, error) {
	switch flavor {
	case "", "default":
		return "default", nil
	case "yaml":
		return "yaml", nil
	default:
		return "", fmt.Errorf("unsupported flavor: %s", flavor)
	}
}

func normalizePreserveComments(preserveComments string) (string, error) {
	switch preserveComments {
	case "", "default":
		return "default", nil
	case "types":
		return "types", nil
	case "none":
		return "none", nil
	default:
		return "", fmt.Errorf("unsupported preserve_comments: %s", preserveComments)
	}
}

func normalizeOptionalType(optional string) (string, error) {
	switch optional {
	case "", "default", "undefined":
		return "undefined", nil
	case "null":
		return "null", nil
	default:
		return "", fmt.Errorf("unsupported optional: %s", optional)
	}
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

// Normalize returns a new PackageConfig with default values set.
func (pc PackageConfig) Normalize() (PackageConfig, error) {
	if pc.Indent == "" {
		pc.Indent = "  "
	}

	if pc.FallbackType == "" {
		pc.FallbackType = defaultFallbackType
	}

	if pc.PreserveComments == "" {
		pc.PreserveComments = defaultPreserveComments
	}

	var err error
	pc.Flavor, err = normalizeFlavor(pc.Flavor)
	if err != nil {
		return pc, fmt.Errorf("invalid flavor config for package %s: %s", pc.Path, err)
	}

	pc.PreserveComments, err = normalizePreserveComments(pc.PreserveComments)
	if err != nil {
		return pc, fmt.Errorf("invalid preserve_comments config for package %s: %s", pc.Path, err)
	}

	pc.OptionalType, err = normalizeOptionalType(pc.OptionalType)
	if err != nil {
		return pc, fmt.Errorf("invalid optional_type config for package %s: %s", pc.Path, err)
	}

	return pc, nil
}

func (c Config) mergeMappings(pkg map[string]string) map[string]string {
	mappings := make(map[string]string)
	for k, v := range c.TypeMappings {
		mappings[k] = v
	}
	for k, v := range pkg {
		mappings[k] = v
	}
	return mappings
}
