```go
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
	*Base          `tstype:",extends,required"`
	Base2[T]       `tstype:",extends"`
	*Base3[X, T]   `tstype:",extends"`
	OtherWithBase  Base                             `                                          json:"otherWithBase"`
	OtherWithBase2 Base2[X]                         `                                          json:"otherWithBase2"`
	OtherValue     string                           `                                          json:"otherValue"`
	Author         bookapp.AuthorWithInheritance[T] `tstype:"bookapp.AuthorWithInheritance<T>" json:"author"`
	bookapp.Book   `tstype:",extends"`
	TextBook       *bookapp.TextBook[T] `tstype:",extends,required"`
}
```

```ts
export interface Base {
  name: string;
}
export interface Base2<T extends string | number /* int */> {
  id: T;
}
export interface Base3<T extends string, X extends number /* int */> {
  class: T;
  level: X;
}
export interface Other<T extends number /* int */, X extends string> extends Base, Base2<T>, Partial<Base3<X, T>>, bookapp.Book, bookapp.TextBook<T> {
  otherWithBase: Base;
  otherWithBase2: Base2<X>;
  otherValue: string;
  author: bookapp.AuthorWithInheritance<T>;
}
```