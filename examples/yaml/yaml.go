package yaml

type YAMLTest struct {
	ID string `yaml:"id"`

	Preferences map[string]struct {
		Foo uint32 `yaml:"foo"`
	} `yaml:"prefs"`

	MaybeFieldWithStar *string `yaml:"address"`
	Nickname           string  `yaml:"nickname,omitempty"`

	unexported bool // Unexported fields won't be in the output
}
