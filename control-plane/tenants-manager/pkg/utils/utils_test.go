package utils_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestCheckHttpResponse_HappyFlow(t *testing.T) {
	httpResponse := &http.Response{
		StatusCode: http.StatusOK,
	}
	res := utils.CheckHttpResponse(httpResponse, nil, "action")
	assert.Nil(t, res)
}

func TestCheckHttpResponse_FailedResponse(t *testing.T) {
	httpResponse := &http.Response{
		StatusCode: http.StatusInternalServerError,
	}
	res := utils.CheckHttpResponse(httpResponse, nil, "action")
	assert.Equal(t, "failed to action", res.Error())
}

func TestCheckHttpResponse_WithError(t *testing.T) {
	httpResponse := &http.Response{
		StatusCode: http.StatusInternalServerError,
	}
	res := utils.CheckHttpResponse(httpResponse, errors.New("Error"), "action")
	assert.Equal(t, "failed to action: Error", res.Error())
}

func TestValidatePassword_ValidPasswords(t *testing.T) {
	validPasswords := []string{
		"Abcd!234",
		"Admin!234",
		"P@ssw0rd",
		"MyP@ss123",
		"Test-1234",
		"Hello_World1",
		"Str0ng#Pass",
		"A1!aaaaa",
	}

	for _, password := range validPasswords {
		t.Run(password, func(t *testing.T) {
			err := utils.ValidatePassword(password)
			assert.Nil(t, err, "password %q should be valid", password)
		})
	}
}

func TestValidatePassword_TooShort(t *testing.T) {
	err := utils.ValidatePassword("Ab!1234")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), utils.InvalidPasswordMessage)
}

func TestValidatePassword_MissingLowercase(t *testing.T) {
	err := utils.ValidatePassword("ABCD!234")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), utils.InvalidPasswordMessage)
}

func TestValidatePassword_MissingUppercase(t *testing.T) {
	err := utils.ValidatePassword("abcd!234")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), utils.InvalidPasswordMessage)
}

func TestValidatePassword_MissingDigit(t *testing.T) {
	err := utils.ValidatePassword("Abcdefg!")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), utils.InvalidPasswordMessage)
}

func TestValidatePassword_MissingSpecialChar(t *testing.T) {
	err := utils.ValidatePassword("Abcd1234")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), utils.InvalidPasswordMessage)
}

func TestValidatePassword_AllSpecialCharsAllowed(t *testing.T) {
	specialChars := []string{"-", "!", "@", "#", "$", "%", "^", "&", "*", "(", ")", "_", "+", "="}

	for _, char := range specialChars {
		password := "Abcd123" + char
		t.Run("special_char_"+char, func(t *testing.T) {
			err := utils.ValidatePassword(password)
			assert.Nil(t, err, "password with special char %q should be valid", char)
		})
	}
}

func TestValidatePassword_EmptyPassword(t *testing.T) {
	err := utils.ValidatePassword("")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), utils.InvalidPasswordMessage)
}

func TestGenerateTemporaryUserPassword_Length(t *testing.T) {
	got := utils.GenerateTemporaryUserPassword()
	assert.Len(t, got, 16, "password should be 16 cryptographically random characters")
}

func TestGenerateTemporaryUserPassword_ValidatesAsPassword(t *testing.T) {
	// Run many times since the character classes are placed at random positions.
	for i := 0; i < 10; i++ {
		got := utils.GenerateTemporaryUserPassword()
		err := utils.ValidatePassword(got)
		assert.Nil(t, err, "generated password must pass ValidatePassword (Keycloak-compliant)")
	}
}

func TestGenerateTemporaryUserPassword_UniqueEachCall(t *testing.T) {
	// Cryptographically random passwords should effectively never collide.
	seen := make(map[string]bool)
	const iterations = 10
	for i := 0; i < iterations; i++ {
		seen[utils.GenerateTemporaryUserPassword()] = true
	}
	assert.Equal(t, iterations, len(seen), "every generated password should be unique")
}
