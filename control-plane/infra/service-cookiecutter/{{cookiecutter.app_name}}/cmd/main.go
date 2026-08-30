package main

import (
	"flag"
	"os"
	"time"

	authClient "github.com/run-ai/runai/control-plane/authorization/pkg/client"
	runaihttp "github.com/run-ai/runai/runai-common-packages/runaihttp"
	"github.com/run-ai/runai/control-plane/{{cookiecutter.app_name}}/pkg/config"
	"github.com/run-ai/runai/control-plane/{{cookiecutter.app_name}}/pkg/router"

CACHE-CC-PREFIX	"github.com/run-ai/runai/runai-common-packages/cache/v2"
DB-CC-PREFIX	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/tracing"
)

func init() {
	logging.Init(nil)
    metrics.Init(nil)
	tracing.Init(nil)
	if err := config.InitEnvironment(); err != nil {
		logging.Fatalf("Failed to init configuration: %v", err)
	}
DB-CC-PREFIX	dao_utils.Init(dao_utils.ConstructConfig("{{cookiecutter.__app_name_underscore_lower}}"))
CACHE-CC-PREFIX	if err := cache.Init(nil); err != nil {
		logging.Errorf("Failed to init cache: %w", err)
	}
	authClient.Init(nil)
}

func main() {
	logging.Infof("Starting Run:AI {{cookiecutter.app_name}}")

	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	runaihttp.ServeHTTP(
		router.CreateRouter(),
		nil,
	)

	logging.Infof("Shutting down {{cookiecutter.app_name}}")
	os.Exit(0)
}
