/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodel

import (
	"fmt"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbtables"
	events "github.com/dsx-ai-factory/run-ai/runai-application-packages/async-events"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/google/uuid"
)

const (
	PVC_OBJECT_KIND                  = "PersistentVolumeClaim"
	WAIT_FOR_FIRST_CONSUMER          = "WaitForFirstConsumer"
	EVENT_TYPE_NORMAL                = "Normal"
	EVENT_TYPE_WARNING               = "Warning"
	EVENT_SOURCE_ASSETS_SERVICE      = "assets-service"
	EVENT_SOURCE_K8S_OBJECTS_TRACKER = "k8s-objects-tracker"
)

// formatEventMessage formats the event message with additional context if needed
func formatEventMessage(event *api.EventSyncFields) string {
	if event.Reason == WAIT_FOR_FIRST_CONSUMER && event.InvolvedObject.Namespace != "" {
		return fmt.Sprintf("%s (%s)", event.Message, event.InvolvedObject.Namespace)
	}
	return event.Message
}

type Event struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantID        uint      `gorm:"not null"`
	ClusterID       uuid.UUID `gorm:"not null"`
	AssetId         uuid.UUID `gorm:"not null"`
	ScopeType       string    `gorm:"not null"`
	ScopeId         string    `gorm:"not null"`
	CreatedAt       time.Time `gorm:"not null"`
	AssetKind       string    `gorm:"not null"`
	ObjectName      string
	ObjectNamespace string
	ObjectKind      string
	ObjectUid       string
	Phase           string
	Type            string `gorm:"not null"`
	Message         string `gorm:"not null"`
	Reason          string
	Source          string `gorm:"not null"`
}

func (Event) TableName() string {
	return dbtables.TableNameEvents
}

func NewEventModelFromSyncApi(tenantId int32, event *api.EventSyncFields, assetId string, ScopeId string, ScopeType string) Event {
	return Event{
		ID:              uuid.MustParse(event.Id),
		TenantID:        uint(tenantId),
		ClusterID:       uuid.MustParse(event.ClusterId),
		CreatedAt:       event.CreatedAt,
		ScopeType:       ScopeType,
		ScopeId:         ScopeId,
		AssetId:         uuid.MustParse(assetId),
		AssetKind:       getAssetKindOfObjectKind(event.InvolvedObject.Kind),
		ObjectName:      event.InvolvedObject.Name,
		ObjectNamespace: event.InvolvedObject.Namespace,
		ObjectKind:      event.InvolvedObject.Kind,
		ObjectUid:       event.InvolvedObject.Uid,
		Type:            event.Type,
		Message:         formatEventMessage(event),
		Reason:          event.Reason,
		Source:          event.Source,
	}
}

func NewEventModelFromPVCEvent(tenantId int32, assetId string, event *events.PvcEvent) Event {
	eventType := ""
	switch event.Status {
	case string(api.ASSETSYNCSTATUSENUM_ISSUES_FOUND):
		eventType = EVENT_TYPE_WARNING
	default:
		eventType = EVENT_TYPE_NORMAL
	}

	return Event{
		ID:        uuid.New(),
		TenantID:  uint(tenantId),
		ClusterID: uuid.MustParse(event.ClusterId),
		CreatedAt: time.Now(),
		ScopeType: string(event.ScopeType),
		ScopeId:   event.ScopeId,
		AssetId:   uuid.MustParse(assetId),
		Reason:    string(event.Status),
		Type:      eventType,
		Message:   fmt.Sprintf("PVC %s status changed to %s", event.ClaimName, event.Status),
		Source:    EVENT_SOURCE_K8S_OBJECTS_TRACKER,
	}
}

func getAssetKindOfObjectKind(objectKind string) string {
	switch objectKind {
	case PVC_OBJECT_KIND:
		return string(api.ASSETKIND_PVC)
	default:
		logging.Warnf("object kind %s is not supported, no matching asset kind found", objectKind)
		return ""
	}
}
