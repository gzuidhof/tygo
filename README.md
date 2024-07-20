# 🎑 tygo

Tygo is a tool for generating Typescript typings from Golang source files that just works.

It preserves comments, understands constants and also supports non-struct `type` expressions. It's perfect for generating equivalent types for a Golang REST API to be used in your front-end codebase.

**🚀 Supports Golang 1.18 generic types and struct inheritance**

## Installation

```shell
go install github.com/gzuidhof/tygo@latest
```

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
	Ignored    bool        `tstype:"-"` // Fields with - are omitted too
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
  [key: string]: {
    [key: number /* uint16 */]: number /* uint32 */ | undefined;
  };
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

### Option A: CLI (recommended)

Create a file `tygo.yaml` in which you specify which packages are to be converted and any special type mappings you want to add.

```yaml
packages:
  - path: "github.com/gzuidhof/tygo/examples/bookstore"
    type_mappings:
      time.Time: "string /* RFC3339 */"
      null.String: "null | string"
      null.Bool: "null | boolean"
      uuid.UUID: "string /* uuid */"
      uuid.NullUUID: "null | string /* uuid */"
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

    # Package that the generates Typescript types should extend. This is useful when
    # attaching your types to a generic ORM.
    extends: "SomeType"
```

See also the source file [tygo/config.go](./tygo/config.go).

## Type hints through tagging

You can tag struct fields with `tstype` to specify their output Typescript type.

### Custom type mapping

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

**Alternative**

You could use the `frontmatter` field in the config to inject `export type Genre = "novel" | "crime" | "fantasy"` at the top of the file, and use `tstype:"Genre"`. I personally prefer that as we may use the `Genre` type more than once.

**`tygo:emit` directive**

Another way to generate types that cannot be directly represented in Go is to use a `//tygo:emit` directive to 
directly emit literal TS code.
The directive can be used in two ways. A `tygo:emit` directive on a struct will emit the remainder of the directive 
text before the struct.
```golang
// Golang input

//tygo:emit export type Genre = "novel" | "crime" | "fantasy"
type Book struct {
	Title    string    `json:"title"`
	Genre    string    `json:"genre" tstype:"Genre"`
}
```

```typescript
export type Genre = "novel" | "crime" | "fantasy"

export interface Book {
  title: string;
  genre: Genre;
}
```

A `//tygo:emit` directive on a string var will emit the contents of the var, useful for multi-line content.
```golang
//tygo:emit
var _ = `export type StructAsTuple=[
  a:number, 
  b:number, 
  c:string,
]
`
type CustomMarshalled struct {
  Content []StructAsTuple `json:"content"`
}
```

```typescript
export type StructAsTuple=[
  a:number, 
  b:number, 
  c:string,
]

export interface CustomMarshalled {
  content: StructAsTuple[];
}

```

Generating types this way is particularly useful for tuple types, because a comma cannot be used in the `tstype` tag.

### Required fields

Pointer type fields usually become optional in the Typescript output, but sometimes you may want to require it regardless.

You can add `,required` to the `tstype` tag to mark a pointer type as required.

```golang
// Golang input
type Nicknames struct {
	Alice   *string `json:"alice"`
	Bob     *string `json:"bob" tstype:"BobCustomType,required"`
	Charlie *string `json:"charlie" tstype:",required"`
}
```

```typescript
// Typescript output
export interface Nicknames {
  alice?: string;
  bob: BobCustomType;
  charlie: string;
}
```

### Readonly fields

Sometimes a field should be immutable, you can add `,readonly` to the `tstype` tag to mark a field as `readonly`.

```golang
// Golang input
type Cat struct {
	Name    string `json:"name,readonly"`
	Owner   string `json:"owner"`
}
```

```typescript
// Typescript output
export interface Cat {
  readonly name: string;
  owner: string;
}
```

## Inheritance

Tygo supports interface inheritance. To extend an `inlined` struct, use the tag `tstype:",extends"` on struct fields you wish to extend. Only `struct` types can be extended.

Struct pointers are optionally extended using `Partial<MyType>`. To mark these structs as required, use the tag `tstype:",extends,required"`.

Named `struct fields` can also be extended.

Example usage [here](examples/inheritance)

```go
// Golang input
import "example.com/external"

type Base struct {
	Name string `json:"name"`
}

type Base2[T string | int] struct {
	ID T `json:"id"`
}

type OptionalPtr struct {
	Field string `json:"field"`
}

type Other[T int] struct {
	*Base                  `       tstype:",extends,required"`
	Base2[T]               `       tstype:",extends"`
	*OptionalPtr           `       tstype:",extends"`
	external.AnotherStruct `       tstype:",extends"`
	OtherValue             string `                  json:"other_value"`
}
```

```typescript
// Typescript output
export interface Base {
  name: string;
}

export interface Base2<T extends string | number /* int */> {
  id: T;
}

export interface OptionalPtr {
  field: string;
}

export interface Other<T extends number /* int */>
  extends Base,
    Base2<T>,
    Partial<OptionalPtr>,
    external.AnotherStruct {
  other_value: string;
}
```

## Generics

Tygo supports generic types (Go version >= 1.18) out of the box.

```go
// Golang input
type UnionType interface {
	uint64 | string
}

type ABCD[A, B string, C UnionType, D int64 | bool] struct {
	A A `json:"a"`
	B B `json:"b"`
	C C `json:"c"`
	D D `json:"d"`
}
```

```typescript
// Typescript output
export type UnionType = number /* uint64 */ | string;

export interface ABCD<
  A extends string,
  B extends string,
  C extends UnionType,
  D extends number /* int64 */ | boolean
> {
  a: A;
  b: B;
  c: C;
  d: D;
}
```

## YAML support

Tygo supports generating typings for YAML-serializable objects that can be understood by Go apps.

By default, Tygo will respect `yaml` Go struct tags, in addition to `json`, but it will not apply any transformations to untagged fields.
However, the default behavior of the popular `gopkg.in/yaml.v2` package for Go structs without tags is to downcase the struct field names.
To emulate this behavior, one can use the `flavor` configuration option:

```yaml
packages:
  - path: "github.com/my/package"
    output_path: "webapp/api/types.ts"
    flavor: "yaml"
```

```go
// Golang input
type Foo struct {
	TaggedField string `yaml:"custom_field_name_in_yaml"`
    UntaggedField string
}
```

```typescript
// Typescript output
export interface Foo {
  custom_field_name_in_yaml: string;
  untaggedfield: string;
}
```

## Related projects

- [**typescriptify-golang-structs**](https://github.com/tkrajina/typescriptify-golang-structs): Probably the most popular choice. The downside of this package is that it relies on reflection rather than parsing, which means that certain things can't be kept such as comments without adding a bunch of tags to your structs. The CLI generates a Go file which is then executed and reflected on. The library requires you to manually specify all types that should be converted.
- [**go2ts**](https://github.com/StirlingMarketingGroup/go2ts): A transpiler with a web interface, this project can be seen as an evolution of this project. It's perfect for quick one-off transpilations. There is no CLI, no support for `const` and there are no ways to customize the output.

**If `tygo` is useful for your project, consider leaving a star.**

## License

[MIT](./LICENSE)
