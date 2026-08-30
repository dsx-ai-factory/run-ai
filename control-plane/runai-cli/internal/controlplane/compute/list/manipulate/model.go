package manipulate

import (
	openapi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
)

type BriefCompute struct {
	Name       string `json:"name" yaml:"name"`
	Scope      string `json:"scope" yaml:"scope"`
	GPUDevices string `json:"gpuDevices,omitempty" yaml:"gpuDevices,omitempty"`
	GPUMemory  string `json:"gpuMemory,omitempty" yaml:"gpuMemory,omitempty"`
	CPUMemory  string `json:"cpuMemory,omitempty" yaml:"cpuMemory,omitempty"`
}

type ComputesAlignedRow struct {
	Name       string                   `json:"Name,omitempty" yaml:"Name,omitempty"`
	Scope      string                   `json:"Scope,omitempty" yaml:"Scope,omitempty"`
	GPUDevices string                   `json:"GPUDevices,omitempty" yaml:"GPUDevices,omitempty"`
	GPUMemory  string                   `json:"GPUMemory,omitempty" yaml:"GPUMemory,omitempty"`
	CPUMemory  string                   `json:"CPUMemory,omitempty" yaml:"CPUMemory,omitempty"`
	Spec       openapi.ComputeAssetSpec `json:"-" yaml:"-"`
}

type ComputesAlignedData struct {
	Computes []ComputesAlignedRow `json:"computes" yaml:"computes"`
}

type Data struct {
	Computes []openapi.ComputeAsset `json:"computes" yaml:"computes"`
}

type BriefData struct {
	Computes []BriefCompute `json:"computes" yaml:"computes"`
}
