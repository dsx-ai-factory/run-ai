/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package describe

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/ui/field"
	uitable "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/ui/table"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

const (
	fieldName        = "Name"
	fieldScope       = "Scope"
	fieldCreatedBy   = "Created By"
	fieldCreatedAt   = "Created At"
	fieldUpdatedAt   = "Last Updated"
	fieldDescription = "Description"
	fieldStatus      = "Status"

	titleGPUResources      = "GPU Resources"
	titleCPUResources      = "CPU Resources"
	titleExtendedResources = "Extended Resources"

	metadataColumnPadding = 28
	resourceColumnWidth   = 25
	valueColumnWidth      = 20

	sectionsSeparator = "\n\n"

	// Column titles
	columnTitleResource = "Resource"
	columnTitleValue    = "Value"
	columnTitleQuantity = "Quantity"

	// GPU resource field names
	fieldGPUDevicesRequest = "GPU Devices Request"
	fieldGPURequestType    = "GPU Request Type"
	fieldGPUMemoryRequest  = "GPU Memory Request"
	fieldGPUMemoryLimit    = "GPU Memory Limit"
	fieldGPUPortionRequest = "GPU Portion Request"
	fieldGPUPortionLimit   = "GPU Portion Limit"

	// CPU resource field names
	fieldCPUCoresRequest  = "CPU Cores Request"
	fieldCPUCoresLimit    = "CPU Cores Limit"
	fieldCPUMemoryRequest = "CPU Memory Request"
	fieldCPUMemoryLimit   = "CPU Memory Limit"
)

type DisplayField struct {
	Name  string
	Value interface{}
}

type DisplaySection struct {
	Displayer uitable.Constructor
}

type ComputeFormatter struct {
	compute *openapi.ComputeAsset
}

func NewComputeFormatter(compute *openapi.ComputeAsset) (*ComputeFormatter, error) {
	if compute == nil {
		return nil, fmt.Errorf("compute is nil")
	}
	return &ComputeFormatter{
		compute: compute,
	}, nil
}

func (f *ComputeFormatter) Displayer() *ComputeDisplayer {
	return NewComputeDisplayer(f.compute)
}

type ComputeDisplayer struct {
	compute *openapi.ComputeAsset
}

func NewComputeDisplayer(compute *openapi.ComputeAsset) *ComputeDisplayer {
	return &ComputeDisplayer{compute: compute}
}

func (cd *ComputeDisplayer) AsTable() error {
	cd.DisplayMetadata()

	sections := []struct {
		show        bool
		displayFunc func() error
	}{
		{cd.hasGPUResources(), cd.displayGPUResourcesTable},
		{cd.hasCPUResources(), cd.displayCPUResourcesTable},
		{cd.hasExtendedResources(), cd.displayExtendedResourcesTable},
	}

	for _, section := range sections {
		if section.show {
			if err := section.displayFunc(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (cd *ComputeDisplayer) DisplayMetadata() {
	compute := cd.compute

	fields := []DisplayField{
		{fieldName, compute.Meta.Name},
		{fieldScope, formatScope(compute.Meta)},
		{fieldCreatedBy, compute.Meta.CreatedBy},
	}

	createdAt, _ := compute.Meta.GetCreatedAtOk()
	fields = append(fields, DisplayField{fieldCreatedAt, field.TimeOrEmpty(createdAt)})

	updatedAt, _ := compute.Meta.GetUpdatedAtOk()
	fields = append(fields, DisplayField{fieldUpdatedAt, field.TimeOrEmpty(updatedAt)})

	if compute.Meta.Description.IsSet() && compute.Meta.Description.Get() != nil && *compute.Meta.Description.Get() != "" {
		fields = append(fields, DisplayField{fieldDescription, *compute.Meta.Description.Get()})
	}

	fields = append(fields, DisplayField{fieldStatus, "Ready"})

	displayFields(fields)
}

func displayFields(fields []DisplayField) {
	for _, f := range fields {
		if value := fmt.Sprintf("%v", f.Value); value != "" && value != "<nil>" {
			fmt.Printf("%-*s %s\n", metadataColumnPadding, f.Name+":", value) //nolint:forbidigo // allowed for user printout
		}
	}
}

func (cd *ComputeDisplayer) displayGPUResourcesTable() error {
	displayer := NewGPUResourcesDisplayer(cd.compute)
	return cd.displaySection(displayer.Displayer)
}

func (cd *ComputeDisplayer) displayCPUResourcesTable() error {
	displayer := NewCPUResourcesDisplayer(cd.compute)
	return cd.displaySection(displayer.Displayer)
}

func (cd *ComputeDisplayer) displayExtendedResourcesTable() error {
	displayer := NewExtendedResourcesDisplayer(cd.compute)
	return cd.displaySection(displayer.Displayer)
}

func (cd *ComputeDisplayer) displaySection(displayer uitable.Constructor) error {
	if len(displayer.Data()) == 0 {
		return nil
	}

	fmt.Printf("%s%s%s", sectionsSeparator, displayer.Subject(), sectionsSeparator)
	tableDisplay := uitable.NewTableDisplay(displayer)
	output, err := tableDisplay.Generate()
	if err != nil {
		return err
	}
	fmt.Println(output)

	return nil
}

type GPUResourcesDisplayer struct {
	compute *openapi.ComputeAsset
}

func NewGPUResourcesDisplayer(compute *openapi.ComputeAsset) *DisplaySection {
	return &DisplaySection{
		Displayer: &GPUResourcesDisplayer{compute: compute},
	}
}

func (g *GPUResourcesDisplayer) Subject() string {
	return titleGPUResources
}

func (g *GPUResourcesDisplayer) Columns() []table.Column {
	return []table.Column{
		{Title: columnTitleResource, Width: resourceColumnWidth},
		{Title: columnTitleValue, Width: valueColumnWidth},
	}
}

func (g *GPUResourcesDisplayer) Data() [][]string {
	var data [][]string
	compute := g.compute

	if compute.Spec.GpuDevicesRequest.IsSet() && compute.Spec.GpuDevicesRequest.Get() != nil {
		data = append(data, []string{fieldGPUDevicesRequest, field.Int32OrEmpty(compute.Spec.GpuDevicesRequest.Get())})
	}

	if compute.Spec.GpuRequestType.IsSet() && compute.Spec.GpuRequestType.Get() != nil {
		data = append(data, []string{fieldGPURequestType, string(*compute.Spec.GpuRequestType.Get())})
	}

	if compute.Spec.GpuMemoryRequest.IsSet() && compute.Spec.GpuMemoryRequest.Get() != nil {
		data = append(data, []string{fieldGPUMemoryRequest, *compute.Spec.GpuMemoryRequest.Get()})
	}

	if compute.Spec.GpuMemoryLimit.IsSet() && compute.Spec.GpuMemoryLimit.Get() != nil {
		data = append(data, []string{fieldGPUMemoryLimit, *compute.Spec.GpuMemoryLimit.Get()})
	}

	if compute.Spec.GpuPortionRequest.IsSet() && compute.Spec.GpuPortionRequest.Get() != nil {
		data = append(data, []string{fieldGPUPortionRequest, field.Float64OrEmpty(compute.Spec.GpuPortionRequest.Get())})
	}

	if compute.Spec.GpuPortionLimit.IsSet() && compute.Spec.GpuPortionLimit.Get() != nil {
		data = append(data, []string{fieldGPUPortionLimit, field.Float64OrEmpty(compute.Spec.GpuPortionLimit.Get())})
	}

	return data
}

type CPUResourcesDisplayer struct {
	compute *openapi.ComputeAsset
}

func NewCPUResourcesDisplayer(compute *openapi.ComputeAsset) *DisplaySection {
	return &DisplaySection{
		Displayer: &CPUResourcesDisplayer{compute: compute},
	}
}

func (c *CPUResourcesDisplayer) Subject() string {
	return titleCPUResources
}

func (c *CPUResourcesDisplayer) Columns() []table.Column {
	return []table.Column{
		{Title: columnTitleResource, Width: resourceColumnWidth},
		{Title: columnTitleValue, Width: valueColumnWidth},
	}
}

func (c *CPUResourcesDisplayer) Data() [][]string {
	var data [][]string
	compute := c.compute

	if compute.Spec.CpuCoreRequest.IsSet() && compute.Spec.CpuCoreRequest.Get() != nil {
		data = append(data, []string{fieldCPUCoresRequest, field.Float64OrEmpty(compute.Spec.CpuCoreRequest.Get())})
	}

	if compute.Spec.CpuCoreLimit.IsSet() && compute.Spec.CpuCoreLimit.Get() != nil {
		data = append(data, []string{fieldCPUCoresLimit, field.Float64OrEmpty(compute.Spec.CpuCoreLimit.Get())})
	}

	if compute.Spec.CpuMemoryRequest.IsSet() && compute.Spec.CpuMemoryRequest.Get() != nil {
		data = append(data, []string{fieldCPUMemoryRequest, *compute.Spec.CpuMemoryRequest.Get()})
	}

	if compute.Spec.CpuMemoryLimit.IsSet() && compute.Spec.CpuMemoryLimit.Get() != nil {
		data = append(data, []string{fieldCPUMemoryLimit, *compute.Spec.CpuMemoryLimit.Get()})
	}

	return data
}

type ExtendedResourcesDisplayer struct {
	compute *openapi.ComputeAsset
}

func NewExtendedResourcesDisplayer(compute *openapi.ComputeAsset) *DisplaySection {
	return &DisplaySection{
		Displayer: &ExtendedResourcesDisplayer{compute: compute},
	}
}

func (e *ExtendedResourcesDisplayer) Subject() string {
	return titleExtendedResources
}

func (e *ExtendedResourcesDisplayer) Columns() []table.Column {
	return []table.Column{
		{Title: columnTitleResource, Width: resourceColumnWidth},
		{Title: columnTitleQuantity, Width: valueColumnWidth},
	}
}

func (e *ExtendedResourcesDisplayer) Data() [][]string {
	var data [][]string

	for _, er := range e.compute.Spec.ExtendedResources {
		if er.Resource.IsSet() && er.Resource.Get() != nil {
			resourceName := *er.Resource.Get()
			value := ""
			if er.Quantity.IsSet() && er.Quantity.Get() != nil {
				value = *er.Quantity.Get()
			}
			if resourceName != "" {
				data = append(data, []string{resourceName, value})
			}
		}
	}

	return data
}

func (cd *ComputeDisplayer) hasGPUResources() bool {
	spec := cd.compute.Spec
	return (spec.GpuDevicesRequest.IsSet() && spec.GpuDevicesRequest.Get() != nil) ||
		(spec.GpuRequestType.IsSet() && spec.GpuRequestType.Get() != nil) ||
		(spec.GpuMemoryRequest.IsSet() && spec.GpuMemoryRequest.Get() != nil) ||
		(spec.GpuMemoryLimit.IsSet() && spec.GpuMemoryLimit.Get() != nil) ||
		(spec.GpuPortionRequest.IsSet() && spec.GpuPortionRequest.Get() != nil) ||
		(spec.GpuPortionLimit.IsSet() && spec.GpuPortionLimit.Get() != nil)
}

func (cd *ComputeDisplayer) hasCPUResources() bool {
	spec := cd.compute.Spec
	return (spec.CpuCoreRequest.IsSet() && spec.CpuCoreRequest.Get() != nil) ||
		(spec.CpuCoreLimit.IsSet() && spec.CpuCoreLimit.Get() != nil) ||
		(spec.CpuMemoryRequest.IsSet() && spec.CpuMemoryRequest.Get() != nil) ||
		(spec.CpuMemoryLimit.IsSet() && spec.CpuMemoryLimit.Get() != nil)
}

func (cd *ComputeDisplayer) hasExtendedResources() bool {
	return len(cd.compute.Spec.ExtendedResources) > 0
}

func (cd *ComputeDisplayer) AsJSON() error {
	asset := ToComputeAsset(cd.compute)
	jsonData, err := json.MarshalIndent(asset, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling to JSON: %w", err)
	}
	fmt.Println(string(jsonData)) //nolint:forbidigo // allowed for user printout
	return nil
}

func (cd *ComputeDisplayer) AsYAML() error {
	asset := ToComputeAsset(cd.compute)
	yamlData, err := yaml.Marshal(asset)
	if err != nil {
		return fmt.Errorf("error marshalling to YAML: %w", err)
	}
	fmt.Println(string(yamlData)) //nolint:forbidigo // allowed for user printout
	return nil
}

func formatScope(meta openapi.AssetMeta) string {
	caser := cases.Title(language.English)
	return caser.String(strings.ToLower(string(meta.Scope)))
}

func RenderOutput(output *ComputeFormatter, outputFormat string) error {
	displayer := output.Displayer()

	switch outputFormat {
	case "table":
		return displayer.AsTable()
	case "json":
		return displayer.AsJSON()
	case "yaml":
		return displayer.AsYAML()
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
}
