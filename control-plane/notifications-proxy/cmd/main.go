package main

import (
	"context"
	"os"

	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/authorization"
	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/events_listeners"
	"github.com/run-ai/runai/runai-common-packages/cache/v2"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub"

	audit "github.com/run-ai/runai/control-plane/audit-service/pkg/client"
	proxy_config "github.com/run-ai/runai/control-plane/notifications-proxy/internal/config"
	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/router"
	notifyclient "github.com/run-ai/runai/control-plane/notifications-service/pkg/client"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/metrics"
	"github.com/run-ai/runai/runai-common-packages/runaihttp"
	"github.com/run-ai/runai/runai-common-packages/tracing"
	"k8s.io/utils/env"

	authorization_client "github.com/run-ai/runai/control-plane/authorization/pkg/client"

	_ "github.com/run-ai/runai/runai-common-packages/fips"
)

func init() {
	logging.Init(nil)
	_, err := metrics.Init(nil)
	if err != nil {
		logging.Errorf("Failed to initialize metrics: %v", err)
	}
	_, err = tracing.Init(nil)
	if err != nil {
		logging.Errorf("Failed to initialize tracing: %v", err)
	}
	err = audit.Init(nil)
	if err != nil {
		logging.Errorf("Failed to initialize audit: %v", err)
	}
	cache.SetLogger(logging.Actual())
	if err = cache.Init(nil); err != nil {
		logging.Fatalf("Failed to initialize cache: %v", err)
	}
	proxy_config.Init()
	runaidb.SetLogger(logging.Actual())
	if err = runaidb.Init(nil); err != nil {
		logging.Fatalf("failed to initialize database: %v", err)
	}
	runaipubsub.SetLogger(logging.Actual())
	if _, err := runaipubsub.InitAsync(nil); err != nil {
		logging.Errorf("Failed to initialize pubsub: %v", err)
	}

	authorization_client.Init(nil)

	if b, _ := env.GetBool("USE_MOCKS", false); b {
		logging.Infof("Using mock for authorization client")
		authorization_client.SetMockAuthorizationClient(&authorization.AuthMock{})
	}
}

func main() {
	logging.Infof(">>>>>>>>>>> Starting notifications proxy <<<<<<<<<<<<")
	ctx, cancel := context.WithCancel(context.Background())

	tenantDeleteListener := events_listeners.NewTenantDeleteListener()
	tenantDeleteListener.Listen(ctx)

	if err := initNotifyClient(ctx); err != nil {
		logging.Warnf("Failed to init notify client: %v", err)
	}

	runaihttp.ServeHTTP(router.CreateRouter(), nil)

	logging.Infof("received stop signal")
	cancel()
	if err := runaipubsub.Close(); err != nil {
		logging.Errorf("Failed to close pubsub: %v", err)
	}
	os.Exit(0)
}

func initNotifyClient(ctx context.Context) error {
	_, err := notifyclient.Init(nil, ctx)
	if err != nil {
		return err
	}

	notifyclient.RegisterErrorHandler(func(err error) {
		logging.Errorf("Error in notify client: %v", err)
	})

	return nil
}
