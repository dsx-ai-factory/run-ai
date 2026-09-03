/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package ri_migrations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	dao "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/go_migrations/ri/source"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/go_migrations/ri/source/kubernetes"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/go_migrations/ri/source/nvidia"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/go_migrations/ri/source/slinky"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	dao_utils "github.com/dsx-ai-factory/run-ai/runai-common-packages/dao-utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/google/uuid"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	RunaiSource = "Run.ai"
)

func getKartasToMigrate() []source.KartaConfiguration {
	return []source.KartaConfiguration{

		// Kubernetes Native
		kubernetes.GetK8sDeploymentRIConfiguration(),
		kubernetes.GetK8sJobRIConfiguration(),
		kubernetes.GetK8sStatefulSetRIConfiguration(),
		kubernetes.GetK8sReplicaSetRIConfiguration(),
		kubernetes.GetK8sPodRIConfiguration(),
		kubernetes.GetK8sCronJobRIConfiguration(),

		source.GetDynamoGraphDeploymentRIConfiguration(),
		source.GetKserveInferenceServiceRIConfiguration(),
		source.GetLeaderWorkerSetRIConfiguration(),
		source.GetSeldonDeploymentRIConfiguration(),
		source.GetKubevirtVMIRIConfiguration(),
		source.GetKNativeServiceRIConfiguration(),
		source.GetDevWorkspaceRIConfiguration(),
		source.GetTektonTaskRunRIConfiguration(),
		source.GetTektonPipelineRunRIConfiguration(),
		source.GetArgoWorkflowRIConfiguration(),
		source.GetAzureAmlJobRIConfiguration(),

		// NVIDIA
		nvidia.GetNIMCacheRIConfiguration(),
		nvidia.GetNIMServiceRIConfiguration(),
		nvidia.GetNVCFSPOTRequestRIConfiguration(),

		// Kubeflow
		source.GetKubeflowScheduledWorkflowRIConfiguration(),
		source.GetKubeflowTFJobRIConfiguration(),
		source.GetKubeflowPyTorchJobRIConfiguration(),
		source.GetKubeflowXGBoostJobRIConfiguration(),
		source.GetKubeflowJAXJobRIConfiguration(),
		source.GetKubeflowMPIJobRIConfiguration(),
		source.GetKubeflowNotebookRIConfiguration(),

		// Ray
		source.GetRayClusterRIConfiguration(),
		source.GetRayJobRIConfiguration(),
		source.GetRayServiceRIConfiguration(),

		// Slinky
		slinky.GetSlinkyNodeSetRIConfiguration(),
	}
}

// InitInternalKartas iterates internal karta Go structs and upserts them to the DB.
// Idempotent and safe to run at startup.
func InitInternalKartas(ctx context.Context) error {
	kartaConfigs := getKartasToMigrate()
	return initInternalKartas(ctx, kartaConfigs)
}

// another internal func - mainly for tests purposes
func initInternalKartas(ctx context.Context, kartaConfigs []source.KartaConfiguration) error {
	existingGroupKinds := map[string]bool{}
	tx := dao_utils.DB().Begin()
	if tx.Error != nil {
		logging.Errorf("RI migrations: failed to begin transaction: %v", tx.Error)
		return tx.Error
	}
	committed := false
	defer dao_utils.FinalizeTransaction(tx, &committed)

	for _, kartaConfig := range kartaConfigs {
		groupKind, err := validateKartaBeforeMigrate(kartaConfig)
		if err != nil {
			logging.Errorf("RI migrations: validation failed for Kartas: %v", err)
			return err
		}

		workloadType, apiErr := dao.DbUpsertWorkloadType(ctx, tx, workloadTypeCreateFields(groupKind.Kind, groupKind.Group))
		if apiErr != nil {
			logging.Errorf("RI migrations: upsert workload type for %s/%s failed: %v", groupKind.Group, groupKind.Kind, err)
			return errors.New(apiErr.Details)
		}

		groupKindStr := groupKind.String()
		if _, ok := existingGroupKinds[groupKindStr]; !ok {
			if err = upsertExternalWorkloadsTypeConfig(tx, workloadType.ID.String(), kartaConfig.Category, kartaConfig.Priority, string(kartaConfig.Preemptibility)); err != nil {
				logging.Errorf("RI migrations: upsert external workload type config for %s/%s failed: %v", groupKind.Group, groupKind.Kind, err)
				return err
			}
			existingGroupKinds[groupKindStr] = true
		}

		for _, karta := range kartaConfig.Kartas {
			gvk := karta.GroupVersionKind()
			if gvk == nil || gvk.Version == "" {
				logging.Warnf("RI migrations: skipping RI with invalid or missing version for %s/%s", groupKind.Group, groupKind.Kind)
				continue
			}
			if err = upsertKarta(tx, karta, workloadType.ID.String()); err != nil {
				logging.Errorf("RI migrations: upsert resourceInterface %s/%s/%s failed: %v", gvk.Group, gvk.Kind, gvk.Version, err)
				return err
			}
			logging.Debugf("RI migrations: ensured %s", gvk.String())
		}
	}

	result := tx.Commit()
	if result.Error != nil {
		logging.Errorf("Failed to commit RI migrations transaction: %v", result.Error)
		return result.Error
	}

	committed = true
	logging.Infof("RI migrations: completed successfully")

	return nil
}

func validateKartaBeforeMigrate(kartaConfig source.KartaConfiguration) (*schema.GroupKind, error) {
	if kartaConfig.Category == "" {
		return nil, fmt.Errorf("invalid KartaConfiguration: category must not be empty")
	}
	if kartaConfig.Priority == "" {
		return nil, fmt.Errorf("invalid KartaConfiguration: priority must not be empty")
	}
	if kartaConfig.Preemptibility == "" || !kartaConfig.Preemptibility.IsValid() {
		return nil, fmt.Errorf("invalid KartaConfiguration: preemptibility must be a valid value")
	}
	groupKind, err := validateGvks(kartaConfig.Kartas)
	if err != nil {
		logging.Errorf("RI migrations: validation failed for Kartas: %v", err)
		return nil, err
	}

	return groupKind, nil
}

func validateGvks(kartas []models.Karta) (*schema.GroupKind, error) {
	if len(kartas) == 0 {
		return nil, fmt.Errorf("no RIs provided for validation")
	}
	groupVersionKind := kartas[0].GroupVersionKind()
	if groupVersionKind == nil {
		return nil, fmt.Errorf("RI has invalid or missing GVK")
	}
	groupKind := groupVersionKind.GroupKind()

	versions := datastructs.Set[string]{}
	for _, karta := range kartas {
		gvk := karta.GroupVersionKind()
		// Note: Group can be empty for core Kubernetes resources (e.g., Pod)
		if gvk == nil || gvk.Kind == "" || gvk.Version == "" || (gvk.Group == "" && gvk.Kind != "Pod") {
			return nil, fmt.Errorf("RI has invalid or missing GVK fields: %v", gvk)
		}
		if gvk.Group != groupKind.Group || gvk.Kind != groupKind.Kind {
			return nil, fmt.Errorf("RI has mismatched group/kind: expected %s/%s but got %s/%s", groupKind.Group, groupKind.Kind, gvk.Group, gvk.Kind)
		}
		if versions.Has(gvk.Version) {
			return nil, fmt.Errorf("duplicate RI version found: %s", gvk.Version)
		}

		converted := kartav1alpha1.Karta(karta)
		validator := kartav1alpha1.NewKartaValidator(&converted)
		err := validator.Validate()
		if err != nil {
			logging.Errorf("RI validation failed for gvk <%s>: %v", gvk.String(), err)
			return nil, fmt.Errorf("RI validation for gvk <%s> failed: %v", gvk.String(), err)
		}

		versions.Add(gvk.Version)
	}
	return &groupKind, nil
}

func upsertExternalWorkloadsTypeConfig(tx *gorm.DB, workloadTypeId, categoryName, priorityName, preemptibility string) error {
	_, err := getExternalWorkloadsTypeConfig(tx, workloadTypeId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return createExternalWorkloadsTypeConfig(tx, workloadTypeId, categoryName, priorityName, preemptibility)
		}
		return err
	}

	return updateExternalWorkloadsTypeConfig(tx, workloadTypeId, categoryName, priorityName, preemptibility)
}

func getExternalWorkloadsTypeConfig(tx *gorm.DB, workloadTypeId string) (*models.ExternalWorkloadsTypeConfig, error) {
	var config models.ExternalWorkloadsTypeConfig
	result := tx.
		Where("workload_type_id = ?", workloadTypeId).
		Where("tenant_id IS NULL").
		Where("scope_type = ?", dao.SystemScopeType).
		First(&config)
	if result.Error != nil {
		return nil, result.Error
	}
	return &config, nil
}

func updateExternalWorkloadsTypeConfig(tx *gorm.DB, workloadTypeId, categoryName, priorityName, preemptibility string) error {
	sql := `
UPDATE external_workloads_type_config
SET category_id = (SELECT id FROM workloads_categories WHERE name = ? AND scope_type = ? LIMIT 1),
    priority_id = (SELECT id FROM priorities WHERE name = ? AND scope_type = ? LIMIT 1),
    preemptibility = ?,
    updated_by  = ?,
    updated_at  = now()
WHERE workload_type_id = ?
  AND scope_type = ?
  AND tenant_id IS NULL
  AND updated_by = ?
  AND (
    category_id IS DISTINCT FROM (SELECT id FROM workloads_categories WHERE name = ? AND scope_type = ? LIMIT 1)
    OR priority_id IS DISTINCT FROM (SELECT id FROM priorities WHERE name = ? AND scope_type = ? LIMIT 1)
  );
`
	result := tx.Exec(sql,
		// category - name and scope_type
		categoryName, dao.SystemScopeType,
		// priority - name and scope_type
		priorityName, dao.SystemScopeType,
		// preemptibility
		preemptibility,
		// updated_by
		RunaiSource,
		// workload_type_id
		workloadTypeId,
		// scope_type
		dao.SystemScopeType,
		// only if record is owned by Run.ai - was updated_by Run.ai
		RunaiSource,
		// change guard: only update if either category_id or priority_id differs
		categoryName, dao.SystemScopeType,
		priorityName, dao.SystemScopeType,
	)
	return result.Error
}

func createExternalWorkloadsTypeConfig(tx *gorm.DB, workloadTypeId, categoryName, priorityName, preemptibility string) error {
	sql := `
INSERT INTO external_workloads_type_config
  (id, tenant_id, scope_type, workload_type_id, category_id, priority_id, preemptibility, created_by, updated_by)
VALUES
  (gen_random_uuid(), NULL, ?, ?, 
   (SELECT id FROM workloads_categories WHERE name = ? AND scope_type = ? LIMIT 1),
   (SELECT id FROM priorities WHERE name = ? AND scope_type = ? LIMIT 1),
   ?, ?, ?);
`
	result := tx.Exec(sql,
		// scope_type
		dao.SystemScopeType,
		// workload_type_id
		workloadTypeId,
		// category - name and scope_type
		categoryName, dao.SystemScopeType,
		// priority - name and scope_type
		priorityName, dao.SystemScopeType,
		// preemptibility
		preemptibility,
		// created_by, updated_by
		RunaiSource, RunaiSource,
	)
	return result.Error
}

// upsertKarta - in system scope
func upsertKarta(tx *gorm.DB, karta models.Karta, workloadTypeId string) error {
	gvk := karta.GroupVersionKind()
	if gvk == nil {
		return fmt.Errorf("RI has no GVK")
	}

	riJSON, err := json.Marshal(karta)
	if err != nil {
		return err
	}

	sql := `
INSERT INTO workload_type_kartas
  (id, tenant_id, scope_type, workload_type_id, version, karta, created_by, created_at, updated_at)
SELECT
  ?, NULL, ?, ?, ?, ?::jsonb, ?, ?, ?
WHERE NOT EXISTS (
  SELECT 1
  FROM workload_type_kartas
  WHERE workload_type_id = ?
    AND scope_type = ?
    AND tenant_id IS NULL
    AND version = ?
    AND created_by <> ?
)
ON CONFLICT (workload_type_id, version, scope_type) WHERE tenant_id IS NULL
DO UPDATE SET
  karta = EXCLUDED.karta,
   updated_at = EXCLUDED.updated_at
   WHERE workload_type_kartas.created_by = ?
     AND workload_type_kartas.karta IS DISTINCT FROM EXCLUDED.karta;
`

	now := time.Now()
	result := tx.Exec(sql,
		// id
		uuid.New(),
		// scope_type
		dao.SystemScopeType,
		// workload_type_id
		workloadTypeId,
		// version
		gvk.Version,
		// resource_interface
		string(riJSON),
		// created_by
		RunaiSource,
		// created_at, updated_at
		now, now,
		// insert guard: block insert if there is any user-owned RI for this WT in system scope
		workloadTypeId,
		dao.SystemScopeType,
		gvk.Version,
		RunaiSource,
		// only update rows owned by Run.ai
		RunaiSource,
	)
	return result.Error
}

func workloadTypeCreateFields(name, group string) dao.WorkloadTypeCreateFieldsParsed {
	return dao.WorkloadTypeCreateFieldsParsed{
		Group: group,
		WorkloadTypeCreateUpdateFieldsParsed: dao.WorkloadTypeCreateUpdateFieldsParsed{
			WorkloadTypeName: name,
		},
	}
}
