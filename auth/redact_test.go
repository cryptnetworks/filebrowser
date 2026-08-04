package auth

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestConfigForDisplayRedactsEverySecretBearingField(t *testing.T) {
	tests := []struct {
		name   string
		auther Auther
		secret string
	}{
		{
			name: "recaptcha secret",
			auther: &JSONAuth{ReCaptcha: &ReCaptcha{
				Host: "https://captcha.example", Key: "public-key", Secret: "private-secret",
			}},
			secret: "private-secret",
		},
		{
			name:   "hook command may embed credentials",
			auther: &HookAuth{Command: "/opt/hooks/login --token embedded-token"},
			secret: "embedded-token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := json.Marshal(ConfigForDisplay(tt.auther))
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(output), tt.secret) {
				t.Fatalf("secret leaked in output: %s", output)
			}
			if !strings.Contains(string(output), RedactedConfigValue) {
				t.Fatalf("redaction marker missing from output: %s", output)
			}
		})
	}
}
