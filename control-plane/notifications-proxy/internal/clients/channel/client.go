package channel

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/channel"
	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/clients"
	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/email"
	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/notification_state"
	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/slack"
	proxyapi "github.com/run-ai/runai/control-plane/notifications-proxy/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/notifications-service/pkg/client/types"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

const (
	NotificationChannelTypeEmail = "email"
	NotificationChannelTypeSlack = "slack"
)

var (
	ErrClientFailure    = fmt.Errorf("client error")
	ErrAlreadyExists    = fmt.Errorf("channel already exist")
	ErrNotFound         = fmt.Errorf("channel not found")
	ErrFailedToParse    = fmt.Errorf("failed to parse server response")
	ErrNotSupportedType = fmt.Errorf("channel not supported type")
	ErrValidationFailed = fmt.Errorf("validation failed")

	errToChannelErrMapping = map[error]error{
		ErrNotFound:         channel.ErrNotFound,
		ErrClientFailure:    channel.ErrClientFailure,
		ErrAlreadyExists:    channel.ErrAlreadyExists,
		ErrFailedToParse:    channel.ErrFailedToParse,
		ErrNotSupportedType: channel.ErrNotSupportedType,
		ErrValidationFailed: channel.ErrValidationFailed,
	}

	errToEmailErrMapping = map[error]error{
		ErrNotFound:         email.ErrNotFound,
		ErrClientFailure:    email.ErrClient,
		ErrAlreadyExists:    email.ErrAlreadyExists,
		ErrFailedToParse:    email.ErrFailedToParse,
		ErrValidationFailed: email.ErrValidationFailed,
	}

	errToSlackErrMapping = map[error]error{
		ErrNotFound:         slack.ErrNotFound,
		ErrClientFailure:    slack.ErrClientFailure,
		ErrAlreadyExists:    slack.ErrAlreadyExists,
		ErrFailedToParse:    slack.ErrFailedToParse,
		ErrValidationFailed: slack.ErrValidationFailed,
	}
)

type SinksClienter interface {
	GetSink(ctx context.Context, channelId string) (*types.SinkConfig, error)
	GetSinks(ctx context.Context) ([]*types.SinkConfig, error)
	DeleteSink(ctx context.Context, channelId string) error
	UpdateSink(ctx context.Context, channelId string, config *types.SinkConfig) error
	CreateSink(ctx context.Context, channelId string, config *types.SinkConfig) error
	ValidateSink(ctx context.Context, channelId string, config *types.SinkConfig) (*clients.SinkValidityResponse, error)
}

func EmailChannelName() string {
	return "runai-email"
}

func SlackChannelName(tenantId string) string {
	return fmt.Sprintf("slack-%s", tenantId)
}

type ChannelClient struct {
	sinksClient SinksClienter
}

func NewClient() *ChannelClient {
	return &ChannelClient{
		sinksClient: clients.NewSinks(),
	}
}

// ------------------------------------------- START Implements NotificationState's ChannelClienter --------------------------
func (cc *ChannelClient) IsExists(ctx context.Context, tenantId string, state notification_state.Stater) (bool, error) {
	sc, err := cc.sinksClient.GetSink(ctx, state.GetChanName())
	if err != nil {
		if errNotFound(err) {
			return false, nil
		}

		return false, errors.Join(fmt.Errorf("failed to get sink config %s for tenant %s: %w", state.GetChanName(), tenantId, err), notification_state.ErrClientFailure)
	}
	if sc == nil {
		return false, nil
	}

	if state.GetType() != sc.Type {
		return false, errors.Join(fmt.Errorf("sink %s of type %s does not match sink of type %s, tenant %s", state.GetChanName(), sc.Type, state.GetType(), tenantId), notification_state.ErrIncompatibleSinkType)
	}

	return true, nil
}

// ------------------------------------------- END Implements NotificationState's ChannelClienter --------------------------

// ------------------------------------------- START Implements Slack's ChannelClienter --------------------------

func (cc *ChannelClient) CreateSlack(ctx context.Context, tenantId string, channel slack.ChannelConfig) (*slack.ChannelConfig, error) {
	channel.Id = SlackChannelName(tenantId)
	if err := cc.create(ctx, channel.Id, slackChannelToSinkConfig(channel), errToSlackErrMapping); err != nil {
		return nil, err
	}

	return &channel, nil
}

func (cc *ChannelClient) DeleteSlack(ctx context.Context, tenantId string) (*slack.ChannelConfig, error) {
	if err := cc.delete(ctx, SlackChannelName(tenantId), errToSlackErrMapping); err != nil {
		return nil, err
	}
	return &slack.ChannelConfig{Id: SlackChannelName(tenantId)}, nil
}

func (cc *ChannelClient) GetSlack(ctx context.Context, tenantId string) (*slack.ChannelConfig, error) {
	cfg, err := cc.getSlack(ctx, tenantId, errToSlackErrMapping)
	if err != nil {
		return nil, err
	}

	return &cfg.Slack, nil
}

func (cc *ChannelClient) PatchSlack(ctx context.Context, tenantId string, channel slack.ChannelConfig) (*slack.ChannelConfig, error) {
	sink, err := cc.patchSlack(ctx, tenantId, channel, errToSlackErrMapping)
	if err != nil {
		return nil, err
	}

	return sink, nil
}

func (cc *ChannelClient) ValidateSlack(ctx context.Context, config slack.ChannelConfig) (slack.ValidationResponse, error) {
	valid, err := cc.validate(ctx, slackChannelToSinkConfig(config), errToSlackErrMapping)
	if err != nil {
		return slack.ValidationResponse{}, err
	}

	return slack.ValidationResponse{Status: valid.Status, Message: valid.Message}, nil
}

// ------------------------------------------- END Implements Slack's ChannelClienter --------------------------

// ------------------------------------------- START Implements Channels Clienter --------------------------
func (cc *ChannelClient) GetTenantChannels(ctx context.Context, tenantId string) ([]channel.NotificationChannelMetadata, error) {
	cfgs, err := cc.sinksClient.GetSinks(ctx)
	if err != nil {
		return nil, errors.Join(err, channel.ErrClientFailure)
	}

	return filterPermittedConfigs(cfgs, tenantId), nil
}

func (cc *ChannelClient) Get(ctx context.Context, tenantId, channelId string) (*channel.NotificationChannel, error) {
	if channelId == SlackChannelName(tenantId) {
		return cc.getSlack(ctx, tenantId, errToChannelErrMapping)
	}

	if channelId == EmailChannelName() {
		return cc.getEmail(ctx, errToChannelErrMapping)
	}

	return nil, errors.Join(fmt.Errorf("unknown channel id: %s", channelId), channel.ErrNotSupportedType)
}

func (cc *ChannelClient) Create(ctx context.Context, tenantId string, cfg channel.NotificationChannel) error {
	if cfg.Type == NotificationChannelTypeEmail {
		return cc.create(ctx, EmailChannelName(), emailConfigToSinkConfig(cfg.Email), errToChannelErrMapping)
	}

	if cfg.Type == NotificationChannelTypeSlack {
		return cc.create(ctx, SlackChannelName(tenantId), slackChannelToSinkConfig(cfg.Slack), errToChannelErrMapping)
	}

	return errors.Join(fmt.Errorf("unsupported channel type: %s", cfg.Type), channel.ErrNotSupportedType)
}

func (cc *ChannelClient) Delete(ctx context.Context, tenantId, channelId string) error {
	if channelId == EmailChannelName() {
		return cc.delete(ctx, EmailChannelName(), errToChannelErrMapping)
	}

	if channelId == SlackChannelName(tenantId) {
		return cc.delete(ctx, SlackChannelName(tenantId), errToChannelErrMapping)
	}

	return errors.Join(fmt.Errorf("channel id not recognized: %s", channelId), channel.ErrChannelNameNotValid)
}

func (cc *ChannelClient) Update(ctx context.Context, tenantId, channelId string, config channel.NotificationChannel) error {
	if channelId == EmailChannelName() {
		return cc.update(ctx, EmailChannelName(), emailConfigToSinkConfig(config.Email), errToChannelErrMapping)
	}

	if channelId == SlackChannelName(tenantId) {
		return cc.update(ctx, SlackChannelName(tenantId), slackChannelToSinkConfig(config.Slack), errToChannelErrMapping)
	}

	return channel.ErrNotSupportedType
}

func (cc *ChannelClient) Patch(ctx context.Context, tenantId, channelId string, config channel.NotificationChannel) error {
	if channelId == EmailChannelName() {
		_, err := cc.patchEmail(ctx, config.Email, errToChannelErrMapping)
		if err != nil {
			return err
		}

		return nil
	}

	if channelId == SlackChannelName(tenantId) {
		_, err := cc.patchSlack(ctx, tenantId, config.Slack, errToChannelErrMapping)
		if err != nil {
			return err
		}

		return nil
	}

	return channel.ErrNotSupportedType
}

func (cc *ChannelClient) Validate(ctx context.Context, tenantId string, cfg channel.NotificationChannel) (channel.ValidationResponse, error) {
	if cfg.Type == NotificationChannelTypeSlack {
		return cc.validate(ctx, slackChannelToSinkConfig(cfg.Slack), errToChannelErrMapping)
	} else if cfg.Type != NotificationChannelTypeEmail {
		return channel.ValidationResponse{}, channel.ErrNotSupportedType
	}

	return cc.validate(ctx, emailConfigToSinkConfig(cfg.Email), errToChannelErrMapping)
}

func (cc *ChannelClient) getEmail(ctx context.Context, errMapper map[error]error) (*channel.NotificationChannel, error) {
	ec, err := cc.get(ctx, EmailChannelName(), errMapper)
	if err != nil {
		return nil, err
	}

	return &channel.NotificationChannel{Id: EmailChannelName(), Type: NotificationChannelTypeEmail, Email: *emailChannelConfig(ec.Config)}, nil
}

func (cc *ChannelClient) getSlack(ctx context.Context, tenantId string, errMapper map[error]error) (*channel.NotificationChannel, error) {
	sink, err := cc.get(ctx, SlackChannelName(tenantId), errMapper)
	if err != nil {
		return nil, err
	}

	return &channel.NotificationChannel{Type: NotificationChannelTypeSlack, Slack: *slackChannelConfig(sink.Config)}, nil
}

func filterPermittedConfigs(configs []*types.SinkConfig, tenantId string) []channel.NotificationChannelMetadata {
	permittedConfigs := []channel.NotificationChannelMetadata{}
	for _, config := range configs {
		if config.Id == EmailChannelName() || config.Id == SlackChannelName(tenantId) {
			permittedConfigs = append(permittedConfigs, sinkConfigToNotificationChannel(config))
		}
	}
	return permittedConfigs
}

// ------------------------------------------- END Implements Channels Clienter ----------------------------------

// ------------------------------------------- START Implements Email Clienter  ----------------------------------

func (cc *ChannelClient) CreateEmail(ctx context.Context, config email.ChannelConfig) error {
	return cc.create(ctx, EmailChannelName(), emailConfigToSinkConfig(config), errToEmailErrMapping)
}

func (cc *ChannelClient) GetEmail(ctx context.Context) (*email.ChannelConfig, error) {
	sink, err := cc.getEmail(ctx, errToEmailErrMapping)
	if err != nil {
		return nil, err
	}

	return &sink.Email, nil
}

func (cc *ChannelClient) DeleteEmail(ctx context.Context) error {
	return cc.delete(ctx, EmailChannelName(), errToEmailErrMapping)
}

func (cc *ChannelClient) UpdateEmail(ctx context.Context, config email.ChannelConfig) error {
	return cc.update(ctx, EmailChannelName(), emailConfigToSinkConfig(config), errToEmailErrMapping)
}

func (cc *ChannelClient) ValidateEmail(ctx context.Context, config email.ChannelConfig) (email.ValidationResponse, error) {
	resp, err := cc.validate(ctx, emailConfigToSinkConfig(config), errToEmailErrMapping)
	if err == nil {
		return email.ValidationResponse{Status: true}, nil
	}
	return email.ValidationResponse{Status: resp.Status, Message: resp.Message}, err
}

func (cc *ChannelClient) PatchEmail(ctx context.Context, config email.ChannelConfig) (*email.ChannelConfig, error) {
	return cc.patchEmail(ctx, config, errToEmailErrMapping)
}

// ------------------------------------------- END Implements Email Clienter  ----------------------------------

// ------------------------------------------- 		 START Internal 		  ----------------------------------

func (cc *ChannelClient) create(ctx context.Context, channelId string, cfg *types.SinkConfig, errMapper map[error]error) error {
	channelConfig, err := cc.sinksClient.GetSink(ctx, channelId) // TODO remove
	if err != nil && !errNotFound(err) {
		logging.Errorf("failed to get sink config %s: %v", channelId, err)
		return errors.Join(err, errMapper[ErrClientFailure])
	}
	if channelConfig != nil {
		logging.Errorf("sink config %s already exists", channelId)
		return errMapper[ErrAlreadyExists]
	}

	if err := cc.sinksClient.CreateSink(ctx, channelId, cfg); err != nil {
		logging.Errorf("failed to create sink config %s: %v", channelId, err)
		if errConflict(err) {
			return errMapper[ErrAlreadyExists]
		}
		return errors.Join(err, errMapper[ErrClientFailure])
	}
	return nil
}

func (cc *ChannelClient) get(ctx context.Context, channelId string, errMapper map[error]error) (*types.SinkConfig, error) {
	sink, err := cc.sinksClient.GetSink(ctx, channelId)
	if err != nil {
		logging.Errorf("failed to get sink config %s: %v", channelId, err)
		if errNotFound(err) {
			return nil, errMapper[ErrNotFound]
		}
		return nil, errors.Join(err, errMapper[ErrClientFailure])
	}
	if sink == nil {
		return nil, errMapper[ErrNotFound]
	}

	return sink, nil
}

func (cc *ChannelClient) update(ctx context.Context, channelId string, cfg *types.SinkConfig, errMapper map[error]error) error {
	if err := cc.sinksClient.DeleteSink(ctx, channelId); err != nil {
		if errNotFound(err) {
			return errMapper[ErrNotFound]
		}

		return errors.Join(err, errMapper[ErrClientFailure])
	}

	if err := cc.sinksClient.CreateSink(ctx, channelId, cfg); err != nil {
		return errors.Join(err, errMapper[ErrClientFailure])
	}
	return nil
}

func (cc *ChannelClient) delete(ctx context.Context, channelId string, errMapper map[error]error) error {
	if err := cc.sinksClient.DeleteSink(ctx, channelId); err != nil {
		if errNotFound(err) {
			return errMapper[ErrNotFound]
		}

		return errors.Join(err, errMapper[ErrClientFailure])
	}

	return nil
}

func (cc *ChannelClient) validate(ctx context.Context, sink *types.SinkConfig, errMapper map[error]error) (channel.ValidationResponse, error) {
	res, err := cc.sinksClient.ValidateSink(ctx, "sink-validation", sink)
	if err != nil || res == nil {
		return channel.ValidationResponse{}, errors.Join(err, errMapper[ErrClientFailure])
	}

	return inferSinkValidity(*res, errMapper)
}

func (cc *ChannelClient) patchSlack(ctx context.Context, tenantId string, config slack.ChannelConfig, errMapper map[error]error) (*slack.ChannelConfig, error) {
	newConf, err := cc.patch(ctx, SlackChannelName(tenantId), slackChannelToSinkConfig(config), errMapper)
	if err != nil {
		return nil, err
	}
	return slackChannelConfig(*newConf), nil
}

func (cc *ChannelClient) patchEmail(ctx context.Context, config email.ChannelConfig, errMapper map[error]error) (*email.ChannelConfig, error) {
	newConf, err := cc.patch(ctx, EmailChannelName(), emailConfigToSinkConfig(config), errMapper)
	if err != nil {
		return nil, err
	}
	return emailChannelConfig(*newConf), nil
}

func (cc *ChannelClient) patch(ctx context.Context, channelId string, config *types.SinkConfig, errMapper map[error]error) (*types.SinkConfigDetails, error) {
	sinkConfig, err := cc.get(ctx, channelId, errMapper)
	if err != nil {
		return nil, err
	}

	mergedConfig := merge(sinkConfig.Config, config.Config)
	if err := cc.sinksClient.UpdateSink(ctx, channelId, &types.SinkConfig{Id: channelId, Type: config.Type, Config: mergedConfig}); err != nil {
		return nil, errors.Join(err, errMapper[ErrClientFailure])
	}

	return &mergedConfig, nil
}

func merge(currentConfig types.SinkConfigDetails, overrideConfig types.SinkConfigDetails) types.SinkConfigDetails {
	mergedConfig := currentConfig
	mergeFields(reflect.ValueOf(&mergedConfig).Elem(), reflect.ValueOf(overrideConfig))
	return mergedConfig
}

func mergeFields(merged, override reflect.Value) {
	for i := 0; i < override.NumField(); i++ {
		field := override.Type().Field(i)
		overrideField := override.Field(i)
		mergedField := merged.FieldByName(field.Name)

		if !mergedField.IsValid() || !mergedField.CanSet() {
			logging.Warnf("failed to merge field %s is not valid", field.Name)
			continue
		}

		if field.Type.Kind() == reflect.Struct {
			mergeFields(mergedField, overrideField)
			continue
		}

		if !overrideField.IsZero() {
			mergedField.Set(overrideField)
		}
	}
}

// ------------------------------------------- 			END Internal 		  ----------------------------------

func inferSinkValidity(res clients.SinkValidityResponse, errMapper map[error]error) (channel.ValidationResponse, error) {
	if res.ErrorCode == nil || *res.ErrorCode == 0 {
		return channel.ValidationResponse{Status: true}, nil
	}

	return channel.ValidationResponse{Status: false, Message: res.ErrorMessage}, errors.Join(errMapper[ErrValidationFailed], fmt.Errorf("failed to validate sink: %+v", res))
}

func sinkConfigToNotificationChannel(sink *types.SinkConfig) channel.NotificationChannelMetadata {
	return channel.NotificationChannelMetadata{
		Id:   sink.Id,
		Type: sink.Type,
	}
}

func errNotFound(err error) bool {
	apiErr, ok := err.(*proxyapi.Error)
	return ok && apiErr.Code == http.StatusNotFound
}

func errConflict(err error) bool {
	apiErr, ok := err.(*proxyapi.Error)
	return ok && apiErr.Code == http.StatusConflict
}

func slackChannelToSinkConfig(channel slack.ChannelConfig) *types.SinkConfig {
	return &types.SinkConfig{
		Type: NotificationChannelTypeSlack,
		Config: types.SinkConfigDetails{
			DirectNotifications: channel.DirectNotifications,
			Recipients:          channel.Recipients,
			TemplatesFolder:     channel.TemplatesFolder,
			TemplateFile:        channel.TemplateFile,
			SlackSinkConfig: types.SlackSinkConfig{
				ApiToken: channel.ApiToken,
			},
		},
	}
}

func slackChannelConfig(sink types.SinkConfigDetails) *slack.ChannelConfig {
	return &slack.ChannelConfig{
		DirectNotifications: sink.DirectNotifications,
		Recipients:          sink.Recipients,
		TemplatesFolder:     sink.TemplatesFolder,
		TemplateFile:        sink.TemplateFile,
		ApiToken:            sink.ApiToken,
	}
}

func emailConfigToSinkConfig(config email.ChannelConfig) *types.SinkConfig {
	return &types.SinkConfig{Type: NotificationChannelTypeEmail, Config: types.SinkConfigDetails{
		EmailSinkConfig: types.EmailSinkConfig{
			User:            config.Username,
			Password:        config.Password,
			AuthType:        config.AuthType,
			SmtpHost:        config.SmtpHost,
			SmtpPort:        int(config.SmtpPort),
			SmtpTlsEnabled:  config.SmtpTlsEnabled,
			From:            config.From,
			FromDisplayName: config.FromDisplayName,
			Subject:         config.Subject,
			UseAttachments:  config.UseAttachments,
		},
		Recipients:          config.Recipients,
		DirectNotifications: config.DirectNotifications,
	},
	}
}

func emailChannelConfig(sink types.SinkConfigDetails) *email.ChannelConfig {
	return &email.ChannelConfig{
		Username:            sink.User,
		Password:            sink.Password,
		SmtpHost:            sink.SmtpHost,
		SmtpPort:            int32(sink.SmtpPort),
		From:                sink.From,
		AuthType:            sink.AuthType,
		SmtpTlsEnabled:      sink.SmtpTlsEnabled,
		Recipients:          sink.Recipients,
		FromDisplayName:     sink.FromDisplayName,
		DirectNotifications: sink.DirectNotifications,
		UseAttachments:      sink.UseAttachments,
		Subject:             sink.Subject,
	}
}
