package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveRootFlagWins(t *testing.T) {
	t.Setenv(EnvRoot, "/env/should/lose")
	got, err := ResolveRoot("flag/root")
	if err != nil {
		t.Fatal(err)
	}
	abs, _ := filepath.Abs("flag/root")
	if got != abs {
		t.Errorf("got %q, want %q", got, abs)
	}
}

func TestResolveRootEnv(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(EnvRoot, dir)
	got, err := ResolveRoot("")
	if err != nil {
		t.Fatal(err)
	}
	if got != dir {
		t.Errorf("got %q, want %q", got, dir)
	}
}

func TestResolveRootWalkUp(t *testing.T) {
	root := t.TempDir()
	mem := filepath.Join(root, "memory")
	if err := os.MkdirAll(mem, 0o755); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv(EnvRoot, "")
	chdir(t, nested)

	got, err := ResolveRoot("")
	if err != nil {
		t.Fatal(err)
	}
	// macOS /tmp is a symlink to /private/tmp; compare resolved paths.
	gotEval, _ := filepath.EvalSymlinks(got)
	wantEval, _ := filepath.EvalSymlinks(mem)
	if gotEval != wantEval {
		t.Errorf("got %q, want %q", gotEval, wantEval)
	}
}

func TestResolveRootDefault(t *testing.T) {
	dir := t.TempDir() // no memory/ here or above within the temp dir
	t.Setenv(EnvRoot, "")
	chdir(t, dir)

	got, err := ResolveRoot("")
	if err != nil {
		t.Fatal(err)
	}
	// Walk-up may find a memory/ dir in a real ancestor of the temp dir; only
	// assert the default when none was found.
	wantDefault, _ := filepath.Abs("memory")
	if got != wantDefault && filepath.Base(got) != "memory" {
		t.Errorf("unexpected root %q", got)
	}
}

func chdir(t *testing.T, dir string) {
	t.Helper()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(prev) })
}
