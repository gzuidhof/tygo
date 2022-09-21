# 0.2.4
* Add support for `fallback_type`, which allows one to specify the type that should be used for unknown types (defaults to `any`).

# 0.2.3
* Add support for `include_files` in the config, which allows one to specify the only files that should be included.

# 0.2.2
* Add support for `readonly` fields.

# 0.2.1
* Add support for single `IndexExpr` in type alias, such as `type X = Y[string]`.

# 0.2.0

* Support for generic types for Golang 1.18 and beyond.

# 0.1.2

* You can now mark pointer type fields as required by adding `,required` to the struct field tag. 

# 0.1.1

* You can now exclude fields with `tstype:"-"`.

# 0.1.0

First public release.
