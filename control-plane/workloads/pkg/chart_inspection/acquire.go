/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package chart_inspection

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	httpsScheme      = "https"
	maxDownloadBytes = 32 << 20
	maxIndexBytes    = 1 << 20
	downloadTimeout  = 30 * time.Second
)

// Error contains a safe public message; outbound URLs, response bodies and credentials never enter it.
type Error struct {
	StatusCode int
	Code       string
	Message    string
}

func (e *Error) Error() string { return e.Message }

// Selection identifies exactly one chart source. Uploaded bytes must come from an authorized callback.
type Selection struct {
	ChartID      string
	Repo         string
	ChartName    string
	ChartVersion string
	ChartURL     string
}

// FetchFunc acquires archive bytes with the supplied context and resolved credential.
// Custom upload fetchers must enforce access to stored bytes and honor cancellation.
type FetchFunc func(context.Context, Selection, string) ([]byte, error)

func (s Selection) validate() error {
	sources := 0
	if s.ChartID != "" {
		sources++
	}
	if s.ChartURL != "" {
		sources++
	}
	hasRepo := s.Repo != "" || s.ChartName != "" || s.ChartVersion != ""
	if hasRepo {
		sources++
	}
	if sources != 1 || (hasRepo && (s.Repo == "" || s.ChartName == "" || s.ChartVersion == "")) {
		return acquisitionError(http.StatusBadRequest, "invalid_chart_source", "Provide exactly one chart source: chartId, chartURL, or repo with chartName and chartVersion.")
	}
	if len(s.ChartID) > 128 || len(s.ChartName) > 512 || len(s.ChartVersion) > 512 {
		return acquisitionError(http.StatusBadRequest, "invalid_chart_source", "Chart source identifiers are too long.")
	}
	return nil
}

func acquisitionError(status int, code, message string) *Error {
	return &Error{StatusCode: status, Code: code, Message: message}
}

type hostResolver interface {
	LookupIPAddr(context.Context, string) ([]net.IPAddr, error)
}
type dialFunc func(context.Context, string, string) (net.Conn, error)
type remoteFetcher struct {
	resolver  hostResolver
	dial      dialFunc
	tlsConfig *tls.Config
}

func newRemoteFetcher(resolver hostResolver, dial dialFunc) *remoteFetcher {
	return &remoteFetcher{resolver: resolver, dial: dial}
}

// FetchRemote downloads an HTTP(S) archive or exact repository version without
// installing it. Private/internal addresses are rejected; saved credentials are
// accepted only for HTTPS NGC sources and are not forwarded across origins.
func FetchRemote(ctx context.Context, selection Selection, apiKey string) ([]byte, error) {
	return newRemoteFetcher(net.DefaultResolver, (&net.Dialer{Timeout: downloadTimeout}).DialContext).fetch(ctx, selection, apiKey)
}

func (f *remoteFetcher) fetch(ctx context.Context, selection Selection, apiKey string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := selection.validate(); err != nil {
		return nil, err
	}
	if selection.ChartID != "" {
		return nil, acquisitionError(http.StatusBadRequest, "invalid_chart_source", "Stored chart inspection requires an authorized uploaded-chart callback.")
	}
	source := selection.ChartURL
	if selection.Repo != "" {
		source = selection.Repo
	}
	u, err := validateSourceURL(source)
	if err != nil {
		return nil, err
	}
	if apiKey != "" && (u.Scheme != httpsScheme || !ngcHost(u.Hostname())) {
		return nil, acquisitionError(http.StatusBadRequest, "unsupported_chart_auth", "Saved NGC credentials require an HTTPS NGC chart source.")
	}
	ctx, cancel := context.WithTimeout(ctx, downloadTimeout)
	defer cancel()
	transport := &http.Transport{
		DialContext: f.safeDial, TLSClientConfig: f.tlsConfig, TLSHandshakeTimeout: 10 * time.Second,
		ResponseHeaderTimeout: downloadTimeout, MaxResponseHeaderBytes: 64 << 10, DisableKeepAlives: true,
	}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: downloadTimeout}
	// Repository entries and redirects may point elsewhere. Keep credentials
	// tied to the original origin even when the target address is public.
	origin := *u
	client.CheckRedirect = func(next *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return acquisitionError(422, "chart_redirect_limit", "Chart source returned too many redirects.")
		}
		if _, err := validateSourceURL(next.URL.String()); err != nil {
			return err
		}
		if !sameOrigin(&origin, next.URL) {
			next.Header.Del("Authorization")
		}
		return nil
	}
	if selection.Repo == "" {
		return f.download(ctx, client, u, apiKey, maxDownloadBytes)
	}
	indexURL := *u
	indexURL.RawPath = strings.TrimRight(indexURL.EscapedPath(), "/") + "/index.yaml"
	indexURL.Path = strings.TrimRight(indexURL.Path, "/") + "/index.yaml"
	data, err := f.download(ctx, client, &indexURL, apiKey, maxIndexBytes)
	if err != nil {
		return nil, err
	}
	boundedIndex, err := yamlObject(data)
	if err != nil {
		return nil, acquisitionError(422, "invalid_chart_index", "Repository index is not valid YAML.")
	}
	jsonIndex, err := json.Marshal(boundedIndex)
	if err != nil {
		return nil, acquisitionError(422, "invalid_chart_index", "Repository index could not be decoded.")
	}
	var index struct {
		Entries map[string][]struct {
			Name    string   `json:"name"`
			Version string   `json:"version"`
			URLs    []string `json:"urls"`
		} `json:"entries"`
	}
	if err := json.Unmarshal(jsonIndex, &index); err != nil || index.Entries == nil {
		return nil, acquisitionError(422, "invalid_chart_index", "Repository index must contain chart entries.")
	}
	for _, entry := range index.Entries[selection.ChartName] {
		if entry.Version != selection.ChartVersion || (entry.Name != "" && entry.Name != selection.ChartName) || len(entry.URLs) == 0 {
			continue
		}
		ref, err := url.Parse(entry.URLs[0])
		if err != nil {
			return nil, acquisitionError(422, "invalid_chart_index", "Repository index contains an invalid chart URL.")
		}
		target := indexURL.ResolveReference(ref)
		if _, err := validateSourceURL(target.String()); err != nil {
			return nil, err
		}
		key := apiKey
		if !sameOrigin(&origin, target) {
			key = ""
		}
		return f.download(ctx, client, target, key, maxDownloadBytes)
	}
	return nil, acquisitionError(422, "chart_version_not_found", "The selected chart name and version are absent from the repository index.")
}

func validateSourceURL(raw string) (*url.URL, error) {
	if len(raw) > 8192 {
		return nil, acquisitionError(400, "invalid_chart_url", "Chart source URL is too long.")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil, acquisitionError(400, "invalid_chart_url", "Chart source URL is malformed.")
	}
	if u.Scheme == "oci" {
		return nil, acquisitionError(400, "unsupported_chart_transport", "OCI chart inspection is not supported.")
	}
	if (u.Scheme != "http" && u.Scheme != httpsScheme) || u.Hostname() == "" || u.User != nil || u.Fragment != "" {
		return nil, acquisitionError(400, "invalid_chart_url", "Use an HTTP or HTTPS chart URL without embedded credentials or a fragment.")
	}
	if ip := net.ParseIP(u.Hostname()); ip != nil && blockedIP(ip) {
		return nil, acquisitionError(400, "blocked_chart_address", "Chart source must not use a private or internal address.")
	}
	return u, nil
}

func ngcHost(host string) bool {
	host = strings.ToLower(strings.TrimSuffix(host, "."))
	return host == "ngc.nvidia.com" || strings.HasSuffix(host, ".ngc.nvidia.com")
}
func sameOrigin(a, b *url.URL) bool {
	return strings.EqualFold(a.Scheme, b.Scheme) && strings.EqualFold(a.Host, b.Host)
}

func blockedIP(ip net.IP) bool {
	if !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
		return true
	}
	if v4 := ip.To4(); v4 != nil {
		return v4[0] == 0 || v4[0] >= 240 || (v4[0] == 100 && v4[1] >= 64 && v4[1] <= 127) || (v4[0] == 198 && (v4[1] == 18 || v4[1] == 19))
	}
	return false
}

// Resolve and validate immediately before connecting, then dial the validated IP itself.
// A separate validation followed by an ordinary hostname dial would permit DNS rebinding.
func (f *remoteFetcher) safeDial(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, acquisitionError(400, "invalid_chart_url", "Chart source has an invalid host or port.")
	}
	ips, err := f.resolver.LookupIPAddr(ctx, host)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, acquisitionError(502, "chart_download_failed", "Could not resolve the chart source host.")
	}
	if len(ips) == 0 {
		return nil, acquisitionError(502, "chart_download_failed", "Chart source host has no usable address.")
	}
	for _, ip := range ips {
		if ip.IP == nil || blockedIP(ip.IP) || ip.Zone != "" {
			return nil, acquisitionError(400, "blocked_chart_address", "Chart source resolves to a private or internal address.")
		}
	}
	for _, ip := range ips {
		connection, err := f.dial(ctx, network, net.JoinHostPort(ip.IP.String(), port))
		if err == nil {
			return connection, nil
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
	}
	return nil, acquisitionError(502, "chart_download_failed", "Could not connect to the chart source.")
}

func (f *remoteFetcher) download(ctx context.Context, client *http.Client, u *url.URL, apiKey string, limit int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, acquisitionError(400, "invalid_chart_url", "Chart source URL is malformed.")
	}
	if apiKey != "" {
		req.SetBasicAuth("$oauthtoken", apiKey)
	}
	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		var public *Error
		if errors.As(err, &public) {
			return nil, public
		}
		return nil, acquisitionError(502, "chart_download_failed", "Could not download the chart source (network, TLS, or redirect failure).")
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, acquisitionError(http.StatusBadGateway, "chart_download_failed", fmt.Sprintf("Chart source returned HTTP %d.", resp.StatusCode))
	}
	if resp.ContentLength > limit {
		return nil, acquisitionError(422, "chart_source_too_large", "Chart source exceeds the inspection size limit.")
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, acquisitionError(502, "chart_download_failed", "Could not read the chart source.")
	}
	if int64(len(data)) > limit {
		return nil, acquisitionError(422, "chart_source_too_large", "Chart source exceeds the inspection size limit.")
	}
	return data, nil
}
