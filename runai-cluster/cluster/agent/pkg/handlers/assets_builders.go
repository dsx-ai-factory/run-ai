/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	assetsApi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	assetIdLabel = "run.ai/asset-id"

	tenantScopeLabel     = "run.ai/tenant-wide"
	clusterScopeLabel    = "run.ai/cluster-wide"
	departmentScopeLabel = "run.ai/department"
	projectScopeLabel    = "run.ai/project"
	userCredentialsLabel = "run.ai/user-credentials"

	pvcKindLabel                   = "pvc"
	dockerRegistryKindLabel        = "docker-registry"
	passwordKindLabel              = "password"
	accessKeyKindLabel             = "access-key"
	genericSecretKindLabel         = "generic"
	resourceClaimTemplateKindLabel = "resource-claim-template"

	assetNameAnnotation = "run.ai/asset-name"

	UpdateCountAnnotation = "run.ai/update-count"

	boolTrueAsString = "true"

	attributesPrefix = "attributes.run.ai"
)

// encryptionKey holds the encryption key used to encrypt credential content
var encryptionKey []byte

func init() {
	key := os.Getenv(constants.RunaiAdsEncryptionKeyEnvVar)
	if key == "" {
		log.Warn().Msgf("Encryption key not found in environment variable %s", constants.RunaiAdsEncryptionKeyEnvVar)
	} else {
		encryptionKey = []byte(key)
	}
}

func (ah *AssetsHandler) composeClusterUpdatesOfDatasources(backendUpdates []AssetsFromBackend) (clusterUpdatesOfAssets, error) {
	clusterUpdates := clusterUpdatesOfAssets{
		UpdateRef: make([]int, 0, len(backendUpdates)),
		ToUpdate:  make([]client.Object, 0, len(backendUpdates)),

		DeleteRef: make([]int, 0, len(backendUpdates)),
		ToDelete:  make([]types.NamespacedName, 0),
	}

	projectToNamespaceMap, err := ah.ProjectToNamespaceList()
	if err != nil {
		return clusterUpdates, err
	}

	for idx, backendUpdate := range backendUpdates {
		namespace := constants.NamespaceRunAi // default namespace for tenant/department scope

		if backendUpdate.Meta.Scope == assetsApi.SCOPE_PROJECT {
			projectIdPtr := backendUpdate.Meta.ProjectId.Get()
			if projectIdPtr != nil {
				var found bool
				projectId := strconv.Itoa(int(*projectIdPtr))
				namespace, found = projectToNamespaceMap[projectId]
				if !found || namespace == "" {
					log.Warn().Msgf("Had an error while getting backend policy namespace, project id: %s", projectId)
					continue
				}
			} else {
				log.Warn().Msgf("project scope %s asset %s missing projectId", ah.assetGroup, backendUpdate.Meta.Name)
				continue
			}
		}

		pvcSpec := backendUpdate.Spec.(*assetsApi.DatasourceUpdateSpec).Pvc.Get()

		claimName := convutils.Deref(pvcSpec.ClaimName.Get(), "")

		if backendUpdate.IsDeleted {
			clusterUpdates.ToDelete = append(clusterUpdates.ToDelete, types.NamespacedName{
				Namespace: namespace,
				Name:      claimName,
			})
			clusterUpdates.DeleteRef = append(clusterUpdates.DeleteRef, idx)
			continue
		}

		clusterPvc := corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      claimName,
				Namespace: namespace,
			},
			Spec: corev1.PersistentVolumeClaimSpec{},
		}

		ah.composeLabelsAndAnnotations(&clusterPvc.ObjectMeta, &backendUpdate, nil)

		ah.composeAttributesAnnotations(&clusterPvc.ObjectMeta, pvcSpec.ClaimInfo.Get())

		ah.composePvcLabels(clusterPvc.Labels, *pvcSpec)

		err := ah.composePvcSpec(&clusterPvc.Spec, pvcSpec)
		if err != nil {
			log.Error().Msgf("compose spec of %s asset %s for update count %d failed: %v", ah.assetGroup, backendUpdate.Meta.Name, ah.syncCounter, err)
			continue
		}

		clusterUpdates.ToUpdate = append(clusterUpdates.ToUpdate, client.Object(&clusterPvc))
		clusterUpdates.UpdateRef = append(clusterUpdates.UpdateRef, idx)
	}

	return clusterUpdates, nil
}

func (ah *AssetsHandler) composeAttributesAnnotations(meta *metav1.ObjectMeta, claimInfo *assetsApi.ClaimInfo) {
	if claimInfo == nil || claimInfo.AddedAttrValues == nil {
		return
	}

	if meta.Annotations == nil {
		meta.Annotations = make(map[string]string)
	}

	for _, attr := range claimInfo.AddedAttrValues {
		annotationKey := fmt.Sprintf("%s/%s", attributesPrefix, attr.Key)
		meta.Annotations[annotationKey] = convutils.Deref(attr.Value, "")
	}
}

func (ah *AssetsHandler) composePvcSpec(spec *corev1.PersistentVolumeClaimSpec, pvcAssetSpec *assetsApi.PVCAssetSpec) error {
	if convutils.BoolIsTrue(pvcAssetSpec.ExistingPvc.Get()) {
		return errors.New("asset of existing PVC, skipped")
	}

	if pvcAssetSpec.ClaimInfo.Get() == nil {
		return errors.New("no claim info, skipped")
	}

	size := pvcAssetSpec.GetClaimInfo().Size.Get()

	if convutils.IsStringHasValue(size) {
		spec.Resources = corev1.VolumeResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse(*size),
			},
		}
	} else {
		return errors.New("no size, skipped")
	}

	accessModes := pvcAssetSpec.GetClaimInfo().AccessModes.Get()

	if accessModes != nil {
		if convutils.BoolIsTrue(accessModes.ReadOnlyMany.Get()) {
			spec.AccessModes = append(spec.AccessModes, corev1.ReadOnlyMany)
		}

		if convutils.BoolIsTrue(accessModes.ReadWriteOnce.Get()) {
			spec.AccessModes = append(spec.AccessModes, corev1.ReadWriteOnce)
		}

		if convutils.BoolIsTrue(accessModes.ReadWriteMany.Get()) {
			spec.AccessModes = append(spec.AccessModes, corev1.ReadWriteMany)
		}

		if len(spec.AccessModes) == 0 {
			spec.AccessModes = append(spec.AccessModes, corev1.ReadOnlyMany)
		}
	}

	storageClassName := pvcAssetSpec.GetClaimInfo().StorageClass.Get()

	if convutils.IsStringHasValue(storageClassName) {
		if *storageClassName == constants.EmptyStorageClass {
			spec.StorageClassName = convutils.Ref("")
		} else {
			spec.StorageClassName = storageClassName
		}
	}

	volumeMode := pvcAssetSpec.GetClaimInfo().VolumeMode.Get()

	if volumeMode != nil && *volumeMode != "" {
		spec.VolumeMode = convutils.Ref(corev1.PersistentVolumeMode(*volumeMode))
	}

	return nil
}

// composeLabelsAndAnnotations send private parameter only for private secrets
func (ah *AssetsHandler) composeLabelsAndAnnotations(meta *metav1.ObjectMeta, backendUpdate *AssetsFromBackend, private *bool) {
	if meta.Labels == nil {
		meta.Labels = make(map[string]string)
	}

	meta.Labels[constants.RunaiAssetLabel] = getAssetKindLabel(backendUpdate.Meta.Kind)
	meta.Labels[assetIdLabel] = backendUpdate.Meta.Id

	if convutils.BoolIsFalse(private) {
		composeScopeLabel(meta.Labels, &backendUpdate.Meta)
	}

	if meta.Annotations == nil {
		meta.Annotations = make(map[string]string)
	}

	meta.Annotations[assetNameAnnotation] = backendUpdate.Meta.Name
	meta.Annotations[UpdateCountAnnotation] = fmt.Sprintf("%d", backendUpdate.UpdateCount)
}

func elbrateNgcApiKeyAsets(backendUpdates []AssetsFromBackend) []AssetsFromBackend {
	assets := make([]AssetsFromBackend, 0, len(backendUpdates))
	for idx := range backendUpdates {
		if backendUpdates[idx].Meta.Kind != assetsApi.ASSETKIND_NGC_API_KEY {
			backendUpdates[idx].refIndex = idx
			assets = append(assets, backendUpdates[idx])
			continue
		}
		// Create two separate assets: generic secret and docker registry
		genericSecret := backendUpdates[idx]
		genericSecret.Meta.Name = fmt.Sprintf("ngcgs-%s", backendUpdates[idx].Meta.Name)
		if genericSecret.Meta.PrevName != nil {
			genericSecret.Meta.PrevName = convutils.Ref(fmt.Sprintf("ngcdr-%s", *backendUpdates[idx].Meta.PrevName))
		}
		genericSecret.Meta.Kind = assetsApi.ASSETKIND_GENERIC_SECRET
		genericSecret.refIndex = idx
		assets = append(assets, genericSecret)

		dockerRegistry := backendUpdates[idx]
		dockerRegistry.Meta.Name = fmt.Sprintf("ngcdr-%s", backendUpdates[idx].Meta.Name)
		if dockerRegistry.Meta.PrevName != nil {
			dockerRegistry.Meta.PrevName = convutils.Ref(fmt.Sprintf("ngcdr-%s", *backendUpdates[idx].Meta.PrevName))
		}
		dockerRegistry.Meta.Kind = assetsApi.ASSETKIND_DOCKER_REGISTRY
		dockerRegistry.refIndex = idx
		assets = append(assets, dockerRegistry)
	}
	return assets
}

func (ah *AssetsHandler) composeClusterUpdatesOfCredentials(backendUpdates []AssetsFromBackend) (clusterUpdatesOfAssets, error) {
	clusterUpdates := clusterUpdatesOfAssets{
		UpdateRef: make([]int, 0, len(backendUpdates)),
		ToUpdate:  make([]client.Object, 0, len(backendUpdates)),

		// for delete we need multiple size due to the rename
		DeleteRef: make([]int, 0, len(backendUpdates)*2),
		ToDelete:  make([]types.NamespacedName, 0, len(backendUpdates)*2),
	}

	projectToNamespaceMap, err := ah.ProjectToNamespaceList()
	if err != nil {
		return clusterUpdates, err
	}

	assets := elbrateNgcApiKeyAsets(backendUpdates)
	for _, backendUpdate := range assets {
		namespace := constants.NamespaceRunAi // default namespace for tenant/department scope

		private := convutils.BoolIsTrue(backendUpdate.Meta.Private.Get())

		// Private assets are stored in default namespace
		if backendUpdate.Meta.Scope == assetsApi.SCOPE_PROJECT && !private {
			projectIdPtr := backendUpdate.Meta.ProjectId.Get()
			if projectIdPtr != nil {
				var found bool
				projectId := strconv.Itoa(int(*projectIdPtr))
				namespace, found = projectToNamespaceMap[projectId]
				if !found || namespace == "" {
					log.Warn().Msgf("Had an error while getting backend policy namespace, project id: %s", projectId)
					continue
				}
			} else {
				log.Warn().Msgf("project scope %s asset %s missing projectId", ah.assetGroup, backendUpdate.Meta.Name)
				continue
			}
		}

		secretName := ah.getSecretResourceName(backendUpdate.Meta.Kind, backendUpdate.Meta.Name, backendUpdate.Meta.Id, private)

		// if asset has been renamed, attempt to remove the old secret
		if backendUpdate.Meta.PrevName != nil {
			clusterUpdates.ToDelete = append(clusterUpdates.ToDelete, types.NamespacedName{
				Namespace: namespace,
				Name:      ah.getSecretResourceName(backendUpdate.Meta.Kind, *backendUpdate.Meta.PrevName, backendUpdate.Meta.Id, private),
			})
			clusterUpdates.DeleteRef = append(clusterUpdates.DeleteRef, backendUpdate.refIndex)
		}

		if backendUpdate.IsDeleted {
			clusterUpdates.ToDelete = append(clusterUpdates.ToDelete, types.NamespacedName{
				Namespace: namespace,
				Name:      secretName,
			})
			clusterUpdates.DeleteRef = append(clusterUpdates.DeleteRef, backendUpdate.refIndex)
			continue
		}

		clusterSecret := corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: namespace,
			},
			Data: make(map[string][]byte),
		}

		// a designated secret type should be set
		if backendUpdate.Meta.Kind == assetsApi.ASSETKIND_DOCKER_REGISTRY && !private {
			clusterSecret.Type = corev1.SecretTypeDockerConfigJson
		}

		ah.composeLabelsAndAnnotations(&clusterSecret.ObjectMeta, &backendUpdate, convutils.Ref(private))

		// mark private secrets with special label
		if private {
			if clusterSecret.Labels == nil {
				clusterSecret.Labels = make(map[string]string)
			}
			clusterSecret.Labels[userCredentialsLabel] = boolTrueAsString
		}

		err := ah.composeSecretData(&clusterSecret, &backendUpdate, private)
		if err != nil {
			log.Error().Msgf("compose spec of %s asset %s for update count %d failed: %v", ah.assetGroup, backendUpdate.Meta.Name, ah.syncCounter, err)
			continue
		}

		clusterUpdates.ToUpdate = append(clusterUpdates.ToUpdate, client.Object(&clusterSecret))
		clusterUpdates.UpdateRef = append(clusterUpdates.UpdateRef, backendUpdate.refIndex)
	}

	return clusterUpdates, nil
}

func (ah *AssetsHandler) composeSecretData(clusterSecret *corev1.Secret, backendUpdate *AssetsFromBackend, isPrivate bool) error {
	credentialsSpec := backendUpdate.Spec.(*assetsApi.CredentialsUpdateSpec)
	if credentialsSpec == nil {
		return errors.New("mandatory credential spec is missing")
	}

	switch backendUpdate.Meta.Kind {
	case assetsApi.ASSETKIND_ACCESS_KEY:
		backendSpec := credentialsSpec.AccessKey.Get()

		if backendSpec.HasExistingSecretName() {
			return errors.New("synced access-key should not have existing secret name")
		}

		if convutils.IsEmptyOrNull(backendSpec.AccessKeyId.Get()) {
			return fmt.Errorf("required accessKeyId is missing")
		}

		if convutils.IsEmptyOrNull(backendSpec.SecretAccessKey.Get()) {
			return fmt.Errorf("required secretAccessKey is missing")
		}

		clusterSecret.Data["AccessKeyId"] = []byte(backendSpec.GetAccessKeyId())
		clusterSecret.Data["SecretKey"] = []byte(backendSpec.GetSecretAccessKey())

	case assetsApi.ASSETKIND_PASSWORD:
		backendSpec := credentialsSpec.Password.Get()

		if backendSpec.HasExistingSecretName() {
			return errors.New("synced password should not have existing secret name")
		}

		if convutils.IsEmptyOrNull(backendSpec.User.Get()) {
			return fmt.Errorf("required user is missing")
		}

		if convutils.IsEmptyOrNull(backendSpec.Password.Get()) {
			return fmt.Errorf("required password is missing")
		}

		clusterSecret.Data["User"] = []byte(backendSpec.GetUser())
		clusterSecret.Data["Password"] = []byte(backendSpec.GetPassword())

	case assetsApi.ASSETKIND_DOCKER_REGISTRY:
		backendSpec := credentialsSpec.DockerRegistry.Get()

		if backendSpec.HasExistingSecretName() {
			return errors.New("synced dockerRegistry should not have existing secret name")
		}

		if convutils.IsEmptyOrNull(backendSpec.Url.Get()) {
			return fmt.Errorf("required url is missing")
		}

		if convutils.IsEmptyOrNull(backendSpec.User.Get()) {
			return fmt.Errorf("required user is missing")
		}

		if convutils.IsEmptyOrNull(backendSpec.Password.Get()) {
			return fmt.Errorf("required password is missing")
		}

		authStr := fmt.Sprintf("%s:%s", backendSpec.GetUser(), backendSpec.GetPassword())
		auth := base64.StdEncoding.EncodeToString([]byte(authStr))

		json := fmt.Sprintf("{\"auths\": {\"%s\": {\"username\": \"%s\", \"password\": \"%s\", \"auth\": \"%s\"}}}",
			backendSpec.GetUrl(), backendSpec.GetUser(), backendSpec.GetPassword(), auth)

		clusterSecret.Data[".dockerconfigjson"] = []byte(json)

	case assetsApi.ASSETKIND_GENERIC_SECRET:
		backendSpec := credentialsSpec.GenericSecret.Get()

		if backendSpec.HasExistingSecretName() {
			return errors.New("synced dockerRegistry should not have existing secret name")
		}

		if !backendSpec.HasKeyValuePairs() {
			return fmt.Errorf("generic secret is missing key-value pairs")
		}

		for _, secret := range backendSpec.KeyValuePairs {
			clusterSecret.Data[secret.GetKey()] = []byte(secret.GetValue())
		}
	}

	// If the asset is private, encrypt all secret data
	if isPrivate {
		for key, value := range clusterSecret.Data {
			encryptedValue, err := encryptData(value)
			if err != nil {
				return fmt.Errorf("failed to encrypt value for key %s: %w", constants.RunaiAdsEncryptionKeyEnvVar, err)
			}
			clusterSecret.Data[key] = encryptedValue
		}
	}

	return nil
}

// composeScopeLabel stamps the scope-wide Run:ai label on an in-flight k8s
// object's label map. The resulting label is what a future replication
// controller selects on when deciding which project namespaces should
// receive a copy of the asset — keep the label vocabulary in sync between
// here and that controller.
func composeScopeLabel(labels map[string]string, meta *assetsApi.AssetMetaForSync) {
	switch meta.Scope {
	case assetsApi.SCOPE_TENANT:
		labels[tenantScopeLabel] = boolTrueAsString
	case assetsApi.SCOPE_CLUSTER:
		labels[clusterScopeLabel] = boolTrueAsString
	case assetsApi.SCOPE_DEPARTMENT:
		labels[departmentScopeLabel] = convutils.Deref(meta.DepartmentId.Get(), "")
	case assetsApi.SCOPE_PROJECT:
		labels[projectScopeLabel] = fmt.Sprintf("%d", convutils.Deref(meta.ProjectId.Get(), 0))
	}
}

// getAssetKindLabel maps an asset kind to the string stamped on the
// `run.ai/asset` label of the k8s object. Keep this exhaustive — unknown
// kinds collapse to an empty label which hides them from label-selectors
// (e.g. replication, cleanup) that key off this label.
func getAssetKindLabel(kind assetsApi.AssetKind) string {
	switch kind {
	case assetsApi.ASSETKIND_PVC:
		return pvcKindLabel
	case assetsApi.ASSETKIND_DOCKER_REGISTRY:
		return dockerRegistryKindLabel
	case assetsApi.ASSETKIND_PASSWORD:
		return passwordKindLabel
	case assetsApi.ASSETKIND_ACCESS_KEY:
		return accessKeyKindLabel
	case assetsApi.ASSETKIND_GENERIC_SECRET:
		return genericSecretKindLabel
	case assetsApi.ASSETKIND_RESOURCE_CLAIM_TEMPLATE:
		return resourceClaimTemplateKindLabel
	}
	return ""
}

func (ah *AssetsHandler) getSecretResourceName(assetKind assetsApi.AssetKind, assetName string, assetId string, isPrivate bool) string {
	if isPrivate {
		// Take first 10 characters of the asset ID
		shortId := assetId
		if len(assetId) > 10 {
			shortId = assetId[:10]
		}
		return fmt.Sprintf("%s-%s", assetName, shortId)
	}
	secretName := fmt.Sprintf("%s-%s", assetKind, assetName)
	return strings.ToLower(secretName)
}

func (ah *AssetsHandler) composePvcLabels(labels map[string]string, pvcSpec assetsApi.PVCAssetSpec) {
	isPvcDataSharing := pvcSpec.DataSharing.Get()
	if isPvcDataSharing != nil && *isPvcDataSharing {
		labels[constants.RunaiDataSharingLabel] = boolTrueAsString
	}
}

func encryptData(data []byte) ([]byte, error) {
	if len(encryptionKey) == 0 {
		return nil, fmt.Errorf("encryption key not found in environment variable %s", constants.RunaiAdsEncryptionKeyEnvVar)
	}

	// Create new AES cipher block
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Create nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt data
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}
