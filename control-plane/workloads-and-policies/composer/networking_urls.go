/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer

import (
	"errors"
	"fmt"
	"maps"
	"net/url"
	"slices"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"k8s.io/utils/ptr"
)

// ComposeAutoExposedURL keeps connection metadata, proxy redirects and networking resources aligned.
func ComposeAutoExposedURL(clusterDomain, projectName, workloadName, suffix string, subdomainSupported bool) (*url.URL, error) {
	if clusterDomain == "" {
		return nil, errors.New("invalid cluster domain: empty")
	}
	parsed, err := url.Parse(clusterDomain)
	if err != nil {
		return nil, fmt.Errorf("invalid cluster domain %q: %w", clusterDomain, err)
	}

	if subdomainSupported {
		subdomain := fmt.Sprintf("%s-%s", projectName, workloadName)
		if suffix != "" {
			subdomain += "-" + suffix
		}
		parsed.Host = shortenNetworkingName(subdomain) + "." + parsed.Host
	} else {
		path := fmt.Sprintf("/%s/%s", projectName, workloadName)
		if suffix != "" {
			path += "/" + suffix
		}
		if parsed.Path != "" {
			// Preserve the existing path concatenation, including its slash boundaries.
			path += "/" + parsed.Path
		}
		parsed.Path = path
	}
	if parsed.Path == "" {
		parsed.Path = "/"
	}
	return parsed, nil
}

func exposedURLsPerNick(exposedURLs *field_types.ExposedUrlsField, projectName, workloadName, clusterDomain string,
	subdomainSupported bool,
) (map[string]*url.URL, error) {
	result := make(map[string]*url.URL)
	noSuffixGenerated := false
	items := exposedURLs.Deref()
	for _, nick := range slices.Sorted(maps.Keys(items)) {
		item := items[nick].Value
		if item == nil {
			continue
		}
		customEnabled := ptr.Deref(item.CustomUrl, false)
		customURL := ptr.Deref(item.Url, "")
		if customEnabled && customURL == "" {
			return nil, errors.New("url is mandatory when customUrl is enabled")
		}
		if !customEnabled && customURL != "" {
			return nil, errors.New("url should only be provided for customUrl")
		}

		var parsed *url.URL
		var err error
		if customURL != "" {
			parsed, err = url.Parse(customURL)
		} else {
			suffix := nick
			if !noSuffixGenerated {
				suffix = ""
				noSuffixGenerated = true
			}
			parsed, err = ComposeAutoExposedURL(clusterDomain, projectName, workloadName, suffix, subdomainSupported)
		}
		if err != nil {
			return nil, err
		}
		result[nick] = parsed
	}
	return result, nil
}
