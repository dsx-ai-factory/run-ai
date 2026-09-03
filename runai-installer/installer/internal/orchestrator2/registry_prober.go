/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package orchestrator2

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// RegistryCheckRequest describes a single read-only registry probe.
type RegistryCheckRequest struct {
	Host      string // registry host without scheme or path, e.g. "harbor.corp.com"
	Repo      string // repository to probe pull/push scope for; may be empty (auth-only)
	Project   string // project to confirm exists (Harbor API); may be empty
	Username  string
	Password  string
	CAPath    string // optional PEM bundle for a private registry CA
	CheckPush bool   // also probe push permission (read-only, via token scope grant)
	// Insecure probes over plain HTTP instead of HTTPS.
	Insecure bool
	// InsecureSkipVerify keeps HTTPS but does not verify the registry certificate.
	InsecureSkipVerify bool
}

// RegistryCheckResult is the outcome of a read-only registry probe. The
// *Checked flags distinguish "verified false" from "could not verify".
type RegistryCheckResult struct {
	Reachable      bool
	Authenticated  bool
	CanPull        bool
	ProjectChecked bool
	ProjectExists  bool
	PushChecked    bool
	CanPush        bool
	// Detail explains a transport/reachability failure (Reachable == false).
	Detail string
	// ResponseError is set when the registry was reached but responded in a way
	// that is neither a successful auth nor a clean credential rejection (e.g. an
	// unexpected HTTP status or an unparseable token response). It is distinct
	// from a credential failure.
	ResponseError string
}

// RegistryProber performs read-only registry checks (reachability, auth, pull,
// push-permission via token scope grant, project existence). It never writes to
// the registry.
type RegistryProber interface {
	Check(ctx context.Context, req RegistryCheckRequest) (RegistryCheckResult, error)
}

type realRegistryProber struct{}

func NewRegistryProber() RegistryProber { return &realRegistryProber{} }

func (p *realRegistryProber) Check(ctx context.Context, req RegistryCheckRequest) (RegistryCheckResult, error) {
	var res RegistryCheckResult
	client, err := registryHTTPClient(req.CAPath, req.InsecureSkipVerify)
	if err != nil {
		return res, err
	}
	base := registryScheme(req) + req.Host

	resp, err := registryGet(ctx, client, base+"/v2/", "")
	if err != nil {
		res.Detail = err.Error()
		return res, nil
	}
	resp.Body.Close()
	res.Reachable = true

	switch resp.StatusCode {
	case http.StatusOK:
		// Anonymous access allowed.
		res.Authenticated = true
		res.CanPull = true
	case http.StatusUnauthorized:
		challenge := resp.Header.Get("WWW-Authenticate")
		if strings.HasPrefix(strings.ToLower(challenge), "bearer") {
			p.tokenFlow(ctx, client, challenge, req, &res)
		} else {
			p.basicFlow(ctx, client, base, req, &res)
		}
	default:
		res.ResponseError = fmt.Sprintf("unexpected status %d from /v2/", resp.StatusCode)
	}

	if req.Project != "" && res.Authenticated {
		p.harborProject(ctx, client, base, req, &res)
	}
	return res, nil
}

// tokenFlow handles Bearer-token registries (Docker distribution, Harbor, JFrog,
// NGC): it requests a token scoped to the repo and reads back the granted
// actions to determine pull/push permission without writing anything.
func (p *realRegistryProber) tokenFlow(ctx context.Context, client *http.Client, challenge string, req RegistryCheckRequest, res *RegistryCheckResult) {
	realm, service := parseBearerChallenge(challenge)
	if realm == "" {
		res.ResponseError = "could not parse bearer challenge"
		return
	}

	scope := ""
	if req.Repo != "" {
		actions := "pull"
		if req.CheckPush {
			actions = "push,pull"
		}
		scope = "repository:" + req.Repo + ":" + actions
	}

	url := realm + "?service=" + service
	if scope != "" {
		url += "&scope=" + scope
	}
	basic := ""
	if req.Username != "" || req.Password != "" {
		basic = req.Username + ":" + req.Password
	}

	resp, err := registryGet(ctx, client, url, basic)
	if err != nil {
		res.Detail = err.Error()
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		// 401 at the token endpoint is a clean credential rejection; any other
		// status is an unexpected registry response.
		if resp.StatusCode != http.StatusUnauthorized {
			res.ResponseError = fmt.Sprintf("token request returned %d", resp.StatusCode)
		}
		return
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	var tok struct {
		Token       string `json:"token"`
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &tok); err != nil {
		res.ResponseError = "could not parse token response"
		return
	}
	raw := tok.Token
	if raw == "" {
		raw = tok.AccessToken
	}

	granted, subject, decodable, hasAccessClaim := decodeRegistryToken(raw, req.Repo)
	credsSupplied := req.Username != "" || req.Password != ""
	hasGrant := granted["pull"] || granted["push"]

	// Some token servers return an anonymous token (HTTP 200, empty subject, no
	// granted scopes) when the supplied credentials are wrong, instead of 401.
	// Detect that so it is reported as invalid credentials rather than as a valid
	// user that merely lacks permissions.
	if credsSupplied && decodable && isAnonymousSubject(subject) && !hasGrant {
		res.Detail = "invalid credentials"
		return // Authenticated stays false
	}
	res.Authenticated = true

	if req.Repo == "" || !decodable || !hasAccessClaim {
		// Auth-only probe, or an opaque token: a token was issued, so treat the
		// credentials as valid and pull as available.
		res.CanPull = true
		return
	}
	res.CanPull = granted["pull"]
	if req.CheckPush {
		res.PushChecked = true
		res.CanPush = granted["push"]
	}
}

func isAnonymousSubject(sub string) bool {
	return sub == "" || strings.EqualFold(sub, "anonymous")
}

// basicFlow handles Basic-auth registries. Push cannot be verified read-only
// here, so PushChecked stays false.
func (p *realRegistryProber) basicFlow(ctx context.Context, client *http.Client, base string, req RegistryCheckRequest, res *RegistryCheckResult) {
	basic := req.Username + ":" + req.Password
	resp, err := registryGet(ctx, client, base+"/v2/", basic)
	if err != nil {
		res.Detail = err.Error()
		return
	}
	resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		res.Authenticated = true
		res.CanPull = true
	case http.StatusUnauthorized:
		// Clean credential rejection; leave Authenticated false.
	default:
		res.ResponseError = fmt.Sprintf("basic auth returned %d", resp.StatusCode)
	}
}

// harborProject checks project existence via the Harbor API. Registries that are
// not Harbor leave ProjectChecked false (cannot verify).
func (p *realRegistryProber) harborProject(ctx context.Context, client *http.Client, base string, req RegistryCheckRequest, res *RegistryCheckResult) {
	basic := req.Username + ":" + req.Password
	resp, err := registryGet(ctx, client, base+"/api/v2.0/projects/"+req.Project, basic)
	if err != nil {
		return
	}
	resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		res.ProjectChecked, res.ProjectExists = true, true
	case http.StatusNotFound:
		res.ProjectChecked, res.ProjectExists = true, false
	}
}

// registryScheme returns the URL scheme to probe with: plain HTTP when the
// registry is marked insecure, HTTPS otherwise.
func registryScheme(req RegistryCheckRequest) string {
	if req.Insecure {
		return "http://"
	}
	return "https://"
}

func registryHTTPClient(caPath string, insecureSkipVerify bool) (*http.Client, error) {
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: insecureSkipVerify}
	if caPath != "" {
		pem, err := os.ReadFile(caPath)
		if err != nil {
			return nil, fmt.Errorf("read registry CA %q: %w", caPath, err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("no certificates parsed from registry CA %q", caPath)
		}
		tlsConfig.RootCAs = pool
	}
	return &http.Client{
		Timeout:   15 * time.Second,
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
	}, nil
}

func registryGet(ctx context.Context, client *http.Client, url, basic string) (*http.Response, error) {
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if basic != "" {
		r.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(basic)))
	}
	return client.Do(r)
}

// parseBearerChallenge extracts realm and service from a WWW-Authenticate header.
func parseBearerChallenge(h string) (realm, service string) {
	h = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(h, "Bearer"), "bearer"))
	for _, part := range strings.Split(h, ",") {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		val := strings.Trim(kv[1], `"`)
		switch strings.ToLower(kv[0]) {
		case "realm":
			realm = val
		case "service":
			service = val
		}
	}
	return realm, service
}

// decodeRegistryToken decodes a registry JWT to report its subject, Docker
// access grants, and whether it contains a Docker access claim.
func decodeRegistryToken(rawToken, repo string) (granted map[string]bool, subject string, decodable bool, hasAccessClaim bool) {
	parts := strings.Split(rawToken, ".")
	if len(parts) != 3 {
		return nil, "", false, false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, "", false, false
	}
	var claims struct {
		Subject string          `json:"sub"`
		Access  json.RawMessage `json:"access"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, "", false, false
	}
	granted = map[string]bool{}
	hasAccessClaim = claims.Access != nil
	if !hasAccessClaim {
		return granted, claims.Subject, true, false
	}
	var access []struct {
		Name    string   `json:"name"`
		Actions []string `json:"actions"`
	}
	if err := json.Unmarshal(claims.Access, &access); err != nil {
		return granted, claims.Subject, true, true
	}
	for _, a := range access {
		if a.Name != repo {
			continue
		}
		for _, act := range a.Actions {
			granted[act] = true
		}
	}
	return granted, claims.Subject, true, true
}
