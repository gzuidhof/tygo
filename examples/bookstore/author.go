package bookapp

type AuthorBookListing struct {
	AuthorName   string `json:"author_name"`
	WrittenBooks []Book `json:"written_books"`
}
