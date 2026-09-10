/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package installer

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const (
	testBashRCMarker = "bash rc marker\n"
	testZshRCMarker  = "zsh rc marker\n"
)

func TestInstallerConfiguresOnlyDetectedShell(t *testing.T) {
	testCases := []struct {
		name              string
		shell             string
		configuredRC      string
		unconfiguredRC    string
		unconfiguredValue string
		completionShell   string
	}{
		{
			name:              "bash",
			shell:             "/bin/bash",
			configuredRC:      ".bashrc",
			unconfiguredRC:    ".zshrc",
			unconfiguredValue: testZshRCMarker,
			completionShell:   "bash",
		},
		{
			name:              "zsh",
			shell:             "/bin/zsh",
			configuredRC:      ".zshrc",
			unconfiguredRC:    ".bashrc",
			unconfiguredValue: testBashRCMarker,
			completionShell:   "zsh",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			homeDir := t.TempDir()
			writeTestFile(t, filepath.Join(homeDir, ".bashrc"), testBashRCMarker, 0o644)
			writeTestFile(t, filepath.Join(homeDir, ".zshrc"), testZshRCMarker, 0o644)

			output := runTestInstaller(t, homeDir, &testCase.shell, "\n\n\nn\n")

			configuredContent := readTestFile(t, filepath.Join(homeDir, testCase.configuredRC))
			installPath := filepath.Join(homeDir, ".runai", "bin")
			softLinkPath := filepath.Join(installPath, "runai")
			assertContainsOnce(t, configuredContent, "export PATH=$PATH:"+installPath)
			assertContainsOnce(t, configuredContent, "source <("+softLinkPath+" completion "+testCase.completionShell+")")

			unconfiguredContent := readTestFile(t, filepath.Join(homeDir, testCase.unconfiguredRC))
			if unconfiguredContent != testCase.unconfiguredValue {
				t.Fatalf("unexpected changes to %s:\n%s", testCase.unconfiguredRC, unconfiguredContent)
			}

			expectedSummary := "source " + filepath.Join(homeDir, testCase.configuredRC)
			if !strings.Contains(output, expectedSummary) {
				t.Fatalf("expected output to contain %q, got:\n%s", expectedSummary, output)
			}
		})
	}
}

func TestInstallerDoesNotModifyStartupFilesForUnsupportedShell(t *testing.T) {
	homeDir := t.TempDir()
	bashRCPath := filepath.Join(homeDir, ".bashrc")
	zshRCPath := filepath.Join(homeDir, ".zshrc")
	writeTestFile(t, bashRCPath, testBashRCMarker, 0o644)
	writeTestFile(t, zshRCPath, testZshRCMarker, 0o644)
	tcsh := "/bin/tcsh"

	output := runTestInstaller(t, homeDir, &tcsh, "n\n")

	if content := readTestFile(t, bashRCPath); content != testBashRCMarker {
		t.Fatalf("unexpected changes to .bashrc:\n%s", content)
	}
	if content := readTestFile(t, zshRCPath); content != testZshRCMarker {
		t.Fatalf("unexpected changes to .zshrc:\n%s", content)
	}
	if !strings.Contains(output, "Does not support adding the CLI to the PATH variable for this shell; supported shells are bash and zsh") {
		t.Fatalf("expected unsupported-shell warning, got:\n%s", output)
	}
	if !strings.Contains(output, "Add '"+filepath.Join(homeDir, ".runai", "bin")+"' to PATH") {
		t.Fatalf("expected manual PATH instructions, got:\n%s", output)
	}
	if strings.Contains(output, "Configure PATH variable?") || strings.Contains(output, "Configure auto-completion?") {
		t.Fatalf("unexpected shell-configuration prompt for unsupported shell:\n%s", output)
	}
	if strings.Contains(output, "To apply shell changes") {
		t.Fatalf("unexpected shell source instruction for unsupported shell:\n%s", output)
	}
}

func TestInstallerHandlesEmptyShell(t *testing.T) {
	homeDir := t.TempDir()
	emptyShell := ""

	output := runTestInstaller(t, homeDir, &emptyShell, "n\n")

	if !strings.Contains(output, "Does not support adding the CLI to the PATH variable for this shell; supported shells are bash and zsh") {
		t.Fatalf("expected unsupported-shell warning, got:\n%s", output)
	}
	if _, err := os.Stat(filepath.Join(homeDir, ".bashrc")); !os.IsNotExist(err) {
		t.Fatalf("expected .bashrc not to be created, got error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(homeDir, ".zshrc")); !os.IsNotExist(err) {
		t.Fatalf("expected .zshrc not to be created, got error: %v", err)
	}
}

func TestInstallerDoesNotCreateMissingStartupFile(t *testing.T) {
	homeDir := t.TempDir()
	bash := "/bin/bash"

	output := runTestInstaller(t, homeDir, &bash, "n\n")

	if _, err := os.Stat(filepath.Join(homeDir, ".bashrc")); !os.IsNotExist(err) {
		t.Fatalf("expected .bashrc not to be created, got error: %v", err)
	}
	if strings.Contains(output, "Configure PATH variable?") || strings.Contains(output, "Configure auto-completion?") {
		t.Fatalf("unexpected shell-configuration prompt without a startup file:\n%s", output)
	}
	if !strings.Contains(output, "Add '"+filepath.Join(homeDir, ".runai", "bin")+"' to PATH") {
		t.Fatalf("expected manual PATH instructions, got:\n%s", output)
	}
}

func TestInstallerConfigurationIsIdempotent(t *testing.T) {
	homeDir := t.TempDir()
	bashRCPath := filepath.Join(homeDir, ".bashrc")
	writeTestFile(t, bashRCPath, testBashRCMarker, 0o644)
	bash := "/bin/bash"

	runTestInstaller(t, homeDir, &bash, "\n\nn\n")
	runTestInstaller(t, homeDir, &bash, "\n\nn\n")

	content := readTestFile(t, bashRCPath)
	installPath := filepath.Join(homeDir, ".runai", "bin")
	softLinkPath := filepath.Join(installPath, "runai")
	assertContainsOnce(t, content, "export PATH=$PATH:"+installPath)
	assertContainsOnce(t, content, "source <("+softLinkPath+" completion bash)")
}

func TestInstallerDoesNotReportFailedStartupFileWritesAsConfigured(t *testing.T) {
	homeDir := t.TempDir()
	bashRCPath := filepath.Join(homeDir, ".bashrc")
	writeTestFile(t, bashRCPath, testBashRCMarker, 0o444)

	probe, err := os.OpenFile(bashRCPath, os.O_APPEND|os.O_WRONLY, 0)
	if err == nil {
		closeErr := probe.Close()
		if closeErr != nil {
			t.Fatalf("close write probe: %v", closeErr)
		}
		t.Skip("test process can write a read-only startup file")
	}

	bash := "/bin/bash"
	output := runTestInstaller(t, homeDir, &bash, "\n\nn\n")

	if content := readTestFile(t, bashRCPath); content != testBashRCMarker {
		t.Fatalf("unexpected changes to .bashrc:\n%s", content)
	}
	if strings.Contains(output, "To apply shell changes") {
		t.Fatalf("unexpected source instruction after failed writes:\n%s", output)
	}
	if !strings.Contains(output, "Add '"+filepath.Join(homeDir, ".runai", "bin")+"' to PATH") {
		t.Fatalf("expected manual PATH instructions after failed write, got:\n%s", output)
	}
	if !strings.Contains(output, "Auto-completion was not configured.") {
		t.Fatalf("expected auto-completion failure summary, got:\n%s", output)
	}
}

func TestInstallerReportsWhenShellConfigurationIsDeclined(t *testing.T) {
	homeDir := t.TempDir()
	bashRCPath := filepath.Join(homeDir, ".bashrc")
	writeTestFile(t, bashRCPath, testBashRCMarker, 0o644)
	bash := "/bin/bash"

	output := runTestInstaller(t, homeDir, &bash, "n\nn\nn\n")

	if !strings.Contains(output, "Shell configuration was skipped.") {
		t.Fatalf("expected skipped-configuration summary, got:\n%s", output)
	}
	if content := readTestFile(t, bashRCPath); content != testBashRCMarker {
		t.Fatalf("unexpected changes to .bashrc:\n%s", content)
	}
}

func runTestInstaller(t *testing.T, homeDir string, shell *string, input string) string {
	t.Helper()

	oldInstallerBasePath := installerBasePath
	installerBasePath = filepath.Join("..", "..", "installer")
	t.Cleanup(func() {
		installerBasePath = oldInstallerBasePath
	})

	installerCode, err := GetInstallerCode(Settings{
		CpURL:      "https://example.test",
		AuthURL:    "https://example.test/auth",
		Token:      "test-token",
		SdkVersion: "2.26.0-test",
		CpVersion:  "2.26.0-test",
	})
	if err != nil {
		t.Fatalf("render installer: %v", err)
	}

	testBinDir := t.TempDir()
	fakeCurlPath := filepath.Join(testBinDir, "curl")
	writeTestFile(t, fakeCurlPath, `#!/bin/bash
output=""
while [ "$#" -gt 0 ]; do
  if [ "$1" = "-o" ]; then
    output="$2"
    shift 2
    continue
  fi
  shift
done
printf '#!/bin/bash\nexit 0\n' > "$output"
chmod 755 "$output"
`, 0o755)

	installerPath := filepath.Join(t.TempDir(), "installer.sh")
	writeTestFile(t, installerPath, string(installerCode), 0o755)

	command := exec.Command("bash", installerPath, "--skip-old-version-validation")
	command.Stdin = strings.NewReader(input)
	command.Env = testEnvironment(homeDir, testBinDir, shell)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("run installer: %v\n%s", err, output)
	}
	return string(output)
}

func testEnvironment(homeDir, testBinDir string, shell *string) []string {
	environment := make([]string, 0, len(os.Environ())+4)
	for _, variable := range os.Environ() {
		if strings.HasPrefix(variable, "HOME=") ||
			strings.HasPrefix(variable, "PATH=") ||
			strings.HasPrefix(variable, "SHELL=") ||
			strings.HasPrefix(variable, "RUNAI_CLI_CONFIG_PATH=") {
			continue
		}
		environment = append(environment, variable)
	}
	environment = append(environment,
		"HOME="+homeDir,
		"PATH="+testBinDir+string(os.PathListSeparator)+os.Getenv("PATH"),
		"RUNAI_CLI_CONFIG_PATH="+filepath.Join(homeDir, ".runai"),
	)
	if shell != nil {
		environment = append(environment, "SHELL="+*shell)
	}
	return environment
}

func writeTestFile(t *testing.T, path, content string, mode os.FileMode) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func readTestFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(content)
}

func assertContainsOnce(t *testing.T, content, value string) {
	t.Helper()
	if count := strings.Count(content, value); count != 1 {
		t.Fatalf("expected %q exactly once, got %d occurrences in:\n%s", value, count, content)
	}
}
