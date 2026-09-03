/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package field_groups

import flds "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"

// +k8s:deepcopy-gen=true
type InteractiveFields struct {

	// Indication if an interactive workload should run jupyter notebook
	IsJupyter *flds.BoolField `json:"jupyter,omitempty"`

	// Indicates that this interactive workload should also run a TensorBoard dashboard
	IsTensorboard *flds.BoolField `json:"tensorboard,omitempty"`

	// A token for connecting to a Jupyter Notebook created for workloads
	// of type Jupyter. When token authentication is enabled, the notebook uses this token to authenticate requests.
	// For more information see:  https://jupyter-notebook.readthedocs.io/en/stable/security.html
	NotebookToken *flds.StringField `json:"notebookToken,omitempty"`

	// The TensorBoard Logs directory
	TensorboardLogdir *flds.StringField `json:"tensorboardLogdir,omitempty"`

	// Deprecated: Use Preemptibility field instead.
	// Specifies that the created workload will be preemptible. Interactive preemptible workloads can be
	// scheduled above the guaranteed quota but may be reclaimed at any time.
	Preemptible *flds.BoolField `json:"preemptible,omitempty"`
}
