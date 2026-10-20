/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package displayer

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/ui/field"
)

// Type definitions
type Container struct {
	Name      string     `json:"name"`
	Image     string     `json:"image"`
	StartedAt *time.Time `json:"startedAt,omitempty"`
}

type Pod struct {
	Name            string      `json:"name"`
	NodeName        *string     `json:"nodeName,omitempty"`
	CompletedAt     *time.Time  `json:"completedAt,omitempty"`
	CurrentNodePool *string     `json:"currentNodePool,omitempty"`
	K8sPhase        string      `json:"k8sPhase"`
	Containers      []Container `json:"containers"`
	Role            *string     `json:"role,omitempty"`
}

// Displayer
type PodsDisplayer struct {
	pods              []Pod
	includeContainers bool
	fields            []string
	runningPods       int32
	replicas          *int32
	workersPerReplica *int32
}

func NewPodsSection(pods []Pod, includeContainers bool, runningPods int32, replicas *int32, workersPerReplica *int32) *Section {
	hasRoles := false
	for _, pod := range pods {
		if pod.Role != nil {
			hasRoles = true
			break
		}
	}

	tableFields := []string{FieldName}
	if hasRoles {
		tableFields = append(tableFields, FieldRole)
	}
	tableFields = append(tableFields, FieldNode, FieldPhase, FieldCompletedAt, FieldNodePool)
	if includeContainers {
		tableFields = append(tableFields, FieldContainers)
	}

	return &Section{
		Displayer: &PodsDisplayer{
			pods:              pods,
			includeContainers: includeContainers,
			fields:            tableFields,
			runningPods:       runningPods,
			replicas:          replicas,
			workersPerReplica: workersPerReplica,
		},
	}
}

func (p *PodsDisplayer) Subject() string {
	return TitlePods
}

func (p *PodsDisplayer) Columns() []table.Column {
	var columns []table.Column
	for _, f := range p.fields {
		width := PodColumnWidth
		if f == FieldRole {
			width = 8
		}
		columns = append(columns, table.Column{Title: f, Width: width})
	}
	return columns
}

func (p *PodsDisplayer) Data() [][]string {
	var data [][]string
	for _, pod := range p.pods {
		var row []string
		for _, fieldType := range p.fields {
			switch fieldType {
			case FieldName:
				row = append(row, pod.Name)
			case FieldRole:
				row = append(row, field.StringOrEmpty(pod.Role))
			case FieldNode:
				row = append(row, field.StringOrEmpty(pod.NodeName))
			case FieldPhase:
				row = append(row, pod.K8sPhase)
			case FieldCompletedAt:
				row = append(row, field.TimeOrEmpty(pod.CompletedAt))
			case FieldNodePool:
				row = append(row, field.StringOrEmpty(pod.CurrentNodePool))
			case FieldContainers:
				containerNames := make([]string, len(pod.Containers))
				for i, container := range pod.Containers {
					containerNames[i] = container.Name
				}
				row = append(row, strings.Join(containerNames, "\n"))
			case FieldImageSingular:
				containerImages := make([]string, len(pod.Containers))
				for i, container := range pod.Containers {
					containerImages[i] = container.Image
				}
				row = append(row, strings.Join(containerImages, ", "))
			}
		}
		data = append(data, row)
	}
	return data
}

func (p *PodsDisplayer) DisplayBeforeTable(indent string) {
	totalPods := len(p.pods)
	columnPadding := GeneralAndRuntimeColumnPadding

	// Build running pods line with optional role breakdown for distributed workloads
	runningPodsLine := fmt.Sprintf("%d/%d", p.runningPods, totalPods)
	if p.hasRoleColumn() {
		leaderCount, workerCount := p.countPodRoles()
		runningPodsLine += fmt.Sprintf(" (%d leaders, %d workers)", leaderCount, workerCount)
	}
	fmt.Printf("%s%-*s %s\n", indent, columnPadding, FieldRunningPods+":", runningPodsLine) //nolint:forbidigo

	// Show requested pods for distributed workloads
	if p.replicas != nil && *p.replicas > 0 {
		workers := int32(0)
		if p.workersPerReplica != nil {
			workers = *p.workersPerReplica
		}
		total := *p.replicas * (1 + workers)
		fmt.Printf("%s%-*s Leaders: %d, Workers: %d, Total: %d\n", indent, columnPadding, FieldRequestedPods+":", *p.replicas, *p.replicas*workers, total) //nolint:forbidigo
	}
}

func (p *PodsDisplayer) countPodRoles() (leaders, workers int) {
	for _, pod := range p.pods {
		if pod.Role != nil {
			switch *pod.Role {
			case RoleLeader:
				leaders++
			case RoleWorker:
				workers++
			}
		}
	}
	return
}

func (p *PodsDisplayer) hasRoleColumn() bool {
	for _, field := range p.fields {
		if field == FieldRole {
			return true
		}
	}
	return false
}
