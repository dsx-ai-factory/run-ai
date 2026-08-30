package subscription

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/clients"
	subpkg "github.com/run-ai/runai/control-plane/notifications-proxy/internal/subscription"
	proxyapi "github.com/run-ai/runai/control-plane/notifications-proxy/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/notifications-service/pkg/client/types"
)

const (
	http_runai_listener_id = "http-runai"
	Type                   = "type"
	Category               = "category"
	TypesDelimiter         = "|"
	InternalRouteFlag      = "is_internal"
)

type UserRoutesClienter interface {
	CreateRoute(ctx context.Context, tenantId, userId string, route *types.UserRouteConfig) error
	UpdateRoute(ctx context.Context, tenantId, userId string, route *types.UserRouteConfig) error
	DeleteRoute(ctx context.Context, tenantId, userId, routeId string) error
	GetRoutes(ctx context.Context, tenantId, userId string) ([]*types.UserRouteConfig, error)
	DeleteNotifyIdFromAllUserRoutes(ctx context.Context, tenantId string, notifyId string) error
}

type UserClient struct {
	userRoutesClient UserRoutesClienter
}

func NewClient() *UserClient {
	return &UserClient{
		userRoutesClient: clients.NewUserRoutes(),
	}
}

func (uc *UserClient) DisableChannel(ctx context.Context, tenantId string, channelId string) error {
	return uc.userRoutesClient.DeleteNotifyIdFromAllUserRoutes(ctx, tenantId, channelId)
}

func (uc *UserClient) Create(ctx context.Context, tenantId, userId string, subscription subpkg.Subscription) (*subpkg.Subscription, error) {
	rc := BuildRoute(tenantId, userId, &subscription)
	if err := uc.userRoutesClient.CreateRoute(ctx, tenantId, userId, rc); err != nil {
		if errAlreadyExists(err) {
			return nil, errors.Join(err, subpkg.ErrSubscriptionAlreadyExists)
		}

		return nil, errors.Join(err, subpkg.ErrClient)
	}

	subscription.Id = rc.RouteId
	return &subscription, nil
}

func (uc *UserClient) Update(ctx context.Context, tenantId, userId string, subscription subpkg.Subscription) error {
	err := uc.userRoutesClient.UpdateRoute(ctx, tenantId, userId, BuildRoute(tenantId, userId, &subscription))
	if err == nil {
		return nil
	}

	if errNotFound(err) {
		return errors.Join(err, subpkg.ErrNotFound)
	}
	return errors.Join(err, subpkg.ErrClient)
}

func (uc *UserClient) GetUserSubscriptions(ctx context.Context, tenantId, userId string) ([]subpkg.Subscription, error) {
	routes, err := uc.userRoutesClient.GetRoutes(ctx, tenantId, userId)
	if err != nil {
		return nil, errors.Join(err, subpkg.ErrClient)
	}
	return toSubscriptions(getNotInternalRoutes(routes)), nil
}

func (uc *UserClient) Delete(ctx context.Context, tenantId, userId, routeId string) error {
	err := uc.userRoutesClient.DeleteRoute(ctx, tenantId, userId, routeId)
	if err == nil {
		return nil
	}
	if errNotFound(err) {
		return errors.Join(err, subpkg.ErrNotFound)
	}
	return errors.Join(err, subpkg.ErrClient)
}

func getNotInternalRoutes(routes []*types.UserRouteConfig) []*types.UserRouteConfig {
	filteredSubscriptions := []*types.UserRouteConfig{}
	for _, r := range routes {
		if !isInternalSubscription(r) {
			filteredSubscriptions = append(filteredSubscriptions, r)
		}
	}
	return filteredSubscriptions
}

func isInternalSubscription(subscription *types.UserRouteConfig) bool {
	return subscription.Metadata != nil && subscription.Metadata[InternalRouteFlag] == "true"
}

func toSubscriptions(routeConf []*types.UserRouteConfig) []subpkg.Subscription {
	out := make([]subpkg.Subscription, 0, len(routeConf))
	for _, r := range routeConf {
		if r == nil {
			continue
		}
		out = append(out, *toSubscription(*r))
	}
	return out
}

func toSubscription(routeConf types.UserRouteConfig) *subpkg.Subscription {
	subscription := subpkg.Subscription{
		Id:       routeConf.RouteId,
		Channels: routeConf.NotifyIds,
		Events:   []subpkg.SubscriptionEvent{},
	}
	for _, f := range routeConf.Filtering.Filters {
		category, types := getCategoryAndTypes(f)
		if category == "" || len(types) == 0 {
			continue
		}

		subscription.Events = append(subscription.Events, subpkg.SubscriptionEvent{
			Category: category, Types: types,
		})
	}
	return &subscription
}

func getCategoryAndTypes(f *types.FilterConfig) (string, []string) {
	category := ""
	types := []string{}
	if f.Matcher == nil {
		return category, types
	}

	for _, matcher := range *f.Matcher {
		switch matcher.Key {
		case Category:
			category = matcher.Value
		case Type:
			for t := range strings.SplitSeq(matcher.Value, TypesDelimiter) {
				types = append(types, t)
			}
		}
	}

	return category, types
}

func BuildRoute(tenantId string, userId string, subscription *subpkg.Subscription) *types.UserRouteConfig {
	return &types.UserRouteConfig{
		UserId: userId,
		BasicRouteConfig: types.BasicRouteConfig{
			RouteId:    getRouteId(subscription.Id),
			TenantId:   tenantId,
			ListenerId: http_runai_listener_id,
			NotifyIds:  subscription.Channels,
			Filtering: types.FilteringConfig{
				Policy:  types.PolicyMatchAny,
				Filters: buildFilters(subscription.Events),
			},
		},
	}
}

func getRouteId(routeId string) string {
	if routeId != "" {
		return routeId
	}
	return uuid.NewString()
}

func buildFilters(subscriptionEvents []subpkg.SubscriptionEvent) []*types.FilterConfig {
	filters := make([]*types.FilterConfig, 0, len(subscriptionEvents))
	for _, event := range subscriptionEvents {
		filter := &types.FilterConfig{
			Type:    types.FilterTypeMatcher,
			Matcher: buildMatchers(event),
		}

		filters = append(filters, filter)
	}
	return filters
}

func buildMatchers(subscription subpkg.SubscriptionEvent) *types.MatcherFilterConfig {
	return &types.MatcherFilterConfig{&types.PropertyMatcher{
		Key:       "category",
		Value:     subscription.Category,
		MatchType: "==",
	}, &types.PropertyMatcher{
		Key:       "type",
		Value:     strings.Join(subscription.Types, "|"),
		MatchType: "=~",
	}}
}

func errNotFound(err error) bool {
	apiErr, ok := err.(*proxyapi.Error)
	return ok && apiErr.Code == http.StatusNotFound
}

func errAlreadyExists(err error) bool {
	apiErr, ok := err.(*proxyapi.Error)
	return ok && apiErr.Code == http.StatusConflict
}
