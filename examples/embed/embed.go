package embed

import bookapp "github.com/vincenzomerolla/tygo/examples/bookstore"

// TokenType Built-in type alias
type TokenType string

type StructEmbed struct {
	Base             `json:",inline" tstype:",extends"` // embed struct with `tstype:"extends"`
	TokenType        `json:"tokenType"`                 // built-in type field without `tstype:"extends"`
	Reference        `json:"reference"`                 // embed struct without `tstype:"extends"`
	OtherReference   Reference                          `json:"other_reference"`
	Bar              string                             `json:"bar"`
	bookapp.Book     `json:"book"`                      // embed external struct without `tstype:"extends"`
	*bookapp.Chapter `json:"chapter"`                   // embed external struct pointer without `tstype:"extends"`
}
