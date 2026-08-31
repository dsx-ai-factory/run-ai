//go:build integration
// +build integration


/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"log"
	"net/url"
	"os"
	"strings"
	"testing"

	openapi "github.com/run-ai/runai/control-plane/notifications-proxy/pkg/client/generated"
	notifyclient "github.com/run-ai/runai/control-plane/notifications-service/pkg/client"
	"github.com/run-ai/runai/control-plane/notifications-service/pkg/client/generated/config_client"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub"
)

var (
	client              *openapi.APIClient
	notificationsClient *config_client.APIClient
)

const (
	// we mock authorization client to return system+tenant scope for tenant id 1 and tenant scope for tenant id 3
	tenantIdWithSystemScope    = "1"
	tenantIdWithoutSystemScope = "3"
	tenantIdWithoutPermissions = "4"
)

func TestMain(m *testing.M) {
	createDirSecretTesting()
	setEnvs()
	defer deleteDirSecretTesting()
	initProxyAPIClient()
	initConfigClient()
	if err := initNotifyClient(); err != nil {
		log.Fatalf("failed to initialize notify client: %v", err)
	}
	if err := runaipubsub.Init(nil); err != nil {
		log.Fatalf("failed to initialize pubsub: %v", err)
	}
	os.Exit(m.Run())
}

func initProxyAPIClient() {
	notificationsProxyHost := GetEnvOrDefault("NOTIFICATIONS_PROXY_HOST", "localhost:8080")
	notificationsProxyScheme := GetEnvOrDefault("NOTIFICATIONS_PROXY_SCHEME", "http")

	clientConfig := openapi.NewConfiguration()
	clientConfig.Host = notificationsProxyHost
	clientConfig.Scheme = notificationsProxyScheme

	client = openapi.NewAPIClient(clientConfig)
}

func initConfigClient() {
	cfgServer := config_client.NewConfiguration()
	url_, _ := url.Parse(GetEnvOrDefault("NOTIFY_CONFIG_URL", "http://localhost:8085"))
	cfgServer.Host = url_.Host
	cfgServer.Scheme = url_.Scheme

	notificationsClient = config_client.NewAPIClient(cfgServer)
}

func initNotifyClient() error {
	configUrl, _ := url.Parse(GetEnvOrDefault("NOTIFY_CONFIG_URL", "http://localhost:8085"))
	notifyUrl, _ := url.Parse(GetEnvOrDefault("NOTIFY_URL", "http://localhost:5000"))
	_, err := notifyclient.Init(&notifyclient.NotifyClientConfig{
		NotifyUrl: notifyUrl.String(),
		ConfigUrl: configUrl.String(),
	}, context.Background())
	return err
}

func GetEnvOrDefault(key, defaultVal string) string {
	if value, ok := os.LookupEnv(key); ok {
		return strings.Trim(value, " ")
	}
	return defaultVal
}

func createDirSecretTesting() error {
	if err := os.MkdirAll("secrets", 0o755); err != nil {
		return err
	}
	return nil
}

func deleteDirSecretTesting() error {
	if err := os.RemoveAll("secrets"); err != nil {
		return err
	}
	return nil
}

func setEnvs() {
	setEnvIfNotSet("SERVICE_NAME", "notifications-proxy-integration-tests")
	setEnvIfNotSet("SECRETS_MOUNT_PATH", "./secrets")
	setEnvIfNotSet("PUBSUB_TYPE", "redis")
	setEnvIfNotSet("PUBSUB_URL", "localhost:6379")
}

func setEnvIfNotSet(key, value string) {
	if _, ok := os.LookupEnv(key); !ok {
		os.Setenv(key, value)
	}
}
