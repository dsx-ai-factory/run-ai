/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	clusterAPI "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	clustersmock "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/clusterservice"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/clusterservice"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models_crud"
	workloadstest "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
	workloadMgrApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerclient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"knative.dev/serving/pkg/apis/autoscaling"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

func TestInferenceMinimumUpdate(t *testing.T) {
	const clusterID = "inference-scaling"
	cluster := workloadstest.BuildDisplayedClusterWithKnative(clusterID, "3.0.0", true, true, true)
	knative := cluster.Status.Get().GetDependencies().Optional["knative"]
	features := (*knative.Components)["knative"].Components
	(*features)[checker.ZeroInitialScaleDependency] = clusterAPI.ClusterDependencyStatus{Available: true}
	priorClient, priorService := checkerclient.Get(), clusterservice.Get()
	// Version checks still see the old report without the startup feature.
	checkerclient.SetMock(checkerclient.NewSimpleMockClient().WithCluster(clusterID, workloadstest.BuildDisplayedClusterWithKnative(clusterID, "3.0.0", true, true, true)))
	checker.InitClustersCache(workloadstest.GetDefaultClustersCacheConfig())
	clusterservice.SetMock(clustersmock.NewSimpleMock().WithCluster(clusterID, &cluster))
	t.Cleanup(func() {
		clusterservice.SetMock(priorService)
		checkerclient.SetMock(priorClient)
		checker.InitClustersCache(workloadstest.GetDefaultClustersCacheConfig())
	})

	var spec workloadsApi.SupersetSpec
	require.NoError(t, json.Unmarshal([]byte(`{"image":"example/model:v1","autoscaling":{"minReplicas":1,"maxReplicas":3,"metric":"concurrency","metricThreshold":10},"environmentVariables":[{"name":"HF_TOKEN","value":"<test-value>"}]}`), &spec))
	wc := &workloadContext{clusterCapabilities: checker.NewClusterCapabilitiesChecker(clusterID)}
	existing := &workloadMgrApi.WorkloadV2GetResponse{}
	existing.Workload.Metadata.ProjectName = "project"
	workload := &models_crud.Workload{Name: "model", CreatedBy: "tester", Namespace: "runai-project", ClusterId: clusterID, Spec: &spec}
	var requiresServing bool
	render := func() *servingv1.Service {
		t.Helper()
		update, updateErr := (&AnyWorkloadSubmitterInference{}).Update(t.Context(), wc, "id", existing, workload)
		require.Nil(t, updateErr)
		requiresServing = update.RequiresServingRevision
		data, err := json.Marshal(update.Request.Manifest)
		require.NoError(t, err)
		var service servingv1.Service
		require.NoError(t, json.Unmarshal(data, &service))
		// The public workloads-manager GET strips these literals from RawManifest.
		public := service.DeepCopy()
		public.Spec.Template.Spec.Containers[0].Env[0].Value = ""
		existing.Workload.RawManifest, err = gojsonq.LoadJson(public)
		require.NoError(t, err)
		return &service
	}
	patch := func(body string) *servingv1.Service {
		t.Helper()
		var request api.InferenceUpdateRequest
		require.NoError(t, json.Unmarshal([]byte(body), &request))
		previous := spec
		wc.previousInferenceSpec = &previous
		require.Nil(t, mergeUpdateSpecIntoFullSpec(t.Context(), &spec, &request.Spec))
		return render()
	}

	baseline := render()
	require.NotContains(t, baseline.Spec.Template.Annotations, autoscaling.InitialScaleAnnotationKey)
	t.Run("minimum-only PATCH cannot roll back an image accepted before the native snapshot was saved", func(t *testing.T) {
		accepted := baseline.DeepCopy()
		accepted.Spec.Template.Spec.Containers[0].Image = "example/model:v2"
		current := *existing
		var err error
		current.Workload.RawManifest, err = gojsonq.LoadJson(accepted)
		require.NoError(t, err)
		// A newer Service field still allows the known images to be compared.
		current.Workload.RawManifest["spec"].(map[string]any)["newServiceField"] = true

		stale, desired := spec, spec
		var request api.InferenceUpdateRequest
		require.NoError(t, json.Unmarshal([]byte(`{"spec":{"autoscaling":{"minReplicas":0}}}`), &request))
		require.Nil(t, mergeUpdateSpecIntoFullSpec(t.Context(), &desired, &request.Spec))
		updateContext := *wc
		updateContext.previousInferenceSpec = &stale
		merged := *workload
		merged.Spec = &desired
		update, updateErr := (&AnyWorkloadSubmitterInference{}).Update(t.Context(), &updateContext, "id", &current, &merged)
		require.Nil(t, update)
		require.NotNil(t, updateErr)
		require.EqualValues(t, http.StatusConflict, updateErr.Code)
		require.Contains(t, updateErr.Message, "image differs from the latest accepted update")
	})
	lowered := patch(`{"spec":{"autoscaling":{"minReplicas":0}}}`)
	require.True(t, requiresServing)
	require.Equal(t, "0", lowered.Spec.Template.Annotations[autoscaling.MinScaleAnnotationKey])
	require.Equal(t, "0", lowered.Spec.Template.Annotations[autoscaling.InitialScaleAnnotationKey])
	require.Equal(t, "3", lowered.Spec.Template.Annotations[autoscaling.MaxScaleAnnotationKey])
	require.Equal(t, "concurrency", lowered.Spec.Template.Annotations[autoscaling.MetricAnnotationKey])
	require.Equal(t, "<test-value>", lowered.Spec.Template.Spec.Containers[0].Env[0].Value)
	require.Nil(t, spec.Autoscaling.Get().InitialReplicas.Get())

	(*features)[checker.ZeroInitialScaleDependency] = clusterAPI.ClusterDependencyStatus{Available: false}
	retry := patch(`{"spec":{"autoscaling":{"minReplicas":0}}}`)
	require.False(t, requiresServing)
	require.Equal(t, lowered.Spec, retry.Spec)
	imageUpdate := patch(`{"spec":{"image":"example/model:v2"}}`)
	require.Equal(t, "0", imageUpdate.Spec.Template.Annotations[autoscaling.MinScaleAnnotationKey])
	require.NotContains(t, imageUpdate.Spec.Template.Annotations, autoscaling.InitialScaleAnnotationKey)
	require.Equal(t, "example/model:v2", imageUpdate.Spec.Template.Spec.Containers[0].Image)

	patch(`{"spec":{"autoscaling":{"minReplicas":1}}}`)
	disabled := patch(`{"spec":{"autoscaling":{"minReplicas":0}}}`)
	require.NotContains(t, disabled.Spec.Template.Annotations, autoscaling.InitialScaleAnnotationKey)
	(*features)[checker.ZeroInitialScaleDependency] = clusterAPI.ClusterDependencyStatus{Available: true}
	patch(`{"spec":{"autoscaling":{"minReplicas":1}}}`)
	combined := patch(`{"spec":{"autoscaling":{"minReplicas":0},"image":"example/model:v3"}}`)
	require.NotContains(t, combined.Spec.Template.Annotations, autoscaling.InitialScaleAnnotationKey)
	patch(`{"spec":{"autoscaling":{"minReplicas":1}}}`)
	credentialUpdate := patch(`{"spec":{"autoscaling":{"minReplicas":0},"environmentVariables":[{"name":"HF_TOKEN","value":"<replacement>"}]}}`)
	require.NotContains(t, credentialUpdate.Spec.Template.Annotations, autoscaling.InitialScaleAnnotationKey)
	patch(`{"spec":{"autoscaling":{"minReplicas":1}}}`)
	explicit := patch(`{"spec":{"autoscaling":{"minReplicas":0,"initialReplicas":1}}}`)
	require.NotEqual(t, "0", explicit.Spec.Template.Annotations[autoscaling.InitialScaleAnnotationKey])
}

func TestInferenceMinimumUpdateEligibility(t *testing.T) {
	latest := servingv1.TrafficTarget{LatestRevision: convutils.Ref(true), Percent: convutils.Ref(int64(100))}
	for _, tc := range []struct {
		name           string
		class          string
		traffic        []servingv1.TrafficTarget
		imageChanged   bool
		canStartAtZero bool
		want           bool
	}{
		{name: "supported", class: autoscaling.KPA, traffic: []servingv1.TrafficTarget{latest}, canStartAtZero: true, want: true},
		{name: "disabled", class: autoscaling.KPA},
		{name: "HPA", class: autoscaling.HPA, imageChanged: true, canStartAtZero: true},
		{name: "pinned traffic", class: autoscaling.KPA, imageChanged: true, traffic: []servingv1.TrafficTarget{{RevisionName: "existing", Percent: convutils.Ref(int64(100))}}, canStartAtZero: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service := &servingv1.Service{}
			service.APIVersion, service.Kind = "serving.knative.dev/v1", "Service"
			service.Spec.Traffic = tc.traffic
			service.Spec.Template.Annotations = map[string]string{autoscaling.ClassAnnotationKey: tc.class, autoscaling.MinScaleAnnotationKey: "1"}
			service.Spec.Template.Spec.Containers = []corev1.Container{{Name: "model", Image: "example/model:v2"}}
			prior, err := gojsonq.LoadJson(service)
			require.NoError(t, err)
			service.Spec.Template.Annotations[autoscaling.MinScaleAnnotationKey] = "0"
			if tc.imageChanged {
				service.Spec.Template.Spec.Containers[0].Image = "example/model:v1"
			}
			guard, updateErr := setInitialScaleOnMinimumUpdate(prior, service, tc.canStartAtZero)
			require.Nil(t, updateErr)
			require.Equal(t, tc.want, guard)
			_, derived := service.Spec.Template.Annotations[autoscaling.InitialScaleAnnotationKey]
			require.Equal(t, tc.want, derived)
		})
	}
}

func TestRequireServingInferenceUpdate(t *testing.T) {
	for _, tc := range []struct {
		name     string
		current  *int64
		serving  int64
		readErr  error
		wantCode int32
	}{
		{name: "serving current revision", current: convutils.Ref(int64(7)), serving: 7},
		{name: "zero is a valid counter", current: convutils.Ref(int64(0)), serving: 0},
		{name: "old image still serving", current: convutils.Ref(int64(8)), serving: 7, wantCode: http.StatusConflict},
		{name: "manager missing version", wantCode: http.StatusConflict},
		{name: "status read failed", current: convutils.Ref(int64(8)), readErr: errors.New("database unavailable"), wantCode: http.StatusServiceUnavailable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			err := requireServingInferenceUpdate(t.Context(), 4, "workload", tc.current, func(_ context.Context, tenant uint, id string, count int64) (bool, error) {
				called = true
				require.Equal(t, uint(4), tenant)
				require.Equal(t, "workload", id)
				return count == tc.serving, tc.readErr
			})
			require.Equal(t, tc.current != nil, called)
			if tc.wantCode == 0 {
				require.Nil(t, err)
				return
			}
			require.NotNil(t, err)
			require.Equal(t, tc.wantCode, err.Code)
			if tc.wantCode == http.StatusConflict {
				require.Contains(t, err.Message, "roll it back")
			}
		})
	}
}
