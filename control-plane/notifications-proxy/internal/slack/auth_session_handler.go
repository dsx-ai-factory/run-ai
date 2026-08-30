package slack

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/run-ai/runai/runai-common-packages/cache/v2"
	"github.com/run-ai/runai/runai-common-packages/cache/v2/options"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

const slackAuthExpirationSeconds = 120

type SlackAuthSessionHandler struct {
}

func NewSlackAuthSessionHandler() *SlackAuthSessionHandler {
	err := cache.RegisterStore[SlackAuthSessionState](
		options.Ttl(5*time.Minute),
		options.MaxStoreSizeMB(1),
		options.MaxRecordSizeKB(5),
	)
	if err != nil {
		logging.Warnf("Failed to register slack state store: %v", err)
	}
	return &SlackAuthSessionHandler{}
}

func (s *SlackAuthSessionHandler) Create(ctx context.Context, tenantId string, email string) (string, error) {
	state := SlackAuthSessionState{
		TenantId:  tenantId,
		Email:     email,
		Timestamp: time.Now().Unix(),
	}

	sessionId := uuid.New().String()
	if err := cache.Set[SlackAuthSessionState](ctx, sessionId, state); err != nil {
		return "", fmt.Errorf("failed to store session state: %w", err)
	}
	return sessionId, nil
}

func (s *SlackAuthSessionHandler) Parse(ctx context.Context, sessionId string) (*SlackAuthSessionState, error) {
	state, err := cache.Get[SlackAuthSessionState](ctx, sessionId)
	if err != nil {
		return nil, fmt.Errorf("failed to get session state: %w", err)
	}
	if err := s.validate(state); err != nil {
		return nil, err
	}

	return state, nil
}

func (s *SlackAuthSessionHandler) validate(state *SlackAuthSessionState) error {
	if state == nil {
		return fmt.Errorf("%w: session state is nil", ErrInvalidState)
	}

	if state.TenantId == "" {
		return fmt.Errorf("%w: tenantId is empty", ErrInvalidState)
	}

	if state.Email == "" {
		return fmt.Errorf("%w: email is empty", ErrInvalidState)
	}

	if time.Now().Unix()-state.Timestamp > slackAuthExpirationSeconds {
		return ErrExpiredState
	}

	return nil
}
