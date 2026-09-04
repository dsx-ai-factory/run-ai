/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"os"
	"os/signal"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/app"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/config/basic"
	runaimetrics "github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics"
	"github.com/rs/zerolog/log"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/config/advanced"
)

const (
	envConfPath       = "CONF_PATH"
	envConfigProvider = "ENV_CONFIG_PROVIDER"
)

func main() {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	app := initApp()
	app.Start()

	<-signalCh
	log.Info().Msg("service received stop signal")

	app.Stop()
}

func initApp() *app.App {
	log.Info().Msg("service initializing")
	runaimetrics.Init(nil)
	configFactory := GetConfigFactory()
	appInstance := app.NewApp(configFactory)
	appInstance.InitializeEventRouters()

	return appInstance
}

func GetConfigFactory() app.ConfigFactory {
	path := getEnv(envConfPath)
	providerType := getEnv(envConfigProvider)
	log.Info().Msgf("requested config provider is: %s", providerType)

	switch providerType {
	case "advanced":
		log.Info().Msg("using advanced config provider")
		return advanced.NewFactory(path)
	default:
		log.Info().Msg("using basic config provider")
		return basic.NewFactory(path)
	}
}

func getEnv(envVar string) string {
	return strings.Trim(os.Getenv(envVar), " ")
}
