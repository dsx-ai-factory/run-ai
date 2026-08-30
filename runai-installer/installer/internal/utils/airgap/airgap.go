package airgap

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/run-ai/runai/runai-installer/installer/internal/logging"
	"github.com/run-ai/runai/runai-installer/installer/internal/utils/linetail"
)

const (
	CustomEnvFile = "custom-env.yaml"
	SetupScript   = "setup.sh"
	// pushTailLines is how many recent setup.sh output lines are shown live.
	pushTailLines = 5
)

func FindChartFile(dir, prefix string) string {
	chartsDir := filepath.Join(dir, "charts")
	matches, _ := filepath.Glob(filepath.Join(chartsDir, prefix+"-*.tgz"))
	if len(matches) > 0 {
		return matches[0]
	}
	exact := filepath.Join(chartsDir, prefix+".tgz")
	if _, err := os.Stat(exact); err == nil {
		return exact
	}
	return filepath.Join(dir, prefix+".tgz")
}

func ExtractTar(ctx context.Context, tarballPath, destDir string) error {
	f, err := os.Open(tarballPath)
	if err != nil {
		return fmt.Errorf("open tarball: %v", err)
	}
	defer f.Close()

	archive, closeArchive, err := decompress(f)
	if err != nil {
		return err
	}
	defer closeArchive()

	// Confines every write below destDir, including through symlinks planted by
	// earlier entries in the same archive.
	root, err := os.OpenRoot(destDir)
	if err != nil {
		return fmt.Errorf("open dest dir: %v", err)
	}
	defer root.Close()

	tr := tar.NewReader(archive)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		header, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("reading tar: %v", err)
		}

		name := filepath.Clean(filepath.FromSlash(header.Name))
		if filepath.IsAbs(name) || escapesRoot(name) {
			return fmt.Errorf("illegal path in tar: %s", header.Name)
		}
		// Drops setuid and friends; os.Root also rejects modes above 0o777.
		mode := os.FileMode(header.Mode).Perm()

		switch header.Typeflag {
		case tar.TypeDir:
			// 0o700 so a restrictive archive mode cannot block nested entries.
			if err := root.MkdirAll(name, mode|0o700); err != nil {
				return fmt.Errorf("mkdir %s: %v", name, err)
			}
		case tar.TypeReg:
			if err := root.MkdirAll(filepath.Dir(name), 0755); err != nil {
				return fmt.Errorf("mkdir parent %s: %v", name, err)
			}
			outFile, err := root.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
			if err != nil {
				return fmt.Errorf("create file %s: %v", name, err)
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return fmt.Errorf("write file %s: %v", name, err)
			}
			if err := outFile.Close(); err != nil {
				return fmt.Errorf("close file %s: %v", name, err)
			}
		case tar.TypeSymlink:
			// Root.Symlink does not validate its target, and setup.sh later runs
			// outside the sandbox. The target check is lexical, so it only holds
			// once the parents are known to be real directories.
			linkname := filepath.FromSlash(header.Linkname)
			if err := requireRealParents(root, name); err != nil {
				return fmt.Errorf("illegal symlink location in tar: %s: %v", header.Name, err)
			}
			if filepath.IsAbs(linkname) || escapesRoot(filepath.Join(filepath.Dir(name), linkname)) {
				return fmt.Errorf("illegal symlink target in tar: %s -> %s", header.Name, header.Linkname)
			}
			if err := root.Symlink(linkname, name); err != nil {
				return fmt.Errorf("symlink %s: %v", name, err)
			}
		case tar.TypeLink:
			// Hardlink targets are relative to the archive root, not to the entry.
			linkname := filepath.Clean(filepath.FromSlash(header.Linkname))
			if filepath.IsAbs(linkname) || escapesRoot(linkname) {
				return fmt.Errorf("illegal hardlink target in tar: %s -> %s", header.Name, header.Linkname)
			}
			if err := root.MkdirAll(filepath.Dir(name), 0755); err != nil {
				return fmt.Errorf("mkdir parent %s: %v", name, err)
			}
			if err := root.Link(linkname, name); err != nil {
				return fmt.Errorf("hardlink %s: %v", name, err)
			}
		case tar.TypeXHeader, tar.TypeXGlobalHeader:
			// Metadata, not a filesystem entry.
		default:
			return fmt.Errorf("unsupported tar entry type %q: %s", header.Typeflag, header.Name)
		}
	}
}

// decompress wraps the archive in a gzip reader when it is compressed. An
// air-gapped bundle is a plain tar, since its image blobs are already
// compressed, while the legacy package is gzipped.
func decompress(f *os.File) (io.Reader, func(), error) {
	var magic [2]byte
	switch _, err := io.ReadFull(f, magic[:]); {
	case err == io.EOF, err == io.ErrUnexpectedEOF:
	case err != nil:
		return nil, nil, fmt.Errorf("read tarball: %v", err)
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return nil, nil, fmt.Errorf("read tarball: %v", err)
	}
	if magic[0] != 0x1f || magic[1] != 0x8b {
		return f, func() {}, nil
	}

	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, nil, fmt.Errorf("gzip reader: %v", err)
	}
	return gz, func() { _ = gz.Close() }, nil
}

// escapesRoot reports whether a relative path leads outside its own tree.
func escapesRoot(path string) bool {
	clean := filepath.Clean(path)
	return clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator))
}

// requireRealParents rejects a path with a symlinked parent, whose depth on disk
// would differ from its lexical depth.
func requireRealParents(root *os.Root, name string) error {
	for dir := filepath.Dir(name); dir != "."; dir = filepath.Dir(dir) {
		fi, err := root.Lstat(dir)
		if err != nil {
			return err
		}
		if fi.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("parent %s is a symlink", dir)
		}
	}
	return nil
}

// RunSetupScript runs the bundle's setup.sh to push images to the mirror
// registry. Output goes to the installer log; when progress is non-nil, the
// latest lines are also streamed to it for live UI activity.
func RunSetupScript(ctx context.Context, extractedDir, registryURL, username, password string, progress func(lines []string)) error {
	logging.Infof("[Airgap] Running %s to push images to %s...", SetupScript, registryURL)

	cmd := exec.CommandContext(ctx, "./"+SetupScript)
	cmd.Dir = extractedDir
	cmd.Env = append(os.Environ(), "REGISTRY_URL="+registryURL)

	if username != "" {
		cmd.Env = append(cmd.Env, "REGISTRY_USERNAME="+username)
	}
	if password != "" {
		cmd.Env = append(cmd.Env, "REGISTRY_PASSWORD="+password)
	}
	stdout := logging.InfoWriter()
	stderr := logging.ErrorWriter()
	if c, ok := stdout.(io.Closer); ok {
		defer c.Close()
	}
	if c, ok := stderr.(io.Closer); ok {
		defer c.Close()
	}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if progress != nil {
		tail := linetail.New(pushTailLines, progress)
		cmd.Stdout = io.MultiWriter(stdout, tail)
		cmd.Stderr = io.MultiWriter(stderr, tail)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to push airgap images: %v", err)
	}

	logging.Infof("[Airgap] Images pushed successfully to %s", registryURL)
	return nil
}
