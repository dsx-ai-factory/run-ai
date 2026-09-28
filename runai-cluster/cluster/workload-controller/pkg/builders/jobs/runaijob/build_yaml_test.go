/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaijob

import (
	"context"
	"strings"
	"testing"

	clusterconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	runaiv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v2"
	workloadconfig "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/config"
	workloadproject "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/project"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/common"
	builderscommon "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/common"
	commontesting "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/common/testing"
	jobstesting "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/jobs/testing"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils"
	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const namespace = "runai-team-a"

func init() {
	if err := kaiorg.AddToScheme(clientscheme.Scheme); err != nil {
		panic(err)
	}
}

func initClientWithResources(t *testing.T, namespace string) client.Client {

	fakeDataVolumePvc := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "data-volume-pvc",
			Namespace: namespace,
			Labels: map[string]string{
				v1alpha1.DataVolumeIdLabel: dataVolumeId,
			},
		},
	}

	fakePvc := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "existing",
			Namespace: namespace,
		},
	}

	fakePvc2 := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "existing2",
			Namespace: namespace,
		},
	}

	fakePvc3 := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "existing3",
			Namespace: namespace,
		},
	}

	volumeMode := corev1.PersistentVolumeBlock
	fakePvcBlock := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "existing-block",
			Namespace: namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			VolumeMode: &volumeMode,
		},
	}

	projectName := strings.TrimPrefix(namespace, "runai-")

	fakeProject := runaiv2.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: projectName,
		},
		Spec: runaiv2.ProjectSpec{
			InteractiveJobTimeLimitSecs:                  120,
			InteractiveJobMaxIdleDurationSecs:            1200,
			InteractivePreemptibleJobMaxIdleDurationSecs: 1800,
			TrainingJobMaxIdleDurationSecs:               2400,
			TrainingJobTimeLimitSecs:                     120,
		},
	}

	// kai.resources twin of fakeProject so the builder finds its fixture whether
	// the kai project carries the limits in annotations.
	fakeKaiProject := kaiorg.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: projectName,
			Annotations: map[string]string{
				clusterconstants.InteractiveJobTimeLimitSecondsAnnotation:                  "120",
				clusterconstants.InteractiveJobMaxIdleDurationSecondsAnnotation:            "1200",
				clusterconstants.InteractivePreemptibleJobMaxIdleDurationSecondsAnnotation: "1800",
				clusterconstants.TrainingJobMaxIdleDurationSecondsAnnotation:               "2400",
				clusterconstants.TrainingJobTimeLimitSecondsAnnotation:                     "120",
			},
		},
	}

	fakeCredential := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-user-credential",
			Namespace: "runai",
		},
		Data: map[string][]byte{
			// byte array with encrypted "private-value-1"
			"my-user-key": []byte{
				168, 221, 244, 73, 135, 225, 25, 104, 145, 32, 118, 20, 196, 29, 72, 165,
				224, 40, 97, 180, 83, 157, 205, 140, 158, 186, 60, 73, 173, 193, 146, 219,
				201, 199, 65, 36, 241, 251, 52, 131, 38, 155, 0,
			},
		},
	}

	kaiDefaultNodePool := kaiorg.NodePool{ObjectMeta: metav1.ObjectMeta{Name: "default"}}
	kaiFakeNodePool := kaiorg.NodePool{
		ObjectMeta: metav1.ObjectMeta{Name: "nodepoola"},
		Spec:       kaiorg.NodePoolSpec{LabelKey: "test-key", LabelValue: "test-value"},
	}

	fakeClient := commontesting.NewFakeClient(t,
		&fakePvc, &fakePvc2, &fakePvc3, &fakePvcBlock,
		&fakeProject, &fakeKaiProject,
		&kaiDefaultNodePool, &kaiFakeNodePool,
		&fakeDataVolumePvc, &fakeCredential)
	return fakeClient
}
func TestBuildersRunaijob(t *testing.T) {
	toTest := []common.TestConfig{
		{TestName: "basic_params"},
		{TestName: "restart_policy_interactive"},
		{TestName: "capabilities"},
		{TestName: "capabilities-complex"},
		{TestName: "data_volumes"},
		{TestName: "basic_memory"},
		{TestName: "ext_and_std_resources"},
		{TestName: "ext_resources"},
		{TestName: "from_workspace"},
		{TestName: "from_training"},
		{TestName: "git_cli"},
		{TestName: "home_of_user"},
		{TestName: "volumes"},
		{TestName: "existing-pvc-multiple-mounts"},
		{TestName: "multi-frac"},
		{TestName: "multi-frac-memory"},
		{TestName: "jupyter_basic"},
		{TestName: "jupyter_change_command"},
		{TestName: "jupyter_from_asset"},
		{TestName: "jupyter_ingress", TestEnvVars: &common.TestEnvVars{RedisFQDN: "cluster-redis.runai.svc.cluster.local:6379"}},
		{TestName: "jupyter_with_authorized_groups", TestEnvVars: &common.TestEnvVars{RedisFQDN: "cluster-redis.runai.svc.cluster.local:6379"}},
		{TestName: "jupyter_with_authorized_users", TestEnvVars: &common.TestEnvVars{RedisFQDN: "cluster-redis.runai.svc.cluster.local:6379"}},
		{TestName: "multi_exposed_urls_with_authorized_users", TestEnvVars: &common.TestEnvVars{RedisFQDN: "cluster-redis.runai.svc.cluster.local:6379"}},
		{TestName: "jupyter_with_authorized_users_and_groups", TestEnvVars: &common.TestEnvVars{RedisFQDN: "cluster-redis.runai.svc.cluster.local:6379"}},
		{TestName: "jupyter_ingress_with_subdomain", TestEnvVars: &common.TestEnvVars{SubdomainSupportEnvVarValue: "true", RedisFQDN: "cluster-redis.runai.svc.cluster.local:6379"}},
		{TestName: "jupyter_ingress_with_annotations", TestEnvVars: &common.TestEnvVars{IngressClassEnvVarValue: "nginx", IngressAnnotationsEnvVarValue: "{\"kubernetes.io/ingress.class\": \"nginx\"}", RedisFQDN: "cluster-redis.runai.svc.cluster.local:6379"}},
		{TestName: "jupyter_run_as_user"},
		{TestName: "jupyter_run_as_user_non_auth", TestEnvVars: &common.TestEnvVars{JupyterUserIdentityEnvVar: "false"}},
		{TestName: "limits_and_more"},
		{TestName: "ports_cluster_ip"},
		{TestName: "ports_forwarding"},
		{TestName: "ports_ingress", TestEnvVars: &common.TestEnvVars{RedisFQDN: "cluster-redis.runai.svc.cluster.local:6379"}},
		{TestName: "ports_ingress_with_subdomain", TestEnvVars: &common.TestEnvVars{SubdomainSupportEnvVarValue: "true", RedisFQDN: "cluster-redis.runai.svc.cluster.local:6379"}},
		{TestName: "ports_mixture"},
		{TestName: "ports_ingress_with_path", TestEnvVars: &common.TestEnvVars{RedisFQDN: "cluster-redis.runai.svc.cluster.local:6379"}},
		{TestName: "ports_load_balancer"},
		{TestName: "pvc_access_mode"},
		{TestName: "s3_bucket"},
		{TestName: "s3_bucket_from_secret"},
		{TestName: "s3_bucket_custom_ca", TestEnvVars: &common.TestEnvVars{CustomCAS3EnvVarValue: "true"}},
		{TestName: "storage"},
		{TestName: "secret_volume"},
		{TestName: "emptydir_volume"},
		{TestName: "tensorboard_basic"},
		{TestName: "tensorboard_ingress"},
		{TestName: "tensorboard_ingress_with_subdomain", TestEnvVars: &common.TestEnvVars{SubdomainSupportEnvVarValue: "true"}},
		{TestName: "terminate_after_preemption"},
		{TestName: "tolerations"},
		{TestName: "ephemeral"},
		{TestName: "ephemeral_and_pvc"},
		{TestName: "ephemeral_default_access_mode"},
		{TestName: "ephemeral_many"},
		{TestName: "suspended_job"},
		{TestName: "configmap_volume"},
		{TestName: "probes"},
		{TestName: "privileged_true"},
		{TestName: "privileged_false"},
		{TestName: "preemptibility"},
		{TestName: "jupyter_ingess_custom_ca", TestEnvVars: &common.TestEnvVars{CustomCAEnvVarValue: "true", RedisFQDN: "cluster-redis.runai.svc.cluster.local:6379"}},
		{TestName: "port-wo-service"},
		{TestName: "image_pull_secrets", TestEnvVars: &common.TestEnvVars{ADSEncryptionKey: common.ADSEncryptionKey}},
	}
	runaiBuilder, ctx, commonBuilder := createBuilders(t, nil, nil)
	for _, curTest := range common.FilterTests(toTest) {

		//   for testing we do not want real uuid... impossible to compare expected vs actual,
		//   so we use a mock uuid generator
		utils.ActiveUuidGenerator = &jobstesting.MockUuidGenerator{}

		common.TestJobBuilder(ctx, t, &curTest, runaiBuilder, commonBuilder, &TestRunaiJobBuilder{}, commonBuilder)
	}
}

func createBuilders(t *testing.T,
	podLabels map[string]string, podAnnotations map[string]string) (common.ComputeBuilder, context.Context, *builderscommon.SharedResourcesBuilder) {
	k8sClient := initClientWithResources(t, namespace)
	runaiBuilder := NewRunaiJobBuilder(k8sClient, podLabels, podAnnotations, true)
	projectName := strings.TrimPrefix(namespace, "runai-")
	ctx := workloadproject.SetProjectInContext(context.Background(), projectName)
	pvcBuilder := common.NewPvcBuilder(k8sClient)
	serviceBuilder := common.NewServiceBuilder(k8sClient)
	urlExporterBuilder := common.NewExternalsUrlsBuilder(k8sClient, workloadconfig.IsOpenshift())
	credentialsBuilder := common.NewCredentialsBuilder(k8sClient)
	commonBuilder := common.NewShardResourcesBuilder(serviceBuilder, urlExporterBuilder, pvcBuilder, credentialsBuilder)
	return runaiBuilder, ctx, commonBuilder
}
