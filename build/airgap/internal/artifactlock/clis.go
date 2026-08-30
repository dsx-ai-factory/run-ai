package artifactlock

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/run-ai/runai/build/airgap/internal/metadata"
	"github.com/run-ai/runai/build/airgap/pkg/lock"
)

type Downloader interface {
	Download(context.Context, string, DownloadCredentials) ([]byte, error)
}

type DownloadCredentials struct {
	Username    string
	Password    string
	BearerToken string
}

type CLISource struct {
	Version     string
	URL         string
	Credentials DownloadCredentials
}

type HTTPDownloader struct {
	Client *http.Client
}

func (downloader HTTPDownloader) Download(ctx context.Context, source string, credentials DownloadCredentials) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
	if err != nil {
		return nil, err
	}
	if credentials.BearerToken != "" {
		request.Header.Set("Authorization", "Bearer "+credentials.BearerToken)
	} else if credentials.Username != "" {
		request.SetBasicAuth(credentials.Username, credentials.Password)
	}
	client := downloader.Client
	if client == nil {
		client = &http.Client{Timeout: 120 * time.Second}
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("NGC download failed for %s", source)
	}
	defer func() {
		_ = response.Body.Close()
	}()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("NGC download failed for %s: %s", source, response.Status)
	}
	contents, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("NGC download failed for %s", source)
	}
	return contents, nil
}

func ResolveCLIs(ctx context.Context, roots []string, version, resourceBaseURL, token string, sources map[string]CLISource, downloader Downloader) ([]lock.CLI, error) {
	if len(roots) == 0 {
		return []lock.CLI{}, nil
	}
	artifacts, err := metadata.Discover(roots, "cli")
	if err != nil {
		return nil, err
	}
	if len(artifacts) == 0 {
		return nil, fmt.Errorf("no CLI artifacts were discovered")
	}
	clis := make([]lock.CLI, 0, len(artifacts))
	for _, artifact := range artifacts {
		resolved := sources[artifact.ServiceName]
		if resolved.Version == "" {
			resolved.Version = version
		}
		if resolved.URL == "" {
			resolved.URL = fmt.Sprintf(
				"%s/%s/versions/%s/files/%s.tgz",
				strings.TrimRight(resourceBaseURL, "/"),
				artifact.ServiceName,
				resolved.Version,
				artifact.ServiceName,
			)
			resolved.Credentials.BearerToken = token
		}
		archive, err := downloader.Download(ctx, resolved.URL, resolved.Credentials)
		if err != nil {
			return nil, err
		}
		files, err := readTarGzipEntries(archive)
		if err != nil {
			return nil, fmt.Errorf("read CLI archive %s: %w", resolved.URL, err)
		}
		var binaries []lock.CLIFile
		seen := map[string]bool{}
		for _, file := range files {
			clean := strings.TrimPrefix(path.Clean(file.Name), "./")
			if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") || path.IsAbs(clean) {
				return nil, fmt.Errorf("invalid CLI archive path %q", file.Name)
			}
			if file.Mode.Perm()&0o111 == 0 {
				continue
			}
			if seen[clean] {
				return nil, fmt.Errorf("duplicate CLI archive path %q", clean)
			}
			seen[clean] = true
			binaries = append(binaries, lock.CLIFile{
				SourcePath: clean,
				BundlePath: "cli/" + artifact.ServiceName + "/" + clean,
				Digest:     lock.Digest(file.Contents),
				Mode:       uint32(file.Mode.Perm()),
			})
		}
		if len(binaries) == 0 {
			return nil, fmt.Errorf("CLI archive %s contains no executable files", resolved.URL)
		}
		sort.Slice(binaries, func(i, j int) bool {
			return binaries[i].SourcePath < binaries[j].SourcePath
		})
		clis = append(clis, lock.CLI{
			Name:         artifact.ServiceName,
			Version:      resolved.Version,
			Source:       resolved.URL,
			SourceDigest: lock.Digest(archive),
			Files:        binaries,
		})
	}
	sort.Slice(clis, func(i, j int) bool {
		return clis[i].Name < clis[j].Name
	})
	return clis, nil
}
