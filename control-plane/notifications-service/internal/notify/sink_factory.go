/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package notify

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/sinks"
	"github.com/rs/zerolog/log"
)

type SinkFactory struct{}

func NewSinkFactory() Factory {
	return &SinkFactory{}
}

func (sf SinkFactory) CreateSinks(configProvider ConfigProvider) ([]sinks.Sinker, error) {
	sinkers := []sinks.Sinker{}
	sinks, err := configProvider.GetSinks()
	if err != nil {
		return sinkers, err
	}

	for sinkId, conf := range sinks {
		if s, err := newSink(sinkId, conf); err != nil {
			log.Err(err).Msg("failed creating a sink")
		} else {
			sinkers = append(sinkers, s)
		}
	}
	return sinkers, nil
}

func newSink(sinkId string, sc *sinks.SinkConfig) (sinks.Sinker, error) {
	switch {
	case sc.Email != nil:
		return sinks.NewEmailSink(sc.Email), nil
	case sc.Slack != nil:
		return sinks.NewSlackSink(sc.Slack), nil
	case sc.Postgres != nil:
		return sinks.NewPostgresSink(sinkId, sc.Postgres), nil
	case sc.History != nil:
		return sinks.NewHistorySink(sinkId, sc.History)
	}

	return nil, fmt.Errorf("no sink config found")
}
