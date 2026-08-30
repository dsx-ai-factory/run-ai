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
package v1alpha1

// QueueResourcesConfig defines resource allocations for GPU, CPU, and Memory
type QueueResourcesConfig struct {
	// GPU resources in fractions (e.g. 0.7 = 70% of a GPU)
	GPU SystemResource `json:"gpu,omitempty"`

	// CPU resources in millicpus (e.g. 1000 = 1 CPU)
	CPU SystemResource `json:"cpu,omitempty"`

	// Memory resources in megabytes (1 MB = 10^6 bytes)
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

type QueueConfig struct {
	// Queue name
	Name string `json:"name"`

	// Nodepool associated with the queue
	Nodepool string `json:"nodepool"`

	// Queue resources definition
	Resources *QueueResourcesConfig `json:"resources,omitempty"`

	// Queue priority
	Priority *int32 `json:"priority,omitempty"`
}
