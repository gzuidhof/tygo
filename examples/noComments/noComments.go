// noComments is a package-level comment. By default, it is preserved.
// with preserveComments configured to "none", it won't be preserved.
package noComments

import "github.com/google/uuid"

// This is a block comment in the package body. By default, it is preserved.
// With preserveComments configured to "none", it won't be preserved.

// Type comments are preserved by default or with "types". It won't be preserved with "none"
type UserRole = string

const (
	// Const comments are preserved by default or with "types". It won't be preserved with "none"
	UserRoleDefault UserRole = "viewer"
	UserRoleEditor  UserRole = "editor" // Line comments are preserved by default. With preserveComments configured to "none", it won't be preserved.
)

type User struct {
	// Struct field comments are preserved by default or with "types". It won't be preserved with "none"
	ID uuid.NullUUID `json:"id" tstype:"string | null"`
}
