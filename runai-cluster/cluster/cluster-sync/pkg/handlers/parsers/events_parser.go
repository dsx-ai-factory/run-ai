/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	assets_openapi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
)

func ParseEventsToSyncEventsRequest(events []corev1.Event) *openapi.EventsSyncRequest {
	parsedEvents := []openapi.EventSyncFields{}
	for _, event := range events {
		// The original behavior of this parser is to not sync events that are related to the runai namespace.
		// After adding the events to assets service which we do need the "runai" namespace events, we need to sync them.
		// So we filter it from here to keep the original behavior.
		if event.InvolvedObject.Namespace == constants.NamespaceRunAi {
			continue
		}

		parsedEvents = append(parsedEvents, openapi.EventSyncFields{
			Id:        string(event.UID),
			ClusterId: config.Config().ClusterUuid,
			Type:      event.Type,
			Message:   event.Message,
			Reason:    event.Reason,
			CreatedAt: event.CreationTimestamp.Time,
			Source:    event.Source.Component,
			InvolvedObject: openapi.InvolvedObject{
				Uid:       string(event.InvolvedObject.UID),
				Kind:      event.InvolvedObject.Kind,
				Name:      event.InvolvedObject.Name,
				Namespace: event.InvolvedObject.Namespace,
			},
		})
	}

	return &openapi.EventsSyncRequest{
		Events: parsedEvents,
	}
}

func ParseEventsToAssetsSyncEventsRequest(events []corev1.Event) *assets_openapi.EventsSyncRequest {
	parsedEvents := []assets_openapi.EventSyncFields{}
	for _, event := range events {
		// We only care about events related to assets-owned resources (currently only PVCs, will add secrets later)
		if event.InvolvedObject.Kind != constants.KindPvc {
			continue
		}

		parsedEvents = append(parsedEvents, assets_openapi.EventSyncFields{
			Id:        string(event.UID),
			ClusterId: config.Config().ClusterUuid,
			Type:      event.Type,
			Message:   event.Message,
			Reason:    event.Reason,
			CreatedAt: event.CreationTimestamp.Time,
			Source:    event.Source.Component,
			InvolvedObject: assets_openapi.InvolvedObject{
				Uid:       string(event.InvolvedObject.UID),
				Kind:      event.InvolvedObject.Kind,
				Name:      event.InvolvedObject.Name,
				Namespace: event.InvolvedObject.Namespace,
			},
		})
	}

	return &assets_openapi.EventsSyncRequest{
		Events: parsedEvents,
	}
}
