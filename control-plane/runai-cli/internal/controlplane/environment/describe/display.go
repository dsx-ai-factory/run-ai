/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package describe

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/ui/field"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

const (
	fieldName               = "Name"
	fieldScope              = "Scope"
	fieldCreatedBy          = "Created By"
	fieldCreatedAt          = "Created At"
	fieldUpdatedAt          = "Last Updated"
	fieldImage              = "Image"
	fieldImagePullPolicy    = "Image Pull Policy"
	fieldWorkingDir         = "Working Directory"
	fieldCommand            = "Command"
	fieldArgs               = "Args"
	fieldTraining           = "Training"
	fieldWorkspace          = "Workspace"
	fieldInference          = "Inference"
	fieldDistributed        = "Distributed"
	fieldUidGidSource       = "UID/GID Source"
	fieldRunAsUid           = "Run As UID"
	fieldRunAsGid           = "Run As GID"
	fieldSupplementalGroups = "Supplemental Groups"
)

type EnvironmentFormatter struct {
	environment *openapi.EnvironmentAsset
}

func NewEnvironmentFormatter(environment *openapi.EnvironmentAsset) (*EnvironmentFormatter, error) {
	// Prevents panic if API returns nil
	if environment == nil {
		return nil, fmt.Errorf("environment is nil")
	}
	return &EnvironmentFormatter{
		environment: environment,
	}, nil
}

func (f *EnvironmentFormatter) Displayer() *EnvironmentDisplayer {
	return NewEnvironmentDisplayer(f.environment)
}

type DisplayWriter struct {
	indent  string
	padding int
}

func newDisplayWriter(indentLevel int, padding int) *DisplayWriter {
	indent := strings.Repeat("  ", indentLevel)
	return &DisplayWriter{
		indent:  indent,
		padding: padding,
	}
}

func (w *DisplayWriter) writeField(name, value string) {
	if value == "" {
		return
	}
	fmt.Printf("%s%-*s %s\n", w.indent, w.padding, name+":", value) //nolint:forbidigo // allowed for user printout
}

func (w *DisplayWriter) writeSection(title string) {
	fmt.Printf("\n%s%s:\n", w.indent, title) //nolint:forbidigo // allowed for user printout
}

func (w *DisplayWriter) writeListItem(value string) {
	fmt.Printf("%s- %s\n", w.indent, value) //nolint:forbidigo // allowed for user printout
}

type EnvironmentDisplayer struct {
	environment *openapi.EnvironmentAsset
}

func NewEnvironmentDisplayer(environment *openapi.EnvironmentAsset) *EnvironmentDisplayer {
	return &EnvironmentDisplayer{environment: environment}
}

func (ed *EnvironmentDisplayer) AsTable() error {
	ed.DisplayGeneral()
	ed.DisplaySpec()
	return nil
}

func (ed *EnvironmentDisplayer) DisplayGeneral() {
	env := ed.environment
	writer := newDisplayWriter(0, 28)

	writer.writeField(fieldName, env.Meta.Name)
	writer.writeField(fieldScope, formatScope(env.Meta))
	writer.writeField(fieldCreatedBy, env.Meta.CreatedBy)
	createdAt, _ := env.Meta.GetCreatedAtOk()
	writer.writeField(fieldCreatedAt, field.TimeOrEmpty(createdAt))
	updatedAt, _ := env.Meta.GetUpdatedAtOk()
	writer.writeField(fieldUpdatedAt, field.TimeOrEmpty(updatedAt))
}

func (ed *EnvironmentDisplayer) DisplaySpec() {
	env := ed.environment
	fmt.Print("\nSpecification:\n") //nolint:forbidigo // allowed for user printout

	specWriter := newDisplayWriter(1, 26)
	if env.Spec.HasImage() {
		specWriter.writeField(fieldImage, env.Spec.GetImage())
	}
	if env.Spec.ImagePullPolicy.IsSet() && env.Spec.ImagePullPolicy.Get() != nil {
		specWriter.writeField(fieldImagePullPolicy, string(*env.Spec.ImagePullPolicy.Get()))
	}
	if env.Spec.WorkingDir.IsSet() && env.Spec.WorkingDir.Get() != nil && *env.Spec.WorkingDir.Get() != "" {
		specWriter.writeField(fieldWorkingDir, *env.Spec.WorkingDir.Get())
	}

	if ed.hasWorkloadTypes() {
		ed.DisplayWorkloadTypes()
	}
	if ed.hasTools() {
		ed.DisplayTools()
	}
	if ed.hasEnvVars() {
		ed.DisplayEnvVars()
	}

	if env.Spec.Command.IsSet() && env.Spec.Command.Get() != nil && *env.Spec.Command.Get() != "" {
		specWriter.writeField(fieldCommand, *env.Spec.Command.Get())
	}
	if env.Spec.Args.IsSet() && env.Spec.Args.Get() != nil && *env.Spec.Args.Get() != "" {
		specWriter.writeField(fieldArgs, *env.Spec.Args.Get())
	}

	if ed.hasSecurity() {
		ed.DisplaySecurity()
	}
	if ed.hasCapabilities() {
		ed.DisplayCapabilities()
	}
}

func (ed *EnvironmentDisplayer) DisplayWorkloadTypes() {
	env := ed.environment
	if !env.Meta.WorkloadSupportedTypes.IsSet() || env.Meta.WorkloadSupportedTypes.Get() == nil {
		return
	}

	wst := env.Meta.WorkloadSupportedTypes.Get()
	sectionWriter := newDisplayWriter(1, 0)
	sectionWriter.writeSection("Workload Types")

	fieldWriter := newDisplayWriter(2, 24)
	fieldWriter.writeField(fieldTraining, formatBool(wst.Training))
	fieldWriter.writeField(fieldWorkspace, formatBool(wst.Workspace))
	fieldWriter.writeField(fieldInference, formatBool(wst.Inference))

	distStr := formatBool(wst.Distributed)
	if wst.Distributed.IsSet() && wst.Distributed.Get() != nil && *wst.Distributed.Get() {
		if wst.DistFramework.IsSet() && wst.DistFramework.Get() != nil && *wst.DistFramework.Get() != "" {
			distStr = "Yes (" + *wst.DistFramework.Get() + ")"
		}
	}
	fieldWriter.writeField(fieldDistributed, distStr)
}

func (ed *EnvironmentDisplayer) DisplayTools() {
	env := ed.environment
	if len(env.Spec.Connections) == 0 {
		return
	}

	sectionWriter := newDisplayWriter(1, 0)
	sectionWriter.writeSection("Tools")

	listWriter := newDisplayWriter(2, 0)
	for _, conn := range env.Spec.Connections {
		if conn.HasInternalToolInfo() {
			toolInfo := conn.GetInternalToolInfo()
			listWriter.writeListItem(string(toolInfo.ToolType))
		} else if conn.HasExternalToolInfo() {
			toolInfo := conn.GetExternalToolInfo()
			listWriter.writeListItem(string(toolInfo.ToolType))
		}
	}
}

func (ed *EnvironmentDisplayer) DisplayEnvVars() {
	env := ed.environment
	if len(env.Spec.EnvironmentVariables) == 0 {
		return
	}

	sectionWriter := newDisplayWriter(1, 0)
	sectionWriter.writeSection("Environment Variables")

	fieldWriter := newDisplayWriter(2, 24)
	for _, envVar := range env.Spec.EnvironmentVariables {
		var name, value string
		if envVar.Name.IsSet() && envVar.Name.Get() != nil {
			name = *envVar.Name.Get()
		}
		if envVar.Value.IsSet() && envVar.Value.Get() != nil {
			value = *envVar.Value.Get()
		}
		if name != "" {
			fieldWriter.writeField(name, value)
		}
	}
}

func (ed *EnvironmentDisplayer) DisplaySecurity() {
	env := ed.environment
	if !env.Spec.UidGidSource.IsSet() || env.Spec.UidGidSource.Get() == nil {
		return
	}

	sectionWriter := newDisplayWriter(1, 0)
	sectionWriter.writeSection("Security")

	fieldWriter := newDisplayWriter(2, 24)
	if env.Spec.UidGidSource.IsSet() && env.Spec.UidGidSource.Get() != nil {
		fieldWriter.writeField(fieldUidGidSource, string(*env.Spec.UidGidSource.Get()))
	}
	if env.Spec.RunAsUid.IsSet() && env.Spec.RunAsUid.Get() != nil {
		fieldWriter.writeField(fieldRunAsUid, field.Int64OrEmpty(env.Spec.RunAsUid.Get()))
	}
	if env.Spec.RunAsGid.IsSet() && env.Spec.RunAsGid.Get() != nil {
		fieldWriter.writeField(fieldRunAsGid, field.Int64OrEmpty(env.Spec.RunAsGid.Get()))
	}
	if env.Spec.SupplementalGroups.IsSet() && env.Spec.SupplementalGroups.Get() != nil {
		fieldWriter.writeField(fieldSupplementalGroups, *env.Spec.SupplementalGroups.Get())
	}
}

func (ed *EnvironmentDisplayer) DisplayCapabilities() {
	env := ed.environment
	if len(env.Spec.Capabilities) == 0 {
		return
	}

	sectionWriter := newDisplayWriter(1, 0)
	sectionWriter.writeSection("Capabilities")

	listWriter := newDisplayWriter(2, 0)
	for _, cap := range env.Spec.Capabilities {
		listWriter.writeListItem(string(cap))
	}
}

func (ed *EnvironmentDisplayer) hasWorkloadTypes() bool {
	return ed.environment.Meta.WorkloadSupportedTypes.IsSet() && ed.environment.Meta.WorkloadSupportedTypes.Get() != nil
}

func (ed *EnvironmentDisplayer) hasEnvVars() bool {
	return len(ed.environment.Spec.EnvironmentVariables) > 0
}

func (ed *EnvironmentDisplayer) hasCapabilities() bool {
	return len(ed.environment.Spec.Capabilities) > 0
}

func (ed *EnvironmentDisplayer) hasTools() bool {
	return len(ed.environment.Spec.Connections) > 0
}

func (ed *EnvironmentDisplayer) hasSecurity() bool {
	return ed.environment.Spec.UidGidSource.IsSet() && ed.environment.Spec.UidGidSource.Get() != nil
}

func (ed *EnvironmentDisplayer) AsJSON() error {
	asset := ToEnvironmentAsset(ed.environment)
	jsonData, err := json.MarshalIndent(asset, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling to JSON: %w", err)
	}
	fmt.Println(string(jsonData)) //nolint:forbidigo // allowed for user printout
	return nil
}

func (ed *EnvironmentDisplayer) AsYAML() error {
	asset := ToEnvironmentAsset(ed.environment)
	yamlData, err := yaml.Marshal(asset)
	if err != nil {
		return fmt.Errorf("error marshalling to YAML: %w", err)
	}
	fmt.Println(string(yamlData)) //nolint:forbidigo // allowed for user printout
	return nil
}

func formatBool(nb openapi.NullableBool) string {
	if nb.IsSet() && *nb.Get() {
		return "Yes"
	}
	return "No"
}

func formatScope(meta openapi.AssetMeta) string {
	caser := cases.Title(language.English)
	return caser.String(strings.ToLower(string(meta.Scope)))
}

func RenderOutput(output *EnvironmentFormatter, outputFormat string) error {
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
