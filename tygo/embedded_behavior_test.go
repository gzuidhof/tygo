package tygo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEmbeddedFieldBehavior_ExtendsRequired tests embedded field behavior with extends,required
func TestEmbeddedFieldBehavior_ExtendsRequired(t *testing.T) {
	t.Parallel()

	goCode := `
type Base struct {
	Name string ` + "`" + `json:"name"` + "`" + `
}

type Derived struct {
	Base
	Age int ` + "`" + `json:"age"` + "`" + `
}
`

	config := PackageConfig{
		EmbeddedFieldBehavior: "extends,required",
	}

	tsCode, err := ConvertGoToTypescript(goCode, config)
	require.NoError(t, err)

	expected := `export interface Base {
  name: string;
}
export interface Derived extends Base {
  age: number /* int */;
}
`
	assert.Equal(t, expected, tsCode)
}

// TestEmbeddedFieldBehavior_Extends tests embedded field behavior with extends (optional)
func TestEmbeddedFieldBehavior_Extends(t *testing.T) {
	t.Parallel()

	goCode := `
type Base struct {
	Name string ` + "`" + `json:"name"` + "`" + `
}

type Derived struct {
	*Base
	Age int ` + "`" + `json:"age"` + "`" + `
}
`

	config := PackageConfig{
		EmbeddedFieldBehavior: "extends",
	}

	tsCode, err := ConvertGoToTypescript(goCode, config)
	require.NoError(t, err)

	expected := `export interface Base {
  name: string;
}
export interface Derived extends Partial<Base> {
  age: number /* int */;
}
`
	assert.Equal(t, expected, tsCode)
}

// TestEmbeddedFieldBehavior_ExtendsRequiredReadonly tests embedded field behavior with extends,required
func TestEmbeddedFieldBehavior_ExtendsRequiredReadonly(t *testing.T) {
	t.Parallel()

	goCode := `
type Base struct {
	Name string ` + "`" + `json:"name"` + "`" + `
}

type Derived struct {
	Base
	Age int ` + "`" + `json:"age"` + "`" + `
}
`

	config := PackageConfig{
		EmbeddedFieldBehavior: "extends,required",
	}

	tsCode, err := ConvertGoToTypescript(goCode, config)
	require.NoError(t, err)

	expected := `export interface Base {
  name: string;
}
export interface Derived extends Base {
  age: number /* int */;
}
`
	assert.Equal(t, expected, tsCode)
}

// TestEmbeddedFieldBehavior_Empty tests embedded field behavior with empty string (default)
func TestEmbeddedFieldBehavior_Empty(t *testing.T) {
	t.Parallel()

	goCode := `
type Base struct {
	Name string ` + "`" + `json:"name"` + "`" + `
}

type Derived struct {
	Base ` + "`" + `json:"base"` + "`" + `
	Age int ` + "`" + `json:"age"` + "`" + `
}
`

	config := PackageConfig{
		EmbeddedFieldBehavior: "",
	}

	tsCode, err := ConvertGoToTypescript(goCode, config)
	require.NoError(t, err)

	expected := `export interface Base {
  name: string;
}
export interface Derived {
  base: Base;
  age: number /* int */;
}
`
	assert.Equal(t, expected, tsCode)
}

// TestEmbeddedFieldBehavior_OverrideWithExplicitTag tests that explicit tags override default behavior
func TestEmbeddedFieldBehavior_OverrideWithExplicitTag(t *testing.T) {
	t.Parallel()

	goCode := `
type Base struct {
	Name string ` + "`" + `json:"name"` + "`" + `
}

type Derived struct {
	Base ` + "`" + `tstype:"-"` + "`" + `
	Age  int ` + "`" + `json:"age"` + "`" + `
}
`

	config := PackageConfig{
		EmbeddedFieldBehavior: "extends,required",
	}

	tsCode, err := ConvertGoToTypescript(goCode, config)
	require.NoError(t, err)

	expected := `export interface Base {
  name: string;
}
export interface Derived {
  age: number /* int */;
}
`
	assert.Equal(t, expected, tsCode)
}

// TestEmbeddedFieldBehavior_MultipleEmbeddedFields tests multiple embedded fields with default behavior
func TestEmbeddedFieldBehavior_MultipleEmbeddedFields(t *testing.T) {
	t.Parallel()

	goCode := `
type Base1 struct {
	Name string ` + "`" + `json:"name"` + "`" + `
}

type Base2 struct {
	Age int ` + "`" + `json:"age"` + "`" + `
}

type Derived struct {
	Base1
	Base2
	Value string ` + "`" + `json:"value"` + "`" + `
}
`

	config := PackageConfig{
		EmbeddedFieldBehavior: "extends,required",
	}

	tsCode, err := ConvertGoToTypescript(goCode, config)
	require.NoError(t, err)

	expected := `export interface Base1 {
  name: string;
}
export interface Base2 {
  age: number /* int */;
}
export interface Derived extends Base1, Base2 {
  value: string;
}
`
	assert.Equal(t, expected, tsCode)
}

// TestEmbeddedFieldBehavior_GenericEmbedded tests generic embedded field with default behavior
func TestEmbeddedFieldBehavior_GenericEmbedded(t *testing.T) {
	t.Parallel()

	goCode := `
type Base[T string | int] struct {
	ID T ` + "`" + `json:"id"` + "`" + `
}

type Derived struct {
	Base[int]
	Name string ` + "`" + `json:"name"` + "`" + `
}
`

	config := PackageConfig{
		EmbeddedFieldBehavior: "extends,required",
	}

	tsCode, err := ConvertGoToTypescript(goCode, config)
	require.NoError(t, err)

	expected := `export interface Base<T extends string | number /* int */> {
  id: T;
}
export interface Derived extends Base<number /* int */> {
  name: string;
}
`
	assert.Equal(t, expected, tsCode)
}
