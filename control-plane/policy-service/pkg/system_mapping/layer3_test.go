/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package system_mapping_test

// Layer 3+4 of the system-mapping test plan: hand-authored policy YAMLs
// that exercise the SET path of the system mapping, plus K8s OpenAPI
// validation of what they produce.
//
// Per-case flow:
//
//   1. Load testing/inputs/<gvk>/skeleton.yaml       — minimal valid manifest
//   2. Load testing/inputs/<gvk>/cases/<case>/policy.yaml — SET rules
//   3. Build the breakdown in-memory from the system mapping (no DB), so
//      the same Go data exercised by Layer 1/2 drives the engine.
//   4. Run the policy engine: ResolveRules + RunRules over the skeleton.
//   5. (Layer 3 assertion) the rendered manifest must equal the case's
//      testing/inputs/<gvk>/cases/<case>/golden.yaml.
//   6. (Layer 4 assertion) the rendered manifest must validate against the
//      checked-in K8s OpenAPI v3 schema for its GVK.
//
// Each (policy.yaml, golden.yaml) pair lives in the same directory so a
// reviewer can read what the policy sets next to what the engine should
// produce — the test is intentionally not a black box.
//
// In addition to the per-case round trips, two schema-loader self-tests
// validate the per-GVK reference.yaml (the comprehensive Layer 1/2 manifest)
// against the OpenAPI schema — proving the goldens we trust in Layer 1/2
// are themselves K8s-valid.

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/yaml"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	policyApi "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/generated"
)

const (
	// schemaName<GVK> is the canonical kube-openapi component name for each
	// built-in type we validate. K8s OpenAPI v3 keeps these under
	// components.schemas in the per-API-group documents under testing/crd/.
	schemaNameAppsDeploymentV1          = "io.k8s.api.apps.v1.Deployment"
	schemaNameBatchJobV1                = "io.k8s.api.batch.v1.Job"
	schemaNameAppsReplicaSetV1          = "io.k8s.api.apps.v1.ReplicaSet"
	schemaNameAppsStatefulSetV1         = "io.k8s.api.apps.v1.StatefulSet"
	schemaNameBatchCronJobV1            = "io.k8s.api.batch.v1.CronJob"
	schemaNameCorePodV1                 = "io.k8s.api.core.v1.Pod"
	schemaNameNimCacheV1alpha1          = "com.nvidia.apps.v1alpha1.NIMCache"
	schemaNameNimServiceV1alpha1        = "com.nvidia.apps.v1alpha1.NIMService"
	schemaNamePyTorchJobV1              = "org.kubeflow.v1.PyTorchJob"
	schemaNameMPIJobV2beta1             = "org.kubeflow.v2beta1.MPIJob"
	schemaNameTFJobV1                   = "org.kubeflow.v1.TFJob"
	schemaNameXGBoostJobV1              = "org.kubeflow.v1.XGBoostJob"
	schemaNameJAXJobV1                  = "org.kubeflow.v1.JAXJob"
	schemaNameLeaderWorkerSetV1         = "io.x-k8s.leaderworkerset.v1.LeaderWorkerSet"
	schemaNameDynamoGraphDeploymentV1a1 = "com.nvidia.v1alpha1.DynamoGraphDeployment"
	schemaNameRayClusterV1              = "io.ray.v1.RayCluster"
	schemaNameRayJobV1                  = "io.ray.v1.RayJob"
	schemaNameRayServiceV1              = "io.ray.v1.RayService"
	schemaNameKnativeServiceV1          = "dev.knative.serving.v1.Service"
	schemaNameRunAiWorkspaceV1          = "ai.run.v1.Workspace"
	schemaNameRunAiDistributedV1        = "ai.run.v1.Distributed"
)

// gvkConfig describes everything the per-GVK Layer 3+4 spec needs:
// the GVK itself, the per-GVK input directory (containing skeleton.yaml,
// reference.yaml, and the cases/ subdirectory), and the OpenAPI schema
// the rendered manifest must validate against.
type gvkConfig struct {
	gvk        api.Gvk
	inputDir   string
	schemaPath string
	schemaName string
	caseNames  []string

	// syntheticSpecRefs, when non-empty, tells the schema loader to build a
	// fictive manifest wrapper schema in code instead of looking up schemaName
	// in the OpenAPI document. It maps each top-level manifest field (e.g.
	// "spec", "masterSpec") to the name of the real component schema in the
	// document it should validate against. This is used for the run.ai/* GVKs,
	// which are fictive wrappers around the workloads service's native spec
	// schemas (WorkspaceSpecSpec, DistributedSpecSpec, MasterSpec) and thus
	// have no dedicated manifest schema in the (read-only) swagger doc.
	syntheticSpecRefs map[string]string
}

var deploymentConfig = gvkConfig{
	gvk:        api.Gvk{Group: convutils.Ref("apps"), Kind: "Deployment", Version: "v1"},
	inputDir:   "testing/inputs/apps_Deployment_v1",
	schemaPath: "testing/crd/apps_v1_openapi.json",
	schemaName: schemaNameAppsDeploymentV1,
	caseNames:  []string{"workload", "pod", "container"},
}

var jobConfig = gvkConfig{
	gvk:        api.Gvk{Group: convutils.Ref("batch"), Kind: "Job", Version: "v1"},
	inputDir:   "testing/inputs/batch_Job_v1",
	schemaPath: "testing/crd/batch_v1_openapi.json",
	schemaName: schemaNameBatchJobV1,
	caseNames:  []string{"workload", "pod", "container"},
}

var replicaSetConfig = gvkConfig{
	gvk:        api.Gvk{Group: convutils.Ref("apps"), Kind: "ReplicaSet", Version: "v1"},
	inputDir:   "testing/inputs/apps_ReplicaSet_v1",
	schemaPath: "testing/crd/apps_v1_openapi.json",
	schemaName: schemaNameAppsReplicaSetV1,
	caseNames:  []string{"workload", "pod", "container"},
}

var statefulSetConfig = gvkConfig{
	gvk:        api.Gvk{Group: convutils.Ref("apps"), Kind: "StatefulSet", Version: "v1"},
	inputDir:   "testing/inputs/apps_StatefulSet_v1",
	schemaPath: "testing/crd/apps_v1_openapi.json",
	schemaName: schemaNameAppsStatefulSetV1,
	caseNames:  []string{"workload", "pod", "container"},
}

var cronJobConfig = gvkConfig{
	gvk:        api.Gvk{Group: convutils.Ref("batch"), Kind: "CronJob", Version: "v1"},
	inputDir:   "testing/inputs/batch_CronJob_v1",
	schemaPath: "testing/crd/batch_v1_openapi.json",
	schemaName: schemaNameBatchCronJobV1,
	caseNames:  []string{"workload", "pod", "container"},
}

var podConfig = gvkConfig{
	gvk:        api.Gvk{Group: convutils.Ref(""), Kind: "Pod", Version: "v1"},
	inputDir:   "testing/inputs/core_Pod_v1",
	schemaPath: "testing/crd/core_v1_openapi.json",
	schemaName: schemaNameCorePodV1,
	caseNames:  []string{"pod", "container"},
}

var nimCacheConfig = gvkConfig{
	gvk:        api.Gvk{Group: convutils.Ref("apps.nvidia.com"), Kind: "NIMCache", Version: "v1alpha1"},
	inputDir:   "testing/inputs/apps_nvidia_com_NIMCache_v1alpha1",
	schemaPath: "testing/crd/apps_nvidia_com_v1alpha1_openapi.json",
	schemaName: schemaNameNimCacheV1alpha1,
	caseNames:  []string{"workload", "container"},
}

var nimServiceConfig = gvkConfig{
	gvk:        api.Gvk{Group: convutils.Ref("apps.nvidia.com"), Kind: "NIMService", Version: "v1alpha1"},
	inputDir:   "testing/inputs/apps_nvidia_com_NIMService_v1alpha1",
	schemaPath: "testing/crd/apps_nvidia_com_v1alpha1_openapi.json",
	schemaName: schemaNameNimServiceV1alpha1,
	caseNames:  []string{"workload", "container"},
}

var pyTorchJobConfig = gvkConfig{
	gvk:        api.Gvk{Group: convutils.Ref("kubeflow.org"), Kind: "PyTorchJob", Version: "v1"},
	inputDir:   "testing/inputs/kubeflow_org_PyTorchJob_v1",
	schemaPath: "testing/crd/kubeflow_org_v1_openapi.json",
	schemaName: schemaNamePyTorchJobV1,
	caseNames:  []string{"workload", "replica", "container"},
}

var mpiJobConfig = gvkConfig{
	gvk:        api.Gvk{Group: convutils.Ref("kubeflow.org"), Kind: "MPIJob", Version: "v2beta1"},
	inputDir:   "testing/inputs/kubeflow_org_MPIJob_v2beta1",
	schemaPath: "testing/crd/kubeflow_org_v2beta1_openapi.json",
	schemaName: schemaNameMPIJobV2beta1,
	caseNames:  []string{"workload", "replica", "container"},
}

var tfJobConfig = gvkConfig{
	gvk:        api.Gvk{Group: convutils.Ref("kubeflow.org"), Kind: "TFJob", Version: "v1"},
	inputDir:   "testing/inputs/kubeflow_org_TFJob_v1",
	schemaPath: "testing/crd/kubeflow_org_v1_openapi.json",
	schemaName: schemaNameTFJobV1,
	caseNames:  []string{"workload", "replica", "container"},
}

var xgBoostJobConfig = gvkConfig{
	gvk:        api.Gvk{Group: convutils.Ref("kubeflow.org"), Kind: "XGBoostJob", Version: "v1"},
	inputDir:   "testing/inputs/kubeflow_org_XGBoostJob_v1",
	schemaPath: "testing/crd/kubeflow_org_v1_openapi.json",
	schemaName: schemaNameXGBoostJobV1,
	caseNames:  []string{"workload", "replica", "container"},
}

var jaxJobConfig = gvkConfig{
	gvk:        api.Gvk{Group: convutils.Ref("kubeflow.org"), Kind: "JAXJob", Version: "v1"},
	inputDir:   "testing/inputs/kubeflow_org_JAXJob_v1",
	schemaPath: "testing/crd/kubeflow_org_v1_openapi.json",
	schemaName: schemaNameJAXJobV1,
	caseNames:  []string{"workload", "replica", "container"},
}

var leaderWorkerSetConfig = gvkConfig{
	gvk:        api.Gvk{Group: convutils.Ref("leaderworkerset.x-k8s.io"), Kind: "LeaderWorkerSet", Version: "v1"},
	inputDir:   "testing/inputs/leaderworkerset_x_k8s_io_LeaderWorkerSet_v1",
	schemaPath: "testing/crd/leaderworkerset_x_k8s_io_v1_openapi.json",
	schemaName: schemaNameLeaderWorkerSetV1,
	caseNames:  []string{"workload", "template", "pod", "container"},
}

var dynamoGraphDeploymentConfig = gvkConfig{
	gvk:        api.Gvk{Group: convutils.Ref("nvidia.com"), Kind: "DynamoGraphDeployment", Version: "v1alpha1"},
	inputDir:   "testing/inputs/nvidia_com_DynamoGraphDeployment_v1alpha1",
	schemaPath: "testing/crd/dynamographdeployment_nvidia_com_v1alpha1_openapi.json",
	schemaName: schemaNameDynamoGraphDeploymentV1a1,
	caseNames:  []string{"workload", "service", "container", "parameterized-dependency"},
}

var rayClusterConfig = gvkConfig{
	gvk:        api.Gvk{Group: convutils.Ref("ray.io"), Kind: "RayCluster", Version: "v1"},
	inputDir:   "testing/inputs/ray_io_RayCluster_v1",
	schemaPath: "testing/crd/ray_io_v1_openapi.json",
	schemaName: schemaNameRayClusterV1,
	caseNames:  []string{"workload", "worker", "container"},
}

var rayJobConfig = gvkConfig{
	gvk:        api.Gvk{Group: convutils.Ref("ray.io"), Kind: "RayJob", Version: "v1"},
	inputDir:   "testing/inputs/ray_io_RayJob_v1",
	schemaPath: "testing/crd/ray_io_v1_openapi.json",
	schemaName: schemaNameRayJobV1,
	caseNames:  []string{"workload", "worker", "container"},
}

var rayServiceConfig = gvkConfig{
	gvk:        api.Gvk{Group: convutils.Ref("ray.io"), Kind: "RayService", Version: "v1"},
	inputDir:   "testing/inputs/ray_io_RayService_v1",
	schemaPath: "testing/crd/ray_io_v1_openapi.json",
	schemaName: schemaNameRayServiceV1,
	caseNames:  []string{"workload", "worker", "container"},
}

var knativeServiceConfig = gvkConfig{
	gvk:        api.Gvk{Group: convutils.Ref("serving.knative.dev"), Kind: "Service", Version: "v1"},
	inputDir:   "testing/inputs/serving_knative_dev_Service_v1",
	schemaPath: "testing/crd/serving_knative_dev_v1_openapi.json",
	schemaName: schemaNameKnativeServiceV1,
	caseNames:  []string{"workload", "pod", "container"},
}

var runAiWorkspaceConfig = gvkConfig{
	gvk:        api.Gvk{Group: convutils.Ref("run.ai"), Kind: "Workspace", Version: "v1"},
	inputDir:   "testing/inputs/run_ai_Workspace_v1",
	schemaPath: "testing/crd/run_ai_v1_openapi.json",
	schemaName: schemaNameRunAiWorkspaceV1,
	caseNames:  []string{"workspaceSpec"},
	// run.ai/Workspace/v1 is a fictive GVK wrapping the workloads service's
	// WorkspaceSpecSpec under .spec.
	syntheticSpecRefs: map[string]string{"spec": "WorkspaceSpecSpec"},
}

var runAiDistributedConfig = gvkConfig{
	gvk:        api.Gvk{Group: convutils.Ref("run.ai"), Kind: "Distributed", Version: "v1"},
	inputDir:   "testing/inputs/run_ai_Distributed_v1",
	schemaPath: "testing/crd/run_ai_v1_openapi.json",
	schemaName: schemaNameRunAiDistributedV1,
	caseNames:  []string{"workerSpec", "masterSpec"},
	// run.ai/Distributed/v1 is a fictive GVK wrapping the workloads service's
	// DistributedSpecSpec under .spec.worker and MasterSpec under .spec.master
	// (mirroring the workloads UnifiedSpec {worker, master}). Keys are dot-paths
	// resolved into nested object schemas by resolveTargetSchema.
	syntheticSpecRefs: map[string]string{"spec.worker": "DistributedSpecSpec", "spec.master": "MasterSpec"},
}

// Serial: runLayer3 sets the package-global gvk_compliance.BrokenRuleHook
// during execution. Running specs in parallel would cause data races.
var _ = Describe("system mapping layer 3+4: hand-authored policies and K8s schema validation", Serial, func() {
	Context("schema loader self-tests", func() {
		It("validates the apps/Deployment/v1 reference manifest", func() {
			schema, err := loadOpenAPISchema(deploymentConfig.schemaPath, deploymentConfig.schemaName)
			Expect(err).To(BeNil())

			ref := loadReferenceSpecGinkgo(deploymentConfig.inputDir + "/reference.yaml")
			errs := validateAgainstOpenAPI(schema, ref)
			Expect(formatOpenAPIErrors(errs)).To(BeEmpty(),
				"reference manifest used by Layer 1/2 should already be K8s-valid")
		})

		It("validates the batch/Job/v1 reference manifest", func() {
			schema, err := loadOpenAPISchema(jobConfig.schemaPath, jobConfig.schemaName)
			Expect(err).To(BeNil())

			ref := loadReferenceSpecGinkgo(jobConfig.inputDir + "/reference.yaml")
			errs := validateAgainstOpenAPI(schema, ref)
			Expect(formatOpenAPIErrors(errs)).To(BeEmpty(),
				"reference manifest used by Layer 1/2 should already be K8s-valid")
		})

		It("validates the apps/ReplicaSet/v1 reference manifest", func() {
			schema, err := loadOpenAPISchema(replicaSetConfig.schemaPath, replicaSetConfig.schemaName)
			Expect(err).To(BeNil())

			ref := loadReferenceSpecGinkgo(replicaSetConfig.inputDir + "/reference.yaml")
			errs := validateAgainstOpenAPI(schema, ref)
			Expect(formatOpenAPIErrors(errs)).To(BeEmpty(),
				"reference manifest used by Layer 1/2 should already be K8s-valid")
		})

		It("validates the apps/StatefulSet/v1 reference manifest", func() {
			schema, err := loadOpenAPISchema(statefulSetConfig.schemaPath, statefulSetConfig.schemaName)
			Expect(err).To(BeNil())

			ref := loadReferenceSpecGinkgo(statefulSetConfig.inputDir + "/reference.yaml")
			errs := validateAgainstOpenAPI(schema, ref)
			Expect(formatOpenAPIErrors(errs)).To(BeEmpty(),
				"reference manifest used by Layer 1/2 should already be K8s-valid")
		})

		It("validates the batch/CronJob/v1 reference manifest", func() {
			schema, err := loadOpenAPISchema(cronJobConfig.schemaPath, cronJobConfig.schemaName)
			Expect(err).To(BeNil())

			ref := loadReferenceSpecGinkgo(cronJobConfig.inputDir + "/reference.yaml")
			errs := validateAgainstOpenAPI(schema, ref)
			Expect(formatOpenAPIErrors(errs)).To(BeEmpty(),
				"reference manifest used by Layer 1/2 should already be K8s-valid")
		})

		It("validates the /Pod/v1 reference manifest", func() {
			schema, err := loadOpenAPISchema(podConfig.schemaPath, podConfig.schemaName)
			Expect(err).To(BeNil())

			ref := loadReferenceSpecGinkgo(podConfig.inputDir + "/reference.yaml")
			errs := validateAgainstOpenAPI(schema, ref)
			Expect(formatOpenAPIErrors(errs)).To(BeEmpty(),
				"reference manifest used by Layer 1/2 should already be K8s-valid")
		})

		It("validates the apps.nvidia.com/NIMCache/v1alpha1 reference manifest", func() {
			schema, err := loadOpenAPISchema(nimCacheConfig.schemaPath, nimCacheConfig.schemaName)
			Expect(err).To(BeNil())

			ref := loadReferenceSpecGinkgo(nimCacheConfig.inputDir + "/reference.yaml")
			errs := validateAgainstOpenAPI(schema, ref)
			Expect(formatOpenAPIErrors(errs)).To(BeEmpty(),
				"reference manifest used by Layer 1/2 should already be K8s-valid")
		})

		It("validates the apps.nvidia.com/NIMService/v1alpha1 reference manifest", func() {
			schema, err := loadOpenAPISchema(nimServiceConfig.schemaPath, nimServiceConfig.schemaName)
			Expect(err).To(BeNil())

			ref := loadReferenceSpecGinkgo(nimServiceConfig.inputDir + "/reference.yaml")
			errs := validateAgainstOpenAPI(schema, ref)
			Expect(formatOpenAPIErrors(errs)).To(BeEmpty(),
				"reference manifest used by Layer 1/2 should already be K8s-valid")
		})

		It("validates the kubeflow.org/PyTorchJob/v1 reference manifest", func() {
			schema, err := loadOpenAPISchema(pyTorchJobConfig.schemaPath, pyTorchJobConfig.schemaName)
			Expect(err).To(BeNil())

			ref := loadReferenceSpecGinkgo(pyTorchJobConfig.inputDir + "/reference.yaml")
			errs := validateAgainstOpenAPI(schema, ref)
			Expect(formatOpenAPIErrors(errs)).To(BeEmpty(),
				"reference manifest used by Layer 1/2 should already be K8s-valid")
		})

		It("validates the kubeflow.org/MPIJob/v2beta1 reference manifest", func() {
			schema, err := loadOpenAPISchema(mpiJobConfig.schemaPath, mpiJobConfig.schemaName)
			Expect(err).To(BeNil())

			ref := loadReferenceSpecGinkgo(mpiJobConfig.inputDir + "/reference.yaml")
			errs := validateAgainstOpenAPI(schema, ref)
			Expect(formatOpenAPIErrors(errs)).To(BeEmpty(),
				"reference manifest used by Layer 1/2 should already be K8s-valid")
		})

		It("validates the kubeflow.org/TFJob/v1 reference manifest", func() {
			schema, err := loadOpenAPISchema(tfJobConfig.schemaPath, tfJobConfig.schemaName)
			Expect(err).To(BeNil())

			ref := loadReferenceSpecGinkgo(tfJobConfig.inputDir + "/reference.yaml")
			errs := validateAgainstOpenAPI(schema, ref)
			Expect(formatOpenAPIErrors(errs)).To(BeEmpty(),
				"reference manifest used by Layer 1/2 should already be K8s-valid")
		})

		It("validates the kubeflow.org/XGBoostJob/v1 reference manifest", func() {
			schema, err := loadOpenAPISchema(xgBoostJobConfig.schemaPath, xgBoostJobConfig.schemaName)
			Expect(err).To(BeNil())

			ref := loadReferenceSpecGinkgo(xgBoostJobConfig.inputDir + "/reference.yaml")
			errs := validateAgainstOpenAPI(schema, ref)
			Expect(formatOpenAPIErrors(errs)).To(BeEmpty(),
				"reference manifest used by Layer 1/2 should already be K8s-valid")
		})

		It("validates the kubeflow.org/JAXJob/v1 reference manifest", func() {
			schema, err := loadOpenAPISchema(jaxJobConfig.schemaPath, jaxJobConfig.schemaName)
			Expect(err).To(BeNil())

			ref := loadReferenceSpecGinkgo(jaxJobConfig.inputDir + "/reference.yaml")
			errs := validateAgainstOpenAPI(schema, ref)
			Expect(formatOpenAPIErrors(errs)).To(BeEmpty(),
				"reference manifest used by Layer 1/2 should already be K8s-valid")
		})

		It("validates the leaderworkerset.x-k8s.io/LeaderWorkerSet/v1 reference manifest", func() {
			schema, err := loadOpenAPISchema(leaderWorkerSetConfig.schemaPath, leaderWorkerSetConfig.schemaName)
			Expect(err).To(BeNil())

			ref := loadReferenceSpecGinkgo(leaderWorkerSetConfig.inputDir + "/reference.yaml")
			errs := validateAgainstOpenAPI(schema, ref)
			Expect(formatOpenAPIErrors(errs)).To(BeEmpty(),
				"reference manifest used by Layer 1/2 should already be K8s-valid")
		})

		It("validates the nvidia.com/DynamoGraphDeployment/v1alpha1 reference manifest", func() {
			schema, err := loadOpenAPISchema(dynamoGraphDeploymentConfig.schemaPath, dynamoGraphDeploymentConfig.schemaName)
			Expect(err).To(BeNil())

			ref := loadReferenceSpecGinkgo(dynamoGraphDeploymentConfig.inputDir + "/reference.yaml")
			errs := validateAgainstOpenAPI(schema, ref)
			Expect(formatOpenAPIErrors(errs)).To(BeEmpty(),
				"reference manifest used by Layer 1/2 should already be K8s-valid")
		})

		It("validates the ray.io/RayCluster/v1 reference manifest", func() {
			schema, err := loadOpenAPISchema(rayClusterConfig.schemaPath, rayClusterConfig.schemaName)
			Expect(err).To(BeNil())

			ref := loadReferenceSpecGinkgo(rayClusterConfig.inputDir + "/reference.yaml")
			errs := validateAgainstOpenAPI(schema, ref)
			Expect(formatOpenAPIErrors(errs)).To(BeEmpty(),
				"reference manifest used by Layer 1/2 should already be K8s-valid")
		})

		It("validates the ray.io/RayJob/v1 reference manifest", func() {
			schema, err := loadOpenAPISchema(rayJobConfig.schemaPath, rayJobConfig.schemaName)
			Expect(err).To(BeNil())

			ref := loadReferenceSpecGinkgo(rayJobConfig.inputDir + "/reference.yaml")
			errs := validateAgainstOpenAPI(schema, ref)
			Expect(formatOpenAPIErrors(errs)).To(BeEmpty(),
				"reference manifest used by Layer 1/2 should already be K8s-valid")
		})

		It("validates the ray.io/RayService/v1 reference manifest", func() {
			schema, err := loadOpenAPISchema(rayServiceConfig.schemaPath, rayServiceConfig.schemaName)
			Expect(err).To(BeNil())

			ref := loadReferenceSpecGinkgo(rayServiceConfig.inputDir + "/reference.yaml")
			errs := validateAgainstOpenAPI(schema, ref)
			Expect(formatOpenAPIErrors(errs)).To(BeEmpty(),
				"reference manifest used by Layer 1/2 should already be K8s-valid")
		})

		It("validates the serving.knative.dev/Service/v1 reference manifest", func() {
			schema, err := loadOpenAPISchema(knativeServiceConfig.schemaPath, knativeServiceConfig.schemaName)
			Expect(err).To(BeNil())

			ref := loadReferenceSpecGinkgo(knativeServiceConfig.inputDir + "/reference.yaml")
			errs := validateAgainstOpenAPI(schema, ref)
			Expect(formatOpenAPIErrors(errs)).To(BeEmpty(),
				"reference manifest used by Layer 1/2 should already be K8s-valid")
		})

		// NOTE: Enum constraints are only present in OpenAPI schemas downloaded
		// from a live K8s cluster. The static files in the kubernetes/kubernetes
		// GitHub repo lack enums — they're generated at runtime by kube-openapi
		// based on Go type annotations. Make sure to download schemas from a
		// cluster (kubectl get --raw /openapi/v3/apis/...) not from GitHub.

		It("rejects invalid enum values in Deployment strategy.type", func() {
			schema, err := loadOpenAPISchema(deploymentConfig.schemaPath, deploymentConfig.schemaName)
			Expect(err).To(BeNil())

			invalidManifest := map[string]any{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata":   map[string]any{"name": "test", "namespace": "default"},
				"spec": map[string]any{
					"selector": map[string]any{"matchLabels": map[string]any{"app": "test"}},
					"template": map[string]any{
						"metadata": map[string]any{"labels": map[string]any{"app": "test"}},
						"spec": map[string]any{
							"containers": []any{
								map[string]any{"name": "c", "image": "nginx"},
							},
						},
					},
					"strategy": map[string]any{
						"type": "InvalidStrategyType", // must be Recreate or RollingUpdate
					},
				},
			}
			errs := validateAgainstOpenAPI(schema, invalidManifest)
			Expect(errs).ToNot(BeEmpty(), "validator should reject invalid enum value")
			errStr := formatOpenAPIErrors(errs)
			Expect(errStr).To(ContainSubstring("strategy.type"),
				"error message should mention the field path")
			Expect(errStr).To(ContainSubstring("Recreate"),
				"error message should list valid enum values")
		})

		It("rejects invalid enum values in container imagePullPolicy", func() {
			schema, err := loadOpenAPISchema(deploymentConfig.schemaPath, deploymentConfig.schemaName)
			Expect(err).To(BeNil())

			invalidManifest := map[string]any{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata":   map[string]any{"name": "test", "namespace": "default"},
				"spec": map[string]any{
					"selector": map[string]any{"matchLabels": map[string]any{"app": "test"}},
					"template": map[string]any{
						"metadata": map[string]any{"labels": map[string]any{"app": "test"}},
						"spec": map[string]any{
							"containers": []any{
								map[string]any{
									"name":            "c",
									"image":           "nginx",
									"imagePullPolicy": "Sometimes", // must be Always, Never, or IfNotPresent
								},
							},
						},
					},
				},
			}
			errs := validateAgainstOpenAPI(schema, invalidManifest)
			Expect(errs).ToNot(BeEmpty(), "validator should reject invalid imagePullPolicy enum")
			errStr := formatOpenAPIErrors(errs)
			Expect(errStr).To(ContainSubstring("imagePullPolicy"),
				"error message should mention the field path")
			Expect(errStr).To(ContainSubstring("IfNotPresent"),
				"error message should list valid enum values")
		})

		It("rejects invalid enum values in Job completionMode", func() {
			schema, err := loadOpenAPISchema(jobConfig.schemaPath, jobConfig.schemaName)
			Expect(err).To(BeNil())

			invalidManifest := map[string]any{
				"apiVersion": "batch/v1",
				"kind":       "Job",
				"metadata":   map[string]any{"name": "test", "namespace": "default"},
				"spec": map[string]any{
					"completionMode": "Parallel", // must be NonIndexed or Indexed
					"template": map[string]any{
						"spec": map[string]any{
							"restartPolicy": "Never",
							"containers": []any{
								map[string]any{"name": "c", "image": "busybox"},
							},
						},
					},
				},
			}
			errs := validateAgainstOpenAPI(schema, invalidManifest)
			Expect(errs).ToNot(BeEmpty(), "validator should reject invalid completionMode enum")
			errStr := formatOpenAPIErrors(errs)
			Expect(errStr).To(ContainSubstring("completionMode"),
				"error message should mention the field path")
			Expect(errStr).To(ContainSubstring("Indexed"),
				"error message should list valid enum values")
		})

		It("rejects wrong type for integer field (replicas as string)", func() {
			schema, err := loadOpenAPISchema(deploymentConfig.schemaPath, deploymentConfig.schemaName)
			Expect(err).To(BeNil())

			invalidManifest := map[string]any{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata":   map[string]any{"name": "test", "namespace": "default"},
				"spec": map[string]any{
					"replicas": "not-an-integer",
					"selector": map[string]any{"matchLabels": map[string]any{"app": "test"}},
					"template": map[string]any{
						"metadata": map[string]any{"labels": map[string]any{"app": "test"}},
						"spec": map[string]any{
							"containers": []any{
								map[string]any{"name": "c", "image": "nginx"},
							},
						},
					},
				},
			}
			errs := validateAgainstOpenAPI(schema, invalidManifest)
			Expect(errs).ToNot(BeEmpty(), "validator should reject string where integer expected")
			Expect(formatOpenAPIErrors(errs)).To(ContainSubstring("replicas"),
				"error message should mention the invalid field")
		})

		It("rejects wrong type for boolean field (hostNetwork as string)", func() {
			schema, err := loadOpenAPISchema(deploymentConfig.schemaPath, deploymentConfig.schemaName)
			Expect(err).To(BeNil())

			invalidManifest := map[string]any{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata":   map[string]any{"name": "test", "namespace": "default"},
				"spec": map[string]any{
					"selector": map[string]any{"matchLabels": map[string]any{"app": "test"}},
					"template": map[string]any{
						"metadata": map[string]any{"labels": map[string]any{"app": "test"}},
						"spec": map[string]any{
							"hostNetwork": "yes", // should be boolean, not string
							"containers": []any{
								map[string]any{"name": "c", "image": "nginx"},
							},
						},
					},
				},
			}
			errs := validateAgainstOpenAPI(schema, invalidManifest)
			Expect(errs).ToNot(BeEmpty(), "validator should reject string where boolean expected")
			Expect(formatOpenAPIErrors(errs)).To(ContainSubstring("hostNetwork"),
				"error message should mention the invalid field")
		})

		It("rejects wrong type for array field (containers as object)", func() {
			schema, err := loadOpenAPISchema(deploymentConfig.schemaPath, deploymentConfig.schemaName)
			Expect(err).To(BeNil())

			invalidManifest := map[string]any{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata":   map[string]any{"name": "test", "namespace": "default"},
				"spec": map[string]any{
					"selector": map[string]any{"matchLabels": map[string]any{"app": "test"}},
					"template": map[string]any{
						"metadata": map[string]any{"labels": map[string]any{"app": "test"}},
						"spec": map[string]any{
							"containers": map[string]any{"name": "c", "image": "nginx"}, // should be array
						},
					},
				},
			}
			errs := validateAgainstOpenAPI(schema, invalidManifest)
			Expect(errs).ToNot(BeEmpty(), "validator should reject object where array expected")
			Expect(formatOpenAPIErrors(errs)).To(ContainSubstring("containers"),
				"error message should mention the invalid field")
		})

		It("rejects wrong type for object field (resources as string)", func() {
			schema, err := loadOpenAPISchema(deploymentConfig.schemaPath, deploymentConfig.schemaName)
			Expect(err).To(BeNil())

			invalidManifest := map[string]any{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata":   map[string]any{"name": "test", "namespace": "default"},
				"spec": map[string]any{
					"selector": map[string]any{"matchLabels": map[string]any{"app": "test"}},
					"template": map[string]any{
						"metadata": map[string]any{"labels": map[string]any{"app": "test"}},
						"spec": map[string]any{
							"containers": []any{
								map[string]any{
									"name":      "c",
									"image":     "nginx",
									"resources": "invalid", // should be object with limits/requests
								},
							},
						},
					},
				},
			}
			errs := validateAgainstOpenAPI(schema, invalidManifest)
			Expect(errs).ToNot(BeEmpty(), "validator should reject string where object expected")
			Expect(formatOpenAPIErrors(errs)).To(ContainSubstring("resources"),
				"error message should mention the invalid field")
		})

		It("rejects Job with wrong type for parallelism", func() {
			schema, err := loadOpenAPISchema(jobConfig.schemaPath, jobConfig.schemaName)
			Expect(err).To(BeNil())

			invalidManifest := map[string]any{
				"apiVersion": "batch/v1",
				"kind":       "Job",
				"metadata":   map[string]any{"name": "test", "namespace": "default"},
				"spec": map[string]any{
					"parallelism": "many", // should be integer
					"template": map[string]any{
						"spec": map[string]any{
							"restartPolicy": "Never",
							"containers": []any{
								map[string]any{"name": "c", "image": "busybox"},
							},
						},
					},
				},
			}
			errs := validateAgainstOpenAPI(schema, invalidManifest)
			Expect(errs).ToNot(BeEmpty(), "validator should reject string where integer expected")
			Expect(formatOpenAPIErrors(errs)).To(ContainSubstring("parallelism"),
				"error message should mention the invalid field")
		})
	})

	for _, cfg := range []gvkConfig{deploymentConfig, jobConfig, replicaSetConfig, statefulSetConfig, cronJobConfig, podConfig, nimCacheConfig, nimServiceConfig, pyTorchJobConfig, mpiJobConfig, tfJobConfig, xgBoostJobConfig, jaxJobConfig, leaderWorkerSetConfig, dynamoGraphDeploymentConfig, rayClusterConfig, rayJobConfig, rayServiceConfig, knativeServiceConfig, runAiWorkspaceConfig, runAiDistributedConfig} {
		cfg := cfg
		Context(cfg.gvk.Kind, func() {
			for _, name := range cfg.caseNames {
				name := name
				// Discover every policy variant in the case directory
				// (policy.yaml is the canonical form; policy1.yaml,
				// policy2.yaml, ... are alternate phrasings that must
				// produce the same golden output — e.g. sub-field SETs
				// that reach the same rendered manifest as the object-
				// level form in policy.yaml).
				for _, policyFile := range discoverPolicyFiles(cfg.inputDir + "/cases/" + name) {
					policyFile := policyFile
					label := policyVariantLabel(policyFile)
					It("case "+name+"/"+label+": rendered matches golden and validates against the schema", func() {
						runLayer3Case(cfg, name, policyFile)
					})
				}
			}
		})
	}
})

// discoverPolicyFiles lists every policy*.yaml file in a case directory,
// sorted lexicographically so policy.yaml is always first. Returns the
// full paths; failing tests if no policy file exists.
func discoverPolicyFiles(caseDir string) []string {
	matches, err := filepath.Glob(caseDir + "/policy*.yaml")
	Expect(err).To(BeNil(), "glob policy files under %s", caseDir)
	Expect(matches).ToNot(BeEmpty(), "expected at least one policy*.yaml under %s", caseDir)
	sort.Strings(matches)
	return matches
}

// policyVariantLabel turns a policy file path into a short label used in
// the test name: "policy.yaml" -> "policy", "policy1.yaml" -> "policy1".
func policyVariantLabel(path string) string {
	base := filepath.Base(path)
	return strings.TrimSuffix(base, ".yaml")
}

// runLayer3Case runs one Layer 3+4 case end-to-end for a specific policy
// variant. Loads the per-GVK skeleton, the case's hand-authored policy and
// golden, applies the policy over the skeleton, asserts equality with the
// golden and OpenAPI compliance of the rendered manifest.
//
// All policy variants in a case (policy.yaml, policy1.yaml, ...) share a
// single golden.yaml; they are meant to be different phrasings that must
// produce the same rendered output.
func runLayer3Case(cfg gvkConfig, caseName, policyFile string) {
	skeleton := loadReferenceSpecGinkgo(cfg.inputDir + "/skeleton.yaml")
	golden := loadReferenceSpecGinkgo(cfg.inputDir + "/cases/" + caseName + "/golden.yaml")

	rules, err := loadLayer3PolicyFile(policyFile)
	Expect(err).To(BeNil(), "load policy")
	Expect(rules).ToNot(BeEmpty(), "policy must declare at least one effective rule")

	policy, err := composeLayer3Policy(cfg.gvk, rules)
	Expect(err).To(BeNil(), "compose breakdown + rules into policy")

	result := runLayer3(context.Background(), policy, skeleton, len(rules)+10)

	Expect(result.BrokenRules).To(BeEmpty(),
		"engine reported broken rules:\n%s", formatLayer3BrokenRules(result.BrokenRules))
	Expect(result.Errors).To(BeEmpty(),
		"engine returned validation errors:\n%s", formatLayer3Errors(result.Errors))

	// Compare via YAML strings so failure messages include a line-by-line
	// unified diff (testify's assert.Equal does this automatically). Both
	// sides go through stripNilLeaves first — the engine substitutes every
	// $ARGS.named.<arg> in a setExpr template with JSON null when the arg
	// is absent, producing explicit null leaves a human-authored golden
	// would simply omit.
	expectedYaml, err := yaml.Marshal(stripNilLeavesMap(golden))
	Expect(err).To(BeNil(), "marshal expected golden to YAML")
	actualYaml, err := yaml.Marshal(stripNilLeavesMap(result.Workload))
	Expect(err).To(BeNil(), "marshal rendered workload to YAML")

	// LAYER3_WRITE_GOLDEN=1 rewrites canonical golden.yaml (policy variant
	// only) with the engine output instead of asserting equality. Use ONLY
	// when authoring a new case or after a deliberate mapping change —
	// always inspect the diff with `git diff` before committing the
	// regenerated goldens.
	if os.Getenv("LAYER3_WRITE_GOLDEN") == "1" && policyVariantLabel(policyFile) == "policy" {
		goldenPath := cfg.inputDir + "/cases/" + caseName + "/golden.yaml"
		Expect(os.WriteFile(goldenPath, actualYaml, 0o644)).To(Succeed(),
			"write golden under LAYER3_WRITE_GOLDEN")
		return
	}

	assert.Equal(GinkgoT(), string(expectedYaml), string(actualYaml),
		"rendered workload does not match golden for %s case %s (%s)",
		cfg.gvk.Kind, caseName, policyVariantLabel(policyFile))

	// Layer 4 (OpenAPI validation) is opt-in per gvkConfig: when schemaPath
	// is empty, the GVK runs Layer 3 (golden) only. Used for CRDs whose
	// OpenAPI schema has not yet been exported from a live cluster.
	if cfg.schemaPath == "" {
		return
	}

	schema, err := loadOpenAPISchemaWithSynthetic(cfg.schemaPath, cfg.schemaName, cfg.syntheticSpecRefs)
	Expect(err).To(BeNil(), "load OpenAPI schema for %s", cfg.schemaName)

	// Strip null leaves before schema validation for the same reason we
	// strip them before the golden comparison: the engine substitutes
	// every $ARGS.named.<arg> reference in setExpr templates with JSON
	// null when the arg is absent, producing artifacts the K8s API
	// server itself does not see (real submitters omit those fields).
	openAPIErrs := validateAgainstOpenAPI(schema, stripNilLeavesMap(result.Workload))
	Expect(formatOpenAPIErrors(openAPIErrs)).To(BeEmpty(),
		"rendered workload failed K8s OpenAPI validation")
}

// stripNilLeaves recursively removes nil values from JSON-shaped data.
// For maps: drops keys whose value is nil after recursion. For slices:
// recurses into elements and drops resulting nil elements. For scalars:
// passes through.
//
// Returning the value (rather than mutating in place) keeps the input
// untouched so failure messages can still print the un-normalized
// rendered manifest if needed.
func stripNilLeaves(v any) any {
	switch x := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, vv := range x {
			cleaned := stripNilLeaves(vv)
			if cleaned == nil {
				continue
			}
			out[k] = cleaned
		}
		if len(out) == 0 {
			// An empty map after stripping would otherwise differ from a
			// missing field; preserve emptiness only when the input was
			// already empty.
			if len(x) == 0 {
				return out
			}
			return nil
		}
		return out
	case []any:
		out := make([]any, 0, len(x))
		for _, vv := range x {
			cleaned := stripNilLeaves(vv)
			if cleaned == nil {
				continue
			}
			out = append(out, cleaned)
		}
		return out
	case nil:
		return nil
	default:
		return v
	}
}

// stripNilLeavesMap is stripNilLeaves typed for map[string]any so callers
// can use it directly on workload manifests.
func stripNilLeavesMap(m map[string]any) map[string]any {
	out, _ := stripNilLeaves(m).(map[string]any)
	return out
}

func formatLayer3Errors(errs []*policyApi.Error) string {
	var s string
	for _, e := range errs {
		if e == nil {
			continue
		}
		s += "  - " + e.Message + "\n"
	}
	return s
}

func formatLayer3BrokenRules(brs []layer3BrokenRule) string {
	var s string
	for _, b := range brs {
		s += "  - " + b.Field + " (" + b.SpecSelector + "): " + b.Message + "\n"
	}
	return s
}
