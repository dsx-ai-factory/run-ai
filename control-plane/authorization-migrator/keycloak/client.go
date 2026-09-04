/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package keycloak

import (
	"context"
	"fmt"
	"time"

	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/control-plane/authorization-migrator/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	gocloak "github.com/Nerzal/gocloak/v13"
)

var (
	conf   *config.ConfigData
	client *gocloak.GoCloak
	token  *gocloak.JWT
)

func Init(configuration *config.ConfigData) {
	conf = configuration
	client = gocloak.NewClient(conf.KeycloakHost + "/auth")
}

func fetchFreshToken() {
	if token != nil {
		returnedToken, err := client.RefreshToken(context.Background(), token.RefreshToken, conf.KeycloakAdminUser, conf.KeycloakAdminPassword, "master")
		if err == nil {
			token = returnedToken
			return
		}
	}
	returnedToken, err := client.LoginAdmin(context.Background(), conf.KeycloakAdminUser, conf.KeycloakAdminPassword, "master")
	if err != nil {
		panic(err)
	}
	token = returnedToken
}

func SetRefreshTokenTicker() *time.Ticker {
	fetchFreshToken()

	ticker := time.NewTicker(time.Duration(token.ExpiresIn-5) * time.Second)
	go func() {
		for range ticker.C {
			fetchFreshToken()
			logging.Infof("Keycloak token refreshed")
		}
	}()

	return ticker
}

func GetAppByKeyCloakId(keyCloakId string, realm string) (*gocloak.Client, error) {
	clients, err := client.GetClients(context.Background(), token.AccessToken, realm, gocloak.GetClientsParams{Max: ptr.To[int](1000)})
	if err != nil {
		logging.Errorf("Failed to get clients from realm <%s>, error: %s", realm, err.Error())
		return nil, err
	}

	for _, returnedClient := range clients {
		if keyCloakId == *returnedClient.ID {
			return returnedClient, nil
		}
	}

	return nil, fmt.Errorf("client with id %s did not found in realm %s", keyCloakId, realm)
}

func GetKeycloakUserById(id string, realm string) (*gocloak.User, error) {
	return client.GetUserByID(context.Background(), token.AccessToken, realm, id)
}

func UpdateUser(user gocloak.User, realm string) error {
	return client.UpdateUser(context.Background(), token.AccessToken, realm, user)
}

func UpdateClient(updatedClient gocloak.Client, realm string) error {
	return client.UpdateClient(context.Background(), token.AccessToken, realm, updatedClient)
}

func GetUsersWithFilter(username string, realm string, filter gocloak.GetUsersParams) ([]*gocloak.User, error) {
	return client.GetUsers(context.Background(), token.AccessToken, realm, filter)
}
