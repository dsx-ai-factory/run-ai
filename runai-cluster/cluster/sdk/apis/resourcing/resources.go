/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// +kubebuilder:object:generate:=true
package resourcing

// QueueResources defines resource allocations for GPU, CPU, and Memory
type QueueResources struct {
	GPU    SystemResource `json:"gpu,omitempty"`
	CPU    SystemResource `json:"cpu,omitempty"`
	Memory SystemResource `json:"memory,omitempty"`
}

// SystemResource defines resource limits and quotas
type SystemResource struct {
	// +optional
	Deserved float64 `json:"deserved"`
	// +optional
	OverQuotaWeight float64 `json:"overQuotaWeight"`
	// +optional
	Limit float64 `json:"limit"`
}
