// Package level
// Second line of package level comment.
package abstract

// DROPPED: Floating comment at the top

// Comment belonging to Foo
type Foo string
type FooInt64 int64

// Comment for the const group declaration
const (
	ConstNumberValue = 123 // Line comment behind field with value 123
	// Individual comment for field ConstStringValue
	ConstStringValue     = "abc"
	ConstFooValue    Foo = "foo_const_value"
) // DROPPED: Line comment after grouped const

const Alice = "Alice"

/*
 DROPPED: Floating multiline comment somewhere in the middle
 Line two
*/

/*
Multiline comment for StructBar
Some more text
*/
type StructBar struct {
	// Comment for field Field of type Foo
	Field                 Foo   `json:"field"` // Line Comment for field Field of type Foo
	FieldWithWeirdJSONTag int64 `json:"weird"`

	FieldThatShouldBeOptional    *string `json:"field_that_should_be_optional"`
	FieldThatShouldNotBeOptional *string `json:"field_that_should_not_be_optional" tstype:",required"`
}

// DROPPED: Floating comment at the end
