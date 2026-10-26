/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package helm_deployer

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/rs/zerolog/log"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/registry"
)

const (
	ChartsDirName        = "helm-charts"
	ChartsDirPermissions = 0750
)

type DownloaderFactory func(*ReleaseClient, *InstallParams) ChartDownloader

type ChartDownloader interface {
	Download() (string, error)

	// Cleanup removes files Download wrote for the chart. Call it once the chart is loaded.
	Cleanup() error
}

// Downloader is responsible for downloading charts or helm related stuff
type Downloader struct {
	// either CharRef or ChartUrl is expected
	ChartRef *v1alpha1.ChartRef
	ChartUrl string

	Settings *cli.EnvSettings

	AuthOptions *AuthOptions

	// RequiresControlPlaneAuth marks ChartUrl as a control-plane-hosted chart that must be fetched
	// with the cluster's control-plane bearer token. ChartAuthClient carries that token; when it is
	// nil, a flagged chart fails fast rather than downloading unauthenticated.
	RequiresControlPlaneAuth bool
	ChartAuthClient          *http.Client

	// tempDir holds a chart Download fetched itself; charts located by Helm stay in Helm's cache.
	tempDir string
}

func NewDownloader(releaseClient *ReleaseClient, params *InstallParams) ChartDownloader {
	return &Downloader{
		ChartRef: params.chartRef,
		ChartUrl: params.chartUrl,

		Settings: releaseClient.Settings,

		AuthOptions: params.authOpts,

		RequiresControlPlaneAuth: params.requiresControlPlaneAuth,
	}
}

// NewAuthenticatedDownloaderFactory returns a DownloaderFactory that injects chartAuthClient (a
// control-plane bearer-token HTTP client) into every Downloader it builds. Charts that do not set
// RequiresControlPlaneAuth are downloaded exactly as before.
func NewAuthenticatedDownloaderFactory(chartAuthClient *http.Client) DownloaderFactory {
	return func(releaseClient *ReleaseClient, params *InstallParams) ChartDownloader {
		downloader := NewDownloader(releaseClient, params).(*Downloader)
		downloader.ChartAuthClient = chartAuthClient
		return downloader
	}
}

func (d *Downloader) Download() (string, error) {
	// ChartURL: direct download from URL (supports auth)
	if d.ChartUrl != "" {
		if d.RequiresControlPlaneAuth {
			return d.downloadControlPlaneChart()
		}
		return d.downloadChartFromURL()
	}
	if d.ChartRef == nil {
		return "", fmt.Errorf("either chartUrl or chartRef must be set")
	}
	if strings.HasPrefix(d.ChartRef.Repository, "oci://") {
		return d.downloadOCIChart()
	}
	return d.downloadHTTPChart()
}

// downloadChartFromURL downloads a chart from a direct URL using Helm's LocateChart
// (same getter + auth path as HTTP repos; supports basic auth, TLS, insecure skip).
func (d *Downloader) downloadChartFromURL() (string, error) {
	chartURL := strings.TrimSpace(d.ChartUrl)
	if chartURL == "" {
		return "", fmt.Errorf("chartUrl is empty")
	}
	if _, err := url.Parse(chartURL); err != nil {
		return "", fmt.Errorf("invalid chartUrl: %w", err)
	}

	if isLocalFileUrl(chartURL) {
		return "", fmt.Errorf("file://... is not supported for chartUrl")
	}

	cpo := d.buildChartPathOptions("", "")

	chartPath, err := cpo.LocateChart(chartURL, d.Settings)
	if err != nil {
		return "", fmt.Errorf("locate chart from URL: %w", err)
	}

	log.Info().Msgf("Download --> chart located via Helm SDK %s -> %s", chartURL, chartPath)

	return chartPath, nil
}

// downloadControlPlaneChart fetches a control-plane-hosted chart over HTTPS using the cluster's
// bearer token and writes it to a temp .tgz for Helm to load. Helm's ChartPathOptions can only
// carry basic-auth/TLS, not a bearer token, so this path bypasses LocateChart and downloads the
// packaged chart directly. HTTPS is required so the token is never sent in the clear.
func (d *Downloader) downloadControlPlaneChart() (_ string, err error) {
	chartURL := strings.TrimSpace(d.ChartUrl)
	if chartURL == "" {
		return "", fmt.Errorf("chartUrl is empty")
	}
	parsed, err := url.Parse(chartURL)
	if err != nil {
		return "", fmt.Errorf("invalid chartUrl: %w", err)
	}
	if parsed.Scheme != "https" {
		return "", fmt.Errorf("control-plane chart download requires https, got scheme %q", parsed.Scheme)
	}
	if d.ChartAuthClient == nil {
		return "", fmt.Errorf("control-plane chart authentication is not configured")
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, chartURL, nil)
	if err != nil {
		return "", fmt.Errorf("build control-plane chart request: %w", err)
	}

	resp, err := d.ChartAuthClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("download control-plane chart: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download control-plane chart: unexpected status %s", resp.Status)
	}

	destDir, err := os.MkdirTemp("", ChartsDirName+"-")
	if err != nil {
		return "", fmt.Errorf("create control-plane chart temp dir: %w", err)
	}
	d.tempDir = destDir
	defer func() {
		if err != nil {
			err = errors.Join(err, d.Cleanup())
		}
	}()

	chartFile, err := os.CreateTemp(destDir, "cp-chart-*.tgz")
	if err != nil {
		return "", fmt.Errorf("create control-plane chart temp file: %w", err)
	}

	_, copyErr := io.Copy(chartFile, resp.Body)
	if err := errors.Join(copyErr, chartFile.Close()); err != nil {
		return "", fmt.Errorf("write control-plane chart temp file: %w", err)
	}

	log.Info().Msgf("Download --> control-plane chart fetched with bearer token -> %s", chartFile.Name())
	return chartFile.Name(), nil
}

// Cleanup removes the temp dir of a chart fetched by downloadControlPlaneChart, if any.
func (d *Downloader) Cleanup() error {
	if d.tempDir == "" {
		return nil
	}
	if err := os.RemoveAll(d.tempDir); err != nil {
		return fmt.Errorf("failed to remove chart temp dir %s: %w", d.tempDir, err)
	}
	d.tempDir = ""
	return nil
}

// downloadOCIChart downloads a chart from an OCI registry using LocateChart.
//
// Auth for OCI: RegistryConfigFile (and default registry config) is always used.
// If AuthOptions has Username/Password, they are also passed to the registry client
// (ClientOptBasicAuth). TLS options (CAFile, CertFile, KeyFile, InsecureSkipTLSVerify)
// are set on ChartPathOptions for LocateChart, but the OCI pull uses the registry client;
// custom TLS for OCI would require building the registry client with ClientOptHTTPClient
// and is not wired here.
func (d *Downloader) downloadOCIChart() (string, error) {
	clientOpts := []registry.ClientOption{
		registry.ClientOptDebug(d.Settings.Debug),
	}

	if d.AuthOptions != nil && d.AuthOptions.RegistryConfigFile != "" {
		clientOpts = append(clientOpts, registry.ClientOptCredentialsFile(d.AuthOptions.RegistryConfigFile))
	} else {
		clientOpts = append(clientOpts, registry.ClientOptCredentialsFile(d.Settings.RegistryConfig))
	}

	if d.AuthOptions != nil && d.AuthOptions.Username != "" && d.AuthOptions.Password != "" {
		clientOpts = append(clientOpts, registry.ClientOptBasicAuth(d.AuthOptions.Username, d.AuthOptions.Password))
	}

	registryClient, err := registry.NewClient(clientOpts...)
	if err != nil {
		return "", fmt.Errorf("failed to create registry client: %w", err)
	}

	ociRef := fmt.Sprintf("%s/%s", strings.TrimSuffix(d.ChartRef.Repository, "/"), d.ChartRef.Name)
	if d.ChartRef.Version != "" {
		ociRef = fmt.Sprintf("%s:%s", ociRef, d.ChartRef.Version)
	}

	cfg := &action.Configuration{RegistryClient: registryClient}

	install := action.NewInstall(cfg)

	cpo := d.buildChartPathOptions("", d.ChartRef.Version)

	install.Version = cpo.Version
	install.Username = cpo.Username
	install.Password = cpo.Password
	install.CaFile = cpo.CaFile
	install.CertFile = cpo.CertFile
	install.KeyFile = cpo.KeyFile
	install.InsecureSkipTLSverify = cpo.InsecureSkipTLSverify

	chartPath, err := install.LocateChart(ociRef, d.Settings)
	if err != nil {
		return "", fmt.Errorf("locate OCI chart %s: %w", ociRef, err)
	}

	log.Info().Msgf("Download --> OCI chart located via Helm SDK %s -> %s", ociRef, chartPath)

	return chartPath, nil
}

// downloadHTTPChart downloads a chart from an HTTP/HTTPS repository using Helm's
// LocateChart (index fetch, version resolution, and download with auth/TLS).
func (d *Downloader) downloadHTTPChart() (string, error) {
	cpo := d.buildChartPathOptions(strings.TrimSuffix(d.ChartRef.Repository, "/"), d.ChartRef.Version)

	chartPath, err := cpo.LocateChart(d.ChartRef.Name, d.Settings)
	if err != nil {
		return "", fmt.Errorf("locate chart %s: %w", d.ChartRef.Name, err)
	}

	log.Info().Msgf("Download --> chart located (index + download via Helm SDK) -> %s", chartPath)

	return chartPath, nil
}

// buildChartPathOptions returns Helm ChartPathOptions from repo URL, version, and AuthOptions.
// For direct URL downloads pass repoURL and version as empty.
func (d *Downloader) buildChartPathOptions(repoURL, version string) action.ChartPathOptions {
	cpo := action.ChartPathOptions{
		RepoURL:               repoURL,
		Version:               version,
		CaFile:                "",
		CertFile:              "",
		KeyFile:               "",
		InsecureSkipTLSverify: false,
		PlainHTTP:             false,
		PassCredentialsAll:    false,
		Username:              "",
		Password:              "",
	}
	if d.AuthOptions != nil {
		cpo.CaFile = d.AuthOptions.CAFile
		cpo.CertFile = d.AuthOptions.CertFile
		cpo.KeyFile = d.AuthOptions.KeyFile
		cpo.InsecureSkipTLSverify = d.AuthOptions.InsecureSkipTLSVerify
		cpo.Username = d.AuthOptions.Username
		cpo.Password = d.AuthOptions.Password
	}
	return cpo
}

type DownloaderForTest struct {
	ChartUrl string
}

func NewDownloaderForTest(_ *ReleaseClient, params *InstallParams) ChartDownloader {
	return &DownloaderForTest{
		ChartUrl: params.chartUrl,
	}
}

func (d *DownloaderForTest) Cleanup() error {
	return nil
}

func (d *DownloaderForTest) Download() (string, error) {
	if !isLocalFileUrl(d.ChartUrl) {
		return "", fmt.Errorf("file://... is expected for chartUrl in tests")
	}

	pathStr := strings.TrimPrefix(d.ChartUrl, "file://")

	info, err := os.Stat(pathStr)
	if err != nil {
		return "", fmt.Errorf("locate chart from file URL %s: %w", d.ChartUrl, err)
	}

	if !info.IsDir() {
		// Allow a .tgz file
		if filepath.Ext(pathStr) != ".tgz" {
			return "", fmt.Errorf("chartUrl file:// must be a directory or .tgz file: %s", d.ChartUrl)
		}
	}

	abs, err := filepath.Abs(pathStr)
	if err != nil {
		return "", fmt.Errorf("chartUrl file://: %w", err)
	}

	log.Info().Msgf("DownloadForTest --> chart located from file URL %s -> %s", d.ChartUrl, abs)
	return abs, nil
}

// isLocalFileUrl returns true if the URL is a local file path.
// in tests, we work with local files, but in production we return error if the url points
// to a local file, because it is a security breach.
func isLocalFileUrl(urlStr string) bool {
	return strings.HasPrefix(urlStr, "file://")
}
