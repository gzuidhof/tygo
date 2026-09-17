package tygo

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

// Generator for one or more input packages, responsible for linking
// them together if necessary.
type Tygo struct {
	conf *Config

	packageGenerators map[string]*PackageGenerator
}

func New(config *Config) *Tygo {
	return &Tygo{
		conf:              config,
		packageGenerators: make(map[string]*PackageGenerator),
	}
}

func (g *Tygo) SetTypeMapping(goType string, tsType string) {
	for _, p := range g.conf.Packages {
		p.TypeMappings[goType] = tsType
	}
}

const (
	syntaxLoadMode = packages.NeedName | packages.NeedSyntax | packages.NeedFiles | packages.NeedCompiledGoFiles
	// typedLoadMode specifies the packages.Load mode with type info, needed for union directives.
	typedLoadMode = syntaxLoadMode | packages.NeedTypes | packages.NeedTypesInfo
)

func (g *Tygo) Generate() error {
	// Load syntax first so packages without union directives avoid type checking.
	pkgs, err := packages.Load(&packages.Config{Mode: syntaxLoadMode}, g.conf.PackageNames()...)
	if err != nil {
		return err
	}

	// Prepare generators and identify the subset that needs type information.
	packageGenerators := make([]*PackageGenerator, 0, len(pkgs))
	typedPackageGenerators := make([]*PackageGenerator, 0)
	for i, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			return fmt.Errorf("%+v", pkg.Errors)
		}

		if len(pkg.GoFiles) == 0 {
			return fmt.Errorf("no input go files for package index %d", i)
		}

		pkgConfig := g.conf.PackageConfig(pkg.ID)

		pkgGen := &PackageGenerator{
			conf:           pkgConfig,
			generatedEnums: make(map[string]bool),
		}
		if err := pkgGen.setPackage(pkg); err != nil {
			return err
		}
		packageGenerators = append(packageGenerators, pkgGen)
		if pkgGen.packageHasUnionDirective() {
			typedPackageGenerators = append(typedPackageGenerators, pkgGen)
		}
	}

	// Reload only annotated packages with the type data needed for method-set checks.
	typedPackagesByPath := make(map[string]*packages.Package, len(typedPackageGenerators))
	if len(typedPackageGenerators) > 0 {
		typedPackagePaths := make([]string, len(typedPackageGenerators))
		for i, pkgGen := range typedPackageGenerators {
			typedPackagePaths[i] = pkgGen.pkg.PkgPath
		}
		typedPackages, err := packages.Load(&packages.Config{Mode: typedLoadMode}, typedPackagePaths...)
		if err != nil {
			return err
		}
		for _, pkg := range typedPackages {
			if len(pkg.Errors) > 0 {
				return fmt.Errorf("%+v", pkg.Errors)
			}
			typedPackagesByPath[pkg.PkgPath] = pkg
		}
	}

	// Add type information to the generators that requested it.
	for _, pkgGen := range typedPackageGenerators {
		typedPkg, ok := typedPackagesByPath[pkgGen.pkg.PkgPath]
		if !ok {
			return fmt.Errorf("failed to load type information for package %s", pkgGen.pkg.PkgPath)
		}
		if err := pkgGen.setPackage(typedPkg); err != nil {
			return err
		}
	}

	// Generate output from fully initialized package generators.
	for _, pkgGen := range packageGenerators {
		g.packageGenerators[pkgGen.pkg.PkgPath] = pkgGen
		code, err := pkgGen.Generate()
		if err != nil {
			return err
		}

		outPath := pkgGen.conf.ResolvedOutputPath(filepath.Dir(pkgGen.pkg.GoFiles[0]))
		err = os.MkdirAll(filepath.Dir(outPath), os.ModePerm)
		if err != nil {
			return nil
		}

		err = os.WriteFile(outPath, []byte(code), 0o664)
		if err != nil {
			return nil
		}
	}
	return nil
}
