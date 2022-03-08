# 🎑 tygo

Tygo is a tool for generating Typescript typings from Golang source files that just works.

Other than reflection-based methods it preserves comments, understands constants and also supports non-struct `type` expressions. It's perfect for generating equivalent types for a Golang REST API to be used in your front-end codebase.

## Example

_Golang input file_

```go
// Comments are kept :)
type ComplexType map[string]map[uint16]*uint32

type UserRole = string
const (
	UserRoleDefault UserRole = "viewer"
	UserRoleEditor  UserRole = "editor" // Line comments are also kept
)

type UserEntry struct {
	// Instead of specifying `tstype` we could also declare the typing
	// for uuid.NullUUID in the config file.
	ID uuid.NullUUID `json:"id" tstype:"string | null"`

	Preferences map[string]struct {
		Foo uint32 `json:"foo"`
		// An unknown type without a `tstype` tag or mapping in the config file
		// becomes `any`
		Bar uuid.UUID `json:"bar"`
	} `json:"prefs"`

	MaybeFieldWithStar *string  `json:"address"`
	Nickname           string   `json:"nickname,omitempty"`
	Role               UserRole `json:"role"`

	Complex    ComplexType `json:"complex"`
	unexported bool        // Unexported fields are omitted
}

type ListUsersResponse struct {
	Users []UserEntry `json:"users"`
}
```

_Typescript output_

```typescript
/**
 * Comments are kept :)
 */
export type ComplexType = {
  [key: string]: { [key: number /* uint16 */]: number /* uint32 */ | undefined };
};
export type UserRole = string;
export const UserRoleDefault: UserRole = "viewer";
export const UserRoleEditor: UserRole = "editor"; // Line comments are also kept
export interface UserEntry {
  /**
   * Instead of specifying `tstype` we could also declare the typing
   * for uuid.NullUUID in the config file.
   */
  id: string | null;
  prefs: {
    [key: string]: {
      foo: number /* uint32 */;
      /**
       * An unknown type without a `tstype` tag or mapping in the config file
       * becomes `any`
       */
      bar: any /* uuid.UUID */;
    };
  };
  address?: string;
  nickname?: string;
  role: UserRole;
  complex: ComplexType;
}
export interface ListUsersResponse {
  users: UserEntry[];
}
```

For a real baptism by fire example, [here is a Gist with output for the Go built-in `net/http` and `time` package](https://gist.github.com/gzuidhof/7e192a2f33d8a4f5bde5b77fb2c5048c).

## Usage

There are two ways to use this project, either you can use the CLI tool `tygo` (recommended) or as a library.

### Option A: CLI (recommended)

Create a file `tygo.yaml`, in that file you specify which packages are to be converted and any type mappings you want to override.

```yaml
packages:
  - path: "github.com/gzuidhof/tygo/examples/bookstore"
    type_mappings:
      time.Time: "string /* RFC 3339 formatted */"
      null.String: "string | null"
      uuid.UUID: "string"
```

Then run

```shell
tygo generate
```

The output Typescript file will be next to the Go source files.

### Option B: Library-mode

```go
config := &tygo.Config{
  Packages: []*tygo.PackageConfig{
      &tygo.PackageConfig{
          Path: "github.com/gzuidhof/tygo/examples/bookstore",
      },
  },
}
gen := tygo.New(config)
err := gen.Generate()
```

## Config

```yaml
# You can specify more than one package
packages:
  # The package path just like you would import it in Go
  - path: "github.com/my/package"

    # Where this output should be written to.
    # If you specify a folder it will be written to a file `index.ts` within that folder. By default it is written into the Golang package folder.
    output_path: "webapp/api/types.ts"

    # Customize the indentation (use \t if you want tabs)
    indent: "    "

    # Specify your own custom type translations, useful for custom types, `time.Time` and `null.String`.
    # Be default unrecognized types will be `any`.
    type_mappings:
      time.Time: "string"
      my.Type: "SomeType"

    # This content will be put at the top of the output Typescript file, useful for importing custom types.
    frontmatter: |
      "import {SomeType} from "../lib/sometype.ts"

    # Filenames of Go source files that should not be included
    # in the output.
    exclude_files:
      - "private_stuff.go"
```

See also the source file [tygo/config.go](./tygo/config.go).

## Type hints through tagging

You can tag struct fields with `tstype` to specify their output Typescript type. For instance:

```golang
// Golang input

type Book struct {
	Title    string    `json:"title"`
	Genre    string    `json:"genre" tstype:"'novel' | 'crime' | 'fantasy'"`
}
```

```typescript
// Typescript output

export interface Book {
  title: string;
  genre: "novel" | "crime" | "fantasy";
}
```

Alternatively you could use the `frontmatter` in the `tygo.yaml` config file to inject `export Genre = "novel" | "crime" | "fantasy"` at the top of the file, and use `tstype:"Genre"`.

## Related projects

- [**typescriptify-golang-structs**](https://github.com/tkrajina/typescriptify-golang-structs): Probably the most popular choice. The issue with this package is that it relies on reflection rather than parsing, which means that certain things can't be kept such comments. The CLI generates Go file which is then executed and reflected on, and its library requires you to manually specify all types that should be converted.
- [**go2ts**](https://github.com/StirlingMarketingGroup/go2ts): A transpiler with a web interface, this project was based off this project. It's perfect for quick one-off transpilations. There is no CLI, no support for `const` and there are no ways to customize the output.

## License

[MIT](./LICENSE)
