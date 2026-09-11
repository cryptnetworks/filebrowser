package auth

import (
	"errors"
	"log"
	"net"
	"net/http"
	"net/netip"
	"os"
	"strings"

	fberrors "github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
)

// MethodProxyAuth is used to identify no auth.
const MethodProxyAuth settings.AuthMethod = "proxy"

// ProxyAuth is a proxy implementation of an auther.
type ProxyAuth struct {
	Header       string   `json:"header"`
	TrustedCIDRs []string `json:"trustedCidrs,omitempty"`
}

// Auth authenticates the user via an HTTP header.
func (a ProxyAuth) Auth(r *http.Request, usr users.Store, setting *settings.Settings, srv *settings.Server) (*users.User, error) {
	if !a.TrustedRequest(r) {
		log.Printf("proxy auth rejected from %q", r.RemoteAddr)
		return nil, os.ErrPermission
	}

	username, ok := a.Username(r)
	if !ok {
		log.Printf("proxy auth rejected due to ambiguous identity header %q", a.Header)
		return nil, os.ErrPermission
	}

	user, err := usr.Get(srv.Root, srv.FollowExternalSymlinks, username)
	if errors.Is(err, fberrors.ErrNotExist) {
		return a.createUser(usr, setting, srv, username)
	}
	return user, err
}

// TrustedRequest reports whether the request came from an explicitly trusted
// proxy network. Proxy auth is disabled unless at least one CIDR is configured.
func (a ProxyAuth) TrustedRequest(r *http.Request) bool {
	if len(a.TrustedCIDRs) == 0 {
		return false
	}

	host := r.RemoteAddr
	if splitHost, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		host = splitHost
	}

	addr, err := netip.ParseAddr(strings.TrimSpace(host))
	if err != nil {
		return false
	}

	for _, cidr := range a.TrustedCIDRs {
		prefix, err := netip.ParsePrefix(strings.TrimSpace(cidr))
		if err != nil {
			continue
		}
		if prefix.Contains(addr) {
			return true
		}
	}

	return false
}

// Username returns the asserted identity from the configured proxy header.
// Ambiguous or duplicated header values are rejected.
func (a ProxyAuth) Username(r *http.Request) (string, bool) {
	values := r.Header.Values(a.Header)
	if len(values) != 1 {
		return "", false
	}

	username := strings.TrimSpace(values[0])
	if username == "" {
		return "", false
	}

	return username, true
}

func (a ProxyAuth) createUser(usr users.Store, setting *settings.Settings, srv *settings.Server, username string) (*users.User, error) {
	const randomPasswordLength = settings.DefaultMinimumPasswordLength + 10
	pwd, err := users.RandomPwd(randomPasswordLength)
	if err != nil {
		return nil, err
	}

	var hashedRandomPassword string
	hashedRandomPassword, err = users.ValidateAndHashPwd(pwd, setting.MinimumPasswordLength)
	if err != nil {
		return nil, err
	}

	user := &users.User{
		Username:     username,
		Password:     hashedRandomPassword,
		LockPassword: true,
	}
	setting.Defaults.Apply(user)
	user.Perm.Admin = false
	user.Perm.Execute = false
	user.Commands = []string{}

	var derivedScope bool
	if derivedScope, err = setting.CreateUserHome(user, srv.Root, false); err != nil {
		return nil, err
	}

	if err = usr.SaveProvisioned(user, derivedScope); err != nil {
		return nil, err
	}

	return user, nil
}

// LoginPage tells that proxy auth doesn't require a login page.
func (a ProxyAuth) LoginPage() bool {
	return false
}
