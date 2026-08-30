package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/describe"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/describe/displayer"
	nimclient "github.com/run-ai/runai/runai-clients/nim-client/generated"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

// DescribeFixtures holds all the mock data needed for describe tests
type DescribeFixtures struct {
	WorkloadList     *openapi.GetWorkloads200Response
	WorkloadDetailed *openapi.WorkloadDetailed
	Pods             []openapi.Pod
	Events           []openapi.Event
	// Type-specific specs
	TrainingSpec       *openapi.Training
	WorkspaceSpec      *openapi.Workspace
	InferenceSpec      *openapi.Inference
	DistributedSpec    *openapi.DistributedWorkload
	DistributedInfSpec *openapi.DistributedInference
	NimSpec            *nimclient.NimServiceResponse
}

// WorkloadType represents the type of workload for fixture loading
type WorkloadType string

const (
	TrainingFixture             WorkloadType = "training"
	WorkspaceFixture            WorkloadType = "workspace"
	InferenceFixture            WorkloadType = "inference"
	DistributedFixture          WorkloadType = "distributed"
	DistributedInferenceFixture WorkloadType = "distributed_inference"
	NimFixture                  WorkloadType = "nim"
	NimAutoscaleFixture         WorkloadType = "nim_autoscale"
	NimMultiNodeFixture         WorkloadType = "nim_multinode"
)

// getFixturesBasePath returns the base path to the fixtures directory (input/ next to tests)
func getFixturesBasePath() (string, error) {
	basePath, err := getTestBasePath()
	if err != nil {
		return "", err
	}
	return basePath, nil
}

// getInputPath returns the input directory path for a workload type
func getInputPath(workloadType WorkloadType) (string, error) {
	basePath, err := getTestBasePath()
	if err != nil {
		return "", err
	}

	// Map workload type to test directory
	var testDir string
	switch workloadType {
	case TrainingFixture:
		testDir = string(TrainingFixture)
	case WorkspaceFixture:
		testDir = string(WorkspaceFixture)
	case InferenceFixture:
		testDir = string(InferenceFixture)
	case DistributedFixture:
		testDir = string(TrainingFixture)
	case DistributedInferenceFixture:
		testDir = string(DistributedInferenceFixture)
	case NimFixture:
		testDir = string(NimFixture)
	case NimAutoscaleFixture:
		testDir = string(NimAutoscaleFixture)
	case NimMultiNodeFixture:
		testDir = string(NimMultiNodeFixture)
	default:
		return "", fmt.Errorf("unknown workload type: %s", workloadType)
	}

	return filepath.Join(basePath, testDir, "input"), nil
}

// LoadDescribeFixtures loads all fixtures for a given workload type
func LoadDescribeFixtures(workloadType WorkloadType) (*DescribeFixtures, error) {
	fixturesPath, err := getInputPath(workloadType)
	if err != nil {
		return nil, fmt.Errorf("failed to get input path: %w", err)
	}

	fixtures := &DescribeFixtures{}

	// Load workload list
	workloadListData, err := loadJSONFile(filepath.Join(fixturesPath, "workload_list.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to load workload_list.json: %w", err)
	}
	fixtures.WorkloadList = &openapi.GetWorkloads200Response{}
	if err := json.Unmarshal(workloadListData, fixtures.WorkloadList); err != nil {
		return nil, fmt.Errorf("failed to unmarshal workload_list.json: %w", err)
	}

	// Load workload detailed
	workloadDetailedData, err := loadJSONFile(filepath.Join(fixturesPath, "workload_detailed.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to load workload_detailed.json: %w", err)
	}
	fixtures.WorkloadDetailed = &openapi.WorkloadDetailed{}
	if err := json.Unmarshal(workloadDetailedData, fixtures.WorkloadDetailed); err != nil {
		return nil, fmt.Errorf("failed to unmarshal workload_detailed.json: %w", err)
	}

	// Load pods
	podsData, err := loadJSONFile(filepath.Join(fixturesPath, "workload_pods.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to load workload_pods.json: %w", err)
	}
	var podsResponse struct {
		Pods []openapi.Pod `json:"pods"`
	}
	if err := json.Unmarshal(podsData, &podsResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal workload_pods.json: %w", err)
	}
	fixtures.Pods = podsResponse.Pods

	// Load events
	eventsData, err := loadJSONFile(filepath.Join(fixturesPath, "workload_events.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to load workload_events.json: %w", err)
	}
	var eventsResponse struct {
		Events []openapi.Event `json:"events"`
	}
	if err := json.Unmarshal(eventsData, &eventsResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal workload_events.json: %w", err)
	}
	fixtures.Events = eventsResponse.Events

	// Load type-specific spec
	specData, err := loadJSONFile(filepath.Join(fixturesPath, "workload_spec.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to load workload_spec.json: %w", err)
	}

	switch workloadType {
	case TrainingFixture:
		fixtures.TrainingSpec = &openapi.Training{}
		if err := json.Unmarshal(specData, fixtures.TrainingSpec); err != nil {
			return nil, fmt.Errorf("failed to unmarshal training spec: %w", err)
		}
	case WorkspaceFixture:
		fixtures.WorkspaceSpec = &openapi.Workspace{}
		if err := json.Unmarshal(specData, fixtures.WorkspaceSpec); err != nil {
			return nil, fmt.Errorf("failed to unmarshal workspace spec: %w", err)
		}
	case InferenceFixture:
		fixtures.InferenceSpec = &openapi.Inference{}
		if err := json.Unmarshal(specData, fixtures.InferenceSpec); err != nil {
			return nil, fmt.Errorf("failed to unmarshal inference spec: %w", err)
		}
	case DistributedFixture:
		fixtures.DistributedSpec = &openapi.DistributedWorkload{}
		if err := json.Unmarshal(specData, fixtures.DistributedSpec); err != nil {
			return nil, fmt.Errorf("failed to unmarshal distributed spec: %w", err)
		}
	case DistributedInferenceFixture:
		fixtures.DistributedInfSpec = &openapi.DistributedInference{}
		if err := json.Unmarshal(specData, fixtures.DistributedInfSpec); err != nil {
			return nil, fmt.Errorf("failed to unmarshal distributed inference spec: %w", err)
		}
	case NimFixture, NimAutoscaleFixture, NimMultiNodeFixture:
		nimSpec := &nimclient.NimServiceResponse{}
		if err := json.Unmarshal(specData, nimSpec); err != nil {
			return nil, fmt.Errorf("failed to unmarshal NIM spec: %w", err)
		}
		fixtures.NimSpec = nimSpec
	}

	return fixtures, nil
}

// loadJSONFile reads and returns the contents of a JSON file
func loadJSONFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// GetExpectedOutputPath returns the path to the expected output file for a workload type
func GetExpectedOutputPath(workloadType WorkloadType) (string, error) {
	basePath, err := getFixturesBasePath()
	if err != nil {
		return "", err
	}
	return filepath.Join(basePath, string(workloadType), "expected_describe_output.txt"), nil
}

// LoadExpectedOutput loads the expected describe output for a workload type
func LoadExpectedOutput(workloadType WorkloadType) (string, error) {
	path, err := GetExpectedOutputPath(workloadType)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // File doesn't exist yet, return empty
		}
		return "", err
	}
	return string(data), nil
}

// SaveExpectedOutput saves the describe output as the expected output for future tests
func SaveExpectedOutput(workloadType WorkloadType, output string) error {
	path, err := GetExpectedOutputPath(workloadType)
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(output), 0600)
}

// getTestBasePath returns the base path to the test unit directory
func getTestBasePath() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to get caller info")
	}

	// Navigate from clients/ to unit/
	// clients/ is in internal/tests/unit/cli/clients/
	// unit/ is in internal/tests/unit/
	dir := filepath.Dir(filename)
	basePath := filepath.Join(dir, "..", "..")

	return basePath, nil
}

// GetExpectedOutputPathWithFormat returns the path to the expected output file for a workload type and format
// Expected files are stored next to test files under /expected/ directory
func GetExpectedOutputPathWithFormat(workloadType WorkloadType, format string) (string, error) {
	basePath, err := getTestBasePath()
	if err != nil {
		return "", err
	}

	// Map workload type to test directory and filename
	var testDir, filename string
	switch workloadType {
	case TrainingFixture:
		testDir = string(TrainingFixture)
		filename = fmt.Sprintf("describe_%s.txt", format)
	case WorkspaceFixture:
		testDir = string(WorkspaceFixture)
		filename = fmt.Sprintf("describe_%s.txt", format)
	case InferenceFixture:
		testDir = string(InferenceFixture)
		filename = fmt.Sprintf("describe_%s.txt", format)
	case DistributedFixture:
		testDir = string(TrainingFixture)
		filename = fmt.Sprintf("distributed_describe_%s.txt", format)
	case DistributedInferenceFixture:
		testDir = string(DistributedInferenceFixture)
		filename = fmt.Sprintf("describe_%s.txt", format)
	case NimFixture:
		testDir = string(NimFixture)
		filename = fmt.Sprintf("describe_%s.txt", format)
	case NimAutoscaleFixture:
		testDir = string(NimAutoscaleFixture)
		filename = fmt.Sprintf("describe_%s.txt", format)
	case NimMultiNodeFixture:
		testDir = string(NimMultiNodeFixture)
		filename = fmt.Sprintf("describe_%s.txt", format)
	default:
		return "", fmt.Errorf("unknown workload type: %s", workloadType)
	}

	expectedPath := filepath.Join(basePath, testDir, "expected", filename)
	return expectedPath, nil
}

// CaptureDescribeTableOutput captures the table output from a describe result
func CaptureDescribeTableOutput(workload *describe.Workload) (string, error) {
	displayOptions := displayer.DisplayOptions{
		ShowGeneral:  true,
		ShowCompute:  true,
		ShowStorage:  true,
		ShowPods:     true,
		ShowEvents:   true,
		ShowNetworks: true,
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	workloadDisplayer := describe.NewWorkloadDisplayer(workload, displayOptions)
	err := workloadDisplayer.AsTable()

	w.Close()
	os.Stdout = old

	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	return buf.String(), nil
}

// CaptureDescribeJSONOutput captures the JSON output from a describe result
func CaptureDescribeJSONOutput(workload *describe.Workload) (string, error) {
	displayOptions := displayer.DisplayOptions{
		ShowGeneral:  true,
		ShowCompute:  true,
		ShowStorage:  true,
		ShowPods:     true,
		ShowEvents:   true,
		ShowNetworks: true,
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	workloadDisplayer := describe.NewWorkloadDisplayer(workload, displayOptions)
	err := workloadDisplayer.AsJSON()

	w.Close()
	os.Stdout = old

	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	return buf.String(), nil
}

// CaptureDescribeYAMLOutput captures the YAML output from a describe result
func CaptureDescribeYAMLOutput(workload *describe.Workload) (string, error) {
	displayOptions := displayer.DisplayOptions{
		ShowGeneral:  true,
		ShowCompute:  true,
		ShowStorage:  true,
		ShowPods:     true,
		ShowEvents:   true,
		ShowNetworks: true,
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	workloadDisplayer := describe.NewWorkloadDisplayer(workload, displayOptions)
	err := workloadDisplayer.AsYAML()

	w.Close()
	os.Stdout = old

	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	return buf.String(), nil
}

// CompareWithExpectedOutput compares the actual output with the expected output file
// If UPDATE_EXPECTED=1 env var is set, it updates the expected file instead of comparing
// If the expected file doesn't exist, it creates it and returns an error suggesting to review it
func CompareWithExpectedOutput(workloadType WorkloadType, format, actualOutput string) error {
	expectedPath, err := GetExpectedOutputPathWithFormat(workloadType, format)
	if err != nil {
		return fmt.Errorf("failed to get expected output path: %w", err)
	}

	// Ensure the expected directory exists
	expectedDir := filepath.Dir(expectedPath)
	if err := os.MkdirAll(expectedDir, 0755); err != nil {
		return fmt.Errorf("failed to create expected directory: %w", err)
	}

	// Check if we should update the expected output
	if os.Getenv("UPDATE_EXPECTED") == "1" {
		if err := os.WriteFile(expectedPath, []byte(actualOutput), 0600); err != nil {
			return fmt.Errorf("failed to update expected output: %w", err)
		}
		return nil
	}

	// Read expected output
	expectedBytes, err := os.ReadFile(expectedPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create the expected file with actual output
			if err := os.WriteFile(expectedPath, []byte(actualOutput), 0600); err != nil {
				return fmt.Errorf("failed to create expected output file: %w", err)
			}
			return fmt.Errorf("expected output file created at %s - please review and re-run the test", expectedPath)
		}
		return fmt.Errorf("failed to read expected output: %w", err)
	}

	expectedOutput := string(expectedBytes)

	// Normalize line endings for comparison
	actualNormalized := normalizeOutput(actualOutput)
	expectedNormalized := normalizeOutput(expectedOutput)

	if actualNormalized != expectedNormalized {
		return fmt.Errorf("output mismatch:\n\n--- EXPECTED ---\n%s\n\n--- ACTUAL ---\n%s\n\n--- DIFF ---\n%s",
			expectedOutput, actualOutput, getDiff(expectedNormalized, actualNormalized))
	}

	return nil
}

// normalizeOutput normalizes the output for comparison (trim whitespace, normalize line endings)
func normalizeOutput(s string) string {
	// Normalize line endings
	s = strings.ReplaceAll(s, "\r\n", "\n")
	// Trim trailing whitespace from each line
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	// Trim trailing empty lines
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return strings.Join(lines, "\n")
}

// getDiff returns a simple diff between expected and actual output
func getDiff(expected, actual string) string {
	expectedLines := strings.Split(expected, "\n")
	actualLines := strings.Split(actual, "\n")

	var diff strings.Builder
	maxLines := len(expectedLines)
	if len(actualLines) > maxLines {
		maxLines = len(actualLines)
	}

	for i := 0; i < maxLines; i++ {
		var expLine, actLine string
		if i < len(expectedLines) {
			expLine = expectedLines[i]
		}
		if i < len(actualLines) {
			actLine = actualLines[i]
		}

		if expLine != actLine {
			diff.WriteString(fmt.Sprintf("Line %d:\n", i+1))
			diff.WriteString(fmt.Sprintf("  - %s\n", expLine))
			diff.WriteString(fmt.Sprintf("  + %s\n", actLine))
		}
	}

	return diff.String()
}
