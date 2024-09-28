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
const myValue = 3
```

```ts
```