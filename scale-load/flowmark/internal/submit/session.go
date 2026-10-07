/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/client"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/core"
)

const cleanupGrace = time.Minute

type Session struct {
	Clients   *client.Clients
	ClusterID string
	RunID     string
	Project   ProjectRef
	timeout   time.Duration
}

func OpenSession(ctx context.Context, target core.Target, limits client.Limits, cluster, projectPrefix string, setupTimeout time.Duration) (*Session, error) {
	setupCtx, cancel := context.WithTimeout(ctx, setupTimeout)
	defer cancel()
	clients, err := client.NewWithLimits(ctx, target, limits)
	if err != nil {
		return nil, err
	}
	if !clients.TokenExpiry.IsZero() {
		log.Printf("%s: control plane token expires in %s; longer runs will fail with 401", projectPrefix, time.Until(clients.TokenExpiry).Round(time.Minute))
	}
	selected, err := clients.Cluster(setupCtx, cluster)
	if err != nil {
		return nil, err
	}
	runID, err := core.NewRunID()
	if err != nil {
		return nil, err
	}
	session := &Session{Clients: clients, ClusterID: selected.Uuid, RunID: runID, timeout: setupTimeout}
	session.Project, err = Project(setupCtx, clients, session.ClusterID, projectPrefix+"-"+runID, setupTimeout)
	if err != nil {
		return nil, errors.Join(err, session.Close())
	}
	return session, nil
}

func (s *Session) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout+cleanupGrace)
	defer cancel()
	name := s.Project.Name
	if name == "" {
		return nil
	}
	return CleanupProject(ctx, s.Clients, s.ClusterID, name, s.timeout)
}
