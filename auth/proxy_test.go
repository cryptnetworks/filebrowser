package auth

import (
	"net/http"
	"os"
	"strings"
	"testing"

	fberrors "github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
)

type mockUserStore struct {
	users map[string]*users.User
}

func (m *mockUserStore) Get(_ string, _ bool, id interface{}) (*users.User, error) {
	if v, ok := id.(string); ok {
		if u, ok := m.users[v]; ok {
			return u, nil
		}
	}
	return nil, fberrors.ErrNotExist
}

func (m *mockUserStore) GetByScope(scope string) (*users.User, error) {
	for _, u := range m.users {
		if strings.EqualFold(u.Scope, scope) {
			return u, nil
		}
	}
	return nil, fberrors.ErrNotExist
}

func (m *mockUserStore) Gets(_ string, _ bool) ([]*users.User, error) { return nil, nil }
func (m *mockUserStore) Update(_ *users.User, _ ...string) error      { return nil }
func (m *mockUserStore) Save(user *users.User) error {
	m.users[user.Username] = user
	return nil
}

func (m *mockUserStore) SaveProvisioned(user *users.User, derivedScope bool) error {
	if derivedScope {
		if _, err := m.GetByScope(user.Scope); err == nil {
			return fberrors.ErrExist
		}
	}
	return m.Save(user)
}

func (m *mockUserStore) Delete(_ interface{}) error { return nil }
func (m *mockUserStore) LastUpdate(_ uint) int64    { return 0 }

func TestProxyAuthCreateUserRestrictsDefaults(t *testing.T) {
	t.Parallel()

	store := &mockUserStore{users: make(map[string]*users.User)}
	srv := &settings.Server{Root: t.TempDir()}

	s := &settings.Settings{
		Key:        []byte("key"),
		AuthMethod: MethodProxyAuth,
		Defaults: settings.UserDefaults{
			Perm: users.Permissions{
				Admin:    true,
				Execute:  true,
				Create:   true,
				Rename:   true,
				Modify:   true,
				Delete:   true,
				Share:    true,
				Download: true,
			},
			Commands: []string{"git", "ls", "cat", "id"},
		},
	}

	auth := ProxyAuth{Header: "X-Remote-User", TrustedCIDRs: []string{"127.0.0.1/32"}, AutoProvision: true}
	req, _ := http.NewRequest(http.MethodGet, "/", http.NoBody)
	req.Header.Set("X-Remote-User", "newproxyuser")
	req.RemoteAddr = "127.0.0.1:12345"

	user, err := auth.Auth(req, store, s, srv)
	if err != nil {
		t.Fatalf("Auth() error: %v", err)
	}

	if user.Perm.Admin {
		t.Error("auto-provisioned proxy user should not have Admin permission")
	}
	if user.Perm.Execute {
		t.Error("auto-provisioned proxy user should not have Execute permission")
	}
	if len(user.Commands) != 0 {
		t.Errorf("auto-provisioned proxy user should have empty Commands, got %v", user.Commands)
	}
	if !user.Perm.Create {
		t.Error("auto-provisioned proxy user should retain Create permission from defaults")
	}
}

func TestProxyAuthRejectsUntrustedOrAmbiguousIdentity(t *testing.T) {
	t.Parallel()

	store := &mockUserStore{users: make(map[string]*users.User)}
	srv := &settings.Server{Root: t.TempDir()}
	s := &settings.Settings{
		Key:        []byte("key"),
		AuthMethod: MethodProxyAuth,
	}

	auth := ProxyAuth{
		Header:       "X-Remote-User",
		TrustedCIDRs: []string{"127.0.0.1/32"},
	}

	t.Run("untrusted remote is rejected", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/", http.NoBody)
		req.RemoteAddr = "192.0.2.10:12345"
		req.Header.Set("X-Remote-User", "newproxyuser")

		if _, err := auth.Auth(req, store, s, srv); !os.IsPermission(err) {
			t.Fatalf("expected permission error for untrusted remote, got %v", err)
		}
		if _, ok := store.users["newproxyuser"]; ok {
			t.Fatal("untrusted request should not provision a user")
		}
	})

	t.Run("duplicated identity header is rejected", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/", http.NoBody)
		req.RemoteAddr = "127.0.0.1:12345"
		req.Header.Add("X-Remote-User", "alice")
		req.Header.Add("X-Remote-User", "bob")

		if _, err := auth.Auth(req, store, s, srv); !os.IsPermission(err) {
			t.Fatalf("expected permission error for duplicated header, got %v", err)
		}
	})
}

// With CreateUserDir enabled, two distinct proxy-authenticated users must each
// receive their own home directory instead of both inheriting the server root.
func TestProxyAuthCreateUserDirIsolatesScope(t *testing.T) {
	t.Parallel()

	store := &mockUserStore{users: make(map[string]*users.User)}
	srv := &settings.Server{Root: t.TempDir()}
	s := &settings.Settings{
		Key:              []byte("key"),
		AuthMethod:       MethodProxyAuth,
		CreateUserDir:    true,
		UserHomeBasePath: "/users",
		Defaults: settings.UserDefaults{
			Scope: ".",
			Perm:  users.Permissions{Create: true},
		},
	}

	auth := ProxyAuth{Header: "X-Remote-User", TrustedCIDRs: []string{"127.0.0.1/32"}, AutoProvision: true}
	provision := func(name string) *users.User {
		req, _ := http.NewRequest(http.MethodGet, "/", http.NoBody)
		req.Header.Set("X-Remote-User", name)
		req.RemoteAddr = "127.0.0.1:12345"
		u, err := auth.Auth(req, store, s, srv)
		if err != nil {
			t.Fatalf("Auth(%q) error: %v", name, err)
		}
		return u
	}

	alice := provision("alice")
	bob := provision("bob")

	if alice.Scope == "/" || bob.Scope == "/" {
		t.Fatalf("provisioned users inherited the server root: alice=%q bob=%q", alice.Scope, bob.Scope)
	}
	if alice.Scope == bob.Scope {
		t.Fatalf("distinct users must get distinct scopes, both got %q", alice.Scope)
	}
	if alice.Scope != "/users/alice" {
		t.Errorf("expected /users/alice, got %q", alice.Scope)
	}
}

func TestProxyProvisioningPolicy(t *testing.T) {
	for _, signup := range []bool{false, true} {
		for _, provision := range []bool{false, true} {
			store := &mockUserStore{users: map[string]*users.User{"existing": {Username: "existing"}}}
			srv := &settings.Server{Root: t.TempDir()}
			set := &settings.Settings{Signup: signup}
			a := ProxyAuth{Header: "X-Remote-User", TrustedCIDRs: []string{"127.0.0.1/32"}, AutoProvision: provision}
			for _, username := range []string{"existing", "new"} {
				req, _ := http.NewRequest(http.MethodGet, "/", http.NoBody)
				req.RemoteAddr = "127.0.0.1:1234"
				req.Header.Set(a.Header, username)
				_, err := a.Auth(req, store, set, srv)
				if username == "new" && !provision {
					if !os.IsPermission(err) || len(store.users) != 1 {
						t.Fatalf("signup=%t: disabled provisioning: err=%v users=%v", signup, err, store.users)
					}
					entries, readErr := os.ReadDir(srv.Root)
					if readErr != nil || len(entries) != 0 {
						t.Fatalf("denied provisioning modified root: %v, %v", entries, readErr)
					}
				} else if err != nil {
					t.Fatalf("signup=%t provision=%t user=%s: %v", signup, provision, username, err)
				}
			}
		}
	}
}

func TestProxyIdentityVariants(t *testing.T) {
	for _, tc := range []struct {
		name   string
		header http.Header
		want   string
	}{
		{"canonical", http.Header{"X-Remote-User": {"alice"}}, "alice"},
		{"lowercase", http.Header{"x-remote-user": {" alice "}}, "alice"},
		{"case collision", http.Header{"X-Remote-User": {"alice"}, "x-remote-user": {"bob"}}, ""},
		{"merged values", http.Header{"X-Remote-User": {"alice, bob"}}, ""},
		{"same value repeated", http.Header{"X-Remote-User": {"alice", "alice"}}, ""},
		{"empty", http.Header{"X-Remote-User": {" "}}, ""},
		{"missing", http.Header{}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := ProxyAuth{Header: "X-Remote-User"}
			got, ok := a.Username(&http.Request{Header: tc.header})
			if got != tc.want || ok != (tc.want != "") {
				t.Fatalf("Username() = %q, %t; want %q", got, ok, tc.want)
			}
		})
	}
}

func TestProxyTrustBoundary(t *testing.T) {
	for _, tc := range []struct {
		name, remote string
		cidrs        []string
		want         bool
	}{
		{"default deny", "127.0.0.1:1234", nil, false},
		{"invalid network", "127.0.0.1:1234", []string{"invalid"}, false},
		{"IPv4", "192.0.2.5:1234", []string{"192.0.2.0/24"}, true},
		{"IPv6", "[2001:db8::1]:1234", []string{"2001:db8::/32"}, true},
		{"outside network", "192.0.3.5:1234", []string{"192.0.2.0/24"}, false},
		{"malformed peer", "unknown:1234", []string{"0.0.0.0/0"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := ProxyAuth{TrustedCIDRs: tc.cidrs}
			r := &http.Request{RemoteAddr: tc.remote, Header: http.Header{"X-Forwarded-For": {"192.0.2.5"}, "Forwarded": {"for=192.0.2.5"}}}
			if got := a.TrustedRequest(r); got != tc.want {
				t.Fatalf("TrustedRequest() = %t, want %t", got, tc.want)
			}
		})
	}
}
