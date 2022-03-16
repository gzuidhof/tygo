package generic

// Comment for UnionType
type UnionType[T any] interface {
	// Comment for fields are possible
	uint64 | string | *bool // comment after

	// Comment for a method
	SomeMethod() string
	AnotherMethod(T) *T
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

type Foo[V any, PT *V, Unused string] struct {
	Val V
	// Comment for ptr field
	Ptr PT // ptr line comment
}

type ABCD[A, B, C string, D int64 | bool] struct {
	A A `json:"a"`
	B B `json:"b"`
	C C `json:"c"`
	D D `json:"d"`
}

// Should not be output as it's a function
func (f Foo[int, Derived, string]) DoSomething() {
	panic("something")
}
