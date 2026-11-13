/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clients

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// On-prem NGC-compatible registries use RFC1918 and IPv6 unique-local
// addresses. Loopback, link-local, and unspecified addresses are not
// registries. The dial uses the addresses checked here so a second lookup
// cannot move the connection onto a blocked address.

func newRegistryHTTPClient(c *ngcClient, timeout time.Duration) *http.Client {
	base, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		base = &http.Transport{}
	}
	transport := base.Clone()
	// A proxy would connect to the checked host itself, so the dial pin
	// would only see the proxy address.
	transport.Proxy = nil
	transport.DialContext = c.dialRegistry
	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			if err := c.validateRegistryBaseURL(req.Context(), req.URL.String()); err != nil {
				return fmt.Errorf("redirect blocked: %w", err)
			}
			return nil
		},
	}
}

func (c *ngcClient) validateRegistryBaseURL(ctx context.Context, rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid registry URL: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("registry URL scheme %q is not allowed", parsed.Scheme)
	}
	host := parsed.Hostname()
	if host == "" || blockedRegistryHost(host) {
		return fmt.Errorf("registry URL host is not allowed")
	}
	ips, err := c.lookupRegistryIPs(ctx, host)
	if err != nil {
		return fmt.Errorf("resolve registry host: %w", err)
	}
	if len(ips) == 0 {
		return fmt.Errorf("registry host resolved to no addresses")
	}
	for _, ip := range ips {
		if blockedRegistryIP(ip) {
			return fmt.Errorf("registry host resolves to a blocked address")
		}
	}
	return nil
}

func (c *ngcClient) lookupRegistryIPs(ctx context.Context, host string) ([]net.IP, error) {
	if ip := net.ParseIP(host); ip != nil {
		return []net.IP{ip}, nil
	}
	if c.lookupHost != nil {
		return c.lookupHost(ctx, host)
	}
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	ips := make([]net.IP, 0, len(addrs))
	for _, addr := range addrs {
		ips = append(ips, addr.IP)
	}
	return ips, nil
}

func (c *ngcClient) dialRegistry(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("parse registry address: %w", err)
	}
	if blockedRegistryHost(host) {
		return nil, fmt.Errorf("registry host is not allowed")
	}
	ips, err := c.lookupRegistryIPs(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("resolve registry host: %w", err)
	}
	for _, ip := range ips {
		if blockedRegistryIP(ip) {
			return nil, fmt.Errorf("registry host resolves to a blocked address")
		}
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("registry host resolved to no addresses")
	}

	// One lookup, then each address that passed it. The request context is
	// the deadline for every attempt, so a later address is not dialed after
	// the caller has given up.
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	var lastErr error
	for _, ip := range ips {
		if err = ctx.Err(); err != nil {
			return nil, fmt.Errorf("dial registry: %w", err)
		}
		target := net.JoinHostPort(ip.String(), port)
		var conn net.Conn
		if c.dialContext != nil {
			conn, err = c.dialContext(ctx, network, target)
		} else {
			conn, err = dialer.DialContext(ctx, network, target)
		}
		if err == nil {
			return conn, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("dial registry: %w", lastErr)
}

func blockedRegistryHost(host string) bool {
	name := strings.TrimSuffix(strings.ToLower(host), ".")
	switch name {
	case "localhost", "metadata.google.internal", "metadata.goog":
		return true
	}
	return strings.HasSuffix(name, ".localhost") ||
		strings.HasSuffix(name, ".metadata.google.internal") ||
		strings.HasSuffix(name, ".metadata.goog")
}

func blockedRegistryIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
		return true
	}
	if v4 := ip.To4(); v4 != nil {
		if v4[0] == 0 {
			return true
		}
		if v4.Equal(net.IPv4(100, 100, 100, 200)) {
			return true
		}
		return false
	}
	return ip.Equal(net.ParseIP("fd00:ec2::254"))
}
