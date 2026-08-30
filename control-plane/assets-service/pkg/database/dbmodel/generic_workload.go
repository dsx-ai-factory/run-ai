package dbmodel

import (
	"fmt"
	"sort"
	"time"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/runai-application-packages/convutils"
)

type GenericWorkload struct {
	WorkloadType string `json:"workloadType" gorm:"not null;"`

	Meta api.WorkloadMetaV1 `json:"meta" gorm:"serializer:json; type:jsonb;not null;"`

	CreatedAt *time.Time

	Spec api.SpecificRunInfoFields `json:"spec" gorm:"serializer:json; type:jsonb;not null;"`

	// A workload ready for submission at the cluster. Further details at https://github.com/run-ai/runai/runai-cluster/tree/master/workload-controller/config/crd\"
	Workload map[string]interface{} `json:"workload,omitempty" gorm:"not null;"`

	Job api.DisplayedJob `json:"job,omitempty" gorm:"serializer:json; type:jsonb;not null;"`
}

func (gw GenericWorkload) ToWorkloadMeta() *api.WorkloadMeta {
	result := api.WorkloadMeta{
		Name:          gw.Meta.Name,
		RequestedName: gw.Meta.Name,
		WorkloadId:    gw.Meta.Id,
		ClusterId:     gw.Meta.ClusterId,
		CreatedBy:     gw.Meta.CreatedBy,
	}

	if gw.Meta.ProjectId != 0 {
		result.ProjectId = fmt.Sprintf("%d", gw.Meta.ProjectId)
	}

	if gw.CreatedAt != nil {
		result.CreatedAt = *gw.CreatedAt
	}

	return &result
}

func AssetsRefToAssetIds(assetsRef *api.AssetsRef) api.AssetsIds {
	var result api.AssetsIds

	if assetsRef == nil {
		return result
	}

	result.Environment = assetsRef.Environment.Id

	if assetsRef.Compute != nil {
		result.Compute = convutils.Ref(assetsRef.Compute.Id)
	}

	if assetsRef.Datasources != nil {
		result.Datasources = make([]api.AssetIdAndKind, 0)

		// sorting is mainly for testing purposes (when we compare two objects)
		sort.Slice(assetsRef.Datasources, func(i, j int) bool {
			return assetsRef.Datasources[i].Id < assetsRef.Datasources[j].Id
		})

		for _, ds := range assetsRef.Datasources {
			result.Datasources = append(result.Datasources, api.AssetIdAndKind{
				Id:   ds.Id,
				Kind: ds.Kind,
			})
		}
	}

	result.WorkloadVolumes = assetsRef.WorkloadVolumes

	return result
}
