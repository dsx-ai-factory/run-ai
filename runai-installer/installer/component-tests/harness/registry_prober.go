package harness

import (
	"context"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
)

// FakeRegistryProber is a configurable orchestrator2.RegistryProber for tests.
// Results is keyed by host; hosts absent from the map use Default.
type FakeRegistryProber struct {
	Results map[string]orchestrator2.RegistryCheckResult
	Default orchestrator2.RegistryCheckResult
	Err     error
	Calls   []orchestrator2.RegistryCheckRequest
}

func (f *FakeRegistryProber) Check(_ context.Context, req orchestrator2.RegistryCheckRequest) (orchestrator2.RegistryCheckResult, error) {
	f.Calls = append(f.Calls, req)
	if f.Err != nil {
		return orchestrator2.RegistryCheckResult{}, f.Err
	}
	if r, ok := f.Results[req.Host]; ok {
		return r, nil
	}
	return f.Default, nil
}