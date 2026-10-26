/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package database

import (
	"fmt"
	"regexp"
	"time"

	migrator_config "github.com/dsx-ai-factory/run-ai/control-plane/cluster-migrator/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-migrator/database/models"
	clusterServiceModels "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"k8s.io/utils/ptr"
)

const (
	TestTenantsCount             = 4
	TestClustersCountPerTenant   = 2
	TestNodepoolsCountPerCluster = 2
	TestNodepoolsCountPerTenant  = TestClustersCountPerTenant * TestNodepoolsCountPerCluster

	TestWeightsSetting  = false
	TestCpuQuotaSetting = true
)

var npId = int32(0)
var schemaNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]{0,62}$`)

func SetupTables(conf *migrator_config.ConfigData) {
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

	err = sourceClient.Actual().Migrator().CreateTable(models.TenantSetting{})
	if err != nil {
		logging.Errorf("Failed to run auto-migrate. err: %v", err)
		return
	}

	nodepools := []*clusterServiceModels.Nodepool{}
	for i := 1; i <= TestTenantsCount; i++ {
		tenantId := int32(i)
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

			for k := 0; k < TestNodepoolsCountPerCluster; k++ {
				nodepools = append(nodepools, newTestNodepool(cluster, tenantId, k == 0))
			}
		}
	}

	if _, err := targetClient.Create(&nodepools); err != nil {
		logging.Errorf("Failed to create nodepools. err: %v", err)
		return
	}
}

func GetTargetTenantNodepools(tenantId int32) ([]clusterServiceModels.Nodepool, error) {
	nodepools := []clusterServiceModels.Nodepool{}
	err := targetClient.Where("tenant_id = ?", tenantId).Find(&nodepools)
	return nodepools, err
}

func newTestNodepool(cluster *clusterServiceModels.Cluster, tenantId int32, isDefault bool) *clusterServiceModels.Nodepool {
	npId += 1
	nodepool := &clusterServiceModels.Nodepool{
		Common: clusterServiceModels.Common{
			CreatedAt: time.Now().Add(-1 * time.Minute),
			UpdatedAt: time.Now(),
		},
		ID:                     npId,
		Name:                   "default",
		TenantId:               tenantId,
		Phase:                  clusterServiceModels.Nodepool{}.ParseNodepoolPhase("Ready"),
		PhaseMessage:           ptr.To("test status message"),
		ClusterId:              cluster.ID,
		ClusterName:            fmt.Sprintf("cluster-%s", cluster.ID.String()),
		IsDefault:              isDefault,
		GpuPlacementStrategy:   clusterServiceModels.Nodepool{}.ParsePlacementStrategy("binpack"),
		CpuPlacementStrategy:   clusterServiceModels.Nodepool{}.ParsePlacementStrategy("binpack"),
		AllowAdjustableWeights: clusterServiceModels.DefaultAllowAdjustableWeights,
		AllowQuota:             clusterServiceModels.DefaultAllowQuota,
		AllowRanks:             clusterServiceModels.DefaultAllowRanks,
		AllowCpuQuota:          clusterServiceModels.DefaultAllowCpuQuota,
	}
	if !isDefault {
		nodepool.Name = fmt.Sprintf("nodepool-%d", npId)
		nodepool.LabelKey = ptr.To(fmt.Sprintf("label-key-%d", npId))
		nodepool.LabelValue = ptr.To(fmt.Sprintf("label-value-%d", npId))
	}
	return nodepool
}
