package migration

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	migrator_config "github.com/run-ai/runai/control-plane/cluster-migrator/config"
	"github.com/run-ai/runai/control-plane/cluster-migrator/database"
	"github.com/run-ai/runai/control-plane/cluster-migrator/database/models"
	clusterServiceModels "github.com/run-ai/runai/control-plane/cluster-service/pkg/database/models"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

func Migrate(conf *migrator_config.ConfigData) bool {
	tenants := database.GetTenantsToMigrate(conf.TenantsToMigrate, conf.TenantsToNotMigrate)

	return migrateTenants(conf, tenants)
}

func migrateTenants(conf *migrator_config.ConfigData, tenants []models.Tenant) bool {
	logging.Infof("Running nodepools migration %s on tenants %+v", conf.MigrationVersion, tenants)
	hadFailure := false

	for _, tenant := range tenants {
		err := migrateTenant(tenant)

		hadFailure = hadFailure || err != nil
	}

	err := database.SetMigrationCompleted()
	if err != nil {
		logging.Errorf("Failed to set migration completed. err: %v", err)
		hadFailure = true
	}

	return hadFailure
}

func migrateTenant(tenant models.Tenant) (err error) {
	defer func() {
		if err != nil {
			logging.Errorf("Had an error migrating tenant %s. error: %v", tenant.Name, err)
			database.SetNodepoolsMigrationStatusForTenant(tenant.Id, models.FAILED)
		} else {
			database.SetNodepoolsMigrationStatusForTenant(tenant.Id, models.COMPLETED)
		}
	}()

	if database.IsNodepoolsMigrationCompletedForTenant(tenant.Id) {
		logging.Infof("Tenant %s has already been migrated", tenant.Name)
		return
	}

	startTime := time.Now()
	logging.Infof("Starting tenant %s nodepools migration", tenant.Name)

	database.SetNodepoolsMigrationStatusForTenant(tenant.Id, models.STARTED)

	nodepools, err := database.GetTenantNodePools(tenant.Id)
	if err != nil {
		return
	}

	clusterNamesMap, err := database.GetTenantClusterNames(tenant.Id)
	if err != nil {
		return
	}

	err = CreateNodepools(tenant, clusterNamesMap, nodepools)
	if err != nil {
		return
	}

	logging.Infof("Tenant %s nodepools migration has been completed in %d milli seconds", tenant.Name, time.Now().UnixMilli()-startTime.UnixMilli())
	return
}

func CreateNodepools(tenant models.Tenant, clusterNamesMap map[string]string, nodepools []models.NodepoolV1) error {
	nodepoolsToCreate := []clusterServiceModels.Nodepool{}
	for _, nodepoolV1 := range nodepools {
		logging.Debugf("Migrating nodepool %s of tenant %s.", nodepoolV1.Name, tenant.Name)
		nodepoolV2, err := MapNodepoolV1ToV2(clusterNamesMap, nodepoolV1)
		if err != nil {
			logging.Warnf("got an error mapping nodepool %s struct of tennat %d. error: %v", nodepoolV1.Name, nodepoolV1.TenantId, err)
		} else {
			nodepoolsToCreate = append(nodepoolsToCreate, nodepoolV2)
		}
	}

	err := database.CreateNodepools(context.Background(), nodepoolsToCreate)
	if err != nil {
		logging.Errorf("Failed to create nodepools of tenant %s. Err: %v", tenant.Name, err)
		return err
	}

	return nil
}

func MapNodepoolV1ToV2(clusterNamesMap map[string]string, nodepoolV1 models.NodepoolV1) (clusterServiceModels.Nodepool, error) {
	var clusterName string
	var ok bool
	if clusterName, ok = clusterNamesMap[nodepoolV1.ClusterId.String()]; !ok {
		return clusterServiceModels.Nodepool{}, fmt.Errorf("cluster %s not found", nodepoolV1.ClusterId.String())
	}

	return clusterServiceModels.Nodepool{
		Common: clusterServiceModels.Common{
			CreatedAt: nodepoolV1.CreatedAt,
			UpdatedAt: nodepoolV1.UpdatedAt,
		},
		ID:                   nodepoolV1.Id,
		Name:                 nodepoolV1.Name,
		TenantId:             nodepoolV1.TenantId,
		LabelKey:             nodepoolV1.LabelKey,
		LabelValue:           nodepoolV1.LabelValue,
		Phase:                clusterServiceModels.Nodepool{}.ParseNodepoolPhase(nodepoolV1.Status),
		PhaseMessage:         nodepoolV1.StatusMessage,
		ClusterId:            nodepoolV1.ClusterId,
		ClusterName:          clusterName,
		DeletedAt:            nodepoolV1.DeletedAt,
		CreatedBy:            convutils.Deref(nodepoolV1.CreatedBy, ""),
		UpdatedBy:            convutils.Deref(nodepoolV1.UpdatedBy, ""),
		DeletedBy:            nil,
		IsDefault:            nodepoolV1.IsDefault,
		GpuPlacementStrategy: clusterServiceModels.Nodepool{}.ParsePlacementStrategy(nodepoolV1.GpuPlacementStrategy),
		CpuPlacementStrategy: clusterServiceModels.Nodepool{}.ParsePlacementStrategy(nodepoolV1.CpuPlacementStrategy),
		NodepoolStatus: clusterServiceModels.NodepoolStatus{
			Nodes: getNodes(nodepoolV1.Nodes),
		},
		AllowAdjustableWeights: clusterServiceModels.DefaultAllowAdjustableWeights,
		AllowQuota:             clusterServiceModels.DefaultAllowQuota,
		AllowRanks:             clusterServiceModels.DefaultAllowRanks,
		AllowCpuQuota:          clusterServiceModels.DefaultAllowCpuQuota,
	}, nil
}

func getNodes(nodes *string) pq.StringArray {
	if nodes == nil {
		return []string{}
	}

	nodesString := strings.Trim(*nodes, "[]")

	nodesArray := []string{}

	if nodesString != "" {
		nodesArray = strings.Split(nodesString, ",")
	}

	return nodesArray
}
