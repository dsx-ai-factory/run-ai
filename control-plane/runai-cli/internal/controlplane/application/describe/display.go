/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package describe

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/ui/field"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"gopkg.in/yaml.v3"
)

const (
	columnPadding = 30

	sectionGeneral      = "GENERAL"
	sectionChartRef     = "CHART REFERENCE"
	sectionRequestedRes = "REQUESTED RESOURCES"
	sectionAllocatedRes = "ALLOCATED RESOURCES"
	sectionResources    = "RESOURCES"
)

func RenderOutput(w io.Writer, app *openapi.AIApplication, format string) error {
	switch format {
	case "table", "":
		return renderTable(w, app)
	case "json":
		return renderJSON(w, app)
	case "yaml":
		return renderYAML(w, app)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

func renderJSON(w io.Writer, app *openapi.AIApplication) error {
	b, err := json.MarshalIndent(app, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling to JSON: %w", err)
	}
	fmt.Fprintln(w, string(b))
	return nil
}

func renderYAML(w io.Writer, app *openapi.AIApplication) error {
	b, err := json.Marshal(app)
	if err != nil {
		return fmt.Errorf("error converting to YAML: %w", err)
	}
	var data interface{}
	if err = json.Unmarshal(b, &data); err != nil {
		return fmt.Errorf("error converting to YAML: %w", err)
	}
	yamlBytes, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshalling to YAML: %w", err)
	}
	fmt.Fprint(w, string(yamlBytes))
	return nil
}

func renderTable(w io.Writer, app *openapi.AIApplication) error {
	printSection(w, sectionGeneral)
	printField(w, "Name", field.StringOrEmpty(app.Name))
	printField(w, "ID", field.StringOrEmpty(app.Id))
	printField(w, "Type", field.StringOrEmpty(app.Type))
	printField(w, "Source", field.StringOrEmpty(app.Source))
	if app.ActionStatus.IsSet() && app.ActionStatus.Get() != nil {
		printField(w, "Status", *app.ActionStatus.Get())
	}
	if app.ActionMessage.IsSet() && app.ActionMessage.Get() != nil {
		printField(w, "Message", *app.ActionMessage.Get())
	}
	printField(w, "Project", field.StringOrEmpty(app.ProjectName))
	printField(w, "Created By", field.StringOrEmpty(app.CreatedBy))
	printField(w, "Created At", field.TimeOrEmpty(app.CreatedAt))

	if app.Spec != nil {
		fmt.Fprintln(w)
		printSection(w, sectionChartRef)
		chartRef := app.Spec.ChartRef
		if chartRef.ChartURL != nil && *chartRef.ChartURL != "" {
			printField(w, "Chart URL", *chartRef.ChartURL)
		} else {
			printField(w, "Chart Name", field.StringOrEmpty(chartRef.ChartName))
			printField(w, "Chart Version", field.StringOrEmpty(chartRef.ChartVersion))
			printField(w, "Repository", field.StringOrEmpty(chartRef.Repo))
		}
	}

	if app.RequestedResources.IsSet() && app.RequestedResources.Get() != nil {
		r := app.RequestedResources.Get()
		gpu := formatCores(r.Gpu)
		gpuMem := formatQuantity(r.GpuMemory)
		cpu := formatCores(r.Cpu)
		cpuMem := formatQuantity(r.CpuMemory)
		if gpu != "" || gpuMem != "" || cpu != "" || cpuMem != "" {
			fmt.Fprintln(w)
			printSection(w, sectionRequestedRes)
			printField(w, "GPU", gpu)
			printField(w, "GPU Memory", gpuMem)
			printField(w, "CPU", cpu)
			printField(w, "CPU Memory", cpuMem)
		}
	}

	if app.AllocatedResources.IsSet() && app.AllocatedResources.Get() != nil {
		a := app.AllocatedResources.Get()
		gpu, gpuMem, cpu, cpuMem := "", "", "", ""
		if a.Gpu.IsSet() && a.Gpu.Get() != nil {
			gpu = strconv.FormatFloat(*a.Gpu.Get(), 'f', -1, 64)
		}
		if a.GpuMemory.IsSet() && a.GpuMemory.Get() != nil {
			gpuMem = *a.GpuMemory.Get()
		}
		if a.Cpu.IsSet() && a.Cpu.Get() != nil {
			cpu = strconv.FormatFloat(*a.Cpu.Get(), 'f', -1, 64)
		}
		if a.CpuMemory.IsSet() && a.CpuMemory.Get() != nil {
			cpuMem = *a.CpuMemory.Get()
		}
		if gpu != "" || gpuMem != "" || cpu != "" || cpuMem != "" {
			fmt.Fprintln(w)
			printSection(w, sectionAllocatedRes)
			printField(w, "GPU", gpu)
			printField(w, "GPU Memory", gpuMem)
			printField(w, "CPU", cpu)
			printField(w, "CPU Memory", cpuMem)
		}
	}

	if len(app.Resources) > 0 {
		fmt.Fprintln(w)
		printSection(w, fmt.Sprintf("%s (%d)", sectionResources, len(app.Resources)))
		fmt.Fprintf(w, "%-30s %-21s %s\n", "NAME", "KIND", "ID")
		for _, r := range app.Resources {
			name := field.StringOrEmpty(r.Name)
			kind := field.StringOrEmpty(r.Kind)
			id := field.StringOrEmpty(r.Id)
			if len(kind) > 20 {
				kind = kind[:17] + "..."
			}
			fmt.Fprintf(w, "%-30s %-21s %s\n", name, kind, id)
		}
	}

	return nil
}

func printSection(w io.Writer, title string) {
	fmt.Fprintln(w, title)
}

func printField(w io.Writer, name, value string) {
	if value == "" {
		return
	}
	fmt.Fprintf(w, "%-*s %s\n", columnPadding, name+":", value)
}

func formatCores(c openapi.NullableRequestResourceCores) string {
	if !c.IsSet() || c.Get() == nil {
		return ""
	}
	cores := c.Get()
	if cores.Request.IsSet() && cores.Request.Get() != nil {
		return strconv.FormatFloat(*cores.Request.Get(), 'f', -1, 64)
	}
	if cores.Limit.IsSet() && cores.Limit.Get() != nil {
		return strconv.FormatFloat(*cores.Limit.Get(), 'f', -1, 64)
	}
	return ""
}

func formatQuantity(q openapi.NullableRequestResourceQuantity) string {
	if !q.IsSet() || q.Get() == nil {
		return ""
	}
	qty := q.Get()
	if qty.Request.IsSet() && qty.Request.Get() != nil {
		return *qty.Request.Get()
	}
	if qty.Limit.IsSet() && qty.Limit.Get() != nil {
		return *qty.Limit.Get()
	}
	return ""
}
