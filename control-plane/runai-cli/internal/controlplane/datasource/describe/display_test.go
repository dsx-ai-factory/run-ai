/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package describe_test

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/datasource/describe"
	"gopkg.in/yaml.v3"
)

// captureOutput captures stdout during the execution of the provided function
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

var _ = Describe("DatasourceDisplayer", func() {
	var (
		datasource *describe.Datasource
		displayer  *describe.DatasourceDisplayer
	)

	BeforeEach(func() {
		datasource = newTestDatasource()
		displayer = describe.NewDatasourceDisplayer(datasource)
	})

	Context("NewDatasourceDisplayer", func() {
		It("should create displayer successfully", func() {
			Expect(displayer).ToNot(BeNil())
		})
	})

	Context("AsTable", func() {
		It("should not return error", func() {
			err := displayer.AsTable()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("AsJSON", func() {
		It("should generate valid JSON", func() {
			err := displayer.AsJSON()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should output valid JSON with expected structure", func() {
			output := captureOutput(func() {
				err := displayer.AsJSON()
				Expect(err).ToNot(HaveOccurred())
			})

			var result map[string]interface{}
			err := json.Unmarshal([]byte(output), &result)
			Expect(err).ToNot(HaveOccurred())

			// Validate meta section exists
			Expect(result["meta"]).ToNot(BeNil())
			meta := result["meta"].(map[string]interface{})
			Expect(meta["name"]).To(Equal("test-nfs"))
			Expect(meta["type"]).To(Equal("nfs"))
			Expect(meta["scope"]).To(Equal("Project"))
		})
	})

	Context("AsYAML", func() {
		It("should generate valid YAML", func() {
			err := displayer.AsYAML()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should output valid YAML with expected structure", func() {
			output := captureOutput(func() {
				err := displayer.AsYAML()
				Expect(err).ToNot(HaveOccurred())
			})

			var result map[string]interface{}
			err := yaml.Unmarshal([]byte(output), &result)
			Expect(err).ToNot(HaveOccurred())

			// Validate meta section exists
			Expect(result["meta"]).ToNot(BeNil())
			meta := result["meta"].(map[string]interface{})
			Expect(meta["name"]).To(Equal("test-nfs"))
			Expect(meta["type"]).To(Equal("nfs"))
		})
	})
})

var _ = Describe("RenderOutput", func() {
	var datasource *describe.Datasource

	BeforeEach(func() {
		datasource = newTestDatasource()
	})

	Context("with table output format", func() {
		It("should render successfully", func() {
			err := describe.RenderOutput(datasource, "table")
			Expect(err).ToNot(HaveOccurred())
		})

		It("should output table with datasource name", func() {
			output := captureOutput(func() {
				err := describe.RenderOutput(datasource, "table")
				Expect(err).ToNot(HaveOccurred())
			})

			Expect(output).To(ContainSubstring("test-nfs"))
			Expect(output).To(ContainSubstring("Name"))
		})
	})

	Context("with json output format", func() {
		It("should render successfully", func() {
			err := describe.RenderOutput(datasource, "json")
			Expect(err).ToNot(HaveOccurred())
		})

		It("should output valid JSON", func() {
			output := captureOutput(func() {
				err := describe.RenderOutput(datasource, "json")
				Expect(err).ToNot(HaveOccurred())
			})

			var result map[string]interface{}
			err := json.Unmarshal([]byte(output), &result)
			Expect(err).ToNot(HaveOccurred())
			Expect(result["meta"]).ToNot(BeNil())
		})
	})

	Context("with yaml output format", func() {
		It("should render successfully", func() {
			err := describe.RenderOutput(datasource, "yaml")
			Expect(err).ToNot(HaveOccurred())
		})

		It("should output valid YAML", func() {
			output := captureOutput(func() {
				err := describe.RenderOutput(datasource, "yaml")
				Expect(err).ToNot(HaveOccurred())
			})

			var result map[string]interface{}
			err := yaml.Unmarshal([]byte(output), &result)
			Expect(err).ToNot(HaveOccurred())
			Expect(result["meta"]).ToNot(BeNil())
		})
	})

	Context("with unsupported output format", func() {
		It("should return error", func() {
			err := describe.RenderOutput(datasource, "unsupported")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unsupported output format"))
		})
	})
})

var _ = Describe("Datasource types", func() {
	Context("NFS datasource", func() {
		It("should display NFS spec", func() {
			datasource := newTestNFSDatasource()
			displayer := describe.NewDatasourceDisplayer(datasource)

			err := displayer.AsTable()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should render NFS as JSON with NFS spec fields", func() {
			datasource := newTestNFSDatasource()

			output := captureOutput(func() {
				err := describe.RenderOutput(datasource, "json")
				Expect(err).ToNot(HaveOccurred())
			})

			var result map[string]interface{}
			err := json.Unmarshal([]byte(output), &result)
			Expect(err).ToNot(HaveOccurred())

			spec := result["spec"].(map[string]interface{})
			Expect(spec["nfs"]).ToNot(BeNil())
			nfs := spec["nfs"].(map[string]interface{})
			Expect(nfs["server"]).To(Equal("nfs.example.com"))
		})

		It("should render NFS as YAML with NFS spec fields", func() {
			datasource := newTestNFSDatasource()

			output := captureOutput(func() {
				err := describe.RenderOutput(datasource, "yaml")
				Expect(err).ToNot(HaveOccurred())
			})

			Expect(output).To(ContainSubstring("nfs:"))
			Expect(output).To(ContainSubstring("server:"))
		})
	})

	Context("PVC datasource", func() {
		It("should display PVC spec", func() {
			datasource := newTestPVCDatasource()
			displayer := describe.NewDatasourceDisplayer(datasource)

			err := displayer.AsTable()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should render PVC as JSON with PVC spec fields", func() {
			datasource := newTestPVCDatasource()

			output := captureOutput(func() {
				err := describe.RenderOutput(datasource, "json")
				Expect(err).ToNot(HaveOccurred())
			})

			var result map[string]interface{}
			err := json.Unmarshal([]byte(output), &result)
			Expect(err).ToNot(HaveOccurred())

			spec := result["spec"].(map[string]interface{})
			Expect(spec["pvc"]).ToNot(BeNil())
			pvc := spec["pvc"].(map[string]interface{})
			Expect(pvc["claimName"]).To(Equal("my-pvc"))
		})

		It("should render PVC as YAML with PVC spec fields", func() {
			datasource := newTestPVCDatasource()

			output := captureOutput(func() {
				err := describe.RenderOutput(datasource, "yaml")
				Expect(err).ToNot(HaveOccurred())
			})

			Expect(output).To(ContainSubstring("pvc:"))
			Expect(output).To(ContainSubstring("claimName:"))
		})
	})

	Context("S3 datasource", func() {
		It("should display S3 spec", func() {
			datasource := newTestS3Datasource()
			displayer := describe.NewDatasourceDisplayer(datasource)

			err := displayer.AsTable()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should render S3 as JSON with S3 spec fields", func() {
			datasource := newTestS3Datasource()

			output := captureOutput(func() {
				err := describe.RenderOutput(datasource, "json")
				Expect(err).ToNot(HaveOccurred())
			})

			var result map[string]interface{}
			err := json.Unmarshal([]byte(output), &result)
			Expect(err).ToNot(HaveOccurred())

			spec := result["spec"].(map[string]interface{})
			Expect(spec["s3"]).ToNot(BeNil())
			s3 := spec["s3"].(map[string]interface{})
			Expect(s3["bucket"]).To(Equal("my-bucket"))
		})

		It("should render S3 as YAML with S3 spec fields", func() {
			datasource := newTestS3Datasource()

			output := captureOutput(func() {
				err := describe.RenderOutput(datasource, "yaml")
				Expect(err).ToNot(HaveOccurred())
			})

			Expect(output).To(ContainSubstring("s3:"))
			Expect(output).To(ContainSubstring("bucket:"))
		})
	})

	Context("Git datasource", func() {
		It("should display Git spec", func() {
			datasource := newTestGitDatasource()
			displayer := describe.NewDatasourceDisplayer(datasource)

			err := displayer.AsTable()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should render Git as JSON with Git spec fields", func() {
			datasource := newTestGitDatasource()

			output := captureOutput(func() {
				err := describe.RenderOutput(datasource, "json")
				Expect(err).ToNot(HaveOccurred())
			})

			var result map[string]interface{}
			err := json.Unmarshal([]byte(output), &result)
			Expect(err).ToNot(HaveOccurred())

			spec := result["spec"].(map[string]interface{})
			Expect(spec["git"]).ToNot(BeNil())
			git := spec["git"].(map[string]interface{})
			Expect(git["repository"]).To(ContainSubstring("github.com"))
		})

		It("should render Git as YAML with Git spec fields", func() {
			datasource := newTestGitDatasource()

			output := captureOutput(func() {
				err := describe.RenderOutput(datasource, "yaml")
				Expect(err).ToNot(HaveOccurred())
			})

			Expect(output).To(ContainSubstring("git:"))
			Expect(output).To(ContainSubstring("repository:"))
		})
	})

	Context("HostPath datasource", func() {
		It("should display HostPath spec", func() {
			datasource := newTestHostPathDatasource()
			displayer := describe.NewDatasourceDisplayer(datasource)

			err := displayer.AsTable()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should render HostPath as JSON with HostPath spec fields", func() {
			datasource := newTestHostPathDatasource()

			output := captureOutput(func() {
				err := describe.RenderOutput(datasource, "json")
				Expect(err).ToNot(HaveOccurred())
			})

			var result map[string]interface{}
			err := json.Unmarshal([]byte(output), &result)
			Expect(err).ToNot(HaveOccurred())

			spec := result["spec"].(map[string]interface{})
			Expect(spec["hostPath"]).ToNot(BeNil())
			hostPath := spec["hostPath"].(map[string]interface{})
			Expect(hostPath["path"]).To(Equal("/host/data"))
		})

		It("should render HostPath as YAML with HostPath spec fields", func() {
			datasource := newTestHostPathDatasource()

			output := captureOutput(func() {
				err := describe.RenderOutput(datasource, "yaml")
				Expect(err).ToNot(HaveOccurred())
			})

			Expect(output).To(ContainSubstring("hostPath:"))
			Expect(output).To(ContainSubstring("path:"))
		})
	})

	Context("ConfigMap datasource", func() {
		It("should display ConfigMap spec", func() {
			datasource := newTestConfigMapDatasource()
			displayer := describe.NewDatasourceDisplayer(datasource)

			err := displayer.AsTable()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should render ConfigMap as JSON with ConfigMap spec fields", func() {
			datasource := newTestConfigMapDatasource()

			output := captureOutput(func() {
				err := describe.RenderOutput(datasource, "json")
				Expect(err).ToNot(HaveOccurred())
			})

			var result map[string]interface{}
			err := json.Unmarshal([]byte(output), &result)
			Expect(err).ToNot(HaveOccurred())

			spec := result["spec"].(map[string]interface{})
			Expect(spec["configMap"]).ToNot(BeNil())
			configMap := spec["configMap"].(map[string]interface{})
			Expect(configMap["configMapName"]).To(Equal("my-config"))
		})

		It("should render ConfigMap as YAML with ConfigMap spec fields", func() {
			datasource := newTestConfigMapDatasource()

			output := captureOutput(func() {
				err := describe.RenderOutput(datasource, "yaml")
				Expect(err).ToNot(HaveOccurred())
			})

			Expect(output).To(ContainSubstring("configMap:"))
			Expect(output).To(ContainSubstring("configMapName:"))
		})
	})

	Context("Secret datasource", func() {
		It("should display Secret spec", func() {
			datasource := newTestSecretDatasource()
			displayer := describe.NewDatasourceDisplayer(datasource)

			err := displayer.AsTable()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should render Secret as JSON with Secret spec fields", func() {
			datasource := newTestSecretDatasource()

			output := captureOutput(func() {
				err := describe.RenderOutput(datasource, "json")
				Expect(err).ToNot(HaveOccurred())
			})

			var result map[string]interface{}
			err := json.Unmarshal([]byte(output), &result)
			Expect(err).ToNot(HaveOccurred())

			spec := result["spec"].(map[string]interface{})
			Expect(spec["secret"]).ToNot(BeNil())
			secret := spec["secret"].(map[string]interface{})
			Expect(secret["credentialAssetId"]).To(Equal("cred-asset-123"))
		})

		It("should render Secret as YAML with Secret spec fields", func() {
			datasource := newTestSecretDatasource()

			output := captureOutput(func() {
				err := describe.RenderOutput(datasource, "yaml")
				Expect(err).ToNot(HaveOccurred())
			})

			Expect(output).To(ContainSubstring("secret:"))
			Expect(output).To(ContainSubstring("credentialAssetId:"))
		})
	})
})

// Helper function to create a basic test datasource (NFS type)
func newTestDatasource() *describe.Datasource {
	return newTestNFSDatasource()
}

// Helper function to create test NFS datasource
func newTestNFSDatasource() *describe.Datasource {
	server := "nfs.example.com"
	path := "/data"
	mountPath := "/mnt/data"
	readOnly := false
	description := "Test NFS datasource"
	projectName := "test-project"
	projectID := int32(123)
	clusterID := "cluster-123"

	return &describe.Datasource{
		Meta: describe.DatasourceMeta{
			Name:        "test-nfs",
			Type:        "nfs",
			Description: &description,
			Scope:       "Project",
			ProjectName: &projectName,
			ProjectID:   &projectID,
			ClusterID:   &clusterID,
			CreatedBy:   "test-user",
			CreatedAt:   time.Now(),
			UpdatedBy:   "test-user",
			UpdatedAt:   time.Now(),
		},
		Spec: &describe.DatasourceSpec{
			NFS: &describe.NFSSpec{
				Server:    &server,
				Path:      &path,
				MountPath: &mountPath,
				ReadOnly:  &readOnly,
			},
		},
	}
}

// Helper function to create test PVC datasource
func newTestPVCDatasource() *describe.Datasource {
	path := "/data"
	existingPVC := true
	claimName := "my-pvc"
	readOnly := false
	ephemeral := false
	dataSharing := true
	size := "10Gi"
	storageClass := "standard"
	volumeMode := "Filesystem"
	rwo := true
	rom := false
	rwm := false
	description := "Test PVC datasource"

	return &describe.Datasource{
		Meta: describe.DatasourceMeta{
			Name:        "test-pvc",
			Type:        "pvc",
			Description: &description,
			Scope:       "Project",
			CreatedBy:   "test-user",
			CreatedAt:   time.Now(),
			UpdatedBy:   "test-user",
			UpdatedAt:   time.Now(),
		},
		Spec: &describe.DatasourceSpec{
			PVC: &describe.PVCSpec{
				Path:        &path,
				ExistingPVC: &existingPVC,
				ClaimName:   &claimName,
				ReadOnly:    &readOnly,
				Ephemeral:   &ephemeral,
				DataSharing: &dataSharing,
				ClaimInfo: &describe.ClaimInfo{
					Size:         &size,
					StorageClass: &storageClass,
					VolumeMode:   &volumeMode,
					AccessModes: &describe.AccessModes{
						ReadWriteOnce: &rwo,
						ReadOnlyMany:  &rom,
						ReadWriteMany: &rwm,
					},
				},
			},
		},
	}
}

// Helper function to create test S3 datasource
func newTestS3Datasource() *describe.Datasource {
	bucket := "my-bucket"
	path := "/data"
	url := "https://s3.amazonaws.com"
	accessKeyAssetID := "access-key-123"
	description := "Test S3 datasource"

	return &describe.Datasource{
		Meta: describe.DatasourceMeta{
			Name:        "test-s3",
			Type:        "s3",
			Description: &description,
			Scope:       "Cluster",
			CreatedBy:   "test-user",
			CreatedAt:   time.Now(),
			UpdatedBy:   "test-user",
			UpdatedAt:   time.Now(),
		},
		Spec: &describe.DatasourceSpec{
			S3: &describe.S3Spec{
				Bucket:           &bucket,
				Path:             &path,
				URL:              &url,
				AccessKeyAssetID: &accessKeyAssetID,
			},
		},
	}
}

// Helper function to create test Git datasource
func newTestGitDatasource() *describe.Datasource {
	repository := "https://github.com/example/repo.git"
	branch := "main"
	revision := "abc123"
	path := "/code"
	credentialID := "cred-123"
	description := "Test Git datasource"

	return &describe.Datasource{
		Meta: describe.DatasourceMeta{
			Name:        "test-git",
			Type:        "git",
			Description: &description,
			Scope:       "Project",
			CreatedBy:   "test-user",
			CreatedAt:   time.Now(),
			UpdatedBy:   "test-user",
			UpdatedAt:   time.Now(),
		},
		Spec: &describe.DatasourceSpec{
			Git: &describe.GitSpec{
				Repository:   &repository,
				Branch:       &branch,
				Revision:     &revision,
				Path:         &path,
				CredentialID: &credentialID,
			},
		},
	}
}

// Helper function to create test HostPath datasource
func newTestHostPathDatasource() *describe.Datasource {
	path := "/host/data"
	mountPath := "/mnt/data"
	readOnly := true
	mountPropagation := "HostToContainer"
	description := "Test HostPath datasource"

	return &describe.Datasource{
		Meta: describe.DatasourceMeta{
			Name:        "test-hostpath",
			Type:        "hostPath",
			Description: &description,
			Scope:       "Cluster",
			CreatedBy:   "test-user",
			CreatedAt:   time.Now(),
			UpdatedBy:   "test-user",
			UpdatedAt:   time.Now(),
		},
		Spec: &describe.DatasourceSpec{
			HostPath: &describe.HostPathSpec{
				Path:             &path,
				MountPath:        &mountPath,
				ReadOnly:         &readOnly,
				MountPropagation: &mountPropagation,
			},
		},
	}
}

// Helper function to create test ConfigMap datasource
func newTestConfigMapDatasource() *describe.Datasource {
	configMapName := "my-config"
	mountPath := "/etc/config"
	subPath := "config.yaml"
	defaultMode := "0644"
	description := "Test ConfigMap datasource"

	return &describe.Datasource{
		Meta: describe.DatasourceMeta{
			Name:        "test-configmap",
			Type:        "configMap",
			Description: &description,
			Scope:       "Project",
			CreatedBy:   "test-user",
			CreatedAt:   time.Now(),
			UpdatedBy:   "test-user",
			UpdatedAt:   time.Now(),
		},
		Spec: &describe.DatasourceSpec{
			ConfigMap: &describe.ConfigMapSpec{
				ConfigMapName: &configMapName,
				MountPath:     &mountPath,
				SubPath:       &subPath,
				DefaultMode:   &defaultMode,
			},
		},
	}
}

// Helper function to create test Secret datasource
func newTestSecretDatasource() *describe.Datasource {
	credentialAssetID := "cred-asset-123"
	mountPath := "/etc/secrets"
	defaultMode := "0600"
	description := "Test Secret datasource"

	return &describe.Datasource{
		Meta: describe.DatasourceMeta{
			Name:        "test-secret",
			Type:        "secret",
			Description: &description,
			Scope:       "Project",
			CreatedBy:   "test-user",
			CreatedAt:   time.Now(),
			UpdatedBy:   "test-user",
			UpdatedAt:   time.Now(),
		},
		Spec: &describe.DatasourceSpec{
			Secret: &describe.SecretSpec{
				CredentialAssetID: &credentialAssetID,
				MountPath:         &mountPath,
				DefaultMode:       &defaultMode,
			},
		},
	}
}
