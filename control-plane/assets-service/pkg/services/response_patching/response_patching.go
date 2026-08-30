package response_patching

import (
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
)

// ResponsePatchingPtrs holds pointers which assist us in storing usage info
// in asset lists. The problem it is trying to solve is that the Meta and usage
// info are the same for all assets, but the Spec is different. see the load
// function below for further info.
type ResponsePatchingPtrs struct {
	Meta           *api.AssetMeta
	UsedBy         **api.AssetUsageInfo
	UsageTimes     **api.UsageTimesInfo
	ComplianceInfo **api.ComplianceInfo
	Spec           interface{}
	Status         **api.AssetClusterStatusInfo
}

// LoadResponsePatchingPtrs receives a list-asset struct and create a map
// of assetId -> usagePatchingPtrs. it will allow us to implement a general
// function for storing usage info for any kind of asset, even though the
// original list structure contains the Spec field.
func LoadResponsePatchingPtrs(assets interface{}) map[string]ResponsePatchingPtrs {
	idToPatchingPtrsMap := make(map[string]ResponsePatchingPtrs)

	switch realEntry := assets.(type) {

	case []*api.AssetMeta:
		// this is intended for project name patching, which is patched on the metadata itself
		for _, metaEntry := range realEntry {
			idToPatchingPtrsMap[metaEntry.Id] = ResponsePatchingPtrs{
				Meta: metaEntry,
			}
		}

	case *api.EnvironmentListResponse:
		for idx, entry := range realEntry.Entries {
			idToPatchingPtrsMap[entry.Meta.Id] = ResponsePatchingPtrs{
				Meta:           &realEntry.Entries[idx].Meta,
				Spec:           &realEntry.Entries[idx].Spec,
				UsedBy:         &realEntry.Entries[idx].UsedBy,
				UsageTimes:     &realEntry.Entries[idx].UsageTimes,
				ComplianceInfo: &realEntry.Entries[idx].Compliance,
				Status:         &realEntry.Entries[idx].Status,
			}
		}

	case *api.EnvironmentAsset:
		idToPatchingPtrsMap[realEntry.Meta.Id] = ResponsePatchingPtrs{
			Meta:           &realEntry.Meta,
			Spec:           &realEntry.Spec,
			UsedBy:         &realEntry.UsedBy,
			UsageTimes:     &realEntry.UsageTimes,
			ComplianceInfo: &realEntry.Compliance,
			Status:         &realEntry.Status,
		}

	case *api.ComputeListResponse:
		for idx, entry := range realEntry.Entries {
			idToPatchingPtrsMap[entry.Meta.Id] = ResponsePatchingPtrs{
				Meta:           &realEntry.Entries[idx].Meta,
				Spec:           &realEntry.Entries[idx].Spec,
				UsedBy:         &realEntry.Entries[idx].UsedBy,
				UsageTimes:     &realEntry.Entries[idx].UsageTimes,
				ComplianceInfo: &realEntry.Entries[idx].Compliance,
			}
		}

	case *api.ComputeAsset:
		idToPatchingPtrsMap[realEntry.Meta.Id] = ResponsePatchingPtrs{
			Meta:           &realEntry.Meta,
			Spec:           &realEntry.Spec,
			UsedBy:         &realEntry.UsedBy,
			UsageTimes:     &realEntry.UsageTimes,
			ComplianceInfo: &realEntry.Compliance,
		}

	case *api.HostPathListResponse:
		for idx, entry := range realEntry.Entries {
			idToPatchingPtrsMap[entry.Meta.Id] = ResponsePatchingPtrs{
				Meta:           &realEntry.Entries[idx].Meta,
				Spec:           &realEntry.Entries[idx].Spec,
				UsedBy:         &realEntry.Entries[idx].UsedBy,
				UsageTimes:     &realEntry.Entries[idx].UsageTimes,
				ComplianceInfo: &realEntry.Entries[idx].Compliance,
			}
		}

	case *api.HostPathAsset:
		idToPatchingPtrsMap[realEntry.Meta.Id] = ResponsePatchingPtrs{
			Meta:           &realEntry.Meta,
			Spec:           &realEntry.Spec,
			UsedBy:         &realEntry.UsedBy,
			UsageTimes:     &realEntry.UsageTimes,
			ComplianceInfo: &realEntry.Compliance,
		}

	case *api.NfsListResponse:
		for idx, entry := range realEntry.Entries {
			idToPatchingPtrsMap[entry.Meta.Id] = ResponsePatchingPtrs{
				Meta:           &realEntry.Entries[idx].Meta,
				Spec:           &realEntry.Entries[idx].Spec,
				UsedBy:         &realEntry.Entries[idx].UsedBy,
				UsageTimes:     &realEntry.Entries[idx].UsageTimes,
				ComplianceInfo: &realEntry.Entries[idx].Compliance,
			}
		}

	case *api.NfsAsset:
		idToPatchingPtrsMap[realEntry.Meta.Id] = ResponsePatchingPtrs{
			Meta:           &realEntry.Meta,
			Spec:           &realEntry.Spec,
			UsedBy:         &realEntry.UsedBy,
			UsageTimes:     &realEntry.UsageTimes,
			ComplianceInfo: &realEntry.Compliance,
		}

	case *api.PvcListResponse:
		for idx, entry := range realEntry.Entries {
			idToPatchingPtrsMap[entry.Meta.Id] = ResponsePatchingPtrs{
				Meta:           &realEntry.Entries[idx].Meta,
				Spec:           &realEntry.Entries[idx].Spec,
				UsedBy:         &realEntry.Entries[idx].UsedBy,
				UsageTimes:     &realEntry.Entries[idx].UsageTimes,
				ComplianceInfo: &realEntry.Entries[idx].Compliance,
				Status:         &realEntry.Entries[idx].Status,
			}
		}

	case *api.PvcAsset:
		idToPatchingPtrsMap[realEntry.Meta.Id] = ResponsePatchingPtrs{
			Meta:           &realEntry.Meta,
			Spec:           &realEntry.Spec,
			UsedBy:         &realEntry.UsedBy,
			UsageTimes:     &realEntry.UsageTimes,
			ComplianceInfo: &realEntry.Compliance,
			Status:         &realEntry.Status,
		}

	case *api.ConfigMapListResponse:
		for idx, entry := range realEntry.Entries {
			idToPatchingPtrsMap[entry.Meta.Id] = ResponsePatchingPtrs{
				Meta:           &realEntry.Entries[idx].Meta,
				Spec:           &realEntry.Entries[idx].Spec,
				UsedBy:         &realEntry.Entries[idx].UsedBy,
				UsageTimes:     &realEntry.Entries[idx].UsageTimes,
				ComplianceInfo: &realEntry.Entries[idx].Compliance,
				Status:         &realEntry.Entries[idx].Status,
			}
		}

	case *api.ConfigMapAsset:
		idToPatchingPtrsMap[realEntry.Meta.Id] = ResponsePatchingPtrs{
			Meta:           &realEntry.Meta,
			Spec:           &realEntry.Spec,
			UsedBy:         &realEntry.UsedBy,
			UsageTimes:     &realEntry.UsageTimes,
			ComplianceInfo: &realEntry.Compliance,
			Status:         &realEntry.Status,
		}

	case *api.SecretAssetListResponse:
		for idx, entry := range realEntry.Entries {
			idToPatchingPtrsMap[entry.Meta.Id] = ResponsePatchingPtrs{
				Meta:           &realEntry.Entries[idx].Meta,
				Spec:           &realEntry.Entries[idx].Spec,
				UsedBy:         &realEntry.Entries[idx].UsedBy,
				UsageTimes:     &realEntry.Entries[idx].UsageTimes,
				ComplianceInfo: &realEntry.Entries[idx].Compliance,
				Status:         &realEntry.Entries[idx].Status,
			}
		}

	case *api.SecretAsset:
		idToPatchingPtrsMap[realEntry.Meta.Id] = ResponsePatchingPtrs{
			Meta:           &realEntry.Meta,
			Spec:           &realEntry.Spec,
			UsedBy:         &realEntry.UsedBy,
			UsageTimes:     &realEntry.UsageTimes,
			ComplianceInfo: &realEntry.Compliance,
			Status:         &realEntry.Status,
		}

	case *api.GitListResponse:
		for idx, entry := range realEntry.Entries {
			idToPatchingPtrsMap[entry.Meta.Id] = ResponsePatchingPtrs{
				Meta:           &realEntry.Entries[idx].Meta,
				Spec:           &realEntry.Entries[idx].Spec,
				UsedBy:         &realEntry.Entries[idx].UsedBy,
				UsageTimes:     &realEntry.Entries[idx].UsageTimes,
				ComplianceInfo: &realEntry.Entries[idx].Compliance,
				Status:         &realEntry.Entries[idx].Status,
			}
		}

	case *api.GitAsset:
		idToPatchingPtrsMap[realEntry.Meta.Id] = ResponsePatchingPtrs{
			Meta:           &realEntry.Meta,
			Spec:           &realEntry.Spec,
			UsedBy:         &realEntry.UsedBy,
			UsageTimes:     &realEntry.UsageTimes,
			ComplianceInfo: &realEntry.Compliance,
			Status:         &realEntry.Status,
		}

	case *api.S3ListResponse:
		for idx, entry := range realEntry.Entries {
			idToPatchingPtrsMap[entry.Meta.Id] = ResponsePatchingPtrs{
				Meta:           &realEntry.Entries[idx].Meta,
				Spec:           &realEntry.Entries[idx].Spec,
				UsedBy:         &realEntry.Entries[idx].UsedBy,
				UsageTimes:     &realEntry.Entries[idx].UsageTimes,
				ComplianceInfo: &realEntry.Entries[idx].Compliance,
				Status:         &realEntry.Entries[idx].Status,
			}
		}

	case *api.S3Asset:
		idToPatchingPtrsMap[realEntry.Meta.Id] = ResponsePatchingPtrs{
			Meta:           &realEntry.Meta,
			Spec:           &realEntry.Spec,
			UsedBy:         &realEntry.UsedBy,
			UsageTimes:     &realEntry.UsageTimes,
			ComplianceInfo: &realEntry.Compliance,
			Status:         &realEntry.Status,
		}

	case []api.DatasourceListResponseEntry:
		for idx, entry := range realEntry {
			idToPatchingPtrsMap[entry.Meta.Id] = ResponsePatchingPtrs{
				Meta:           &realEntry[idx].Meta,
				Spec:           &realEntry[idx].Spec,
				UsedBy:         &realEntry[idx].UsedBy,
				UsageTimes:     &realEntry[idx].UsageTimes,
				ComplianceInfo: &realEntry[idx].Compliance,
				Status:         &realEntry[idx].Status,
			}
		}

	case []api.CredentialsListResponseEntry:
		for idx, entry := range realEntry {
			idToPatchingPtrsMap[entry.Meta.Id] = ResponsePatchingPtrs{
				Meta:   &realEntry[idx].Meta,
				Spec:   &realEntry[idx].Spec,
				UsedBy: &realEntry[idx].UsedBy,
				Status: &realEntry[idx].Status,
			}
		}

	case *api.AccessKey:
		idToPatchingPtrsMap[realEntry.Meta.Id] = ResponsePatchingPtrs{
			Meta:   &realEntry.Meta,
			Spec:   &realEntry.Spec,
			UsedBy: &realEntry.UsedBy,
			Status: &realEntry.Status,
		}

	case []api.AccessKey:
		for idx, entry := range realEntry {
			idToPatchingPtrsMap[entry.Meta.Id] = ResponsePatchingPtrs{
				Meta:   &realEntry[idx].Meta,
				Spec:   &realEntry[idx].Spec,
				UsedBy: &realEntry[idx].UsedBy,
				Status: &realEntry[idx].Status,
			}
		}

	case *api.Password:
		idToPatchingPtrsMap[realEntry.Meta.Id] = ResponsePatchingPtrs{
			Meta:   &realEntry.Meta,
			Spec:   &realEntry.Spec,
			UsedBy: &realEntry.UsedBy,
			Status: &realEntry.Status,
		}

	case []api.Password:
		for idx, entry := range realEntry {
			idToPatchingPtrsMap[entry.Meta.Id] = ResponsePatchingPtrs{
				Meta:   &realEntry[idx].Meta,
				Spec:   &realEntry[idx].Spec,
				UsedBy: &realEntry[idx].UsedBy,
				Status: &realEntry[idx].Status,
			}
		}

	case *api.DockerRegistry:
		idToPatchingPtrsMap[realEntry.Meta.Id] = ResponsePatchingPtrs{
			Meta:   &realEntry.Meta,
			Spec:   &realEntry.Spec,
			Status: &realEntry.Status,
		}

	case []api.DockerRegistry:
		for idx, entry := range realEntry {
			idToPatchingPtrsMap[entry.Meta.Id] = ResponsePatchingPtrs{
				Meta:   &realEntry[idx].Meta,
				Spec:   &realEntry[idx].Spec,
				Status: &realEntry[idx].Status,
			}
		}
	case *api.GenericSecret:
		idToPatchingPtrsMap[realEntry.Meta.Id] = ResponsePatchingPtrs{
			Meta:   &realEntry.Meta,
			Spec:   &realEntry.Spec,
			UsedBy: &realEntry.UsedBy,
			Status: &realEntry.Status,
		}
	case []api.GenericSecret:
		for idx, entry := range realEntry {
			idToPatchingPtrsMap[entry.Meta.Id] = ResponsePatchingPtrs{
				Meta:   &realEntry[idx].Meta,
				Spec:   &realEntry[idx].Spec,
				UsedBy: &realEntry[idx].UsedBy,
				Status: &realEntry[idx].Status,
			}
		}
	case *api.NgcApiKey:
		idToPatchingPtrsMap[realEntry.Meta.Id] = ResponsePatchingPtrs{
			Meta:   &realEntry.Meta,
			Spec:   &realEntry.Spec,
			UsedBy: &realEntry.UsedBy,
			Status: &realEntry.Status,
		}
	case []api.NgcApiKey:
		for idx, entry := range realEntry {
			idToPatchingPtrsMap[entry.Meta.Id] = ResponsePatchingPtrs{
				Meta:   &realEntry[idx].Meta,
				Spec:   &realEntry[idx].Spec,
				UsedBy: &realEntry[idx].UsedBy,
				Status: &realEntry[idx].Status,
			}
		}
	}
	return idToPatchingPtrsMap
}
