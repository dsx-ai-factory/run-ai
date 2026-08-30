package generator

import (
	"fmt"
	"regexp"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

const (
	natsServerConfigKey = "nats-server.conf"
)

// authBlockRe matches the content of an authorization { ... } block (no nested braces).
var authBlockRe = regexp.MustCompile(`(?s)\bauthorization\s*\{([^{}]*)\}`)

// authTokenRe matches a token assignment inside an authorization block,
// capturing either a quoted value (group 1) or an unquoted value (group 2).
var authTokenRe = regexp.MustCompile(`(?m)^\s*token\s*[:=]\s*(?:"([^"]+)"|(\S+))`)

type NatsMigrationHandler struct{}

func (h *NatsMigrationHandler) Migrate(spec SecretSpec, legacySecret *corev1.Secret, legacyName string) (map[string][]byte, error) {
	configData, ok := legacySecret.Data[natsServerConfigKey]
	if !ok {
		return nil, fmt.Errorf("legacy NATS secret %s is missing expected key %s; "+
			"manual intervention required", legacyName, natsServerConfigKey)
	}

	token := extractNatsToken(string(configData))
	if token == "" {
		return nil, fmt.Errorf("legacy NATS secret %s contains %s but no authorization token could be extracted; "+
			"manual intervention required", legacyName, natsServerConfigKey)
	}

	return map[string][]byte{
		"AUTH_TOKEN": []byte(token),
	}, nil
}

// extractNatsToken scans all authorization { ... } blocks in a NATS server config
// and returns the token value from the first block that contains one.
//
// It uses regex rather than the NATS conf library to avoid failures when the config
// contains $ENV_VAR references that are only defined inside the NATS server process
// (e.g. $NATS_SERVER_NAME). It iterates all blocks rather than taking the first one
// because the cluster section also contains a nested authorization block (with
// user/password, not token), which may appear before the top-level block.
//
// Returns "" if no token is found or if the token value is itself an unresolved
// env var reference (e.g. token: $SOME_TOKEN).
func extractNatsToken(config string) string {
	for _, blockMatch := range authBlockRe.FindAllStringSubmatch(config, -1) {
		tokenMatch := authTokenRe.FindStringSubmatch(blockMatch[1])
		if tokenMatch == nil {
			continue
		}
		if tokenMatch[1] != "" {
			return tokenMatch[1] // quoted: token: "value"
		}
		// Unquoted token — reject env var references since we can't resolve them.
		if strings.HasPrefix(tokenMatch[2], "$") {
			return ""
		}
		return tokenMatch[2]
	}
	return ""
}
