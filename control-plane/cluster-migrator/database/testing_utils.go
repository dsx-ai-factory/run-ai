/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package database

import (
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	migrator_config "github.com/dsx-ai-factory/run-ai/control-plane/cluster-migrator/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-migrator/database/models"
	clusterServiceModels "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"k8s.io/utils/ptr"
)

const (
	TestTenantsCount             = 4
	TestClustersCountPerTenant   = 2
	TestNodepoolsCountPerCluster = 2
	TestNodepoolsCountPerTenant  = TestClustersCountPerTenant * TestNodepoolsCountPerCluster
	TestUser1                    = "user1@run.ai"
	TestUser2                    = "user2@run.ai"

	TestWeightsSetting  = false
	TestCpuQuotaSetting = true
)

var MapTenantIdToClusterIds = map[int32][]string{}
var npId = int32(0)
var schemaNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]{0,62}$`)

func SetupTables(conf *migrator_config.ConfigData, newMigration bool) {
	if !schemaNamePattern.MatchString(conf.SchemaSource) {
		logging.Errorf("Invalid source schema name: %q", conf.SchemaSource)
		return
	}
	quotedSchemaSource := pq.QuoteIdentifier(conf.SchemaSource)
	if _, err := targetClient.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", quotedSchemaSource)); err != nil {
		logging.Errorf("Failed to create schema. err: %v", err)
		return
	}

	if _, err := targetClient.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", models.TenantsSchema)); err != nil {
		logging.Errorf("Failed to create schema. err: %v", err)
		return
	}

	err := globalClient.Actual().Migrator().CreateTable(models.Tenant{})
	if err != nil {
		logging.Errorf("Failed to run auto-migrate. err: %v", err)
		return
	}

	err = sourceClient.Actual().Migrator().CreateTable(models.NodepoolsMigrationControl{})
	if err != nil {
		logging.Errorf("Failed to run auto-migrate. err: %v", err)
		return
	}

	err = sourceClient.Actual().Migrator().CreateTable(models.NodepoolV1{})
	if err != nil {
		logging.Errorf("Failed to run auto-migrate. err: %v", err)
		return
	}

	err = sourceClient.Actual().Migrator().CreateTable(models.TenantSetting{})
	if err != nil {
		logging.Errorf("Failed to run auto-migrate. err: %v", err)
		return
	}

	nodepools := []*models.NodepoolV1{}
	for i := 1; i <= TestTenantsCount; i++ {
		tenantId := int32(i)
		tenantClusters := make([]string, TestClustersCountPerTenant)
		if _, err := globalClient.Create(&models.Tenant{
			Id:   tenantId,
			Name: fmt.Sprintf("test-%d", i),
		}); err != nil {
			logging.Errorf("Failed to create tenant. err: %v", err)
			return
		}

		if _, err := sourceClient.Create(&models.TenantSetting{
			TenantId: tenantId,
			Key:      "project.quota.cpu_resources",
			Value:    fmt.Sprintf("%t", TestCpuQuotaSetting),
		}); err != nil {
			logging.Errorf("Failed to create tenant setting. err: %v", err)
			return
		}
		if _, err := sourceClient.Create(&models.TenantSetting{
			TenantId: tenantId,
			Key:      "project.quota.manual_over_quota",
			Value:    fmt.Sprintf("%t", TestWeightsSetting),
		}); err != nil {
			logging.Errorf("Failed to create tenant setting. err: %v", err)
			return
		}

		for j := int32(1); j <= TestClustersCountPerTenant; j++ {
			cluster := &clusterServiceModels.Cluster{
				Name:     uuid.NewString(),
				TenantId: tenantId,
				Domain:   "xyz",
				Version:  "xyz",
			}
			if _, err := targetClient.Create(cluster); err != nil {
				logging.Errorf("Failed to create cluster. err: %v", err)
				return
			}
			tenantClusters[j-1] = cluster.ID.String()

			nodepoolId := getNodepoolId()
			nodepools = append(nodepools, &models.NodepoolV1{
				Id:                   nodepoolId,
				Name:                 "default",
				LabelKey:             nil,
				LabelValue:           nil,
				Status:               "Ready",
				StatusMessage:        ptr.To("test status message - default"),
				ClusterId:            cluster.ID,
				CreatedAt:            time.Now().Add(-1 * time.Minute),
				UpdatedAt:            time.Now(),
				CreatedBy:            nil,
				UpdatedBy:            nil,
				IsDefault:            true,
				TenantId:             tenantId,
				GpuPlacementStrategy: "spread",
				CpuPlacementStrategy: "binpack",
			})

			for k := int32(1); k <= TestNodepoolsCountPerCluster-1; k++ {
				nodepoolId := getNodepoolId()
				nodepools = append(nodepools, &models.NodepoolV1{
					Id:                   nodepoolId,
					Name:                 fmt.Sprintf("nodepool-%d", nodepoolId),
					LabelKey:             ptr.To(fmt.Sprintf("label-key-%d", nodepoolId)),
					LabelValue:           ptr.To(fmt.Sprintf("label-value-%d", nodepoolId)),
					Status:               "Ready",
					StatusMessage:        ptr.To("test status message"),
					ClusterId:            cluster.ID,
					CreatedAt:            time.Now().Add(-1 * time.Minute),
					UpdatedAt:            time.Now(),
					CreatedBy:            ptr.To(TestUser1),
					UpdatedBy:            ptr.To(TestUser2),
					IsDefault:            false,
					TenantId:             tenantId,
					GpuPlacementStrategy: "binpack",
					CpuPlacementStrategy: "spread",
				})
			}
		}

		MapTenantIdToClusterIds[tenantId] = tenantClusters
	}

	if !newMigration {
		if _, err := sourceClient.Create(&nodepools); err != nil {
			logging.Errorf("Failed to create nodepools. err: %v", err)
			return
		}
		return
	}

	nodepoolsV2 := []*clusterServiceModels.Nodepool{}
	for _, np := range nodepools {
		mappedNp, err := mapNodepoolV1ToV2(*np)
		if err != nil {
			logging.Errorf("Failed to map nodepool v1 to v2. err: %v", err)
			return
		}
		nodepoolsV2 = append(nodepoolsV2, &mappedNp)
	}
	if _, err := targetClient.Create(&nodepoolsV2); err != nil {
		logging.Errorf("Failed to create nodepools. err: %v", err)
		return
	}
}

func GetTargetTenantNodepools(tenantId int32) ([]clusterServiceModels.Nodepool, error) {
	nodepools := []clusterServiceModels.Nodepool{}
	err := targetClient.Where("tenant_id = ?", tenantId).Find(&nodepools)
	return nodepools, err
}

func getNodepoolId() int32 {
	npId += 1

	return npId
}

func mapNodepoolV1ToV2(nodepoolV1 models.NodepoolV1) (clusterServiceModels.Nodepool, error) {
	clusterName := fmt.Sprintf("cluster-%s", nodepoolV1.ClusterId.String())

	return clusterServiceModels.Nodepool{
		Common: clusterServiceModels.Common{
			CreatedAt: nodepoolV1.CreatedAt,
			UpdatedAt: nodepoolV1.UpdatedAt,
		},
		ID:                     nodepoolV1.Id,
		Name:                   nodepoolV1.Name,
		TenantId:               nodepoolV1.TenantId,
		LabelKey:               nodepoolV1.LabelKey,
		LabelValue:             nodepoolV1.LabelValue,
		Phase:                  clusterServiceModels.Nodepool{}.ParseNodepoolPhase(nodepoolV1.Status),
		PhaseMessage:           nodepoolV1.StatusMessage,
		ClusterId:              nodepoolV1.ClusterId,
		ClusterName:            clusterName,
		DeletedAt:              nodepoolV1.DeletedAt,
		CreatedBy:              convutils.Deref(nodepoolV1.CreatedBy, ""),
		UpdatedBy:              convutils.Deref(nodepoolV1.UpdatedBy, ""),
		DeletedBy:              nil,
		IsDefault:              nodepoolV1.IsDefault,
		GpuPlacementStrategy:   clusterServiceModels.Nodepool{}.ParsePlacementStrategy(nodepoolV1.GpuPlacementStrategy),
		CpuPlacementStrategy:   clusterServiceModels.Nodepool{}.ParsePlacementStrategy(nodepoolV1.CpuPlacementStrategy),
		AllowAdjustableWeights: clusterServiceModels.DefaultAllowAdjustableWeights,
		AllowQuota:             clusterServiceModels.DefaultAllowQuota,
		AllowRanks:             clusterServiceModels.DefaultAllowRanks,
		AllowCpuQuota:          clusterServiceModels.DefaultAllowCpuQuota,
	}, nil
}
