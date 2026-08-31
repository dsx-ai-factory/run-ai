/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controllers

import (
	"context"
	"time"

	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/constants"
	commonConstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"go.uber.org/zap"

	backendclient "github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/handlers"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/util"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type EventController struct {
	eventsHandler *handlers.EventHandler
	backendClient *backendclient.BackendClient
	cachedReader  client.Reader

	lastSyncTime metav1.Time
}

func NewEventController(cachedReader client.Reader, backendClient *backendclient.BackendClient) *EventController {
	eventController := &EventController{
		eventsHandler: handlers.NewEventHandler(backendClient),
		backendClient: backendClient,
		cachedReader:  cachedReader,
	}

	zap.S().Debug("Event controller finished initialization")
	return eventController
}

func (controller *EventController) StartSyncingPeriodically(stopCh <-chan struct{}) {
	zap.S().Debug("The Agent Event controller has started")

	ticker := time.NewTicker(time.Second * (time.Duration(config.Config().SyncEventsIntervalSeconds)))
	defer ticker.Stop()
	//This loop synchs events every interval (from config), and enters on first iteration immediately
loop:
	for {
		controller.syncRunaiNamespacesEvents()
		select {
		case <-ticker.C:
			continue
		case <-stopCh:
			break loop
		}
	}

}

func (controller *EventController) syncRunaiNamespacesEvents() {
	namespaces, err := util.GetAllRunAiManagedNamespaces(controller.cachedReader)
	if err != nil {
		zap.S().Errorf("Failed syncing events in runai-managed namespaces, error: %v", err)
		return
	}

	namespaces = append(namespaces, corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: commonConstants.NamespaceRunAi,
		},
	})
	// consider sync time with a buffer of half the sample period - just in case
	syncTimeWithBuffer := time.Now().Add(-time.Second * (time.Duration(config.Config().SyncEventsIntervalSeconds / 2)))

	var eventsToSync []corev1.Event
	for _, namespace := range namespaces {
		events, err := controller.getEventsForNamespace(namespace.Name)
		if err != nil {
			continue
		}

		filteredEvents := controller.filterEventsByTimestamp(events.Items)
		eventsToSync = append(eventsToSync, filteredEvents...)
	}

	err = controller.eventsHandler.HandleSync(eventsToSync)
	genericControllerErrorHandler(err, constants.SyncOperation, constants.EventsResourceName, nil)

	// only update last sync time if request succeeded
	if err == nil {
		controller.lastSyncTime = metav1.Time{Time: syncTimeWithBuffer}
	}
}

func (controller *EventController) getEventsForNamespace(namespace string) (*corev1.EventList, error) {
	events := &corev1.EventList{}
	err := controller.cachedReader.List(context.Background(), events, &client.ListOptions{Namespace: namespace})
	if err != nil {
		zap.S().Errorf("Failed listing events in namespace <%s>, error: %s", namespace, err.Error())
		return nil, err
	}
	return events, nil
}

func (controller *EventController) filterEventsByTimestamp(events []corev1.Event) []corev1.Event {
	if controller.lastSyncTime.IsZero() {
		return events
	}

	var filteredEvents []corev1.Event
	for _, event := range events {
		if !event.LastTimestamp.Before(&controller.lastSyncTime) && event.InvolvedObject.UID != "" {
			filteredEvents = append(filteredEvents, event)
		}
	}

	return filteredEvents
}
