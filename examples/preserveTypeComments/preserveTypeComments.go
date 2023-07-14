// preservetypecomments is a package-level comment. By default, it is preserved.
// with preserveComments configured to "types", it won't be preserved.
package preservetypecomments

import "github.com/google/uuid"

// This is a block comment in the package body. By default, it is preserved.
// With preserveComments configured to "types", it won't be preserved.

// Type comments are kept, unless configured to "none"
type UserRole = string

const (
	// UserRoleDefault is "viewer"
	UserRoleDefault UserRole = "viewer"
	// UserRoleEditor can edit other users
	UserRoleEditor UserRole = "editor" // Line comments are preserved by default. With preserveComments configured to "types", it won't be preserved.
)

type User struct {
	// Struct field comments are preserved unless configured to "none"
	ID uuid.NullUUID `json:"id" tstype:"string | null"`
}
