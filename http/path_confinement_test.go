package fbhttp

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/filebrowser/filebrowser/v2/diskcache"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
)

// CodeQL reports the filesystem calls below individually because it cannot
// infer the confinement provided by ScopedFs. Exercise the HTTP entry points as
// one security boundary: an authenticated request must never dereference a
// symlink whose target is outside the user's scope.
func TestAuthenticatedHandlersRejectEscapingSymlink(t *testing.T) {
	root := t.TempDir()
	scope := filepath.Join(root, "scope")
	outside := filepath.Join(root, "outside")
	for _, dir := range []string{scope, outside} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	secretPath := filepath.Join(outside, "secret.srt")
	if err := os.WriteFile(secretPath, []byte("out-of-scope"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(scope, "escape")); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	key := []byte("test-signing-key")
	perm := users.Permissions{
		Create: true, Modify: true, Delete: true, Download: true,
	}
	st := scopedUserStorage(t, scope, perm, key)
	token := signToken(t, perm, key)

	tests := []struct {
		name    string
		method  string
		path    string
		body    string
		handler handleFunc
	}{
		{name: "resource read", method: http.MethodGet, path: "/escape/secret.srt", handler: resourceGetHandler},
		{name: "raw read", method: http.MethodGet, path: "/escape/secret.srt", handler: rawHandler},
		{name: "subtitle read", method: http.MethodGet, path: "/escape/secret.srt", handler: subtitleHandler},
		{name: "resource delete", method: http.MethodDelete, path: "/escape/secret.srt", handler: resourceDeleteHandler(diskcache.NewNoOp())},
		{name: "resource create", method: http.MethodPost, path: "/escape/new.txt", body: "new", handler: resourcePostHandler(diskcache.NewNoOp())},
		{name: "resource update", method: http.MethodPut, path: "/escape/secret.srt", body: "changed", handler: resourcePutHandler},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("X-Auth", token)
			rec := httptest.NewRecorder()
			handle(tt.handler, "", st, &settings.Server{}).ServeHTTP(rec, req)

			if rec.Code != http.StatusForbidden {
				t.Fatalf("got status %d body=%q, want 403", rec.Code, rec.Body.String())
			}
			content, err := os.ReadFile(secretPath)
			if err != nil {
				t.Fatalf("out-of-scope file was removed: %v", err)
			}
			if string(content) != "out-of-scope" {
				t.Fatalf("out-of-scope file was modified: %q", content)
			}
			if _, err := os.Stat(filepath.Join(outside, "new.txt")); !os.IsNotExist(err) {
				t.Fatalf("out-of-scope file was created: %v", err)
			}
		})
	}
}
