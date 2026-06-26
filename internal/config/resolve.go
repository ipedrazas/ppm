// Package config resolves where the memory root lives.
package config

import (
	"os"
	"path/filepath"
)

// EnvRoot is the environment variable consulted after the --root flag.
const EnvRoot = "PPM_MEMORY_ROOT"

// ResolveRoot picks the memory root directory using, in order:
//  1. the --root flag (flagRoot), if non-empty
//  2. the PPM_MEMORY_ROOT environment variable
//  3. the nearest ancestor of the cwd containing an existing memory/ dir
//  4. the default ./memory
//
// The returned path is absolute.
func ResolveRoot(flagRoot string) (string, error) {
	if flagRoot != "" {
		return filepath.Abs(flagRoot)
	}
	if env := os.Getenv(EnvRoot); env != "" {
		return filepath.Abs(env)
	}
	if found, ok := walkUp(); ok {
		return found, nil
	}
	return filepath.Abs("memory")
}

// walkUp searches the cwd and its ancestors for an existing memory/ directory.
func walkUp() (string, bool) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", false
	}
	for dir := cwd; ; {
		candidate := filepath.Join(dir, "memory")
		if fi, err := os.Stat(candidate); err == nil && fi.IsDir() {
			return candidate, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}
