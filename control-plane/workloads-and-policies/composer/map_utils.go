package composer

import (
	"sort"

	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

type mapValue interface {
	string |
		workloadsApi.Connection |
		workloadsApi.Capability |
		workloadsApi.ExtendedResource |
		workloadsApi.HostPathInstance |
		workloadsApi.NfsInstance |
		workloadsApi.GitInstance |
		workloadsApi.PvcInstance |
		workloadsApi.S3Instance |
		workloadsApi.ConfigMapInstance |
		workloadsApi.EmptyDirInstance |
		workloadsApi.DataVolumeInstance |
		workloadsApi.SecretInstance |
		workloadsApi.Toleration |
		workloadsApi.ExposedUrl |
		workloadsApi.Port |
		GitSyncItem |
		ExposedUrlInfoItem |
		S3Item
}

func SortedMapKeys[T mapValue](unsorted map[string]*T) []string {
	result := make([]string, 0, len(unsorted))
	for k := range unsorted {
		result = append(result, k)
	}
	sort.Strings(result)
	return result
}

func CreateAssetMappingMap(assetMappings []workloadsApi.AssetsMapping) map[string]workloadsApi.AssetsMapping {
	assetsMap := make(map[string]workloadsApi.AssetsMapping)
	for _, assetMapping := range assetMappings {
		if assetMapping.Name != nil {
			assetsMap[*assetMapping.Name] = assetMapping
		}
	}
	return assetsMap
}
