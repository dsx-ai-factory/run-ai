package auth

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
)

const AppTimer = "app_login"

var authenticator Authenticator

func SetAuthenticator(authenticatorSet Authenticator) {
	authenticator = authenticatorSet
}

func ApplicationLogin(txn *transaction.Txn, name, secret, authUrl, tokenFilePath string) (string, error) {
	var err error
	timerID := txn.AddTimer(AppTimer)
	defer txn.StopTimer(timerID)

	if authenticator == nil {
		authenticator, err = NewApplicationAuthenticator(authUrl, name, secret)
		if err != nil {
			return FailedAuthAppMsg, err
		}
	}
	ctx := txn.GetCtx()
	token, err := authenticator.Authenticate(ctx)
	if err != nil {
		return FailedAuthAppMsg, err
	}
	writer := NewFileBased64TokenReadWriter()
	err = writer.WriteToken(tokenFilePath, token)
	if err != nil {
		return FailedSaveTokenMsg, err
	}
	return AppLoginSuccessfullyMsg, nil
}
