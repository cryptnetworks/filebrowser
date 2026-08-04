package cmd

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/filebrowser/filebrowser/v2/auth"
	"github.com/filebrowser/filebrowser/v2/settings"
)

// TestEnvCollisions ensures that there are no collisions in the produced environment
// variable names for all commands and their flags.
func TestEnvCollisions(t *testing.T) {
	testEnvCollisions(t, rootCmd)
}

func testEnvCollisions(t *testing.T, cmd *cobra.Command) {
	for _, cmd := range cmd.Commands() {
		testEnvCollisions(t, cmd)
	}

	replacements := generateEnvKeyReplacements(cmd)
	envVariables := []string{}

	for i := range replacements {
		if i%2 != 0 {
			envVariables = append(envVariables, replacements[i])
		}
	}

	duplicates := lo.FindDuplicates(envVariables)

	if len(duplicates) > 0 {
		t.Errorf("Found duplicate environment variable keys for command %q: %v", cmd.Name(), duplicates)
	}
}

// TestGetSettingsFollowExternalSymlinks ensures that the followExternalSymlinks
// flag is persisted to the server config when set via "config set".
func TestGetSettingsFollowExternalSymlinks(t *testing.T) {
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	addConfigFlags(flags)

	if err := flags.Parse([]string{"--followExternalSymlinks"}); err != nil {
		t.Fatal(err)
	}

	set := &settings.Settings{AuthMethod: auth.MethodJSONAuth}
	ser := &settings.Server{}

	if _, err := getSettings(flags, set, ser, &auth.JSONAuth{}, false); err != nil {
		t.Fatal(err)
	}

	if !ser.FollowExternalSymlinks {
		t.Error("expected FollowExternalSymlinks to be persisted as true")
	}
}

func TestAuthConfigForDisplayRedactsSecrets(t *testing.T) {
	config := auth.ConfigForDisplay(&auth.JSONAuth{ReCaptcha: &auth.ReCaptcha{
		Host:   "https://captcha.example",
		Key:    "public-site-key",
		Secret: "private-secret",
	}})

	b, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	output := string(b)
	if strings.Contains(output, "private-secret") {
		t.Fatalf("authentication secret leaked in output: %s", output)
	}
	if !strings.Contains(output, auth.RedactedConfigValue) {
		t.Fatalf("redaction marker missing from output: %s", output)
	}
	if !strings.Contains(output, "public-site-key") {
		t.Fatalf("non-secret site key unexpectedly omitted: %s", output)
	}
}

func TestParseUsernameOrIDBoundaries(t *testing.T) {
	maxUint := ^uint(0)
	overflow := "4294967296"
	if strconv.IntSize == 64 {
		overflow = "18446744073709551616"
	}

	tests := []struct {
		name         string
		input        string
		wantUsername string
		wantID       uint
		wantRangeErr bool
	}{
		{name: "zero", input: "0", wantID: 0},
		{name: "one", input: "1", wantID: 1},
		{name: "maximum platform value", input: strconv.FormatUint(uint64(maxUint), 10), wantID: maxUint},
		{name: "platform overflow", input: overflow, wantRangeErr: true},
		{name: "uint64 overflow", input: "18446744073709551616", wantRangeErr: true},
		{name: "negative value remains a username", input: "-1", wantUsername: "-1"},
		{name: "malformed value remains a username", input: "123alice", wantUsername: "123alice"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			username, id, err := parseUsernameOrID(tt.input)
			if tt.wantRangeErr {
				if !errors.Is(err, strconv.ErrRange) {
					t.Fatalf("got error %v, want strconv.ErrRange", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if username != tt.wantUsername || id != tt.wantID {
				t.Fatalf("got username %q and id %d, want username %q and id %d", username, id, tt.wantUsername, tt.wantID)
			}
		})
	}
}

func TestMarshalUsesOwnerOnlyPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not expose POSIX permission bits")
	}

	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte("old"), 0o666); err != nil {
		t.Fatal(err)
	}
	if err := marshal(path, map[string]string{"secret": "value"}); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("got permissions %O, want 0600", got)
	}
}
