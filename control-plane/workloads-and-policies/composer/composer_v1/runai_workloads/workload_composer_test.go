/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runai_workloads

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/composer_v2"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/notifications"
	k8sObjectsApi "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"

	dinfwv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ghodss/yaml"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/stretchr/testify/assert"
)

const assetUuid = "123e4567-e89b-42d3-a456-ab1234567890"

type ComposerTest struct {
	workloadType string

	specFname          string // values of the workload OR defaults of the policy
	expectedError      string
	expectedFname      string
	storageClassConfig string

	expectedRequirements composer.WorkloadRequirements

	assetsInfo *composer.ComplimentaryAssetInfo

	clusterVersion string

	distributed *DistributedTest
}

type DistributedTest struct {
	masterSpecFname string

	numWorkers int32
	spec       workloadsApi.DistributedSpecSpec

	masterAssetsInfo *composer.ComplimentaryAssetInfo
}

var ComposerTests = append([]TableEntry{
	Entry("basic workspace", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "basic",
		expectedFname: "basic_ws",
	}),

	Entry("basic workspace", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED),
		specFname:     "basic",
		expectedError: "mandatory distributedFramework is not provided",
	}),

	Entry("basic workspace PC", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "basic_pc_preemptible",
		expectedFname: "basic_pc_preemptible_ws",
	}),

	Entry("env empty", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "env_empty_value",
		expectedFname: "env_empty_value_ws",
	}),

	Entry("env empty", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "env_no_name",
		expectedError: "environmentVariables: name is required",
	}),

	Entry("env no value", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "env_no_value",
		expectedError: "environmentVariables 'HOME': value, userCredential, configMap, podFieldRef or secret must be set",
	}),

	Entry("parallelism", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "parallelism",
		expectedFname: "parallelism_ws",
	}),

	Entry("completions", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "completions",
		expectedFname: "completions_ws",
	}),

	Entry("category", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "category",
		expectedFname: "category_ws",
	}),

	Entry("annotations", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "annotations",
		expectedFname: "annotations_ws",
	}),

	Entry("annotations no name", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "annotations_no_name",
		expectedError: "annotations: name must be provided",
	}),

	Entry("labels no name", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "labels_no_name",
		expectedError: "labels: name must be provided",
	}),

	Entry("auto_deletion", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_TRAINING),
		specFname:     "auto_deletion",
		expectedFname: "auto_deletion_tr",
	}),

	Entry("labels", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_TRAINING),
		specFname:     "labels",
		expectedFname: "labels_tr",
	}),

	Entry("rich", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "rich",
		expectedFname: "rich_ws",
	}),

	Entry("node_type", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_TRAINING),
		specFname:     "node_type",
		expectedFname: "node_type_tr",
	}),

	Entry("ports", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "ports",
		expectedFname: "ports_ws",
	}),

	Entry("urls", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "urls",
		expectedFname: "urls_ws",
	}),

	Entry("urls_syntax", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "urls_syntax",
		expectedError: "malformed url: parse \"http://example.com:badport\": invalid port",
	}),

	Entry("urls_host_syntax", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "urls_host_syntax",
		expectedError: "malformed url host",
	}),

	Entry("urls_public", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "urls_authenticated_users",
		expectedFname: "urls_authenticated_users_ws",
	}),

	Entry("urls_by_group_no_groups", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "urls_by_groups_with_no_groups",
		expectedError: "exposedUrls: authorizedGroups must be set when authorizationType is authorizedGroups",
	}),

	Entry("urls_by_users_no_users", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "urls_by_users_no_users",
		expectedError: "exposedUrls: authorizedUsers must be set when authorizationType is authorizedUsers",
	}),

	Entry("urls_not_abs", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "urls_not_abs",
		expectedFname: "urls_not_abs_ws",
		expectedError: "Invalid url <https://compute.on-demand.io> must be an absolute path",
	}),

	Entry("urls_invalid", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_TRAINING),
		specFname:     "urls_invalid_authorized",
		expectedError: "authorizedUsers and authorizedGroups are mutually exclusive",
	}),

	Entry("ports_and_urls", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "ports_and_urls",
		expectedFname: "ports_and_urls_ws",
	}),

	Entry("priority_class_build", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "priority_class_build",
		expectedFname: "priority_class_build_ws",
	}),

	Entry("priority_class_build_tr", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_TRAINING),
		specFname:     "priority_class_build",
		expectedFname: "priority_class_build_tr",
	}),

	Entry("priority_class_preemptible_tr", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_TRAINING),
		specFname:     "priority_class_preemptible",
		expectedFname: "priority_class_preemptible_tr",
	}),

	Entry("priority_class_train", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_TRAINING),
		specFname:     "priority_class_train",
		expectedFname: "priority_class_train_tr",
	}),

	Entry("priority_class_preemptible", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "priority_class_preemptible",
		expectedFname: "priority_class_preemptible_ws",
	}),

	Entry("preemptibility", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "preemptibility",
		expectedFname: "preemptibility_ws",
	}),

	Entry("ports_tf", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED),
		specFname:     "ports",
		expectedFname: "ports_tf",

		distributed: &DistributedTest{
			numWorkers: 3,
			spec: workloadsApi.DistributedSpecSpec{
				DistributedFramework: *workloadsApi.NewNullableDistributedFramework(convutils.Ref(workloadsApi.DISTRIBUTEDFRAMEWORK_TF)),
			},
			masterSpecFname: "ports_and_urls",
		},
	}),

	Entry("distributed ports urls", ComposerTest{
		workloadType: string(workloadsApi.WORKLOADTYPE_DISTRIBUTED),
		specFname:    "dist_ports_urls_worker",

		distributed: &DistributedTest{
			numWorkers: 3,
			spec: workloadsApi.DistributedSpecSpec{
				DistributedFramework: *workloadsApi.NewNullableDistributedFramework(convutils.Ref(workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH)),
			},
			masterSpecFname: "dist_ports_urls_master",
		},

		expectedFname: "dist_ports_urls_pytorch",
	}),

	Entry("pod affinity - required", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "pod_affinity_required",
		expectedFname: "pod_affinity_required_ws",
	}),

	Entry("privileged true", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "privileged_true",
		expectedFname: "privileged_true",
	}),

	Entry("privileged true and allow privileged false", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "privileged_and_allow_privilege_escalation",
		expectedError: "allowPrivilegeEscalation can't be false when privileged is true",
	}),

	Entry("privileged true and run as non root true", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "privileged_and_run_as_non_root",
		expectedError: "runAsNonRoot can't be true when privileged is true",
	}),

	Entry("pod affinity - preferred", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "pod_affinity_preferred",
		expectedFname: "pod_affinity_preferred_ws",
	}),

	Entry("pod affinity - no type", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "pod_affinity_no_type",
		expectedError: "podAffinity: type must be provided",
	}),

	Entry("pod affinity - no key", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "pod_affinity_no_key",
		expectedError: "podAffinity: key must be provided",
	}),

	Entry("tolerations", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "tolerations",
		expectedFname: "tolerations_ws",
	}),

	Entry("tolerations - Any effect", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "tolerations_effect_any",
		expectedFname: "tolerations_effect_any_ws",
	}),

	Entry("tolerations - invalid - no operator", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "tolerations_invalid_no_operator",
		expectedError: "tolerations: operator must be provided",
	}),

	Entry("tolerations - invalid - no key", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "tolerations_invalid_no_key",
		expectedError: "tolerations: key must be provided",
	}),

	Entry("tolerations - invalid - no effect", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "tolerations_invalid_no_effect",
		expectedError: "tolerations: effect must be provided",
	}),

	Entry("tolerations - invalid - equal operator without value", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "tolerations_invalid_equal_without_value",
		expectedError: "tolerations: value must be provided when operator is set to Equal",
	}),

	Entry("tolerations - invalid - exists operator with value", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "tolerations_invalid_exists_with_value",
		expectedError: "tolerations: value is not allowed when operator is set to Exists",
	}),

	Entry("tolerations - invalid - seconds with wrong effect", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "tolerations_invalid_seconds_with_wrong_effect",
		expectedError: "tolerations: seconds is not allowed when effect is not set to NoExecute",
	}),

	Entry("probes", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "probes",
		expectedFname: "probes_ws",
	}),

	Entry("probes - only required fields", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "probes_only_required",
		expectedFname: "probes_only_required_ws",
	}),

	Entry("probes - invalid - no handler", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "probes_invalid_no_handler",
		expectedError: "probes.readiness: handler must be provided",
	}),

	Entry("probes - invalid - httpGet handler without port", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "probes_invalid_no_httpget_port",
		expectedError: "probes.readiness.handler.httpGet: port must be provided",
	}),

	Entry("secret_and_config_env", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "secret_and_config_env",
		expectedFname: "secret_and_config_env",
	}),

	Entry("user_credentials_env", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "user_credentials_env",
		expectedFname: "user_credentials_env",
	}),

	Entry("user_credentials_env_no_asset", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "user_credentials_env_no_asset",
		expectedError: "asset mapping not found for credential name: my-user-credential-not-found",
	}),

	Entry("secret_env_err", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "secret_env_err",
		expectedError: "environmentVariables 'PASSWORD': value, userCredential, configMap, podFieldRef and secret are mutually exclusive",
	}),

	Entry("config_env_err", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "config_env_err",
		expectedError: "environmentVariables 'PASSWORD': value, userCredential, configMap, podFieldRef and secret are mutually exclusive",
	}),

	Entry("pod_field_ref_env_err", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "pod_field_ref_env_err",
		expectedError: "environmentVariables 'PASSWORD': value, userCredential, configMap, podFieldRef and secret are mutually exclusive",
	}),

	Entry("node_affinity", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "node_affinity",
		expectedFname: "node_affinity_ws",
	}),

	Entry("node_affinity_invalid", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "node_affinity_invalid",
		expectedError: "nodeAffinityRequired: Values must be empty when Operator is Exists",
	}),

	Entry("image_pull_secrets", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "image_pull_secrets",
		expectedFname: "image_pull_secrets_ws",
	}),
},
	append(computeTests,
		append(distributedTests,
			append(storageTests,
				append(storageClassConfigTests,
					append(inferenceTests,
						distributedInferenceTests...,
					)...,
				)...,
			)...,
		)...,
	)...,
)

var lettersAndColon = regexp.MustCompile(`([A-Za-z]+:)`)

type ActualResults struct {
	JobSpec      *jobs.JobSpec
	Dinfw        *dinfwv1.DistributedInferenceWorkload
	MasterSpec   *jobs.JobSpec
	Notification *workloadsApi.WorkloadCreationFields
}

func RunComposerTest(testDetails ComposerTest, basePath string) {
	testBrief := fmt.Sprintf("Value: %s, Rules: , Type: %s", testDetails.specFname, testDetails.workloadType)

	testDetailsStr := lettersAndColon.ReplaceAllString(fmt.Sprintf("%+v", testDetails), "\n\t$1")

	var actualResults ActualResults
	var actualRequirements composer.WorkloadRequirements

	var supersetSpec workloadsApi.SupersetSpec

	if testDetails.specFname != testutils.NA {
		specYaml, err := testutils.LoadInputFile(testDetails.specFname, basePath)
		Expect(err).To(BeNil(), fmt.Sprintf("failed to load of %s.yaml", testDetails.specFname))

		err = yaml.Unmarshal(specYaml, &supersetSpec)
		if err != nil {
			if os.IsNotExist(err) {
				err = fmt.Errorf("file %s.yaml not found", testDetails.specFname)
			}
		}
		Expect(err).To(BeNil(), fmt.Sprintf("failed to parse of %s.yaml", testDetails.specFname))
	}

	var masterSpec *workloadsApi.MasterSpec

	if testDetails.distributed != nil && testDetails.distributed.masterSpecFname != "" {
		specYaml, err := testutils.LoadInputFile(testDetails.distributed.masterSpecFname, basePath)
		Expect(err).To(BeNil(), fmt.Sprintf("failed to load of %s.yaml", testDetails.distributed.masterSpecFname))

		masterSpec = &workloadsApi.MasterSpec{}

		err = yaml.Unmarshal(specYaml, &masterSpec)
		Expect(err).To(BeNil(), fmt.Sprintf("failed to parse of %s.yaml", testDetails.distributed.masterSpecFname))
	}

	clusterVersion := testutils.DefaultClusterVersionForTesting
	if testDetails.clusterVersion != "" {
		clusterVersion = testDetails.clusterVersion
	}

	testutils.RandStringTestMode("p")

	var storageClassConfig map[string]k8sObjectsApi.StorageClassConfiguration

	if testDetails.storageClassConfig != "" {
		fname := "storage-class-config/" + testDetails.storageClassConfig

		configYaml, err := testutils.LoadInputFile(fname, basePath)
		Expect(err).To(BeNil(), fmt.Sprintf("expecting to load %s.yaml", fname))

		err = yaml.Unmarshal(configYaml, &storageClassConfig)
		Expect(err).To(BeNil(), fmt.Sprintf("failed to parse of %s.yaml", fname))
	}

	clusterCapabilities, checkerErr := checker.NewClusterCapabilitiesCheckerForVersion(clusterVersion)
	Expect(checkerErr).To(BeNil())

	assetsMapping := []workloadsApi.AssetsMapping{
		{
			Name: convutils.Ref("my-user-credential"),
			Id:   convutils.Ref(assetUuid),
			Kind: convutils.Ref(workloadsApi.ASSETKIND_NGC_API_KEY),
		},
	}

	options := composer.FieldsComposerOptions{
		UserCredentialsSupported: true,
	}

	checkerErr = composer_v2.SetClusterDependentOptions(context.Background(), &options, clusterCapabilities)
	Expect(checkerErr).To(BeNil(), "failed to set cluster dependent options")

	notificationsComposer := notifications.NewNotificationsComposer(
		testDetails.workloadType,
		"test",
		fmt.Sprintf("id-%s", testBrief),
		"test",
		"cluster-123-456",
		"project-654-321",
		"department-456-123",
		time.Time{},
		nil,
		options).WithConnectionsPreProcessor("project", "")

	var err *workloadsApi.Error

	switch testDetails.workloadType {
	case string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE):
		dInfComposer := NewDinfWorkloadComposer(
			testDetails.specFname,
			"test",
			notificationsComposer,
			options)

		dInfComposer.fc.WithStorageClassConfiguration(storageClassConfig)

		actualResults.JobSpec = &jobs.JobSpec{} // needed only for backward compatibility with the old tests
		actualResults.Dinfw = &dinfwv1.DistributedInferenceWorkload{}

		err = dInfComposer.Compose(context.Background(), &supersetSpec, actualResults.Dinfw)

		actualRequirements = dInfComposer.Requirements()
	case string(workloadsApi.WORKLOADTYPE_DISTRIBUTED):
		if testDetails.distributed != nil {
			supersetSpec.NumWorkers = *workloadsApi.NewNullableInt32(&testDetails.distributed.numWorkers)
			supersetSpec.DistributedFramework = testDetails.distributed.spec.DistributedFramework
			supersetSpec.SlotsPerWorker = testDetails.distributed.spec.SlotsPerWorker
			supersetSpec.SshAuthMountPath = testDetails.distributed.spec.SshAuthMountPath
			supersetSpec.MpiLauncherCreationPolicy = testDetails.distributed.spec.MpiLauncherCreationPolicy
			supersetSpec.MinReplicas = testDetails.distributed.spec.MinReplicas
			supersetSpec.MaxReplicas = testDetails.distributed.spec.MaxReplicas
			supersetSpec.CleanPodPolicy = testDetails.distributed.spec.CleanPodPolicy
		}

		distributedComposer := NewDistributedWorkloadComposer(
			testDetails.specFname, "test", notificationsComposer, options)

		distributedComposer.
			WithAssetMapping(assetsMapping).
			WithComplimentaryAssetInfo(workloadsApi.REPLICATYPE_WORKER, testDetails.assetsInfo)

		if testDetails.distributed != nil {
			distributedComposer.WithComplimentaryAssetInfo(workloadsApi.REPLICATYPE_MASTER, testDetails.distributed.masterAssetsInfo)
		}

		distributedComposer.fc.WithStorageClassConfiguration(storageClassConfig)

		distributedWorkload := workloadsApi.SupersetWorkload{
			Spec: &supersetSpec,
		}

		if masterSpec != nil {
			distributedWorkload.MasterSpec = *workloadsApi.NewNullableMasterSpec(masterSpec)
		}

		var targetWorkload jobs.JobWorkload

		err = distributedComposer.Compose(context.Background(), &distributedWorkload, &targetWorkload)
		if err == nil {
			actualResults.JobSpec = &targetWorkload.Spec
			actualResults.MasterSpec = targetWorkload.MasterSpec
		}

		actualRequirements = distributedComposer.Requirements()
	default:
		jobSpecComposer := NewSupersetToJobSpecComposer(
			testDetails.workloadType,
			testDetails.specFname,
			"test",
			nil,
			assetsMapping,
			options)

		jobSpecComposer.fc.WithStorageClassConfiguration(storageClassConfig)

		actualResults.JobSpec = &jobs.JobSpec{}

		err = jobSpecComposer.Compose(context.Background(), convutils.Ref(supersetSpec), actualResults.JobSpec)()
		if err == nil {
			err = notificationsComposer.Compose(context.Background(), nil, testDetails.assetsInfo, convutils.Ref(supersetSpec))()
		}
		actualRequirements = jobSpecComposer.WorkloadRequirements
	}

	//
	//   notification
	//
	actualResults.Notification = notificationsComposer.Output

	// for testing purposes only we change the name to contain some information about the test
	actualResults.Notification.Name = testBrief

	if noResourcesRequested(actualResults.Notification.RequestedResources.Get()) {
		actualResults.Notification.RequestedResources.Unset()
	}

	//
	//   check composer error
	//
	if testDetails.expectedError != testutils.NA {
		Expect(err).NotTo(BeNil(), fmt.Sprintf("expected message was not returned: %s", testDetailsStr))
		if err != nil {
			Expect(err.Message).To(ContainSubstring(testDetails.expectedError), testDetailsStr)
		}
		return
	} else {
		Expect(err).To(BeNil(), testDetailsStr)
	}

	CompareFiles(testDetailsStr, &actualResults, testDetails.expectedFname, basePath)

	Expect(actualRequirements).To(Equal(testDetails.expectedRequirements), testDetailsStr)
}

func CompareFiles(testDetailsStr string, actualResults *ActualResults, expectedFname string, basePath string) {
	var actualYaml []byte
	actualYaml, marshalError := yaml.Marshal(actualResults)
	Expect(marshalError).To(BeNil(), testDetailsStr)

	expectedYaml := testutils.LoadExpectedFile(expectedFname, actualYaml, basePath)

	isOk := assert.Equal(GinkgoT(), string(expectedYaml), string(actualYaml), testDetailsStr)
	if !isOk {
		tmpPath := fmt.Sprintf("%s/%s.yaml", os.TempDir(), expectedFname)
		osError := os.WriteFile(tmpPath, actualYaml, 0600)
		if osError != nil {
			assert.Fail(GinkgoT(), fmt.Sprintf("failed to backup the result to %s: %v", tmpPath, osError.Error()))
		} else {
			assert.Fail(GinkgoT(), fmt.Sprintf("actual result backed up to %s", tmpPath))
		}
	}
}

var _ = Describe("superset composer validations", func() {
	Context("input VS expected tests", func() {
		DescribeTable("should create proper workloads or return sensible errors", func(testDetails ComposerTest) {
			RunComposerTest(testDetails, ".")
		}, ComposerTests)
	})
})

func noResourcesRequested(resources *workloadsApi.WorkloadRequestResources) bool {
	if resources == nil {
		return true
	}
	if resources.GpuRequestType.IsSet() || resources.Gpu.IsSet() || resources.GpuMemory.IsSet() {
		return false
	}
	if resources.Cpu.IsSet() || resources.CpuMemory.IsSet() {
		return false
	}
	if len(resources.ExtendedResources) != 0 {
		return false
	}
	return true
}
