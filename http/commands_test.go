package fbhttp

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/filebrowser/filebrowser/v2/users"
	"github.com/spf13/afero"
)

func TestCommandWorkingDirectoryRejectsEscapingSymlink(t *testing.T) {
	root := t.TempDir()
	scope := filepath.Join(root, "scope")
	outside := filepath.Join(root, "outside")
	for _, dir := range []string{scope, outside} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(outside, filepath.Join(scope, "escape")); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	user := &users.User{Fs: files.NewScopedFs(afero.NewOsFs(), scope)}
	if _, err := commandWorkingDirectory(user, "/escape"); !os.IsPermission(err) {
		t.Fatalf("escaping working directory returned %v, want permission error", err)
	}

	got, err := commandWorkingDirectory(user, "/")
	if err != nil {
		t.Fatal(err)
	}
	if got != scope {
		t.Fatalf("got working directory %q, want %q", got, scope)
	}
}
