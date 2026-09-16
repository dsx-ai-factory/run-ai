/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package releaselock

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/pkg/artifactlock"
)

const (
	DefaultRepository = "dsx-ai-factory/run-ai"
	DefaultAPIBaseURL = "https://api.github.com"

	releasesPerPage = 100
	maxPages        = 10
	requestTimeout  = 2 * time.Minute
)

var releaseVersion = regexp.MustCompile(`^v?(\d+)\.(\d+)(?:\.(\d+))?$`)

var ErrNotReleaseVersion = errors.New("not a release version")

type NotFoundError struct {
	Version string
	Tag     string
	Asset   string
}

func (e *NotFoundError) Error() string {
	if e.Tag == "" {
		return fmt.Sprintf("release for version %s not found", e.Version)
	}
	return fmt.Sprintf("release %s has no lock file %s", e.Tag, e.Asset)
}

type Result struct {
	Lock  *artifactlock.ArtifactLock
	Tag   string
	Asset string
	URL   string
}

type Fetcher struct {
	APIBaseURL string
	Repository string
	HTTPClient *http.Client
}

func NewFetcher() *Fetcher {
	return &Fetcher{
		APIBaseURL: DefaultAPIBaseURL,
		Repository: DefaultRepository,
		HTTPClient: &http.Client{Timeout: requestTimeout},
	}
}

func AssetName(tag, arch string) string {
	return fmt.Sprintf("runai-%s-standard-linux-%s.yaml", tag, arch)
}

func IsReleaseVersion(version string) bool {
	return releaseVersion.MatchString(version)
}

type release struct {
	TagName    string  `json:"tag_name"`
	Draft      bool    `json:"draft"`
	Prerelease bool    `json:"prerelease"`
	Assets     []asset `json:"assets"`
}

type asset struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	DownloadURL string `json:"browser_download_url"`
}

func (f *Fetcher) Fetch(ctx context.Context, version, arch string) (*Result, error) {
	match := releaseVersion.FindStringSubmatch(version)
	if match == nil {
		return nil, fmt.Errorf("%w: %q", ErrNotReleaseVersion, version)
	}

	releases, err := f.listReleases(ctx)
	if err != nil {
		return nil, err
	}

	selected := selectRelease(releases, match)
	if selected == nil {
		return nil, &NotFoundError{Version: version}
	}

	name := AssetName(selected.TagName, arch)
	var found *asset
	for i := range selected.Assets {
		if selected.Assets[i].Name == name {
			found = &selected.Assets[i]
			break
		}
	}
	if found == nil {
		return nil, &NotFoundError{Version: version, Tag: selected.TagName, Asset: name}
	}

	lock, err := f.downloadLock(ctx, found)
	if err != nil {
		return nil, err
	}
	return &Result{Lock: lock, Tag: selected.TagName, Asset: name, URL: found.DownloadURL}, nil
}

func selectRelease(releases []release, match []string) *release {
	major, _ := strconv.ParseUint(match[1], 10, 64)
	minor, _ := strconv.ParseUint(match[2], 10, 64)
	exact := match[3] != ""
	var patch uint64
	if exact {
		patch, _ = strconv.ParseUint(match[3], 10, 64)
	}

	var best *release
	var bestVersion *semver.Version
	for i := range releases {
		r := &releases[i]
		if r.Draft || r.Prerelease {
			continue
		}
		v, err := semver.StrictNewVersion(r.TagName)
		if err != nil || v.Prerelease() != "" || v.Major() != major || v.Minor() != minor {
			continue
		}
		if exact {
			if v.Patch() == patch {
				return r
			}
			continue
		}
		if bestVersion == nil || v.GreaterThan(bestVersion) {
			best, bestVersion = r, v
		}
	}
	return best
}

func (f *Fetcher) listReleases(ctx context.Context) ([]release, error) {
	var all []release
	for page := 1; page <= maxPages; page++ {
		endpoint := fmt.Sprintf("%s/repos/%s/releases?per_page=%d&page=%d", f.APIBaseURL, f.Repository, releasesPerPage, page)
		body, err := f.get(ctx, endpoint, "application/vnd.github+json")
		if err != nil {
			return nil, fmt.Errorf("list releases of %s: %w", f.Repository, err)
		}
		var releases []release
		if err := json.Unmarshal(body, &releases); err != nil {
			return nil, fmt.Errorf("list releases of %s: decode response: %w", f.Repository, err)
		}
		all = append(all, releases...)
		if len(releases) < releasesPerPage {
			break
		}
	}
	return all, nil
}

func (f *Fetcher) downloadLock(ctx context.Context, a *asset) (*artifactlock.ArtifactLock, error) {
	if _, err := url.Parse(a.URL); err != nil || a.URL == "" {
		return nil, fmt.Errorf("release asset %s has no download URL", a.Name)
	}
	body, err := f.get(ctx, a.URL, "application/octet-stream")
	if err != nil {
		return nil, fmt.Errorf("download %s: %w", a.Name, err)
	}
	return artifactlock.Parse(bytes.NewReader(body), a.Name)
}

func (f *Fetcher) get(ctx context.Context, endpoint, accept string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", accept)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	client := f.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: requestTimeout}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", endpoint, resp.Status)
	}
	return body, nil
}
