/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package describe

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/ui/field"
	"gopkg.in/yaml.v3"
)

var datasourceTypeMap = map[string]string{
	"nfs":       "Nfs",
	"pvc":       "Pvc",
	"s3":        "S3",
	"git":       "Git",
	"hostPath":  "HostPath",
	"configMap": "ConfigMap",
	"secret":    "Secret",
}

const (
	// Field names - General
	fieldName        = "Name"
	fieldType        = "Type"
	fieldScope       = "Scope"
	fieldDescription = "Description"
	fieldCreatedBy   = "Created By"
	fieldCreatedAt   = "Created At"
	fieldUpdatedBy   = "Last Updated"
	fieldStatus      = "Status"

	// Field names - NFS
	fieldNFSServer    = "Server"
	fieldNFSPath      = "Path"
	fieldNFSMountPath = "Mount Path"
	fieldNFSReadOnly  = "Read Only"

	// Field names - PVC
	fieldPVCPath          = "Container Path"
	fieldPVCExistingPVC   = "Existing PVC"
	fieldPVCClaimName     = "Claim Name"
	fieldPVCReadOnly      = "Read Only"
	fieldPVCEphemeral     = "Ephemeral"
	fieldPVCDataSharing   = "Data Sharing"
	fieldPVCSize          = "Size"
	fieldPVCStorageClass  = "Storage Class"
	fieldPVCAccessModes   = "Access Modes"
	fieldPVCReadWriteOnce = "Read Write Once"
	fieldPVCReadOnlyMany  = "Read Only Many"
	fieldPVCReadWriteMany = "Read Write Many"
	fieldPVCVolumeMode    = "Volume Mode"

	// Field names - S3
	fieldS3Bucket     = "Bucket"
	fieldS3Path       = "Container Path"
	fieldS3URL        = "URL"
	fieldS3Credential = "Credential"

	// Field names - Git
	fieldGitRepository = "Repository"
	fieldGitBranch     = "Branch"
	fieldGitRevision   = "Revision"
	fieldGitPath       = "Container Path"
	fieldGitCredential = "Credential"

	fieldHostPathPath             = "Host Path"
	fieldHostPathMountPath        = "Mount Path"
	fieldHostPathReadOnly         = "Read Only"
	fieldHostPathMountPropagation = "Mount Propagation"

	fieldConfigMapName        = "ConfigMap Name"
	fieldConfigMapMountPath   = "Mount Path"
	fieldConfigMapSubPath     = "Sub Path"
	fieldConfigMapDefaultMode = "Default Mode"

	// Field names - Secret
	fieldSecretCredential  = "Credential"
	fieldSecretMountPath   = "Mount Path"
	fieldSecretDefaultMode = "Default Mode"

	// Section titles
	titleNFSSpec        = "NFS Specification:"
	titlePVCSpec        = "PVC Specification:"
	titlePVCClaimInfo   = "Claim Info:"
	titleS3Spec         = "S3 Specification:"
	titleGitSpec        = "Git Specification:"
	titleHostPathSpec   = "Host Path Specification:"
	titleConfigMapSpec  = "ConfigMap Specification:"
	titleSecretSpec     = "Secret Specification:"
	titleAdditionalAttr = "Additional Attributes:"

	// Date format
	dateFormat = time.RFC3339

	// Column padding
	generalColumnPadding     = 35
	specSectionColumnPadding = 25

	// Section separator
	sectionsSeparator = "\n\n"
)

type DisplayField struct {
	Name  string
	Value interface{}
}

type Displayer interface {
	AsTable() error
	AsJSON() error
	AsYAML() error
}

type DatasourceDisplayer struct {
	datasource *Datasource
}

func NewDatasourceDisplayer(datasource *Datasource) *DatasourceDisplayer {
	return &DatasourceDisplayer{
		datasource: datasource,
	}
}

func (dd *DatasourceDisplayer) AsTable() error {
	return dd.renderTable(dd.datasource)
}

func (dd *DatasourceDisplayer) AsJSON() error {
	return dd.renderJSON(dd.datasource)
}

func (dd *DatasourceDisplayer) AsYAML() error {
	return dd.renderYAML(dd.datasource)
}

func (dd *DatasourceDisplayer) renderJSON(data interface{}) error {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling to JSON: %w", err)
	}
	fmt.Println(string(jsonBytes)) //nolint:forbidigo // CLI output to user
	return nil
}

func (dd *DatasourceDisplayer) renderYAML(data interface{}) error {
	yamlBytes, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshalling to YAML: %w", err)
	}
	fmt.Println(string(yamlBytes)) //nolint:forbidigo // CLI output to user
	return nil
}

func (dd *DatasourceDisplayer) renderTable(datasource *Datasource) error {
	dd.DisplayGeneral(datasource)

	if datasource.Spec != nil {
		if datasource.Spec.NFS != nil {
			dd.displayNFSSpec(datasource.Spec.NFS)
		} else if datasource.Spec.PVC != nil {
			dd.displayPVCSpec(datasource.Spec.PVC)
		} else if datasource.Spec.S3 != nil {
			dd.displayS3Spec(datasource.Spec.S3)
		} else if datasource.Spec.Git != nil {
			dd.displayGitSpec(datasource.Spec.Git)
		} else if datasource.Spec.HostPath != nil {
			dd.displayHostPathSpec(datasource.Spec.HostPath)
		} else if datasource.Spec.ConfigMap != nil {
			dd.displayConfigMapSpec(datasource.Spec.ConfigMap)
		} else if datasource.Spec.Secret != nil {
			dd.displaySecretSpec(datasource.Spec.Secret)
		}
	}

	if len(datasource.AdditionalAttributes) > 0 {
		dd.displayAdditionalAttributes(datasource.AdditionalAttributes)
	}

	return nil
}

func (dd *DatasourceDisplayer) DisplayGeneral(datasource *Datasource) {
	meta := &datasource.Meta

	fields := []DisplayField{
		{fieldName, meta.Name},
		{fieldType, formatDatasourceType(meta.Type)},
		{fieldScope, formatScope(meta.Scope)},
		{fieldCreatedBy, meta.CreatedBy},
		{fieldCreatedAt, meta.CreatedAt.Format(dateFormat)},
		{fieldUpdatedBy, meta.UpdatedBy},
		{fieldUpdatedBy, meta.UpdatedAt.Format(dateFormat)},
	}

	if meta.Description != nil && *meta.Description != "" {
		fields = append(fields, DisplayField{fieldDescription, *meta.Description})
	}

	if datasource.Status != nil && datasource.Status.Status != nil && *datasource.Status.Status != "" {
		fields = append(fields, DisplayField{fieldStatus, *datasource.Status.Status})
	}

	displayFields(fields)
}

func (dd *DatasourceDisplayer) displayNFSSpec(spec *NFSSpec) {
	fmt.Printf("%s%s\n", sectionsSeparator, titleNFSSpec) //nolint:forbidigo // CLI output to user

	fields := []DisplayField{
		{fieldNFSServer, field.StringOrEmpty(spec.Server)},
		{fieldNFSPath, field.StringOrEmpty(spec.Path)},
		{fieldNFSMountPath, field.StringOrEmpty(spec.MountPath)},
		{fieldNFSReadOnly, formatBoolYesNo(spec.ReadOnly)},
	}

	displayIndentedFields(fields)
}

func (dd *DatasourceDisplayer) displayPVCSpec(spec *PVCSpec) {
	fmt.Printf("%s%s\n", sectionsSeparator, titlePVCSpec) //nolint:forbidigo // CLI output to user

	fields := []DisplayField{
		{fieldPVCPath, field.StringOrEmpty(spec.Path)},
		{fieldPVCExistingPVC, formatBoolYesNo(spec.ExistingPVC)},
		{fieldPVCClaimName, field.StringOrEmpty(spec.ClaimName)},
		{fieldPVCReadOnly, formatBoolYesNo(spec.ReadOnly)},
		{fieldPVCEphemeral, formatBoolYesNo(spec.Ephemeral)},
		{fieldPVCDataSharing, formatBoolYesNo(spec.DataSharing)},
	}

	displayIndentedFields(fields)

	if spec.ClaimInfo != nil {
		dd.displayClaimInfo(spec.ClaimInfo)
	}
}

func (dd *DatasourceDisplayer) displayClaimInfo(claimInfo *ClaimInfo) {
	fmt.Printf("\n%s\n", titlePVCClaimInfo) //nolint:forbidigo // CLI output to user

	fields := []DisplayField{
		{fieldPVCSize, field.StringOrEmpty(claimInfo.Size)},
		{fieldPVCStorageClass, field.StringOrEmpty(claimInfo.StorageClass)},
	}

	if claimInfo.AccessModes != nil {
		fields = append(fields,
			DisplayField{fieldPVCReadWriteOnce, formatBoolYesNo(claimInfo.AccessModes.ReadWriteOnce)},
			DisplayField{fieldPVCReadOnlyMany, formatBoolYesNo(claimInfo.AccessModes.ReadOnlyMany)},
			DisplayField{fieldPVCReadWriteMany, formatBoolYesNo(claimInfo.AccessModes.ReadWriteMany)},
		)
	}

	fields = append(fields, DisplayField{fieldPVCVolumeMode, field.StringOrEmpty(claimInfo.VolumeMode)})

	displayIndentedFields(fields)
}

func (dd *DatasourceDisplayer) displayS3Spec(spec *S3Spec) {
	fmt.Printf("%s%s\n", sectionsSeparator, titleS3Spec) //nolint:forbidigo // CLI output to user

	fields := []DisplayField{
		{fieldS3Bucket, field.StringOrEmpty(spec.Bucket)},
		{fieldS3Path, field.StringOrEmpty(spec.Path)},
		{fieldS3URL, field.StringOrEmpty(spec.URL)},
		{fieldS3Credential, field.StringOrEmpty(spec.AccessKeyAssetID)},
	}

	displayIndentedFields(fields)
}

func (dd *DatasourceDisplayer) displayGitSpec(spec *GitSpec) {
	fmt.Printf("%s%s\n", sectionsSeparator, titleGitSpec) //nolint:forbidigo // CLI output to user

	fields := []DisplayField{
		{fieldGitRepository, field.StringOrEmpty(spec.Repository)},
		{fieldGitBranch, field.StringOrEmpty(spec.Branch)},
		{fieldGitRevision, field.StringOrEmpty(spec.Revision)},
		{fieldGitPath, field.StringOrEmpty(spec.Path)},
		{fieldGitCredential, field.StringOrEmpty(spec.CredentialID)},
	}

	displayIndentedFields(fields)
}

func (dd *DatasourceDisplayer) displayHostPathSpec(spec *HostPathSpec) {
	fmt.Printf("%s%s\n", sectionsSeparator, titleHostPathSpec) //nolint:forbidigo // CLI output to user

	fields := []DisplayField{
		{fieldHostPathPath, field.StringOrEmpty(spec.Path)},
		{fieldHostPathMountPath, field.StringOrEmpty(spec.MountPath)},
		{fieldHostPathReadOnly, formatBoolYesNo(spec.ReadOnly)},
		{fieldHostPathMountPropagation, field.StringOrEmpty(spec.MountPropagation)},
	}

	displayIndentedFields(fields)
}

func (dd *DatasourceDisplayer) displayConfigMapSpec(spec *ConfigMapSpec) {
	fmt.Printf("%s%s\n", sectionsSeparator, titleConfigMapSpec) //nolint:forbidigo // CLI output to user

	fields := []DisplayField{
		{fieldConfigMapName, field.StringOrEmpty(spec.ConfigMapName)},
		{fieldConfigMapMountPath, field.StringOrEmpty(spec.MountPath)},
		{fieldConfigMapSubPath, field.StringOrEmpty(spec.SubPath)},
		{fieldConfigMapDefaultMode, field.StringOrEmpty(spec.DefaultMode)},
	}

	displayIndentedFields(fields)
}

func (dd *DatasourceDisplayer) displaySecretSpec(spec *SecretSpec) {
	fmt.Printf("%s%s\n", sectionsSeparator, titleSecretSpec) //nolint:forbidigo // CLI output to user

	fields := []DisplayField{
		{fieldSecretCredential, field.StringOrEmpty(spec.CredentialAssetID)},
		{fieldSecretMountPath, field.StringOrEmpty(spec.MountPath)},
		{fieldSecretDefaultMode, field.StringOrEmpty(spec.DefaultMode)},
	}

	displayIndentedFields(fields)
}

func (dd *DatasourceDisplayer) displayAdditionalAttributes(attrs map[string]string) {
	if len(attrs) == 0 {
		return
	}

	fmt.Printf("%s%s\n", sectionsSeparator, titleAdditionalAttr) //nolint:forbidigo // CLI output to user

	var fields []DisplayField
	for key, value := range attrs {
		fields = append(fields, DisplayField{key, value})
	}

	displayIndentedFields(fields)
}

func displayFields(fields []DisplayField) {
	for _, field := range fields {
		if value := fmt.Sprintf("%v", field.Value); value != "" && value != "<nil>" {
			fmt.Printf("%-*s %s\n", generalColumnPadding, field.Name+":", value) //nolint:forbidigo // CLI output to user
		}
	}
}

func displayIndentedFields(fields []DisplayField) {
	for _, field := range fields {
		if value := fmt.Sprintf("%v", field.Value); value != "" {
			fmt.Printf("  %-*s %s\n", specSectionColumnPadding, field.Name+":", value) //nolint:forbidigo // CLI output to user
		}
	}
}

func formatDatasourceType(dsType string) string {
	if displayType, ok := datasourceTypeMap[dsType]; ok {
		return displayType
	}
	return ""
}

func formatScope(scope string) string {
	return field.CapitalizeFirst(scope)
}

func formatBoolYesNo(b *bool) string {
	if b == nil {
		return ""
	}
	if *b {
		return "Yes"
	}
	return "No"
}

func RenderOutput(datasource *Datasource, outputFormat string) error {
	displayer := NewDatasourceDisplayer(datasource)

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
