/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package displayer

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/ui/field"
)

// Type definitions
type Storage struct {
	ConfigMaps []StorageVolume `json:"configMaps,omitempty"`
	Secrets    []StorageVolume `json:"secrets,omitempty"`
	PVCs       []StorageVolume `json:"pvcs,omitempty"`
	EmptyDirs  []StorageVolume `json:"emptyDirs,omitempty"`
	Gits       []StorageVolume `json:"gits,omitempty"`
	HostPaths  []StorageVolume `json:"hostPaths,omitempty"`
	S3s        []StorageVolume `json:"s3s,omitempty"`
}

type StorageVolume struct {
	Name         string  `json:"name"`
	MountPath    string  `json:"mountPath"`
	Size         string  `json:"size,omitempty"`
	ReadOnly     *bool   `json:"readOnly,omitempty"`
	Ephemeral    *bool   `json:"ephemeral,omitempty"`
	DataSharing  *bool   `json:"dataSharing,omitempty"`
	StorageClass *string `json:"storageClass,omitempty"`
}

// Displayer
type StorageDisplayer struct {
	storage *Storage
}

func NewStorageSection(storage *Storage) *Section {
	return &Section{
		Displayer: &StorageDisplayer{
			storage: storage,
		},
	}
}

func (s *StorageDisplayer) Subject() string {
	return TitleStorage
}

func (s *StorageDisplayer) Columns() []table.Column {
	return []table.Column{
		{Title: "Type", Width: StorageTypeColumnWidth},
		{Title: FieldName, Width: StorageNameColumnWidth},
		{Title: "Mount Path", Width: StorageMountPathColumnWidth},
		{Title: "Size", Width: StorageSizeColumnWidth},
		{Title: "ReadOnly", Width: StorageReadOnlyColumnWidth},
		{Title: "Ephemeral", Width: StorageEphemeralColumnWidth},
		{Title: "Data Sharing", Width: StorageDataSharingColumnWidth},
		{Title: "Storage Class", Width: StorageStorageClassColumnWidth},
	}
}

func (s *StorageDisplayer) Data() [][]string {
	var data [][]string

	if s.storage == nil {
		return data
	}

	for _, pvc := range s.storage.PVCs {
		data = append(data, []string{
			"PVC",
			pvc.Name,
			pvc.MountPath,
			field.StringOrDash(pvc.Size),
			field.BoolToYesNo(pvc.ReadOnly),
			field.BoolToYesNo(pvc.Ephemeral),
			field.BoolToEnabled(pvc.DataSharing),
			field.PtrStringOrDash(pvc.StorageClass),
		})
	}
	for _, cm := range s.storage.ConfigMaps {
		data = append(data, []string{"ConfigMap", cm.Name, cm.MountPath, "-", "-", "-", "-", "-"})
	}
	for _, secret := range s.storage.Secrets {
		data = append(data, []string{"Secret", secret.Name, secret.MountPath, "-", "-", "-", "-", "-"})
	}
	for _, ed := range s.storage.EmptyDirs {
		data = append(data, []string{
			"EmptyDir",
			ed.Name,
			ed.MountPath,
			field.StringOrDash(ed.Size),
			"-",
			field.BoolToYesNo(ed.Ephemeral),
			"-",
			"-",
		})
	}
	for _, git := range s.storage.Gits {
		data = append(data, []string{"Git", git.Name, git.MountPath, "-", "-", "-", "-", "-"})
	}
	for _, hp := range s.storage.HostPaths {
		data = append(data, []string{
			"HostPath",
			hp.Name,
			hp.MountPath,
			"-",
			field.BoolToYesNo(hp.ReadOnly),
			"-",
			"-",
			"-",
		})
	}
	for _, s3 := range s.storage.S3s {
		data = append(data, []string{"S3", s3.Name, s3.MountPath, "-", "-", "-", "-", "-"})
	}

	return data
}
