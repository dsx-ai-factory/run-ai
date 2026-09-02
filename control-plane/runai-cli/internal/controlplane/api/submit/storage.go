/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/specset/storage/cmv"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/specset/storage/datavolume"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/specset/storage/emptydirvolume"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/specset/storage/gitsync"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/specset/storage/hostpath"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/specset/storage/nfs"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/specset/storage/pvc"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/specset/storage/s3"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/specset/storage/secretvolume"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/spec/spectype"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/spf13/pflag"
)

type Storage struct {
	ConfigMapVolume []string
	ExistingPvc     []string
	NewPvc          []string
	GitSync         []string
	Nfs             []string
	S3              []string
	HostPath        []string
	Volume          []string
	// this is for backward compatibility, should be removed once we deprecated the old pvc format completely
	BCStylePvc     []string
	MasterNoPVC    bool
	SecretVolume   []string
	EmptyDirVolume []string
	DataVolume     []string
	// ResolvedDataVolumes holds the parsed and resolved data volumes (name resolved to ID)
	ResolvedDataVolumes []openapi.DataVolumeInstance

	// Distributed Inference (Worker). The leader/worker spec schema only supports
	// configMapVolume, emptyDirVolume, pvc, secretVolume - so we expose only those flags.
	WorkerExistingPvc        []string
	WorkerNewPvc             []string
	WorkerSecretVolume       []string
	WorkerConfigmapMapVolume []string
	WorkerEmptyDirVolume     []string
}

func (s *Storage) GenerateStorageSpec() (*openapi.WorkspaceSpecSpecAllOfStorage, error) {
	storage := openapi.NewWorkspaceSpecSpecAllOfStorage()
	fieldsSet := false

	err := s.setConfigMapVolumeList(storage)
	if err != nil {
		return nil, err
	}
	fieldsSet = fieldsSet || len(storage.ConfigMapVolume) > 0

	err = s.setPvcList(storage)
	if err != nil {
		return nil, err
	}
	fieldsSet = fieldsSet || len(storage.Pvc) > 0

	err = s.setGitSyncList(storage)
	if err != nil {
		return nil, err
	}
	fieldsSet = fieldsSet || len(storage.Git) > 0

	err = s.setNfsList(storage)
	if err != nil {
		return nil, err
	}
	fieldsSet = fieldsSet || len(storage.Nfs) > 0

	err = s.setS3List(storage)
	if err != nil {
		return nil, err
	}
	fieldsSet = fieldsSet || len(storage.S3) > 0

	err = s.setHostPathList(storage)
	if err != nil {
		return nil, err
	}
	fieldsSet = fieldsSet || len(storage.HostPath) > 0

	err = s.setSecretVolumeList(storage)
	if err != nil {
		return nil, err
	}
	fieldsSet = fieldsSet || len(storage.SecretVolume) > 0

	err = s.setEmptyDirVolumeList(storage)
	if err != nil {
		return nil, err
	}
	fieldsSet = fieldsSet || len(storage.EmptyDirVolume) > 0

	s.setDataVolumeList(storage)
	fieldsSet = fieldsSet || len(storage.DataVolume) > 0

	if !fieldsSet {
		return nil, nil
	}

	return storage, nil
}

func (s *Storage) setConfigMapVolumeList(storage *openapi.WorkspaceSpecSpecAllOfStorage) error {
	if len(s.ConfigMapVolume) > 0 {
		cmvObjList, err := cmv.NewConfigMapVolumeList(s.ConfigMapVolume)
		if err != nil {
			return err
		}
		storage.SetConfigMapVolume(cmvObjList)
	}
	return nil
}

func (s *Storage) setPvcList(storage *openapi.WorkspaceSpecSpecAllOfStorage) error {
	if len(s.NewPvc) > 0 {
		newPvcList, err := pvc.NewPvcList(s.NewPvc, false)
		if err != nil {
			return err
		}
		storage.Pvc = append(storage.Pvc, newPvcList...)
	}
	if len(s.ExistingPvc) > 0 {
		existPvcList, err := pvc.NewPvcList(s.ExistingPvc, true)
		if err != nil {
			return err
		}
		storage.Pvc = append(storage.Pvc, existPvcList...)
	}
	// this is for backward compatibility, should be removed once we deprecated the old pvc format completely
	if len(s.BCStylePvc) > 0 {
		bcPvcList, err := pvc.ListDeprecatedOldStyle(s.BCStylePvc)
		if err != nil {
			return err
		}
		storage.Pvc = append(storage.Pvc, bcPvcList...)
	}
	return nil
}

func (s *Storage) setGitSyncList(storage *openapi.WorkspaceSpecSpecAllOfStorage) error {
	if len(s.GitSync) > 0 {
		gitSyncList, err := gitsync.NewGitSyncList(s.GitSync)
		if err != nil {
			return err
		}
		storage.SetGit(gitSyncList)
	}
	return nil
}

func (s *Storage) setNfsList(storage *openapi.WorkspaceSpecSpecAllOfStorage) error {
	if len(s.Nfs) > 0 {
		nfsList, err := nfs.NewNfsList(s.Nfs)
		if err != nil {
			return err
		}
		storage.SetNfs(nfsList)
	}
	return nil
}

func (s *Storage) setS3List(storage *openapi.WorkspaceSpecSpecAllOfStorage) error {
	if len(s.S3) > 0 {
		s3List, err := s3.NewS3List(s.S3)
		if err != nil {
			return err
		}
		storage.SetS3(s3List)
	}
	return nil
}
func (s *Storage) setSecretVolumeList(storage *openapi.WorkspaceSpecSpecAllOfStorage) error {
	if len(s.SecretVolume) > 0 {
		SecretVolume, err := secretvolume.NewList(s.SecretVolume)
		if err != nil {
			return err
		}
		storage.SetSecretVolume(SecretVolume)
	}
	return nil
}

func (s *Storage) setHostPathList(storage *openapi.WorkspaceSpecSpecAllOfStorage) error {
	if len(s.HostPath) > 0 {
		hostPathList, err := hostpath.NewList(s.HostPath)
		if err != nil {
			return err
		}
		storage.SetHostPath(hostPathList)
	}
	return nil
}

func (s *Storage) setEmptyDirVolumeList(storage *openapi.WorkspaceSpecSpecAllOfStorage) error {
	if len(s.EmptyDirVolume) > 0 {
		emptyDirVolumeList, err := emptydirvolume.NewEmptyDirVolumeList(s.EmptyDirVolume)
		if err != nil {
			return err
		}
		storage.SetEmptyDirVolume(emptyDirVolumeList)
	}
	return nil
}

func (s *Storage) setDataVolumeList(storage *openapi.WorkspaceSpecSpecAllOfStorage) {
	if len(s.ResolvedDataVolumes) > 0 {
		storage.SetDataVolume(s.ResolvedDataVolumes)
	}
}

func (s *Storage) GetRawDataVolumes() []string {
	return s.DataVolume
}

func (s *Storage) SetResolvedDataVolumes(resolved []openapi.DataVolumeInstance) {
	s.ResolvedDataVolumes = resolved
}

func (s *Storage) ParseDataVolumes() ([]datavolume.ParsedDataVolume, error) {
	return datavolume.ParseDataVolumeList(s.DataVolume)
}

func StorageFlagSet(flagsStorage *pflag.FlagSet, specStorage *Storage, getType spectype.Workload, framework api.Framework) {
	// Distributed Inference only supports a restricted set of storage types: configMapVolume,
	// emptyDirVolume, pvc, secretVolume (see DistributedInferenceLeaderWorkerSpec.yaml). Flags
	// for unsupported types (hostPath, nfs, git-sync, s3) are intentionally not registered to
	// avoid the CLI silently dropping them at the JSON-marshal step into the leader/worker spec.
	if getType != spectype.Submit {
		// new style (like `runai workspace submit ...`)
		flagsStorage.StringArrayVar(&specStorage.ConfigMapVolume, consts.AttributeNameConfigMapVolume, []string{}, consts.AttributeUsageConfigMapVolume)
		if getType != spectype.InferenceDistributed {
			flagsStorage.StringArrayVar(&specStorage.HostPath, consts.AttributeNameHostPath, []string{}, consts.AttributeUsageHostPath)
			flagsStorage.StringArrayVar(&specStorage.Nfs, consts.AttributeNameNfs, []string{}, consts.AttributeUsageNfs)
		}
	} else {
		// backward compatibility (like `runai submit ... `)
		flagsStorage.StringArrayVar(&specStorage.ConfigMapVolume, consts.AttributeBCNameConfigMapVolume, []string{}, consts.AttributeUsageConfigMapVolume)
		flagsStorage.StringArrayVarP(&specStorage.Volume, consts.AttributeBCNameVolume, "v", []string{}, consts.AttributeUsageHostPath)
	}

	flagsStorage.StringArrayVar(&specStorage.ExistingPvc, consts.AttributeNameExistingPvc, []string{}, consts.AttributeUsageExistingPvc)
	flagsStorage.StringArrayVar(&specStorage.NewPvc, consts.AttributeNameNewPvc, []string{}, consts.AttributeUsageNewPvc)
	if getType != spectype.InferenceDistributed {
		flagsStorage.StringArrayVar(&specStorage.GitSync, consts.AttributeNameGitSync, []string{}, consts.AttributeUsageGitSync)
	}
	flagsStorage.StringArrayVar(&specStorage.SecretVolume, consts.AttributeNameSecretVolume, []string{}, consts.AttributeUsageSecretVolume)
	flagsStorage.StringArrayVar(&specStorage.EmptyDirVolume, consts.AttributeNameEmptyDirVolume, []string{}, consts.AttributeUsageEmptyDirVolume)
	flagsStorage.StringArrayVar(&specStorage.DataVolume, consts.AttributeNameDataVolume, []string{}, consts.AttributeUsageDataVolume)

	if getType != spectype.Inference && getType != spectype.InferenceDistributed {
		flagsStorage.StringArrayVar(&specStorage.S3, consts.AttributeNameS3, []string{}, consts.AttributeUsageS3)
	}

	if getType == spectype.InferenceDistributed {
		// Distributed Inference worker-specific storage flags
		DistributedInferenceStorageFlagSet(flagsStorage, specStorage)
	}

	if getType == spectype.TrainingDistributed && framework != api.Jax {
		// Distributed Training master-specific storage flags
		flagsStorage.BoolVar(&specStorage.MasterNoPVC, consts.AttributeNameMasterNoPvc, false, consts.AttributeUsageMasterNoPvc)
	}
}
