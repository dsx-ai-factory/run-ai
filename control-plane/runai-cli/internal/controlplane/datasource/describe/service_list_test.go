/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package describe_test

import (
	"context"
	"errors"

	cpmocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/datasource/describe"
	assetsgenerated "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Datasource Describe Service", func() {
	var (
		mockCtrl   *gomock.Controller
		mockLister *cpmocks.MockDatasourceDescribeLister
		ctx        context.Context
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockLister = cpmocks.NewMockDatasourceDescribeLister(mockCtrl)
		ctx = context.Background()
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Context("FindByName", func() {
		It("should find datasource by name", func() {
			datasource := newTestDatasourceEntry("test-ds", "nfs", "Project")
			listResponse := &assetsgenerated.DatasourceListResponse{
				Entries: []assetsgenerated.DatasourceListResponseEntry{datasource},
			}
			mockLister.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(listResponse, nil)

			projectID := int32(123)
			clusterID := "cluster-123"
			result, err := describe.FindByName(ctx, mockLister, "test-ds", &projectID, &clusterID, nil)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.Meta.Name).To(Equal("test-ds"))
		})

		It("should return error when datasource not found", func() {
			listResponse := &assetsgenerated.DatasourceListResponse{
				Entries: []assetsgenerated.DatasourceListResponseEntry{},
			}
			mockLister.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(listResponse, nil)

			projectID := int32(123)
			result, err := describe.FindByName(ctx, mockLister, "nonexistent-ds", &projectID, nil, nil)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found"))
			Expect(result).To(BeNil())
		})

		It("should return type-specific error when datasource with type not found", func() {
			listResponse := &assetsgenerated.DatasourceListResponse{
				Entries: []assetsgenerated.DatasourceListResponseEntry{},
			}
			mockLister.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(listResponse, nil)

			projectID := int32(123)
			dsType := "nfs"
			result, err := describe.FindByName(ctx, mockLister, "test-ds", &projectID, nil, &dsType)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("of type"))
			Expect(err.Error()).To(ContainSubstring("nfs"))
			Expect(result).To(BeNil())
		})

		It("should return error when list fails", func() {
			mockLister.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("list error"))

			projectID := int32(123)
			result, err := describe.FindByName(ctx, mockLister, "test-ds", &projectID, nil, nil)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to search for datasource"))
			Expect(result).To(BeNil())
		})

		It("should return first matching datasource when multiple exist", func() {
			datasource1 := newTestDatasourceEntry("test-ds", "nfs", "Project")
			datasource2 := newTestDatasourceEntry("test-ds", "pvc", "Cluster")
			listResponse := &assetsgenerated.DatasourceListResponse{
				Entries: []assetsgenerated.DatasourceListResponseEntry{datasource1, datasource2},
			}
			mockLister.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(listResponse, nil)

			projectID := int32(123)
			result, err := describe.FindByName(ctx, mockLister, "test-ds", &projectID, nil, nil)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.Meta.Kind).To(Equal(assetsgenerated.AssetKind("nfs")))
		})

		It("should return error when name doesn't match any entry", func() {
			datasource := newTestDatasourceEntry("other-ds", "nfs", "Project")
			listResponse := &assetsgenerated.DatasourceListResponse{
				Entries: []assetsgenerated.DatasourceListResponseEntry{datasource},
			}
			mockLister.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(listResponse, nil)

			projectID := int32(123)
			result, err := describe.FindByName(ctx, mockLister, "test-ds", &projectID, nil, nil)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found"))
			Expect(result).To(BeNil())
		})
	})

	Context("ListByName with name", func() {
		It("should resolve datasource by name with project filter", func() {
			datasource := newTestDatasourceEntry("test-ds", "nfs", "Project")
			listResponse := &assetsgenerated.DatasourceListResponse{
				Entries: []assetsgenerated.DatasourceListResponseEntry{datasource},
			}
			mockLister.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(listResponse, nil)

			projectID := int32(123)
			clusterID := "cluster-123"
			result, err := describe.ListByName(ctx, mockLister, clusterID, "test-ds", &projectID, nil)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.Meta.Name).To(Equal("test-ds"))
		})

		It("should resolve datasource by name with cluster filter", func() {
			datasource := newTestDatasourceEntry("test-ds", "nfs", "Cluster")
			listResponse := &assetsgenerated.DatasourceListResponse{
				Entries: []assetsgenerated.DatasourceListResponseEntry{datasource},
			}
			mockLister.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(listResponse, nil)

			clusterID := "cluster-123"
			result, err := describe.ListByName(ctx, mockLister, clusterID, "test-ds", nil, nil)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.Meta.Name).To(Equal("test-ds"))
		})

		It("should return error when datasource not found by name", func() {
			listResponse := &assetsgenerated.DatasourceListResponse{
				Entries: []assetsgenerated.DatasourceListResponseEntry{},
			}
			mockLister.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(listResponse, nil)

			projectID := int32(123)
			clusterID := "cluster-123"
			result, err := describe.ListByName(ctx, mockLister, clusterID, "nonexistent-ds", &projectID, nil)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("does not exist"))
			Expect(result).To(BeNil())
		})

		It("should return error when list fails", func() {
			mockLister.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("list error"))

			projectID := int32(123)
			clusterID := "cluster-123"
			result, err := describe.ListByName(ctx, mockLister, clusterID, "test-ds", &projectID, nil)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to fetch datasource"))
			Expect(result).To(BeNil())
		})
	})

	Context("ListByName with multiple matching names", func() {
		It("should return the first matching datasource", func() {
			datasource1 := newTestDatasourceEntry("test-ds", "nfs", "Project")
			datasource2 := newTestDatasourceEntry("test-ds", "pvc", "Cluster")
			listResponse := &assetsgenerated.DatasourceListResponse{
				Entries: []assetsgenerated.DatasourceListResponseEntry{datasource1, datasource2},
			}
			mockLister.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(listResponse, nil)

			projectID := int32(123)
			clusterID := "cluster-123"
			result, err := describe.ListByName(ctx, mockLister, clusterID, "test-ds", &projectID, nil)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.Meta.Name).To(Equal("test-ds"))
		})
	})

	Context("ListByName with type filter", func() {
		It("should pass type filter to API and return matching datasource", func() {
			datasource := newTestPVCDatasourceEntry()
			listResponse := &assetsgenerated.DatasourceListResponse{
				Entries: []assetsgenerated.DatasourceListResponseEntry{datasource},
			}
			dsType := "pvc"
			// Verify that the type filter is passed correctly to the API
			mockLister.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), &dsType).Return(listResponse, nil)

			projectID := int32(123)
			clusterID := "cluster-123"
			result, err := describe.ListByName(ctx, mockLister, clusterID, "test-pvc", &projectID, &dsType)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.Spec.PVC).ToNot(BeNil())
		})

		It("should return error when type filter matches no datasource", func() {
			listResponse := &assetsgenerated.DatasourceListResponse{
				Entries: []assetsgenerated.DatasourceListResponseEntry{},
			}
			dsType := "nfs"
			mockLister.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), &dsType).Return(listResponse, nil)

			projectID := int32(123)
			clusterID := "cluster-123"
			result, err := describe.ListByName(ctx, mockLister, clusterID, "test-ds", &projectID, &dsType)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("does not exist"))
			Expect(result).To(BeNil())
		})

		It("should filter by nfs type correctly", func() {
			datasource := newTestNFSDatasourceEntry()
			listResponse := &assetsgenerated.DatasourceListResponse{
				Entries: []assetsgenerated.DatasourceListResponseEntry{datasource},
			}
			dsType := "nfs"
			mockLister.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), &dsType).Return(listResponse, nil)

			projectID := int32(123)
			clusterID := "cluster-123"
			result, err := describe.ListByName(ctx, mockLister, clusterID, "test-nfs", &projectID, &dsType)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.Spec.NFS).ToNot(BeNil())
		})

		It("should filter by s3 type correctly", func() {
			datasource := newTestS3DatasourceEntry()
			listResponse := &assetsgenerated.DatasourceListResponse{
				Entries: []assetsgenerated.DatasourceListResponseEntry{datasource},
			}
			dsType := "s3"
			mockLister.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), &dsType).Return(listResponse, nil)

			projectID := int32(123)
			clusterID := "cluster-123"
			result, err := describe.ListByName(ctx, mockLister, clusterID, "test-s3", &projectID, &dsType)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.Spec.S3).ToNot(BeNil())
		})

		It("should filter by git type correctly", func() {
			datasource := newTestGitDatasourceEntry()
			listResponse := &assetsgenerated.DatasourceListResponse{
				Entries: []assetsgenerated.DatasourceListResponseEntry{datasource},
			}
			dsType := "git"
			mockLister.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), &dsType).Return(listResponse, nil)

			projectID := int32(123)
			clusterID := "cluster-123"
			result, err := describe.ListByName(ctx, mockLister, clusterID, "test-git", &projectID, &dsType)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.Spec.Git).ToNot(BeNil())
		})

		It("should filter by hostPath type correctly", func() {
			datasource := newTestHostPathDatasourceEntry()
			listResponse := &assetsgenerated.DatasourceListResponse{
				Entries: []assetsgenerated.DatasourceListResponseEntry{datasource},
			}
			dsType := "hostPath"
			mockLister.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), &dsType).Return(listResponse, nil)

			projectID := int32(123)
			clusterID := "cluster-123"
			result, err := describe.ListByName(ctx, mockLister, clusterID, "test-hostpath", &projectID, &dsType)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.Spec.HostPath).ToNot(BeNil())
		})

		It("should filter by configMap type correctly", func() {
			datasource := newTestConfigMapDatasourceEntry()
			listResponse := &assetsgenerated.DatasourceListResponse{
				Entries: []assetsgenerated.DatasourceListResponseEntry{datasource},
			}
			dsType := "configMap"
			mockLister.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), &dsType).Return(listResponse, nil)

			projectID := int32(123)
			clusterID := "cluster-123"
			result, err := describe.ListByName(ctx, mockLister, clusterID, "test-configmap", &projectID, &dsType)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.Spec.ConfigMap).ToNot(BeNil())
		})

		It("should filter by secretVolume type correctly", func() {
			datasource := newTestSecretDatasourceEntry()
			listResponse := &assetsgenerated.DatasourceListResponse{
				Entries: []assetsgenerated.DatasourceListResponseEntry{datasource},
			}
			dsType := "secretVolume"
			mockLister.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), &dsType).Return(listResponse, nil)

			projectID := int32(123)
			clusterID := "cluster-123"
			result, err := describe.ListByName(ctx, mockLister, clusterID, "test-secret", &projectID, &dsType)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.Spec.Secret).ToNot(BeNil())
		})
	})

	Context("ListByName with different datasource types", func() {
		It("should handle NFS datasource", func() {
			datasource := newTestNFSDatasourceEntry()
			listResponse := &assetsgenerated.DatasourceListResponse{
				Entries: []assetsgenerated.DatasourceListResponseEntry{datasource},
			}
			mockLister.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(listResponse, nil)

			projectID := int32(123)
			clusterID := "cluster-123"
			result, err := describe.ListByName(ctx, mockLister, clusterID, "test-nfs", &projectID, nil)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.Spec.NFS).ToNot(BeNil())
		})

		It("should handle PVC datasource", func() {
			datasource := newTestPVCDatasourceEntry()
			listResponse := &assetsgenerated.DatasourceListResponse{
				Entries: []assetsgenerated.DatasourceListResponseEntry{datasource},
			}
			mockLister.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(listResponse, nil)

			projectID := int32(123)
			clusterID := "cluster-123"
			result, err := describe.ListByName(ctx, mockLister, clusterID, "test-pvc", &projectID, nil)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.Spec.PVC).ToNot(BeNil())
		})

		It("should handle S3 datasource", func() {
			datasource := newTestS3DatasourceEntry()
			listResponse := &assetsgenerated.DatasourceListResponse{
				Entries: []assetsgenerated.DatasourceListResponseEntry{datasource},
			}
			mockLister.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(listResponse, nil)

			projectID := int32(123)
			clusterID := "cluster-123"
			result, err := describe.ListByName(ctx, mockLister, clusterID, "test-s3", &projectID, nil)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.Spec.S3).ToNot(BeNil())
		})

		It("should handle Git datasource", func() {
			datasource := newTestGitDatasourceEntry()
			listResponse := &assetsgenerated.DatasourceListResponse{
				Entries: []assetsgenerated.DatasourceListResponseEntry{datasource},
			}
			mockLister.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(listResponse, nil)

			projectID := int32(123)
			clusterID := "cluster-123"
			result, err := describe.ListByName(ctx, mockLister, clusterID, "test-git", &projectID, nil)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.Spec.Git).ToNot(BeNil())
		})

		It("should handle HostPath datasource", func() {
			datasource := newTestHostPathDatasourceEntry()
			listResponse := &assetsgenerated.DatasourceListResponse{
				Entries: []assetsgenerated.DatasourceListResponseEntry{datasource},
			}
			mockLister.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(listResponse, nil)

			projectID := int32(123)
			clusterID := "cluster-123"
			result, err := describe.ListByName(ctx, mockLister, clusterID, "test-hostpath", &projectID, nil)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.Spec.HostPath).ToNot(BeNil())
		})

		It("should handle ConfigMap datasource", func() {
			datasource := newTestConfigMapDatasourceEntry()
			listResponse := &assetsgenerated.DatasourceListResponse{
				Entries: []assetsgenerated.DatasourceListResponseEntry{datasource},
			}
			mockLister.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(listResponse, nil)

			projectID := int32(123)
			clusterID := "cluster-123"
			result, err := describe.ListByName(ctx, mockLister, clusterID, "test-configmap", &projectID, nil)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.Spec.ConfigMap).ToNot(BeNil())
		})

		It("should handle Secret datasource", func() {
			datasource := newTestSecretDatasourceEntry()
			listResponse := &assetsgenerated.DatasourceListResponse{
				Entries: []assetsgenerated.DatasourceListResponseEntry{datasource},
			}
			mockLister.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(listResponse, nil)

			projectID := int32(123)
			clusterID := "cluster-123"
			result, err := describe.ListByName(ctx, mockLister, clusterID, "test-secret", &projectID, nil)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.Spec.Secret).ToNot(BeNil())
		})
	})
})

// Helper function to create basic test datasource entry
func newTestDatasourceEntry(name, kind, scope string) assetsgenerated.DatasourceListResponseEntry {
	meta := assetsgenerated.AssetMeta{
		Name:  name,
		Kind:  assetsgenerated.AssetKind(kind),
		Scope: assetsgenerated.Scope(scope),
	}

	return assetsgenerated.DatasourceListResponseEntry{
		Meta: meta,
		Spec: assetsgenerated.DatasourceListResponseAssetSpec{},
	}
}

// Helper function to create test NFS datasource entry
func newTestNFSDatasourceEntry() assetsgenerated.DatasourceListResponseEntry {
	server := "nfs.example.com"
	path := "/data"
	mountPath := "/mnt/data"

	nfsSpec := assetsgenerated.NFSAssetSpec{}
	nfsSpec.SetServer(server)
	nfsSpec.SetPath(path)
	nfsSpec.SetMountPath(mountPath)
	nfsSpec.SetReadOnly(false)

	spec := assetsgenerated.DatasourceListResponseAssetSpec{}
	spec.SetNfs(nfsSpec)

	meta := assetsgenerated.AssetMeta{
		Name:  "test-nfs",
		Kind:  assetsgenerated.AssetKind("nfs"),
		Scope: assetsgenerated.SCOPE_PROJECT,
	}

	return assetsgenerated.DatasourceListResponseEntry{
		Meta: meta,
		Spec: spec,
	}
}

// Helper function to create test PVC datasource entry
func newTestPVCDatasourceEntry() assetsgenerated.DatasourceListResponseEntry {
	path := "/data"
	existingPvc := true
	claimName := "my-pvc"
	readOnly := false
	ephemeral := false
	dataSharing := true

	pvcSpec := assetsgenerated.PVCAssetSpec{}
	pvcSpec.SetPath(path)
	pvcSpec.SetExistingPvc(existingPvc)
	pvcSpec.SetClaimName(claimName)
	pvcSpec.SetReadOnly(readOnly)
	pvcSpec.SetEphemeral(ephemeral)
	pvcSpec.SetDataSharing(dataSharing)

	spec := assetsgenerated.DatasourceListResponseAssetSpec{}
	spec.SetPvc(pvcSpec)

	meta := assetsgenerated.AssetMeta{
		Name:  "test-pvc",
		Kind:  assetsgenerated.AssetKind("pvc"),
		Scope: assetsgenerated.SCOPE_PROJECT,
	}

	return assetsgenerated.DatasourceListResponseEntry{
		Meta: meta,
		Spec: spec,
	}
}

// Helper function to create test S3 datasource entry
func newTestS3DatasourceEntry() assetsgenerated.DatasourceListResponseEntry {
	bucket := "my-bucket"
	path := "/data"
	url := "https://s3.amazonaws.com"

	s3Spec := assetsgenerated.S3AssetSpec{}
	s3Spec.SetBucket(bucket)
	s3Spec.SetPath(path)
	s3Spec.SetUrl(url)

	spec := assetsgenerated.DatasourceListResponseAssetSpec{}
	spec.SetS3(s3Spec)

	meta := assetsgenerated.AssetMeta{
		Name:  "test-s3",
		Kind:  assetsgenerated.AssetKind("s3"),
		Scope: assetsgenerated.SCOPE_CLUSTER,
	}

	return assetsgenerated.DatasourceListResponseEntry{
		Meta: meta,
		Spec: spec,
	}
}

// Helper function to create test Git datasource entry
func newTestGitDatasourceEntry() assetsgenerated.DatasourceListResponseEntry {
	repository := "https://github.com/example/repo.git"
	branch := "main"
	revision := "abc123"
	path := "/code"

	gitSpec := assetsgenerated.GitAssetSpec{}
	gitSpec.SetRepository(repository)
	gitSpec.SetBranch(branch)
	gitSpec.SetRevision(revision)
	gitSpec.SetPath(path)

	spec := assetsgenerated.DatasourceListResponseAssetSpec{}
	spec.SetGit(gitSpec)

	meta := assetsgenerated.AssetMeta{
		Name:  "test-git",
		Kind:  assetsgenerated.AssetKind("git"),
		Scope: assetsgenerated.SCOPE_PROJECT,
	}

	return assetsgenerated.DatasourceListResponseEntry{
		Meta: meta,
		Spec: spec,
	}
}

// Helper function to create test HostPath datasource entry
func newTestHostPathDatasourceEntry() assetsgenerated.DatasourceListResponseEntry {
	path := "/host/data"
	mountPath := "/mnt/data"
	readOnly := true

	hostPathSpec := assetsgenerated.HostPathSpec{}
	hostPathSpec.SetPath(path)
	hostPathSpec.SetMountPath(mountPath)
	hostPathSpec.SetReadOnly(readOnly)

	spec := assetsgenerated.DatasourceListResponseAssetSpec{}
	spec.SetHostPath(hostPathSpec)

	meta := assetsgenerated.AssetMeta{
		Name:  "test-hostpath",
		Kind:  assetsgenerated.AssetKind("hostPath"),
		Scope: assetsgenerated.SCOPE_CLUSTER,
	}

	return assetsgenerated.DatasourceListResponseEntry{
		Meta: meta,
		Spec: spec,
	}
}

// Helper function to create test ConfigMap datasource entry
func newTestConfigMapDatasourceEntry() assetsgenerated.DatasourceListResponseEntry {
	configMap := "my-config"
	mountPath := "/etc/config"
	subPath := "config.yaml"

	configMapSpec := assetsgenerated.ConfigMapAssetSpec{}
	configMapSpec.SetConfigMap(configMap)
	configMapSpec.SetMountPath(mountPath)
	configMapSpec.SetSubPath(subPath)

	spec := assetsgenerated.DatasourceListResponseAssetSpec{}
	spec.SetConfigMap(configMapSpec)

	meta := assetsgenerated.AssetMeta{
		Name:  "test-configmap",
		Kind:  assetsgenerated.AssetKind("configMap"),
		Scope: assetsgenerated.SCOPE_PROJECT,
	}

	return assetsgenerated.DatasourceListResponseEntry{
		Meta: meta,
		Spec: spec,
	}
}

// Helper function to create test Secret datasource entry
func newTestSecretDatasourceEntry() assetsgenerated.DatasourceListResponseEntry {
	credentialAssetId := "cred-asset-123"
	mountPath := "/etc/secrets"

	secretSpec := assetsgenerated.SecretAssetSpec{}
	secretSpec.SetCredentialAssetId(credentialAssetId)
	secretSpec.SetMountPath(mountPath)

	spec := assetsgenerated.DatasourceListResponseAssetSpec{}
	spec.SetSecret(secretSpec)

	meta := assetsgenerated.AssetMeta{
		Name:  "test-secret",
		Kind:  assetsgenerated.AssetKind("secret"),
		Scope: assetsgenerated.SCOPE_PROJECT,
	}

	return assetsgenerated.DatasourceListResponseEntry{
		Meta: meta,
		Spec: spec,
	}
}
