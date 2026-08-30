package airgap

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunSetupScript_StreamsProgress(t *testing.T) {
	dir := t.TempDir()
	script := "#!/bin/sh\nfor i in 1 2 3 4 5 6 7; do echo line$i; done\nexit 0\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, SetupScript), []byte(script), 0o755))

	var mu sync.Mutex
	var last []string
	report := func(l []string) {
		mu.Lock()
		last = append([]string(nil), l...)
		mu.Unlock()
	}

	require.NoError(t, RunSetupScript(context.Background(), dir, "registry.local", "", "", report))

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, []string{"line3", "line4", "line5", "line6", "line7"}, last,
		"progress surfaces the last pushTailLines lines of setup.sh output")
}

// tarEntry is one member of a tarball; Size is derived from body.
type tarEntry struct {
	hdr  tar.Header
	body string
}

// writeTarGzEntries builds a gzipped tarball with the given entries, in order.
func writeTarGzEntries(t *testing.T, entries []tarEntry) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "bundle.tar.gz")
	f, err := os.Create(path)
	require.NoError(t, err)
	defer f.Close()

	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)
	for _, e := range entries {
		hdr := e.hdr
		hdr.Size = int64(len(e.body))
		require.NoError(t, tw.WriteHeader(&hdr))
		_, err := tw.Write([]byte(e.body))
		require.NoError(t, err)
	}
	require.NoError(t, tw.Close())
	require.NoError(t, gw.Close())
	return path
}

// writeTarGz builds a gzipped tarball of regular files (name -> contents).
func writeTarGz(t *testing.T, files map[string]string) string {
	t.Helper()
	entries := make([]tarEntry, 0, len(files))
	for name, content := range files {
		entries = append(entries, tarEntry{
			hdr:  tar.Header{Typeflag: tar.TypeReg, Name: name, Mode: 0644},
			body: content,
		})
	}
	return writeTarGzEntries(t, entries)
}

func TestExtractTar_ExtractsFiles(t *testing.T) {
	tarball := writeTarGz(t, map[string]string{
		"custom-env.yaml":          "registry: test\n",
		"charts/control-plane.tgz": "chart",
	})
	dest := t.TempDir()

	require.NoError(t, ExtractTar(context.Background(), tarball, dest))

	got, err := os.ReadFile(filepath.Join(dest, "custom-env.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "registry: test\n", string(got))
	_, err = os.Stat(filepath.Join(dest, "charts", "control-plane.tgz"))
	require.NoError(t, err)
}

// An air-gapped bundle is a plain tar: its image blobs are already compressed.
func TestExtractTar_ExtractsUncompressedArchive(t *testing.T) {
	tarball := filepath.Join(t.TempDir(), "bundle.tar")
	f, err := os.Create(tarball)
	require.NoError(t, err)
	tw := tar.NewWriter(f)
	require.NoError(t, tw.WriteHeader(&tar.Header{Typeflag: tar.TypeReg, Name: "artifact-lock.yaml", Mode: 0644, Size: 5}))
	_, err = tw.Write([]byte("lock\n"))
	require.NoError(t, err)
	require.NoError(t, tw.Close())
	require.NoError(t, f.Close())

	dest := t.TempDir()
	require.NoError(t, ExtractTar(context.Background(), tarball, dest))

	got, err := os.ReadFile(filepath.Join(dest, "artifact-lock.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "lock\n", string(got))
}

func TestExtractTar_ContextCancelled(t *testing.T) {
	tarball := writeTarGz(t, map[string]string{"custom-env.yaml": "registry: test\n"})
	dest := t.TempDir()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := ExtractTar(ctx, tarball, dest)
	require.ErrorIs(t, err, context.Canceled)
}

func TestExtractTar_MissingTarball(t *testing.T) {
	err := ExtractTar(context.Background(), filepath.Join(t.TempDir(), "nope.tar.gz"), t.TempDir())
	require.Error(t, err)
}

// nestedDest returns an empty extraction directory and its parent.
func nestedDest(t *testing.T) (dest, outside string) {
	t.Helper()
	outside = t.TempDir()
	dest = filepath.Join(outside, "extract")
	require.NoError(t, os.Mkdir(dest, 0o755))
	return dest, outside
}

func TestExtractTar_RejectsPathTraversal(t *testing.T) {
	tarball := writeTarGzEntries(t, []tarEntry{
		{hdr: tar.Header{Typeflag: tar.TypeReg, Name: "../escape.txt", Mode: 0644}, body: "owned"},
	})
	dest, outside := nestedDest(t)

	assert.Error(t, ExtractTar(context.Background(), tarball, dest))
	assert.NoFileExists(t, filepath.Join(outside, "escape.txt"))
}

func TestExtractTar_RejectsSymlinkEscape(t *testing.T) {
	for name, linkname := range map[string]string{
		"absolute": "/etc/passwd",
		"relative": "../../outside",
	} {
		t.Run(name, func(t *testing.T) {
			tarball := writeTarGzEntries(t, []tarEntry{
				{hdr: tar.Header{Typeflag: tar.TypeSymlink, Name: "link", Linkname: linkname, Mode: 0777}},
			})
			dest, _ := nestedDest(t)

			require.Error(t, ExtractTar(context.Background(), tarball, dest))
		})
	}
}

// A target that is lexically inside can still lead outside once an earlier
// symlink is followed.
func TestExtractTar_RejectsSymlinkChainEscape(t *testing.T) {
	tarball := writeTarGzEntries(t, []tarEntry{
		{hdr: tar.Header{Typeflag: tar.TypeSymlink, Name: "a", Linkname: ".", Mode: 0777}},
		{hdr: tar.Header{Typeflag: tar.TypeSymlink, Name: "a/b", Linkname: "..", Mode: 0777}},
		{hdr: tar.Header{Typeflag: tar.TypeReg, Name: "a/b/pwned", Mode: 0644}, body: "owned"},
	})
	dest, outside := nestedDest(t)

	assert.Error(t, ExtractTar(context.Background(), tarball, dest))
	assert.NoFileExists(t, filepath.Join(outside, "pwned"))
}

// Without a write through the chain, the link itself is the escape.
func TestExtractTar_RejectsSymlinkPlantedViaSymlinkedParent(t *testing.T) {
	tarball := writeTarGzEntries(t, []tarEntry{
		{hdr: tar.Header{Typeflag: tar.TypeSymlink, Name: "a", Linkname: ".", Mode: 0777}},
		{hdr: tar.Header{Typeflag: tar.TypeSymlink, Name: "a/b", Linkname: "..", Mode: 0777}},
	})
	dest, _ := nestedDest(t)

	assert.Error(t, ExtractTar(context.Background(), tarball, dest))

	// Lstat, not Stat: the residual entry would be a symlink Stat would resolve.
	_, err := os.Lstat(filepath.Join(dest, "b"))
	assert.ErrorIs(t, err, os.ErrNotExist, "no symlink may be left pointing above the destination")
}

func TestExtractTar_RestrictiveDirModeStillExtracts(t *testing.T) {
	tarball := writeTarGzEntries(t, []tarEntry{
		{hdr: tar.Header{Typeflag: tar.TypeDir, Name: "locked", Mode: 0500}},
		{hdr: tar.Header{Typeflag: tar.TypeReg, Name: "locked/chart.tgz", Mode: 0644}, body: "chart"},
	})
	dest := t.TempDir()

	require.NoError(t, ExtractTar(context.Background(), tarball, dest))

	got, err := os.ReadFile(filepath.Join(dest, "locked", "chart.tgz"))
	require.NoError(t, err)
	assert.Equal(t, "chart", string(got))
}

func TestExtractTar_ExtractsHardlink(t *testing.T) {
	tarball := writeTarGzEntries(t, []tarEntry{
		{hdr: tar.Header{Typeflag: tar.TypeReg, Name: "charts/real.tgz", Mode: 0644}, body: "chart"},
		{hdr: tar.Header{Typeflag: tar.TypeLink, Name: "copy.tgz", Linkname: "charts/real.tgz", Mode: 0644}},
	})
	dest := t.TempDir()

	require.NoError(t, ExtractTar(context.Background(), tarball, dest))

	got, err := os.ReadFile(filepath.Join(dest, "copy.tgz"))
	require.NoError(t, err)
	assert.Equal(t, "chart", string(got))
}

func TestExtractTar_RejectsHardlinkEscape(t *testing.T) {
	tarball := writeTarGzEntries(t, []tarEntry{
		{hdr: tar.Header{Typeflag: tar.TypeLink, Name: "copy", Linkname: "../../outside", Mode: 0644}},
	})
	dest, _ := nestedDest(t)

	assert.Error(t, ExtractTar(context.Background(), tarball, dest))
}

// Go surfaces global PAX headers to the caller.
func TestExtractTar_SkipsArchiveMetadata(t *testing.T) {
	tarball := writeTarGzEntries(t, []tarEntry{
		{hdr: tar.Header{Typeflag: tar.TypeXGlobalHeader, Name: "pax_global_header",
			PAXRecords: map[string]string{"comment": "abc"}, Format: tar.FormatPAX}},
		{hdr: tar.Header{Typeflag: tar.TypeReg, Name: "custom-env.yaml", Mode: 0644}, body: "registry: test\n"},
	})
	dest := t.TempDir()

	require.NoError(t, ExtractTar(context.Background(), tarball, dest))

	got, err := os.ReadFile(filepath.Join(dest, "custom-env.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "registry: test\n", string(got))
}

func TestExtractTar_RejectsUnsupportedEntryType(t *testing.T) {
	tarball := writeTarGzEntries(t, []tarEntry{
		{hdr: tar.Header{Typeflag: tar.TypeFifo, Name: "pipe", Mode: 0644}},
	})

	err := ExtractTar(context.Background(), tarball, t.TempDir())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported tar entry type")
}

func TestExtractTar_ExtractsSymlink(t *testing.T) {
	tarball := writeTarGzEntries(t, []tarEntry{
		{hdr: tar.Header{Typeflag: tar.TypeReg, Name: "charts/real.tgz", Mode: 0644}, body: "chart"},
		{hdr: tar.Header{Typeflag: tar.TypeSymlink, Name: "latest.tgz", Linkname: "charts/real.tgz", Mode: 0777}},
	})
	dest := t.TempDir()

	require.NoError(t, ExtractTar(context.Background(), tarball, dest))

	got, err := os.ReadFile(filepath.Join(dest, "latest.tgz"))
	require.NoError(t, err)
	assert.Equal(t, "chart", string(got))
}
