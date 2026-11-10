/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/clusterservice"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/topology"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/clients/clusters"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"

	clusterApi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/clients/org_unit"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/clients/workloads"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/database/models"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	clusterconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var _ = Describe("WorkloadsV2Impl", func() {
	var (
		impl *WorkloadsV2Impl

		ctx context.Context

		randoms *testutils.RandomValues

		sampleManifest map[string]interface{}
		csclient       *clusterservice.SimpleMock
	)

	BeforeEach(func() {
		randoms = testutils.NewRandomValues()

		csclient = clusterservice.NewSimpleMock()
		csclient.WithDefaultNodePool(randoms.ClusterId)
		clusters.SetMock(csclient)

		jsonErr := json.Unmarshal([]byte(sampleManifestSpec), &sampleManifest)
		Expect(jsonErr).To(BeNil())

		k8sUtil := NewK8sResourceUtil()
		impl = &WorkloadsV2Impl{
			k8sUtil: k8sUtil,
		}

		ctx = setContext("1", "user-a", []string{})

		authMock.SetNoPermissions().AddPermittedTenants(randoms)

		org_unit.SetMock(org_unit.NewSimpleMock().WithRandoms(randoms))

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, clusterInfoOf(randoms, testutils.DefaultClusterVersionForTesting)))

		checker.InitClusterCapabilitiesChecker("", defaultClusterCacheConfig(), "")

		// Setup default workloads client mock for existing tests
		defaultWorkloadsMock := workloads.NewSimpleMock()
		// Add default workload types for common test cases
		readyPhase := "Ready"
		jobWorkloadType := workloadsApi.WorkloadTypeConfig{
			Name:       "Job",
			Group:      "batch",
			CategoryId: "default-category",
			PriorityId: "default-priority",
			ClusterStatuses: []workloadsApi.WorkloadTypeStatus{
				{
					ClusterId: convutils.Ref(randoms.ClusterId),
					Version:   convutils.Ref("v1"),
					Phase:     &readyPhase,
				},
			},
		}
		podWorkloadType := workloadsApi.WorkloadTypeConfig{
			Name:       "Pod",
			Group:      "",
			CategoryId: "default-category",
			PriorityId: "default-priority",
			ClusterStatuses: []workloadsApi.WorkloadTypeStatus{
				{
					ClusterId: convutils.Ref(randoms.ClusterId),
					Version:   convutils.Ref("v1"),
					Phase:     &readyPhase,
				},
			},
		}
		deploymentWorkloadType := workloadsApi.WorkloadTypeConfig{
			Name:       "Deployment",
			Group:      "apps",
			CategoryId: "default-category",
			PriorityId: "default-priority",
			ClusterStatuses: []workloadsApi.WorkloadTypeStatus{
				{
					ClusterId: convutils.Ref(randoms.ClusterId),
					Version:   convutils.Ref("v1"),
					Phase:     &readyPhase,
				},
			},
		}
		pyTorchWorkloadType := workloadsApi.WorkloadTypeConfig{
			Name:       "PyTorchJob",
			Group:      "kubeflow.org",
			CategoryId: "default-category",
			PriorityId: "default-priority",
			ClusterStatuses: []workloadsApi.WorkloadTypeStatus{
				{
					ClusterId: convutils.Ref(randoms.ClusterId),
					Version:   convutils.Ref("v1"),
					Phase:     &readyPhase,
				},
			},
		}
		defaultWorkloadsMock.
			WithWorkloadType(jobWorkloadType).
			WithWorkloadType(podWorkloadType).
			WithWorkloadType(deploymentWorkloadType).
			WithWorkloadType(pyTorchWorkloadType).
			WithPrioritiesAndCategories(randoms.ClusterId, defaultPrioritiesAndCategories)
		workloads.SetMock(defaultWorkloadsMock)

		policies.SetMock(policies.SimpleMockClientFactory{})

		dao.RestartUpdateCountForTest()
	})

	AfterEach(func() {
		cleanTables()
	})

	Context("json manifest", func() {
		It("should create workload successfully with JSON manifest", func() {
			request := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "test-workload",
					ProjectId: randoms.ProjectIdStr,
				},
				Manifest: map[string]interface{}{
					"apiVersion": "batch/v1",
					"kind":       "Job",
					"metadata": map[string]interface{}{
						"name": "test-job",
					},
					"spec": map[string]interface{}{
						"template": map[string]interface{}{
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":  "test-container",
										"image": "nginx",
									},
								},
								"restartPolicy": "Never",
							},
						},
					},
				},
			}

			result, apiErrs := impl.CreateWorkloadV2(ctx, request, false)
			Expect(apiErrs).To(BeEmpty())
			Expect(result).ToNot(BeNil())

			Expect(result.Metadata.Name).To(Equal("test-workload"))
			Expect(result.Metadata.ProjectId).To(Equal(randoms.ProjectIdStr))
			Expect(result.Metadata.Id).ToNot(BeEmpty())

			Expect(convutils.Deref(result.Metadata.Gvk.Group, "")).To(Equal("batch"))
			Expect(result.Metadata.Gvk.Version).To(Equal("v1"))
			Expect(result.Metadata.Gvk.Kind).To(Equal("Job"))

			Expect(result.DesiredPhase).To(Equal(api.DESIREDPHASE_RUNNING))
			Expect(result.WorkloadManifest).ToNot(BeNil())
			Expect(result.RawManifest).ToNot(BeNil())

			workloadMetadata := result.WorkloadManifest["metadata"].(map[string]interface{})
			Expect(workloadMetadata["name"]).To(Equal("test-workload"))
			Expect(workloadMetadata["namespace"]).To(Equal(fmt.Sprintf("runai-%s", randoms.ProjectIdStr)))

			labels := workloadMetadata["labels"].(map[string]interface{})
			Expect(labels[ControlPlaneManagedLabel]).To(Equal(TrueAsString))

			annotations := workloadMetadata["annotations"].(map[string]interface{})
			Expect(annotations[WorkloadIdAnnotation]).To(Equal(result.Metadata.Id))
		})
	})

	Context("CP pod ownership stamping", func() {
		var stampedType workloadsApi.WorkloadTypeConfig
		BeforeEach(func() {
			data, err := os.ReadFile("testdata/ownership/krt-v1-kubeflow-org-pytorchjob.json")
			Expect(err).NotTo(HaveOccurred())
			var karta kartav1alpha1.Karta
			Expect(json.Unmarshal(data, &karta)).To(Succeed())
			stampedType = *createWorkloadTypeWithRi(&karta)
			stampedType.Name, stampedType.Group = "PyTorchJob", "kubeflow.org"
			stampedType.CategoryId, stampedType.PriorityId = "default-category", "default-priority"
			stampedType.ClusterStatuses = []workloadsApi.WorkloadTypeStatus{{ClusterId: convutils.Ref(randoms.ClusterId), Version: convutils.Ref("v1"), Phase: convutils.Ref("Ready")}}
			workloads.SetMock(workloads.NewSimpleMock().WithWorkloadType(stampedType).WithPrioritiesAndCategories(randoms.ClusterId, defaultPrioritiesAndCategories))
		})
		submit := func() *api.WorkloadV2 {
			result, errs := impl.CreateWorkloadV2(ctx, api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{Name: "ownership-test", ProjectId: randoms.ProjectIdStr, Category: convutils.Ref("train"), Priority: convutils.Ref("medium")},
				Manifest: newPyTorchJobManifest(nil),
			}, false)
			Expect(errs).To(BeEmpty())
			return result
		}
		checkLabels := func(result *api.WorkloadV2) {
			labels, found, err := unstructured.NestedStringMap(result.WorkloadManifest, "spec", "pytorchReplicaSpecs", "Worker", "template", "metadata", "labels")
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(labels).To(HaveKeyWithValue("run.ai/top-owner-uid", result.Metadata.Id))
			Expect(labels).To(HaveKeyWithValue("run.ai/workload-id", result.Metadata.Id))
			Expect(labels).NotTo(HaveKey("run.ai/cluster-top-owner-uid"))
			_, master, _ := unstructured.NestedFieldNoCopy(result.WorkloadManifest, "spec", "pytorchReplicaSpecs", "Master")
			Expect(master).To(BeFalse())
		}
		It("stamps the real PyTorch child template without changing raw input or creating Master", func() {
			result := submit()
			checkLabels(result)
			_, found, _ := unstructured.NestedFieldNoCopy(result.RawManifest, "spec", "pytorchReplicaSpecs", "Worker", "template", "metadata", "labels", "run.ai/top-owner-uid")
			Expect(found).To(BeFalse())
		})
		It("preserves stamping across explicit configuration replacement and original YAML edits", func() {
			created := submit()
			updated, err := impl.UpdateWorkloadV2(ctx, created.Metadata.Id, api.WorkloadV2UpdateRequest{Metadata: api.WorkloadV2MetadataUpdateParams{Configuration: &api.WorkloadV2Configuration{}}, Manifest: newPyTorchJobManifest(nil)})
			Expect(err).To(BeNil())
			checkLabels(updated)
			updated, err = impl.UpdateWorkloadV2(ctx, created.Metadata.Id, api.WorkloadV2UpdateRequest{Manifest: newPyTorchJobManifest(nil)})
			Expect(err).To(BeNil())
			checkLabels(updated)
		})
		It("does not retrofit old records when a mapping becomes available", func() {
			oldType := stampedType
			oldType.Kartas = nil
			workloads.SetMock(workloads.NewSimpleMock().WithWorkloadType(oldType).WithPrioritiesAndCategories(randoms.ClusterId, defaultPrioritiesAndCategories))
			created := submit()
			workloads.SetMock(workloads.NewSimpleMock().WithWorkloadType(stampedType).WithPrioritiesAndCategories(randoms.ClusterId, defaultPrioritiesAndCategories))
			updated, err := impl.UpdateWorkloadV2(ctx, created.Metadata.Id, api.WorkloadV2UpdateRequest{Manifest: newPyTorchJobManifest(nil)})
			Expect(err).To(BeNil())
			_, found, _ := unstructured.NestedFieldNoCopy(updated.WorkloadManifest, "spec", "pytorchReplicaSpecs", "Worker", "template", "metadata", "labels", "run.ai/top-owner-uid")
			Expect(found).To(BeFalse())
		})

		It("refuses an edit if the stamped workload loses its mapping", func() {
			created := submit()
			stampedType.Kartas = nil
			workloads.SetMock(workloads.NewSimpleMock().WithWorkloadType(stampedType).WithPrioritiesAndCategories(randoms.ClusterId, defaultPrioritiesAndCategories))
			_, err := impl.UpdateWorkloadV2(ctx, created.Metadata.Id, api.WorkloadV2UpdateRequest{Manifest: newPyTorchJobManifest(nil)})
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("no longer has a usable pod ownership mapping"))
		})

		DescribeTable("validates ownership selectors during dry run", func(conflicting bool) {
			data, err := os.ReadFile("testdata/ownership/krt-v1-apps-deployment.json")
			Expect(err).NotTo(HaveOccurred())
			var karta kartav1alpha1.Karta
			Expect(json.Unmarshal(data, &karta)).To(Succeed())
			deploymentType := *createWorkloadTypeWithRi(&karta)
			deploymentType.Name, deploymentType.Group = "Deployment", "apps"
			deploymentType.CategoryId, deploymentType.PriorityId = "default-category", "default-priority"
			deploymentType.ClusterStatuses = stampedType.ClusterStatuses
			workloads.SetMock(workloads.NewSimpleMock().WithWorkloadType(deploymentType).WithPrioritiesAndCategories(randoms.ClusterId, defaultPrioritiesAndCategories))
			var manifest map[string]interface{}
			Expect(json.Unmarshal([]byte(`{"apiVersion":"apps/v1","kind":"Deployment","metadata":{"name":"dryrun-owner"},"spec":{"replicas":0,"selector":{"matchLabels":{"app":"test"}},"template":{"metadata":{"labels":{"app":"test"}},"spec":{"containers":[{"name":"main","image":"busybox"}]}}}}`), &manifest)).To(Succeed())
			if conflicting {
				Expect(unstructured.SetNestedField(manifest, "caller-owner", "spec", "selector", "matchLabels", "run.ai/top-owner-uid")).To(Succeed())
				Expect(unstructured.SetNestedField(manifest, "caller-owner", "spec", "template", "metadata", "labels", "run.ai/top-owner-uid")).To(Succeed())
			}
			original, err := json.Marshal(manifest)
			Expect(err).NotTo(HaveOccurred())
			result, errs := impl.CreateWorkloadV2WithDryRun(ctx, api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{Name: "dryrun-owner", ProjectId: randoms.ProjectIdStr, Category: convutils.Ref("train"), Priority: convutils.Ref("medium")},
				Manifest: manifest,
			}, CreateWorkloadV2Options{DryRun: true})
			if conflicting {
				Expect(result).To(BeNil())
				Expect(errs).To(HaveLen(1))
				Expect(errs[0].Code).To(BeNumerically("==", 400))
				Expect(errs[0].Message).To(ContainSubstring("spec.selector conflicts"))
			} else {
				Expect(errs).To(BeEmpty())
				Expect(result.Metadata.Id).To(BeEmpty())
				Expect(result.WorkloadManifest).To(BeNil())
			}
			unchanged, err := json.Marshal(manifest)
			Expect(err).NotTo(HaveOccurred())
			Expect(unchanged).To(Equal(original))
		}, Entry("rejects a conflicting reserved selector", true), Entry("allows an ordinary app selector without mutating input", false))

		It("finalizes identity after update record modifiers", func() {
			created := submit()
			updated, err := impl.UpdateWorkloadV2(ctx, created.Metadata.Id,
				api.WorkloadV2UpdateRequest{Manifest: newPyTorchJobManifest(nil)},
				func(record *models.WorkloadsQueueRecord) {
					Expect(unstructured.SetNestedStringMap(record.EffectiveManifest, map[string]string{
						"run.ai/top-owner-uid": "wrong", "run.ai/workload-id": "wrong", "run.ai/cluster-top-owner-uid": "stale-cluster-uid", "user-label": "keep",
					}, "spec", "pytorchReplicaSpecs", "Worker", "template", "metadata", "labels")).To(Succeed())
				})
			Expect(err).To(BeNil())
			checkLabels(updated)
			labels, _, _ := unstructured.NestedStringMap(updated.WorkloadManifest, "spec", "pytorchReplicaSpecs", "Worker", "template", "metadata", "labels")
			Expect(labels).To(HaveKeyWithValue("user-label", "keep"))
		})
	})

	Context("Heimdall workflow-id stamping", func() {
		labelsOf := func(w *api.WorkloadV2) map[string]interface{} {
			return w.WorkloadManifest["metadata"].(map[string]interface{})["labels"].(map[string]interface{})
		}
		setOperand := func(connected, operandPresent, ready bool) {
			cluster := clusterInfoOf(randoms, testutils.DefaultClusterVersionForTesting)
			status := clusterApi.NewClusterDisplayedStatus()
			if connected {
				status.SetState(clusterApi.CLUSTERSTATE_CONNECTED)
			} else {
				status.SetState(clusterApi.CLUSTERSTATE_DISCONNECTED)
			}
			if operandPresent {
				status.SetOperands(map[string]clusterApi.ClusterReportedStatusOperandsValue{
					"advice-restart-controller": *clusterApi.NewClusterReportedStatusOperandsValue(ready),
				})
			}
			cluster.SetStatus(*status)
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().WithCluster(randoms.ClusterId, cluster))
			checker.InitClusterCapabilitiesChecker("", defaultClusterCacheConfig(), "")
		}
		submitNative := func(name string, manifest map[string]interface{}) (*api.WorkloadV2, []api.SubmissionErrorV2) {
			metadata := manifest["metadata"].(map[string]interface{})
			manifestLabels, ok := metadata["labels"].(map[string]interface{})
			if !ok {
				manifestLabels = map[string]interface{}{}
				metadata["labels"] = manifestLabels
			}
			manifestLabels[ConvertedFromNativeLabel] = TrueAsString
			return impl.CreateWorkloadV2WithDryRun(ctx, api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{Name: name, ProjectId: randoms.ProjectIdStr},
				Manifest: manifest,
			}, CreateWorkloadV2Options{TranslatedFromV1: true})
		}
		submitRaw := func(name string, manifest map[string]interface{}) (*api.WorkloadV2, []api.SubmissionErrorV2) {
			return impl.CreateWorkloadV2(ctx, api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{Name: name, ProjectId: randoms.ProjectIdStr},
				Manifest: manifest,
			}, false)
		}

		It("stamps a native PyTorchJob with workflow-id=<workloadId> when the operand is ready", func() {
			setOperand(true, true, true)
			created, errs := submitNative("pt-native", newPyTorchJobManifest(nil))
			Expect(errs).To(BeEmpty())
			Expect(labelsOf(created)).To(HaveKeyWithValue(clusterconstants.HeimdallWorkflowIdLabel, created.Metadata.Id))
		})

		It("does NOT stamp a raw v2 PyTorchJob even when the operand is ready", func() {
			setOperand(true, true, true)
			created, errs := submitRaw("pt-raw", newPyTorchJobManifest(nil))
			Expect(errs).To(BeEmpty())
			Expect(labelsOf(created)).ToNot(HaveKey(clusterconstants.HeimdallWorkflowIdLabel))
		})

		It("stamps optimistically when the operand is enabled but not ready yet", func() {
			setOperand(true, true, false)
			created, errs := submitNative("pt-notready", newPyTorchJobManifest(nil))
			Expect(errs).To(BeEmpty())
			Expect(labelsOf(created)).To(HaveKeyWithValue(clusterconstants.HeimdallWorkflowIdLabel, created.Metadata.Id))
		})

		It("does NOT stamp when the cluster is not connected", func() {
			setOperand(false, true, true)
			created, errs := submitNative("pt-disc", newPyTorchJobManifest(nil))
			Expect(errs).To(BeEmpty())
			Expect(labelsOf(created)).ToNot(HaveKey(clusterconstants.HeimdallWorkflowIdLabel))
		})

		It("does NOT stamp when the operand is absent", func() {
			setOperand(true, false, false)
			created, errs := submitNative("pt-noop", newPyTorchJobManifest(nil))
			Expect(errs).To(BeEmpty())
			Expect(labelsOf(created)).ToNot(HaveKey(clusterconstants.HeimdallWorkflowIdLabel))
		})

		It("does NOT stamp a non-PyTorchJob even when the operand is ready", func() {
			setOperand(true, true, true)
			jobManifest := map[string]interface{}{
				"apiVersion": "batch/v1", "kind": "Job",
				"metadata": map[string]interface{}{"name": "j"},
				"spec": map[string]interface{}{"template": map[string]interface{}{"spec": map[string]interface{}{
					"containers":    []interface{}{map[string]interface{}{"name": "c", "image": "nginx"}},
					"restartPolicy": "Never",
				}}},
			}
			created, errs := submitNative("job-nostamp", jobManifest)
			Expect(errs).To(BeEmpty())
			Expect(labelsOf(created)).ToNot(HaveKey(clusterconstants.HeimdallWorkflowIdLabel))
		})

		It("honors the disable-ajr-integration opt-out even when the operand is ready", func() {
			setOperand(true, true, true)
			m := newPyTorchJobManifest(nil)
			m["metadata"].(map[string]interface{})["annotations"] = map[string]interface{}{
				clusterconstants.DisableAJRAnnotation: "true",
			}
			created, errs := submitNative("pt-optout", m)
			Expect(errs).To(BeEmpty())
			Expect(labelsOf(created)).ToNot(HaveKey(clusterconstants.HeimdallWorkflowIdLabel))
		})

		It("leaves a client-supplied workflow-id label untouched when not stamping", func() {
			setOperand(true, false, false)
			m := newPyTorchJobManifest(nil)
			m["metadata"].(map[string]interface{})["labels"] = map[string]interface{}{
				clusterconstants.HeimdallWorkflowIdLabel: "client-bogus-value",
			}
			created, errs := submitNative("pt-inject", m)
			Expect(errs).To(BeEmpty())
			Expect(labelsOf(created)).To(HaveKeyWithValue(clusterconstants.HeimdallWorkflowIdLabel, "client-bogus-value"))
		})

		It("overwrites a client-supplied workflow-id label with the workloadId when stamping", func() {
			setOperand(true, true, true)
			m := newPyTorchJobManifest(nil)
			m["metadata"].(map[string]interface{})["labels"] = map[string]interface{}{
				clusterconstants.HeimdallWorkflowIdLabel: "client-bogus-value",
			}
			created, errs := submitNative("pt-inject-ready", m)
			Expect(errs).To(BeEmpty())
			Expect(labelsOf(created)).To(HaveKeyWithValue(clusterconstants.HeimdallWorkflowIdLabel, created.Metadata.Id))
		})
	})

	Context("segmented PyTorch manifest", func() {
		BeforeEach(func() {
			csclient = clusterservice.NewSimpleMock().
				WithNodePoolsWithTopology(randoms.ClusterId, "cluster-topology", "default").
				WithTopology("cluster-topology", &clusterApi.NetworkTopology{
					Name:   "cluster-topology",
					Levels: []string{"zone", "rack"},
				})
			clusters.SetMock(csclient)
		})

		It("defaults create and update in the effective manifest without changing raw input", func() {
			createManifest := newPyTorchJobManifest(map[string]interface{}{
				topology.KAISegmentSizeAnnotationKey: "2",
			})
			created, submissionErrs := impl.CreateWorkloadV2(ctx, api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "segmented-pytorch",
					ProjectId: randoms.ProjectIdStr,
					Category:  convutils.Ref("train"),
					Priority:  convutils.Ref("medium"),
				},
				Manifest: createManifest,
			}, false)
			Expect(submissionErrs).To(BeEmpty())
			Expect(created).ToNot(BeNil())

			assertPyTorchSegmentManifestOwnership(created, "2")

			updateManifest := newPyTorchJobManifest(map[string]interface{}{
				topology.KAISegmentSizeAnnotationKey: "4",
			})
			updated, updateErr := impl.UpdateWorkloadV2(ctx, created.Metadata.Id, api.WorkloadV2UpdateRequest{
				Manifest: updateManifest,
			})
			Expect(updateErr).To(BeNil())
			Expect(updated).ToNot(BeNil())
			assertPyTorchSegmentManifestOwnership(updated, "4")
		})

		It("uses a manifest-selected node pool instead of the project default", func() {
			const (
				projectTopology  = "project-topology"
				manifestTopology = "manifest-topology"
				manifestLevel    = "block"
				nodePoolLabelKey = "node-pool.example.com/type"
			)
			orgUnitMock := org_unit.NewSimpleMock().WithRandoms(randoms)
			project := orgUnitMock.ProjectInfoById[randoms.ProjectIdStr]
			project.DefaultNodePools = []string{"pool-a"}
			project.Effective.DefaultNodePools = []string{"pool-a"}
			orgUnitMock.WithProject(randoms.ProjectIdStr, project)
			org_unit.SetMock(orgUnitMock)

			setDetailedTopologyMock(
				randoms.ClusterId,
				[]clusterApi.Nodepool{
					newTopologyNodePool("pool-a", nodePoolLabelKey, "a", projectTopology),
					newTopologyNodePool("pool-b", nodePoolLabelKey, "b", manifestTopology),
				},
				map[string][]string{
					projectTopology:  {"zone", "rack"},
					manifestTopology: {"zone", manifestLevel},
				},
			)
			manifest := newPyTorchJobManifest(map[string]interface{}{
				topology.KAISegmentSizeAnnotationKey: "2",
				commonconstants.AnnotationNodepools:  "pool-b",
			})
			Expect(unstructured.SetNestedStringMap(
				manifest,
				map[string]string{clusterconstants.NodePoolNameLabel: "pool-b"},
				pyTorchReplicaTemplatePath(pyTorchReplicaTypeWorker, "metadata", "labels")...,
			)).To(Succeed())

			created, submissionErrs := impl.CreateWorkloadV2(ctx, api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "manifest-node-pool-pytorch",
					ProjectId: randoms.ProjectIdStr,
				},
				Manifest: manifest,
			}, false)

			Expect(submissionErrs).To(BeEmpty())
			Expect(created).ToNot(BeNil())
			effectiveRoot := created.WorkloadManifest["metadata"].(map[string]interface{})["annotations"].(map[string]interface{})
			Expect(effectiveRoot).To(HaveKeyWithValue(topology.KAITopologyAnnotationKey, manifestTopology))
			Expect(effectiveRoot).To(HaveKeyWithValue(topology.KAITopologyAnnotationPreferredKey, manifestLevel))
			effectiveWorker, found, err := unstructured.NestedStringMap(created.WorkloadManifest, pyTorchWorkerTemplateAnnotationsPath...)
			Expect(err).To(BeNil())
			Expect(found).To(BeTrue())
			Expect(effectiveWorker).To(HaveKeyWithValue(topology.KAISegmentTopologyPreferredAnnotationKey, manifestLevel))

			rawWorker, found, err := unstructured.NestedStringMap(created.RawManifest, pyTorchWorkerTemplateAnnotationsPath...)
			Expect(err).To(BeNil())
			Expect(found).To(BeTrue())
			Expect(rawWorker).To(HaveKeyWithValue(commonconstants.AnnotationNodepools, "pool-b"))
			Expect(rawWorker).NotTo(HaveKey(topology.KAISegmentTopologyPreferredAnnotationKey))
			Expect(created.RawManifest["metadata"].(map[string]interface{})).NotTo(HaveKey("annotations"))
		})

		It("preserves topology already resolved by the trusted native flow", func() {
			clusters.SetMock(clusterservice.NewSimpleMock().
				WithNodePoolsWithTopology(randoms.ClusterId, "project-default-topology", "default"))
			manifest := newPyTorchJobManifest(map[string]interface{}{
				topology.KAITopologyAnnotationKey:                 "native-selected-topology",
				topology.KAISegmentSizeAnnotationKey:              "2",
				topology.KAISegmentTopologyPreferredAnnotationKey: "native-rack",
			})
			manifest["metadata"].(map[string]interface{})["labels"] = map[string]interface{}{
				ConvertedFromNativeLabel: TrueAsString,
			}
			manifest["metadata"].(map[string]interface{})["annotations"] = map[string]interface{}{
				topology.KAITopologyAnnotationKey:          "native-selected-topology",
				topology.KAITopologyAnnotationPreferredKey: "native-rack",
			}

			validated, dryRunErrs := impl.CreateWorkloadV2WithDryRun(ctx, api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "native-segmented-pytorch",
					ProjectId: randoms.ProjectIdStr,
				},
				Manifest: manifest,
			}, CreateWorkloadV2Options{DryRun: true, TranslatedFromV1: true})
			Expect(dryRunErrs).To(BeEmpty())
			Expect(validated).ToNot(BeNil())

			created, submissionErrs := impl.CreateWorkloadV2WithDryRun(ctx, api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "native-segmented-pytorch",
					ProjectId: randoms.ProjectIdStr,
				},
				Manifest: manifest,
			}, CreateWorkloadV2Options{TranslatedFromV1: true})

			Expect(submissionErrs).To(BeEmpty())
			Expect(created).ToNot(BeNil())
			rootAnnotations := created.WorkloadManifest["metadata"].(map[string]interface{})["annotations"].(map[string]interface{})
			Expect(rootAnnotations).To(HaveKeyWithValue(topology.KAITopologyAnnotationKey, "native-selected-topology"))
			Expect(rootAnnotations).To(HaveKeyWithValue(topology.KAITopologyAnnotationPreferredKey, "native-rack"))
			workerAnnotations, found, err := unstructured.NestedStringMap(created.WorkloadManifest, pyTorchWorkerTemplateAnnotationsPath...)
			Expect(err).To(BeNil())
			Expect(found).To(BeTrue())
			Expect(workerAnnotations).To(HaveKeyWithValue(topology.KAITopologyAnnotationKey, "native-selected-topology"))
			Expect(workerAnnotations).To(HaveKeyWithValue(topology.KAISegmentTopologyPreferredAnnotationKey, "native-rack"))
		})

		It("rejects an incomplete trusted native manifest during dry-run", func() {
			manifest := newPyTorchJobManifest(map[string]interface{}{
				topology.KAISegmentSizeAnnotationKey: "2",
			})
			manifest["metadata"].(map[string]interface{})["labels"] = map[string]interface{}{
				ConvertedFromNativeLabel: TrueAsString,
			}

			validated, submissionErrs := impl.CreateWorkloadV2WithDryRun(ctx, api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "incomplete-native-segmented-pytorch",
					ProjectId: randoms.ProjectIdStr,
				},
				Manifest: manifest,
			}, CreateWorkloadV2Options{DryRun: true, TranslatedFromV1: true})

			Expect(validated).To(BeNil())
			Expect(submissionErrs).To(HaveLen(1))
			Expect(submissionErrs[0].Code).To(Equal(int32(400)))
			Expect(submissionErrs[0].Message).To(ContainSubstring("converted native segmented PyTorch manifest is missing"))
		})

		It("validates a segmented dry-run without changing the submitted manifest", func() {
			manifest := newPyTorchJobManifest(map[string]interface{}{
				topology.KAISegmentSizeAnnotationKey: "2",
			})

			validated, submissionErrs := impl.CreateWorkloadV2WithDryRun(ctx, api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "segmented-pytorch-dry-run",
					ProjectId: randoms.ProjectIdStr,
				},
				Manifest: manifest,
			}, CreateWorkloadV2Options{DryRun: true})

			Expect(submissionErrs).To(BeEmpty())
			Expect(validated).ToNot(BeNil())
			workerAnnotations, found, err := unstructured.NestedStringMap(manifest, pyTorchWorkerTemplateAnnotationsPath...)
			Expect(err).To(BeNil())
			Expect(found).To(BeTrue())
			Expect(workerAnnotations).NotTo(HaveKey(topology.KAISegmentTopologyPreferredAnnotationKey))
			Expect(manifest["metadata"].(map[string]interface{})).NotTo(HaveKey("annotations"))
		})

		It("rejects different project node-pool topologies during dry-run", func() {
			orgUnitMock := org_unit.NewSimpleMock().WithRandoms(randoms)
			project := orgUnitMock.ProjectInfoById[randoms.ProjectIdStr]
			project.DefaultNodePools = []string{"pool-a", "pool-b"}
			project.Effective.DefaultNodePools = []string{"pool-a", "pool-b"}
			orgUnitMock.WithProject(randoms.ProjectIdStr, project)
			org_unit.SetMock(orgUnitMock)

			clusters.SetMock(clusterservice.NewSimpleMock().
				WithNodePoolsWithTopology(randoms.ClusterId, "topology-a", "pool-a").
				WithNodePoolsWithTopology(randoms.ClusterId, "topology-b", "pool-b"))
			manifest := newPyTorchJobManifest(map[string]interface{}{
				topology.KAITopologyAnnotationKey:                 "topology-a",
				topology.KAISegmentSizeAnnotationKey:              "2",
				topology.KAISegmentTopologyPreferredAnnotationKey: "rack",
			})
			manifest["metadata"].(map[string]interface{})["annotations"] = map[string]interface{}{
				topology.KAITopologyAnnotationKey:          "topology-a",
				topology.KAITopologyAnnotationPreferredKey: "rack",
			}
			manifest["metadata"].(map[string]interface{})["labels"] = map[string]interface{}{
				ConvertedFromNativeLabel: TrueAsString,
			}

			validated, submissionErrs := impl.CreateWorkloadV2WithDryRun(ctx, api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "segmented-pytorch-dry-run",
					ProjectId: randoms.ProjectIdStr,
				},
				Manifest: manifest,
			}, CreateWorkloadV2Options{DryRun: true})

			Expect(validated).To(BeNil())
			Expect(submissionErrs).To(HaveLen(1))
			Expect(submissionErrs[0].Code).To(Equal(int32(400)))
			Expect(submissionErrs[0].Message).To(ContainSubstring("same topology"))
			Expect(submissionErrs[0].ComplianceIssues).To(HaveLen(1))
			Expect(submissionErrs[0].ComplianceIssues[0].Field).To(Equal(ErrorFieldManifest))
		})

		It("rejects segmented create on a cluster older than 2.27", func() {
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, clusterInfoOf(randoms, "2.26.9")))
			checker.InitClusterCapabilitiesChecker("", defaultClusterCacheConfig(), "")

			created, submissionErrs := impl.CreateWorkloadV2(ctx, api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "segmented-pytorch",
					ProjectId: randoms.ProjectIdStr,
				},
				Manifest: newPyTorchJobManifest(map[string]interface{}{
					topology.KAISegmentSizeAnnotationKey: "2",
				}),
			}, false)

			Expect(created).To(BeNil())
			Expect(submissionErrs).To(HaveLen(1))
			Expect(submissionErrs[0].Code).To(Equal(int32(400)))
			Expect(submissionErrs[0].Message).To(ContainSubstring("requires cluster version 2.27.0 or later"))
			Expect(submissionErrs[0].ComplianceIssues).To(HaveLen(1))
			Expect(submissionErrs[0].ComplianceIssues[0].Field).To(Equal(ErrorFieldCluster))
		})

		It("allows ordinary PyTorch create but rejects adding segmentation by update below 2.27", func() {
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, clusterInfoOf(randoms, "2.26.9")))
			checker.InitClusterCapabilitiesChecker("", defaultClusterCacheConfig(), "")

			created, submissionErrs := impl.CreateWorkloadV2(ctx, api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "ordinary-pytorch",
					ProjectId: randoms.ProjectIdStr,
				},
				Manifest: newPyTorchJobManifest(map[string]interface{}{}),
			}, false)
			Expect(submissionErrs).To(BeEmpty())
			Expect(created).ToNot(BeNil())

			updated, updateErr := impl.UpdateWorkloadV2(ctx, created.Metadata.Id, api.WorkloadV2UpdateRequest{
				Manifest: newPyTorchJobManifest(map[string]interface{}{
					topology.KAISegmentSizeAnnotationKey: "2",
				}),
			})
			Expect(updated).To(BeNil())
			Expect(updateErr).ToNot(BeNil())
			Expect(updateErr.Code).To(Equal(int32(400)))
			Expect(updateErr.Message).To(ContainSubstring("requires cluster version 2.27.0 or later"))
		})
	})

	Context("yaml manifest", func() {
		It("should create workload successfully with base64 YAML manifest", func() {
			yamlManifest := `
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec:
  containers:
  - name: test-container
    image: nginx
`
			encodedManifest := base64.StdEncoding.EncodeToString([]byte(yamlManifest))

			priority := "medium"
			category := "workspace"
			request := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "test-workspace",
					ProjectId: randoms.ProjectIdStr,
					Priority:  &priority,
					Category:  &category,
				},
				ManifestBase64Encoded: encodedManifest,
			}

			result, apiErrs := impl.CreateWorkloadV2(ctx, request, false)
			Expect(apiErrs).To(BeEmpty())
			Expect(result).ToNot(BeNil())

			Expect(result.Metadata.Priority).To(Equal(&priority))
			Expect(result.Metadata.Category).To(Equal(&category))

			workloadMetadata := result.WorkloadManifest["metadata"].(map[string]interface{})
			labels := workloadMetadata["labels"].(map[string]interface{})
			Expect(labels[PriorityLabel]).To(Equal("medium"))
			Expect(labels[CategoryLabel]).To(Equal("workspace"))
			Expect(labels[ControlPlaneManagedLabel]).To(Equal(TrueAsString))
		})
	})

	Context("validity of manifest", func() {
		It("should fail with invalid manifest", func() {
			request := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "test-workload",
					ProjectId: randoms.ProjectIdStr,
				},
				Manifest: map[string]interface{}{
					"invalid": "manifest",
				},
			}

			result, apiErrs := impl.CreateWorkloadV2(ctx, request, false)
			Expect(result).To(BeNil())
			Expect(apiErrs).ToNot(BeEmpty())
			Expect(apiErrs[0].Code).To(Equal(int32(400)))
			Expect(apiErrs[0].Message).To(ContainSubstring("missing required field: apiVersion"))
		})

		It("should fail when project is not found", func() {
			request := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "test-workload",
					ProjectId: "nonexistent-project",
				},
				Manifest: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Pod",
					"metadata": map[string]interface{}{
						"name": "test-pod",
					},
				},
			}

			result, apiErrs := impl.CreateWorkloadV2(ctx, request, false)
			Expect(result).To(BeNil())
			Expect(apiErrs).ToNot(BeEmpty())
			Expect(apiErrs[0].Code).To(Equal(int32(404)))
		})
	})

	Context("GVK workload validation and mutations", func() {
		var (
			workloadsMock *workloads.SimpleMock
			jobManifest   map[string]interface{}
			podManifest   map[string]interface{}
		)

		BeforeEach(func() {
			// Setup workloads client mock
			workloadsMock = workloads.NewSimpleMock()
			workloads.SetMock(workloadsMock)

			// Setup test manifests
			jobManifest = map[string]interface{}{
				"apiVersion": "batch/v1",
				"kind":       "Job",
				"metadata": map[string]interface{}{
					"name": "test-job",
				},
				"spec": map[string]interface{}{
					"template": map[string]interface{}{
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"name":  "test-container",
									"image": "nginx",
								},
							},
							"restartPolicy": "Never",
						},
					},
				},
			}

			podManifest = map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata": map[string]interface{}{
					"name": "test-pod",
				},
				"spec": map[string]interface{}{
					"containers": []interface{}{
						map[string]interface{}{
							"name":  "test-container",
							"image": "nginx",
						},
					},
				},
			}
		})

		Context("successful validation", func() {
			It("should validate Job workload type successfully", func() {
				// Setup mock with ready Job workload type
				readyPhase := "Ready"
				workloadType := workloadsApi.WorkloadTypeConfig{
					Name:       "Job",
					Group:      "batch",
					CategoryId: "default-category",
					PriorityId: "default-priority",
					ClusterStatuses: []workloadsApi.WorkloadTypeStatus{
						{
							ClusterId: convutils.Ref(randoms.ClusterId),
							Version:   convutils.Ref("v1"),
							Phase:     &readyPhase,
						},
					},
				}
				workloadsMock.WithWorkloadType(workloadType)

				request := api.WorkloadV2CreateRequest{
					Metadata: api.WorkloadV2MetadataCreateParams{
						Name:      "test-job-workload",
						ProjectId: randoms.ProjectIdStr,
					},
					Manifest: jobManifest,
				}

				result, apiErrs := impl.CreateWorkloadV2(ctx, request, false)
				Expect(apiErrs).To(BeEmpty())
				Expect(result).ToNot(BeNil())
				Expect(convutils.Deref(result.Metadata.Gvk.Group, "")).To(Equal("batch"))
				Expect(result.Metadata.Gvk.Kind).To(Equal("Job"))
				Expect(result.Metadata.Gvk.Version).To(Equal("v1"))
			})

			It("should validate Pod workload type successfully", func() {
				// Setup mock with ready Pod workload type
				readyPhase := "Ready"
				workloadType := workloadsApi.WorkloadTypeConfig{
					Name:       "Pod",
					Group:      "",
					CategoryId: "default-category",
					PriorityId: "default-priority",
					ClusterStatuses: []workloadsApi.WorkloadTypeStatus{
						{
							ClusterId: convutils.Ref(randoms.ClusterId),
							Version:   convutils.Ref("v1"),
							Phase:     &readyPhase,
						},
					},
				}
				workloadsMock.WithWorkloadType(workloadType)

				request := api.WorkloadV2CreateRequest{
					Metadata: api.WorkloadV2MetadataCreateParams{
						Name:      "test-pod-workload",
						ProjectId: randoms.ProjectIdStr,
					},
					Manifest: podManifest,
				}

				result, apiErrs := impl.CreateWorkloadV2(ctx, request, false)
				Expect(apiErrs).To(BeEmpty())
				Expect(result).ToNot(BeNil())
				Expect(convutils.Deref(result.Metadata.Gvk.Group, "")).To(Equal(""))
				Expect(result.Metadata.Gvk.Kind).To(Equal("Pod"))
			})
		})
	})
})

func assertPyTorchSegmentManifestOwnership(workload *api.WorkloadV2, expectedSegmentSize string) {
	rawWorkerAnnotations, found, err := unstructured.NestedStringMap(workload.RawManifest, pyTorchWorkerTemplateAnnotationsPath...)
	Expect(err).To(BeNil())
	Expect(found).To(BeTrue())
	Expect(rawWorkerAnnotations).To(HaveKeyWithValue(topology.KAISegmentSizeAnnotationKey, expectedSegmentSize))
	Expect(rawWorkerAnnotations).NotTo(HaveKey(topology.KAISegmentTopologyPreferredAnnotationKey))

	effectiveWorkerAnnotations, found, err := unstructured.NestedStringMap(workload.WorkloadManifest, pyTorchWorkerTemplateAnnotationsPath...)
	Expect(err).To(BeNil())
	Expect(found).To(BeTrue())
	Expect(effectiveWorkerAnnotations).To(HaveKeyWithValue(topology.KAISegmentSizeAnnotationKey, expectedSegmentSize))
	Expect(effectiveWorkerAnnotations).To(HaveKeyWithValue(topology.KAISegmentTopologyPreferredAnnotationKey, "rack"))

	effectiveMetadata, ok := workload.WorkloadManifest["metadata"].(map[string]interface{})
	Expect(ok).To(BeTrue())
	effectiveRootAnnotations, ok := effectiveMetadata["annotations"].(map[string]interface{})
	Expect(ok).To(BeTrue())
	Expect(effectiveRootAnnotations).To(HaveKeyWithValue(topology.KAITopologyAnnotationKey, "cluster-topology"))
	Expect(effectiveRootAnnotations).To(HaveKeyWithValue(topology.KAITopologyAnnotationPreferredKey, "rack"))

	rawMetadata, ok := workload.RawManifest["metadata"].(map[string]interface{})
	Expect(ok).To(BeTrue())
	Expect(rawMetadata).NotTo(HaveKey("annotations"))
}
