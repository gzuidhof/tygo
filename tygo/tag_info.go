package tygo

import (
	"go/ast"

	"github.com/fatih/structtag"
)

// FieldTagInfo holds parsed and cached tag information for a field
type FieldTagInfo struct {
	// Parsed tags
	jsonTag   *structtag.Tag
	yamlTag   *structtag.Tag
	tstypeTag *structtag.Tag

	// Derived properties
	name       string
	optional   bool
	required   bool
	readonly   bool
	skip       bool // true if field should be omitted
	isExtend   bool // true if field should be treated as extends
	isEmbedded bool // true if field is anonymous/embedded
}

// TagCache caches parsed tag information per field
type TagCache struct {
	cache map[*ast.Field]*FieldTagInfo
}

// NewTagCache creates a new tag cache
func NewTagCache() *TagCache {
	return &TagCache{
		cache: make(map[*ast.Field]*FieldTagInfo),
	}
}

// Get returns cached tag info for a field, or nil if not cached
func (c *TagCache) Get(f *ast.Field) *FieldTagInfo {
	return c.cache[f]
}

// Set stores tag info for a field in the cache
func (c *TagCache) Set(f *ast.Field, info *FieldTagInfo) {
	c.cache[f] = info
}

// isAnonymousField checks if field is embedded/anonymous
func isAnonymousField(f *ast.Field) bool {
	return len(f.Names) == 0
}
