/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth

import (
	"errors"
	"net/url"
	"os"
	"strings"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/config"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/auth/pkce"

	api "github.com/run-ai/runai/control-plane/identity-manager/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
)

func Login(txn *transaction.Txn, authTimerName string, authenticator Authenticator, tokenFilePath string) error {
	timerID := txn.AddTimer(authTimerName)
	defer txn.StopTimer(timerID)

	ctx := txn.GetCtx()
	token, err := authenticator.Authenticate(ctx)
	if err != nil {
		if strings.HasSuffix(err.Error(), "EOF") {
			return errors.New(EOFAuthErr)
		}
		return err
	}
	writer := NewFileBased64TokenReadWriter()
	err = writer.WriteToken(tokenFilePath, token)
	if err != nil {
		return err
	}
	return nil
}

func NewLoginClient(baseUrl string) (*api.APIClient, error) {
	conf, err := config.CreateLoginClientConfig(baseUrl)
	if err != nil {
		return nil, err
	}
	return api.NewAPIClient(conf), nil
}

func InitBrowserLogin(baseUrl string, address string, printURLOnly bool) (Authenticator, error) {
	loginClient, err := NewLoginClient(baseUrl)
	if err != nil {
		return nil, err
	}
	redirectUrl := url.URL{
		Scheme: Http,
		Host:   address,
	}
	redirectSuccess, err := url.JoinPath(baseUrl, consts.SuccessLoginUrlPath)
	if err != nil {
		// This should never happen
		return nil, err
	}
	pkceParams, err := pkce.New()
	if err != nil {
		return nil, err
	}
	urlGetter := NewControlPlaneUrlGetter(loginClient.LoginsAPI, config.GetUserAgent(), redirectUrl.String(), false, pkceParams)
	webserver := NewLocalWebserverExchangeCodeGetter(address, redirectSuccess)
	handler := NewBrowserAuthExchangeCodeHandler(webserver, printURLOnly)
	exchanger := NewControlPlaneTokenExchanger(loginClient.TokensAPI, pkceParams.CodeVerifier)
	return NewIDPAuthenticator(urlGetter, exchanger, handler), nil
}

func InitRemoteBrowserLogin(baseUrl string) (Authenticator, error) {
	loginClient, err := NewLoginClient(baseUrl)
	if err != nil {
		return nil, err
	}
	pkceParams, err := pkce.New()
	if err != nil {
		return nil, err
	}
	urlGetter := NewControlPlaneUrlGetter(loginClient.LoginsAPI, config.GetUserAgent(), "", true, pkceParams)
	handler := NewRemoteBrowserHandler(os.Stdin)
	exchanger := NewControlPlaneTokenExchanger(loginClient.TokensAPI, pkceParams.CodeVerifier)
	return NewIDPAuthenticator(urlGetter, exchanger, handler), nil
}
