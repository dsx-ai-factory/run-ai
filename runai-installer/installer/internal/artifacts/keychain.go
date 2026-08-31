/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package artifacts downloads the images and charts an artifact lock pins and
// assembles them into an air-gapped bundle.
package artifacts

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
)

// Credential authenticates to one registry. Prefix is the match key: a host, or
// a host and path prefix scoping the credentials to one project.
type Credential struct {
	Prefix             string
	Username           string
	Password           string
	BearerToken        string
	CAChainPath        string
	Insecure           bool
	InsecureSkipVerify bool
}

// Credentials is the configured credential set.
type Credentials []Credential

// Match returns the credential whose prefix is the longest one covering the target.
func (c Credentials) Match(target string) *Credential {
	target = NormalizeTarget(target)

	var best *Credential
	var bestLen int
	for i := range c {
		prefix := NormalizeTarget(c[i].Prefix)
		if prefix == "" || !covers(prefix, target) {
			continue
		}
		if best == nil || len(prefix) > bestLen {
			best, bestLen = &c[i], len(prefix)
		}
	}
	return best
}

func covers(prefix, target string) bool {
	if !strings.HasPrefix(target, prefix) {
		return false
	}
	rest := target[len(prefix):]
	if rest == "" || strings.HasPrefix(rest, "/") {
		return true
	}
	if !strings.Contains(prefix, "/") {
		return false
	}
	return strings.HasPrefix(rest, ":") || strings.HasPrefix(rest, "@")
}

// NormalizeTarget canonicalizes the registry host of a target.
func NormalizeTarget(target string) string {
	host, path, hasPath := strings.Cut(target, "/")
	if host != defaultRegistryAlias {
		return target
	}
	if !hasPath {
		return name.DefaultRegistry
	}
	return name.DefaultRegistry + "/" + path
}

const defaultRegistryAlias = "docker.io"

// ParseImageTarget returns an image reference's registry host and host-qualified repository.
func ParseImageTarget(ref string) (host, repository string, err error) {
	parsed, err := name.ParseReference(ref)
	if err != nil {
		return "", "", fmt.Errorf("invalid image reference %q: %w", ref, err)
	}
	repo := parsed.Context()
	return repo.RegistryStr(), repo.RegistryStr() + "/" + repo.RepositoryStr(), nil
}

// Keychain resolves the credentials and transport for a target. It is not an
// authn.Keychain: that resolves by host alone and so cannot honour credentials
// scoped to one project of a shared registry.
type Keychain struct {
	creds Credentials

	mu         sync.Mutex
	transports map[string]http.RoundTripper
}

func NewKeychain(creds Credentials) *Keychain {
	return &Keychain{creds: creds, transports: map[string]http.RoundTripper{}}
}

// AuthFor resolves the credentials for a reference: the longest matching
// configured prefix, then the ambient docker configuration, then anonymous.
func (k *Keychain) AuthFor(ref name.Reference) authn.Authenticator {
	repo := ref.Context()
	if c := k.creds.Match(repo.RegistryStr() + "/" + repo.RepositoryStr()); c != nil {
		return &authn.Basic{Username: c.Username, Password: c.Password}
	}
	auth, err := authn.DefaultKeychain.Resolve(repo)
	if err != nil {
		return authn.Anonymous
	}
	return auth
}

// BasicAuthFor returns the configured credentials for a target, for the chart
// downloads that do not go through an image registry client.
func (k *Keychain) BasicAuthFor(target string) (username, password string) {
	if c := k.creds.Match(target); c != nil {
		return c.Username, c.Password
	}
	return "", ""
}

// BasicAuthForURL returns the credentials for a chart URL, withholding them from
// a plaintext URL unless the matched credential sets insecure.
func (k *Keychain) BasicAuthForURL(source string) (username, password string) {
	c := k.creds.Match(hostOf(source))
	if c == nil {
		return "", ""
	}
	if strings.HasPrefix(source, "http://") && !c.Insecure {
		return "", ""
	}
	return c.Username, c.Password
}

func (k *Keychain) BearerTokenForURL(source string) string {
	c := k.creds.Match(hostOf(source))
	if c == nil || strings.HasPrefix(source, "http://") && !c.Insecure {
		return ""
	}
	return c.BearerToken
}

// ParseReference parses an image reference, marking its registry insecure when
// it is configured that way.
func (k *Keychain) ParseReference(ref string) (name.Reference, error) {
	var opts []name.Option
	if c := k.creds.Match(ref); c != nil && c.Insecure {
		opts = append(opts, name.Insecure)
	}
	parsed, err := name.ParseReference(ref, opts...)
	if err != nil {
		return nil, fmt.Errorf("invalid image reference %q: %w", ref, err)
	}
	return parsed, nil
}

// TransportFor builds the HTTP transport for a target, honouring its CA chain
// and TLS settings. Transports are cached so connections are reused.
func (k *Keychain) TransportFor(target string) (http.RoundTripper, error) {
	c := k.creds.Match(target)
	if c == nil {
		return http.DefaultTransport, nil
	}

	k.mu.Lock()
	defer k.mu.Unlock()
	if t, ok := k.transports[c.Prefix]; ok {
		return t, nil
	}

	transport, err := newTransport(*c)
	if err != nil {
		return nil, err
	}
	k.transports[c.Prefix] = transport
	return transport, nil
}

func newTransport(c Credential) (http.RoundTripper, error) {
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: c.InsecureSkipVerify}
	if c.CAChainPath != "" {
		pem, err := os.ReadFile(c.CAChainPath)
		if err != nil {
			return nil, fmt.Errorf("read caChainPath for %s: %w", c.Prefix, err)
		}
		pool, err := x509.SystemCertPool()
		if err != nil || pool == nil {
			pool = x509.NewCertPool()
		}
		if !pool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("caChainPath for %s contains no certificates", c.Prefix)
		}
		tlsConfig.RootCAs = pool
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = tlsConfig
	return transport, nil
}
