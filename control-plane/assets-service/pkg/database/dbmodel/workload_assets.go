package dbmodel

// AssetsContainedInWorkload includes the specs of all the assets that can be directly contained in a workload
// For example: environment or compute but not credentials, as workload cannot directly contain credential assets
type AssetsContainedInWorkload struct {
	AssetMetaRec       `gorm:"embedded"`
	AssetsEnvironment  *EnvironmentAssetRec `gorm:"foreignKey:AssetId"`
	AssetsCompute      *ComputeAssetRec     `gorm:"foreignKey:AssetId"`
	AssetsS3           *S3AssetRec          `gorm:"foreignKey:AssetId"`
	AssetsPvc          *PVCAssetRec         `gorm:"foreignKey:AssetId"`
	AssetsGit          *GitAssetRec         `gorm:"foreignKey:AssetId"`
	AssetsHostPath     *HostPathAssetRec    `gorm:"foreignKey:AssetId"`
	AssetsNfs          *NFSAssetRec         `gorm:"foreignKey:AssetId"`
	AssetsConfigMap    *ConfigMapAssetRec   `gorm:"foreignKey:AssetId"` // AssetsConfigMap holds ConfigMap record
	AssetsSecretVolume *SecretAssetRec      `gorm:"foreignKey:AssetId"` // AssetsSecretVolume holds Secret record

	IsImposed bool
}
