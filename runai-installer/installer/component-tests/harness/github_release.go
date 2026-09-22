/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package harness

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/pkg/artifactlock"
	"gopkg.in/yaml.v3"
)

type GitHubRelease struct {
	Tag        string
	Draft      bool
	Prerelease bool
	Assets     map[string][]byte
}

func (r GitHubRelease) WithLock(arch string, lock *artifactlock.ArtifactLock) GitHubRelease {
	out, err := yaml.Marshal(lock)
	if err != nil {
		panic(err)
	}
	if r.Assets == nil {
		r.Assets = map[string][]byte{}
	}
	r.Assets[fmt.Sprintf("runai-%s-standard-linux-%s.yaml", r.Tag, arch)] = out
	return r
}

type GitHubReleaseServer struct {
	*httptest.Server
	Repository string

	mu       sync.Mutex
	releases []GitHubRelease
	requests []*http.Request
}

func NewGitHubReleaseServer(t *testing.T, repository string, releases ...GitHubRelease) *GitHubReleaseServer {
	t.Helper()
	s := &GitHubReleaseServer{Repository: repository, releases: releases}
	s.Server = httptest.NewServer(http.HandlerFunc(s.handle))
	t.Cleanup(s.Close)
	return s
}

func (s *GitHubReleaseServer) Requests() []*http.Request {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]*http.Request(nil), s.requests...)
}

func (s *GitHubReleaseServer) DownloadURL(tag, name string) string {
	return fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", s.Repository, tag, name)
}

func (s *GitHubReleaseServer) AssetName(tag, arch string) string {
	return fmt.Sprintf("runai-%s-standard-linux-%s.yaml", tag, arch)
}

func (s *GitHubReleaseServer) handle(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	s.requests = append(s.requests, r)
	s.mu.Unlock()

	listPath := "/repos/" + s.Repository + "/releases"
	assetPrefix := listPath + "/assets/"
	switch {
	case r.URL.Path == listPath:
		s.serveList(w, r)
	case len(r.URL.Path) > len(assetPrefix) && r.URL.Path[:len(assetPrefix)] == assetPrefix:
		s.serveAsset(w, r, r.URL.Path[len(assetPrefix):])
	default:
		http.NotFound(w, r)
	}
}

func (s *GitHubReleaseServer) serveList(w http.ResponseWriter, r *http.Request) {
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if perPage <= 0 {
		perPage = 30
	}
	if page <= 0 {
		page = 1
	}

	type asset struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		URL         string `json:"url"`
		DownloadURL string `json:"browser_download_url"`
	}
	type release struct {
		TagName    string  `json:"tag_name"`
		Draft      bool    `json:"draft"`
		Prerelease bool    `json:"prerelease"`
		Assets     []asset `json:"assets"`
	}

	var out []release
	start := (page - 1) * perPage
	for i := start; i < len(s.releases) && i < start+perPage; i++ {
		rel := s.releases[i]
		item := release{TagName: rel.Tag, Draft: rel.Draft, Prerelease: rel.Prerelease, Assets: []asset{}}
		for name := range rel.Assets {
			id := i*1000 + len(item.Assets) + 1
			item.Assets = append(item.Assets, asset{
				ID:          id,
				Name:        name,
				URL:         fmt.Sprintf("%s/repos/%s/releases/assets/%d/%s", s.URL, s.Repository, id, name),
				DownloadURL: s.DownloadURL(rel.Tag, name),
			})
		}
		out = append(out, item)
	}
	if out == nil {
		out = []release{}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func (s *GitHubReleaseServer) serveAsset(w http.ResponseWriter, r *http.Request, rest string) {
	var name string
	for i := len(rest) - 1; i >= 0; i-- {
		if rest[i] == '/' {
			name = rest[i+1:]
			break
		}
	}
	if r.Header.Get("Accept") != "application/octet-stream" {
		http.Error(w, "asset downloads require Accept: application/octet-stream", http.StatusUnsupportedMediaType)
		return
	}
	for _, rel := range s.releases {
		if content, ok := rel.Assets[name]; ok {
			w.Header().Set("Content-Type", "application/octet-stream")
			_, _ = w.Write(content)
			return
		}
	}
	http.NotFound(w, r)
}
