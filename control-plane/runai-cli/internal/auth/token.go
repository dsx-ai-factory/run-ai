package auth

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
)

// CustomClaims represents the custom claims in the JWT.
type CustomClaims struct {
	jwt.RegisteredClaims
	UserID              string `json:"subject_uuid"`
	UserEmail           string `json:"email"`
	Uid                 string `json:"uid,omitempty"`
	Gid                 string `json:"gid,omitempty"`
	SupplementaryGroups []int  `json:"supplementarygroups,omitempty"`
}

const (
	PlaintextFormat  = "plaintext"
	KubeconfigFormat = "kubeconfig"
)

func GetToken(ctx context.Context) (string, error) {
	manager := GetManager()
	tokenStr, err := manager.GetAccessToken(ctx)
	if err != nil {
		return "", fmt.Errorf("Authentication failed. %w \n%s", err, consts.ReLoginMsg) //nolint:staticcheck // user facing error
	}
	return tokenStr, nil
}

func GetTokenDisplay(ctx context.Context, format string) (string, error) {
	tokenStr, err := GetToken(ctx)
	if err != nil {
		return "", err
	}

	switch format {
	case PlaintextFormat:
		return tokenStr, nil
	case KubeconfigFormat:
		cred := NewExecCredential(tokenStr)
		byteArr, jsonErr := json.MarshalIndent(cred, "", "\t")
		if jsonErr != nil {
			return "", jsonErr
		}

		return string(byteArr), nil
	default:
		return "", fmt.Errorf("unknown token display format: %s", format)
	}
}

func GetClaims(tokenStr string) (*CustomClaims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenStr, &CustomClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token. %w", err)
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, fmt.Errorf("failed to parse claims")
	}
	return claims, nil
}
