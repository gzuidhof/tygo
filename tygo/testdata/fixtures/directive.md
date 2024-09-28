This fixture is here to reproduce Github issue #26.

```go
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

const A = "a"
const B = "a"
const C = "c"
```

```ts
/**
 * Comment above a directive
 */
export const SomeValue = 3;
/**
 * Empty Comment
 */
export const AnotherValue = 4;

export const DirectiveOnly = 5;
/**
 * RepoIndexerType specifies the repository indexer type
 */
export type RepoIndexerType = number /* int */;
/**
 * RepoIndexerTypeCode code indexer
 */
export const RepoIndexerTypeCode: RepoIndexerType = 0; // 0
/**
 * RepoIndexerTypeStats repository stats indexer
 */
export const RepoIndexerTypeStats: RepoIndexerType = 1; // 1
export const A = "a";
export const B = "a";
export const C = "c";
```