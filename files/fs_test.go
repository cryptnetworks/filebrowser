package files

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
)

// TestNewFs verifies that NewFs picks the right implementation and that the
// follow-external-symlinks toggle flips whether a symlink pointing outside the
// scope is honored.
func TestNewFs(t *testing.T) {
	base := t.TempDir()
	scope := filepath.Join(base, "srv")
	outside := filepath.Join(base, "outside")
	for _, d := range []string{scope, outside} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("secret"), 0o644); err != nil {
		t.Fatal(err)
	}
	// A symlink lexically inside the scope whose target resolves outside it.
	if err := os.Symlink(outside, filepath.Join(scope, "escape")); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	t.Run("disabled returns a ScopedFs that rejects the escaping symlink", func(t *testing.T) {
		fs := NewFs(afero.NewOsFs(), scope, false)
		if _, ok := fs.(*ScopedFs); !ok {
			t.Fatalf("expected *ScopedFs, got %T", fs)
		}
		if _, err := fs.Stat("/escape"); !os.IsPermission(err) {
			t.Fatalf("expected stat of escaping symlink to be rejected, got %v", err)
		}
	})

	t.Run("enabled returns a BasePathFs that follows the escaping symlink", func(t *testing.T) {
		fs := NewFs(afero.NewOsFs(), scope, true)
		if _, ok := fs.(*afero.BasePathFs); !ok {
			t.Fatalf("expected *afero.BasePathFs, got %T", fs)
		}
		if _, err := fs.Stat("/escape"); err != nil {
			t.Fatalf("expected escaping symlink to be followed, got %v", err)
		}
		b, err := afero.ReadFile(fs, "/escape/secret.txt")
		if err != nil {
			t.Fatalf("expected to read through escaping symlink, got %v", err)
		}
		if string(b) != "secret" {
			t.Fatalf("got %q, want %q", b, "secret")
		}

		// The link must also appear in a directory listing (the symptom in #5998).
		entries, err := afero.ReadDir(fs, "/")
		if err != nil {
			t.Fatal(err)
		}
		var found bool
		for _, e := range entries {
			if e.Name() == "escape" {
				found = true
			}
		}
		if !found {
			t.Fatal("expected escaping symlink to appear in the listing")
		}
	})
}

func TestScopedFsRejectsLexicalSiblingEscape(t *testing.T) {
	base := t.TempDir()
	scope := filepath.Join(base, "scope")
	sibling := filepath.Join(base, "scope-sibling")
	for _, dir := range []string{scope, sibling} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(sibling, "secret.txt"), []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}

	fs := NewScopedFs(afero.NewOsFs(), scope)
	for _, name := range []string{"../scope-sibling/secret.txt", "/../scope-sibling/secret.txt"} {
		t.Run(name, func(t *testing.T) {
			if _, err := fs.Open(name); !os.IsPermission(err) {
				t.Fatalf("open %q returned %v, want permission error", name, err)
			}
		})
	}
}

func TestScopedFsTraversalCannotReachSibling(t *testing.T) {
	base := t.TempDir()
	scope := filepath.Join(base, "scope")
	outside := filepath.Join(base, "outside")
	for _, dir := range []string{scope, outside} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	secret := filepath.Join(outside, "secret.txt")
	if err := os.WriteFile(secret, []byte("outside"), 0o600); err != nil {
		t.Fatal(err)
	}
	inside := filepath.Join(scope, "inside.txt")
	if err := os.WriteFile(inside, []byte("inside"), 0o600); err != nil {
		t.Fatal(err)
	}

	fs := NewScopedFs(afero.NewOsFs(), scope)
	tests := []struct {
		name string
		run  func() error
	}{
		{name: "parent read", run: func() error { _, err := fs.Open("../outside/secret.txt"); return err }},
		{name: "repeated parent read", run: func() error { _, err := fs.Open("../../outside/secret.txt"); return err }},
		{name: "absolute host path read", run: func() error { _, err := fs.Open(secret); return err }},
		{name: "encoded parent read", run: func() error { _, err := fs.Open("%2e%2e/outside/secret.txt"); return err }},
		{name: "double encoded parent read", run: func() error { _, err := fs.Open("%252e%252e/outside/secret.txt"); return err }},
		{name: "parent create", run: func() error { _, err := fs.Create("../outside/new.txt"); return err }},
		{name: "parent mkdir", run: func() error { return fs.MkdirAll("../outside/newdir", 0o755) }},
		{name: "parent rename", run: func() error { return fs.Rename("/inside.txt", "../outside/moved.txt") }},
		{name: "parent delete", run: func() error { return fs.RemoveAll("../outside/secret.txt") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = tt.run()
			content, err := os.ReadFile(secret)
			if err != nil || string(content) != "outside" {
				t.Fatalf("out-of-scope secret changed: content=%q err=%v", content, err)
			}
			for _, name := range []string{"new.txt", "newdir", "moved.txt"} {
				if _, err := os.Stat(filepath.Join(outside, name)); !os.IsNotExist(err) {
					t.Fatalf("out-of-scope target %q was created: %v", name, err)
				}
			}
			if content, err := os.ReadFile(inside); err != nil || string(content) != "inside" {
				t.Fatalf("in-scope source changed: content=%q err=%v", content, err)
			}
		})
	}
}

// TestBasePath verifies BasePath extracts the underlying *afero.BasePathFs from
// either filesystem NewFs may return, so User.FullPath keeps working.
func TestBasePath(t *testing.T) {
	root := t.TempDir()
	osFs := afero.NewOsFs()

	for _, tc := range []struct {
		name           string
		followExternal bool
	}{
		{"ScopedFs", false},
		{"BasePathFs", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fs := NewFs(osFs, root, tc.followExternal)
			base := BasePath(fs)
			if base == nil {
				t.Fatalf("expected non-nil base for %T", fs)
			}
			got := afero.FullBaseFsPath(base, "/x")
			want := filepath.Join(root, "x")
			if got != want {
				t.Fatalf("FullBaseFsPath: got %q, want %q", got, want)
			}
		})
	}

	if got := BasePath(osFs); got != nil {
		t.Fatalf("expected nil base for a plain OsFs, got %v", got)
	}
}

// TestFileInfoRealPathUsesBasePathFsRealPath mirrors
// TestFileInfoRealPathUsesScopedFsRealPath for the follow-external-symlinks case,
// where the user filesystem is a bare BasePathFs.
func TestFileInfoRealPathUsesBasePathFsRealPath(t *testing.T) {
	root := t.TempDir()
	file := &FileInfo{
		Fs:   NewFs(afero.NewOsFs(), root, true),
		Path: "/root/downloads",
	}

	got := file.RealPath()
	want := filepath.Join(root, "root", "downloads")
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
