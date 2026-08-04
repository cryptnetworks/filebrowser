package auth

import "fmt"

// RedactedConfigValue is used in human-readable configuration output whenever
// the stored value is authentication material or may embed such material.
const RedactedConfigValue = "[REDACTED]"

// ConfigForDisplay returns a serialization-safe view of an authentication
// configuration with secret-bearing values removed. It is the single boundary
// used by CLI and support-oriented output; persistent exports remain encrypted
// only by their surrounding storage controls and are deliberately separate.
func ConfigForDisplay(auther Auther) any {
	switch value := auther.(type) {
	case JSONAuth:
		return jsonConfigForDisplay(value)
	case *JSONAuth:
		return jsonConfigForDisplay(*value)
	case ProxyAuth:
		return struct {
			Header string `json:"header"`
		}{Header: value.Header}
	case *ProxyAuth:
		return struct {
			Header string `json:"header"`
		}{Header: value.Header}
	case *HookAuth:
		command := ""
		if value.Command != "" {
			command = RedactedConfigValue
		}
		return struct {
			Command string `json:"command"`
		}{Command: command}
	case NoAuth, *NoAuth:
		return struct{}{}
	default:
		return struct {
			Type string `json:"type"`
		}{Type: fmt.Sprintf("%T", auther)}
	}
}

func jsonConfigForDisplay(value JSONAuth) any {
	if value.ReCaptcha == nil {
		return struct {
			ReCaptcha any `json:"recaptcha"`
		}{ReCaptcha: nil}
	}

	return struct {
		ReCaptcha struct {
			Host   string `json:"host"`
			Key    string `json:"key"`
			Secret string `json:"secret"`
		} `json:"recaptcha"`
	}{
		ReCaptcha: struct {
			Host   string `json:"host"`
			Key    string `json:"key"`
			Secret string `json:"secret"`
		}{
			Host:   value.ReCaptcha.Host,
			Key:    value.ReCaptcha.Key,
			Secret: RedactedConfigValue,
		},
	}
}
