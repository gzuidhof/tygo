Structs that would generate an interface with no fields are dropped.

```yaml
exclude_empty_structs: true
```

```go
type Marker struct{}

type Data struct {
	Value string `json:"value"`
}
```

```ts
export interface Data {
  value: string;
}
```

Combined with `match_tags`, a struct whose fields are all filtered out is
dropped as well, while a struct that still has a field is kept.

```yaml
match_tags:
  - json
exclude_empty_structs: true
```

```go
type Included struct {
	Name string `json:"name"`
}

type Internal struct {
	secret string
	Cache  map[string]int
}
```

```ts
export interface Included {
  name: string;
}
```

Structs that extend another type are kept even when they have no fields of
their own.

```yaml
exclude_empty_structs: true
```

```go
type Base struct {
	ID string `json:"id"`
}

type WithBase struct {
	Base `tstype:",extends"`
}
```

```ts
export interface Base {
  id: string;
}
export interface WithBase extends Base {
}
```
