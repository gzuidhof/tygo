Only fields carrying one of the configured tags are emitted.

```yaml
match_tags:
  - json
```

```go
type User struct {
	Username string `json:"username"`
	Password string
}
```

```ts
export interface User {
  username: string;
}
```

The tag used for matching does not have to be the one that decides the field
name, and more than one tag can be listed.

```yaml
match_tags:
  - db
```

```go
type Account struct {
	ID    int    `json:"id" db:"id"`
	Name  string `json:"name"`
	Token string `db:"token"`
}
```

```ts
export interface Account {
  id: number /* int */;
  Token: string;
}
```

Without any `match_tags` configured every field is emitted as before.

```go
type Plain struct {
	Username string `json:"username"`
	Password string
}
```

```ts
export interface Plain {
  username: string;
  Password: string;
}
```
