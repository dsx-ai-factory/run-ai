package notification

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/config"
	notifyclient "github.com/run-ai/runai/control-plane/notifications-service/pkg/client"
	"github.com/run-ai/runai/control-plane/notifications-service/pkg/client/generated/config_client"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

const (
	TenantId   = "keycloak"
	ListenerId = "smtp-runai"
	RouteId    = "keycloak-route"
)

var NotifyIds = []string{"runai-email"}

func Init(ctx context.Context) {
	notifyConfig := runai_config.Get[*config.Notify]()

	if !notifyConfig.Enabled {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Notifications disabled by configuration, skipping initialization")
		return
	}

	retryAttempts := 3
	retryDelaySeconds := 5
	err := retry.Do(func() error {
		return initNotificationClientWithRegistration(ctx, notifyConfig)
	}, retry.Delay(time.Duration(retryDelaySeconds)*time.Second), retry.Attempts(uint(retryAttempts)), retry.OnRetry(func(attempt uint, err error) {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to init notification client, retrying, attempt: %d, error: %s", attempt+1, err.Error())
	}))

	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to init notification client after %d retries, error: %s. Service will continue without notifications.", retryAttempts, err.Error())
	} else {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Successfully initialized and registered notification client")
	}

	if err != nil {
		return
	}

	if routeErr := updateTenantRoute(ctx, notifyConfig.ConfigUrl); routeErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to update tenant route: %v", routeErr)
	} else {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Successfully updated tenant route for keycloak notifications")
	}
}

func initNotificationClientWithRegistration(ctx context.Context, notifyConfig *config.Notify) error {
	err := initNotificationClient(ctx, notifyConfig)
	if err != nil {
		return fmt.Errorf("failed to init notification client, error: %s", err.Error())
	}

	err = registerNotifications()
	if err != nil {
		return fmt.Errorf("failed to register notifications, error: %s", err.Error())
	}

	return nil
}

func initNotificationClient(ctx context.Context, notifyConfig *config.Notify) error {
	notifyClientConfig := &notifyclient.NotifyClientConfig{
		NotifyUrl: notifyConfig.NotifyUrl,
		ConfigUrl: notifyConfig.ConfigUrl,
	}

	_, err := notifyclient.Init(notifyClientConfig, ctx)
	return err
}

func registerNotifications() error {
	enableNotifications := true
	notifyConfig := runai_config.Get[*config.Notify]()
	notificationKinds := []*notifyclient.NotificationKind{
		{
			Category:    "keycloak",
			Type:        "reset",
			Description: "Password reset notification",
			Template:    notifyConfig.ResetPasswordEmailTemplate,
		},
		{
			Category:    "keycloak",
			Type:        "welcome",
			Description: "Welcome email notification",
			Template:    notifyConfig.WelcomeEmailTemplate,
		},
	}

	return notifyclient.RegisterNotification("identity-manager", notificationKinds, &enableNotifications)
}

func updateTenantRoute(ctx context.Context, configUrl string) error {
	configClient := createConfigClient(configUrl)
	routeConfig := buildRouteDataFromConfig()

	_, httpResponse, err := configClient.TenantRouteAPI.UpdateRoute(ctx, "keycloak", RouteId).RouteDataForUpdate(config_client.RouteDataForUpdate{
		TenantId:    TenantId,
		RouteConfig: routeConfig,
	}).Execute()
	if err == nil {
		return nil
	}

	if httpResponse == nil {
		return err
	}

	if httpResponse.StatusCode != http.StatusNotFound {
		return err
	}

	_, _, err = configClient.TenantRouteAPI.CreateRoute(ctx, "keycloak").RouteData(config_client.RouteData{
		TenantId:    TenantId,
		RouteConfig: routeConfig,
		RouteId:     RouteId,
	}).Execute()
	return err
}

func createConfigClient(configUrl string) *config_client.APIClient {
	host, scheme := getHostAndScheme(configUrl)

	cfgServer := config_client.NewConfiguration()
	cfgServer.Host = host
	cfgServer.Scheme = scheme

	return config_client.NewAPIClient(cfgServer)
}

func getHostAndScheme(configUrl string) (string, string) {
	parsedUrl, err := url.Parse(configUrl)
	if err != nil {
		return "", ""
	}

	host := parsedUrl.Host
	scheme := parsedUrl.Scheme

	if scheme == "" {
		scheme = "http"
	}

	return host, scheme
}

func buildRouteDataFromConfig() map[string]interface{} {
	return map[string]interface{}{
		"enrich": map[string]interface{}{
			"static":        nil,
			"kube_metadata": nil,
		},
		"metadata":    nil,
		"tenant_id":   TenantId,
		"notify_ids":  NotifyIds,
		"listener_id": ListenerId,
		"filtering": map[string]interface{}{
			"policy": "any",
			"filters": []map[string]interface{}{
				{
					"type": "matcher",
					"matchers": []map[string]interface{}{
						{"key": "category", "value": "keycloak", "match_type": "=="},
						{"key": "type", "value": "reset|welcome", "match_type": "=~"},
					},
					"prometheus_matchers": nil,
				},
			},
		},
	}
}
