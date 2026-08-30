//go:build e2e

package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// CollectDiagnostics gathers cluster state into files under logsDir/postmortem-<timestamp>/.
// Intended to be called immediately on test failure while the cluster state is still live.
// All errors are printed and skipped — this must never fail a test.
func CollectDiagnostics(logsDir string) {
	timestamp := time.Now().Format("20060102-150405")
	dir := filepath.Join(logsDir, fmt.Sprintf("postmortem-%s", timestamp))
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("[diagnostics] failed to create dir %s: %v\n", dir, err)
		return
	}

	fmt.Printf("[diagnostics] collecting post-mortem data into %s\n", dir)

	// Node conditions and resource allocations — key for DiskPressure detection.
	runToFile(dir, "nodes.txt", "kubectl", "describe", "nodes")
	// All events sorted by time — captures evictions, OOM kills, disk pressure triggers.
	runToFile(dir, "events.txt", "kubectl", "get", "events", "-A", "--sort-by=.lastTimestamp")
	// Pod overview across all namespaces.
	runToFile(dir, "pods.txt", "kubectl", "get", "pods", "-A", "-o", "wide")
	// Actual disk consumption on the kind host (kind nodes are Docker containers).
	runToFile(dir, "docker-disk.txt", "docker", "system", "df", "-v")

	collectPodLogs(dir)
}

// collectPodLogs fetches logs for every non-Running/Succeeded pod cluster-wide,
// and all pods in runai* namespaces (regardless of state).
// Uses --all-containers to avoid needing to enumerate container names separately,
// which sidesteps jsonpath fragility when containerStatuses is nil (e.g. Pending pods).
func collectPodLogs(dir string) {
	logsDir := filepath.Join(dir, "pod-logs")
	// Create the directory before the kubectl listing call so it always exists in the
	// artifact even if the listing fails or all log fetches return empty — an absent
	// pod-logs/ directory is otherwise indistinguishable from "we never tried".
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		fmt.Printf("[diagnostics] failed to create pod-logs dir: %v\n", err)
		return
	}

	// Use simple column output instead of jsonpath — jsonpath breaks when
	// containerStatuses is nil (Pending pods) and is sensitive to API server errors.
	out, err := exec.Command("kubectl", "get", "pods", "-A", "--no-headers",
		"-o", "custom-columns=NS:.metadata.namespace,NAME:.metadata.name,PHASE:.status.phase").
		CombinedOutput()
	if err != nil {
		fmt.Printf("[diagnostics] failed to list pods: %v\nOutput: %s\n", err, string(out))
		return
	}

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}
		namespace, podName, phase := parts[0], parts[1], parts[2]

		isRunaiNS := strings.HasPrefix(namespace, "runai")
		isFailed := phase != "Running" && phase != "Succeeded"

		if !isRunaiNS && !isFailed {
			continue
		}

		fetchPodLog(logsDir, namespace, podName, false)
		// Previous logs are critical for crash-looping containers.
		fetchPodLog(logsDir, namespace, podName, true)
	}
}

// fetchPodLog runs kubectl logs --all-containers for a pod and writes output to a file.
// previous=true fetches the last terminated container's logs (-p flag).
func fetchPodLog(logsDir, namespace, pod string, previous bool) {
	args := []string{"logs", "-n", namespace, pod, "--all-containers", "--tail=500", "--prefix"}
	suffix := ""
	if previous {
		args = append(args, "-p")
		suffix = ".previous"
	}

	out, err := exec.Command("kubectl", args...).CombinedOutput()
	if err != nil {
		if !previous {
			// Current logs should always be available — log the failure so it's visible in CI output.
			fmt.Printf("[diagnostics] failed to fetch logs for %s/%s: %v\n", namespace, pod, err)
		}
		// For previous logs, errors are expected when no prior container exists — skip silently.
		return
	}
	if len(strings.TrimSpace(string(out))) == 0 {
		return
	}

	filename := fmt.Sprintf("%s_%s%s.log", namespace, pod, suffix)
	path := filepath.Join(logsDir, filename)
	if writeErr := os.WriteFile(path, out, 0644); writeErr != nil {
		fmt.Printf("[diagnostics] failed to write log %s: %v\n", path, writeErr)
	}
}

// runToFile executes a command and writes its combined output to filename inside dir.
func runToFile(dir, filename string, args ...string) {
	out, err := exec.Command(args[0], args[1:]...).CombinedOutput()
	path := filepath.Join(dir, filename)
	if err != nil {
		out = append(out, []byte(fmt.Sprintf("\n[error: %v]\n", err))...)
	}
	if writeErr := os.WriteFile(path, out, 0644); writeErr != nil {
		fmt.Printf("[diagnostics] failed to write %s: %v\n", path, writeErr)
	}
}
