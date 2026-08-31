/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inference

import (
	"context"
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/controller"
	testingclient "github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/controller/testing/client"
	inferenceaggregator "github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/status/aggregator"
	statusupdater "github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/status/updater"
	"github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/utils/spec_checksum"
	utilsconverter "github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/pkg/utils/converter"
	runaiJobV1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	runaiv2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v2"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	workloadconfig "github.com/run-ai/runai/runai-cluster/cluster/workload-common/config"
	workloadproject "github.com/run-ai/runai/runai-cluster/cluster/workload-common/project"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	infw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	job "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	wcCommon "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/common"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/status/update"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/register"
	resourceUtils "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource/workload"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	projectName  = "test-project"
	namespace    = "runai-test-project"
	workloadName = "test-infw-workload"
)

var fakeCredential = &v1.Secret{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "my-user-credential",
		Namespace: "runai",
	},
	Data: map[string][]byte{
		// byte array with encrypted "private-value-1"
		"my-user-key": {
			168, 221, 244, 73, 135, 225, 25, 104, 145, 32, 118, 20, 196, 29, 72, 165,
			224, 40, 97, 180, 83, 157, 205, 140, 158, 186, 60, 73, 173, 193, 146, 219,
			201, 199, 65, 36, 241, 251, 52, 131, 38, 155, 0,
		},
	},
}

var _ = Describe("Reconcile inference workloads", func() {
	var (
		k8sClient    k8sclient.Client
		reconcile    *Controller
		ctx          context.Context
		project      *runaiv2.Project
		namespaceObj *v1.Namespace
	)

	scheme := runtime.NewScheme()
	register.InitSchema(scheme)

	BeforeEach(func() {
		project = resourceUtils.CreateProject(projectName)
		namespaceObj = resourceUtils.CreateNamespace(namespace, projectName)
		k8sClient = testingclient.NewFakeClientWithConflictsErrors(scheme, []k8sclient.Object{&runaiJobV1.RunaiJob{}, &infw.InferenceWorkload{}, &v2alpha1.DistributedWorkload{}}, project, namespaceObj, fakeCredential)
		ctx = workloadproject.SetProjectInContext(context.Background(), projectName)
		err := os.Setenv(workloadconfig.EnvVarClusterDomain, "https://cluster-domain.mock")
		Expect(err).ToNot(HaveOccurred())
	})

	Context("Inference Workload", func() {
		var (
			inferenceWorkload *infw.InferenceWorkload
			jobWorkload       *job.JobWorkload
			statusUpdater     update.StatusUpdater
		)

		BeforeEach(func() {
			statusUpdater = statusupdater.NewInferenceWorkloadUpdater(k8sClient, controller.TypeInferenceWorkload, utilsconverter.NewInferenceWorkloadConvertor())
			reconcile = createController(k8sClient, statusUpdater)
		})

		Context("Sanity test", func() {
			Context("External URL authentication", func() {
				It("routes restricted external URLs through their OAuth proxy", func() {
					Expect(os.Setenv(workloadconfig.EnvVarExternalAuthUrlEnabled, "true")).To(Succeed())
					DeferCleanup(os.Unsetenv, workloadconfig.EnvVarExternalAuthUrlEnabled)

					inferenceWorkload = buildBasicInferenceWorkload(workloadName, namespace)
					inferenceWorkload.Spec.ServingPortAuthorizationType = &field_types.StringField{
						Value: ptr.To(constants.ServingPortAuthorizationTypeAuthorizedUsersOrGroups),
					}
					inferenceWorkload.Spec.ServingPortAuthorizedUsers = &field_types.StringField{
						Value: ptr.To("serving-port-user@run.ai"),
					}
					inferenceWorkload.Spec.ExposedUrls = &field_types.ExposedUrlsField{
						Value: &map[string]field_types.ExposedUrlItemAndRules{
							"idx1": {
								Value: &field_types.ExposedUrlItemValue{
									ContainerPort:   ptr.To[int32](3000),
									AuthorizedUsers: ptr.To("test@run.ai"),
								},
							},
						},
					}
					Expect(k8sClient.Create(ctx, inferenceWorkload)).To(Succeed())

					jobWorkload = workload.NewJobWorkloadBuilder().WithInference().Build()
					inferenceWorkload.CopyTo(jobWorkload)
					_, err := reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					externalURLService := &v1.Service{}
					Expect(k8sClient.Get(ctx, k8sclient.ObjectKey{
						Name:      fmt.Sprintf("infw-%s-0", workloadName),
						Namespace: namespace,
					}, externalURLService)).To(Succeed())
					Expect(externalURLService.Spec.Ports[0].TargetPort.IntVal).To(Equal(constants.ExternalUrlAuthContainerDefaultPort + 1))
				})
			})

			Context("First workload reconcile", func() {
				BeforeEach(func() {
					// build inference workload
					inferenceWorkload = buildBasicInferenceWorkload(workloadName, namespace)

					// create inference workload
					err := k8sClient.Create(ctx, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					// first reconcile
					jobWorkload = workload.NewJobWorkloadBuilder().WithInference().Build()
					inferenceWorkload.CopyTo(jobWorkload)
					_, err = reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())
				})

				It("Should set first ack time", func() {
					tmp := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
					Expect(tmp.Status.FirstAcknowledgedTime).ToNot(BeNil())
				})

				It("Verify workload spec checksum annotation is set", func() {
					tmp := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
					Expect(tmp.Annotations[spec_checksum.SpecChecksumAnnotation]).ToNot(BeNil())
				})

				It("Verify Knative service creation", func() {
					tmp := getKnativeService(ctx, k8sClient, workloadName, namespace)
					Expect(tmp).ToNot(BeNil())
				})

				It("Verify Knative service Topowner", func() {
					tmp := getKnativeService(ctx, k8sClient, workloadName, namespace)
					Expect(tmp.GetOwnerReferences()[0].Name).To(Equal(workloadName))
				})
			})

			Context("Test status fields", func() {
				BeforeEach(func() {
					// build inference workload
					inferenceWorkload = buildBasicInferenceWorkload(workloadName, namespace)
				})

				It("Verify inference status fields", func() {
					inferenceWorkload.Spec.NetworkingFields.Ports = &field_types.PortsField{
						Value: &map[string]field_types.PortItemAndRules{
							"serving-port": {
								Value: &field_types.PortItemValue{
									ServiceType:   ptr.To(constants.ServiceTypeServingPort),
									ContainerPort: ptr.To(8080),
								},
							},
						},
					}
					inferenceWorkload.Spec.ExposedUrls = &field_types.ExposedUrlsField{
						Value: &map[string]field_types.ExposedUrlItemAndRules{
							"idx1": {
								Value: &field_types.ExposedUrlItemValue{
									ContainerPort: ptr.To[int32](3000),
								},
							},
						},
					}

					// adding an image pull secret, in order to test status calculation with a secret resource present
					_ = os.Setenv(constants.RunaiAdsEncryptionKeyEnvVar, wcCommon.ADSEncryptionKey)
					defer func() { _ = os.Unsetenv(constants.RunaiAdsEncryptionKeyEnvVar) }()
					inferenceWorkload.Spec.ImagePullSecrets = &field_types.StringField{
						Value: ptr.To(fmt.Sprintf("USERCREDENTIAL:%s", fakeCredential.Name)),
					}

					err := k8sClient.Create(ctx, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					// first reconcile
					inferenceWorkload.CopyTo(jobWorkload)
					_, err = reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					inferenceWorkloadAfterTheFirstReconcile := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
					kntvServiceAfterTheFirstReconcile := getKnativeService(ctx, k8sClient, workloadName, namespace)
					kntvServiceAfterTheFirstReconcile.Status.LatestCreatedRevisionName = "revision-1"
					kntvServiceAfterTheFirstReconcile.Status.LatestReadyRevisionName = "revision-1"
					url := apis.URL{
						Host: "host",
					}
					kntvServiceAfterTheFirstReconcile.Status.URL = &url
					Expect(k8sClient.Update(ctx, kntvServiceAfterTheFirstReconcile)).To(Succeed())

					inferenceWorkloadAfterTheFirstReconcile.CopyTo(jobWorkload)
					_, err = reconcile.Reconcile(ctx, jobWorkload, inferenceWorkloadAfterTheFirstReconcile)
					Expect(err).ToNot(HaveOccurred())

					inferenceWorkloadAfterTheSecondeReconcile := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
					Expect(inferenceWorkloadAfterTheSecondeReconcile.Status.LatestCreatedRevisionName).ToNot(BeNil())
					Expect(inferenceWorkloadAfterTheSecondeReconcile.Status.LatestReadyRevisionName).ToNot(BeNil())
					Expect(inferenceWorkloadAfterTheSecondeReconcile.Status.Phase).ToNot(BeNil())
					Expect(inferenceWorkloadAfterTheSecondeReconcile.Status.PhaseMessage).To(Equal(""))
					Expect(inferenceWorkloadAfterTheSecondeReconcile.Status.ExternalUrls[0]).To(Equal("https://cluster-domain.mock/test-project/test-infw-workload"))
					Expect(inferenceWorkloadAfterTheSecondeReconcile.Status.ExternalUrls[1]).To(Equal("//host"))
				})

				It("Verify OwnedResources updated", func() {
					inferenceWorkload.Spec.NetworkingFields.Ports = &field_types.PortsField{
						Value: &map[string]field_types.PortItemAndRules{
							"serving-port": {
								Value: &field_types.PortItemValue{
									ServiceType:   ptr.To(constants.ServiceTypeServingPort),
									ContainerPort: ptr.To(8080),
								},
							},
						},
					}
					inferenceWorkload.Spec.ExposedUrls = &field_types.ExposedUrlsField{
						Value: &map[string]field_types.ExposedUrlItemAndRules{
							"idx1": {
								Value: &field_types.ExposedUrlItemValue{
									ContainerPort: ptr.To[int32](3000),
								},
							},
						},
					}

					err := k8sClient.Create(ctx, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					// first reconcile
					jobWorkload = workload.NewJobWorkloadBuilder().WithInference().Build()
					inferenceWorkload.CopyTo(jobWorkload)
					_, err = reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					tmp := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
					Expect(tmp.Status.OwnedResources).ToNot(BeNil())
					Expect(len(tmp.Status.OwnedResources)).To(Equal(3))

					// check if the content of the owned resources
					found := checkOwnedResource(tmp.Status.OwnedResources, fmt.Sprintf("%s/%s, Kind=%s", constants.KnativeServiceGroup, constants.KnativeServiceVersion, constants.KnativeServiceKind))
					Expect(found).To(BeTrue(), "Expected to find a Knative Service in owned resources")

					found = checkOwnedResource(tmp.Status.OwnedResources, "Ingress")
					Expect(found).To(BeTrue(), "Expected to find a Ingress in owned resources")

					found = checkOwnedResource(tmp.Status.OwnedResources, "Service")
					Expect(found).To(BeTrue(), "Expected to find a Service in owned resources")

					// second reconcile
					// delete the knative service to verify that the reconcile will create it again and will not add it twice to the owned resources
					existKnativeService := getKnativeService(ctx, k8sClient, workloadName, namespace)
					Expect(k8sClient.Delete(ctx, existKnativeService)).To(Succeed())

					// get the inference workload after the first reconcile
					inferenceWorkloadAfterReconcile := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
					inferenceWorkloadAfterReconcile.CopyTo(jobWorkload)
					_, err = reconcile.Reconcile(ctx, jobWorkload, inferenceWorkloadAfterReconcile)
					Expect(err).ToNot(HaveOccurred())
					Expect(inferenceWorkloadAfterReconcile.Status.OwnedResources).ToNot(BeNil())
					Expect(len(inferenceWorkloadAfterReconcile.Status.OwnedResources)).To(Equal(3))

					knativeService := getKnativeService(ctx, k8sClient, workloadName, namespace)
					Expect(knativeService).ToNot(BeNil())
				})
			})

			Context("Reconcile with failure", func() {
				Context("Service with the same name as the workload service already exist", func() {
					BeforeEach(func() {
						inferenceWorkload = buildBasicInferenceWorkload(workloadName, namespace)
						err := k8sClient.Create(ctx, inferenceWorkload)
						Expect(err).ToNot(HaveOccurred())

						// add a service with the same name as the knative service
						err = k8sClient.Create(ctx, &v1.Service{
							ObjectMeta: metav1.ObjectMeta{
								Name:      fmt.Sprintf("infw-%s-0", workloadName),
								Namespace: namespace,
								Labels: map[string]string{
									"app":          "service",
									"release":      "workload-1",
									"workloadKind": "InferenceWorkload",
									"workloadName": workloadName,
								},
							},
						})
						Expect(err).ToNot(HaveOccurred())
					})

					It("reconcile success, workload failed", func() {
						// add exposed url to make sure to create service during reconcile the workload and get the error
						inferenceWorkload.Spec.ExposedUrls = &field_types.ExposedUrlsField{
							Value: &map[string]field_types.ExposedUrlItemAndRules{
								"idx1": {
									Value: &field_types.ExposedUrlItemValue{
										ContainerPort: ptr.To[int32](3000),
									},
								},
							},
						}

						inferenceWorkload.CopyTo(jobWorkload)
						_, err := reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
						Expect(err).ToNot(HaveOccurred()) // for service conflict error we dont want to re-queue the reconcile so we expect no error

						inferenceWorkloadAfterReconcile := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
						Expect(inferenceWorkloadAfterReconcile.Status.InferenceStatus.Phase).To(Equal(inferenceaggregator.Failed))
					})
				})

				Context("Service with the same name as the knative service already exist", func() {
					BeforeEach(func() {
						inferenceWorkload = buildBasicInferenceWorkload(workloadName, namespace)
						err := k8sClient.Create(ctx, inferenceWorkload)
						Expect(err).ToNot(HaveOccurred())
					})

					It("reconcile success, workload failed", func() {
						// Create an already existing Knative service (to simulate the conflict)
						alreadyExistService := &servingv1.Service{
							ObjectMeta: metav1.ObjectMeta{
								Name:      workloadName,
								Namespace: namespace,
							},
							// Add relevant spec fields
						}
						Expect(k8sClient.Create(ctx, alreadyExistService)).To(Succeed())

						inferenceWorkload.CopyTo(jobWorkload)
						_, err := reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
						Expect(err).ToNot(HaveOccurred())

						inferenceWorkloadAfterReconcile := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
						Expect(inferenceWorkloadAfterReconcile.Status.InferenceStatus.Phase).To(Equal(inferenceaggregator.Failed))
					})
				})

				Context("knative service conflict", func() {
					BeforeEach(func() {
						// set the conflict on update flag to true
						k8sClient.(*testingclient.FakeClientWithConflictsErrors).SetConflictOnUpdate(true)

						// create inference workload
						inferenceWorkload = buildBasicInferenceWorkload(workloadName, namespace)
						err := k8sClient.Create(ctx, inferenceWorkload)
						Expect(err).ToNot(HaveOccurred())
					})

					AfterEach(func() {
						// set the conflict on update flag to false
						k8sClient.(*testingclient.FakeClientWithConflictsErrors).SetConflictOnUpdate(false)
					})

					It("reconcile failed", func() {
						inferenceWorkload.CopyTo(jobWorkload)
						_, _ = reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)

						inferenceWorkloadAfterReconcile := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
						inferenceWorkloadAfterReconcile.CopyTo(jobWorkload)

						// rebuilds are now gated by the spec checksum; clear it so the
						// knative service is rebuilt and the injected conflict on update surfaces.
						delete(jobWorkload.Annotations, spec_checksum.SpecChecksumAnnotation)

						// the second reconcile to make sure that we get the error
						// (in the second reconcile we update the knative for sure)
						_, err := reconcile.Reconcile(ctx, jobWorkload, inferenceWorkloadAfterReconcile)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("Conflict"))
						inferenceWorkloadAfterReconcile = getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
						Expect(inferenceWorkloadAfterReconcile.Status.InferenceStatus.Phase).To(Equal(inferenceaggregator.Creating))

						// and try again to reconcile with other phase just to make sure that the error will be the same
						inferenceWorkloadAfterReconcile.CopyTo(jobWorkload)
						delete(jobWorkload.Annotations, spec_checksum.SpecChecksumAnnotation)
						_, err = reconcile.Reconcile(ctx, jobWorkload, inferenceWorkloadAfterReconcile)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("Conflict"))
						inferenceWorkloadAfterReconcile = getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
						Expect(inferenceWorkloadAfterReconcile.Status.InferenceStatus.Phase).To(Equal(inferenceaggregator.Creating))

						// clean SetConflictOnUpdate
						// and expect the reconcile to success and the phase to be the same
						k8sClient.(*testingclient.FakeClientWithConflictsErrors).SetConflictOnUpdate(false)
						_, _ = reconcile.Reconcile(ctx, jobWorkload, inferenceWorkloadAfterReconcile)
						_, err = reconcile.Reconcile(ctx, jobWorkload, inferenceWorkloadAfterReconcile)
						Expect(err).ToNot(HaveOccurred())
						inferenceWorkloadAfterReconcile = getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
						Expect(inferenceWorkloadAfterReconcile.Status.InferenceStatus.Phase).To(Equal(inferenceaggregator.Creating))
					})
				})
			})

			Context("Should rebuild knative service and update checksum", func() {
				BeforeEach(func() {
					// build inference workload
					inferenceWorkload = buildBasicInferenceWorkload(workloadName, namespace)
				})

				It("New workload without annotations", func() {
					// with no annotations
					err := k8sClient.Create(ctx, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					inferenceWorkload.CopyTo(jobWorkload)
					_, err = reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					// get the reconciled inference workload
					reconciledInferenceWorkload := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
					// verify the checksum annotation
					expectedCheckSum, err := spec_checksum.CalculateSpecChecksum(inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())
					Expect(reconciledInferenceWorkload.Annotations[spec_checksum.SpecChecksumAnnotation]).To(Equal(expectedCheckSum))
				})

				It("New workload with checksum annotations (not migrated)", func() {
					// with spec checksum annotations
					inferenceWorkload.Annotations = map[string]string{
						spec_checksum.SpecChecksumAnnotation: "checksum",
					}

					// create the workload
					err := k8sClient.Create(ctx, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					// reconcile the workload
					inferenceWorkload.CopyTo(jobWorkload)
					_, err = reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					// get the reconciled inference workload
					reconciledInferenceWorkload := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
					// verify the checksum annotation
					expectedCheckSum, err := spec_checksum.CalculateSpecChecksum(inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())
					Expect(reconciledInferenceWorkload.Annotations[spec_checksum.SpecChecksumAnnotation]).To(Equal(expectedCheckSum))
				})

				It("New workload with checksum annotations (not migrated) with same/updated checksum", func() {
					expectedCheckSum, err := spec_checksum.CalculateSpecChecksum(inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					// with spec checksum annotations
					inferenceWorkload.Annotations = map[string]string{
						spec_checksum.SpecChecksumMigratedAnnotation: "true",
						spec_checksum.SpecChecksumAnnotation:         expectedCheckSum,
					}

					// create the workload
					err = k8sClient.Create(ctx, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					// reconcile the workload
					inferenceWorkload.CopyTo(jobWorkload)
					_, err = reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					// get the reconciled inference workload
					reconciledInferenceWorkload := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
					// verify the checksum annotation
					Expect(reconciledInferenceWorkload.Annotations[spec_checksum.SpecChecksumAnnotation]).To(Equal(expectedCheckSum))
				})

				It("New workload with checksum annotations (not migrated) with different checksum", func() {
					// with spec checksum annotations
					inferenceWorkload.Annotations = map[string]string{
						spec_checksum.SpecChecksumMigratedAnnotation: "true",
						spec_checksum.SpecChecksumAnnotation:         "checksum",
					}

					// create the workload
					err := k8sClient.Create(ctx, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					// verify there is knative service yet
					knativeService := getKnativeService(ctx, k8sClient, workloadName, namespace)
					Expect(knativeService.Name).To(Equal(""))

					// reconcile the workload
					inferenceWorkload.CopyTo(jobWorkload)
					_, err = reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					// get the reconciled inference workload
					reconciledInferenceWorkload := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)

					// verify the checksum annotation
					expectedCheckSum, err := spec_checksum.CalculateSpecChecksum(inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())
					Expect(reconciledInferenceWorkload.Annotations[spec_checksum.SpecChecksumAnnotation]).To(Equal(expectedCheckSum))
					knativeService = getKnativeService(ctx, k8sClient, workloadName, namespace)

					// verify the knative service was created/rebuilt
					Expect(knativeService).ToNot(BeNil())
					Expect(knativeService.GetObjectMeta().GetResourceVersion()).ToNot(BeEmpty())
				})
			})

			// before the fix, shouldRebuildKnativeService returned true
			// unconditionally for every non-migrated workload, so the controller rebuilt and
			// re-Updated the Knative service on every reconcile even when the spec was unchanged.
			// On a version upgrade the builder output drifts, so that unconditional rebuild
			// pushes a new revision that can exceed the Knative progress deadline and take the
			// service offline. These specs pin the checksum-gated behavior.
			Context("Unnecessary rebuild for unchanged workload", func() {
				It("should not rebuild an unchanged non-migrated workload on a subsequent reconcile", func() {
					inferenceWorkload = buildBasicInferenceWorkload(workloadName, namespace)
					Expect(k8sClient.Create(ctx, inferenceWorkload)).To(Succeed())

					// first reconcile: builds the Knative service and records the spec checksum
					jobWorkload = workload.NewJobWorkloadBuilder().WithInference().Build()
					inferenceWorkload.CopyTo(jobWorkload)
					_, err := reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					reconciled := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
					Expect(reconciled.Annotations[spec_checksum.SpecChecksumAnnotation]).ToNot(BeEmpty())

					// nothing about the workload changed since the checksum was recorded
					secondJob := workload.NewJobWorkloadBuilder().WithInference().Build()
					reconciled.CopyTo(secondJob)

					shouldRebuild, _, err := reconcile.shouldRebuildKnativeService(ctx, secondJob)
					Expect(err).ToNot(HaveOccurred())
					Expect(shouldRebuild).To(BeFalse(),
						"unchanged workload must not be rebuilt; the buggy controller rebuilds every reconcile")
				})

				It("should rebuild an unchanged workload when its knative service was deleted", func() {
					inferenceWorkload = buildBasicInferenceWorkload(workloadName, namespace)
					Expect(k8sClient.Create(ctx, inferenceWorkload)).To(Succeed())

					// first reconcile: builds the Knative service and records the spec checksum
					jobWorkload = workload.NewJobWorkloadBuilder().WithInference().Build()
					inferenceWorkload.CopyTo(jobWorkload)
					_, err := reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					reconciled := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
					Expect(reconciled.Annotations[spec_checksum.SpecChecksumAnnotation]).ToNot(BeEmpty())

					// the spec is unchanged, but the knative service was deleted out from under the controller
					knativeService := getKnativeService(ctx, k8sClient, workloadName, namespace)
					Expect(knativeService.Name).ToNot(BeEmpty())
					Expect(k8sClient.Delete(ctx, knativeService)).To(Succeed())

					unchangedJob := workload.NewJobWorkloadBuilder().WithInference().Build()
					reconciled.CopyTo(unchangedJob)

					shouldRebuild, _, err := reconcile.shouldRebuildKnativeService(ctx, unchangedJob)
					Expect(err).ToNot(HaveOccurred())
					Expect(shouldRebuild).To(BeTrue(),
						"a deleted knative service must be recreated even when the workload spec is unchanged")
				})

				It("should rebuild when the workload spec changed", func() {
					inferenceWorkload = buildBasicInferenceWorkload(workloadName, namespace)
					Expect(k8sClient.Create(ctx, inferenceWorkload)).To(Succeed())

					jobWorkload = workload.NewJobWorkloadBuilder().WithInference().Build()
					inferenceWorkload.CopyTo(jobWorkload)
					_, err := reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					changed := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
					Expect(changed.Annotations[spec_checksum.SpecChecksumAnnotation]).ToNot(BeEmpty())

					// mutate the spec so its checksum no longer matches the recorded one
					changed.Spec.Image = field_types.NewOptionalStringField("nginx:latest", "")

					changedJob := workload.NewJobWorkloadBuilder().WithInference().Build()
					changed.CopyTo(changedJob)

					shouldRebuild, _, err := reconcile.shouldRebuildKnativeService(ctx, changedJob)
					Expect(err).ToNot(HaveOccurred())
					Expect(shouldRebuild).To(BeTrue(), "a changed workload spec must trigger a rebuild")
				})
			})

			Context("Test all expected phases", func() {
				BeforeEach(func() {
					// create inference workload
					inferenceWorkload = buildBasicInferenceWorkload(workloadName, namespace)
					err := k8sClient.Create(ctx, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())
				})

				It("Creating phase", func() {
					// first reconcile
					inferenceWorkload.CopyTo(jobWorkload)
					_, err := reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					tmp := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
					Expect(tmp.Status.InferenceStatus.Phase).To(Equal(inferenceaggregator.Creating))

					// For now, we don't have any messages in the Creating phase
					Expect(tmp.Status.InferenceStatus.PhaseMessage).To(Equal(""))
				})

				It("Pending phase", func() {
					// first reconcile
					inferenceWorkload.CopyTo(jobWorkload)
					_, err := reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					// set the knative lCRevision to not "" (latestCreatedRevisionName)
					knativeService := getKnativeService(ctx, k8sClient, workloadName, namespace)
					knativeService.Status.LatestCreatedRevisionName = "not-empty-0001"
					Expect(k8sClient.Update(ctx, knativeService)).To(Succeed())

					// second reconcile to get to pending phase
					_, err = reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					tmp := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
					Expect(tmp.Status.InferenceStatus.Phase).To(Equal(inferenceaggregator.Pending))
					Expect(tmp.Status.InferenceStatus.LatestReadyRevisionName).To(BeNil())
					Expect(tmp.Status.InferenceStatus.LatestCreatedRevisionName).To(Not(BeNil()))

					// For now, we don't have any messages in the Pending phase
					Expect(tmp.Status.InferenceStatus.PhaseMessage).To(Equal(""))
				})

				It("Pending phase while at least one pod in Pending ", func() {
					// first reconcile
					inferenceWorkload.CopyTo(jobWorkload)
					_, err := reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					// set the knative lCRevision to not "" (latestCreatedRevisionName)
					knativeService := getKnativeService(ctx, k8sClient, workloadName, namespace)
					revisionName := fmt.Sprintf("%s-00001", workloadName)
					knativeService.Status.LatestCreatedRevisionName = revisionName
					Expect(k8sClient.Update(ctx, knativeService)).To(Succeed())

					// create pod with the knative service revision
					podWithContainerCreatingPhase := buildPodWithRevisionLabel("pod-1", namespace, revisionName, v1.PodPending)
					podWithContainerCreatingPhase.Annotations = map[string]string{
						constants.PodStatusAnnotation: "ContainerCreating",
					}
					Expect(k8sClient.Create(ctx, podWithContainerCreatingPhase)).To(Succeed())

					// create pod with the knative service revision
					podWithPendingPhase := buildPodWithRevisionLabel("pod-2", namespace, revisionName, v1.PodPending)
					Expect(k8sClient.Create(ctx, podWithPendingPhase)).To(Succeed())

					// second reconcile to get to the Initializing phase
					_, err = reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					tmp := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
					Expect(tmp.Status.InferenceStatus.Phase).To(Equal(inferenceaggregator.Pending))
				})

				It("Initializing phase (based on pod phase)", func() {
					// first reconcile
					inferenceWorkload.CopyTo(jobWorkload)
					_, err := reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					// set the knative lCRevision to not "" (latestCreatedRevisionName)
					knativeService := getKnativeService(ctx, k8sClient, workloadName, namespace)
					revisionName := fmt.Sprintf("%s-00001", workloadName)
					knativeService.Status.LatestCreatedRevisionName = revisionName
					Expect(k8sClient.Update(ctx, knativeService)).To(Succeed())

					// create pod with the knative service revision
					pod := buildPodWithRevisionLabel("pod-1", namespace, revisionName, v1.PodRunning)
					Expect(k8sClient.Create(ctx, pod)).To(Succeed())

					// second reconcile to get to Initializing phase
					_, err = reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					tmp := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
					Expect(tmp.Status.InferenceStatus.Phase).To(Equal(inferenceaggregator.Initializing))
				})

				It("Initializing phase (based on pod annotations 'ContainerCreating' and pod phase 'PodPending')", func() {
					// first reconcile
					inferenceWorkload.CopyTo(jobWorkload)
					_, err := reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					// set the knative lCRevision to not "" (latestCreatedRevisionName)
					knativeService := getKnativeService(ctx, k8sClient, workloadName, namespace)
					revisionName := fmt.Sprintf("%s-00001", workloadName)
					knativeService.Status.LatestCreatedRevisionName = revisionName
					Expect(k8sClient.Update(ctx, knativeService)).To(Succeed())

					// create pod with the knative service revision
					pod := buildPodWithRevisionLabel("pod-1", namespace, revisionName, v1.PodPending)
					pod.Annotations = map[string]string{
						constants.PodStatusAnnotation: "ContainerCreating",
					}
					Expect(k8sClient.Create(ctx, pod)).To(Succeed())

					// second reconcile to get to the Initializing phase
					_, err = reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					tmp := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
					Expect(tmp.Status.InferenceStatus.Phase).To(Equal(inferenceaggregator.Initializing))
				})

				It("Ready phase", func() {
					// first reconcile
					inferenceWorkload.CopyTo(jobWorkload)
					_, err := reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					// set the knative lCRevision and lRRevision to be the same
					knativeService := getKnativeService(ctx, k8sClient, workloadName, namespace)
					revisionName := fmt.Sprintf("%s-00001", workloadName)
					knativeService.Status.LatestCreatedRevisionName = revisionName
					knativeService.Status.LatestReadyRevisionName = revisionName
					Expect(k8sClient.Update(ctx, knativeService)).To(Succeed())

					// second reconcile to get to Ready phase
					_, err = reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					tmp := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
					Expect(tmp.Status.InferenceStatus.Phase).To(Equal(inferenceaggregator.Ready))

					// For now, we don't have any messages in the Ready phase (when no update is in progress).
					Expect(tmp.Status.InferenceStatus.PhaseMessage).To(Equal(""))
				})

				It("Ready(Progressing) phase", func() {
					// first reconcile
					inferenceWorkload.CopyTo(jobWorkload)
					_, err := reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					// set the knative lCRevision and lRRevision to be the same
					knativeService := getKnativeService(ctx, k8sClient, workloadName, namespace)
					latestCreateRevisionName := fmt.Sprintf("%s-00001", workloadName)
					latestReadyRevisionName := fmt.Sprintf("%s-00002", workloadName)
					knativeService.Status.LatestCreatedRevisionName = latestCreateRevisionName
					knativeService.Status.LatestReadyRevisionName = latestReadyRevisionName
					Expect(k8sClient.Update(ctx, knativeService)).To(Succeed())

					// second reconcile to get to Ready(Progressing) phase
					_, err = reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					tmp := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
					Expect(tmp.Status.InferenceStatus.Phase).To(Equal(inferenceaggregator.Ready))
				})

				It("Ready(Progressing) phase while having build error", func() {
					alreadyExistService := &servingv1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Name:      workloadName,
							Namespace: namespace,
						},
					}
					Expect(k8sClient.Create(ctx, alreadyExistService)).To(Succeed())

					// reconcile
					knativeService := getKnativeService(ctx, k8sClient, workloadName, namespace)
					latestCreateRevisionName := fmt.Sprintf("%s-00001", workloadName)
					latestReadyRevisionName := fmt.Sprintf("%s-00002", workloadName)
					knativeService.Status.LatestCreatedRevisionName = latestCreateRevisionName
					knativeService.Status.LatestReadyRevisionName = latestReadyRevisionName
					Expect(k8sClient.Update(ctx, knativeService)).To(Succeed())

					inferenceWorkload.CopyTo(jobWorkload)
					_, err := reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					tmp := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
					Expect(tmp.Status.InferenceStatus.Phase).To(Equal(inferenceaggregator.Ready))
				})

				It("Failed phase caused by conditionFailedFound", func() {
					// first reconcile
					inferenceWorkload.CopyTo(jobWorkload)
					_, err := reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					// set the knative lCRevision and lRRevision to be the same
					knativeService := getKnativeService(ctx, k8sClient, workloadName, namespace)
					latestCreateRevisionName := fmt.Sprintf("%s-00001", workloadName)
					knativeService.Status.LatestCreatedRevisionName = latestCreateRevisionName
					knativeService.Status.Conditions = duckv1.Conditions{
						{
							Type:   servingv1.ServiceConditionConfigurationsReady,
							Status: v1.ConditionFalse,
							Reason: "RevisionFailed",
						},
					}
					Expect(k8sClient.Update(ctx, knativeService)).To(Succeed())

					// second reconcile to get to Failed phase
					_, err = reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					tmp := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
					Expect(tmp.Status.InferenceStatus.Phase).To(Equal(inferenceaggregator.Failed))
				})

				It("Failed phase (instead of Creating) caused by isBuildError", func() {
					// already exist service to simulate the isBuildError on our service creation
					alreadyExistService := &servingv1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Name:      workloadName,
							Namespace: namespace,
						},
						// Add relevant spec fields
					}
					Expect(k8sClient.Create(ctx, alreadyExistService)).To(Succeed())

					inferenceWorkload.CopyTo(jobWorkload)
					_, err := reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					tmp := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
					Expect(tmp.Status.InferenceStatus.Phase).To(Equal(inferenceaggregator.Failed))
				})

				It("Degraded phase (instead of Ready) caused by isBuildError", func() {
					// first reconcile
					inferenceWorkload.CopyTo(jobWorkload)
					_, err := reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					// set the knative lCRevision and lRRevision to be the same
					knativeService := getKnativeService(ctx, k8sClient, workloadName, namespace)
					revisionName := fmt.Sprintf("%s-00001", workloadName)
					knativeService.Status.LatestCreatedRevisionName = revisionName
					knativeService.Status.LatestReadyRevisionName = revisionName
					Expect(k8sClient.Update(ctx, knativeService)).To(Succeed())

					// rebuilds are now gated by the spec checksum; clear the recorded
					// checksum so this reconcile still attempts the knative update whose failure
					// drives the Degraded phase.
					delete(jobWorkload.Annotations, spec_checksum.SpecChecksumAnnotation)

					k8sClient.(*testingclient.FakeClientWithConflictsErrors).SetUnavailableOnUpdate(true)
					// second reconcile to get to build Error
					_, err = reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					k8sClient.(*testingclient.FakeClientWithConflictsErrors).SetUnavailableOnUpdate(false)

					tmp := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
					Expect(tmp.Status.InferenceStatus.Phase).To(Equal(inferenceaggregator.Degraded))
				})

				It("Failed phase (instead of Pending/Initializing) caused by isBuildError", func() {
					// first reconcile
					alreadyExistService := &servingv1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Name:      workloadName,
							Namespace: namespace,
						},
						// Add relevant spec fields
					}
					Expect(k8sClient.Create(ctx, alreadyExistService)).To(Succeed())
					inferenceWorkload.CopyTo(jobWorkload)
					_, err := reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					// set the knative lCRevision to not "" (latestCreatedRevisionName)
					knativeService := getKnativeService(ctx, k8sClient, workloadName, namespace)
					knativeService.Status.LatestCreatedRevisionName = "not-empty-0001"
					Expect(k8sClient.Update(ctx, knativeService)).To(Succeed())

					// second reconcile to get to failed phase
					_, err = reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					tmp := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
					Expect(tmp.Status.InferenceStatus.Phase).To(Equal(inferenceaggregator.Failed))
				})

				It("Deleting phase", func() {
					now := metav1.NewTime(time.Now())
					inferenceWorkload.DeletionTimestamp = &now

					inferenceWorkload.CopyTo(jobWorkload)
					_, err := reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					inferenceWorkloadAfterReconcile := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
					Expect(inferenceWorkloadAfterReconcile.Status.InferenceStatus.Phase).To(Equal(inferenceaggregator.Deleting))

					// we don't have any messages in the Ready Deleting
					Expect(inferenceWorkloadAfterReconcile.Status.InferenceStatus.PhaseMessage).To(Equal(""))
				})

				It("Deleting phase with error should ignore error and just delete ", func() {
					// first reconcile
					inferenceWorkload.CopyTo(jobWorkload)
					_, err := reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					// set RevisionFailed to create the error
					knativeService := getKnativeService(ctx, k8sClient, workloadName, namespace)
					latestCreateRevisionName := fmt.Sprintf("%s-00001", workloadName)
					knativeService.Status.LatestCreatedRevisionName = latestCreateRevisionName
					knativeService.Status.Conditions = duckv1.Conditions{
						{
							Type:   servingv1.ServiceConditionConfigurationsReady,
							Status: v1.ConditionFalse,
							Reason: "RevisionFailed",
						},
					}
					Expect(k8sClient.Update(ctx, knativeService)).To(Succeed())

					// set as deleting
					now := metav1.NewTime(time.Now())
					inferenceWorkload.DeletionTimestamp = &now

					// second reconcile to get to deleting phase (ignore the error)
					inferenceWorkload.CopyTo(jobWorkload)
					_, err = reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					inferenceWorkloadAfterReconcile := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
					Expect(inferenceWorkloadAfterReconcile.Status.InferenceStatus.Phase).To(Equal(inferenceaggregator.Deleting))
				})
			})

			Context("Test all expected phase message", func() {
				BeforeEach(func() {
					// create inference workload
					inferenceWorkload = buildBasicInferenceWorkload(workloadName, namespace)
					err := k8sClient.Create(ctx, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					jobWorkload = workload.NewJobWorkloadBuilder().WithInference().Build()
				})

				It("Initializing phase - Workload is pulling container images", func() {
					// first reconcile
					inferenceWorkload.CopyTo(jobWorkload)
					_, err := reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					// set the knative lCRevision to not "" (latestCreatedRevisionName) e
					latestCreateRevisionName := fmt.Sprintf("%s-00001", workloadName)
					setKnativeService(ctx, k8sClient, workloadName, namespace, nil, &latestCreateRevisionName, nil)

					// create pod with the knative service revision
					pod := buildPodWithRevisionLabel("pod-1", namespace, latestCreateRevisionName, v1.PodUnknown)
					pod.Annotations = map[string]string{
						constants.PodStatusAnnotation: "ContainerCreating",
					}
					Expect(k8sClient.Create(ctx, pod)).To(Succeed())

					// second reconcile to get to Initializing phase
					_, err = reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					tmp := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
					Expect(tmp.Status.InferenceStatus.Phase).To(Equal(inferenceaggregator.Initializing))
					Expect(tmp.Status.InferenceStatus.PhaseMessage).To(Equal("Workload is pulling container images."))
				})

				It("Initializing phase - failed to pull container image, retrying", func() {
					// first reconcile
					inferenceWorkload.CopyTo(jobWorkload)
					_, err := reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					// set Knative latestCreatedRevisionName
					latestCreatedRevisionName := fmt.Sprintf("%s-00001", workloadName)
					setKnativeService(ctx, k8sClient, workloadName, namespace, nil, &latestCreatedRevisionName, nil)

					// create pod with the knative service revision and ImagePullBackOff status
					pod := buildPodWithRevisionLabel("pod-errimagepull", namespace, latestCreatedRevisionName, v1.PodPending)
					pod.Status.ContainerStatuses = []v1.ContainerStatus{
						{
							State: v1.ContainerState{
								Waiting: &v1.ContainerStateWaiting{
									Reason: "ImagePullBackOff",
								},
							},
						},
					}
					Expect(k8sClient.Create(ctx, pod)).To(Succeed())

					// second reconcile
					_, err = reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					tmp := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
					Expect(tmp.Status.InferenceStatus.Phase).To(Equal(inferenceaggregator.Initializing))
					Expect(tmp.Status.InferenceStatus.PhaseMessage).To(Equal("Failed to pull container image. Retrying..."))
				})

				It("Initializing phase - Container images pulled successfully. Workload is initializing", func() {
					// first reconcile
					inferenceWorkload.CopyTo(jobWorkload)
					_, err := reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					// set the knative lCRevision to not "" (latestCreatedRevisionName)
					latestCreateRevisionName := fmt.Sprintf("%s-00001", workloadName)
					setKnativeService(ctx, k8sClient, workloadName, namespace, nil, &latestCreateRevisionName, nil)

					// create pod with the knative service revision
					pod := buildPodWithRevisionLabel("pod-1", namespace, latestCreateRevisionName, v1.PodRunning)
					Expect(k8sClient.Create(ctx, pod)).To(Succeed())

					// second reconcile to get to Initializing phase
					_, err = reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					tmp := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
					Expect(tmp.Status.InferenceStatus.Phase).To(Equal(inferenceaggregator.Initializing))
					Expect(tmp.Status.InferenceStatus.PhaseMessage).To(Equal("Container images pulled successfully. Workload is initializing."))
				})

				It("Failed phase - Failed to pull container image", func() {
					// first reconcile
					inferenceWorkload.CopyTo(jobWorkload)
					_, err := reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					// set the knative
					latestCreateRevisionName := fmt.Sprintf("%s-00001", workloadName)
					message := fmt.Sprintf("Revision \"%s\"  failed with message: Unable to fetch image \"no-image\": failed to resolve image to ...).; ", latestCreateRevisionName)
					conditions := buildKnativeServiceCondition(servingv1.ServiceConditionConfigurationsReady, v1.ConditionFalse, "RevisionFailed", message)
					setKnativeService(ctx, k8sClient, workloadName, namespace, nil, &latestCreateRevisionName, &conditions)

					// create pod with the knative service revision
					pod := buildPodWithRevisionLabel("pod-1", namespace, latestCreateRevisionName, v1.PodUnknown)
					pod.Annotations = map[string]string{
						constants.PodStatusAnnotation: "ContainerCreating",
					}
					Expect(k8sClient.Create(ctx, pod)).To(Succeed())

					// second reconcile to get to Ready(Progressing) phase
					_, err = reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					tmp := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
					Expect(tmp.Status.InferenceStatus.Phase).To(Equal(inferenceaggregator.Failed))
					Expect(tmp.Status.InferenceStatus.PhaseMessage).To(Equal("Failed to pull container image."))
				})

				It("Failed phase - Timed out waiting for the workload to become ready", func() {
					// first reconcile
					inferenceWorkload.CopyTo(jobWorkload)
					_, err := reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					// set the knative
					latestCreateRevisionName := fmt.Sprintf("%s-00001", workloadName)
					conditions := buildKnativeServiceCondition(servingv1.ServiceConditionConfigurationsReady, v1.ConditionFalse, "RevisionFailed", "")
					setKnativeService(ctx, k8sClient, workloadName, namespace, nil, &latestCreateRevisionName, &conditions)

					// create pod with the knative service revision
					pod := buildPodWithRevisionLabel("pod-1", namespace, latestCreateRevisionName, v1.PodRunning)
					Expect(k8sClient.Create(ctx, pod)).To(Succeed())

					// second reconcile to get to Failed phase
					_, err = reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					tmp := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
					Expect(tmp.Status.InferenceStatus.Phase).To(Equal(inferenceaggregator.Failed))
					Expect(tmp.Status.InferenceStatus.PhaseMessage).To(Equal("Timed out waiting for the workload to become ready."))
				})

				It("Failed phase - already exists", func() {
					// already exist service to simulate the isBuildError on our service creation
					alreadyExistService := &servingv1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Name:      workloadName,
							Namespace: namespace,
						},
						// Add relevant spec fields
					}
					Expect(k8sClient.Create(ctx, alreadyExistService)).To(Succeed())

					inferenceWorkload.CopyTo(jobWorkload)
					_, err := reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					tmp := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
					Expect(tmp.Status.InferenceStatus.Phase).To(Equal(inferenceaggregator.Failed))
					Expect(tmp.Status.InferenceStatus.PhaseMessage).To(Equal("Failed to create workload: [Service:test-infw-workload]  \"test-infw-workload\" already exists"))
				})

				// "Unavailable" is the error message, its just and example for any error that issuer generate will show in the phase message
				It("Degraded phase - Unavailable", func() {
					// first reconcile
					inferenceWorkload.CopyTo(jobWorkload)
					_, err := reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					// set the knative lCRevision and lRRevision to be the same
					latestCreateRevisionName := fmt.Sprintf("%s-00001", workloadName)
					conditions := buildKnativeServiceCondition(servingv1.ServiceConditionConfigurationsReady, v1.ConditionFalse, "RevisionFailed", "")
					setKnativeService(ctx, k8sClient, workloadName, namespace, &latestCreateRevisionName, &latestCreateRevisionName, &conditions)

					// create pod with the knative service revision
					pod := buildPodWithRevisionLabel("pod-1", namespace, latestCreateRevisionName, v1.PodRunning)
					Expect(k8sClient.Create(ctx, pod)).To(Succeed())

					// rebuilds are now gated by the spec checksum; clear the recorded
					// checksum so this reconcile still attempts the knative update whose failure
					// drives the Degraded phase.
					delete(jobWorkload.Annotations, spec_checksum.SpecChecksumAnnotation)

					k8sClient.(*testingclient.FakeClientWithConflictsErrors).SetUnavailableOnUpdate(true)
					// second reconcile to get to build Error
					_, err = reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					k8sClient.(*testingclient.FakeClientWithConflictsErrors).SetUnavailableOnUpdate(false)

					tmp := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
					Expect(tmp.Status.InferenceStatus.Phase).To(Equal(inferenceaggregator.Degraded))
					Expect(tmp.Status.InferenceStatus.PhaseMessage).To(Equal("Workload is ready. Update failed: [Service:test-infw-workload] Unavailable"))
				})

				It("Ready phase - waiting for resources", func() {
					// first reconcile
					inferenceWorkload.CopyTo(jobWorkload)
					_, err := reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					// set the knative
					latestCreateRevisionName := fmt.Sprintf("%s-00002", workloadName)
					latestReadyRevisionName := fmt.Sprintf("%s-00001", workloadName)
					setKnativeService(ctx, k8sClient, workloadName, namespace, &latestReadyRevisionName, &latestCreateRevisionName, nil)

					// create pod with the knative service revision
					pod := buildPodWithRevisionLabel("pod-1", namespace, latestCreateRevisionName, v1.PodPending)
					Expect(k8sClient.Create(ctx, pod)).To(Succeed())

					// second reconcile to get to Ready(Progressing) phase
					_, err = reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					tmp := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
					Expect(tmp.Status.InferenceStatus.Phase).To(Equal(inferenceaggregator.Ready))
					Expect(tmp.Status.InferenceStatus.PhaseMessage).To(Equal("Workload is ready. Update in progress: waiting for resources."))
				})

				It("Ready phase - pulling container images", func() {
					// first reconcile
					inferenceWorkload.CopyTo(jobWorkload)
					_, err := reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					// set the knative
					latestCreateRevisionName := fmt.Sprintf("%s-00002", workloadName)
					latestReadyRevisionName := fmt.Sprintf("%s-00001", workloadName)
					setKnativeService(ctx, k8sClient, workloadName, namespace, &latestReadyRevisionName, &latestCreateRevisionName, nil)

					// create pod with the knative service revision
					pod := buildPodWithRevisionLabel("pod-1", namespace, latestCreateRevisionName, v1.PodUnknown)
					pod.Annotations = map[string]string{
						constants.PodStatusAnnotation: "ContainerCreating",
					}
					Expect(k8sClient.Create(ctx, pod)).To(Succeed())

					// second reconcile to get to Ready(Progressing) phase
					_, err = reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					tmp := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
					Expect(tmp.Status.InferenceStatus.Phase).To(Equal(inferenceaggregator.Ready))
					Expect(tmp.Status.InferenceStatus.PhaseMessage).To(Equal("Workload is ready. Update in progress: pulling container images."))
				})

				It("Ready phase - container images pulled successfully, new revision is initializing", func() {
					// first reconcile
					inferenceWorkload.CopyTo(jobWorkload)
					_, err := reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					// set the knative
					latestCreateRevisionName := fmt.Sprintf("%s-00002", workloadName)
					latestReadyRevisionName := fmt.Sprintf("%s-00001", workloadName)
					setKnativeService(ctx, k8sClient, workloadName, namespace, &latestReadyRevisionName, &latestCreateRevisionName, nil)

					// create pod with the knative service revision
					pod := buildPodWithRevisionLabel("pod-1", namespace, latestCreateRevisionName, v1.PodRunning)
					Expect(k8sClient.Create(ctx, pod)).To(Succeed())

					// second reconcile to get to Ready(Progressing) phase
					_, err = reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					tmp := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
					Expect(tmp.Status.InferenceStatus.Phase).To(Equal(inferenceaggregator.Ready))
					Expect(tmp.Status.InferenceStatus.PhaseMessage).To(Equal("Workload is ready. Update in progress: container images pulled successfully, new revision is initializing."))
				})

				It("Ready phase - Timed out waiting for the workload to become ready", func() {
					// first reconcile
					inferenceWorkload.CopyTo(jobWorkload)
					_, err := reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					// set the knative
					latestCreateRevisionName := fmt.Sprintf("%s-00002", workloadName)
					latestReadyRevisionName := fmt.Sprintf("%s-00001", workloadName)
					message := "any thing"
					conditions := buildKnativeServiceCondition(servingv1.ServiceConditionConfigurationsReady, v1.ConditionFalse, "RevisionFailed", message)
					setKnativeService(ctx, k8sClient, workloadName, namespace, &latestReadyRevisionName, &latestCreateRevisionName, &conditions)

					// create pod with the knative service revision
					pod := buildPodWithRevisionLabel("pod-1", namespace, latestCreateRevisionName, v1.PodPending)

					Expect(k8sClient.Create(ctx, pod)).To(Succeed())

					// second reconcile to get to Ready(Progressing) phase
					_, err = reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					tmp := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
					Expect(tmp.Status.InferenceStatus.Phase).To(Equal(inferenceaggregator.Ready))
					Expect(tmp.Status.InferenceStatus.PhaseMessage).To(Equal("Workload is ready. Update failed: Timed out waiting for the workload to become ready."))
				})

				It("Ready phase - build error ", func() {
					// already exist service to simulate the isBuildError on our service creation
					alreadyExistService := &servingv1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Name:      workloadName,
							Namespace: namespace,
						},
						// Add relevant spec fields
					}
					Expect(k8sClient.Create(ctx, alreadyExistService)).To(Succeed())

					// first reconcile
					latestCreateRevisionName := fmt.Sprintf("%s-00001", workloadName)
					latestReadyRevisionName := fmt.Sprintf("%s-00002", workloadName)

					setKnativeService(ctx, k8sClient, workloadName, namespace, &latestReadyRevisionName, &latestCreateRevisionName, nil)

					inferenceWorkload.CopyTo(jobWorkload)
					_, err := reconcile.Reconcile(ctx, jobWorkload, inferenceWorkload)
					Expect(err).ToNot(HaveOccurred())

					tmp := getInferenceWorkload(ctx, k8sClient, workloadName, namespace)
					Expect(tmp.Status.InferenceStatus.Phase).To(Equal(inferenceaggregator.Ready))
					Expect(tmp.Status.InferenceStatus.PhaseMessage).To(Equal("Workload is ready. Update failed: [Service:test-infw-workload]  \"test-infw-workload\" already exists"))
				})
			})
		})
	})
})
