// This example is here to reproduce Github issue #26
// Directive comments should not be output.
//
// See https://github.com/gzuidhof/tygo/issues/26
package directive

// Comment above a directive
//
//go:foo
//go:bar
const SomeValue = 3 //comment:test

// Empty Comment
const AnotherValue = 4 //

//go:something
const DirectiveOnly = 5

// RepoIndexerType specifies the repository indexer type
type RepoIndexerType int //revive:disable-line:exported

const (
	// RepoIndexerTypeCode code indexer
	RepoIndexerTypeCode RepoIndexerType = iota // 0
	// RepoIndexerTypeStats repository stats indexer
	RepoIndexerTypeStats // 1
)
