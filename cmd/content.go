package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// contentFlags backs the --content (primary) / --file (fallback) pair. The
// backing fields are shared across commands because exactly one command runs per
// process invocation.
type contentFlags struct {
	content string
	file    string
}

var contentInput contentFlags

// flagName backs the optional --name slug on collection "add" commands.
var flagName string

// registerContent attaches --content and --file to a command.
func registerContent(cmd *cobra.Command) {
	cmd.Flags().StringVar(&contentInput.content, "content", "", "entry content (primary input)")
	cmd.Flags().StringVar(&contentInput.file, "file", "", "read content from a file (fallback)")
}

// registerName attaches an optional --name slug.
func registerName(cmd *cobra.Command) {
	cmd.Flags().StringVar(&flagName, "name", "", "explicit entry slug (else auto-generated)")
}

// resolve returns the content, requiring exactly one of --content/--file.
func (c *contentFlags) resolve() (string, error) {
	if c.content != "" && c.file != "" {
		return "", fmt.Errorf("provide --content or --file, not both")
	}
	if c.file != "" {
		return readContentFile(c.file)
	}
	if c.content == "" {
		return "", fmt.Errorf("content required: pass --content or --file")
	}
	return c.content, nil
}

// resolveOptional is like resolve but allows neither source (returns "").
func (c *contentFlags) resolveOptional() (string, error) {
	if c.content != "" && c.file != "" {
		return "", fmt.Errorf("provide --content or --file, not both")
	}
	if c.file != "" {
		return readContentFile(c.file)
	}
	return c.content, nil
}

func readContentFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
