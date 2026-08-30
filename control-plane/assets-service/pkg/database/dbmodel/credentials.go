package dbmodel

import (
	"fmt"
	"strings"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/runai-application-packages/convutils"
)

type CredentialsAssetSpecs struct {
	AssetMetaRec        `gorm:"embedded"`
	AssetsAccessKey     *AccessKeyRecord     `gorm:"foreignKey:AssetId"`
	AssetsPassword      *PasswordRecord      `gorm:"foreignKey:AssetId"`
	AssetsGenericSecret *GenericSecretRecord `gorm:"foreignKey:AssetId"`
	AssetsNgcApiKey     *NgcApiKeyRecord     `gorm:"foreignKey:AssetId"`
}

// SecretTypeMap is a maps the secret type to the prefix that the cluster adds to the secret name
var SecretTypeMap = map[api.AssetKind]string{
	api.ASSETKIND_PASSWORD:        "password",
	api.ASSETKIND_ACCESS_KEY:      "accesskey",
	api.ASSETKIND_DOCKER_REGISTRY: "dockerregistry",
	api.ASSETKIND_GENERIC_SECRET:  "genericsecret",
	api.ASSETKIND_NGC_API_KEY:     "ngc",
}

// SecretName returns the name of the underlying secret k8s resource
func (c *CredentialsAssetSpecs) SecretName() *string {
	var existingName *string
	switch c.Kind {
	case api.ASSETKIND_PASSWORD:
		existingName = c.AssetsPassword.ExistingSecretName
	case api.ASSETKIND_ACCESS_KEY:
		existingName = c.AssetsAccessKey.ExistingSecretName
	case api.ASSETKIND_GENERIC_SECRET:
		existingName = c.AssetsGenericSecret.ExistingSecretName
	}

	return GetSecretName(c.Kind, c.Name, existingName)
}

func GetSecretName(kind api.AssetKind, name string, existingName *string) *string {
	if existingName == nil {
		existingName = convutils.Ref(strings.ToLower(fmt.Sprintf("%s-%s", SecretTypeMap[kind], name)))
	}
	return existingName
}

func GetSecretClusterInfo(assetMeta *api.AssetMeta, existingSecret *string) *api.ClusterInfo {
	return &api.ClusterInfo{
		Resources: []api.ClusterResourceInfo{
			{
				Name: GetSecretName(assetMeta.Kind, assetMeta.Name, existingSecret),
			},
		},
	}
}
