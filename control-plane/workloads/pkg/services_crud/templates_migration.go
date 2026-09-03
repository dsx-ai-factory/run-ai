/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"fmt"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	clients "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/backend"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao_crud"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	"github.com/modern-go/reflect2"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/assets"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	assetsApi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

var templates TemplatesService

// the value that we store in the migration info db after completing asset-to-flat templates
// migrations. we prefer it over boolean because numeric value can serve as revision indicator
// in the future.
const assetsToFlatMigrationIndicator = 1

// to prevent API + db access we hold local set of tenants for which templates migration completed
// it will be reset upon service startup
var assetsMigratedToFlat = datastructs.NewSet[int32]()

func migrateAssetTemplatesIfNeeded(ctx context.Context, backendClient clients.TenantSettingsClient) {
	tenantId := auth.TenantIdOf(ctx)

	// already migrated
	if assetsMigratedToFlat.Has(tenantId) {
		return
	}

	migrationInfo, dbErr := dao_crud.GetMigrationsInfo(ctx)
	if dbErr != nil {
		return
	}

	if migrationInfo.AssetsToFlatTemplates == assetsToFlatMigrationIndicator {
		assetsMigratedToFlat.Add(tenantId)
		return
	}

	// we need backend client. the only situation that we might not have it is during testing
	if reflect2.IsNil(backendClient) {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("cannot validate tenant hybrid submission enabled flag: backend client is not initialized")
		return
	}

	hybridEnabled, clientErr := backendClient.GetHybridSubmissionEnabled(ctx)
	if clientErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("cannot validate tenant hybrid submission enabled flag: %v", clientErr)
		return
	}

	if !hybridEnabled {
		return
	}

	migrateErr := MigrateTemplates(ctx)
	if migrateErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to migrate templates: %s: %v", migrateErr.Message, migrateErr.Details)
	} else {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("successfuly migrated templates")
	}

	dbErr = dao_crud.SetAssetsToFlatMigration(ctx, assetsToFlatMigrationIndicator)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to update template migration flat: %s: %v", dbErr.Message, dbErr.Details)
	}

	assetsMigratedToFlat.Add(tenantId)
}

// MigrateTemplates request assets service for a list of old templates and store them in our DB
// as flat templates
func MigrateTemplates(ctx context.Context) *api.Error {
	templates, err := assets.Get().ListTemplatesOfTenant(ctx)
	if err != nil {
		return api.NewError(int(err.Code), err.Message, fmt.Errorf("list templates failed: %v", err.Details))
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("template migrator found %d templates of tenant", len(templates.Entries))

	if len(templates.Entries) == 0 {
		return nil
	}

	migratedToWorkspace := 0
	migratedToTraining := 0

	for _, template := range templates.Entries {
		if !convutils.HasValue(template.Meta.WorkloadSupportedTypes) || template.Meta.WorkloadSupportedTypes.Get().GetWorkspace() {
			err := migrateTemplateForType(ctx, &template, api.WORKLOADTYPE_WORKSPACE)
			if err != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("template migrator failed to migrate %s template %s [%s] for workspaces: %v",
					template.Meta.Scope, template.Meta.Id, template.Meta.Name, err.Details)
			} else {
				migratedToWorkspace += 1
			}
		}

		if !convutils.HasValue(template.Meta.WorkloadSupportedTypes) || template.Meta.WorkloadSupportedTypes.Get().GetTraining() {
			err := migrateTemplateForType(ctx, &template, api.WORKLOADTYPE_TRAINING)
			if err != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("template migrator failed to migrate %s template %s [%s] for training: %v",
					template.Meta.Scope, template.Meta.Id, template.Meta.Name, err.Details)
			} else {
				migratedToTraining += 1
			}
		}
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("template migrator created %d workspace and %d training templates", migratedToWorkspace, migratedToTraining)

	return nil
}

func migrateTemplateForType(ctx context.Context, template *assetsApi.WorkloadTemplate, workloadType api.WorkloadType) *api.Error {
	migratedTemplate := api.SupersetTemplate{
		Meta: api.TemplateMetaFields{
			Name:        fmt.Sprintf("%s-%s", template.Meta.Name, strings.ToLower(string(workloadType))),
			Description: template.Meta.Description.Get(),
			ScopeType:   api.ScopeType(template.Meta.Scope),
		},
		Assets: &api.TemplateAssetsIdsOptional{
			Environment: convutils.Ref(template.Spec.Assets.Environment.Id),
		},
	}

	//
	//   add scope id whenever relevant
	//
	switch template.Meta.Scope {
	case assetsApi.SCOPE_CLUSTER:
		migratedTemplate.Meta.ScopeId = convutils.Deref(template.Meta.ClusterId.Get(), "")
	case assetsApi.SCOPE_DEPARTMENT:
		migratedTemplate.Meta.ScopeId = convutils.Deref(template.Meta.DepartmentId.Get(), "")
	case assetsApi.SCOPE_PROJECT:
		migratedTemplate.Meta.ScopeId = fmt.Sprintf("%d", convutils.Deref(template.Meta.ProjectId.Get(), 0))
	}

	if convutils.HasValue(template.Spec.Assets.Compute) {
		migratedTemplate.Assets.Compute = convutils.Ref(template.Spec.Assets.Compute.Get().Id)
	}

	if template.Spec.Assets.Datasources != nil {
		for _, datasource := range template.Spec.Assets.Datasources {
			migratedTemplate.Assets.Datasources = append(migratedTemplate.Assets.Datasources, datasource.Id)
		}
	}

	//
	//   we need to make sure that the template creator is the original user of the template
	//   and not the one triggering the migration
	//
	migrationContext := context.WithValue(ctx, runaictx.CfSubjectID, template.Meta.CreatedBy)

	request, err := NewUnifiedRequest(migrationContext, workloadType, migratedTemplate)
	if err != nil {
		return err
	}

	//
	//   this is an indication that template creation is done as part of migrating process so we need to bypass
	//   the authorization check (user who run 'list' does not necessarily have write access to all the scopes
	//   that templates will be migrated to, so we bypass the authorization for migration process)
	//
	request.migrating = true

	//
	//   add the template to the DB
	//
	err = templates.CreateTemplate(ctx, request)
	if err != nil {
		return err
	}

	return nil
}
