//go:build integration
// +build integration

/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/config/basic/route"
	notifyclient "github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/pkg/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/pkg/client/generated/config_client"
	history_client "github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/pkg/client/generated/history"
	"github.com/google/uuid"
	"github.com/sethvargo/go-retry"
	"github.com/stretchr/testify/require"
)

var historyClient *history_client.APIClient

func initHistoryClient() {
	cfg := history_client.NewConfiguration()
	cfg.Host = fmt.Sprintf("%s:%s", notificationHost, ConfigPort)
	cfg.Scheme = "http"
	historyClient = history_client.NewAPIClient(cfg)
}

func TestGetEvents(t *testing.T) {
	userID := historyTestUserID()
	setupHistoryRoute(t)
	defer deleteHistoryEvents(t, userID)

	eventID := seedHistoryEvent(t, userID, t.Name())

	got := getHistoryEvents(t, TenantId, userID, nil)
	require.Len(t, got, 1)
	require.Equal(t, eventID, got[0].Id)
	require.Equal(t, TenantId, got[0].Properties["tenant_id"])
	require.Equal(t, userID, got[0].Properties["user_id"])
	require.Equal(t, t.Name(), got[0].Properties["description"])
	require.False(t, got[0].Read)
}

func TestUpdateEventRead(t *testing.T) {
	userID := historyTestUserID()
	setupHistoryRoute(t)
	defer deleteHistoryEvents(t, userID)

	eventID := seedHistoryEvent(t, userID, t.Name())

	update := history_client.NewEventReadUpdate()
	update.SetRead(true)
	_, httpRes, err := historyClient.EventsAPI.UpdateEventRead(context.Background(), eventID).
		EventReadUpdate(*update).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpRes.StatusCode)

	readEvents := waitForHistoryEventCount(t, TenantId, userID, ptr.To(true), 1)
	require.Equal(t, eventID, readEvents[0].Id)
	require.True(t, readEvents[0].Read)

	update.SetRead(false)
	_, httpRes, err = historyClient.EventsAPI.UpdateEventRead(context.Background(), eventID).
		EventReadUpdate(*update).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpRes.StatusCode)

	unreadEvents := waitForHistoryEventCount(t, TenantId, userID, ptr.To(false), 1)
	require.Equal(t, eventID, unreadEvents[0].Id)
	require.False(t, unreadEvents[0].Read)
}

func TestUpdateEventRead_NoReadStatusChange(t *testing.T) {
	userID := historyTestUserID()
	setupHistoryRoute(t)
	defer deleteHistoryEvents(t, userID)

	eventID := seedHistoryEvent(t, userID, t.Name())

	update := history_client.NewEventReadUpdate()
	_, httpRes, err := historyClient.EventsAPI.UpdateEventRead(context.Background(), eventID).
		EventReadUpdate(*update).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusUnprocessableEntity, httpRes.StatusCode)
}

func TestDeleteEvent(t *testing.T) {
	userID := historyTestUserID()
	setupHistoryRoute(t)
	eventID := seedHistoryEvent(t, userID, t.Name())

	_, httpRes, err := historyClient.EventsAPI.DeleteEvent(context.Background(), eventID).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpRes.StatusCode)
}

func TestDeleteEvent_NotFound(t *testing.T) {
	eventID := uuid.NewString()
	_, httpRes, err := historyClient.EventsAPI.DeleteEvent(context.Background(), eventID).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, httpRes.StatusCode)
}

func TestDeleteEvents_NoEvents(t *testing.T) {
	userID := historyTestUserID()

	_, httpRes, err := historyClient.EventsAPI.DeleteEvents(context.Background()).
		TenantId(TenantId).
		UserId(userID).
		Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpRes.StatusCode)
}

func TestDeleteEvents_TenantAndUserFilters(t *testing.T) {
	userID := historyTestUserID()

	setupHistoryRoute(t)
	seedHistoryEvent(t, userID, t.Name())
	publishHistoryNotification(t, userID, t.Name())
	waitForHistoryEventCount(t, TenantId, userID, nil, 2)

	_, httpRes, err := historyClient.EventsAPI.DeleteEvents(context.Background()).
		TenantId(TenantId).
		UserId(userID).
		Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpRes.StatusCode)

	waitForHistoryEventCount(t, TenantId, userID, nil, 0)
}

func TestDeleteEvents_readFilter(t *testing.T) {
	userID := historyTestUserID()
	setupHistoryRoute(t)

	eventID := seedHistoryEvent(t, userID, t.Name())

	update := history_client.NewEventReadUpdate()
	update.SetRead(true)
	_, httpRes, err := historyClient.EventsAPI.UpdateEventRead(context.Background(), eventID).
		EventReadUpdate(*update).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpRes.StatusCode)

	publishHistoryNotification(t, userID, t.Name())
	waitForHistoryEventCount(t, TenantId, userID, nil, 2)
	waitForHistoryEventCount(t, TenantId, userID, ptr.To(false), 1)

	_, httpRes, err = historyClient.EventsAPI.DeleteEvents(context.Background()).
		TenantId(TenantId).
		UserId(userID).
		Read(true).
		Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpRes.StatusCode)

	waitForHistoryEventCount(t, TenantId, userID, nil, 1)
	waitForHistoryEventCount(t, TenantId, userID, ptr.To(false), 1)
	waitForHistoryEventCount(t, TenantId, userID, ptr.To(true), 0)
	deleteHistoryEvents(t, userID)
}

func TestCleanup_MaxAge(t *testing.T) {
	userID := historyTestUserID()
	setupHistoryRoute(t)
	defer deleteHistoryEvents(t, userID)

	seedHistoryEvent(t, userID, t.Name()+"-stale")

	waitForHistoryEventCountWithTimeout(t, TenantId, userID, nil, 0, 5*time.Minute)
}

func TestCleanup_MaxEventsPerUser(t *testing.T) {
	userID := historyTestUserID()
	setupHistoryRoute(t)

	for i := range 10 {
		publishHistoryNotification(t, userID, fmt.Sprintf("event-%s-%d", t.Name(), i))
	}
	waitForHistoryEventCount(t, TenantId, userID, nil, 10)

	waitForHistoryEventCountWithTimeout(t, TenantId, userID, nil, 5, time.Minute)
	deleteHistoryEvents(t, userID)
}

func historyTestUserID() string {
	return "history-user-" + uuid.NewString()
}

func setupHistoryRoute(t *testing.T) {
	routeData := config_client.RouteData{
		RouteId:  t.Name(),
		TenantId: TenantId,
		RouteConfig: StructToMap(route.BasicRouteConfiguration{
			ListenerId: ListenerId,
			NotifyIds:  []string{"history"},
		}),
	}
	_, httpRes, err := configClient.TenantRouteAPI.CreateRoute(context.Background(), TenantId).
		RouteData(routeData).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpRes.StatusCode)

	t.Cleanup(func() {
		configClient.TenantRouteAPI.DeleteRoute(context.Background(), TenantId, t.Name()).Execute()
	})
}

func seedHistoryEvent(t *testing.T, userID, description string) string {
	publishHistoryNotification(t, userID, description)
	events := waitForHistoryEventCount(t, TenantId, userID, nil, 1)
	return events[0].Id
}

func deleteHistoryEvents(t *testing.T, userID string) {
	_, _, err := historyClient.EventsAPI.DeleteEvents(context.Background()).
		TenantId(TenantId).
		UserId(userID).
		Execute()
	require.NoError(t, err)
}

func publishHistoryNotification(t *testing.T, userID, description string) {
	notifyclient.BuildNotification("name", "type").
		WithTenant(TenantId).
		WithUser(userID).
		WithDescription(description).
		WithProperty("test_name", t.Name()).
		Notify()
}

func getHistoryEvents(t *testing.T, tenantID, userID string, read *bool) []history_client.Event {
	req := historyClient.EventsAPI.GetEvents(context.Background()).
		TenantId(tenantID).
		UserId(userID)
	if read != nil {
		req = req.Read(*read)
	}
	events, httpRes, err := req.Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpRes.StatusCode)
	return events
}

func waitForHistoryEventCount(t *testing.T, tenantID, userID string, read *bool, wantCount int) []history_client.Event {
	return waitForHistoryEventCountWithTimeout(t, tenantID, userID, read, wantCount, 35*time.Second)
}

func waitForHistoryEventCountWithTimeout(
	t *testing.T,
	tenantID, userID string,
	read *bool,
	wantCount int,
	timeout time.Duration,
) []history_client.Event {
	var events []history_client.Event
	err := retry.Do(context.Background(), retry.WithMaxDuration(timeout, retry.NewConstant(time.Second)), func(ctx context.Context) error {
		req := historyClient.EventsAPI.GetEvents(ctx).
			TenantId(tenantID).
			UserId(userID)
		if read != nil {
			req = req.Read(*read)
		}
		var httpRes *http.Response
		var err error
		events, httpRes, err = req.Execute()
		if err != nil {
			return retry.RetryableError(err)
		}
		if httpRes.StatusCode != http.StatusOK {
			return retry.RetryableError(fmt.Errorf("unexpected status %d", httpRes.StatusCode))
		}
		if len(events) != wantCount {
			return retry.RetryableError(fmt.Errorf("expected %d events, got %d", wantCount, len(events)))
		}
		return nil
	})
	require.NoError(t, err)
	return events
}

func openHistoryEventsDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := fmt.Sprintf(
		"host=%s port=5432 user=%s password=%s dbname=runai sslmode=disable",
		GetEnv("POSTGRES_HOST", "postgres"),
		GetEnv("POSTGRES_USER", "runai"),
		GetEnv("POSTGRES_PASSWORD", "admin"),
	)
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	return db
}
