/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package init_containers

import (
	clustersApi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	flds "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"

	featuregate "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/feature-gate"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
)

var _ = Describe("BuildS3InitContainers", func() {
	It("gates every S3 mount using readiness support advertised for the exact image", func() {
		cfg := s3TestConfig().(composer.StaticClusterStorageConfig)
		cfg.SetS3FsMountCheckImage(cfg.S3FsImage())
		info := &composer.S3Info{Config: cfg, Items: map[string]*composer.S3Item{
			"first":  {Bucket: "first", Path: "/data set"},
			"second": {Bucket: "second", Path: "/second"},
			"third":  {Bucket: "third", Path: "data"},
		}}
		containers, err := BuildS3InitContainers(info)
		Expect(err).To(BeNil())
		Expect(containers).To(HaveLen(3))
		for i, key := range []string{"first", "second", "third"} {
			item := info.Items[key]
			Expect(containers[i].Command[len(containers[i].Command)-4:]).To(Equal([]string{"--ready-file", "/runai-goofys-ready", item.Bucket, item.Path}))
			Expect(containers[i].RestartPolicy).To(Equal(ptr.To(corev1.ContainerRestartPolicyAlways)))
			Expect(containers[i].StartupProbe).NotTo(BeNil())
			Expect(containers[i].StartupProbe.Exec.Command).To(Equal([]string{"goofys", "--check-ready", "/runai-goofys-ready", item.Path}))
			Expect(containers[i].StartupProbe.PeriodSeconds).To(Equal(int32(2)))
			Expect(containers[i].StartupProbe.TimeoutSeconds).To(Equal(int32(2)))
			Expect(containers[i].StartupProbe.FailureThreshold).To(Equal(int32(150)))
		}
	})

	DescribeTable("keeps legacy behavior when readiness support is unavailable",
		func(advertised string) {
			cfg := s3TestConfig().(composer.StaticClusterStorageConfig)
			cfg.SetS3FsMountCheckImage(advertised)
			containers, err := BuildS3InitContainers(&composer.S3Info{Config: cfg,
				Items: map[string]*composer.S3Item{"s3": {Bucket: "bucket", Path: "/storage"}},
			})
			Expect(err).To(BeNil())
			Expect(containers[0].StartupProbe).To(BeNil())
			Expect(containers[0].Command).NotTo(ContainElement("--ready-file"))
		},
		Entry("older cluster without a capability", ""),
		Entry("stale capability for a different image", "goofys:older"),
	)

	It("uses a capability reported by the cluster for a custom image", func() {
		config := clustersApi.ClusterReportedConfig{}
		config.SetS3FileSystemImage("custom.example/goofys:with-readiness")
		config.SetS3FileSystemMountCheckImage("custom.example/goofys:with-readiness")
		containers, err := BuildS3InitContainers(&composer.S3Info{
			Config: composer.NewClusterStorageConfigFromCluster(composer.ClusterWithConfig(config)),
			Items:  map[string]*composer.S3Item{"s3": {Bucket: "bucket", Path: "/storage"}},
		})
		Expect(err).To(BeNil())
		Expect(containers[0].StartupProbe).NotTo(BeNil())
		Expect(containers[0].StartupProbe.Exec.Command).To(Equal([]string{"goofys", "--check-ready", "/runai-goofys-ready", "/storage"}))
		Expect(containers[0].Command).To(ContainElement("--ready-file"))
	})

	It("should emit secret-backed credential env and a goofys command with endpoint, bucket and path", func() {
		info := &composer.S3Info{
			Config: s3TestConfig(),
			Items: map[string]*composer.S3Item{
				"s3-0": {
					Bucket:             "private-bucket",
					Path:               "/mnt/data",
					Url:                "https://s3.amazonaws.com",
					SecretName:         "my-secret",
					AccessKeyIdKey:     "AccessKeyId",
					SecretAccessKeyKey: "SecretKey",
				},
			},
		}

		containers, err := BuildS3InitContainers(info)
		Expect(err).To(BeNil())
		Expect(containers).To(HaveLen(1))

		c := containers[0]
		Expect(c.Name).To(Equal("s3-fs-0"))
		Expect(c.Command).To(Equal([]string{
			"goofys", "-f", "--dir-mode", "0777", "--file-mode", "0777", "-o", "allow_other", "--debug_s3",
			"--endpoint", "https://s3.amazonaws.com", "private-bucket", "/mnt/data",
		}))

		Expect(c.Env).To(HaveLen(2))
		Expect(c.Env[0].Name).To(Equal("AWS_ACCESS_KEY_ID"))
		Expect(c.Env[0].ValueFrom.SecretKeyRef.Name).To(Equal("my-secret"))
		Expect(c.Env[0].ValueFrom.SecretKeyRef.Key).To(Equal("AccessKeyId"))
		Expect(c.Env[1].Name).To(Equal("AWS_SECRET_ACCESS_KEY"))
		Expect(c.Env[1].ValueFrom.SecretKeyRef.Key).To(Equal("SecretKey"))
	})

	It("should add a plaintext session token env when present", func() {
		info := &composer.S3Info{
			Config: s3TestConfig(),
			Items: map[string]*composer.S3Item{
				"s3-0": {
					Bucket:             "b",
					Path:               "/p",
					SecretName:         "s",
					AccessKeyIdKey:     "AccessKeyId",
					SecretAccessKeyKey: "SecretKey",
					SessionToken:       "tok",
				},
			},
		}

		containers, err := BuildS3InitContainers(info)
		Expect(err).To(BeNil())
		Expect(containers[0].Env).To(HaveLen(3))
		Expect(containers[0].Env[2].Name).To(Equal("AWS_SESSION_TOKEN"))
		Expect(containers[0].Env[2].Value).To(Equal("tok"))
	})

	It("should omit credential env when no secret is provided", func() {
		info := &composer.S3Info{
			Config: s3TestConfig(),
			Items:  map[string]*composer.S3Item{"s3-0": {Bucket: "public", Path: "/p"}},
		}

		containers, err := BuildS3InitContainers(info)
		Expect(err).To(BeNil())
		Expect(containers[0].Env).To(BeEmpty())
	})

	It("should thread run-as uid and gid into the goofys command", func() {
		info := &composer.S3Info{
			Config:   s3TestConfig(),
			RunAsUid: ptrInt64(1000),
			RunAsGid: ptrInt64(2000),
			Items:    map[string]*composer.S3Item{"s3-0": {Bucket: "b", Path: "/p"}},
		}

		containers, err := BuildS3InitContainers(info)
		Expect(err).To(BeNil())
		Expect(containers[0].Command).To(ContainElements("--uid", "1000", "--gid", "2000"))
	})

	It("should mount the custom CA when enabled and index volumes per item", func() {
		cfg := composer.StaticClusterStorageConfig{}
		cfg.SetS3FsImage("goofys")
		cfg.SetIsCustomCAS3Enabled(true)
		cfg.SetCustomCAS3Secret("ca-secret")
		cfg.SetCustomCAS3Key("ca.pem")

		info := &composer.S3Info{
			Config: cfg,
			Items: map[string]*composer.S3Item{
				"s3-0": {Bucket: "b0", Path: "/p0"},
				"s3-1": {Bucket: "b1", Path: "/p1"},
			},
		}

		containers, err := BuildS3InitContainers(info)
		Expect(err).To(BeNil())
		Expect(containers).To(HaveLen(2))
		Expect(containers[0].Name).To(Equal("s3-fs-0"))
		Expect(containers[1].Name).To(Equal("s3-fs-1"))

		Expect(containers[0].VolumeMounts).To(HaveLen(2))
		Expect(containers[0].VolumeMounts[0].Name).To(Equal("s3-0"))
		Expect(*containers[0].VolumeMounts[0].MountPropagation).To(Equal(corev1.MountPropagationBidirectional))
		Expect(containers[0].VolumeMounts[1].Name).To(Equal("ca-secret"))
		Expect(containers[0].VolumeMounts[1].SubPath).To(Equal("ca.pem"))

		volumes := BuildS3Volumes(info)
		Expect(volumes).To(HaveLen(3))
		Expect(volumes[0].Name).To(Equal("s3-0"))
		Expect(volumes[1].Name).To(Equal("s3-1"))
		Expect(volumes[2].Name).To(Equal("ca-secret"))
		Expect(volumes[2].Secret).NotTo(BeNil())
	})

	It("ships the cluster-reported S3 image on the init container", func() {
		config := clustersApi.ClusterReportedConfig{}
		config.SetS3FileSystemImage("my.registry/goofys:custom")
		info := &composer.S3Info{
			Config: composer.NewClusterStorageConfigFromCluster(composer.ClusterWithConfig(config)),
			Items:  map[string]*composer.S3Item{"s3-0": {Bucket: "bucket", Path: "/data"}},
		}

		containers, err := BuildS3InitContainers(info)
		Expect(err).To(BeNil())
		Expect(containers).To(HaveLen(1))
		Expect(containers[0].Image).To(Equal("my.registry/goofys:custom"))
	})

	It("applies configured resources to every S3 instance", func() {
		cfg := composer.StaticClusterStorageConfig{}
		cfg.SetS3FsImage(s3ParityDefaultImage)
		expected := corev1.ResourceRequirements{Limits: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse("256Mi"),
		}}
		cfg.SetS3FileSystemResources(expected)
		info := &composer.S3Info{
			Config: cfg,
			Items: map[string]*composer.S3Item{
				"s3-0": {Bucket: "a", Path: "/a"},
				"s3-1": {Bucket: "b", Path: "/b"},
			},
		}

		containers, err := BuildS3InitContainers(info)

		Expect(err).To(BeNil())
		Expect(containers).To(HaveLen(2))
		Expect(containers[0].Resources).To(Equal(expected))
		Expect(containers[1].Resources).To(Equal(expected))
	})

	It("uses empty resources when unset or explicitly empty", func() {
		item := map[string]*composer.S3Item{"s3-0": {Bucket: "a", Path: "/a"}}
		unsetConfig := composer.StaticClusterStorageConfig{}
		unsetConfig.SetS3FsImage(s3ParityDefaultImage)
		unset, unsetErr := BuildS3InitContainers(&composer.S3Info{
			Config: unsetConfig,
			Items:  item,
		})
		explicitConfig := composer.StaticClusterStorageConfig{}
		explicitConfig.SetS3FsImage(s3ParityDefaultImage)
		explicitConfig.SetS3FileSystemResources(corev1.ResourceRequirements{})
		explicit, explicitErr := BuildS3InitContainers(&composer.S3Info{
			Config: explicitConfig,
			Items:  item,
		})

		Expect(unsetErr).To(BeNil())
		Expect(explicitErr).To(BeNil())
		Expect(unset[0].Resources).To(Equal(corev1.ResourceRequirements{}))
		Expect(explicit[0].Resources).To(Equal(corev1.ResourceRequirements{}))
	})

	It("reports invalid quantities with the S3 field name", func() {
		config := clustersApi.ClusterReportedConfig{}
		config.SetS3FileSystemImage(s3ParityDefaultImage)
		config.SetS3FileSystemResources(composerReportedResources(nil, map[string]string{"memory": "invalid"}))
		info := &composer.S3Info{
			Config: composer.NewClusterStorageConfigFromCluster(composer.ClusterWithConfig(config)),
			Items:  map[string]*composer.S3Item{"s3-0": {Bucket: "a", Path: "/a"}},
		}

		_, err := BuildS3InitContainers(info)

		Expect(err).NotTo(BeNil())
		Expect(err.Message).To(ContainSubstring("s3FileSystemResources"))
		Expect(err.Message).To(ContainSubstring(`limits.memory has invalid quantity "invalid"`))
	})

	It("rejects an injected container without a reported image", func() {
		containers, err := BuildS3InitContainers(&composer.S3Info{
			Config: composer.DefaultClusterStorageConfig(),
			Items:  map[string]*composer.S3Item{"s3-0": {Bucket: "a", Path: "/a"}},
		})
		Expect(containers).To(BeNil())
		Expect(err).NotTo(BeNil())
		Expect(err.Message).To(ContainSubstring("s3FileSystemImage"))
	})
})

func composerReportedResources(requests, limits map[string]string) clustersApi.ClusterReportedResourceRequirements {
	resources := clustersApi.ClusterReportedResourceRequirements{}
	resources.SetRequests(requests)
	resources.SetLimits(limits)
	return resources
}

func ptrInt64(v int64) *int64 {
	return &v
}

var s3ParityEnvVars = []string{
	"S3_FILE_SYSTEM_IMAGE",
	"CUSTOM_CA_S3_ENABLED",
	"CUSTOM_CA_S3_KEY",
	"CUSTOM_CA_S3_SECRET",
}

const s3ParityDefaultImage = "nvcr.io/nvidia/runai/goofys:2.26.30"

func s3TestConfig() composer.ClusterStorageConfig {
	config := composer.DefaultClusterStorageConfig().(composer.StaticClusterStorageConfig)
	config.SetS3FsImage(s3ParityDefaultImage)
	return config
}

func newS3WorkloadFields() *field_groups.WorkloadFields {
	wf := &field_groups.WorkloadFields{}
	wf.S3buckets = &flds.S3bucketField{Value: &map[string]flds.S3bucketItemAndRules{}}
	return wf
}

func addS3Bucket(wf *field_groups.WorkloadFields, nick string, v *flds.S3bucketItemValue) {
	(*wf.S3buckets.Value)[nick] = flds.S3bucketItemAndRules{Value: v}
}

// s3InfoFromWorkloadFields is the test twin of ComposeS3: it copies the cluster's
// already-resolved WorkloadFields 1:1 into the CP S3Info (Key->AccessKeyIdKey,
// Secret->SecretAccessKeyKey) without re-applying defaults, so both builders see
// identical inputs. The reserved-internal per-bucket Image is deliberately not
// copied; the CP has no such field.
func s3InfoFromWorkloadFields(wf *field_groups.WorkloadFields, cfg composer.ClusterStorageConfig) *composer.S3Info {
	items := wf.S3buckets.Deref()
	info := &composer.S3Info{
		Items:  make(map[string]*composer.S3Item, len(items)),
		Config: cfg,
	}
	if wf.RunAsUid != nil {
		info.RunAsUid = convutils.Ref(int64(wf.RunAsUid.Deref()))
	}
	if wf.RunAsGid != nil {
		info.RunAsGid = convutils.Ref(int64(wf.RunAsGid.Deref()))
	}
	for nick, item := range items {
		v := item.Value
		if v == nil {
			info.Items[nick] = nil
			continue
		}
		info.Items[nick] = &composer.S3Item{
			Bucket:             convutils.Deref(v.Bucket, ""),
			Path:               convutils.Deref(v.Path, ""),
			Url:                convutils.Deref(v.Url, ""),
			SecretName:         convutils.Deref(v.SecretName, ""),
			AccessKeyIdKey:     convutils.Deref(v.Key, ""),
			SecretAccessKeyKey: convutils.Deref(v.Secret, ""),
			SessionToken:       convutils.Deref(v.SessionToken, ""),
		}
	}
	return info
}

var _ = Describe("BuildS3InitContainers parity with the cluster builder", func() {
	var (
		wf     *field_groups.WorkloadFields
		cfg    composer.ClusterStorageConfig
		fgPrev bool
	)

	BeforeEach(func() {
		for _, name := range s3ParityEnvVars {
			GinkgoT().Setenv(name, "")
		}
		// The cluster builder resolves the goofys image from S3_FILE_SYSTEM_IMAGE;
		// pin it to the same default the CP config uses so parity isolates builder logic.
		GinkgoT().Setenv("S3_FILE_SYSTEM_IMAGE", s3ParityDefaultImage)
		fg := featuregate.GetFeatureGateUtilInstance()
		fgPrev = fg.FeatureGateSidecarContainersEnable
		fg.FeatureGateSidecarContainersEnable = true

		wf = newS3WorkloadFields()
		cfg = s3TestConfig()
	})

	AfterEach(func() {
		featuregate.GetFeatureGateUtilInstance().FeatureGateSidecarContainersEnable = fgPrev
	})

	expectParity := func() {
		GinkgoHelper()
		clusterOut := legacyOutput[[]corev1.Container]("GetS3container")
		cpOut, err := BuildS3InitContainers(s3InfoFromWorkloadFields(wf, cfg))
		Expect(err).To(BeNil())
		Expect(cpOut).To(Equal(clusterOut))
	}

	Context("no buckets", func() {
		It("should produce nothing on both sides when there are no s3 buckets", func() {
			clusterOut := legacyOutput[[]corev1.Container]("GetS3container")
			cpOut, err := BuildS3InitContainers(s3InfoFromWorkloadFields(wf, cfg))
			Expect(err).To(BeNil())
			Expect(clusterOut).To(BeEmpty())
			Expect(cpOut).To(BeEmpty())
		})
	})

	Context("credential variations", func() {
		It("public bucket with endpoint and no credentials", func() {
			addS3Bucket(wf, "s3-0", &flds.S3bucketItemValue{
				Bucket: ptr.To("public-bucket"),
				Path:   ptr.To("/container/everyones"),
				Url:    ptr.To("https://uswest.s3.amazonaws"),
			})
			expectParity()
		})

		It("plaintext access key id and secret access key", func() {
			addS3Bucket(wf, "s3-0", &flds.S3bucketItemValue{
				Bucket: ptr.To("b"),
				Path:   ptr.To("/p"),
				Url:    ptr.To("https://s3.amazonaws.com"),
				Key:    ptr.To("AKIAEXAMPLE"),
				Secret: ptr.To("plaintext-secret"),
			})
			expectParity()
		})

		It("secret-backed access key id and secret access key", func() {
			addS3Bucket(wf, "s3-0", &flds.S3bucketItemValue{
				Bucket:     ptr.To("private-bucket"),
				Path:       ptr.To("/container/my-bucket"),
				Url:        ptr.To("https://s3.amazonaws.com"),
				SecretName: ptr.To("my-access-key-secret"),
				Key:        ptr.To("AccessKeyId"),
				Secret:     ptr.To("SecretKey"),
			})
			expectParity()
		})

		It("secret-backed credentials with a session token", func() {
			addS3Bucket(wf, "s3-0", &flds.S3bucketItemValue{
				Bucket:       ptr.To("b"),
				Path:         ptr.To("/p"),
				SecretName:   ptr.To("s"),
				Key:          ptr.To("AccessKeyId"),
				Secret:       ptr.To("SecretKey"),
				SessionToken: ptr.To("my-session-token"),
			})
			expectParity()
		})

		It("session token without any access credentials", func() {
			addS3Bucket(wf, "s3-0", &flds.S3bucketItemValue{
				Bucket:       ptr.To("b"),
				Path:         ptr.To("/p"),
				SessionToken: ptr.To("tok"),
			})
			expectParity()
		})

		It("secret-backed with only the access key id key set", func() {
			addS3Bucket(wf, "s3-0", &flds.S3bucketItemValue{
				Bucket:     ptr.To("b"),
				Path:       ptr.To("/p"),
				SecretName: ptr.To("s"),
				Key:        ptr.To("AccessKeyId"),
			})
			expectParity()
		})

		It("secret-backed with only the secret access key key set", func() {
			addS3Bucket(wf, "s3-0", &flds.S3bucketItemValue{
				Bucket:     ptr.To("b"),
				Path:       ptr.To("/p"),
				SecretName: ptr.To("s"),
				Secret:     ptr.To("SecretKey"),
			})
			expectParity()
		})
	})

	Context("endpoint handling", func() {
		It("omits the endpoint flag when no url is provided", func() {
			addS3Bucket(wf, "s3-0", &flds.S3bucketItemValue{
				Bucket: ptr.To("b"),
				Path:   ptr.To("/p"),
			})
			expectParity()
		})
	})

	Context("run-as uid and gid", func() {
		It("threads both uid and gid into the goofys command", func() {
			wf.RunAsUid = flds.NewOptionalIntField(ptr.To(1000))
			wf.RunAsGid = flds.NewOptionalIntField(ptr.To(2000))
			addS3Bucket(wf, "s3-0", &flds.S3bucketItemValue{Bucket: ptr.To("b"), Path: ptr.To("/p")})
			expectParity()
		})

		It("threads only uid when gid is unset", func() {
			wf.RunAsUid = flds.NewOptionalIntField(ptr.To(1000))
			addS3Bucket(wf, "s3-0", &flds.S3bucketItemValue{Bucket: ptr.To("b"), Path: ptr.To("/p")})
			expectParity()
		})

		It("threads only gid when uid is unset", func() {
			wf.RunAsGid = flds.NewOptionalIntField(ptr.To(2000))
			addS3Bucket(wf, "s3-0", &flds.S3bucketItemValue{Bucket: ptr.To("b"), Path: ptr.To("/p")})
			expectParity()
		})
	})

	Context("multiple buckets and exclusions", func() {
		It("mirrors the cluster across three buckets in sorted order", func() {
			addS3Bucket(wf, "s3-0", &flds.S3bucketItemValue{
				Bucket: ptr.To("private-bucket"), Path: ptr.To("/container/my-bucket"), Url: ptr.To("https://s3.amazonaws.com"),
				SecretName: ptr.To("k"), Key: ptr.To("AccessKeyId"), Secret: ptr.To("SecretKey"),
			})
			addS3Bucket(wf, "s3-1", &flds.S3bucketItemValue{
				Bucket: ptr.To("public-bucket"), Path: ptr.To("/container/everyones"), Url: ptr.To("https://uswest.s3.amazonaws"),
			})
			addS3Bucket(wf, "s3-2", &flds.S3bucketItemValue{
				Bucket: ptr.To("token-bucket"), Path: ptr.To("/container/secret"), Url: ptr.To("https://s3.amazonaws.com"),
				SecretName: ptr.To("k"), Key: ptr.To("AccessKeyId"), Secret: ptr.To("SecretKey"), SessionToken: ptr.To("tok"),
			})
			expectParity()
		})

		It("keeps s3-fs index alignment when a nil item sorts before a real one", func() {
			addS3Bucket(wf, "s3-0", nil)
			addS3Bucket(wf, "s3-1", &flds.S3bucketItemValue{Bucket: ptr.To("b"), Path: ptr.To("/p")})
			expectParity()
		})
	})

	Context("custom CA", func() {
		It("mirrors the cluster when the s3 custom CA is enabled", func() {
			GinkgoT().Setenv("CUSTOM_CA_S3_ENABLED", "true")
			GinkgoT().Setenv("CUSTOM_CA_S3_SECRET", "runai-ca-cert")
			GinkgoT().Setenv("CUSTOM_CA_S3_KEY", "runai-ca.pem")

			caConfig := composer.StaticClusterStorageConfig{}
			caConfig.SetS3FsImage(s3ParityDefaultImage)
			caConfig.SetIsCustomCAS3Enabled(true)
			caConfig.SetCustomCAS3Secret("runai-ca-cert")
			caConfig.SetCustomCAS3Key("runai-ca.pem")
			cfg = caConfig

			addS3Bucket(wf, "s3-0", &flds.S3bucketItemValue{Bucket: ptr.To("b"), Path: ptr.To("/p")})
			expectParity()
		})
	})

	Context("documented per-bucket image override divergence", func() {
		It("CP uses the configured image while the cluster honors the reserved per-bucket image", func() {
			addS3Bucket(wf, "s3-0", &flds.S3bucketItemValue{
				Bucket: ptr.To("b"), Path: ptr.To("/p"), Image: ptr.To("custom-goofys:override"),
			})
			clusterOut := legacyOutput[[]corev1.Container]("GetS3container")
			cpOut, err := BuildS3InitContainers(s3InfoFromWorkloadFields(wf, cfg))
			Expect(err).To(BeNil())
			Expect(clusterOut[0].Image).To(Equal("custom-goofys:override"))
			Expect(cpOut[0].Image).To(Equal(s3ParityDefaultImage))

			clusterOut[0].Image = cpOut[0].Image
			Expect(cpOut[0]).To(Equal(clusterOut[0]), "modulo the image, CP and cluster output must be identical")
		})
	})
})

var _ = Describe("BuildS3Volumes and BuildS3MainVolumeMounts parity with the cluster builder", func() {
	var (
		wf  *field_groups.WorkloadFields
		cfg composer.ClusterStorageConfig
	)

	BeforeEach(func() {
		for _, name := range s3ParityEnvVars {
			GinkgoT().Setenv(name, "")
		}
		wf = newS3WorkloadFields()
		cfg = s3TestConfig()
	})

	// The CA volume is intentionally out of scope: the cluster adds it from its
	// generic volume path, not GetS3Volumes, so these specs cover the non-CA
	// shape only. CA mount parity is covered on the init container above.
	expectVolumeAndMountParity := func() {
		GinkgoHelper()
		info := s3InfoFromWorkloadFields(wf, cfg)
		Expect(BuildS3Volumes(info)).To(Equal(legacyOutput[[]corev1.Volume]("GetS3Volumes")))
		Expect(BuildS3MainVolumeMounts(info)).To(Equal(legacyOutput[[]corev1.VolumeMount]("GetS3VolumeMounts")))
	}

	It("single bucket", func() {
		addS3Bucket(wf, "s3-0", &flds.S3bucketItemValue{Bucket: ptr.To("b"), Path: ptr.To("/mnt/data")})
		expectVolumeAndMountParity()
	})

	It("multiple buckets in sorted order", func() {
		addS3Bucket(wf, "s3-0", &flds.S3bucketItemValue{Bucket: ptr.To("b0"), Path: ptr.To("/p0")})
		addS3Bucket(wf, "s3-1", &flds.S3bucketItemValue{Bucket: ptr.To("b1"), Path: ptr.To("/p1")})
		expectVolumeAndMountParity()
	})

	It("keeps index alignment when a nil item sorts before a real one", func() {
		addS3Bucket(wf, "s3-0", nil)
		addS3Bucket(wf, "s3-1", &flds.S3bucketItemValue{Bucket: ptr.To("b"), Path: ptr.To("/p")})
		expectVolumeAndMountParity()
	})
})
