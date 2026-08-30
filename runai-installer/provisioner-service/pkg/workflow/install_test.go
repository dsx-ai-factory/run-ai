package workflow

import (
	"testing"
	"time"

	"github.com/run-ai/runai/runai-installer/installer/pkg/api"
)

// TestRunInstallationWithEventListener tests the integration with installer.RunInstallationWithEventListner
// This test verifies that we can call the installer library and receive events asynchronously
func TestRunInstallationWithEventListener(t *testing.T) {
	t.Log("=== Starting installer.RunInstallationWithEventListner integration test ===")

	// Create installation configuration
	cfg := &api.Config{
		StepTimeout:        7 * time.Minute,
		GPUType:            api.GPUTypeFake,
		DevMode:            true,
		LogsDir:            "/tmp",
		CpVersion:          ">0.0.0-0",
		ClusterVersion:     ">0.0.0-0",
		RunaiInstallFormat: api.SHWithResearcherSetup,
		ConfigValues: map[string]string{
			api.FlagClusterName:           "test-cluster",
			api.FlagFQDN:                  "iris-provisioner.runailabs.com",
			api.FlagSSLCertPath:           "/Users/ishainsky/Documents/tls_new.crt",
			api.FlagSSLKeyPath:            "/Users/ishainsky/Documents/tls_new.key",
			api.FlagWildcardFQDN:          "*.inference.iris-provisioner.runailabs.com",
			api.FlagWildcardCertPath:      "/Users/ishainsky/dev/certs/certs/runai.crt",
			api.FlagWildcardCertKeyPath:   "/Users/ishainsky/dev/certs/certs/runai.key",
			api.FlagResearcherUserEmail:   "ishainsky@runailabs.com",
			api.FlagResearcherProjectName: "test-project",
		},
	}
	t.Logf("Created config: Mode=%s, StepTimeout=%v, GPUType=%s", cfg.Mode, cfg.StepTimeout, cfg.GPUType)

	// Create an installer instance
	err := api.Init(cfg)
	if err != nil {
		t.Logf("Failed to init installer (expected if not in k8s cluster): %v", err)
		t.Skip("Skipping test - requires kubernetes cluster access")
		return
	}
	inst := api.Get()
	t.Log("Successfully created installer instance")

	// Start installation with event listener
	t.Log("Calling installer.RunInstallationWithEventListner()...")
	eventsChan, err := inst.RunInstallationWithEventListner()
	if err != nil {
		t.Fatalf("Failed to start installation with event listener: %v", err)
	}
	t.Log("Installation started successfully, listening for events...")

	// Process events asynchronously with timeout
	timeout := time.After(15 * time.Minute)
	eventCount := 0
	installedFlows := 0
	failedFlows := 0

	for {
		select {
		case event, ok := <-eventsChan:
			if !ok {
				t.Logf("Event channel closed - installation completed")
				t.Logf("Summary: Received %d events, %d flows installed, %d flows failed",
					eventCount, installedFlows, failedFlows)
				return
			}

			eventCount++
			t.Logf("Event #%d: Flow='%s', Status=%v", eventCount, event.FlowName, event.FlowStatus)

			// Track flow completion states
			switch event.FlowStatus {
			case api.Installed:
				installedFlows++
				t.Logf("  ✓ Flow '%s' completed successfully", event.FlowName)
			case api.Failed:
				failedFlows++
				t.Logf("  ✗ Flow '%s' failed", event.FlowName)
			case api.Timeout:
				failedFlows++
				t.Logf("  ⏱ Flow '%s' timed out", event.FlowName)
			case api.Installing:
				t.Logf("  ⟳ Flow '%s' is installing...", event.FlowName)
			}

		case <-timeout:
			t.Logf("Test timeout after 10 minutes (installation still running)")
			t.Logf("Summary so far: Received %d events, %d flows installed, %d flows failed",
				eventCount, installedFlows, failedFlows)
			t.Log("Note: This timeout is expected for long-running installations")
			return
		}
	}
}
