// Copyright 2019 The Kubeflow Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v2beta1

import (
	common "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type MPIJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              MPIJobSpec       `json:"spec,omitempty"`
	Status            common.JobStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type MPIJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []MPIJob `json:"items"`
}

type MPIJobSpec struct {

	// Specifies the number of slots per worker used in hostfile.
	// +optional
	// +kubebuilder:default:=1
	SlotsPerWorker *int32 `json:"slotsPerWorker,omitempty"`

	// RunPolicy encapsulates various runtime policies of the job.
	RunPolicy common.RunPolicy `json:"runPolicy,omitempty"`

	// MPIReplicaSpecs contains maps from `MPIReplicaType` to `ReplicaSpec` that
	// specify the MPI replicas to run.
	MPIReplicaSpecs map[MPIReplicaType]*common.ReplicaSpec `json:"mpiReplicaSpecs"`

	// SSHAuthMountPath is the directory where SSH keys are mounted.
	// +kubebuilder:default:="/root/.ssh"
	SSHAuthMountPath string `json:"sshAuthMountPath,omitempty"`

	// launcherCreationPolicy if WaitForWorkersReady, the launcher is created only after all workers are in Ready state. Defaults to AtStartup.
	// +kubebuilder:validation:Enum:AtStartup;WaitForWorkersReady
	// +kubebuilder:default:=AtStartup
	LauncherCreationPolicy LauncherCreationPolicy `json:"launcherCreationPolicy,omitempty"`

	// MPIImplementation is the MPI implementation.
	// Options are "OpenMPI" (default) and "Intel".
	// +kubebuilder:validation:Enum:=OpenMPI;Intel
	// +kubebuilder:default:=OpenMPI
	MPIImplementation MPIImplementation `json:"mpiImplementation,omitempty"`
}

// MPIReplicaType is the type for MPIReplica.
type MPIReplicaType common.ReplicaType

const (
	// MPIReplicaTypeLauncher is the type for launcher replica.
	MPIReplicaTypeLauncher MPIReplicaType = "Launcher"

	// MPIReplicaTypeWorker is the type for worker replicas.
	MPIReplicaTypeWorker MPIReplicaType = "Worker"
)

type LauncherCreationPolicy string

const (
	// LauncherCreationPolicyAtStartup is default behavior
	// when Launcher is started in parallel with workers
	LauncherCreationPolicyAtStartup LauncherCreationPolicy = "AtStartup"

	// LauncherCreationPolicyWaitForWorkersReady makes Launcher reation
	// postponed until all workers are in ready state so that the Launcher
	// does not fail trying to connect to worker.
	LauncherCreationPolicyWaitForWorkersReady LauncherCreationPolicy = "WaitForWorkersReady"
)

type MPIImplementation string

const (
	MPIImplementationOpenMPI MPIImplementation = "OpenMPI"
	MPIImplementationIntel   MPIImplementation = "Intel"
)
