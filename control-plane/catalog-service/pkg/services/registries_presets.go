package services

import (
	"context"

	"github.com/google/uuid"
	authorization_api "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	authorization_utils "github.com/run-ai/runai/control-plane/authorization/pkg/client/helpers"
	"github.com/run-ai/runai/control-plane/catalog-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/catalog-service/pkg/constants"
	"github.com/run-ai/runai/control-plane/catalog-service/pkg/database/dao"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

const (
	createdByRunai  = "NVIDIA Run:ai"
	systemSubjectID = "system"
)

var (
	systemGroups          = []string{"System"}
	systemPermittedScopes = &authorization_api.PermittedScopes{
		System:  convutils.Ref(true),
		Tenants: []string{"0"},
	}
)

var ngcrPreset = api.RegistryIntegrationCreationRequest{
	ScopeType:   api.SCOPETYPE_SYSTEM,
	Name:        constants.InternalNgcIntegrationName,
	Url:         constants.NgcImagesBaseURL,
	Description: "NVIDIA GPU Cloud Registry",
	Version:     "1.0.0",
}

var presets = map[string]api.RegistryIntegrationCreationRequest{
	ngcrPreset.Name: ngcrPreset,
}

func InitRegistriesPresets() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, systemSubjectID)
	ctx = context.WithValue(ctx, runaictx.CfGroups, systemGroups)
	ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, systemPermittedScopes)

	systemRegistries, dbErr := dao.ListPresetRegistries(ctx)
	if dbErr != nil {
		logging.Errorf("error listing registries: %v", dbErr)
		return
	}

	rs := registriesAPIService{}

	for name, preset := range presets {
		existingRegistry := systemRegistries[name]
		if existingRegistry == nil {
			rs.createPreset(ctx, name, preset)
			continue
		}

		if existingRegistry.Version == nil || *existingRegistry.Version != preset.Version {
			rs.updatePreset(ctx, existingRegistry.ID, name, preset)
		}
	}
}

func (rs *registriesAPIService) createPreset(ctx context.Context, name string, preset api.RegistryIntegrationCreationRequest) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("creating preset registry: %s", name)

	_, err := rs.createRegistry(ctx, preset, true)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("error creating preset registry: %v", err)
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("created preset registry: %s", name)
}

func (rs *registriesAPIService) updatePreset(ctx context.Context, id uuid.UUID, name string, preset api.RegistryIntegrationCreationRequest) {
	rec, err := dao.GetRegistryById(ctx, id.String())
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("error getting preset registry: %v", err)
		return
	}

	oldVersion := "0.0.0"
	if rec.Version != nil {
		oldVersion = *rec.Version
	}

	rec.Name = name
	rec.Url = preset.Url
	rec.Description = &preset.Description
	rec.Version = &preset.Version

	apiErr := dao.UpdateRegistry(ctx, rec)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("error updating preset registry: %v", apiErr)
		return
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("updated preset registry %s from version %s to %s", id, oldVersion, preset.Version)
}
