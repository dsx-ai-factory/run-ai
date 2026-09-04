/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package filter

import (
	"fmt"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
)

type KeyName string

type Operator string

// Filters keys --> (type|name|clusterId|projectId|projectName|departmentId|departmentName|createdAt|deletedAt|submittedBy|phase|completedAt|nodepool|distributedFramework|allocatedGPU|)
// Operators --> (==|!=|<=|>=|=@|!@|=\\^|=\\$)
const (
	Name                 KeyName = "name"
	ProjectName          KeyName = "projectName"
	Type                 KeyName = "type"
	Status               KeyName = "phase"
	ClusterID            KeyName = "clusterId"
	Source               KeyName = "source"
	DistributedFramework KeyName = "distributedFramework"
	CreatedAt            KeyName = "createdAt"

	EqualOp        Operator = "=="
	NoEqualOp      Operator = "!="
	GreaterEqualOp Operator = ">="
)

type Filter struct {
	filter   KeyName
	value    string
	operator Operator
}

func NewProjectFilter(project string) *Filter {
	return &Filter{ProjectName, project, EqualOp}
}

func NewWorkloadTypeFilter(workloadType string) *Filter {
	return &Filter{Type, workloadType, EqualOp}
}

func NewWorkloadStatusFilter(workloadState string) *Filter {
	return &Filter{Status, workloadState, EqualOp}
}

func NameFilter(workloadName string) *Filter {
	return &Filter{Name, workloadName, EqualOp}
}

func NewCreatedAtFilter(createdAt string) *Filter {
	return &Filter{CreatedAt, createdAt, GreaterEqualOp}
}

func NewClusterIDFilter(clusterID string) *Filter {
	return &Filter{ClusterID, clusterID, EqualOp}
}

func NewSourceFilter(source string) *Filter {
	return &Filter{Source, source, EqualOp}
}

func NewDistributedFrameworkFilter(framework string) *Filter {
	return &Filter{DistributedFramework, framework, EqualOp}
}

func (f *Filter) Get() string {
	return fmt.Sprintf("%s%s%s", f.filter, f.operator, f.value)
}

func (f *Filter) Type() KeyName {
	return f.filter
}

func (f *Filter) Value() string {
	return f.value
}

// GetFrameworkFilters returns the filters for the given framework
// Note: until Suricata will support global framework, only distributed frameworks have filters
// so for now, runai will return all list of not distributed frameworks
func GetFrameworkFilters(framework api.Framework) []Filter {
	switch framework {
	case api.MPI, api.PyTorch, api.TF, api.XGBoost, api.Jax:
		return []Filter{*NewDistributedFrameworkFilter(distributedFrameworkAliases(framework))}
	case api.Runai:
		// Distributed workloads may store a short framework name (PyTorch) or a legacy job kind
		// (PyTorchJob), so exclude both representations from the standard training list.
		var filters []Filter
		for _, distributedFramework := range []api.Framework{api.MPI, api.PyTorch, api.Jax, api.TF, api.XGBoost} {
			jobKind := distributedFramework.JobKind()
			filters = append(filters,
				Filter{DistributedFramework, strings.TrimSuffix(jobKind, "Job"), NoEqualOp},
				Filter{DistributedFramework, jobKind, NoEqualOp},
			)
		}
		return filters

	default:
		return []Filter{}
	}
}

func distributedFrameworkAliases(framework api.Framework) string {
	jobKind := framework.JobKind()
	return strings.Join([]string{strings.TrimSuffix(jobKind, "Job"), jobKind}, "||")
}
