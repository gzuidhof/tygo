package cmd

import (
	"fmt"
	"runtime"
)

// Field injected by goreleaser
var (
	version    = "<unknown>"
	commitDate = "date unknown"
	commit     = ""
)

func Version() string {
	return version
}

func CommitDate() string {
	return commitDate
}

func Commit() string {
	return commit
}

func Target() string {
	return runtime.GOOS
}

func FullVersion() string {
	return fmt.Sprintf("%s %s/%s %s (%s) %s",
		version, runtime.GOOS, runtime.GOARCH, runtime.Version(), commitDate, commit)
}
