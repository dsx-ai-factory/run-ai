package clients

import (
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"
	proxyapi "github.com/run-ai/runai/control-plane/notifications-proxy/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/notifications-service/pkg/client/generated/config_client"
	"github.com/run-ai/runai/control-plane/notifications-service/pkg/client/types"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

type Sinks struct {
	sinksClient *config_client.SinksAPIService
}

func NewSinks() *Sinks {
	return &Sinks{sinksClient: GetConfigClient().SinksAPI}
}

func (sc *Sinks) GetSink(ctx context.Context, channelId string) (*types.SinkConfig, error) {
	config, httpResponse, err := sc.sinksClient.GetSink(ctx, channelId).Execute()
	if err != nil {
		return nil, proxyapi.NewErrorFromResponse(httpResponse, err)
	}

	innerConfig := types.SinkConfigDetails{}
	if err := DecodeFromMap(config.Config, &innerConfig); err != nil {
		logging.Warnf("failed to decode sink config for sink %s: %v", config.Id, err)
		return nil, fmt.Errorf("failed to decode sink config: %v", err)
	}
	return &types.SinkConfig{
		Id:     config.Id,
		Type:   config.Type,
		Config: innerConfig,
	}, nil
}

func (sc *Sinks) GetSinks(ctx context.Context) ([]*types.SinkConfig, error) {
	request := sc.sinksClient.GetSinks(ctx)
	configs, _, err := request.Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get sinks: %v", err)
	}

	sinks := make([]*types.SinkConfig, 0)
	for _, c := range configs {
		innerConfig := types.SinkConfigDetails{}
		if err := DecodeFromMap(c.Config, &innerConfig); err != nil {
			logging.Warnf("failed to decode sink config for sink %s: %v", c.Id, err)
			continue
		}

		sinks = append(sinks, &types.SinkConfig{
			Id:     c.Id,
			Type:   c.Type,
			Config: innerConfig,
		})
	}
	return sinks, nil
}

func (sc *Sinks) DeleteSink(ctx context.Context, channelId string) error {
	_, httpResponse, err := sc.sinksClient.DeleteSink(ctx, channelId).Execute()
	if err != nil {
		return proxyapi.NewErrorFromResponse(httpResponse, err)
	}
	return nil
}

func (sc *Sinks) CreateSink(ctx context.Context, channelId string, config *types.SinkConfig) error {
	cfg, err := DecodeToMap(config.Config)
	if err != nil {
		return fmt.Errorf("failed to decode sink config: %v", err)
	}

	_, httpResponse, err := sc.sinksClient.CreateSink(ctx).SinkData(config_client.SinkData{
		Id:     channelId,
		Type:   string(config.Type),
		Config: cfg,
	}).Execute()
	if err != nil {
		return proxyapi.NewErrorFromResponse(httpResponse, err)
	}
	return nil
}

func (sc *Sinks) UpdateSink(ctx context.Context, channelId string, config *types.SinkConfig) error {
	cfg, err := DecodeToMap(config.Config)
	if err != nil {
		return fmt.Errorf("failed to decode sink config: %v", err)
	}

	_, httpResponse, err := sc.sinksClient.UpdateSink(ctx, channelId).SinkDataForUpdate(config_client.SinkDataForUpdate{
		Type:   string(config.Type),
		Config: cfg,
	}).Execute()
	if err != nil {
		return proxyapi.NewErrorFromResponse(httpResponse, err)
	}
	return nil
}

func (sc *Sinks) ValidateSink(ctx context.Context, channelId string, config *types.SinkConfig) (*SinkValidityResponse, error) {
	cfg, err := DecodeToMap(config.Config)
	if err != nil {
		return nil, err
	}

	response, _, err := sc.sinksClient.ValidateSink(ctx).SinkData(config_client.SinkData{
		Id:     channelId,
		Type:   string(config.Type),
		Config: cfg,
	}).Execute()

	logging.Infof("validate sink response: %+v, %+v", response, err)
	if err != nil || response == nil {
		return nil, err
	}

	msg := ""
	if response.ErrorMessage != nil {
		msg = *response.ErrorMessage
	}

	return &SinkValidityResponse{
		ErrorCode:    response.ErrorCode,
		ErrorMessage: msg,
	}, err
}

func DecodeToMap(config any) (map[string]interface{}, error) {
	cfg := map[string]interface{}{}
	err := mapstructure.Decode(config, &cfg)
	if err != nil {
		return cfg, err
	}
	return cfg, nil
}

func DecodeFromMap(cfg map[string]interface{}, config any) error {
	err := mapstructure.Decode(cfg, config)
	if err != nil {
		return err
	}
	return nil
}

type SinkValidityResponse struct {
	ErrorCode    *int32 `json:"errorCode,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
}
