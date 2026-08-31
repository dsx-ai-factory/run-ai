/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package webhook

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sync"
	"time"
)

var logger = log.Log.WithName("migration.webhook.health")

type InternalServer interface {
	Start(ctx context.Context) error
	StartedChecker() healthz.Checker
}
type ExternalAccessCheckerServer struct {
	mu              sync.Mutex
	ExternalAddress string
	ExternalTimeout time.Duration
	InternalServer
}

func NewExternalAccessCheckerServer(internalServer InternalServer, externalAddress string, externalTimeout time.Duration) *ExternalAccessCheckerServer {
	return &ExternalAccessCheckerServer{
		ExternalAddress: externalAddress,
		ExternalTimeout: externalTimeout,
		InternalServer:  internalServer,
	}
}

// StartedChecker returns an healthz.Checker which is healthy after connection to the
// server has been started.
func (eas *ExternalAccessCheckerServer) StartedChecker() healthz.Checker {
	config := &tls.Config{
		InsecureSkipVerify: true, //nolint:gosec // config is used to connect to our own webhook port.
	}
	return func(req *http.Request) error {
		eas.mu.Lock()
		defer eas.mu.Unlock()
		logger.V(4).Info("checking internal access")
		internalChecker := eas.InternalServer.StartedChecker()
		if err := internalChecker(req); err != nil {
			return fmt.Errorf("webhook server is not reachable internally: %w", err)
		}
		logger.V(4).Info("internal access is ok")
		logger.V(4).Info("checking external access")
		d := &net.Dialer{Timeout: eas.ExternalTimeout}
		conn, err := tls.DialWithDialer(d, "tcp", eas.ExternalAddress, config)
		if err != nil {
			return fmt.Errorf("webhook server is not reachable externally: %w", err)
		}
		if err = conn.Close(); err != nil {
			return fmt.Errorf("webhook server is not reachable externally: closing connection: %w", err)
		} else {
			logger.V(4).Info("external access is ok")
			return nil
		}

	}
}
