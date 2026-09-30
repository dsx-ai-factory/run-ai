/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package integration_test

import (
	"context"
	"fmt"
	"sort"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/composer_v1/runai_workloads"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/composer_v2"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/composer_v2/iw_tw_workloads"
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"

	clustersApi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants/oauth2proxy"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/builders"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/config"
	featuregate "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/feature-gate"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/project"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	corev1 "k8s.io/api/core/v1"
)

const inputsBasePath = "../composer_v2/iw_tw_workloads"

const gitSyncParityImage = "registry.example/git-sync:test"

// Cluster builder reads these from process env; anyworkload reads from
// ClusterStorageConfig. Clear them so a stray runner override can't flake parity.
var gitSyncEnvVars = []string{
	"GIT_SYNC_IMAGE",
	"CUSTOM_CA_GIT_ENABLED",
	"CUSTOM_CA_GIT_KEY",
	"CUSTOM_CA_GIT_SECRET",
}

var _ = Describe("native vs anyworkload composer parity for git-sync", func() {
	BeforeEach(func() {
		for _, name := range gitSyncEnvVars {
			GinkgoT().Setenv(name, "")
		}
		GinkgoT().Setenv(config.EnvVarGitSyncImage, gitSyncParityImage)
	})

	DescribeTable("init containers from both composers must match for the same SupersetSpec",
		func(fixtureName string) {
			ctx := context.Background()
			spec := loadSupersetSpec(fixtureName)
			options := buildSharedOptions(ctx)

			clusterInit := initContainersFromNativePath(ctx, spec, options, fixtureName)
			anyworkloadInit := initContainersFromAnyworkloadPath(ctx, spec, options, fixtureName)

			Expect(anyworkloadInit).To(Equal(clusterInit),
				"anyworkload composer's init containers diverged from what the cluster "+
					"would build off the native composer's StorageFields.GitSync output",
			)
		},
		Entry("git-only, default branch", "git"),
		Entry("git with explicit branch", "git_with_branch"),
		Entry("git with password auth", "git_with_password"),
		Entry("git with SSH key auth", "git_with_ssh_key"),
	)
})

// The cluster builders read these from env; the anyworkload path reads them from
// ClusterStorageConfig. Clear them, then pin the proxy image env to match the
// reported image so both sides agree.
var oauth2EnvVars = []string{
	"RUNAI_OAUTH2_PROXY_IMAGE",
	"RUNAI_OAUTH2_PROXY_WORKLOAD_AUTH_SECRET_NAME",
	"RUNAI_OAUTH2_PROXY_SESSION_STORE_TYPE",
	"RUNAI_OAUTH2_PROXY_REDIS_FQDN",
	"RUNAI_OAUTH2_PROXY_USERNAME_LOADER_IMAGE",
	"OAUTH2_PROXY_OIDC_ISSUER_URL",
	"ENDPOINT",
	"IS_OPENSHIFT",
	"SUBDOMAIN_SUPPORT",
	"CUSTOM_CA_ENABLED",
	"EXTERNAL_AUTH_URL_ENABLED",
}

var _ = Describe("native vs anyworkload composer parity for oauth2-proxy", func() {
	var fgPrev bool

	BeforeEach(func() {
		for _, name := range oauth2EnvVars {
			GinkgoT().Setenv(name, "")
		}
		GinkgoT().Setenv(builders.ExternalUrlAuthContainerImageEnv, oauth2ParityImage)
		GinkgoT().Setenv(config.EnvVarOauth2ProxyUsernameLoaderImage, oauth2ParityUsernameLoaderImage)
		GinkgoT().Setenv(oauth2proxy.TemplatesConfigMapEnvVar, oauth2proxy.ConfigMapName)
		// The cluster only stamps the oauth2-proxy sidecar RestartPolicy when the
		// sidecar-containers feature gate is on; the anyworkload constructor sets
		// it unconditionally at the K8s 1.29 floor. Enable the gate so both agree.
		fg := featuregate.GetFeatureGateUtilInstance()
		fgPrev = fg.FeatureGateSidecarContainersEnable
		fg.FeatureGateSidecarContainersEnable = true
	})

	AfterEach(func() {
		featuregate.GetFeatureGateUtilInstance().FeatureGateSidecarContainersEnable = fgPrev
	})

	DescribeTable("oauth2-proxy and username-loader init containers from both composers must match for the same SupersetSpec",
		func(fixtureName string, workloadType string, subdomain bool) {
			ctx := project.SetProjectInContext(context.Background(), oauth2ParityProject)
			GinkgoT().Setenv(config.EnvVarClusterDomain, oauth2ParityClusterDomain)
			if subdomain {
				GinkgoT().Setenv(config.EnvVarSubdomainSupport, "true")
			}

			spec := loadSupersetSpec(fixtureName)
			options := buildSharedOptions(ctx)

			clusterInit := oauth2InitContainersFromNativePath(ctx, spec, options, fixtureName, workloadType)
			anyworkloadInit := oauth2InitContainersFromAnyworkloadPath(ctx, spec, options, fixtureName, workloadType, subdomain)

			Expect(anyworkloadInit).To(Equal(clusterInit),
				"anyworkload composer's oauth2 init containers diverged from what the cluster "+
					"would build off the native composer's ExposedUrls output",
			)
		},
		Entry("auth-only workspace", "auth_url", string(workloadsApi.WORKLOADTYPE_WORKSPACE), false),
		Entry("auth-only training", "auth_url", string(workloadsApi.WORKLOADTYPE_TRAINING), false),
		Entry("auth with authorized users workspace", "auth_url_with_users", string(workloadsApi.WORKLOADTYPE_WORKSPACE), false),
		Entry("auth with authorized users training", "auth_url_with_users", string(workloadsApi.WORKLOADTYPE_TRAINING), false),
		Entry("auto-generated url with path routing workspace", "auth_url_auto", string(workloadsApi.WORKLOADTYPE_WORKSPACE), false),
		Entry("auto-generated url with path routing training", "auth_url_auto", string(workloadsApi.WORKLOADTYPE_TRAINING), false),
		Entry("auto-generated url with subdomain routing workspace", "auth_url_auto", string(workloadsApi.WORKLOADTYPE_WORKSPACE), true),
		Entry("auto-generated urls only the first suffixless workspace", "auth_url_auto_multi", string(workloadsApi.WORKLOADTYPE_WORKSPACE), false),
	)
})

// Cluster builder reads these from process env; anyworkload reads from
// ClusterStorageConfig. Clear them so a stray override can't flake parity.
var s3EnvVars = []string{
	"S3_FILE_SYSTEM_IMAGE",
	"CUSTOM_CA_S3_ENABLED",
	"CUSTOM_CA_S3_KEY",
	"CUSTOM_CA_S3_SECRET",
}

const s3ParityImage = "test.registry/goofys:parity"

var _ = Describe("native vs anyworkload composer parity for s3", func() {
	BeforeEach(func() {
		for _, name := range s3EnvVars {
			GinkgoT().Setenv(name, "")
		}
		// The cluster builder resolves the goofys image from S3_FILE_SYSTEM_IMAGE,
		// while the anyworkload path takes it from ClusterStorageConfig. Point the
		// env var at that same default so the comparison isolates builder logic.
		GinkgoT().Setenv("S3_FILE_SYSTEM_IMAGE", s3ParityImage)
	})

	DescribeTable("s3 init containers from both composers must match for the same SupersetSpec",
		func(fixtureName string, workloadType string) {
			ctx := context.Background()
			spec := loadSupersetSpec(fixtureName)
			options := buildSharedOptions(ctx)

			clusterInit := s3InitContainersFromNativePath(ctx, spec, options, fixtureName, workloadType)
			anyworkloadInit := s3InitContainersFromAnyworkloadPath(ctx, spec, options, fixtureName, workloadType)

			Expect(anyworkloadInit).To(Equal(clusterInit),
				"anyworkload composer's s3 init containers diverged from what the cluster "+
					"would build off the native composer's StorageFields.S3buckets output",
			)
		},
		Entry("s3 workspace", "s3", string(workloadsApi.WORKLOADTYPE_WORKSPACE)),
		Entry("s3 training", "s3", string(workloadsApi.WORKLOADTYPE_TRAINING)),
		Entry("s3 with run-as uid/gid", "s3_run_as_user", string(workloadsApi.WORKLOADTYPE_WORKSPACE)),
	)
})

func loadSupersetSpec(fixtureName string) workloadsApi.SupersetSpec {
	GinkgoHelper()
	specYaml, err := testutils.LoadInputFile(fixtureName, inputsBasePath)
	Expect(err).NotTo(HaveOccurred(), "failed to read fixture %s.yaml", fixtureName)

	var spec workloadsApi.SupersetSpec
	Expect(yaml.Unmarshal(specYaml, &spec)).To(Succeed(),
		"failed to unmarshal fixture %s.yaml", fixtureName)
	return spec
}

func loadInferenceSpec(fixtureName string) workloadsApi.InferenceSpecSpec {
	GinkgoHelper()
	specYaml, err := testutils.LoadInputFile(fixtureName, inputsBasePath)
	Expect(err).NotTo(HaveOccurred(), "failed to read fixture %s.yaml", fixtureName)

	var spec workloadsApi.InferenceSpecSpec
	Expect(yaml.Unmarshal(specYaml, &spec)).To(Succeed(),
		"failed to unmarshal fixture %s.yaml", fixtureName)
	return spec
}

// Both composers must receive identical options for parity to be meaningful.
// SetClusterDependentOptions flips version-gated flags (notably
// GitSyncWithSshKeySupported, needed by the SSH-key Entry).
func buildSharedOptions(ctx context.Context) composer.FieldsComposerOptions {
	GinkgoHelper()
	caps, err := checker.NewClusterCapabilitiesCheckerForVersion(testutils.DefaultClusterVersionForTesting)
	Expect(err).NotTo(HaveOccurred(), "failed to build cluster capabilities checker")

	options := composer.FieldsComposerOptions{UserCredentialsSupported: true}
	Expect(composer_v2.SetClusterDependentOptions(ctx, &options, caps)).To(Succeed(),
		"failed to set cluster-dependent composer options")
	return options
}

// SecurityContext nil-out and isJupyter=false compensate for the two
// documented CP-vs-cluster divergences listed on BuildGitSyncInitContainers.
func initContainersFromNativePath(
	ctx context.Context,
	spec workloadsApi.SupersetSpec,
	options composer.FieldsComposerOptions,
	workloadName string,
) []corev1.Container {
	GinkgoHelper()
	oldComposer := runai_workloads.NewSupersetToJobSpecComposer(
		string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		workloadName,
		"tester",
		nil,
		nil,
		options,
	)

	jobSpec := &jobs.JobSpec{}
	Expect(oldComposer.Compose(ctx, convutils.Ref(spec), jobSpec)()).To(BeNil(),
		"native composer returned an error for fixture %s", workloadName)

	containers := builders.GetGitSyncInitContainers(
		ctx,
		&jobSpec.ContainerFields,
		&jobSpec.StorageFields,
		false,
	)

	for i := range containers {
		containers[i].SecurityContext = nil
	}
	return containers
}

func initKnativeFromNativePath(
	ctx context.Context,
	spec workloadsApi.InferenceSpecSpec,
	options composer.FieldsComposerOptions,
	workloadName string,
) []corev1.Container {
	GinkgoHelper()
	oldComposer := runai_workloads.NewSupersetToJobSpecComposer(
		string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		workloadName,
		"tester",
		nil,
		nil,
		options,
	)

	var supersetSpec workloadsApi.SupersetSpec
	err := convutils.CopySpecs(&supersetSpec, &spec)
	Expect(err).NotTo(HaveOccurred(), "failed to copy native spec to SupersetSpec")

	jobSpec := &jobs.JobSpec{}
	Expect(oldComposer.Compose(ctx, &supersetSpec, jobSpec)()).To(BeNil(),
		"native composer returned an error for fixture %s", workloadName)

	containers := builders.GetGitSyncInitContainers(
		ctx,
		&jobSpec.ContainerFields,
		&jobSpec.StorageFields,
		false,
	)

	for i := range containers {
		containers[i].SecurityContext = nil
	}
	return containers
}

func initContainersFromAnyworkloadPath(
	ctx context.Context,
	spec workloadsApi.SupersetSpec,
	options composer.FieldsComposerOptions,
	workloadName string,
) []corev1.Container {
	GinkgoHelper()
	oldComposer := iw_tw_workloads.NewInterTrainJobComposer(
		string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		workloadName,
		"test",
		"runai-test",
		"tester",
		nil,
		options,
		nil,
	).WithClusterStorageConfig(gitSyncParityStorageConfig())

	target := iw_tw_workloads.NewMinimalBatchJobTarget()
	Expect(oldComposer.Compose(ctx, convutils.Ref(spec), target)()).To(BeNil(),
		"anyworkload composer returned an error for fixture %s", workloadName)

	return target.Job.Spec.Template.Spec.InitContainers
}

const (
	oauth2ParityNamespace           = "runai-test"
	oauth2ParityProject             = "test"
	oauth2ParityClusterDomain       = "https://test.example.com"
	oauth2ParityImage               = "test.registry/oauth2-proxy:parity"
	oauth2ParityUsernameLoaderImage = "test.registry/username-loader:parity"
)

func gitSyncParityStorageConfig() composer.ClusterStorageConfig {
	config := composer.DefaultClusterStorageConfig().(composer.StaticClusterStorageConfig)
	config.SetGitSyncImage(gitSyncParityImage)
	return config
}

func oauth2ParityStorageConfig(subdomain bool) composer.ClusterStorageConfig {
	reported := clustersApi.ClusterReportedConfig{}
	reported.SetRunaiOauth2ProxyImage(oauth2ParityImage)
	reported.SetRunaiOauth2ProxyUsernameLoaderImage(oauth2ParityUsernameLoaderImage)
	reported.SetSubdomainEnabled(subdomain)

	status := clustersApi.ClusterDisplayedStatus{}
	status.SetConfig(reported)

	cluster := clustersApi.DisplayedCluster{}
	cluster.Status.Set(&status)
	cluster.SetDomain(oauth2ParityClusterDomain)
	return composer.NewClusterStorageConfigFromCluster(&cluster)
}

// oauth2InitContainersFromNativePath reproduces how the cluster's runaijob pod
// builder turns the native composer's ExposedUrls field into init containers:
// the username-loader first (BuildPodSpecInitContainers) followed by one
// oauth2-proxy sidecar per exposed URL (BuildPodSpecContainers, appended to
// init containers when the sidecar gate is on).
func oauth2InitContainersFromNativePath(
	ctx context.Context,
	spec workloadsApi.SupersetSpec,
	options composer.FieldsComposerOptions,
	workloadName string,
	workloadType string,
) []corev1.Container {
	GinkgoHelper()
	oldComposer := runai_workloads.NewSupersetToJobSpecComposer(
		workloadType,
		workloadName,
		"tester",
		nil,
		nil,
		options,
	)

	jobSpec := &jobs.JobSpec{}
	Expect(oldComposer.Compose(ctx, convutils.Ref(spec), jobSpec)()).To(BeNil(),
		"native composer returned an error for fixture %s", workloadName)

	exposedUrls := jobSpec.WorkloadFields.ExposedUrls.Deref()

	result := make([]corev1.Container, 0)
	if loader := builders.GetUsernameLoaderInitContainer(exposedUrls, nil); loader != nil {
		result = append(result, *loader)
	}

	nicks := make([]string, 0, len(exposedUrls))
	for nick := range exposedUrls {
		nicks = append(nicks, nick)
	}
	sort.Strings(nicks)

	noSuffixUrlGenerated := false
	for idx, nick := range nicks {
		exposedUrl := exposedUrls[nick]
		if exposedUrl.Value == nil {
			continue
		}

		jobUrl, err := builders.ComposeOauthIngressUrl(ctx, exposedUrl, nick, &noSuffixUrlGenerated, workloadName)
		Expect(err).NotTo(HaveOccurred(), "ComposeOauthIngressUrl failed for fixture %s", workloadName)

		cookiePrefix := fmt.Sprintf("%s-%s-auth-proxy", workloadName, nick)
		container, err := builders.GetOauth2ProxyContainer(jobUrl, *exposedUrl.Value.ContainerPort,
			exposedUrl.Value.AuthorizedUsers, exposedUrl.Value.AuthorizedGroups, idx, nick, false, false, cookiePrefix)
		Expect(err).NotTo(HaveOccurred(), "GetOauth2ProxyContainer failed for fixture %s", workloadName)

		result = append(result, *container)
	}

	return result
}

func oauth2InitContainersFromAnyworkloadPath(
	ctx context.Context,
	spec workloadsApi.SupersetSpec,
	options composer.FieldsComposerOptions,
	workloadName string,
	workloadType string,
	subdomain bool,
) []corev1.Container {
	GinkgoHelper()
	oldComposer := iw_tw_workloads.NewInterTrainJobComposer(
		workloadType,
		workloadName,
		oauth2ParityProject,
		oauth2ParityNamespace,
		"tester",
		nil,
		options,
		nil,
	)
	oldComposer.WithClusterStorageConfig(oauth2ParityStorageConfig(subdomain))

	target := iw_tw_workloads.NewMinimalBatchJobTarget()
	Expect(oldComposer.Compose(ctx, convutils.Ref(spec), target)()).To(BeNil(),
		"anyworkload composer returned an error for fixture %s", workloadName)

	return target.Job.Spec.Template.Spec.InitContainers
}

// The cluster builder only stamps RestartPolicy when the sidecar-containers
// feature gate is on; the anyworkload constructor sets it unconditionally at the
// K8s 1.29 floor. Enable the gate here so both agree. Unlike git-sync, the s3
// container's SecurityContext (Privileged) is part of the contract and is kept.
func s3InitContainersFromNativePath(
	ctx context.Context,
	spec workloadsApi.SupersetSpec,
	options composer.FieldsComposerOptions,
	workloadName string,
	workloadType string,
) []corev1.Container {
	GinkgoHelper()
	oldComposer := runai_workloads.NewSupersetToJobSpecComposer(
		workloadType,
		workloadName,
		"tester",
		nil,
		nil,
		options,
	)

	jobSpec := &jobs.JobSpec{}
	Expect(oldComposer.Compose(ctx, convutils.Ref(spec), jobSpec)()).To(BeNil(),
		"native composer returned an error for fixture %s", workloadName)

	featuregate.GetFeatureGateUtilInstance().FeatureGateSidecarContainersEnable = true

	return builders.GetS3container(ctx, &jobSpec.WorkloadFields)
}

func s3InitContainersFromAnyworkloadPath(
	ctx context.Context,
	spec workloadsApi.SupersetSpec,
	options composer.FieldsComposerOptions,
	workloadName string,
	workloadType string,
) []corev1.Container {
	GinkgoHelper()
	oldComposer := iw_tw_workloads.NewInterTrainJobComposer(
		workloadType,
		workloadName,
		"test",
		"runai-test",
		"tester",
		nil,
		options,
		nil,
	).WithClusterStorageConfig(s3ParityStorageConfig())

	target := iw_tw_workloads.NewMinimalBatchJobTarget()
	Expect(oldComposer.Compose(ctx, convutils.Ref(spec), target)()).To(BeNil(),
		"anyworkload composer returned an error for fixture %s", workloadName)

	return target.Job.Spec.Template.Spec.InitContainers
}

func s3ParityStorageConfig() composer.ClusterStorageConfig {
	config := composer.DefaultClusterStorageConfig().(composer.StaticClusterStorageConfig)
	config.SetS3FsImage(s3ParityImage)
	return config
}
