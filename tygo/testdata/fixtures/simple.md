```go
type MyUint8 uint8
type MyInt int
type MyString string
type MyAny any

// Should be a number in TypeScript.
type MyRune rune
```

```ts
export type MyUint8 = number /* uint8 */;
export type MyInt = number /* int */;
export type MyString = string;
export type MyAny = any;
/**
 * Should be a number in TypeScript.
 */
export type MyRune = number /* rune */;
```


Struct with some comments
```go
// Comment for a struct
type MyStruct struct {
    SomeField any `json:"some_field"`
    // Comment for a field
    OtherField bool // Comment after line
    FieldWithImportedType some.Type
}
```

```ts
/**
 * Comment for a struct
 */
export interface MyStruct {
  some_field: any;
  /**
   * Comment for a field
   */
  OtherField: boolean; // Comment after line
  FieldWithImportedType: any /* some.Type */;
}
```

No preserve comments
```yaml
preserve_comments: "none"
```

```go
// Foo
type MyValue int // Bar
```

```ts
export type MyValue = number /* int */;
```

Empty file
```go
```

```ts
```

Unexported

```go
// A comment on an unexported constant
const myValue = 3

// A comment on an unexported type
type myType struct {
  // A comment on an unexported field
  field string
}

// Mixed unexported and exported 
const (
  unexportedValue = 7 // A comment on an unexported constant
  ExportedValue = 42 // A comment on an exported constant
)

// Unexported group
const (
  unexportedValue1 = 1 // A comment on an unexported constant
  unexportedValue2 = 2 // Another comment on an unexported constant
)
```

```ts
/**
 * Mixed unexported and exported
 */
export const ExportedValue = 42; // A comment on an exported constant
```


Comma

```go
type A struct {
  Foo, Bar, baz string // A comment on fields separated by a comma
}

type B struct {
  // A comment above the fields separated by a comma
  Foo, Bar, baz string
}


```
```ts
export interface A {
  Foo: string; // A comment on fields separated by a comma
  Bar: string; // A comment on fields separated by a comma
}
export interface B {
  /**
   * A comment above the fields separated by a comma
   */
  Foo: string;
  /**
   * A comment above the fields separated by a comma
   */
  Bar: string;
}
```

```go
const Pi, E = 3.14, 2.71 // A comment on constants separated by a comma
```
```ts
export const Pi = 3.14; // A comment on constants separated by a comma
export const E = 2.71; // A comment on constants separated by a comma
```