package artifacts

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/run-ai/runai/runai-installer/installer/pkg/artifactlock"
)

type CLIFetcher interface {
	Fetch(context.Context, artifactlock.CLI, string) error
	Probe(context.Context, artifactlock.CLI) error
}

type cliFetcher struct {
	keychain *Keychain
}

func NewCLIFetcher(keychain *Keychain) CLIFetcher {
	return &cliFetcher{keychain: keychain}
}

func (f *cliFetcher) Fetch(ctx context.Context, cli artifactlock.CLI, workDir string) error {
	archive, err := f.download(ctx, cli.Source)
	if err != nil {
		return err
	}
	if digestBytes(archive) != cli.SourceDigest {
		return fmt.Errorf("CLI %s source digest does not match the lock", cli.Name)
	}
	files, err := cliArchiveFiles(archive)
	if err != nil {
		return fmt.Errorf("read CLI archive %s: %w", cli.Name, err)
	}
	locked := cli.Files
	if len(locked) == 0 && cli.SourcePath != "" {
		locked = []artifactlock.CLIFile{{SourcePath: cli.SourcePath, BundlePath: cli.BundlePath, Digest: cli.Digest, Mode: 0o755}}
	}
	for _, file := range locked {
		content, ok := files[file.SourcePath]
		if !ok {
			return fmt.Errorf("CLI %s archive does not contain %s", cli.Name, file.SourcePath)
		}
		if digestBytes(content) != file.Digest {
			return fmt.Errorf("CLI %s file %s digest does not match the lock", cli.Name, file.SourcePath)
		}
		destination, err := CLIPath(workDir, file.BundlePath)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(destination, content, os.FileMode(file.Mode)); err != nil {
			return err
		}
	}
	return nil
}

func (f *cliFetcher) Probe(ctx context.Context, cli artifactlock.CLI) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, cli.Source, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Range", "bytes=0-0")
	f.authenticate(request)
	transport, err := f.keychain.TransportFor(hostOf(cli.Source))
	if err != nil {
		return err
	}
	response, err := (&http.Client{Transport: transport}).Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("%s", response.Status)
	}
	return nil
}

func (f *cliFetcher) download(ctx context.Context, source string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
	if err != nil {
		return nil, err
	}
	f.authenticate(request)
	transport, err := f.keychain.TransportFor(hostOf(source))
	if err != nil {
		return nil, err
	}
	response, err := (&http.Client{Transport: transport}).Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s", response.Status)
	}
	return io.ReadAll(response.Body)
}

func (f *cliFetcher) authenticate(request *http.Request) {
	if token := f.keychain.BearerTokenForURL(request.URL.String()); token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
		return
	}
	if username, password := f.keychain.BasicAuthForURL(request.URL.String()); username != "" {
		request.SetBasicAuth(username, password)
	}
}

func cliArchiveFiles(archive []byte) (map[string][]byte, error) {
	compressed, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, err
	}
	defer compressed.Close()
	reader := tar.NewReader(compressed)
	files := map[string][]byte{}
	for {
		header, err := reader.Next()
		if err == io.EOF {
			return files, nil
		}
		if err != nil {
			return nil, err
		}
		if !header.FileInfo().Mode().IsRegular() {
			continue
		}
		name := strings.TrimPrefix(path.Clean(header.Name), "./")
		if path.IsAbs(name) || name == "." || name == ".." || strings.HasPrefix(name, "../") {
			return nil, fmt.Errorf("invalid archive path %q", header.Name)
		}
		if _, exists := files[name]; exists {
			return nil, fmt.Errorf("duplicate archive path %q", name)
		}
		content, err := io.ReadAll(reader)
		if err != nil {
			return nil, err
		}
		files[name] = content
	}
}

func digestBytes(content []byte) string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256(content))
}
