package constants

const (
	// RunaiAssetLabel - asset sync owned resources. values: password, access-key, docker-registry, pvc
	RunaiAssetLabel = "run.ai/asset"
	// RunaiResourceLabel - user owned secrets. values: password, access-key, docker-registry
	RunaiResourceLabel = "run.ai/resource"
	// RunaiDataSharingLabel - PVCs that are sharing their data (should not be replicated since the shared-objects-controller share them)
	RunaiDataSharingLabel = "run.ai/data-sharing-pvc"
	// RunaiAssetIdLabel - the id of the asset that owns the resource
	RunaiAssetIdLabel = "run.ai/asset-id"
)
