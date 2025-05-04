# Union types and empty interfaces and types
```yaml
fallback_type: "unknown"
```

```go
// Comment for UnionType
type UnionType interface {
    // Comment for fields are possible
    uint64 | string | *bool // comment after

    // Comment for a method
    SomeMethod() string
}

type Derived interface {
    ~int | string // Line comment
}

type Any interface {
    string | any
}

type Empty interface{}

type Something any

type EmptyStruct struct{}
```

```ts
/**
 * Comment for UnionType
 */
export type UnionType = 
    /**
     * Comment for fields are possible
     */
    number /* uint64 */ | string | boolean | undefined // comment after
;
export type Derived = 
    number /* int */ | string // Line comment
;
export type Any = 
    string | unknown;
export type Empty = unknown;
export type Something = any;
export interface EmptyStruct {
}
```

# Values and pointers

```yaml
fallback_type: "unknown"
```

```go

type ValAndPtr[V any, PT *V, Unused ~uint64] struct {
    Val V
    // Comment for ptr field
    Ptr PT // ptr line comment
}

type ABCD[A, B string, C UnionType, D int64 | bool] struct {
    A A `json:"a"`
    B B `json:"b"`
    C C `json:"c"`
    D D `json:"d"`
}

type Foo[A string | uint64, B *A] struct {
    Bar A
    Boo B
}

type WithFooGenericTypeArg[A Foo[string, *string]] struct {
    SomeField A `json:"some_field"`
}

// Should not be output as it's a function
func (f Foo[int, Derived]) DoSomething() {
    panic("something")
}
```

```ts
export interface ValAndPtr<V extends unknown, PT extends (V | undefined), Unused extends number /* uint64 */> {
  Val: V;
  /**
   * Comment for ptr field
   */
  Ptr: PT; // ptr line comment
}
export interface ABCD<A extends string, B extends string, C extends UnionType, D extends number /* int64 */ | boolean> {
  a: A;
  b: B;
  c: C;
  d: D;
}
export interface Foo<A extends string | number /* uint64 */, B extends (A | undefined)> {
  Bar: A;
  Boo: B;
}
export interface WithFooGenericTypeArg<A extends Foo<string, string | undefined>> {
  some_field: A;
}
```

# Single

```go
type Single[S string | uint] struct {
    Field S
}

type SingleSpecific = Single[string]
```

```ts
export interface Single<S extends string | number /* uint */> {
  Field: S;
}
export type SingleSpecific = Single<string>;
```

# Any field

Example for https://github.com/okaris/tygo/issues/65.
```go
type AnyStructField[T any] struct {
  Value     T
  SomeField string
}
```
```ts
export interface AnyStructField<T extends any> {
  Value: T;
  SomeField: string;
}
```

```go
type JsonArray[T any] []T
```

```ts
export type JsonArray<T extends any> = T[];
```