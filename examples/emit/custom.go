package emit

// CustomMarshalled illustrates getting tygo to emit literal text/
// This solves the problem of a struct field being marshalled into a tuple.
//
//tygo:emit export type StructAsTuple=[a:number, b:number, c:string]
type CustomMarshalled struct {
	Content []StructAsTuple `json:"content"`
}
