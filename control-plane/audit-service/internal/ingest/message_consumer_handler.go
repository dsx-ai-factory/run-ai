/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package ingest

import (
	"context"
	"strconv"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/audit-service/internal/database/models"
	"github.com/run-ai/runai/control-plane/audit-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/runai-application-packages/dbmodels"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

type MessageConsumer interface {
	Start(ctx context.Context, handleLog func(log *api.AuditLog) error) error
	Stop() error
	GetName() string
}

type MessageConsumerHandler struct {
	next            Handler
	messageConsumer MessageConsumer
}

func NewMessageConsumerHandler(next Handler, consumer MessageConsumer) (Handler, error) {
	logging.Infof("starting consumer on audit service: %s", consumer.GetName())
	mc := MessageConsumerHandler{next: next, messageConsumer: consumer}

	err := mc.InitConsumer()
	if err != nil {
		logging.Errorf("failed to start consumer on audit service: %s, err: %v", consumer.GetName(), err)
		return nil, err
	}

	return &mc, nil
}

func (mc *MessageConsumerHandler) InitConsumer() error {
	return mc.messageConsumer.Start(context.Background(), mc.handleLog)
}

func (mc *MessageConsumerHandler) Stop() error {
	err := mc.messageConsumer.Stop()
	if err != nil {
		return err
	}
	return mc.next.Stop()
}

func (mc *MessageConsumerHandler) Handle(log *models.AuditLog) error {
	return mc.next.Handle(log)
}

func (mc *MessageConsumerHandler) Next() Handler {
	return mc.next
}

func (mc *MessageConsumerHandler) handleLog(log *api.AuditLog) error {
	parsedLog, err := ParseToModelLog(log)
	if err != nil {
		return err
	}
	err = mc.Handle(parsedLog)
	if err != nil {
		logging.Errorf("failed to handle audit log: %v", err)
	}
	return err
}

func ParseToModelLog(log *api.AuditLog) (*models.AuditLog, error) {
	uid, err := uuid.Parse(log.RequestId)
	if err != nil {
		log.RequestId = uuid.NewString()
	}

	tenant, err := strconv.Atoi(log.TenantId)
	if err != nil {
		logging.Errorf("audit is either missing mandatory field TenantId for URL %q, or failed to parse it: %v", log.Url, err)
		return nil, err
	}

	// not a mandatory field
	clusterId, err := parseClusterId(log)
	if err != nil {
		logging.Errorf("failed to parse ClusterId for URL %q: %v", log.Url, err)
		return nil, err
	}

	return &models.AuditLog{
		LogId:          uuid.New(),
		TenantId:       tenant,
		ClusterId:      clusterId,
		ClusterName:    log.ClusterName,
		RequestId:      uid,
		Timestamp:      log.Timestamp,
		Subject:        log.Subject,
		SubjectType:    log.SubjectType,
		SourceIp:       log.SourceIp,
		Action:         log.Action,
		HttpMethod:     log.HttpMethod,
		Url:            log.Url,
		EntityType:     log.EntityType,
		EntityName:     log.EntityName,
		EntityId:       log.EntityId,
		HttpStatusCode: int(log.HttpStatusCode),
		Result:         parseResult(log),
		Origin:         "",
		Metadata:       convertToJsonBMap(log.Metadata),
	}, nil
}

func convertToJsonBMap(logMetadata map[string]string) dbmodels.JsonBMap {
	jsonb := dbmodels.JsonBMap{}
	for key, value := range logMetadata {
		jsonb[key] = value
	}
	return jsonb
}

func parseClusterId(log *api.AuditLog) (uuid.UUID, error) {
	if log.ClusterId == "" {
		return uuid.UUID{}, nil
	}
	return uuid.Parse(log.ClusterId)
}

func parseResult(log *api.AuditLog) string {
	if log.Result == "" {
		return parseStatusCodeToResult(log.HttpStatusCode)
	}

	return log.Result
}

func parseStatusCodeToResult(code int32) string {
	if code >= 200 && code < 300 {
		return "Succeeded"
	}

	return "Failed"
}
