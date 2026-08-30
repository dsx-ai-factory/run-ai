package dao

import (
	"fmt"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbtables"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
)

func createWorkloadSupportedTypes(tx orm.DB, id string, _ api.AssetKind, assetMeta api.AssetCreationRequest) (*dbmodel.WorkloadSupportedTypes, error) {
	WorkloadSupportedTypesRec := dbmodel.WorkloadSupportedTypesRec{
		WorkloadSupportedTypes: dbmodel.WorkloadSupportedTypes{
			AssetId:       id,
			Workspace:     assetMeta.WorkloadSupportedTypes.Workspace,
			Training:      assetMeta.WorkloadSupportedTypes.Training,
			Inference:     assetMeta.WorkloadSupportedTypes.Inference,
			Distributed:   assetMeta.WorkloadSupportedTypes.Distributed,
			DistFramework: assetMeta.WorkloadSupportedTypes.DistFramework,
		},
	}

	_, err := tx.Create(&WorkloadSupportedTypesRec)
	if err != nil {
		return nil, err
	}

	return &WorkloadSupportedTypesRec.WorkloadSupportedTypes, nil
}

func WorkloadSupportedTypesOf(tx orm.DB, id string) (*dbmodel.WorkloadSupportedTypes, error) {
	WorkloadSupportedTypesRec := dbmodel.WorkloadSupportedTypesRec{
		WorkloadSupportedTypes: dbmodel.WorkloadSupportedTypes{AssetId: id},
	}

	selectedColumns := fmt.Sprintf("%s.*", dbtables.TableNameWorkloadSupportedTypes)

	_, err := tx.
		Model(dbmodel.WorkloadSupportedTypesRec{}).
		Select(selectedColumns).
		Scan(&WorkloadSupportedTypesRec)

	return &WorkloadSupportedTypesRec.WorkloadSupportedTypes, err
}

func DeleteWorkloadSupportedTypesOfByIds(tx orm.DB, assetIds []string) (int64, error) {
	if len(assetIds) == 0 {
		return 0, nil
	}

	return tx.Where(fmt.Sprintf("%s IN ?", dbtables.WorkloadSupportedTypesColumnAssetId), assetIds).
		Delete(&dbmodel.WorkloadSupportedTypesRec{})
}
