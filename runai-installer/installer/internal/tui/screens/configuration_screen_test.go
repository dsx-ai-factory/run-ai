package screens

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/run-ai/runai/runai-installer/installer/internal/config"
	"github.com/run-ai/runai/runai-installer/installer/internal/inputs"
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	component "github.com/run-ai/runai/runai-installer/installer/internal/prereq/components"
	prereqids "github.com/run-ai/runai/runai-installer/installer/internal/prereq/ids"
	viewscommon "github.com/run-ai/runai/runai-installer/installer/internal/tui/screens/common"
	"github.com/run-ai/runai/runai-installer/installer/internal/tui/screens/common/mocks"
	k8smocks "github.com/run-ai/runai/runai-installer/installer/internal/utils/k8sclient/mocks"
	pkgvalidation "github.com/run-ai/runai/runai-installer/installer/pkg/validation"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func newRunAIData(dev bool) *model.RunAIInstallData {
	opts := &config.InstallConfig{DevMode: dev}
	r, _ := model.NewRunAIInstallData(opts)
	r.Prerequisites = []string{prereqids.PrerequisiteIdentifierIngress}
	// Initialize empty inputs map to ensure clean state
	r.Inputs = make(map[inputs.InstallInputType]string)

	return r
}

func newMockNextScreen(t *testing.T) *mocks.MockScreen {
	mockScreen := mocks.NewMockScreen(t)
	return mockScreen
}

func newFakeK8s(t *testing.T) *k8smocks.MockK8sClients {
	cl := fake.NewClientBuilder().Build()
	m := k8smocks.NewMockK8sClients(t)
	m.On("GetClient").Return(cl).Maybe()
	return m
}

func newFakePrerequisitesMap(k8sClient *k8smocks.MockK8sClients) map[string]component.Prerequisite {
	return map[string]component.Prerequisite{
		prereqids.PrerequisiteIdentifierIngress:          component.NewNgnixIngress(k8sClient),
		prereqids.PrerequisiteIdentifierKnative:          component.NewKnative(k8sClient),
		prereqids.PrerequisiteIdentifierGpuOperator:      component.NewGpuOperator(k8sClient),
		prereqids.PrerequisiteIdentifierPrometheus:       component.NewPrometheus(k8sClient),
		prereqids.PrerequisiteIdentifierTrainingOperator: component.NewTrainingOperator(k8sClient),
		prereqids.PrerequisiteIdentifierNetworkOperator:  component.NewNetworkOperator(k8sClient),
		prereqids.PrerequisiteIdentifierLWS:              component.NewLWS(k8sClient),
	}
}

func TestConfigurationScreen_InitAndOrder(t *testing.T) {
	runaiData := newRunAIData(false)
	k8sClient := newFakeK8s(t)
	s := NewConfigurationScreen(newMockNextScreen(t), k8sClient, newFakePrerequisitesMap(k8sClient))
	_ = s.Init(runaiData)

	var got []inputs.InstallInputType
	for _, in := range s.displayedInputs {
		got = append(got, in.GetInputIdentifier())
	}

	// Expected unique inputs for SH + Ingress
	expected := []inputs.InstallInputType{
		inputs.InstallInputClusterName,                // order 0
		inputs.InstallInputJfrogCreds,                 // order 1
		inputs.InstallInputFQDN,                       // order 2
		inputs.InstallInputIngressControllerHTTPSPort, // order 3 (from Ingress)
		inputs.InstallInputSSLPath,                    // order 4
		inputs.InstallInputSSLKeyPath,                 // order 5
		inputs.InstallInputCAChainPath,                // order 6 (optional)
		inputs.InstallInputWorkloadRouting,             // order 10 (optional)
		inputs.InstallInputHostRoutingCertPath,         // order 11 (optional when routing!=host)
		inputs.InstallInputHostRoutingKeyPath,          // order 12 (optional when routing!=host)
	}

	if len(got) != len(expected) {
		t.Fatalf("unexpected input count: got %d want %d; inputs=%v", len(got), len(expected), got)
	}
	for i := range expected {
		if got[i] != expected[i] {
			t.Fatalf("unexpected order at %d: got %v want %v; full=%v", i, got[i], expected[i], got)
		}
	}
}

func TestConfigurationScreen_CanProceed(t *testing.T) {
	// Should require all required inputs
	runaiData := newRunAIData(false)
	k8sClient := newFakeK8s(t)
	s := NewConfigurationScreen(newMockNextScreen(t), k8sClient, newFakePrerequisitesMap(k8sClient))
	_ = s.Init(runaiData)

	if s.CanProceed() {
		t.Fatalf("expected CanProceed=false when required inputs are empty")
	}

	// Fill only required inputs
	for _, in := range s.displayedInputs {
		if in.IsRequired() {
			in.SetValue("x")
		}
	}
	if !s.CanProceed() {
		t.Fatalf("expected CanProceed=true after filling required inputs")
	}

	// Test with dev mode - should still require all mandatory inputs
	runaiDataDev := newRunAIData(true)
	k8sClientDev := newFakeK8s(t)
	sDev := NewConfigurationScreen(newMockNextScreen(t), k8sClientDev, newFakePrerequisitesMap(k8sClientDev))
	_ = sDev.Init(runaiDataDev)

	// Empty required inputs should not allow proceeding, even in dev mode
	if sDev.CanProceed() {
		t.Fatalf("expected CanProceed=false when required inputs are empty, even in dev mode")
	}

	// Fill required inputs
	for _, in := range sDev.displayedInputs {
		if in.IsRequired() {
			in.SetValue("x")
		}
	}
	if !sDev.CanProceed() {
		t.Fatalf("expected CanProceed=true after filling required inputs in dev mode")
	}
}

func TestConfigurationScreen_UpdateSelection_WritesInputs(t *testing.T) {
	runaiData := newRunAIData(false)
	k8sClient := newFakeK8s(t)
	s := NewConfigurationScreen(newMockNextScreen(t), k8sClient, newFakePrerequisitesMap(k8sClient))
	_ = s.Init(runaiData)

	// Set values for a subset including optional
	for _, in := range s.displayedInputs {
		id := in.GetInputIdentifier()
		if id == inputs.InstallInputFQDN || id == inputs.InstallInputClusterName || id == inputs.InstallInputCAChainPath {
			in.SetValue("val")
		}
	}

	s.UpdateSelection(runaiData)

	if runaiData.Inputs[inputs.InstallInputFQDN] != "val" {
		t.Fatalf("expected FQDN to be saved")
	}
	if runaiData.Inputs[inputs.InstallInputClusterName] != "val" {
		t.Fatalf("expected ClusterName to be saved")
	}
	if runaiData.Inputs[inputs.InstallInputCAChainPath] != "val" {
		t.Fatalf("expected optional CAChainPath to be saved when set")
	}
	// Optional values may be present depending on component requirements; only assert explicitly set ones
}

func TestConfigurationScreen_Update_KeysAndInput(t *testing.T) {
	runaiData := newRunAIData(false)
	k8sClient := newFakeK8s(t)
	s := NewConfigurationScreen(newMockNextScreen(t), k8sClient, newFakePrerequisitesMap(k8sClient))
	_ = s.Init(runaiData)

	// Ensure the first input is empty before we start typing
	initialValue := s.displayedInputs[0].GetValue()
	if initialValue != "" {
		t.Fatalf("expected first input to start empty, got %q", initialValue)
	}

	// Type a character into the first input (ClusterName)
	cmd := s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if cmd == nil {
		// typing a rune shouldn't move selection, so no ScreenSelectionChange
		// but ensure the value updated
		if s.displayedInputs[0].GetValue() != "a" {
			t.Fatalf("expected first input value 'a', got %q", s.displayedInputs[0].GetValue())
		}
	} else {
		// Even if a cmd returned, value must be set
		if s.displayedInputs[0].GetValue() != "a" {
			t.Fatalf("expected first input value 'a', got %q", s.displayedInputs[0].GetValue())
		}
	}

	// Navigate down; expect cursor moves and selection change cmd
	cmd = s.Update(tea.KeyMsg{Type: tea.KeyDown})
	if s.GetCursor() != 1 {
		t.Fatalf("expected cursor=1 after KeyDown, got %d", s.GetCursor())
	}
	if cmd == nil {
		t.Fatalf("expected a command on selection change")
	}
	msg := cmd()
	if _, ok := msg.(viewscommon.ScreenSelectionChangeMsg); !ok {
		t.Fatalf("expected ScreenSelectionChangeMsg, got %T", msg)
	}
}

func TestConfigurationScreen_InitWithMultiplePrerequisites(t *testing.T) {
	runaiData := newRunAIData(false)
	// Add knative (adds WildcardFQDN, WildcardCertPath, and WildcardCertKeyPath inputs)
	runaiData.Prerequisites = []string{prereqids.PrerequisiteIdentifierIngress, prereqids.PrerequisiteIdentifierKnative}

	k8sClient := newFakeK8s(t)
	s := NewConfigurationScreen(newMockNextScreen(t), k8sClient, newFakePrerequisitesMap(k8sClient))
	_ = s.Init(runaiData)

	var got []inputs.InstallInputType
	for _, in := range s.displayedInputs {
		got = append(got, in.GetInputIdentifier())
	}

	expected := []inputs.InstallInputType{
		inputs.InstallInputClusterName,
		inputs.InstallInputJfrogCreds,
		inputs.InstallInputFQDN,
		inputs.InstallInputIngressControllerHTTPSPort,
		inputs.InstallInputSSLPath,
		inputs.InstallInputSSLKeyPath,
		inputs.InstallInputCAChainPath,
		inputs.InstallInputWildcardFQDN,
		inputs.InstallInputWildcardCertPath,
		inputs.InstallInputWildcardCertKeyPath,
		inputs.InstallInputWorkloadRouting,
		inputs.InstallInputHostRoutingCertPath,
		inputs.InstallInputHostRoutingKeyPath,
	}

	if len(got) != len(expected) {
		t.Fatalf("unexpected input count: got %d want %d; inputs=%v", len(got), len(expected), got)
	}
	for i := range expected {
		if got[i] != expected[i] {
			t.Fatalf("unexpected order at %d: got %v want %v; full=%v", i, got[i], expected[i], got)
		}
	}
}

func TestConfigurationScreen_AllPrerequisites_NoExtraInputs(t *testing.T) {
	runaiData := newRunAIData(false)
	// Select all prerequisites; only Ingress and knative add inputs currently
	runaiData.Prerequisites = []string{
		prereqids.PrerequisiteIdentifierIngress,
		prereqids.PrerequisiteIdentifierKnative,
		prereqids.PrerequisiteIdentifierGpuOperator,
		prereqids.PrerequisiteIdentifierPrometheus,
		prereqids.PrerequisiteIdentifierTrainingOperator,
		prereqids.PrerequisiteIdentifierNetworkOperator,
	}

	k8sClient := newFakeK8s(t)
	s := NewConfigurationScreen(newMockNextScreen(t), k8sClient, newFakePrerequisitesMap(k8sClient))
	_ = s.Init(runaiData)

	var got []inputs.InstallInputType
	for _, in := range s.displayedInputs {
		got = append(got, in.GetInputIdentifier())
	}

	expected := []inputs.InstallInputType{
		inputs.InstallInputClusterName,
		inputs.InstallInputJfrogCreds,
		inputs.InstallInputFQDN,
		inputs.InstallInputIngressControllerHTTPSPort,
		inputs.InstallInputSSLPath,
		inputs.InstallInputSSLKeyPath,
		inputs.InstallInputCAChainPath,
		inputs.InstallInputWildcardFQDN,
		inputs.InstallInputWildcardCertPath,
		inputs.InstallInputWildcardCertKeyPath,
		inputs.InstallInputWorkloadRouting,
		inputs.InstallInputHostRoutingCertPath,
		inputs.InstallInputHostRoutingKeyPath,
	}

	if len(got) != len(expected) {
		t.Fatalf("unexpected input count: got %d want %d; inputs=%v", len(got), len(expected), got)
	}
	for i := range expected {
		if got[i] != expected[i] {
			t.Fatalf("unexpected order at %d: got %v want %v; full=%v", i, got[i], expected[i], got)
		}
	}
}

func TestConfigurationScreen_HostRoutingMakesCertsMandatory(t *testing.T) {
	runaiData := newRunAIData(false)
	runaiData.Inputs[inputs.InstallInputWorkloadRouting] = pkgvalidation.WorkloadRoutingHost

	k8sClient := newFakeK8s(t)
	s := NewConfigurationScreen(newMockNextScreen(t), k8sClient, newFakePrerequisitesMap(k8sClient))
	_ = s.Init(runaiData)

	certRequired := false
	keyRequired := false
	for _, in := range s.displayedInputs {
		switch in.GetInputIdentifier() {
		case inputs.InstallInputHostRoutingCertPath:
			certRequired = in.IsRequired()
		case inputs.InstallInputHostRoutingKeyPath:
			keyRequired = in.IsRequired()
		}
	}

	if !certRequired {
		t.Error("expected host routing cert to be required when workload_routing=host")
	}
	if !keyRequired {
		t.Error("expected host routing key to be required when workload_routing=host")
	}
}

func TestConfigurationScreen_PathRoutingKeepsCertsOptional(t *testing.T) {
	runaiData := newRunAIData(false)

	k8sClient := newFakeK8s(t)
	s := NewConfigurationScreen(newMockNextScreen(t), k8sClient, newFakePrerequisitesMap(k8sClient))
	_ = s.Init(runaiData)

	for _, in := range s.displayedInputs {
		switch in.GetInputIdentifier() {
		case inputs.InstallInputHostRoutingCertPath:
			if in.IsRequired() {
				t.Error("expected host routing cert to be optional when workload_routing is not set")
			}
		case inputs.InstallInputHostRoutingKeyPath:
			if in.IsRequired() {
				t.Error("expected host routing key to be optional when workload_routing is not set")
			}
		}
	}
}

func TestConfigurationScreen_Render(t *testing.T) {
	runaiData := newRunAIData(false)
	k8sClient := newFakeK8s(t)
	s := NewConfigurationScreen(newMockNextScreen(t), k8sClient, newFakePrerequisitesMap(k8sClient))
	_ = s.Init(runaiData)

	// Test initial render (empty inputs)
	output := s.Render()
	if output == "" {
		t.Fatalf("expected non-empty render output")
	}

	// Verify that the output contains newlines (one for each input)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	expectedInputCount := len(s.displayedInputs)
	if len(lines) != expectedInputCount {
		t.Fatalf("expected %d lines in render output, got %d", expectedInputCount, len(lines))
	}

	// Test render with filled inputs
	for _, input := range s.displayedInputs {
		input.SetValue("test-value")
	}
	filledOutput := s.Render()
	if filledOutput == "" {
		t.Fatalf("expected non-empty render output with filled inputs")
	}

	// Verify that filled output is different from empty output
	if output == filledOutput {
		t.Fatalf("expected render output to change when inputs are filled")
	}

	// Verify that filled output contains the test values
	if !strings.Contains(filledOutput, "test-value") {
		t.Fatalf("expected render output to contain input values")
	}
}

func TestConfigurationScreen_Render_WithCursor(t *testing.T) {
	runaiData := newRunAIData(false)
	k8sClient := newFakeK8s(t)
	s := NewConfigurationScreen(newMockNextScreen(t), k8sClient, newFakePrerequisitesMap(k8sClient))
	_ = s.Init(runaiData)

	// Test render with cursor on first input (default)
	firstCursorOutput := s.Render()
	if firstCursorOutput == "" {
		t.Fatalf("render output should not be empty")
	}
	if s.GetCursor() != 0 {
		t.Fatalf("cursor should start at position 0, got %d", s.GetCursor())
	}

	// Move cursor to second input and verify it changes
	if len(s.displayedInputs) > 1 {
		s.cursor = 1
		secondCursorOutput := s.Render()

		if s.GetCursor() != 1 {
			t.Fatalf("cursor should be at position 1, got %d", s.GetCursor())
		}
		if secondCursorOutput == "" {
			t.Fatalf("render output should not be empty after cursor move")
		}
	}
}
