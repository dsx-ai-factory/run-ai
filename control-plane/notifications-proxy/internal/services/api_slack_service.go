package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"

	audit "github.com/run-ai/runai/control-plane/audit-service/pkg/client"
	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/clients/tenant_migration"
	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/notification_state"

	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/clients/channel"
	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/clients/notify"
	slackclient "github.com/run-ai/runai/control-plane/notifications-proxy/internal/clients/slack"
	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/clients/state"
	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/clients/subscription"
	proxy_config "github.com/run-ai/runai/control-plane/notifications-proxy/internal/config"
	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/database/dao"
	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/slack"
	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

type SlackManager interface {
	AuthorizeApp(ctx context.Context, code string, state string) error
	CreateApp(ctx context.Context, accessToken string) (string, error)
	DeleteApp(ctx context.Context) error
	ConnectWorkspace(ctx context.Context, tenantId string, subjectEmail string) (string, error)
	DisconnectWorkspace(ctx context.Context, tenantId string) error
	GetSlackConfig(ctx context.Context, tenantId string) (*slack.ChannelConfig, error)
	PatchSlackConfig(ctx context.Context, tenantId string, cfg slack.ChannelConfig) (*slack.ChannelConfig, error)
	ValidateSlack(ctx context.Context, cfg slack.ChannelConfig) (slack.ValidationResponse, error)
}

type SlackNotificationStateManager interface {
	GetSlackNotificationState(ctx context.Context, tenantId string) (notification_state.Stater, error)
	GetDetailedSlackNotificationState(ctx context.Context, tenantId string) (notification_state.Stater, error)
	UpdateSlackNotificationState(ctx context.Context, tenantId string, notificationState notification_state.Stater) (notification_state.Stater, error)
}

type SlackAPIService struct {
	manager      SlackManager
	stateManager SlackNotificationStateManager
}

func NewSlackAPIService() api.SlackAPIServicer {
	conf := config.Get[*proxy_config.ServiceConfig]()
	slackClient := slackclient.NewSlackClient(conf.RunaiEnvDomain)
	stateClient := state.NewClient()
	notifier := notify.NewNotifier()
	slackAuthSessionHandler := slack.NewSlackAuthSessionHandler()
	channelClient := channel.NewClient()
	slackCredentialsDao := dao.NewSlackCredentials()
	slackManager := slack.NewSlackManager(stateClient, channelClient, slackClient, slackAuthSessionHandler, notifier, slackCredentialsDao, subscription.NewClient())

	err := slackManager.SyncSecrets(context.Background())
	if err != nil {
		logging.Errorf("Failed to sync slack secrets: %v", err)
	}
	return &SlackAPIService{
		manager:      slackManager,
		stateManager: notification_state.NewNotificationStateManager(stateClient, channelClient, subscription.NewClient(), tenant_migration.NewClient()),
	}
}

func (ss *SlackAPIService) AuthorizeSlackApp(ctx context.Context, code string, state string) (api.ImplResponse, error) {
	err := ss.manager.AuthorizeApp(ctx, code, state)
	if err == nil {
		return api.Response(http.StatusOK, "Slack app authorized successfully. You can now return to the Runai platform."), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to authorize slack app: %v", err)
	if errors.Is(err, slack.ErrMissingAppCredentials) {
		return api.ImplResponse{Code: http.StatusBadRequest}, fmt.Errorf("missing slack app credentials")
	}
	if errors.Is(err, slack.ErrExpiredState) {
		return api.ImplResponse{Code: http.StatusBadRequest}, fmt.Errorf("slack state expired")
	}
	if errors.Is(err, slack.ErrInvalidState) {
		return api.ImplResponse{Code: http.StatusBadRequest}, fmt.Errorf("invalid slack state")
	}
	return api.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("failed to authorize slack app")
}

func (ss *SlackAPIService) CreateSlackApp(ctx context.Context, body api.CreateSlackAppRequestBody) (api.ImplResponse, error) {
	ss.addDataToAudit(ctx, "Create")
	if !isAccessTokenValid(body.AccessToken) {
		return api.ImplResponse{Code: http.StatusBadRequest}, errors.New("accessToken is invalid, must be a Slack token format (e.g., xoxp-123... or xoxe.xoxp-123...)")
	}

	appUrl, err := ss.manager.CreateApp(ctx, body.AccessToken)
	if err == nil {
		return api.Response(http.StatusOK, api.CreateSlackAppResponse{SlackAppUrl: appUrl}), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to create slack app: %v", err)
	return api.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("failed to create slack app")
}

func (ss *SlackAPIService) DeleteSlackApp(ctx context.Context) (api.ImplResponse, error) {
	ss.addDataToAudit(ctx, "Delete")
	err := ss.manager.DeleteApp(ctx)
	if err == nil {
		return api.ImplResponse{Code: http.StatusOK}, nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to delete slack app, error %v", err)
	return api.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("failed to delete slack app")
}

func (ss *SlackAPIService) ConnectSlackWorkspace(ctx context.Context) (api.ImplResponse, error) {
	ss.addDataToAudit(ctx, "Connect")
	tenantId, err := runaictx.ParseTenantId(ctx)
	if err != nil {
		return api.ImplResponse{Code: http.StatusBadRequest}, fmt.Errorf("missing tenant id in request context")
	}
	subjectEmail, err := runaictx.ParseSubjectId(ctx)
	if err != nil {
		return api.ImplResponse{Code: http.StatusBadRequest}, fmt.Errorf("missing subject id in request context")
	}

	redirectUrl, err := ss.manager.ConnectWorkspace(ctx, tenantId, subjectEmail)
	if err == nil {
		return api.Response(http.StatusOK, api.ConnectSlackWorkspace200Response{
			RedirectUrl: redirectUrl,
		}), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to add slack app: %v", err)
	if errors.Is(err, slack.ErrMissingAppCredentials) {
		return api.ImplResponse{Code: http.StatusBadRequest}, fmt.Errorf("missing slack app credentials")
	}
	return api.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("failed to connect slack workspace")
}

func (ss *SlackAPIService) DisconnectSlackWorkspace(ctx context.Context) (api.ImplResponse, error) {
	ss.addDataToAudit(ctx, "Disconnect")
	tenantId, err := runaictx.ParseTenantId(ctx)
	if err != nil {
		return api.ImplResponse{Code: http.StatusBadRequest}, fmt.Errorf("missing tenant id in request context")
	}

	err = ss.manager.DisconnectWorkspace(ctx, tenantId)
	if err == nil {
		return api.ImplResponse{Code: http.StatusOK}, nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to disconnect workspace: %v", err)
	return api.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("failed to disconnect workspace")
}

func (ss *SlackAPIService) GetSlack(ctx context.Context) (api.ImplResponse, error) {
	tenantId, err := runaictx.ParseTenantId(ctx)
	if err != nil {
		return api.ImplResponse{Code: http.StatusBadRequest}, fmt.Errorf("missing tenant id in request context")
	}

	stackConfig, err := ss.manager.GetSlackConfig(ctx, tenantId)
	if err == nil {
		return api.Response(http.StatusOK, toApi(stackConfig)), nil
	}
	if errors.Is(err, slack.ErrNotFound) {
		return api.ImplResponse{Code: http.StatusNotFound}, fmt.Errorf("slack app not found")
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get slack app: %v", err)
	return api.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("failed to get slack app")
}

func (ss *SlackAPIService) PatchSlack(ctx context.Context, request api.SlackNotificationChannelPatchRequest) (api.ImplResponse, error) {
	tenantId, err := runaictx.ParseTenantId(ctx)
	if err != nil {
		return api.ImplResponse{Code: http.StatusBadRequest}, fmt.Errorf("missing tenant id in request context")
	}

	updated, err := ss.manager.PatchSlackConfig(ctx, tenantId, toSlack(request.Config))
	if err == nil {
		return api.Response(http.StatusOK, toApi(updated)), nil
	}

	if errors.Is(err, slack.ErrNotFound) {
		return api.ImplResponse{Code: http.StatusNotFound}, fmt.Errorf("slack app not found")
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to patch slack app: %v", err)
	return api.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("failed to patch slack app")
}

func (sas *SlackAPIService) ValidateSlack(ctx context.Context, request api.SlackNotificationChannelConfig) (api.ImplResponse, error) {
	resp, err := sas.manager.ValidateSlack(ctx, slack.ChannelConfig{
		ApiToken:            request.ApiToken,
		DirectNotifications: request.DirectNotifications,
		Recipients:          request.Recipients,
	})
	if err == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("successfully validated slack")
		return api.ImplResponse{Code: http.StatusOK, Body: api.ValidateSinkResponse{Status: resp.Status, Message: resp.Message}}, nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("slack validation failed %v", err)
	if errors.Is(err, slack.ErrValidationFailed) {
		return api.ImplResponse{Code: http.StatusOK, Body: api.ValidateSinkResponse{Status: resp.Status, Message: resp.Message}}, nil
	}
	return api.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("slack validation failed")
}

func (sas *SlackAPIService) GetSlackNotificationState(ctx context.Context) (api.ImplResponse, error) {
	return sas.getSlackNotificationState(ctx, false)
}

func (sas *SlackAPIService) GetDetailedSlackNotificationState(ctx context.Context) (api.ImplResponse, error) {
	return sas.getSlackNotificationState(ctx, true)
}

func (sas *SlackAPIService) UpdateSlackNotificationState(ctx context.Context, status api.NotificationStateStatus) (api.ImplResponse, error) {
	tenantId, err := runaictx.ParseTenantId(ctx)
	if err != nil {
		return api.ImplResponse{Code: http.StatusBadRequest}, fmt.Errorf("missing tenant id in request context")
	}

	res, err := sas.stateManager.UpdateSlackNotificationState(ctx, tenantId, &state.SlackState{TenantId: tenantId, Enabled: status.Enabled})
	if err == nil {
		return api.ImplResponse{Code: http.StatusOK, Body: getAPIState(res)}, nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update slack notifications state: %v", err)
	if errors.Is(err, notification_state.ErrFailedToRemoveSinkFromUserRoutes) {
		return api.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("failed to remove slack sink from user routes")
	}
	if errors.Is(err, notification_state.ErrNoSinkConfigured) {
		return api.ImplResponse{Code: http.StatusNotFound}, fmt.Errorf("slack app not configured to update state")
	}
	return api.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("failed to update slack notifications state")
}

// ------------------------------------------- START DEPRECATED METHODS ------------------------------------------- //
func (ss *SlackAPIService) CreateSlackAppDeprecated(ctx context.Context, body api.CreateSlackAppRequestBody) (api.ImplResponse, error) {
	return ss.CreateSlackApp(ctx, body)
}

func (ss *SlackAPIService) AuthorizeSlackAppDeprecated(ctx context.Context, code string, state string) (api.ImplResponse, error) {
	return ss.AuthorizeSlackApp(ctx, code, state)
}

func (ss *SlackAPIService) AddSlackAppDeprecated(ctx context.Context) (api.ImplResponse, error) {
	return ss.ConnectSlackWorkspace(ctx)
}

// ------------------------------------------- END DEPRECATED METHODS ------------------------------------------- //

func (sas *SlackAPIService) getSlackNotificationState(ctx context.Context, withDetails bool) (api.ImplResponse, error) {
	tenantId, err := runaictx.ParseTenantId(ctx)
	if err != nil {
		return api.ImplResponse{Code: http.StatusBadRequest}, fmt.Errorf("missing tenant id in request context")
	}
	stateGetter := sas.stateManager.GetSlackNotificationState
	if withDetails {
		stateGetter = sas.stateManager.GetDetailedSlackNotificationState
	}
	sns, err := stateGetter(ctx, tenantId)
	if err == nil || errors.Is(err, notification_state.ErrFailedToRemoveSinkFromUserRoutes) {
		if withDetails {
			return api.ImplResponse{Code: http.StatusOK, Body: getAPISlackStateWithDetails(sns)}, nil
		}
		return api.ImplResponse{Code: http.StatusOK, Body: getAPIState(sns)}, nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get slack notifications state: %v", err)
	if errors.Is(err, notification_state.ErrNoSinkConfigured) {
		return api.ImplResponse{Code: http.StatusNotFound}, fmt.Errorf("slack app not configured to update state")
	}
	return api.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("failed to get slack state")
}

func (*SlackAPIService) addDataToAudit(ctx context.Context, action string) *audit.AuditLog {
	tenantId, _ := runaictx.ParseTenantId(ctx)
	return audit.AuditFromCtx(ctx).WithEntityType("Notification channel").WithEntityName("Slack").WithEntityId(channel.SlackChannelName(tenantId)).WithAction(action)
}

func isAccessTokenValid(accessToken string) bool {
	// Pattern: optional "xoxe." prefix, then "xoxp-", then alphanumeric/hyphens
	regex := regexp.MustCompile(`^(xoxe\.)?xoxp-[a-zA-Z0-9-]+$`)
	return regex.MatchString(accessToken)
}

func toApi(channel *slack.ChannelConfig) api.SlackNotificationChannel {
	return api.SlackNotificationChannel{Config: api.SlackNotificationChannelConfig{
		DirectNotifications: channel.DirectNotifications,
		Recipients:          channel.Recipients,
	},
	}
}

func toSlack(config api.SlackNotificationChannelPatchRequestConfig) slack.ChannelConfig {
	return slack.ChannelConfig{
		DirectNotifications: config.DirectNotifications,
		Recipients:          config.Recipients,
	}
}
