package bookapp

type AuthorBookListing struct {
	AuthorName   string `json:"author_name"`
	WrittenBooks []Book `json:"written_books"`
}

type AuthorWithInheritance[T int] struct {
	ID T `json:"id"`
}
