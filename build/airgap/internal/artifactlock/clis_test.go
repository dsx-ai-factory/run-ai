/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package artifactlock

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dsx-ai-factory/run-ai/build/airgap/pkg/lock"
)

type fakeDownloader struct {
	archive     []byte
	credentials DownloadCredentials
	source      string
}

func (downloader *fakeDownloader) Download(_ context.Context, source string, credentials DownloadCredentials) ([]byte, error) {
	downloader.source = source
	downloader.credentials = credentials
	return downloader.archive, nil
}

func TestResolveCLIsLocksAllExecutableFiles(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "installer")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	metadata := `{"service_name":"runai-installer","artifact_type":"cli","path_to_artifacts":"bin"}`
	if err := os.WriteFile(filepath.Join(directory, "artifact-metadata.json"), []byte(metadata), 0o644); err != nil {
		t.Fatal(err)
	}
	binary := []byte("installer-binary")
	archive := archiveEntries(t, []archiveFile{
		{Name: "linux/amd64/runai-installer", Contents: binary, Mode: 0o755},
		{Name: "linux/arm64/runai-installer", Contents: []byte("arm"), Mode: 0o755},
		{Name: "darwin/arm64/runai-installer", Contents: []byte("darwin"), Mode: 0o755},
		{Name: "README.md", Contents: []byte("docs"), Mode: 0o644},
	})
	downloader := &fakeDownloader{archive: archive}

	clis, err := ResolveCLIs(
		context.Background(),
		[]string{root},
		"3.0.0",
		"https://api.example.com/resources/",
		"secret",
		nil,
		downloader,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(clis[0].Files) != 3 || clis[0].Files[0].SourcePath != "darwin/arm64/runai-installer" {
		t.Fatalf("CLI = %#v", clis[0])
	}
	if clis[0].SourceDigest != lock.Digest(archive) || clis[0].Files[1].Digest != lock.Digest(binary) {
		t.Fatalf("CLI = %#v", clis[0])
	}
	if downloader.credentials.BearerToken != "secret" || !strings.HasSuffix(downloader.source, "/runai-installer/versions/3.0.0/files/runai-installer.tgz") {
		t.Fatalf("download = %s, %#v", downloader.source, downloader.credentials)
	}
}

func TestResolveCLIsRejectsDuplicatePlatformBinary(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "installer")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	metadata := `{"service_name":"runai-installer","artifact_type":"cli"}`
	if err := os.WriteFile(filepath.Join(directory, "artifact-metadata.json"), []byte(metadata), 0o644); err != nil {
		t.Fatal(err)
	}
	archive := archiveEntries(t, []archiveFile{
		{Name: "linux/amd64/runai-installer", Contents: []byte("first"), Mode: 0o755},
		{Name: "linux/amd64/runai-installer", Contents: []byte("second"), Mode: 0o755},
	})
	_, err := ResolveCLIs(
		context.Background(),
		[]string{root},
		"3.0.0",
		"https://api.example.com/resources",
		"secret",
		nil,
		&fakeDownloader{archive: archive},
	)
	if err == nil || !strings.Contains(err.Error(), "duplicate CLI archive path") {
		t.Fatalf("error = %v", err)
	}
}

func TestHTTPDownloaderUsesBearerToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer secret" {
			t.Errorf("authorization = %s", request.Header.Get("Authorization"))
		}
		_, _ = fmt.Fprint(response, "archive")
	}))
	defer server.Close()

	contents, err := (HTTPDownloader{}).Download(context.Background(), server.URL, DownloadCredentials{BearerToken: "secret"})
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "archive" {
		t.Fatalf("contents = %q", contents)
	}
}

func TestHTTPDownloaderDoesNotExposeToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		http.Error(response, "secret", http.StatusUnauthorized)
	}))
	defer server.Close()

	_, err := (HTTPDownloader{}).Download(context.Background(), server.URL, DownloadCredentials{BearerToken: "secret"})
	if err == nil || strings.Contains(err.Error(), "secret") {
		t.Fatalf("error = %v", err)
	}
}
