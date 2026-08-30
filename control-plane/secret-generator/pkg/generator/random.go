package generator

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/google/uuid"
	"github.com/sethvargo/go-password/password"
)

const (
	DefaultPasswordLength = 32

	letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

// pwGen produces passwords with the same character classes as the previous
// home-rolled implementation: standard letters/digits plus a narrow symbol
// set chosen to be safe in shell, YAML, env vars, and connection strings.
var pwGen = func() *password.Generator {
	g, err := password.NewGenerator(&password.GeneratorInput{
		Symbols: "!@#%^*()-_=+",
	})
	if err != nil {
		panic(fmt.Errorf("failed to construct password generator: %w", err))
	}
	return g
}()

// GeneratePassword generates a cryptographically secure random password.
// The first character is constrained to a letter so the value is safe to
// embed unquoted in configs that expand env-var references and re-parse the
// result (e.g. NATS HOCON), where a leading digit or special character would
// be misinterpreted by the parser. The remainder is produced by
// sethvargo/go-password and contains at least one digit and one symbol.
func GeneratePassword(length int) (string, error) {
	if length < 8 {
		length = 8
	}

	first, err := randomLetter()
	if err != nil {
		return "", err
	}
	rest, err := pwGen.Generate(length-1, 1, 1, false, true)
	if err != nil {
		return "", fmt.Errorf("failed to generate password body: %w", err)
	}
	return string(first) + rest, nil
}

// GenerateAlphanumeric returns a random alphanumeric string of the requested
// length, with the first character constrained to a letter (see
// GeneratePassword for rationale).
func GenerateAlphanumeric(length int) (string, error) {
	if length < 1 {
		length = DefaultPasswordLength
	}

	first, err := randomLetter()
	if err != nil {
		return "", err
	}
	if length == 1 {
		return string(first), nil
	}
	rest, err := pwGen.Generate(length-1, length/4, 0, false, true)
	if err != nil {
		return "", fmt.Errorf("failed to generate alphanumeric body: %w", err)
	}
	return string(first) + rest, nil
}

func randomLetter() (byte, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
	if err != nil {
		return 0, fmt.Errorf("failed to draw random letter: %w", err)
	}
	return letters[n.Int64()], nil
}

func GenerateHex(length int) (string, error) {
	if length < 1 {
		length = DefaultPasswordLength
	}

	byteLength := (length + 1) / 2
	bytes := make([]byte, byteLength)

	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	hexStr := hex.EncodeToString(bytes)
	if len(hexStr) > length {
		hexStr = hexStr[:length]
	}
	return hexStr, nil
}

func GenerateUUID() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate UUID: %w", err)
	}
	return id.String(), nil
}

// GenerateValue generates a value based on the key spec.
// ChartValue (non-empty) takes priority over random generation.
// Static type is handled upstream by the KeyResolver chain and should not reach here.
func GenerateValue(spec SecretKeySpec) (string, error) {
	if spec.ChartValue != nil && *spec.ChartValue != "" {
		return *spec.ChartValue, nil
	}

	length := spec.Length
	if length == 0 {
		length = DefaultPasswordLength
	}

	switch spec.Type {
	case "password", "":
		return GeneratePassword(length)
	case "alphanumeric":
		return GenerateAlphanumeric(length)
	case "hex":
		return GenerateHex(length)
	case "uuid":
		return GenerateUUID()
	default:
		return GeneratePassword(length)
	}
}
