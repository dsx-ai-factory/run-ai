/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package orchestrator2

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

const tlsProbeTimeout = 10 * time.Second

// TLSProbeResult is the outcome of a read-only TLS certificate probe. Reachable
// distinguishes "could not connect" (the front door / DNS is not ready) from
// "connected but the served certificate did not verify".
type TLSProbeResult struct {
	Reachable bool
	Verified  bool
	// Detail explains a connection failure or a verification failure.
	Detail string
}

// TLSProber performs a read-only TLS handshake against an endpoint and reports
// whether the served certificate verifies for the given host against the given
// trust anchor (a CA PEM file, or the system roots when empty). It never sends
// application data.
type TLSProber interface {
	ProbeCertificate(ctx context.Context, host string, port int, caPEMPath string) (TLSProbeResult, error)
}

type realTLSProber struct{}

func NewTLSProber() TLSProber { return &realTLSProber{} }

func (p *realTLSProber) ProbeCertificate(ctx context.Context, host string, port int, caPEMPath string) (TLSProbeResult, error) {
	roots, err := tlsProbeRoots(caPEMPath)
	if err != nil {
		return TLSProbeResult{}, err
	}
	addr := net.JoinHostPort(host, strconv.Itoa(port))

	// A plain TCP dial first, so "front door not up / DNS not resolvable"
	// (Reachable=false → warning) is distinguished from "up but serving an
	// untrusted certificate" (Reachable=true, Verified=false → failure).
	conn, err := (&net.Dialer{Timeout: tlsProbeTimeout}).DialContext(ctx, "tcp", addr)
	if err != nil {
		return TLSProbeResult{Reachable: false, Detail: err.Error()}, nil
	}
	_ = conn.Close()

	tlsConn, err := (&tls.Dialer{
		NetDialer: &net.Dialer{Timeout: tlsProbeTimeout},
		Config:    &tls.Config{ServerName: host, RootCAs: roots},
	}).DialContext(ctx, "tcp", addr)
	if err != nil {
		return TLSProbeResult{Reachable: true, Verified: false, Detail: err.Error()}, nil
	}
	_ = tlsConn.Close()
	return TLSProbeResult{Reachable: true, Verified: true}, nil
}

// tlsProbeRoots returns the trust anchor for verification: a pool containing
// only the given CA PEM, or nil (the system roots) when no CA is provided.
// This mirrors how the installer's control-plane client builds its trust, so
// the probe reproduces exactly what the later health check will do.
func tlsProbeRoots(caPEMPath string) (*x509.CertPool, error) {
	if caPEMPath == "" {
		return nil, nil
	}
	pem, err := os.ReadFile(caPEMPath)
	if err != nil {
		return nil, fmt.Errorf("read CA %q: %w", caPEMPath, err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pem) {
		return nil, fmt.Errorf("no certificates found in CA %q", caPEMPath)
	}
	return pool, nil
}
