package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	openapi "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	clustersApi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/clients/clusters"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/database/dbtesting"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/testutils"
)

var storageClassConfigurationService = NewStorageClassConfigurationAPIService()

var _ = Describe("storage class configuration test", func() {
	var randomScopes *RandomScopes

	BeforeEach(func() {
		router = CreateRouter()
		randomScopes = NewRandomsSet()

		clusters.SetMock(clusters.NewSimpleMock().WithCluster(randomScopes.ClusterId, &clustersApi.DisplayedCluster{}))

		dbtesting.CleanTables()
	})

	Context("permissions", func() {
		It("create storage class configuration", func() {
			ctx := SetContext(randomScopes.TenantIdStr, "test@run.ai")
			storageClassConfiguration := &api.StorageClassConfigurationCreationRequest{
				TenantId:  &randomScopes.TenantId,
				ScopeType: string(openapi.SCOPETYPE_CLUSTER),
				ScopeId:   randomScopes.ClusterId,
				Name:      "test-storage-class",
				Permissions: &api.StorageClassPermissions{
					AllowedForAssets:           convutils.Ref(true),
					AllowedForWorkloads:        convutils.Ref(false),
					AllowedForEphemeralVolumes: convutils.Ref(true),
				},
			}
			resp, err := storageClassConfigurationService.CreateStorageClassConfiguration(ctx, *storageClassConfiguration)
			Expect(err).To(BeNil())
			Expect(resp).ToNot(BeNil())
			storageClassConfigurationResp := resp.Body.(api.StorageClassConfiguration)
			Expect(storageClassConfigurationResp.Name).To(Equal(storageClassConfiguration.Name))
			Expect(storageClassConfigurationResp.Permissions).To(Equal(storageClassConfiguration.Permissions))
			Expect(storageClassConfigurationResp.ScopeType).To(Equal(storageClassConfiguration.ScopeType))
			Expect(storageClassConfigurationResp.ScopeId).To(Equal(storageClassConfiguration.ScopeId))
			Expect(storageClassConfigurationResp.TenantId).To(Equal(*storageClassConfiguration.TenantId))
		})

		It("create storage class configuration and update it", func() {
			ctx := SetContext(randomScopes.TenantIdStr, "test@run.ai")

			storageClassConfiguration := &api.StorageClassConfigurationCreationRequest{
				ScopeType: string(openapi.SCOPETYPE_CLUSTER),
				ScopeId:   randomScopes.ClusterId,
				Name:      "test-storage-class",
				Permissions: &api.StorageClassPermissions{
					AllowedForAssets:           convutils.Ref(true),
					AllowedForEphemeralVolumes: convutils.Ref(true),
					AllowDataSharing:           convutils.Ref(true),
				},
			}

			resp, err := storageClassConfigurationService.CreateStorageClassConfiguration(ctx, *storageClassConfiguration)
			Expect(err).To(BeNil())
			Expect(resp).ToNot(BeNil())
			storageClassConfigurationResp := resp.Body.(api.StorageClassConfiguration)
			Expect(storageClassConfigurationResp.Name).To(Equal(storageClassConfiguration.Name))
			Expect(storageClassConfigurationResp.ScopeType).To(Equal(storageClassConfiguration.ScopeType))
			Expect(storageClassConfigurationResp.ScopeId).To(Equal(storageClassConfiguration.ScopeId))
			Expect(storageClassConfigurationResp.TenantId).To(Equal(randomScopes.TenantId))
			Expect(storageClassConfigurationResp.Permissions).To(Equal(storageClassConfiguration.Permissions))

			patchRequest := api.StorageClassConfigurationPatchRequest{
				Permissions: &api.StorageClassPermissions{
					AllowedForAssets:            convutils.Ref(false),
					AllowedForPersistentVolumes: convutils.Ref(true),
					AllowDataSharing:            convutils.Ref(false),
				},
			}

			storageClassConfigurationResp = patchConfiguration(ctx, storageClassConfigurationResp.Id, patchRequest, randomTenantId)
			Expect(storageClassConfigurationResp.Name).To(Equal(storageClassConfiguration.Name))
			Expect(storageClassConfigurationResp.ScopeType).To(Equal(storageClassConfiguration.ScopeType))
			Expect(storageClassConfigurationResp.ScopeId).To(Equal(storageClassConfiguration.ScopeId))
			Expect(storageClassConfigurationResp.TenantId).To(Equal(randomTenantId))

			Expect(storageClassConfigurationResp.Permissions).NotTo(BeNil())
			Expect(*storageClassConfigurationResp.Permissions.AllowedForAssets).To(Equal(false))
			Expect(*storageClassConfigurationResp.Permissions.AllowedForWorkloads).To(Equal(true))
			Expect(*storageClassConfigurationResp.Permissions.AllowedForEphemeralVolumes).To(Equal(true))
			Expect(*storageClassConfigurationResp.Permissions.AllowedForPersistentVolumes).To(Equal(true))
			Expect(*storageClassConfigurationResp.Permissions.AllowDataSharing).To(Equal(false))
		})
	})

	Context("General API behavior", func() {
		It("create storage class configuration with invalid cluster id", func() {
			ctx := SetContext(randomScopes.TenantIdStr, "test@run.ai")
			storageClassConfiguration := &api.StorageClassConfigurationCreationRequest{
				TenantId:  &randomScopes.TenantId,
				ScopeType: string(openapi.SCOPETYPE_CLUSTER),
				ScopeId:   uuid.NewString(),
				Name:      "test-storage-class",
			}

			resp, err := storageClassConfigurationService.CreateStorageClassConfiguration(ctx, *storageClassConfiguration)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusNotFound))
		})

		It("create storage class configuration and delete it", func() {
			ctx := SetContext(randomScopes.TenantIdStr, "test@run.ai")
			storageClassConfiguration := &api.StorageClassConfigurationCreationRequest{
				TenantId:  &randomScopes.TenantId,
				ScopeType: string(openapi.SCOPETYPE_CLUSTER),
				ScopeId:   randomScopes.ClusterId,
				Name:      "test-storage-class",
			}

			createResp, err := storageClassConfigurationService.CreateStorageClassConfiguration(ctx, *storageClassConfiguration)
			Expect(err).To(BeNil())
			Expect(createResp).ToNot(BeNil())
			storageClassConfigurationResp := createResp.Body.(api.StorageClassConfiguration)
			deleteResp, err := storageClassConfigurationService.DeleteStorageClassConfigurationById(ctx, storageClassConfigurationResp.Id, *storageClassConfiguration.TenantId)
			Expect(err).To(BeNil())
			Expect(deleteResp).ToNot(BeNil())
			Expect(deleteResp.Code).To(Equal(http.StatusNoContent))
			// validate that the storage class configuration was deleted
			listResp, err := storageClassConfigurationService.ListStorageClassConfiguration(ctx, nil, "", "", 0, 20)
			Expect(err).To(BeNil())
			Expect(listResp).ToNot(BeNil())
			listRespBody := listResp.Body.(api.StorageClassConfigurationListResponse)
			Expect(len(listRespBody.StorageClassConfigurations)).To(Equal(0))
		})

		It("create 2 storage class configurations and validate in list", func() {
			ctx := SetContext(randomScopes.TenantIdStr, "test@run.ai")
			storageClassConfiguration1 := &api.StorageClassConfigurationCreationRequest{
				TenantId:  &randomScopes.TenantId,
				ScopeType: string(openapi.SCOPETYPE_CLUSTER),
				ScopeId:   randomScopes.ClusterId,
				Name:      "test-storage-class-1",
				Permissions: &api.StorageClassPermissions{
					AllowedForAssets:            convutils.Ref(true),
					AllowedForWorkloads:         convutils.Ref(false),
					AllowedForEphemeralVolumes:  convutils.Ref(true),
					AllowedForPersistentVolumes: convutils.Ref(false),
				},
			}

			storageClassConfiguration2 := &api.StorageClassConfigurationCreationRequest{
				TenantId:  &randomScopes.TenantId,
				ScopeType: string(openapi.SCOPETYPE_CLUSTER),
				ScopeId:   randomScopes.ClusterId,
				Name:      "test-storage-class-2",
				Permissions: &api.StorageClassPermissions{
					AllowedForAssets:            convutils.Ref(false),
					AllowedForWorkloads:         convutils.Ref(true),
					AllowedForEphemeralVolumes:  convutils.Ref(false),
					AllowedForPersistentVolumes: convutils.Ref(true),
				},
			}

			_, err := storageClassConfigurationService.CreateStorageClassConfiguration(ctx, *storageClassConfiguration1)
			Expect(err).To(BeNil())
			_, err = storageClassConfigurationService.CreateStorageClassConfiguration(ctx, *storageClassConfiguration2)
			Expect(err).To(BeNil())

			listResp, err := storageClassConfigurationService.ListStorageClassConfiguration(ctx, nil, "name", "", 0, 20)
			Expect(err).To(BeNil())
			Expect(listResp).ToNot(BeNil())
			listRespBody := listResp.Body.(api.StorageClassConfigurationListResponse)
			Expect(len(listRespBody.StorageClassConfigurations)).To(Equal(2))

			Expect(listRespBody.StorageClassConfigurations[0].Name).To(Equal(storageClassConfiguration1.Name))
			Expect(listRespBody.StorageClassConfigurations[1].Name).To(Equal(storageClassConfiguration2.Name))
			Expect(listRespBody.StorageClassConfigurations[0].Permissions).To(Equal(storageClassConfiguration1.Permissions))
			Expect(listRespBody.StorageClassConfigurations[1].Permissions).To(Equal(storageClassConfiguration2.Permissions))
		})

		It("create 2 storage class configurations and validate in list with filter", func() {
			ctx := SetContext(randomScopes.TenantIdStr, "test@run.ai")
			storageClassConfiguration1 := &api.StorageClassConfigurationCreationRequest{
				TenantId:  &randomScopes.TenantId,
				ScopeType: string(openapi.SCOPETYPE_CLUSTER),
				ScopeId:   randomScopes.ClusterId,
				Name:      "test-storage-class-1",
				Permissions: &api.StorageClassPermissions{
					AllowedForAssets:            convutils.Ref(true),
					AllowedForWorkloads:         convutils.Ref(false),
					AllowedForEphemeralVolumes:  convutils.Ref(true),
					AllowedForPersistentVolumes: convutils.Ref(false),
				},
			}

			storageClassConfiguration2 := &api.StorageClassConfigurationCreationRequest{
				TenantId:  &randomScopes.TenantId,
				ScopeType: string(openapi.SCOPETYPE_CLUSTER),
				ScopeId:   randomScopes.ClusterId,
				Name:      "test-storage-class-2",
				Permissions: &api.StorageClassPermissions{
					AllowedForAssets:            convutils.Ref(true),
					AllowedForWorkloads:         convutils.Ref(false),
					AllowedForEphemeralVolumes:  convutils.Ref(true),
					AllowedForPersistentVolumes: convutils.Ref(false),
				},
			}

			_, err := storageClassConfigurationService.CreateStorageClassConfiguration(ctx, *storageClassConfiguration1)
			Expect(err).To(BeNil())
			_, err = storageClassConfigurationService.CreateStorageClassConfiguration(ctx, *storageClassConfiguration2)
			Expect(err).To(BeNil())

			filterBy := []string{fmt.Sprintf("name==%s", storageClassConfiguration1.Name)}
			listResp, err := storageClassConfigurationService.ListStorageClassConfiguration(ctx, filterBy, "", "", 0, 20)
			Expect(err).To(BeNil())
			Expect(listResp).ToNot(BeNil())

			listRespBody := listResp.Body.(api.StorageClassConfigurationListResponse)
			Expect(len(listRespBody.StorageClassConfigurations)).To(Equal(1))
			Expect(listRespBody.StorageClassConfigurations[0].Name).To(Equal(storageClassConfiguration1.Name))
			Expect(listRespBody.StorageClassConfigurations[0].Permissions).To(Equal(storageClassConfiguration1.Permissions))
		})
	})

	Context("customization patching", func() {
		var ctx context.Context
		var currentId string

		BeforeEach(func() {
			ctx = SetContext(randomScopes.TenantIdStr, "test@run.ai")

			storageClassConfiguration := &api.StorageClassConfigurationCreationRequest{
				ScopeType: string(openapi.SCOPETYPE_CLUSTER),
				ScopeId:   randomScopes.ClusterId,
				Name:      "test-storage-class",
			}

			resp, err := storageClassConfigurationService.CreateStorageClassConfiguration(ctx, *storageClassConfiguration)
			Expect(err).To(BeNil())
			Expect(resp).ToNot(BeNil())

			storageClassConfigurationResp := resp.Body.(api.StorageClassConfiguration)
			Expect(storageClassConfigurationResp.Name).To(Equal(storageClassConfiguration.Name))
			Expect(storageClassConfigurationResp.ScopeType).To(Equal(storageClassConfiguration.ScopeType))
			Expect(storageClassConfigurationResp.ScopeId).To(Equal(storageClassConfiguration.ScopeId))
			Expect(storageClassConfigurationResp.TenantId).To(Equal(randomScopes.TenantId))
			Expect(storageClassConfigurationResp.Permissions).To(BeNil())
			Expect(storageClassConfigurationResp.Customization).To(BeNil())

			currentId = storageClassConfigurationResp.Id
		})

		It("empty update patching", func() {
			//
			//   empty customization does not cause any change
			//
			patchRequest := api.StorageClassConfigurationPatchRequest{
				Customization: &api.StorageClassConfigurationPatchRequestCustomization{},
			}
			result := patchConfiguration(ctx, currentId, patchRequest, randomTenantId)
			Expect(result.Customization).To(BeNil())

			patchRequest = api.StorageClassConfigurationPatchRequest{
				Customization: &api.StorageClassConfigurationPatchRequestCustomization{
					AccessMode: &api.AccessModeCustomization{},
					VolumeMode: &api.VolumeModeCustomization{},
					ClaimSize:  &api.ClaimSizeCustomization{},
				},
			}
			result = patchConfiguration(ctx, currentId, patchRequest, randomTenantId)
			Expect(result.Customization).To(BeNil())
		})

		It("update replaces the existing fields", func() {
			updateRequest := api.StorageClassConfigurationUpdateRequest{
				Customization: &api.StorageClassCustomization{
					ClaimSize: &api.ClaimSizeCustomization{
						SupportedUnits: []string{
							"TB",
						},
					},
					Attributes: []api.PvcAddedAttribute{
						{
							Key:             "dnsname",
							Format:          convutils.Ref(api.PVCADDEDATTRIBUTEFORMAT_TEXT),
							ValidationError: convutils.Ref("Invalid DNS name"),
						},
					},
				},
			}
			result := updateConfiguration(ctx, currentId, updateRequest, randomTenantId)
			Expect(*result.Customization).To(Equal(*updateRequest.Customization))
		})

		It("update to empty", func() {
			updateRequest := api.StorageClassConfigurationUpdateRequest{}
			result := updateConfiguration(ctx, currentId, updateRequest, randomTenantId)
			Expect(result.Permissions).To(BeNil())
			Expect(result.Customization).To(BeNil())
		})

		It("attributes patching - missing format", func() {
			updateRequest := api.StorageClassConfigurationUpdateRequest{
				Customization: &api.StorageClassCustomization{
					Attributes: []api.PvcAddedAttribute{
						{
							Key: "a",
						},
					},
				},
			}
			updateConfigurationError(ctx, currentId, updateRequest, randomTenantId, http.StatusBadRequest, "attribute 'a': mandatory format is not provided")
		})

		It("attributes patching - min/max range", func() {
			updateRequest := api.StorageClassConfigurationUpdateRequest{
				Customization: &api.StorageClassCustomization{
					Attributes: []api.PvcAddedAttribute{
						{
							Key:      "a",
							Format:   convutils.Ref(api.PVCADDEDATTRIBUTEFORMAT_NUMBER),
							MinValue: convutils.Ref(float32(5.5)),
							MaxValue: convutils.Ref(float32(5.4)),
						},
					},
				},
			}
			updateConfigurationError(ctx, currentId, updateRequest, randomTenantId, http.StatusBadRequest, "attribute 'a': minValue should be no less than maxValue")
		})

		It("attributes patching - step ok", func() {
			updateRequest := api.StorageClassConfigurationUpdateRequest{
				Customization: &api.StorageClassCustomization{
					Attributes: []api.PvcAddedAttribute{
						{
							Key:          "a",
							Format:       convutils.Ref(api.PVCADDEDATTRIBUTEFORMAT_NUMBER),
							MinValue:     convutils.Ref(float32(6)),
							MaxValue:     convutils.Ref(float32(9)),
							Step:         convutils.Ref(float32(1.5)),
							DefaultValue: convutils.Ref("7.5"),
						},
					},
				},
			}
			updateConfiguration(ctx, currentId, updateRequest, randomTenantId)
		})

		It("attributes patching - step not ok", func() {
			updateRequest := api.StorageClassConfigurationUpdateRequest{
				Customization: &api.StorageClassCustomization{
					Attributes: []api.PvcAddedAttribute{
						{
							Key:          "a",
							Format:       convutils.Ref(api.PVCADDEDATTRIBUTEFORMAT_NUMBER),
							MinValue:     convutils.Ref(float32(6)),
							MaxValue:     convutils.Ref(float32(9)),
							Step:         convutils.Ref(float32(1.5)),
							DefaultValue: convutils.Ref("7.6"),
						},
					},
				},
			}
			updateConfigurationError(ctx, currentId, updateRequest, randomTenantId, http.StatusBadRequest, "default of attribute 'a': value does not comply step of 1.5 from 6")
		})

		It("attributes patching - default small", func() {
			updateRequest := api.StorageClassConfigurationUpdateRequest{
				Customization: &api.StorageClassCustomization{
					Attributes: []api.PvcAddedAttribute{
						{
							Key:          "a",
							Format:       convutils.Ref(api.PVCADDEDATTRIBUTEFORMAT_NUMBER),
							MinValue:     convutils.Ref(float32(6)),
							DefaultValue: convutils.Ref("5.9"),
						},
					},
				},
			}
			updateConfigurationError(ctx, currentId, updateRequest, randomTenantId, http.StatusBadRequest, "default of attribute 'a': value must be no less then 6")
		})

		It("attributes patching - default big", func() {
			updateRequest := api.StorageClassConfigurationUpdateRequest{
				Customization: &api.StorageClassCustomization{
					Attributes: []api.PvcAddedAttribute{
						{
							Key:          "a",
							Format:       convutils.Ref(api.PVCADDEDATTRIBUTEFORMAT_NUMBER),
							MaxValue:     convutils.Ref(float32(6)),
							DefaultValue: convutils.Ref("6.1"),
						},
					},
				},
			}
			updateConfigurationError(ctx, currentId, updateRequest, randomTenantId, http.StatusBadRequest, "default of attribute 'a': value must be no greater then 6")
		})

		It("attributes patching - inconsistent min", func() {
			updateRequest := api.StorageClassConfigurationUpdateRequest{
				Customization: &api.StorageClassCustomization{
					Attributes: []api.PvcAddedAttribute{
						{
							Key:      "a",
							Format:   convutils.Ref(api.PVCADDEDATTRIBUTEFORMAT_TEXT),
							MinValue: convutils.Ref(float32(5.3)),
						},
					},
				},
			}
			updateConfigurationError(ctx, currentId, updateRequest, randomTenantId, http.StatusBadRequest, "attribute 'a': minValue, maxValue and step are only supported for number attributes")
		})

		It("attributes patching - inconsistent regexp", func() {
			updateRequest := api.StorageClassConfigurationUpdateRequest{
				Customization: &api.StorageClassCustomization{
					Attributes: []api.PvcAddedAttribute{
						{
							Key:              "a",
							Format:           convutils.Ref(api.PVCADDEDATTRIBUTEFORMAT_BOOLEAN),
							ValidationRegexp: convutils.Ref("xxx"),
						},
					},
				},
			}
			updateConfigurationError(ctx, currentId, updateRequest, randomTenantId, http.StatusBadRequest, "attribute 'a': validationRegexp is only supported for text attributes")
		})

		It("attributes patching - correct boolean", func() {
			updateRequest := api.StorageClassConfigurationUpdateRequest{
				Customization: &api.StorageClassCustomization{
					Attributes: []api.PvcAddedAttribute{
						{
							Key:          "a",
							Format:       convutils.Ref(api.PVCADDEDATTRIBUTEFORMAT_BOOLEAN),
							DefaultValue: convutils.Ref("true"),
						},
					},
				},
			}
			updateConfiguration(ctx, currentId, updateRequest, randomTenantId)
		})

		It("attributes patching - correct boolean", func() {
			updateRequest := api.StorageClassConfigurationUpdateRequest{
				Customization: &api.StorageClassCustomization{
					Attributes: []api.PvcAddedAttribute{
						{
							Key:          "a",
							Format:       convutils.Ref(api.PVCADDEDATTRIBUTEFORMAT_BOOLEAN),
							DefaultValue: convutils.Ref("false"),
						},
					},
				},
			}
			updateConfiguration(ctx, currentId, updateRequest, randomTenantId)
		})

		It("attributes patching - incorrect number", func() {
			updateRequest := api.StorageClassConfigurationUpdateRequest{
				Customization: &api.StorageClassCustomization{
					Attributes: []api.PvcAddedAttribute{
						{
							Key:          "a",
							Format:       convutils.Ref(api.PVCADDEDATTRIBUTEFORMAT_NUMBER),
							DefaultValue: convutils.Ref("xxx"),
						},
					},
				},
			}
			updateConfigurationError(ctx, currentId, updateRequest, randomTenantId, http.StatusBadRequest, "default of attribute 'a': invalid value for numeric attribute")
		})

		It("attributes patching - incorrect regexp", func() {
			updateRequest := api.StorageClassConfigurationUpdateRequest{
				Customization: &api.StorageClassCustomization{
					Attributes: []api.PvcAddedAttribute{
						{
							Key:              "a",
							Format:           convutils.Ref(api.PVCADDEDATTRIBUTEFORMAT_TEXT),
							ValidationRegexp: convutils.Ref("["),
							DefaultValue:     convutils.Ref("xxx"),
						},
					},
				},
			}
			updateConfigurationError(ctx, currentId, updateRequest, randomTenantId, http.StatusBadRequest, "attribute 'a': invalid validationRegexp: [")
		})

		It("attributes patching - value does not match regexp", func() {
			updateRequest := api.StorageClassConfigurationUpdateRequest{
				Customization: &api.StorageClassCustomization{
					Attributes: []api.PvcAddedAttribute{
						{
							Key:              "a",
							Format:           convutils.Ref(api.PVCADDEDATTRIBUTEFORMAT_TEXT),
							ValidationRegexp: convutils.Ref("^x+$"),
							DefaultValue:     convutils.Ref("xyx"),
						},
					},
				},
			}
			updateConfigurationError(ctx, currentId, updateRequest, randomTenantId, http.StatusBadRequest, "default of attribute 'a': malformed value")
		})

		It("update + patch", func() {
			updateRequest := api.StorageClassConfigurationUpdateRequest{
				Customization: &api.StorageClassCustomization{
					Attributes: []api.PvcAddedAttribute{
						{
							Key:    "a",
							Format: convutils.Ref(api.PVCADDEDATTRIBUTEFORMAT_BOOLEAN),
						},
						{
							Key:    "b",
							Format: convutils.Ref(api.PVCADDEDATTRIBUTEFORMAT_TEXT),
						},
						{
							Key:    "c",
							Format: convutils.Ref(api.PVCADDEDATTRIBUTEFORMAT_NUMBER),
						},
					},
				},
			}
			result := updateConfiguration(ctx, currentId, updateRequest, randomTenantId)
			Expect(result.Customization).To(Equal(updateRequest.Customization))

			patchRequest := api.StorageClassConfigurationPatchRequest{
				Customization: &api.StorageClassConfigurationPatchRequestCustomization{
					Attributes: []api.PvcAddedAttribute{
						{
							Key:             "c",
							Format:          convutils.Ref(api.PVCADDEDATTRIBUTEFORMAT_TEXT),
							ValidationError: convutils.Ref("error of c"),
						},
						{
							Key:             "d",
							Format:          convutils.Ref(api.PVCADDEDATTRIBUTEFORMAT_NUMBER),
							ValidationError: convutils.Ref("error of d"),
						},
					},
					AttributesToRemove: []string{
						"b",
					},
				},
			}

			expected := &api.StorageClassCustomization{
				Attributes: []api.PvcAddedAttribute{
					{
						Key:    "a",
						Format: convutils.Ref(api.PVCADDEDATTRIBUTEFORMAT_BOOLEAN),
					},
					{
						Key:             "c",
						Format:          convutils.Ref(api.PVCADDEDATTRIBUTEFORMAT_TEXT),
						ValidationError: convutils.Ref("error of c"),
					},
					{
						Key:             "d",
						Format:          convutils.Ref(api.PVCADDEDATTRIBUTEFORMAT_NUMBER),
						ValidationError: convutils.Ref("error of d"),
					},
				},
			}

			result = patchConfiguration(ctx, currentId, patchRequest, randomTenantId)
			testutils.Expect(result.Customization).ToEqual(expected)
		})

		It("update + patch causes error", func() {
			updateRequest := api.StorageClassConfigurationUpdateRequest{
				Customization: &api.StorageClassCustomization{
					Attributes: []api.PvcAddedAttribute{
						{
							Key:    "a",
							Format: convutils.Ref(api.PVCADDEDATTRIBUTEFORMAT_BOOLEAN),
						},
					},
				},
			}
			result := updateConfiguration(ctx, currentId, updateRequest, randomTenantId)
			Expect(result.Customization).To(Equal(updateRequest.Customization))

			patchRequest := api.StorageClassConfigurationPatchRequest{
				Customization: &api.StorageClassConfigurationPatchRequestCustomization{
					Attributes: []api.PvcAddedAttribute{
						{
							Key: "a",
						},
					},
					AttributesToRemove: []string{
						"b",
					},
				},
			}

			patchConfigurationError(ctx, currentId, patchRequest, randomTenantId, http.StatusBadRequest, "attribute 'a': mandatory format is not provided")
		})

		It("gradual build the customization", func() {
			patchRequest := api.StorageClassConfigurationPatchRequest{
				Customization: &api.StorageClassConfigurationPatchRequestCustomization{
					AccessMode: &api.AccessModeCustomization{
						Default: &api.PvcAccessModes{
							ReadOnlyMany: convutils.Ref(true),
						},
						SupportedValues: &api.PvcSupportedAccessModes{
							ReadWriteMany: convutils.Ref(false),
						},
					},
					VolumeMode: &api.VolumeModeCustomization{
						Default: convutils.Ref(api.PVCVOLUMEMODE_BLOCK),
					},
				},
			}
			result := patchConfiguration(ctx, currentId, patchRequest, randomTenantId)
			Expect(result.Customization).NotTo(BeNil())

			expectedResp := api.StorageClassCustomization{
				AccessMode: &api.AccessModeCustomization{
					Required: convutils.Ref(false), // unspecified = false
					Default: &api.PvcAccessModes{
						ReadOnlyMany:  convutils.Ref(true),
						ReadWriteOnce: convutils.Ref(false),
						ReadWriteMany: convutils.Ref(false), // unspecified = false (in defaults)

					},
					SupportedValues: &api.PvcSupportedAccessModes{
						ReadOnlyMany:  convutils.Ref(true), // unspecified = true (in supported)
						ReadWriteOnce: convutils.Ref(true), // unspecified = true
						ReadWriteMany: convutils.Ref(false),
					},
				},
				VolumeMode: &api.VolumeModeCustomization{
					Required: convutils.Ref(false), // unspecified = false.
					Default:  convutils.Ref(api.PVCVOLUMEMODE_BLOCK),
					SupportedValues: &api.VolumeModeCustomizationSupportedValues{
						Filesystem: convutils.Ref(true), // unspecified = true
						Block:      convutils.Ref(true), // unspecified = true
					},
				},
			}

			Expect(*result.Customization).To(Equal(expectedResp))

			//
			//   add claim size info and modify some flags
			//
			patchRequest = api.StorageClassConfigurationPatchRequest{
				Customization: &api.StorageClassConfigurationPatchRequestCustomization{
					AccessMode: &api.AccessModeCustomization{
						Default: &api.PvcAccessModes{
							ReadWriteMany: convutils.Ref(true),
						},
						SupportedValues: &api.PvcSupportedAccessModes{
							ReadOnlyMany: convutils.Ref(false),
						},
					},
					VolumeMode: &api.VolumeModeCustomization{
						Required: convutils.Ref(true),
						Default:  convutils.Ref(api.PVCVOLUMEMODE_FILESYSTEM),
						SupportedValues: &api.VolumeModeCustomizationSupportedValues{
							Block: convutils.Ref(false),
						},
					},
					ClaimSize: &api.ClaimSizeCustomization{
						Min:  convutils.Ref("2G"),
						Step: convutils.Ref("500M"),
					},
				},
			}
			result = patchConfiguration(ctx, currentId, patchRequest, randomTenantId)
			Expect(result.Customization).NotTo(BeNil())

			expectedResp = api.StorageClassCustomization{
				AccessMode: &api.AccessModeCustomization{
					Required: convutils.Ref(false), // unspecified = false
					Default: &api.PvcAccessModes{
						ReadOnlyMany:  convutils.Ref(false), // was true, but became unsupported so changed to false
						ReadWriteOnce: convutils.Ref(false),
						ReadWriteMany: convutils.Ref(false), // changed to true, but it is unsupported so remains false
					},
					SupportedValues: &api.PvcSupportedAccessModes{
						ReadOnlyMany:  convutils.Ref(false), // changed to false
						ReadWriteOnce: convutils.Ref(true),  // unspecified = true
						ReadWriteMany: convutils.Ref(false),
					},
				},
				VolumeMode: &api.VolumeModeCustomization{
					Required: convutils.Ref(true),                         // changed to true
					Default:  convutils.Ref(api.PVCVOLUMEMODE_FILESYSTEM), // default changed
					SupportedValues: &api.VolumeModeCustomizationSupportedValues{
						Filesystem: convutils.Ref(true),  // unspecified = true
						Block:      convutils.Ref(false), // change to false
					},
				},
				ClaimSize: &api.ClaimSizeCustomization{
					Min:            convutils.Ref("2G"),
					Step:           convutils.Ref("500M"),
					SupportedUnits: []string{"MB", "MiB", "GB", "GiB", "TB", "TiB"}, // all supported
				},
			}

			Expect(*result.Customization).To(Equal(expectedResp))

			//
			// add max to the claim size
			// modify the list of supported units
			//
			patchRequest = api.StorageClassConfigurationPatchRequest{
				Customization: &api.StorageClassConfigurationPatchRequestCustomization{
					ClaimSize: &api.ClaimSizeCustomization{
						Max:            convutils.Ref("8G"),
						SupportedUnits: []string{"MB", "MiB", "GB", "GiB"}, // all supported
					},
				},
			}
			result = patchConfiguration(ctx, currentId, patchRequest, randomTenantId)
			Expect(result.Customization).NotTo(BeNil())

			expectedResp = api.StorageClassCustomization{
				AccessMode: &api.AccessModeCustomization{
					Required: convutils.Ref(false), // unspecified = false
					Default: &api.PvcAccessModes{
						ReadOnlyMany:  convutils.Ref(false),
						ReadWriteOnce: convutils.Ref(false),
						ReadWriteMany: convutils.Ref(false),
					},
					SupportedValues: &api.PvcSupportedAccessModes{
						ReadOnlyMany:  convutils.Ref(false),
						ReadWriteOnce: convutils.Ref(true),
						ReadWriteMany: convutils.Ref(false),
					},
				},
				VolumeMode: &api.VolumeModeCustomization{
					Required: convutils.Ref(true),
					Default:  convutils.Ref(api.PVCVOLUMEMODE_FILESYSTEM),
					SupportedValues: &api.VolumeModeCustomizationSupportedValues{
						Filesystem: convutils.Ref(true),
						Block:      convutils.Ref(false),
					},
				},
				ClaimSize: &api.ClaimSizeCustomization{
					Min:            convutils.Ref("2G"),
					Step:           convutils.Ref("500M"),
					Max:            convutils.Ref("8G"),                // added
					SupportedUnits: []string{"MB", "MiB", "GB", "GiB"}, // all supported
				},
			}

			Expect(*result.Customization).To(Equal(expectedResp))
		})

		It("call rejected when default volume mode not supported", func() {
			patchRequest := api.StorageClassConfigurationPatchRequest{
				Customization: &api.StorageClassConfigurationPatchRequestCustomization{
					VolumeMode: &api.VolumeModeCustomization{
						Default: convutils.Ref(api.PVCVOLUMEMODE_FILESYSTEM),
					},
				},
			}
			result := patchConfiguration(ctx, currentId, patchRequest, randomTenantId)
			Expect(result.Customization).NotTo(BeNil())

			expectedResp := api.StorageClassCustomization{
				VolumeMode: &api.VolumeModeCustomization{
					Required: convutils.Ref(false), // unspecified = false.
					Default:  convutils.Ref(api.PVCVOLUMEMODE_FILESYSTEM),
					SupportedValues: &api.VolumeModeCustomizationSupportedValues{
						Filesystem: convutils.Ref(true), // unspecified = true
						Block:      convutils.Ref(true), // unspecified = true
					},
				},
			}

			Expect(*result.Customization).To(Equal(expectedResp))

			//
			//  make filesystem not supported
			//
			patchRequest = api.StorageClassConfigurationPatchRequest{
				Customization: &api.StorageClassConfigurationPatchRequestCustomization{
					VolumeMode: &api.VolumeModeCustomization{
						SupportedValues: &api.VolumeModeCustomizationSupportedValues{
							Filesystem: convutils.Ref(false),
						},
					},
				},
			}
			patchConfigurationError(ctx, currentId, patchRequest, randomTenantId, http.StatusBadRequest, "default value 'Filesystem' is not marked as supported")
		})

		It("remove all + required", func() {
			patchRequest := api.StorageClassConfigurationPatchRequest{
				Customization: &api.StorageClassConfigurationPatchRequestCustomization{
					VolumeMode: &api.VolumeModeCustomization{
						SupportedValues: &api.VolumeModeCustomizationSupportedValues{
							Filesystem: convutils.Ref(false),
						},
					},
				},
			}
			patchConfiguration(ctx, currentId, patchRequest, randomTenantId)

			patchRequest = api.StorageClassConfigurationPatchRequest{
				Customization: &api.StorageClassConfigurationPatchRequestCustomization{
					VolumeMode: &api.VolumeModeCustomization{
						Required: convutils.Ref(true),
						SupportedValues: &api.VolumeModeCustomizationSupportedValues{
							Block: convutils.Ref(false),
						},
					},
				},
			}

			patchConfigurationError(ctx, currentId, patchRequest, randomTenantId, http.StatusBadRequest, "at least one volumeMode should be supported when volumeMode is required")
		})

		It("remove all access modes + required", func() {
			patchRequest := api.StorageClassConfigurationPatchRequest{
				Customization: &api.StorageClassConfigurationPatchRequestCustomization{
					AccessMode: &api.AccessModeCustomization{
						SupportedValues: &api.PvcSupportedAccessModes{
							ReadWriteOnce: convutils.Ref(false),
							ReadWriteMany: convutils.Ref(false),
							ReadOnlyMany:  convutils.Ref(false),
						},
					},
				},
			}
			patchConfiguration(ctx, currentId, patchRequest, randomTenantId)

			patchRequest = api.StorageClassConfigurationPatchRequest{
				Customization: &api.StorageClassConfigurationPatchRequestCustomization{
					AccessMode: &api.AccessModeCustomization{
						Required: convutils.Ref(true),
					},
				},
			}

			patchConfigurationError(ctx, currentId, patchRequest, randomTenantId, http.StatusBadRequest, "at least one accessMode should be supported when accessMode is required")
		})

		It("validation of units becoming not supported", func() {
			patchRequest := api.StorageClassConfigurationPatchRequest{
				Customization: &api.StorageClassConfigurationPatchRequestCustomization{
					ClaimSize: &api.ClaimSizeCustomization{
						Default: convutils.Ref("10M"),
					},
				},
			}
			result := patchConfiguration(ctx, currentId, patchRequest, randomTenantId)
			Expect(result.Customization).NotTo(BeNil())

			expectedResp := api.StorageClassCustomization{
				ClaimSize: &api.ClaimSizeCustomization{
					Default:        convutils.Ref("10M"),
					SupportedUnits: []string{"MB", "MiB", "GB", "GiB", "TB", "TiB"},
				},
			}

			Expect(*result.Customization).To(Equal(expectedResp))

			//
			//  remove MB from supported units
			//
			patchRequest = api.StorageClassConfigurationPatchRequest{
				Customization: &api.StorageClassConfigurationPatchRequestCustomization{
					ClaimSize: &api.ClaimSizeCustomization{
						SupportedUnits: []string{"GB", "GiB", "TB", "TiB"},
					}},
			}
			patchConfigurationError(ctx, currentId, patchRequest, randomTenantId, http.StatusBadRequest, "default: unit suffix 'M' is not supported")
		})

		It("validation of units becoming not supported", func() {
			patchRequest := api.StorageClassConfigurationPatchRequest{
				Customization: &api.StorageClassConfigurationPatchRequestCustomization{
					ClaimSize: &api.ClaimSizeCustomization{
						Default: convutils.Ref("10M"),
					},
				},
			}
			result := patchConfiguration(ctx, currentId, patchRequest, randomTenantId)
			Expect(result.Customization).NotTo(BeNil())

			expectedResp := api.StorageClassCustomization{
				ClaimSize: &api.ClaimSizeCustomization{
					Default:        convutils.Ref("10M"),
					SupportedUnits: []string{"MB", "MiB", "GB", "GiB", "TB", "TiB"},
				},
			}

			Expect(*result.Customization).To(Equal(expectedResp))

			//
			//  remove MB from supported units
			//
			patchRequest = api.StorageClassConfigurationPatchRequest{
				Customization: &api.StorageClassConfigurationPatchRequestCustomization{
					ClaimSize: &api.ClaimSizeCustomization{
						SupportedUnits: []string{"GB", "GiB", "TB", "TiB"},
					}},
			}
			patchConfigurationError(ctx, currentId, patchRequest, randomTenantId, http.StatusBadRequest, "default: unit suffix 'M' is not supported")

			//
			//  remove all units
			//
			patchRequest = api.StorageClassConfigurationPatchRequest{
				Customization: &api.StorageClassConfigurationPatchRequestCustomization{
					ClaimSize: &api.ClaimSizeCustomization{
						SupportedUnits: []string{},
					}},
			}
			patchConfigurationError(ctx, currentId, patchRequest, randomTenantId, http.StatusBadRequest, "claimSize: at least one unit should be supported")
		})

		It("validation of min/max/step", func() {
			patchRequest := api.StorageClassConfigurationPatchRequest{
				Customization: &api.StorageClassConfigurationPatchRequestCustomization{
					ClaimSize: &api.ClaimSizeCustomization{
						Min: convutils.Ref("10G"),
						Max: convutils.Ref("2G"),
					},
				},
			}
			patchConfigurationError(ctx, currentId, patchRequest, randomTenantId, http.StatusBadRequest, "claimSize: max should be no less than min")

			patchRequest = api.StorageClassConfigurationPatchRequest{
				Customization: &api.StorageClassConfigurationPatchRequestCustomization{
					ClaimSize: &api.ClaimSizeCustomization{
						Min:     convutils.Ref("2G"),
						Max:     convutils.Ref("10G"),
						Default: convutils.Ref("1900M"),
					},
				},
			}
			patchConfigurationError(ctx, currentId, patchRequest, randomTenantId, http.StatusBadRequest, "claimSize: default should be no less than min")

			patchRequest = api.StorageClassConfigurationPatchRequest{
				Customization: &api.StorageClassConfigurationPatchRequestCustomization{
					ClaimSize: &api.ClaimSizeCustomization{
						Min:  convutils.Ref("2G"),
						Max:  convutils.Ref("10G"),
						Step: convutils.Ref("8100M"),
					},
				},
			}
			patchConfigurationError(ctx, currentId, patchRequest, randomTenantId, http.StatusBadRequest, "claimSize: invalid value for step: should be no greater than the difference between max and min")

			patchRequest = api.StorageClassConfigurationPatchRequest{
				Customization: &api.StorageClassConfigurationPatchRequestCustomization{
					ClaimSize: &api.ClaimSizeCustomization{
						Min:     convutils.Ref("2G"),
						Max:     convutils.Ref("10G"),
						Default: convutils.Ref("10001M"),
					},
				},
			}
			patchConfigurationError(ctx, currentId, patchRequest, randomTenantId, http.StatusBadRequest, "claimSize: default should be no greater than max")

			patchRequest = api.StorageClassConfigurationPatchRequest{
				Customization: &api.StorageClassConfigurationPatchRequestCustomization{
					ClaimSize: &api.ClaimSizeCustomization{
						Min:     convutils.Ref("2G"),
						Max:     convutils.Ref("10G"),
						Step:    convutils.Ref("500M"),
						Default: convutils.Ref("2400M"),
					},
				},
			}
			patchConfigurationError(ctx, currentId, patchRequest, randomTenantId, http.StatusBadRequest, "claimSize: default value is not aligned with 500M steps from the minimal value 2G")

			patchRequest = api.StorageClassConfigurationPatchRequest{
				Customization: &api.StorageClassConfigurationPatchRequestCustomization{
					ClaimSize: &api.ClaimSizeCustomization{
						Max:     convutils.Ref("10G"),
						Step:    convutils.Ref("500M"),
						Default: convutils.Ref("600M"),
					},
				},
			}
			patchConfigurationError(ctx, currentId, patchRequest, randomTenantId, http.StatusBadRequest, "claimSize: default value is not aligned with steps of 500M")

			patchRequest = api.StorageClassConfigurationPatchRequest{
				Customization: &api.StorageClassConfigurationPatchRequestCustomization{
					ClaimSize: &api.ClaimSizeCustomization{
						Max: convutils.Ref("0M"),
					},
				},
			}
			patchConfigurationError(ctx, currentId, patchRequest, randomTenantId, http.StatusBadRequest, "claimSize: default, min, max and step have positive values")

			patchRequest = api.StorageClassConfigurationPatchRequest{
				Customization: &api.StorageClassConfigurationPatchRequestCustomization{
					ClaimSize: &api.ClaimSizeCustomization{
						Default: convutils.Ref("0M"),
					},
				},
			}
			patchConfigurationError(ctx, currentId, patchRequest, randomTenantId, http.StatusBadRequest, "claimSize: default, min, max and step have positive values")

			patchRequest = api.StorageClassConfigurationPatchRequest{
				Customization: &api.StorageClassConfigurationPatchRequestCustomization{
					ClaimSize: &api.ClaimSizeCustomization{
						Min:     convutils.Ref("10G"),
						Max:     convutils.Ref("20G"),
						Step:    convutils.Ref("500M"),
						Default: convutils.Ref("11500M"),
					},
				},
			}

			result := patchConfiguration(ctx, currentId, patchRequest, randomTenantId)
			Expect(result.Customization).NotTo(BeNil())

			expectedResp := api.StorageClassCustomization{
				ClaimSize: &api.ClaimSizeCustomization{
					Min:            convutils.Ref("10G"),
					Max:            convutils.Ref("20G"),
					Step:           convutils.Ref("500M"),
					Default:        convutils.Ref("11500M"),
					SupportedUnits: []string{"MB", "MiB", "GB", "GiB", "TB", "TiB"},
				},
			}

			Expect(*result.Customization).To(Equal(expectedResp))
		})
	})
})

func patchConfiguration(ctx context.Context, id string, patchRequest api.StorageClassConfigurationPatchRequest,
	tenantId int32) api.StorageClassConfiguration {
	resp, err := storageClassConfigurationService.PatchStorageClassConfigurationById(ctx,
		id, patchRequest, tenantId)
	Expect(err).To(BeNil())
	Expect(resp).ToNot(BeNil())
	Expect(resp.Code).To(Equal(http.StatusOK), fmt.Sprintf("%+v", resp.Body))
	return resp.Body.(api.StorageClassConfiguration)
}

func patchConfigurationError(ctx context.Context, id string, patchRequest api.StorageClassConfigurationPatchRequest,
	tenantId int32, expectedCode int, expectedMessage string) {
	resp, err := storageClassConfigurationService.PatchStorageClassConfigurationById(ctx,
		id, patchRequest, tenantId)
	Expect(err).To(BeNil())
	Expect(resp).ToNot(BeNil())
	Expect(resp.Code).To(Equal(expectedCode), fmt.Sprintf("%+v", resp.Body))
	Expect(resp.Body.(api.Error).Message).To(ContainSubstring(expectedMessage))
}

func updateConfiguration(ctx context.Context, id string, updateRequest api.StorageClassConfigurationUpdateRequest,
	tenantId int32) api.StorageClassConfiguration {
	resp, err := storageClassConfigurationService.UpdateStorageClassConfigurationById(ctx,
		id, updateRequest, tenantId)
	Expect(err).To(BeNil())
	Expect(resp).ToNot(BeNil())
	Expect(resp.Code).To(Equal(http.StatusOK), fmt.Sprintf("%+v", resp.Body))
	return resp.Body.(api.StorageClassConfiguration)
}

func updateConfigurationError(ctx context.Context, id string, updateRequest api.StorageClassConfigurationUpdateRequest,
	tenantId int32, expectedCode int, expectedMessage string) {
	resp, err := storageClassConfigurationService.UpdateStorageClassConfigurationById(ctx,
		id, updateRequest, tenantId)
	Expect(err).To(BeNil())
	Expect(resp).ToNot(BeNil())
	Expect(resp.Code).To(Equal(expectedCode), fmt.Sprintf("%+v", resp.Body))
	Expect(resp.Body.(*api.Error).Message).To(ContainSubstring(expectedMessage))
}
