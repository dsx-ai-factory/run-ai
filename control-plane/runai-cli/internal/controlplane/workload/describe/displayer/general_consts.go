/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package displayer

const (
	// Phase enum values
	Creating     Phase = "Creating"
	Initializing Phase = "Initializing"
	Resuming     Phase = "Resuming"
	Pending      Phase = "Pending"
	Deleting     Phase = "Deleting"
	Running      Phase = "Running"
	Updating     Phase = "Updating"
	Stopped      Phase = "Stopped"
	Stopping     Phase = "Stopping"
	Degraded     Phase = "Degraded"
	Failed       Phase = "Failed"
	Completed    Phase = "Completed"
	Terminating  Phase = "Terminating"
	Unknown      Phase = "Unknown"

	// Field names
	FieldName                 = "Name"
	FieldID                   = "ID"
	FieldType                 = "Type"
	FieldPhase                = "Phase"
	FieldPhaseMessage         = "Phase Message"
	FieldCreatedAt            = "Created At"
	FieldUpdatedAt            = "Updated At"
	FieldProject              = "Project"
	FieldDepartment           = "Department"
	FieldSubmittedBy          = "Submitted By"
	FieldPriority             = "Priority"
	FieldCategory             = "Category"
	FieldTopology             = "Topology"
	FieldDistributedFramework = "Distributed Framework"
	FieldCLICommand           = "CLI Command"
	FieldLabels               = "Labels"
	FieldAnnotations          = "Annotations"
	FieldWorkerLabels         = "Worker Labels"
	FieldMasterLabels         = "Master Labels"
)

var AllowedPhaseEnumValues = []Phase{
	"Creating",
	"Initializing",
	"Resuming",
	"Pending",
	"Deleting",
	"Running",
	"Updating",
	"Stopped",
	"Stopping",
	"Degraded",
	"Failed",
	"Completed",
	"Terminating",
	"Unknown",
}
