/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"log"
	"net/url"
	"strings"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/common"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/global"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// BuildParentRef creates a Gateway API ParentReference from the GatewayAPI config.
// If override is non-nil and has a non-empty Name, it takes precedence over the global config.
func BuildParentRef(gatewayAPI *global.GatewayAPI, override *common.GatewayRef) gwv1.ParentReference {
	name := ""
	namespace := ""

	if override != nil && override.Name != "" {
		name = override.Name
		namespace = override.Namespace
	} else if gatewayAPI != nil {
		name = gatewayAPI.Name
		namespace = gatewayAPI.Namespace
	}

	if name == "" {
		return gwv1.ParentReference{}
	}

	ref := gwv1.ParentReference{
		Name: gwv1.ObjectName(name),
	}
	if namespace != "" {
		ns := gwv1.Namespace(namespace)
		ref.Namespace = &ns
	}
	return ref
}

// SecurityHeaderFilter returns an HTTPRoute filter that sets security response headers.
func SecurityHeaderFilter() gwv1.HTTPRouteFilter {
	return gwv1.HTTPRouteFilter{
		Type: gwv1.HTTPRouteFilterResponseHeaderModifier,
		ResponseHeaderModifier: &gwv1.HTTPHeaderFilter{
			Set: []gwv1.HTTPHeader{
				{Name: "X-Content-Type-Options", Value: "nosniff"},
				{Name: "X-Frame-Options", Value: "DENY"},
				{Name: "Referrer-Policy", Value: "strict-origin-when-cross-origin"},
			},
		},
	}
}

// HTTPRouteHostname parses the hostname from a cluster URL for use in HTTPRoute.
func HTTPRouteHostname(clusterURL string) gwv1.Hostname {
	parsed := ParseClusterDomain(clusterURL)
	if parsed == nil {
		return ""
	}
	return gwv1.Hostname(parsed.Host)
}

// ParseClusterDomain parses a cluster domain URL, handling missing scheme.
func ParseClusterDomain(clusterDomain string) *url.URL {
	domain, err := url.Parse(clusterDomain)
	if err != nil {
		log.Printf("failed to parse cluster domain '%s'. error: %s", clusterDomain, err.Error())
		return nil
	}
	if domain.Host == "" {
		// If clusterDomain doesn't have a scheme - the parser treats it as a relative path...
		if domain, err = url.Parse("https://" + clusterDomain); err != nil {
			log.Printf("failed to parse cluster domain '%s', after prefixing with 'https://'. error: %s", "https://"+clusterDomain, err.Error())
			return nil
		}
	}

	// ignore port when creating ingress (Hostname() handles IPv6 correctly)
	domain.Host = domain.Hostname()

	return domain
}

// GetIngressPath defines the ingress path for a given cluster domain and prefix.
func GetIngressPath(clusterDomain, prefix string) string {
	defaultPath := strings.TrimSuffix("/"+prefix, "/")
	if clusterDomain == "" {
		return defaultPath
	} else {
		parsedDomain := ParseClusterDomain(clusterDomain)
		if parsedDomain != nil && parsedDomain.Path != "" && parsedDomain.Path != "/" {
			return strings.TrimSuffix(parsedDomain.Path, "/")
		} else {
			return defaultPath
		}
	}
}
