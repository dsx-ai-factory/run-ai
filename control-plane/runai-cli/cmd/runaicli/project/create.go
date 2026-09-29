/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package project

import (
	"context"
	"fmt"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmdtxn"
	cpproject "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/project"

	departmentservice "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/department"

	openapi2 "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/nodepool"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cli"
	"k8s.io/utils/ptr"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper/validators"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

var description = "Resources for the project (can be specified multiple times). Format: nodepool=<name>,gpu-deserved=<num>,gpu-limit=<num>,priority=<level>"

var usageExamples = `# Basic Project Creation
runai project create research-team-a --resources gpu-deserved=2

# Project with Single Nodepool
runai project create ml-team \
  --resources nodepool=gpu-pool,gpu-deserved=2,gpu-limit=4,priority=High \
  --description "Machine Learning Team Project"

# Project with Multiple Nodepools
runai project create data-science \
  --resources nodepool=gpu-pool-a,gpu-deserved=2,gpu-limit=4,priority=High \
  --resources nodepool=gpu-pool-b,gpu-deserved=1,gpu-limit=2,priority=Medium \
  --description "Data Science Team Project"

# Project with Custom Namespace
runai project create research \
  --resources nodepool=gpu-pool,gpu-deserved=1,gpu-limit=2,priority=Medium \
  --namespace research-ns \
  --description "Research Team Project"

# Project with Over-Quota
runai project create over-quota \
  --resources gpu-deserved=5,gpu-limit=10,priority=Medium

# Project with Unlimited Over-Quota
runai project create unlimited-over-quota \
  --resources gpu-deserved=0,gpu-limit=-1`

type Config struct {
	clusterId string
	domain    string
	token     string
}

func NewConfig() *Config {
	c := &Config{}
	instance = config.GetInstance()
	c.token = instance.GetToken()
	c.domain = instance.GetConfigValue(config.ControlPlaneUrl)
	c.clusterId = instance.GetConfigValue(config.ClusterUUID)
	return c
}

type createOptions struct {
	name           string
	namespace      string
	departmentName string
	departmentId   string
	resources      []openapi.ResourcesNullable
	description    string
}

type creationHandler struct {
	options *createOptions
	config  *Config
}

const (
	FailedToCreateProject = "Project creation failed"
	DefaultDepartmentName = "default"
)

var (
	handler *creationHandler
)

func Create() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create PROJECT_NAME",
		Short:   "Create a new project",
		Example: usageExamples,
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			handler.options.name = args[0]
			handler.initConfig()
			msg, err := handler.validateProject()
			if err != nil {
				return cmdtxn.CmdErrorValidationMessage(cmd, msg, err)
			}

			if err = handler.validateNodepoolNamesUniqueness(); err != nil {
				return handler.CmdErrorPrintAndSet(cmd, err)
			}

			err = handler.populateNodepoolIds()
			if err != nil {
				return handler.CmdErrorPrintAndSet(cmd, err)
			}

			err = handler.populateDepartmentId()
			if err != nil {
				return handler.CmdErrorPrintAndSet(cmd, err)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {

			prj, err := handler.createProject()
			if err != nil {
				return handler.CmdErrorPrintAndSet(cmd, err)
			}
			cmd.Println(fmt.Sprintf("Creating project %s ...\n", prj.Name))
			cmd.Println("To track the project's status, run 'runai project list'")
			return nil
		},
		PostRunE: func(cmd *cobra.Command, args []string) error {
			handler.clean()
			return nil
		},
	}

	return cmd
}
func (h *creationHandler) CmdErrorPrintAndSet(cmd *cobra.Command, err error) error {
	h.clean()
	return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToCreateProject, err)
}

func (h *creationHandler) validateProject() (string, error) {
	validator := validators.NewProjectValidator(h.options.name)
	return validator.Execute()
}

func (h *creationHandler) validateNodepoolNamesUniqueness() error {
	seenNames := make(map[string]int)
	var duplicateNames []string

	// count occurrences of each name
	for _, resource := range h.options.resources {
		if !resource.HasNodePool() {
			continue
		}
		nodePool := resource.GetNodePool()
		if !nodePool.HasName() {
			continue
		}
		name := nodePool.GetName()
		seenNames[name]++
		if seenNames[name] == 2 {
			duplicateNames = append(duplicateNames, name)
		}
	}

	// If no duplicates found, return early
	if len(duplicateNames) == 0 {
		return nil
	}

	return fmt.Errorf("duplicate nodepool names found: %s", strings.Join(duplicateNames, ", "))
}

func emptyCreateOptions() *createOptions {
	return &createOptions{
		departmentName: DefaultDepartmentName,
	}
}

func newCreationHandler() *creationHandler {
	handler = &creationHandler{
		options: emptyCreateOptions(),
		config:  nil,
	}
	return handler
}

func (h *creationHandler) initConfig() {
	h.config = NewConfig()
}

func (h *creationHandler) clean() {
	h.options.name = ""
	h.options.namespace = ""
	h.options.departmentName = DefaultDepartmentName
	h.options.departmentId = ""
	h.options.resources = []openapi.ResourcesNullable{}
	h.options.description = ""
}

func (h *creationHandler) populateDepartmentId() error {
	c := h.config
	departmentId, err := departmentservice.GetDepartmentId(txn, c.domain, c.token, c.clusterId, h.options.departmentName)
	if err != nil {
		return err
	}
	h.options.departmentId = departmentId
	return nil
}

func (h *creationHandler) populateNodepoolIds() error {
	nodepoolsMetadata, err := h.getNodepoolsMetadata(txn.GetCtx())
	if err != nil {
		return err
	}

	for _, resource := range h.options.resources {
		nodepoolName := *resource.NodePool.Get().Name
		np := findNodepoolByName(nodepoolsMetadata, nodepoolName)
		if np == nil {
			return fmt.Errorf("node pool %s not found", nodepoolName)
		}
		resource.NodePool.Get().SetId(np.Id)
	}
	return nil
}

func (h *creationHandler) getNodepoolsMetadata(ctx context.Context) ([]openapi2.NodepoolMinimal, error) {
	c := h.config

	nodepoolService, err := nodepool.GetNodepoolService(c.domain, c.token)
	if err != nil {
		return nil, err
	}

	response, err := nodepoolService.GetNodepoolMinimalList(ctx, c.clusterId)
	if err != nil {
		return nil, err
	}

	return response.Nodepools, nil
}

func findNodepoolByName(nodepoolsMetadata []openapi2.NodepoolMinimal, nodepoolName string) *openapi2.NodepoolMinimal {
	for _, np := range nodepoolsMetadata {
		if np.Name == nodepoolName {
			return &np
		}
	}
	return nil
}

func (h *creationHandler) createProject() (*openapi.Project, error) {
	creationRequest := h.createProjectCreationRequest()
	c := h.config
	return cpproject.CreateProject(txn, c.domain, c.token, creationRequest)
}

func (h *creationHandler) createProjectCreationRequest() *openapi.ProjectCreationRequest {
	options := h.options
	request := &openapi.ProjectCreationRequest{
		Name:      options.name,
		ParentId:  *openapi.NewNullableString(ptr.To(options.departmentId)),
		Resources: options.resources,
		ClusterId: h.config.clusterId,
	}
	if options.description != "" {
		request.Description = ptr.To(options.description)
	}
	if options.namespace != "" {
		request.RequestedNamespace = ptr.To(options.namespace)
	}
	return request
}

func init() {
	handler = newCreationHandler()
	cmd := Create()
	Cmd.AddCommand(cmd)
	handler.SetCreateCommandFlags(cmd)
}

func (h *creationHandler) SetCreateCommandFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&h.options.namespace, "namespace", "", "Namespace for the project")
	cmd.Flags().StringVar(&h.options.departmentName, "department", DefaultDepartmentName, "Department for the project")

	cmd.Flags().Var(cli.NewNullableResourcesFlag(&h.options.resources), "resources", description)
	cmd.Flags().StringVar(&h.options.description, "description", "", "Description for the project")
}
