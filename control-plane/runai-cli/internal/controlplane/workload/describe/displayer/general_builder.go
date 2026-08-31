/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package displayer

import (
	"time"
)

// Builder
type GeneralBuilder struct {
	general General
}

func NewGeneralBuilder() *GeneralBuilder {
	return &GeneralBuilder{general: General{}}
}

func (b *GeneralBuilder) Name(name string) *GeneralBuilder {
	b.general.Name = name
	return b
}

func (b *GeneralBuilder) ID(id string) *GeneralBuilder {
	b.general.ID = id
	return b
}

func (b *GeneralBuilder) DistributedFramework(framework string) *GeneralBuilder {
	b.general.DistributedFramework = &framework
	return b
}

func (b *GeneralBuilder) Type(t string) *GeneralBuilder {
	b.general.Type = t
	return b
}

func (b *GeneralBuilder) Phase(phase Phase) *GeneralBuilder {
	b.general.Phase = phase
	return b
}

func (b *GeneralBuilder) PhaseMessage(message string) *GeneralBuilder {
	b.general.PhaseMessage = &message
	return b
}

func (b *GeneralBuilder) Project(project string) *GeneralBuilder {
	b.general.Project = project
	return b
}

func (b *GeneralBuilder) Department(department string) *GeneralBuilder {
	b.general.Department = department
	return b
}

func (b *GeneralBuilder) CreatedAt(createdAt time.Time) *GeneralBuilder {
	b.general.CreatedAt = createdAt
	return b
}

func (b *GeneralBuilder) UpdatedAt(updatedAt time.Time) *GeneralBuilder {
	b.general.UpdatedAt = updatedAt
	return b
}

func (b *GeneralBuilder) SubmittedBy(submittedBy string) *GeneralBuilder {
	b.general.SubmittedBy = &submittedBy
	return b
}

func (b *GeneralBuilder) Priority(priority string) *GeneralBuilder {
	b.general.Priority = &priority
	return b
}

func (b *GeneralBuilder) Category(category string) *GeneralBuilder {
	b.general.Category = category
	return b
}

func (b *GeneralBuilder) Topology(topology string) *GeneralBuilder {
	b.general.Topology = &topology
	return b
}

func (b *GeneralBuilder) EnvironmentVariables(envVars map[string]string) *GeneralBuilder {
	b.general.EnvironmentVariables = &envVars
	return b
}

func (b *GeneralBuilder) Arguments(arguments string) *GeneralBuilder {
	b.general.Arguments = &arguments
	return b
}

func (b *GeneralBuilder) Command(command string) *GeneralBuilder {
	b.general.Command = &command
	return b
}

func (b *GeneralBuilder) RunaiCliCommand(cliCommand string) *GeneralBuilder {
	b.general.RunaiCliCommand = &cliCommand
	return b
}

func (b *GeneralBuilder) Build() General {
	return b.general
}
