/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package advanced

import (
	"os"
	"testing"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/config"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/config/service"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/history"
	history_api "github.com/run-ai/runai/control-plane/notifications-service/internal/history/api"
	history_mocks "github.com/run-ai/runai/control-plane/notifications-service/internal/history/mocks"
	"github.com/stretchr/testify/require"
)

func TestCreateRoutesRegisters_HistoryDisabled(t *testing.T) {
	factory := newTestFactory(t, false, nil)

	registers := factory.CreateRoutesRegisters()

	require.Len(t, registers, 1)
	register, ok := registers[0].(*RoutesRegister)
	require.True(t, ok)
	require.Equal(t, "advanced_routes_register", register.GetName())
}

func TestCreateRoutesRegisters_HistoryEnabled(t *testing.T) {
	repo := history_mocks.NewEventsRepository(t)
	factory := newTestFactory(t, true, repo)

	registers := factory.CreateRoutesRegisters()

	require.Len(t, registers, 2)
	_, ok := registers[0].(*RoutesRegister)
	require.True(t, ok)
	historyRegister, ok := registers[1].(*history_api.RoutesRegister)
	require.True(t, ok)
	require.Equal(t, "history_routes", historyRegister.GetName())
}

func TestCreateCleaners_HistoryDisabled(t *testing.T) {
	factory := newTestFactory(t, false, nil)

	cleaners := factory.CreateCleaners()

	require.Empty(t, cleaners)
}

func TestCreateCleaners_HistoryEnabled(t *testing.T) {
	repo := history_mocks.NewEventsRepository(t)
	factory := newTestFactory(t, true, repo)

	cleaners := factory.CreateCleaners()

	require.Len(t, cleaners, 1)
	historyCleaner, ok := cleaners[0].(*history.EventsCleaner)
	require.True(t, ok)
	require.NotNil(t, historyCleaner)
}

func newTestFactory(t *testing.T, historyEnabled bool, eventsRepository history.EventsRepository) *Factory {
	configContent := ""
	if historyEnabled {
		configContent = `
history:
  enabled: true
`
	}
	file := config.CreateTempConfigFile(configContent)
	t.Cleanup(func() { _ = os.Remove(file.Name()) })

	return &Factory{
		serviceConfig:       service.NewConfig(file.Name()),
		secretFieldsHandler: secretFieldHandler,
		updatesChan:         make(chan string, DefaultUpdateEventsChannelSize),
		eventsRepository:    eventsRepository,
	}
}
