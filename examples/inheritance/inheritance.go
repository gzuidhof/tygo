package inheritance

import bookapp "github.com/gzuidhof/tygo/examples/bookstore"

type Base struct {
	Name string `json:"name"`
}

type Base2[T string | int] struct {
	ID T `json:"id"`
}

type Base3[T string, X int] struct {
	Class T `json:"class"`
	Level X `json:"level"`
}

type Other[T int, X string] struct {
	Base                `tstype:",extends"`
	Base2[T]            `tstype:",extends"`
	Base3[X, T]         `tstype:",extends"`
	OtherWithBase       Base                             `                                          json:"otherWithBase"`
	OtherWithBase2      Base2[X]                         `                                          json:"otherWithBase2"`
	OtherValue          string                           `                                          json:"otherValue"`
	Author              bookapp.AuthorWithInheritance[T] `tstype:"bookapp.AuthorWithInheritance<T>" json:"author"`
	bookapp.Book        `tstype:",extends"`
	bookapp.TextBook[T] `tstype:",extends"`
}
