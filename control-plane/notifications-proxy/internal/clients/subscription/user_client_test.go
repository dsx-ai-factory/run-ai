package subscription

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/clients/subscription/mocks"
	subpkg "github.com/run-ai/runai/control-plane/notifications-proxy/internal/subscription"
	proxyapi "github.com/run-ai/runai/control-plane/notifications-proxy/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/notifications-service/pkg/client/types"
)

var user = "test@runai.com"

func TestCreateSubscription_Failed(t *testing.T) {
	ctx := context.Background()
	client := mocks.NewUserRoutesClienter(t)
	subscriptionsClient := UserClient{userRoutesClient: client}

	client.On("CreateRoute", ctx, "1", user, mock.Anything).Return(fmt.Errorf("failed to create route"))
	route, err := subscriptionsClient.Create(ctx, "1", user, subpkg.Subscription{
		Id:       "some_id",
		Events:   []subpkg.SubscriptionEvent{},
		Channels: []string{},
	})
	require.Error(t, err)
	require.ErrorContains(t, err, subpkg.ErrClient.Error())
	require.ErrorContains(t, err, "failed to create route")
	require.Nil(t, route)
}

func TestCreateSubscription_AlreadyExists(t *testing.T) {
	ctx := context.Background()
	client := mocks.NewUserRoutesClienter(t)
	subscriptionsClient := UserClient{userRoutesClient: client}

	client.On("CreateRoute", ctx, "1", user, mock.Anything).Return(&proxyapi.Error{Code: http.StatusConflict})
	route, err := subscriptionsClient.Create(ctx, "1", user, subpkg.Subscription{
		Id:       "some_id",
		Events:   []subpkg.SubscriptionEvent{},
		Channels: []string{},
	})
	require.Error(t, err)
	require.ErrorContains(t, err, subpkg.ErrSubscriptionAlreadyExists.Error())
	require.ErrorContains(t, err, "subscription with same id already exist")
	require.Nil(t, route)
}

func TestCreateSubscription_Succeed(t *testing.T) {
	ctx := context.Background()
	client := mocks.NewUserRoutesClienter(t)
	subscriptionsClient := UserClient{userRoutesClient: client}

	client.On("CreateRoute", ctx, "1", user, mock.Anything).Return(nil)
	res, err := subscriptionsClient.Create(ctx, "1", user, subpkg.Subscription{
		Id:       "some_id",
		Events:   []subpkg.SubscriptionEvent{},
		Channels: []string{},
	})
	require.NoError(t, err)
	require.Equal(t, res.Id, "some_id")
}

func TestDeleteSubscription_Failed(t *testing.T) {
	ctx := context.Background()
	id, _ := uuid.NewUUID()
	client := mocks.NewUserRoutesClienter(t)
	subscriptionsClient := UserClient{userRoutesClient: client}

	client.On("DeleteRoute", ctx, "1", user, id.String()).Return(fmt.Errorf("failed to delete"))
	err := subscriptionsClient.Delete(ctx, "1", user, id.String())
	require.Error(t, err)
	require.ErrorContains(t, err, subpkg.ErrClient.Error())
}

func TestDeleteSubscription_Succeed(t *testing.T) {
	ctx := context.Background()
	id, _ := uuid.NewUUID()
	client := mocks.NewUserRoutesClienter(t)
	subscriptionsClient := UserClient{userRoutesClient: client}

	client.On("DeleteRoute", ctx, "1", user, id.String()).Return(nil)
	err := subscriptionsClient.Delete(ctx, "1", user, id.String())
	require.NoError(t, err)
}

func TestUpdateSubscription_Failed(t *testing.T) {
	ctx := context.Background()
	id, _ := uuid.NewUUID()
	client := mocks.NewUserRoutesClienter(t)
	subscriptionsClient := UserClient{userRoutesClient: client}

	client.On("UpdateRoute", ctx, "1", user, mock.Anything).Return(fmt.Errorf("failed to update route"))
	err := subscriptionsClient.Update(ctx, "1", user, subpkg.Subscription{Id: id.String()})
	require.Error(t, err)
	require.ErrorContains(t, err, subpkg.ErrClient.Error())
}

func TestUpdateSubscription_NotFound(t *testing.T) {
	ctx := context.Background()
	id, _ := uuid.NewUUID()
	client := mocks.NewUserRoutesClienter(t)
	subscriptionsClient := UserClient{userRoutesClient: client}

	client.On("UpdateRoute", ctx, "1", user, mock.Anything).Return(&proxyapi.Error{Code: http.StatusNotFound})
	err := subscriptionsClient.Update(ctx, "1", user, subpkg.Subscription{Id: id.String()})
	require.Error(t, err)
	require.ErrorContains(t, err, subpkg.ErrNotFound.Error())
}

func TestUpdateSubscription_Succeed(t *testing.T) {
	ctx := context.Background()
	id, _ := uuid.NewUUID()
	client := mocks.NewUserRoutesClienter(t)
	subscriptionsClient := UserClient{userRoutesClient: client}

	client.On("UpdateRoute", ctx, "1", user, mock.Anything).Return(nil)
	err := subscriptionsClient.Update(ctx, "1", user, subpkg.Subscription{Id: id.String()})
	require.NoError(t, err)
}

func TestGetUserSubscriptions_Failed(t *testing.T) {
	ctx := context.Background()
	client := mocks.NewUserRoutesClienter(t)
	subscriptionsClient := UserClient{userRoutesClient: client}

	client.On("GetRoutes", ctx, "1", user).Return(nil, fmt.Errorf("failed to get routes"))
	routes, err := subscriptionsClient.GetUserSubscriptions(ctx, "1", user)
	require.Error(t, err)
	require.ErrorContains(t, err, subpkg.ErrClient.Error())
	require.Nil(t, routes)
}

func TestGetUserSubscriptions_Succeed(t *testing.T) {
	ctx := context.Background()
	client := mocks.NewUserRoutesClienter(t)
	subscriptionsClient := UserClient{userRoutesClient: client}

	client.On("GetRoutes", ctx, "1", user).Return([]*types.UserRouteConfig{}, nil)
	routes, err := subscriptionsClient.GetUserSubscriptions(ctx, "1", user)
	require.NoError(t, err)
	require.Empty(t, routes)
}

func TestGetUserSubscriptions_ShouldFilterInternalRoutes(t *testing.T) {
	ctx := context.Background()
	client := mocks.NewUserRoutesClienter(t)
	subscriptionsClient := UserClient{userRoutesClient: client}

	client.On("GetRoutes", ctx, "1", user).Return([]*types.UserRouteConfig{
		{BasicRouteConfig: types.BasicRouteConfig{RouteId: "route_1", Metadata: map[string]string{InternalRouteFlag: "true"}}},
		{BasicRouteConfig: types.BasicRouteConfig{RouteId: "route_2", Metadata: map[string]string{InternalRouteFlag: "false"}}},
	}, nil)
	routes, err := subscriptionsClient.GetUserSubscriptions(ctx, "1", user)
	require.NoError(t, err)
	require.Len(t, routes, 1)
	require.Equal(t, "route_2", routes[0].Id)
}

func TestGetUserSubscriptions_ShouldFilterInternalRoutes_WithMetadataNil(t *testing.T) {
	ctx := context.Background()
	client := mocks.NewUserRoutesClienter(t)
	subscriptionsClient := UserClient{userRoutesClient: client}

	client.On("GetRoutes", ctx, "1", user).Return([]*types.UserRouteConfig{
		{BasicRouteConfig: types.BasicRouteConfig{RouteId: "route_1", Metadata: map[string]string{InternalRouteFlag: "true"}}},
		{BasicRouteConfig: types.BasicRouteConfig{RouteId: "route_2"}},
	}, nil)
	routes, err := subscriptionsClient.GetUserSubscriptions(ctx, "1", user)
	require.NoError(t, err)
	require.Len(t, routes, 1)
	require.Equal(t, "route_2", routes[0].Id)
}

func TestToRouteConfig(t *testing.T) {
	subscription := &subpkg.Subscription{
		Id:       "test-123-user1",
		Channels: []string{"test"},
		Events: []subpkg.SubscriptionEvent{
			{
				Category: "category1",
				Types:    []string{"type1", "type2"},
			},
			{
				Category: "category2",
				Types:    []string{"type1", "type3"},
			},
		},
	}

	route := BuildRoute("123", "user1", subscription)

	expectedRoute := types.UserRouteConfig{
		UserId: "user1",
		BasicRouteConfig: types.BasicRouteConfig{
			RouteId:    "test-123-user1",
			ListenerId: http_runai_listener_id,
			TenantId:   "123",
			NotifyIds:  []string{"test"},
			Filtering: types.FilteringConfig{
				Policy: "any",
				Filters: []*types.FilterConfig{
					{
						Type: "matcher",
						Matcher: &types.MatcherFilterConfig{
							{
								Key:       "category",
								Value:     "category1",
								MatchType: "==",
							},
							{
								Key:       "type",
								Value:     "type1|type2",
								MatchType: "=~",
							},
						},
					},
					{
						Type: "matcher",
						Matcher: &types.MatcherFilterConfig{
							{
								Key:       "category",
								Value:     "category2",
								MatchType: "==",
							},
							{
								Key:       "type",
								Value:     "type1|type3",
								MatchType: "=~",
							},
						},
					},
				},
			},
		},
	}
	assert.Equal(t, expectedRoute, *route)
}

func TestToRouteConfig_NoId(t *testing.T) {
	subscription := &subpkg.Subscription{
		Channels: []string{"test"},
		Events: []subpkg.SubscriptionEvent{
			{
				Category: "category1",
				Types:    []string{"type1", "type2"},
			},
			{
				Category: "category2",
				Types:    []string{"type1", "type3"},
			},
		},
	}

	route := BuildRoute("123", "user1", subscription)
	routeId := route.RouteId
	_, err := uuid.Parse(routeId)
	assert.NoError(t, err)
	expectedRoute := types.UserRouteConfig{
		UserId: "user1",
		BasicRouteConfig: types.BasicRouteConfig{
			RouteId:    routeId,
			ListenerId: http_runai_listener_id,
			TenantId:   "123",
			NotifyIds:  []string{"test"},
			Filtering: types.FilteringConfig{
				Policy: "any",
				Filters: []*types.FilterConfig{
					{
						Type: "matcher",
						Matcher: &types.MatcherFilterConfig{
							{
								Key:       "category",
								Value:     "category1",
								MatchType: "==",
							},
							{
								Key:       "type",
								Value:     "type1|type2",
								MatchType: "=~",
							},
						},
					},
					{
						Type: "matcher",
						Matcher: &types.MatcherFilterConfig{
							{
								Key:       "category",
								Value:     "category2",
								MatchType: "==",
							},
							{
								Key:       "type",
								Value:     "type1|type3",
								MatchType: "=~",
							},
						},
					},
				},
			},
		},
	}
	assert.Equal(t, expectedRoute, *route)
}
