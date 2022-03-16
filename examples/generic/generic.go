package generic

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
