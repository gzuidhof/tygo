Basic enum generation with enum_style: "enum"

```yaml
enum_style: "enum"
```

```go
type UserRole = string
const (
    UserRoleDefault UserRole = "viewer"
    UserRoleEditor  UserRole = "editor"
)
```

```ts
export enum UserRole {
  Default = "viewer",
  Editor = "editor",
}
```

Enum with comments

```yaml
enum_style: "enum"
```

```go
// User role enumeration
type Status = string
const (
    // Default status for new users
    StatusActive Status = "active"
    StatusInactive Status = "inactive" // User is temporarily disabled
)
```

```ts
/**
 * User role enumeration
 */
export enum Status {
  /**
   * Default status for new users
   */
  Active = "active",
  Inactive = "inactive", // User is temporarily disabled
}
```

Numeric enum with iota

```yaml
enum_style: "enum"
```

```go
type Priority int
const (
    PriorityLow Priority = iota
    PriorityMedium
    PriorityHigh
)
```

```ts
export enum Priority {
  Low = 0,
  Medium,
  High,
}
```

Mixed const block (partial enum)

```yaml
enum_style: "enum"
```

```go
type UserRole = string
const (
    UserRoleAdmin UserRole = "admin"
    UserRoleGuest UserRole = "guest"
    MaxRetries = 5
    DefaultTimeout = 30
)
```

```ts
export enum UserRole {
  Admin = "admin",
  Guest = "guest",
}
export const MaxRetries = 5;
export const DefaultTimeout = 30;
```

Default behavior (enum_style: "const")

```yaml
enum_style: "const"
```

```go
type UserRole = string
const (
    UserRoleDefault UserRole = "viewer"
    UserRoleEditor  UserRole = "editor"
)
```

```ts
export type UserRole = string;
export const UserRoleDefault: UserRole = "viewer";
export const UserRoleEditor: UserRole = "editor";
```

No enum_style configured (defaults to "const")

```go
type UserRole = string
const (
    UserRoleDefault UserRole = "viewer"
    UserRoleEditor  UserRole = "editor"
)
```

```ts
export type UserRole = string;
export const UserRoleDefault: UserRole = "viewer";
export const UserRoleEditor: UserRole = "editor";
```

Basic enum generation with enum_style: "union"

```yaml
enum_style: "union"
```

```go
type UserRole = string
const (
    UserRoleDefault UserRole = "viewer"
    UserRoleEditor  UserRole = "editor"
)
```

```ts
export type UserRole = "viewer" | "editor";
```

Union enum with comments

```yaml
enum_style: "union"
```

```go
// User role enumeration
type Status = string
const (
    // Default status for new users
    StatusActive Status = "active"
    StatusInactive Status = "inactive" // User is temporarily disabled
)
```

```ts
/**
 * User role enumeration
 */
export type Status = "active" | "inactive";
```

Numeric union with iota

```yaml
enum_style: "union"
```

```go
type Priority int
const (
    PriorityLow Priority = iota
    PriorityMedium
    PriorityHigh
)
```

```ts
export type Priority = 0 | 1 | 2;
```

Mixed const block (partial union)

```yaml
enum_style: "union"
```

```go
type UserRole = string
const (
    UserRoleAdmin UserRole = "admin"
    UserRoleGuest UserRole = "guest"
    MaxRetries = 5
    DefaultTimeout = 30
)
```

```ts
export type UserRole = "admin" | "guest";
export const MaxRetries = 5;
export const DefaultTimeout = 30;
```
