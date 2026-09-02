package mocks

import (
	"net/http"

	gocloak "github.com/Nerzal/gocloak/v13"
)

func GetMockJWT() *gocloak.JWT {
	jwt := gocloak.JWT{
		AccessToken:      "123",
		IDToken:          "123",
		ExpiresIn:        123,
		RefreshExpiresIn: 60,
		RefreshToken:     "124",
		TokenType:        "123",
		NotBeforePolicy:  123,
		SessionState:     "123",
		Scope:            "123",
	}
	return &jwt
}

func GetMockAccessToken() *string {
	return gocloak.StringP("123")

}

func GetKeycloakError() *gocloak.APIError {
	return &gocloak.APIError{
		Code:    http.StatusInternalServerError,
		Message: "failed",
	}
}
