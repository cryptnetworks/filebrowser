package runner

import (
	"errors"
	"slices"

	"github.com/filebrowser/filebrowser/v2/settings"
)

var ErrCommandNotAllowed = errors.New("command not allowed")

// ParseCommand parses the command taking in account if the current
// instance uses a shell to run the commands or just calls the binary
// directly.
func ParseCommand(s *settings.Settings, raw string) (command []string, name string, err error) {
	name, args, err := SplitCommandAndArgs(raw)
	if err != nil {
		return
	}

	if len(s.Shell) == 0 || s.Shell[0] == "" {
		command = append(command, name)
		command = append(command, args...)
	} else {
		command = append(command, s.Shell...)
		command = append(command, raw)
	}

	return command, name, nil
}

// ParseAllowedCommand parses an interactive command and returns the canonical
// argv from the administrator-managed allowlist. The requested command must
// match an allowlist entry exactly, including its arguments, and is never
// passed through the configured shell.
func ParseAllowedCommand(raw string, allowlist []string) ([]string, error) {
	name, args, err := SplitCommandAndArgs(raw)
	if err != nil {
		return nil, err
	}
	requested := append([]string{name}, args...)

	for _, allowed := range allowlist {
		allowedName, allowedArgs, err := SplitCommandAndArgs(allowed)
		if err != nil {
			continue
		}
		candidate := append([]string{allowedName}, allowedArgs...)
		if slices.Equal(requested, candidate) {
			return candidate, nil
		}
	}

	return nil, ErrCommandNotAllowed
}
