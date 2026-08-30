package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

const (
	MaxSecretSize = 1024 * 4
)

type FileBased64TokenReadWriter struct {
}

func NewFileBased64TokenReadWriter() *FileBased64TokenReadWriter {
	return &FileBased64TokenReadWriter{}
}

type JsonTime struct {
	time.Time
}

func (jt *JsonTime) UnmarshalJSON(b []byte) (err error) {
	value := strings.Trim(string(b), `"`) // get rid of "
	if value == "" || value == "null" {
		return nil
	}

	t, err := time.Parse(time.RFC3339, value) // parse time
	if err != nil {
		return err
	}
	jt.Time = t
	return nil
}

func (jt *JsonTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + jt.Format(time.RFC3339) + `"`), nil
}

type Tokens struct {
	AccessToken  string   `json:"accessToken"`
	RefreshToken string   `json:"refreshToken,omitempty"`
	Expiration   JsonTime `json:"expiration"`
}

func (fbrw *FileBased64TokenReadWriter) ReadToken(path string) (*oauth2.Token, error) {
	logger := config.GetInstance().GetLogger()
	fileObj, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer func(fileObj *os.File) {
		closeError := fileObj.Close()
		if closeError != nil {
			logger.Error("failed to close file", zap.Error(closeError))
		}
	}(fileObj)
	rawTokens := Tokens{}
	jsonParser := json.NewDecoder(fileObj)
	if err = jsonParser.Decode(&rawTokens); err != nil {
		return nil, fmt.Errorf("failed to decode token from '%v': %w", path, err)
	}
	return fbrw.decodeToken(rawTokens)
}

func (fbrw *FileBased64TokenReadWriter) decodeToken(rawTokens Tokens) (*oauth2.Token, error) {
	accessToken, err := base64.StdEncoding.DecodeString(rawTokens.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to load access token: %w", err)
	}

	var refreshToken string
	if rawTokens.RefreshToken != "" {
		decodedRefresh, err := base64.StdEncoding.DecodeString(rawTokens.RefreshToken)
		if err != nil {
			return nil, fmt.Errorf("failed to load refresh token: %w", err)
		}
		refreshToken = string(decodedRefresh)
	}

	token := &oauth2.Token{
		AccessToken:  string(accessToken),
		RefreshToken: refreshToken,
		Expiry:       rawTokens.Expiration.Time,
	}

	idTokenKeyMap := map[string]interface{}{
		AccessTokenKey: accessToken,
	}
	token = token.WithExtra(idTokenKeyMap)
	return token, nil
}

func (fbrw *FileBased64TokenReadWriter) WriteToken(path string, token *oauth2.Token) error {
	logger := config.GetInstance().GetLogger()
	fileObj, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", path, err)
	}
	defer func(fileObj *os.File) {
		closeError := fileObj.Close()
		if closeError != nil {
			logger.Error("failed to close file", zap.Error(closeError))
		}
	}(fileObj)

	expiry := token.Expiry
	if expiry.IsZero() {
		var parsedClaims *CustomClaims
		parsedClaims, err = GetClaims(token.AccessToken)
		if err != nil {
			logger.Error("failed to get claims, does not set expiration time", zap.Error(err))
		} else {
			expiry = parsedClaims.ExpiresAt.Time
		}
	}

	encodedToken := Tokens{
		AccessToken:  base64.StdEncoding.EncodeToString([]byte(token.AccessToken)),
		RefreshToken: base64.StdEncoding.EncodeToString([]byte(token.RefreshToken)),
		Expiration:   JsonTime{expiry},
	}

	jsonParser := json.NewEncoder(fileObj)
	jsonParser.SetIndent("", "\t")
	if err = jsonParser.Encode(encodedToken); err != nil {
		return fmt.Errorf("failed to encode token: %w", err)
	}
	return nil
}

func ReadApplicationSecret(path string) (string, error) {
	logger := config.GetInstance().GetLogger()
	fileObj, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer func(fileObj *os.File) {
		closeError := fileObj.Close()
		if closeError != nil {
			logger.Error("failed to close file", zap.Error(closeError))
		}
	}(fileObj)

	secret := make([]byte, MaxSecretSize)
	n, err := fileObj.Read(secret)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	return string(secret[:n]), nil
}
