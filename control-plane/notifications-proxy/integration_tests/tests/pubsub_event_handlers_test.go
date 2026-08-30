//go:build integration
// +build integration

package tests

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	openapi "github.com/run-ai/runai/control-plane/notifications-proxy/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/notifications-service/pkg/client/generated/config_client"
	async_events "github.com/run-ai/runai/runai-application-packages/async-events"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub"
	"github.com/stretchr/testify/require"
)

func TestTenantDeleteEventHandler_ShouldDeleteTenantRouteAndUserRoutes(t *testing.T) {
	routeId := uuid.NewString()
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdWithSystemScope)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, tenantIdWithSystemScope)

	_, _, err := notificationsClient.TenantRouteAPI.CreateRoute(ctx, tenantIdWithSystemScope).RouteData(config_client.RouteData{
		RouteId:  routeId,
		TenantId: tenantIdWithSystemScope,
		RouteConfig: map[string]interface{}{
			"listener_id": "runai-http",
		},
	}).Execute()
	require.NoError(t, err)
	defer notificationsClient.TenantRouteAPI.DeleteRoute(ctx, tenantIdWithSystemScope, routeId).Execute()

	tenantRoutes, _, err := notificationsClient.TenantRouteAPI.GetRoutes(ctx, tenantIdWithSystemScope).Execute()
	require.NoError(t, err)
	require.Equal(t, 1, len(tenantRoutes))
	require.Equal(t, tenantRoutes[0].RouteId, routeId)

	_, _, err = client.SubscriptionsAPI.CreateSubscription(ctx).Subscription(openapi.Subscription{
		Events:   []openapi.SubscriptionEvent{},
		Channels: []string{"some-channel"},
		Id:       &routeId,
	}).Execute()
	require.NoError(t, err)
	defer client.SubscriptionsAPI.DeleteSubscription(ctx, routeId).Execute()

	subscriptions, _, err := client.SubscriptionsAPI.GetUserSubscriptions(ctx).Execute()
	require.NoError(t, err)
	require.Greater(t, len(subscriptions), 0)

	conf := openapi.NewNotificationChannelConfigWithDefaults()
	conf.SetApiToken("xoxb-1234567890")
	response, err := client.NotificationChannelsAPI.CreateNotificationChannel(ctx).NotificationChannel(openapi.NotificationChannel{
		Type:   "slack",
		Config: *conf,
	}).Execute()
	require.NoError(t, err)
	defer client.NotificationChannelsAPI.DeleteNotificationChannel(ctx, fmt.Sprintf("slack-%s", tenantIdWithSystemScope)).Execute()
	require.Equal(t, http.StatusOK, response.StatusCode)

	channel, _, err := client.NotificationChannelsAPI.GetNotificationChannel(ctx, fmt.Sprintf("slack-%s", tenantIdWithSystemScope)).Execute()
	require.NoError(t, err)
	require.NotNil(t, channel)

	err = runaipubsub.Publish(ctx, async_events.NewTenantDelete(tenantIdWithSystemScope))
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		tenantRoutes, _, err := notificationsClient.TenantRouteAPI.GetRoutes(ctx, tenantIdWithSystemScope).Execute()
		return err == nil && len(tenantRoutes) == 0
	}, time.Second*60, time.Second*1)

	require.Eventually(t, func() bool {
		subscriptions, _, err = client.SubscriptionsAPI.GetUserSubscriptions(ctx).Execute()
		return err == nil && len(subscriptions) == 0
	}, time.Second*60, time.Second*1)

	require.Eventually(t, func() bool {
		channel, resp, err := client.NotificationChannelsAPI.GetNotificationChannel(ctx, fmt.Sprintf("slack-%s", tenantIdWithSystemScope)).Execute()
		return err != nil && resp.StatusCode == http.StatusNotFound && channel == nil
	}, time.Second*60, time.Second*1)
}
