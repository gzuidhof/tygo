package tygo

import (
	"fmt"
	"os/exec"
	"strings"
)

func runFormatter(formatter *FormatterConfig, outputPaths []string) error {
	if formatter == nil {
		return nil
	}
	if formatter.Command == "" {
		return fmt.Errorf("formatter command must not be empty")
	}

	args := append(append([]string{}, formatter.Args...), outputPaths...)
	cmd := exec.Command(formatter.Command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			return fmt.Errorf("formatter %q failed: %w", formatter.Command, err)
		}
		return fmt.Errorf("formatter %q failed: %w: %s", formatter.Command, err, message)
	}
	return nil
}
