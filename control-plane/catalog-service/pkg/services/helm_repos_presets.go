/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"

	"github.com/google/uuid"
	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

type helmRepoPreset struct {
	Request api.HelmRepoCreationRequest
	Version string
}

const helmRepoPresetVersion = "1.0.0"

func newHelmRepoPreset(name string, url string, description string, isPrivate bool) helmRepoPreset {
	return helmRepoPreset{
		Request: api.HelmRepoCreationRequest{
			ScopeType:   api.SCOPETYPE_SYSTEM,
			Name:        name,
			Url:         url,
			Description: description,
			IsPrivate:   isPrivate,
		},
		Version: helmRepoPresetVersion,
	}
}

var nvidiaPreset = newHelmRepoPreset(
	"NVIDIA",
	"https://helm.ngc.nvidia.com/nvidia",
	"NVIDIA Helm Repository",
	false,
)

var blueprintsPreset = newHelmRepoPreset(
	"Nvidia Blueprints",
	"https://helm.ngc.nvidia.com/nvidia/blueprint",
	"NVIDIA Blueprints Helm Repository",
	false,
)

var claraPreset = newHelmRepoPreset(
	"Clara",
	"https://helm.ngc.nvidia.com/nvidia/clara",
	"NVIDIA Clara Helm Repository",
	false,
)

var docaPreset = newHelmRepoPreset(
	"DOCA",
	"https://helm.ngc.nvidia.com/nvidia/doca",
	"NVIDIA DOCA Helm Repository",
	false,
)

var omniversePreset = newHelmRepoPreset(
	"Omniverse",
	"https://helm.ngc.nvidia.com/nvidia/omniverse",
	"NVIDIA Omniverse Helm Repository",
	false,
)

var taoPreset = newHelmRepoPreset(
	"TAO",
	"https://helm.ngc.nvidia.com/nvidia/tao",
	"NVIDIA TAO Helm Repository",
	false,
)

var nemoPreset = newHelmRepoPreset(
	"Nemo Microservices",
	"https://helm.ngc.nvidia.com/nvidia/nemo-microservices",
	"NVIDIA Nemo Microservices Helm Repository",
	true,
)

var holoscanForMediaPreset = newHelmRepoPreset(
	"Holoscan for Media",
	"https://helm.ngc.nvidia.com/nvidia/holoscan-for-media",
	"NVIDIA Holoscan for Media Helm Repository",
	true,
)

var isaacPreset = newHelmRepoPreset(
	"Isaac",
	"https://helm.ngc.nvidia.com/nvidia/isaac",
	"NVIDIA Isaac Helm Repository",
	true,
)

var morpheusPreset = newHelmRepoPreset(
	"Morpheus",
	"https://helm.ngc.nvidia.com/nvidia/morpheus",
	"NVIDIA Morpheus Helm Repository",
	true,
)

var deepstreamPreset = newHelmRepoPreset(
	"DeepStream",
	"https://helm.ngc.nvidia.com/nvidia/deepstream",
	"NVIDIA DeepStream Helm Repository",
	true,
)

var usdsearchPreset = newHelmRepoPreset(
	"USD Search",
	"https://helm.ngc.nvidia.com/nvidia/usdsearch",
	"NVIDIA USD Search Helm Repository",
	true,
)

var rivaPreset = newHelmRepoPreset(
	"Riva",
	"https://helm.ngc.nvidia.com/nvidia/riva",
	"NVIDIA Riva Helm Repository",
	true,
)

var nimPreset = newHelmRepoPreset(
	"NIM",
	"https://helm.ngc.nvidia.com/nim",
	"NVIDIA NIM Helm Repository",
	true,
)

var nvaiePreset = newHelmRepoPreset(
	"NVAIE",
	"https://helm.ngc.nvidia.com/nvaie",
	"NVIDIA AI Enterprise Helm Repository",
	true,
)

var helmRepoPresets = map[string]helmRepoPreset{
	nvidiaPreset.Request.Name:           nvidiaPreset,
	blueprintsPreset.Request.Name:       blueprintsPreset,
	claraPreset.Request.Name:            claraPreset,
	docaPreset.Request.Name:             docaPreset,
	omniversePreset.Request.Name:        omniversePreset,
	taoPreset.Request.Name:              taoPreset,
	nemoPreset.Request.Name:             nemoPreset,
	holoscanForMediaPreset.Request.Name: holoscanForMediaPreset,
	isaacPreset.Request.Name:            isaacPreset,
	morpheusPreset.Request.Name:         morpheusPreset,
	deepstreamPreset.Request.Name:       deepstreamPreset,
	usdsearchPreset.Request.Name:        usdsearchPreset,
	rivaPreset.Request.Name:             rivaPreset,
	nimPreset.Request.Name:              nimPreset,
	nvaiePreset.Request.Name:            nvaiePreset,
}

func InitHelmRepoPresets() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, systemSubjectID)
	ctx = context.WithValue(ctx, runaictx.CfGroups, systemGroups)
	ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
		System:  convutils.Ref(true),
		Tenants: []string{"0"},
	})

	existingHelmRepos, dbErr := dao.ListPresetHelmRepos(ctx)
	if dbErr != nil {
		logging.Errorf("error listing preset helm repos: %v", dbErr)
		return
	}

	hs := helmReposAPIService{}

	for name, preset := range helmRepoPresets {
		existing := existingHelmRepos[name]
		if existing == nil {
			hs.createHelmRepoPreset(ctx, name, preset)
			continue
		}

		if existing.Version == nil || *existing.Version != preset.Version {
			hs.updateHelmRepoPreset(ctx, existing.ID, name, preset)
		}
	}
}

func (s *helmReposAPIService) createHelmRepoPreset(ctx context.Context, name string, preset helmRepoPreset) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("creating preset helm repo: %s", name)

	_, err := s.createHelmRepo(ctx, preset.Request, true)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("error creating preset helm repo: %v", err)
		return
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("created preset helm repo: %s", name)
}

func (s *helmReposAPIService) updateHelmRepoPreset(ctx context.Context, id uuid.UUID, name string, preset helmRepoPreset) {
	rec, err := dao.GetHelmRepoById(ctx, id.String())
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("error getting preset helm repo: %v", err)
		return
	}

	oldVersion := "0.0.0"
	if rec.Version != nil {
		oldVersion = *rec.Version
	}

	rec.Name = name
	rec.Url = preset.Request.Url
	rec.IsPrivate = preset.Request.IsPrivate
	rec.Description = &preset.Request.Description
	rec.Version = &preset.Version

	apiErr := dao.UpdateHelmRepo(ctx, rec)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("error updating preset helm repo: %v", apiErr)
		return
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("updated preset helm repo %s from version %s to %s", id, oldVersion, preset.Version)
}
