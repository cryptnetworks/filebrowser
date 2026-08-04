package runner

import (
	"errors"
	"path/filepath"
	"slices"

	"github.com/filebrowser/filebrowser/v2/settings"
)

var ErrCommandNotAllowed = errors.New("command not allowed")
var ErrCommandPathNotAbsolute = errors.New("command executable must use an absolute path")

// ParseDirectCommand parses an administrator-configured command into an
// explicit argument vector. Requiring an absolute executable prevents the
// process environment (especially PATH) from selecting a different binary.
func ParseDirectCommand(raw string) ([]string, error) {
	name, args, err := SplitCommandAndArgs(raw)
	if err != nil {
		return nil, err
	}
	if !filepath.IsAbs(name) {
		return nil, ErrCommandPathNotAbsolute
	}

	return append([]string{name}, args...), nil
}

// ParseCommand parses an event hook as a direct process invocation. The
// legacy shell setting is deliberately ignored.
func ParseCommand(_ *settings.Settings, raw string) (command []string, name string, err error) {
	command, err = ParseDirectCommand(raw)
	if err != nil {
		return nil, "", err
	}

	return command, command[0], nil
}

// ParseAllowedCommand parses an interactive command and returns the canonical
// argv from the administrator-managed allowlist. The requested command must
// match an allowlist entry exactly, including its arguments, and is never
// passed through the configured shell.
func ParseAllowedCommand(raw string, allowlist []string) ([]string, error) {
	requested, err := ParseDirectCommand(raw)
	if err != nil {
		return nil, ErrCommandNotAllowed
	}

	for _, allowed := range allowlist {
		candidate, err := ParseDirectCommand(allowed)
		if err != nil {
			continue
		}
		if slices.Equal(requested, candidate) {
			return candidate, nil
		}
	}

	return nil, ErrCommandNotAllowed
}
