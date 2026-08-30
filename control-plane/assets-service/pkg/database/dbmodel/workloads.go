package dbmodel

import (
	"time"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbtables"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
)

type CombinedWorkloadsRec struct {
	WorkloadRec        WorkloadRec        `gorm:"embedded"`
	WorkloadsAssetsRec WorkloadsAssetsRec `gorm:"embedded"`

	AssetName string
}

func (CombinedWorkloadsRec) TableName() string {
	return dbtables.TableNameWorkloadsAssets
}

type DistSpecificEnv struct {
	NumWorkers int32 `json:"numWorkers,omitempty" gorm:"not null;"`

	NoMaster *bool `json:"noMaster,omitempty" gorm:""`

	MasterSpecificRunParams *api.SpecificRunParams `json:"master,omitempty" gorm:"serializer:json; type:jsonb;"`
}

// WorkloadRec models a record in the workloads table. this is the table that assists us
// in constructing the list workload, which show a combination of information from both jobs
// and workloads tables
type WorkloadRec struct {
	api.WorkloadMetaV1 `gorm:"embedded"`
	CreatedAt          time.Time                 `json:"createdAt"`
	WorkloadType       string                    `json:"workloadType" gorm:"not null;"`
	SpecificEnv        *api.SpecificRunParams    `json:"specificEnv" gorm:"serializer:json; type:jsonb;"`
	Workload           jobs.JobWorkload          `json:"workload" gorm:"serializer:json; type:jsonb;"`
	DistSpecificEnv    *DistSpecificEnv          `json:"distSpecificEnv" gorm:"serializer:json; type:jsonb;"`
	DistFramework      *api.DistributedFramework `json:"distFramework" gorm:""`
}

func (WorkloadRec) TableName() string {
	return dbtables.TableNameWorkloads
}

// WorkloadsAssetsRec models a record in the workloads_assets table, which maps
// between a workload and the assets it contains0.
type WorkloadsAssetsRec struct {
	WorkloadId      string                   `json:"workloadId" gorm:"not null;"`
	AssetId         string                   `json:"assetId" gorm:"not null;"`
	AssetKind       api.AssetKind            `json:"assetKind" gorm:"not null;"`
	DistReplicaType DistReplicaType          `json:"distReplicaType" gorm:""`
	AutoDelete      bool                     `json:"autoDelete" gorm:"not null;"`
	Overrides       *api.DataSourceOverrides `json:"overrides" gorm:"serializer:json; type:jsonb;"`
}

func (WorkloadsAssetsRec) TableName() string {
	return dbtables.TableNameWorkloadsAssets
}

type DistReplicaType string

const (
	DistReplicaTypeMaster DistReplicaType = "master"
	DistReplicaTypeWorker DistReplicaType = "worker"
)
