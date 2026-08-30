package utils

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"html/template"
	"math/big"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/run-ai/runai/runai-common-packages/logging"
)

var (
	lowercaseRegex   = regexp.MustCompile(`[a-z]`)
	uppercaseRegex   = regexp.MustCompile(`[A-Z]`)
	digitRegex       = regexp.MustCompile(`[0-9]`)
	specialCharRegex = regexp.MustCompile(`[-!@#$%^&*()_+=]`)
)

const InvalidPasswordMessage = "password must be: 8 characters long, contain at least 1 lowercase letter, 1 uppercase letter, 1 digit, and 1 special character"

func ValidatePassword(password string) error {
	if len(password) < 8 || !lowercaseRegex.MatchString(password) || !uppercaseRegex.MatchString(password) || !digitRegex.MatchString(password) || !specialCharRegex.MatchString(password) {
		return errors.New(InvalidPasswordMessage)
	}
	return nil
}

func CheckHttpResponse(resp *http.Response, err error, action string) error {
	if err != nil {
		return fmt.Errorf("failed to %s: %s", action, err)
	}
	if resp.StatusCode <= 199 || resp.StatusCode >= 300 {
		return fmt.Errorf("failed to %s", action)
	}
	return nil
}

func DisplayNameToRealm(name string) string {
	return strings.ReplaceAll(strings.ToLower(name), " ", "")
}

const (
	temporaryPasswordLength = 16
	lowercaseChars          = "abcdefghijklmnopqrstuvwxyz"
	uppercaseChars          = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digitChars              = "0123456789"
	specialChars            = "-@_+="
	allChars                = lowercaseChars + uppercaseChars + digitChars + specialChars
)

// GenerateTemporaryUserPassword returns a cryptographically secure random
// password suitable for use as a Keycloak temporary password. It always
// satisfies ValidatePassword: at least one lowercase letter, one uppercase
// letter, one digit, and one special character.
func GenerateTemporaryUserPassword() string {
	chars := []byte{
		randomChar(lowercaseChars),
		randomChar(uppercaseChars),
		randomChar(digitChars),
		randomChar(specialChars),
	}
	for len(chars) < temporaryPasswordLength {
		chars = append(chars, randomChar(allChars))
	}
	// Shuffle so the guaranteed characters are not in fixed positions.
	for i := len(chars) - 1; i > 0; i-- {
		j := randomInt(i + 1)
		chars[i], chars[j] = chars[j], chars[i]
	}
	return string(chars)
}

func randomChar(charset string) byte {
	return charset[randomInt(len(charset))]
}

// randomInt returns a uniform random integer in [0, n) using crypto/rand.
func randomInt(n int) int {
	i, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		logging.Errorf("failed getting random int err: %+v. returninig 1", err)
		return n
	}
	return int(i.Int64())
}

func StringToUint(s string) (uint, error) {
	i, err := strconv.Atoi(s)
	if err != nil {
		logging.Errorf("failed parsing string to uint: %s", s)
		return 0, err
	}
	return uint(i), nil
}

func TemplateBaseUrl(hostTemplate string, name string) (string, error) {
	t, err := template.New("hostTemplate").Parse(hostTemplate)
	if err != nil {
		return "", err
	}
	var value bytes.Buffer
	err = t.Execute(&value, strings.ToLower(name))
	if err != nil {
		return "", err
	}

	return value.String(), nil
}

func UintToString(i int32) string {
	return strconv.FormatInt(int64(i), 10)
}
