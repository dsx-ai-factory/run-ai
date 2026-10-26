/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package describe

import (
	"encoding/json"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/ui/field"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"gopkg.in/yaml.v3"
)

const (
	fieldName           = "Name"
	fieldID             = "ID"
	fieldCategory       = "Category"
	fieldPriority       = "Priority"
	fieldPreemptibility = "Preemptibility"
	fieldGroup          = "Group"
	fieldCreatedAt      = "Created At"
	fieldCreatedBy      = "Created By"
	fieldUpdatedAt      = "Updated At"
	fieldUpdatedBy      = "Updated By"

	titleKartas = "Kartas"

	metadataColumnPadding = 28
)

type DisplayField struct {
	Name  string
	Value string
}

type BriefWorkloadType struct {
	Name               string                                                   `json:"name"`
	ID                 string                                                   `json:"id,omitempty"`
	Category           string                                                   `json:"category,omitempty"`
	Priority           string                                                   `json:"priority,omitempty"`
	Preemptibility     string                                                   `json:"preemptibility"`
	Group              string                                                   `json:"group"`
	Kartas             []openapi.WorkloadTypeKartasWithVersionInner             `json:"kartas,omitempty"`
	ResourceInterfaces []openapi.WorkloadTypeResourceInterfacesWithVersionInner `json:"resourceInterfaces,omitempty"`
}

type WorkloadTypeDisplayer struct {
	wt *openapi.WorkloadTypeConfig
}

func NewWorkloadTypeDisplayer(wt *openapi.WorkloadTypeConfig) *WorkloadTypeDisplayer {
	return &WorkloadTypeDisplayer{wt: wt}
}

func (d *WorkloadTypeDisplayer) AsTable() error {
	d.displayMetadata()

	if len(d.wt.Kartas) > 0 {
		if err := d.displayKartas(d.wt.Kartas); err != nil {
			return err
		}
	} else if len(d.wt.ResourceInterfaces) > 0 {
		if err := d.displayKartas(d.wt.ResourceInterfaces); err != nil {
			return err
		}
	}

	return nil
}

func (d *WorkloadTypeDisplayer) displayMetadata() {
	wt := d.wt

	fields := []DisplayField{
		{fieldName, wt.Name},
		{fieldID, field.PtrStringOrDash(wt.Id)},
		{fieldCategory, field.PtrStringOrDash(wt.CategoryName)},
		{fieldPriority, field.PtrStringOrDash(wt.PriorityName)},
		{fieldPreemptibility, string(wt.Preemptibility)},
		{fieldGroup, wt.Group},
		{fieldCreatedAt, field.TimeOrEmpty(wt.CreatedAt)},
		{fieldCreatedBy, field.PtrStringOrDash(wt.CreatedBy)},
		{fieldUpdatedAt, field.TimeOrEmpty(wt.UpdatedAt)},
		{fieldUpdatedBy, field.PtrStringOrDash(wt.UpdatedBy)},
	}

	for _, f := range fields {
		if f.Value != "" && f.Value != "<nil>" {
			fmt.Printf("%-*s %s\n", metadataColumnPadding, f.Name+":", f.Value) //nolint:forbidigo // allowed for user printout
		}
	}
}

func (d *WorkloadTypeDisplayer) displayKartas(kartas any) error {
	fmt.Printf("\n%s:\n", titleKartas) //nolint:forbidigo // allowed for user printout

	data, err := yaml.Marshal(kartas)
	if err != nil {
		return fmt.Errorf("error marshalling kartas: %w", err)
	}
	fmt.Println(string(data)) //nolint:forbidigo // allowed for user printout
	return nil
}

func (d *WorkloadTypeDisplayer) AsJSON() error {
	brief := toBrief(d.wt)
	data, err := json.MarshalIndent(brief, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling to JSON: %w", err)
	}
	fmt.Println(string(data)) //nolint:forbidigo // allowed for user printout
	return nil
}

func (d *WorkloadTypeDisplayer) AsYAML() error {
	brief := toBrief(d.wt)
	data, err := yaml.Marshal(brief)
	if err != nil {
		return fmt.Errorf("error marshalling to YAML: %w", err)
	}
	fmt.Print(string(data)) //nolint:forbidigo // allowed for user printout
	return nil
}

func toBrief(wt *openapi.WorkloadTypeConfig) BriefWorkloadType {
	brief := BriefWorkloadType{
		Name:           wt.Name,
		ID:             field.StringOrEmpty(wt.Id),
		Category:       field.StringOrEmpty(wt.CategoryName),
		Priority:       field.StringOrEmpty(wt.PriorityName),
		Preemptibility: string(wt.Preemptibility),
		Group:          wt.Group,
	}
	if len(wt.Kartas) > 0 {
		brief.Kartas = wt.Kartas
	} else {
		brief.ResourceInterfaces = wt.ResourceInterfaces
	}
	return brief
}

func RenderOutput(wt *openapi.WorkloadTypeConfig, outputFormat string) error {
	displayer := NewWorkloadTypeDisplayer(wt)

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
