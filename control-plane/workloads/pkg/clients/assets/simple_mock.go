package assets

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/run-ai/runai/runai-application-packages/datastructs"
	"github.com/run-ai/runai/runai-application-packages/jsonutils/conjungo"
	"github.com/run-ai/runai/runai-application-packages/jsonutils/gojsonq"

	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"

	assetsService "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
)

type AssetSpecToFlatKey struct {
	WorkloadType assetsService.WorkloadType
	ReplicaType  api.ReplicaType
	AssetId      string
}

// SimpleMock is a simple table-based mapping of asset id + workload type to spec
// assuming same spec for both worker and master. it also contains meta information for each asset
type SimpleMock struct {
	AssetMeta map[string]*assetsService.AssetMeta

	Templates map[string]*assetsService.SpecificRunInfoFields

	AssetSpecToFlat map[AssetSpecToFlatKey]*assetsService.SupersetSpec

	Credentials []assetsService.CredentialsListResponseEntry
}

func NewSimpleMock() *SimpleMock {
	return &SimpleMock{
		AssetSpecToFlat: make(map[AssetSpecToFlatKey]*assetsService.SupersetSpec),
		AssetMeta:       make(map[string]*assetsService.AssetMeta),
		Templates:       make(map[string]*assetsService.SpecificRunInfoFields),
		Credentials:     make([]assetsService.CredentialsListResponseEntry, 0),
	}
}

func (smc *SimpleMock) WithMapping(workloadType api.WorkloadType, assetId string, flatFormat *assetsService.SupersetSpec) *SimpleMock {
	smc.AssetSpecToFlat[AssetSpecToFlatKey{
		WorkloadType: assetsService.WorkloadType(workloadType),
		AssetId:      assetId,
	}] = flatFormat

	return smc
}

func (smc *SimpleMock) WithTemplate(template assetsService.WorkloadTemplate) *SimpleMock {
	smc.AssetMeta[template.Meta.Id] = &template.Meta
	smc.Templates[template.Meta.Id] = &template.Spec

	return smc
}

func (smc *SimpleMock) WithAssetMeta(assetMeta *assetsService.AssetMeta) *SimpleMock {
	smc.AssetMeta[assetMeta.Id] = assetMeta

	return smc
}

func (smc *SimpleMock) RequestAssetFlatFormat(ctx context.Context, request assetsService.AssetFlatRequest) (*assetsService.AssetFlatResponse, *assetsService.Error) {
	assetIds := make([]string, 0)
	if request.AssetId != "" {
		assetIds = append(assetIds, request.AssetId)
	}
	if request.AssetIds != nil {
		assetIds = append(assetIds, request.AssetIds...)
	}

	resultAsMap := make(map[string]interface{}, 0)

	assetMetas := make([]assetsService.AssetMeta, 0, len(request.AssetIds)+1)
	assetIdHandled := datastructs.NewSet[string]()

	for _, assetId := range assetIds {
		if !assetIdHandled.Has(assetId) {
			if assetMeta, ok := smc.AssetMeta[assetId]; ok {
				assetMetas = append(assetMetas, *assetMeta)
			}
			assetIdHandled.Add(assetId)
		}

		workerSpec := smc.AssetSpecToFlat[AssetSpecToFlatKey{
			WorkloadType: request.WorkloadType,
			AssetId:      assetId,
		}]
		if workerSpec == nil {
			if smc.AssetMeta[assetId] == nil {
				return nil, assetsService.NewError(http.StatusNotFound, "asset not found")
			} else {
				workerSpec = &assetsService.SupersetSpec{}
			}
		}

		assetSpecAsMap, jsonErr := gojsonq.LoadJson(workerSpec)
		if jsonErr != nil {
			return nil, assetsService.NewError(http.StatusInternalServerError, jsonErr.Error())
		}

		mergeErr := conjungo.Merge(&resultAsMap, assetSpecAsMap, conjungo.NewOptions())
		if mergeErr != nil {
			return nil, assetsService.NewError(http.StatusInternalServerError, mergeErr.Error())
		}
	}

	var workerSpec assetsService.SupersetSpec
	jsonErr := json.Unmarshal(gojsonq.ToBytes(resultAsMap), &workerSpec)
	if jsonErr != nil {
		return nil, assetsService.NewError(http.StatusInternalServerError, jsonErr.Error())
	}
	response := assetsService.AssetFlatResponse{
		AssetsMeta: assetMetas,
		Worker:     &workerSpec,
	}

	if request.WorkloadType == assetsService.WORKLOADTYPE_DISTRIBUTED {
		asBytes, copyErr := json.Marshal(workerSpec)
		if copyErr != nil {
			return nil, assetsService.NewError(http.StatusInternalServerError, fmt.Sprintf("copy specs failed: %v", copyErr.Error()))
		}

		var masterSpec assetsService.MasterSpec
		copyErr = json.Unmarshal(asBytes, &masterSpec)
		if copyErr != nil {
			return nil, assetsService.NewError(http.StatusInternalServerError, fmt.Sprintf("load specs during copy failed: %v", copyErr.Error()))
		}

		response.Master = *assetsService.NewNullableMasterSpec(&masterSpec)
	}

	return &response, nil
}

func (smc *SimpleMock) ListTemplatesOfTenant(_ context.Context) (*assetsService.WorkloadTemplateListResponse, *assetsService.Error) {
	result := assetsService.WorkloadTemplateListResponse{
		Entries: make([]assetsService.WorkloadTemplate, 0),
	}

	for assetId, spec := range smc.Templates {
		result.Entries = append(result.Entries, assetsService.WorkloadTemplate{
			Meta: *smc.AssetMeta[assetId],
			Spec: *spec,
		})
	}

	return &result, nil
}

func (smc *SimpleMock) ListCredentialsAssets(_ context.Context, privateOnly bool) (*assetsService.CredentialsListResponse, *assetsService.Error) {
	// dummy implementation
	return &assetsService.CredentialsListResponse{
		Entries: smc.Credentials,
	}, nil
	//return nil, assetsService.NewError(http.StatusBadRequest, "simple mock doesn't mock ListCredentialsAssets, implement or use mockery")
}

func (smc *SimpleMock) WithCredentials(credentials []assetsService.CredentialsListResponseEntry) *SimpleMock {
	smc.Credentials = credentials
	return smc
}
