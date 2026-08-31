/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package integrity

import (
	"path/filepath"
	"slices"
	"sort"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/logger"
	"sigs.k8s.io/yaml"
)

// runaiAuthMiddleware is the Traefik middleware that fronts a router with the
// authentication-handler forwardAuth. A router without it is served to
// unauthenticated callers.
const runaiAuthMiddleware = "runai-auth"

// publicRouters are the Traefik routers in templates/traefik/configmap.yaml that
// are intentionally exposed WITHOUT runai-auth. Adding a name here is a
// conscious, reviewable decision to serve traffic unauthenticated -- either the
// route is public by design or auth is enforced downstream (keycloak,
// researcher-service, cluster-api). Every other router MUST carry runai-auth;
// TestTraefikRoutersRequireAuth fails the build otherwise, so a new route that
// forgets the middleware can't be silently exposed.
var publicRouters = map[string]string{
	"frontend-router":                                  "public web UI",
	"identity-manager-unauthenticated-router":          "pre-auth identity endpoints (login, tenant discovery)",
	"identity-manager-unauthenticated-hostname-router": "pre-auth identity endpoints resolved by hostname",
	"cli-exposer-unauthenticated-router":               "CLI bootstrap endpoints used before a token exists",
	"notifications-proxy-unauthenticated-router":       "unauthenticated notifications endpoints",
	"bff-service-unauthenticated-router":               "unauthenticated BFF endpoints",
	"unleash-edge-router":                              "feature-flag frontend API (/api/frontend)",
	"keycloak-router":                                  "keycloak enforces its own auth",
	"keycloak-workload-router":                         "keycloak workload flow, guarded by workload-redirect-validation",
	"oidc-error-router":                                "keycloak OIDC broker error handling",
	"researcher-router":                                "proxied to researcher-service, auth enforced downstream",
	"cluster-api-router":                               "proxied to cluster-api backend, auth enforced downstream",
	"thanos-router":                                    "internal metrics remote-write receive",
}

type traefikRoutes struct {
	HTTP struct {
		Routers map[string]struct {
			Middlewares []string `json:"middlewares"`
		} `json:"routers"`
	} `json:"http"`
}

// TestTraefikRoutersRequireAuth renders the Traefik ConfigMap with every
// optional router enabled and asserts that each router either carries the
// runai-auth middleware or is on the publicRouters allowlist. It also flags
// stale allowlist entries so the list can't rot as routers are renamed/removed.
func TestTraefikRoutersRequireAuth(t *testing.T) {
	chartAbs, err := filepath.Abs(chartDir)
	if err != nil {
		t.Fatalf("resolve chart dir: %v", err)
	}

	opts := &helm.Options{
		// Enable every value-gated router so they all render and get checked.
		SetValues: map[string]string{
			"thanos.enabled":              "true",
			"tempBackendSettings.enabled": "true",
			"global.multiTenancy.enabled": "true",
		},
		Logger: logger.Discard,
	}
	rendered := helm.RenderTemplate(t, opts, chartAbs, releaseName,
		[]string{"templates/traefik/configmap.yaml"})

	routes := extractTraefikRoutes(t, rendered)
	if len(routes.HTTP.Routers) == 0 {
		t.Fatalf("no routers found in rendered traefik routes.yaml")
	}

	seen := map[string]bool{}
	var missing []string
	for name, r := range routes.HTTP.Routers {
		seen[name] = true
		if _, public := publicRouters[name]; public {
			continue
		}
		if !slices.Contains(r.Middlewares, runaiAuthMiddleware) {
			missing = append(missing, name)
		}
	}
	sort.Strings(missing)
	for _, name := range missing {
		t.Errorf("router %q has no %q middleware and is not on the publicRouters allowlist -- "+
			"add the middleware, or if the route is intentionally public add it to publicRouters "+
			"in traefik_auth_test.go with a justification", name, runaiAuthMiddleware)
	}

	var stale []string
	for name := range publicRouters {
		if !seen[name] {
			stale = append(stale, name)
		}
	}
	sort.Strings(stale)
	for _, name := range stale {
		t.Errorf("publicRouters allowlist entry %q does not match any rendered router -- "+
			"remove the stale entry from traefik_auth_test.go", name)
	}
}

func extractTraefikRoutes(t *testing.T, rendered string) traefikRoutes {
	t.Helper()
	manifests := parseManifests(t, rendered)
	for _, cm := range manifests.ConfigMaps {
		raw, ok := cm.Data["routes.yaml"]
		if !ok {
			continue
		}
		var routes traefikRoutes
		if err := yaml.Unmarshal([]byte(raw), &routes); err != nil {
			t.Fatalf("parse routes.yaml from ConfigMap %q: %v", cm.Name, err)
		}
		return routes
	}
	t.Fatalf("no ConfigMap with a routes.yaml key in rendered output")
	return traefikRoutes{}
}
