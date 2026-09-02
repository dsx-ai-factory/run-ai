/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package service

import (
	"encoding/json"
	"fmt"

	config "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/generated"
	"go.uber.org/zap"

	"github.com/charmbracelet/bubbles/table"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	table2 "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/ui/table"
	"gopkg.in/yaml.v3"
)

const (
	DefaultWidthName            = 10
	DefaultWidthGPUQuota        = 10
	DefaultWidthAllocatedGPUs   = 15
	DefaultWidthAllocationRatio = 20

	Percent = 100.0
)

type BriefNodepoolAssignedGPU struct {
	Deserved float32 `json:"deserved,omitempty"`
}

type BriefNodePoolAssignedResources struct {
	Gpu BriefNodepoolAssignedGPU `json:"gpu,omitempty"`
}

type BriefResources struct {
	Gpu openapi.ResourcesGpu `json:"gpu,omitempty"`
}

type BriefTotalResources struct {
	GpuQuota openapi.NullableFloat64 `json:"gpuQuota,omitempty"`
}

type BriefDepartmentStatus struct {
	QuotaStatus BriefQuotaStatus `json:"quotaStatus,omitempty"`
}

type BriefQuotaStatusAllocation struct {
	Gpu float32 `json:"gpu,omitempty"`
}

type BriefQuotaStatus struct {
	Allocated BriefQuotaStatusAllocation `json:"allocated,omitempty"`
}

type BriefDepartment struct {
	Id             string                `json:"id"`
	Name           string                `json:"name"`
	Resources      []BriefResources      `json:"resource,omitempty"`
	Status         BriefDepartmentStatus `json:"status"`
	TotalResources BriefTotalResources   `json:"totalResources,omitempty"`
}

type Data struct {
	Departments []BriefDepartment `json:"departments,omitempty" yaml:"departments"`
	NextToken   *string           `json:"nextToken,omitempty" yaml:"nextToken,omitempty"`
}

type BriefDepartmentHandler struct {
	departmentList *openapi.GetDepartments200Response
	data           Data
}

func NewBriefDepartmentHandler(departmentList *openapi.GetDepartments200Response) *BriefDepartmentHandler {
	brief := BriefDepartmentHandler{departmentList: departmentList}
	if err := brief.ReduceStruct(); err != nil {
		return nil
	}
	return &brief
}

func (bph *BriefDepartmentHandler) ReduceStruct() error {
	logger := config.GetInstance().GetLogger()
	if bph.departmentList == nil {
		logger.Error("department list is missing")
		return fmt.Errorf("list attribute is missing")
	}

	marshal, err := json.Marshal(bph.departmentList.GetDepartments())
	if err != nil {
		logger.Error("failed to marshal department list", zap.Error(err))
		return err
	}

	err = json.Unmarshal(marshal, &bph.data.Departments)
	if err != nil {
		logger.Error("failed to unmarshal department list", zap.String("error", err.Error()))
		return err
	}
	return nil
}

func (bph *BriefDepartmentHandler) Subject() string {
	return consts.DepartmentStr
}

func (bph *BriefDepartmentHandler) Columns() []table.Column {
	return []table.Column{
		{Title: "Name", Width: DefaultWidthName},
		{Title: "GPU Quota", Width: DefaultWidthGPUQuota}, // UI use the value status so we keep the same
		{Title: "Allocated GPUs", Width: DefaultWidthAllocatedGPUs},
		{Title: "GPU Allocation Ratio", Width: DefaultWidthAllocationRatio},
	}
}

func (bph *BriefDepartmentHandler) Data() [][]string {
	var data [][]string

	for _, brief := range bph.data.Departments {
		var row []string
		row = append(row, brief.Name)

		sumQuotaGpu := 0.0
		if brief.TotalResources.GpuQuota.IsSet() {
			sumQuotaGpu = *brief.TotalResources.GpuQuota.Get()
		}

		sumAllocatedGpu := float64(brief.Status.QuotaStatus.Allocated.Gpu)

		row = append(row, fmt.Sprintf("%.2f", sumQuotaGpu))

		allocationRatioPercent := 0.0
		if sumQuotaGpu > 0 {
			allocationRatioPercent = (sumAllocatedGpu / sumQuotaGpu) * Percent
		}
		row = append(row, fmt.Sprintf("%.2f", sumAllocatedGpu))
		row = append(row, fmt.Sprintf("%.2f%%", allocationRatioPercent))

		data = append(data, row)
	}
	return data
}

func (bph *BriefDepartmentHandler) GenerateTable() (string, error) {
	return table2.NewTableDisplay(bph).Generate()
}

func (bph *BriefDepartmentHandler) GenerateJson() (string, error) {
	output, err := json.MarshalIndent(bph.data, "", "  ")
	return string(output), err
}

func (bph *BriefDepartmentHandler) GenerateYaml() (string, error) {
	output, err := yaml.Marshal(bph.data)
	return string(output), err
}

func (bph *BriefDepartmentHandler) GetList() []BriefDepartment {
	return bph.data.Departments
}
