package cmd

import (
	"encoding/json"
	"errors"
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
	config := authConfigForDisplay(&auth.JSONAuth{ReCaptcha: &auth.ReCaptcha{
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
	if !strings.Contains(output, redactedConfigValue) {
		t.Fatalf("redaction marker missing from output: %s", output)
	}
	if !strings.Contains(output, "public-site-key") {
		t.Fatalf("non-secret site key unexpectedly omitted: %s", output)
	}
}

func TestParseUsernameOrIDRejectsOverflow(t *testing.T) {
	overflow := "4294967296"
	if strconv.IntSize == 64 {
		overflow = "18446744073709551616"
	}

	_, _, err := parseUsernameOrID(overflow)
	if !errors.Is(err, strconv.ErrRange) {
		t.Fatalf("got error %v, want strconv.ErrRange", err)
	}
}

func TestParseUsernameOrIDPreservesUsernames(t *testing.T) {
	username, id, err := parseUsernameOrID("123alice")
	if err != nil {
		t.Fatal(err)
	}
	if username != "123alice" || id != 0 {
		t.Fatalf("got username %q and id %d", username, id)
	}
}
