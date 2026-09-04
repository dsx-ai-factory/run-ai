/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cmd

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/filter"
	apilist "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/list"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/list"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/spec"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction"
	"github.com/spf13/cobra"
)

type ActionCommand interface {
	GetCommand() *cobra.Command
	RegisterFlags() error
}

type CommandMetadata interface {
	GetShort() string
	GetExample() string
	GetLong() string
	GetAliases() []string
}

type SearchableCommandTexts interface {
	CommandTexts
	GetNoResultsErrorMsg() string
	GetSearchableErrorMsg() string
	GetSuccessMsg(workloadName string) string
}

type CommandTexts interface {
	CommandMetadata
	GetBaseErrorMsg() string
}

type SearchWorkloadFn func(txn *transaction.Txn, domain, token, name, project, cluster string, fb *filter.Builder) (*list.Output, error)
type RunOperationFn func(txn *transaction.Txn, cpUrl, token, uuid string) error
type ListWorkloadsFn func(txn *transaction.Txn, domain, token string, pagingOptions *apilist.PageOptions, fb *filter.Builder, deleted bool) (*list.Output, error)
type UpdateWorkloadFn func(txn *transaction.Txn, domain, token, uuid, clusterId string, spec *spec.GeneralSpec) (string, error)

type Factory interface {
	IsSupported(framework api.Framework, workloadType api.WorkloadType) bool
	GetCommand(framework api.Framework, workloadType api.WorkloadType) (ActionCommand, error)
	GetAliasCommand(framework api.Framework, workloadType api.WorkloadType) ([]ActionCommand, error)
	IsAliasSupported(framework api.Framework, workloadType api.WorkloadType) bool
}
