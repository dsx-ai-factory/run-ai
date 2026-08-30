package distributed_status

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/builders"

	nvapi "github.com/NVIDIA/k8s-dra-driver-gpu/api/nvidia.com/resource/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	distresources "github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/resources"
	mnnvltypes "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/mnnvl"
	wccommon "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	dinfw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	lwsv1 "sigs.k8s.io/lws/api/leaderworkerset/v1"
)

const (
	testPhaseMessageOneReady   = "Number of replicas: 1; Ready: 1"
	testPhaseMessageThreeReady = "Number of replicas: 3; Ready: 3"
)

// Mock implementations
type mockLwsStatusCalculator struct {
	phase         string
	replicas      int32
	readyReplicas int32
	phaseMessage  string
	err           error
}

func (m *mockLwsStatusCalculator) GetStatus(ctx context.Context, workload *dinfw.DistributedInferenceWorkload, lws *lwsv1.LeaderWorkerSet) (string, int32, int32, string, error) {
	return m.phase, m.replicas, m.readyReplicas, m.phaseMessage, m.err
}

type mockDistributedResourcesManager struct {
	resources        distresources.DistributedResources
	missingResources []client.Object
	err              error
}

func (m *mockDistributedResourcesManager) GetResources(ctx context.Context, builtResources []client.Object) (distresources.DistributedResources, []client.Object, error) {
	return m.resources, m.missingResources, m.err
}

var _ = Describe("DistributedStatusUpdater", func() {
	var (
		ctx               context.Context
		k8sClient         client.Client
		statusUpdater     *DistributedStatusUpdater
		mockLwsCalculator *mockLwsStatusCalculator
		mockResourcesMgr  *mockDistributedResourcesManager
		workload          *dinfw.DistributedInferenceWorkload
		originalWorkload  *dinfw.DistributedInferenceWorkload
		builtResources    []client.Object
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme := runtime.NewScheme()
		_ = dinfw.AddToScheme(scheme)
		_ = corev1.AddToScheme(scheme)
		_ = lwsv1.AddToScheme(scheme)
		k8sClient = fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(&dinfw.DistributedInferenceWorkload{}).Build()

		mockLwsCalculator = &mockLwsStatusCalculator{}
		mockResourcesMgr = &mockDistributedResourcesManager{}

		statusUpdater = &DistributedStatusUpdater{
			k8sClient:           k8sClient,
			lwsStatusCalculator: mockLwsCalculator,
			resourcesManager:    mockResourcesMgr,
		}

		workload = &dinfw.DistributedInferenceWorkload{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "test-workload",
				Namespace:  "test-namespace",
				Generation: 1,
			},
			Status: dinfw.DistributedInferenceWorkloadStatus{},
		}

		originalWorkload = workload.DeepCopy()
		builtResources = []client.Object{}
	})

	Describe("updateStatusFields", func() {
		Context("when workload is not being deleted", func() {
			It("should set default phase and message", func() {
				mockResourcesMgr.resources = distresources.DistributedResources{}
				mockResourcesMgr.missingResources = []client.Object{}
				mockResourcesMgr.err = nil

				err := statusUpdater.updateStatusFields(ctx, workload, builtResources, nil)

				Expect(err).NotTo(HaveOccurred())
				Expect(workload.Status.Phase).To(Equal(PhaseCreating))
				Expect(workload.Status.PhaseMessage).To(Equal(""))
			})

			It("should set first acknowledged time when it's zero", func() {
				mockResourcesMgr.resources = distresources.DistributedResources{}
				mockResourcesMgr.missingResources = []client.Object{}
				mockResourcesMgr.err = nil

				beforeTime := time.Now()
				err := statusUpdater.updateStatusFields(ctx, workload, builtResources, nil)
				afterTime := time.Now()

				Expect(err).NotTo(HaveOccurred())
				Expect(workload.Status.FirstAcknowledgedTime).NotTo(BeNil())
				Expect(workload.Status.FirstAcknowledgedTime.Time).To(BeTemporally(">=", beforeTime))
				Expect(workload.Status.FirstAcknowledgedTime.Time).To(BeTemporally("<=", afterTime))
			})

			It("should not overwrite existing first acknowledged time", func() {
				existingTime := metav1.NewTime(time.Now().Add(-1 * time.Hour))
				workload.Status.FirstAcknowledgedTime = &existingTime
				mockResourcesMgr.resources = distresources.DistributedResources{}
				mockResourcesMgr.missingResources = []client.Object{}
				mockResourcesMgr.err = nil

				err := statusUpdater.updateStatusFields(ctx, workload, builtResources, nil)

				Expect(err).NotTo(HaveOccurred())
				Expect(workload.Status.FirstAcknowledgedTime).To(Equal(&existingTime))
			})

			It("should set observed generation to workload generation", func() {
				workload.Generation = 5
				mockResourcesMgr.resources = distresources.DistributedResources{}
				mockResourcesMgr.missingResources = []client.Object{}
				mockResourcesMgr.err = nil

				err := statusUpdater.updateStatusFields(ctx, workload, builtResources, nil)

				Expect(err).NotTo(HaveOccurred())
				Expect(workload.Status.ObservedGeneration).To(Equal(int64(5)))
			})

			It("should set last status change time", func() {
				mockResourcesMgr.resources = distresources.DistributedResources{}
				mockResourcesMgr.missingResources = []client.Object{}
				mockResourcesMgr.err = nil

				beforeTime := time.Now()
				err := statusUpdater.updateStatusFields(ctx, workload, builtResources, nil)
				afterTime := time.Now()

				Expect(err).NotTo(HaveOccurred())
				Expect(workload.Status.LastStatusChangeTime).NotTo(BeNil())
				Expect(workload.Status.LastStatusChangeTime.Time).To(BeTemporally(">=", beforeTime))
				Expect(workload.Status.LastStatusChangeTime.Time).To(BeTemporally("<=", afterTime))
			})
		})

		Context("when workload is being deleted", func() {
			BeforeEach(func() {
				mockResourcesMgr.resources = distresources.DistributedResources{}
				mockResourcesMgr.missingResources = []client.Object{}
				mockResourcesMgr.err = nil

				now := metav1.NewTime(time.Now())
				workload.DeletionTimestamp = &now
			})

			It("should set phase to deleting and return early", func() {
				err := statusUpdater.updateStatusFields(ctx, workload, builtResources, nil)

				Expect(err).NotTo(HaveOccurred())
				Expect(workload.Status.Phase).To(Equal(PhaseDeleting))
			})
		})

		Context("when resources manager returns error", func() {
			It("should return the error", func() {
				expectedErr := errors.New("resource manager error")
				mockResourcesMgr.resources = distresources.DistributedResources{}
				mockResourcesMgr.missingResources = []client.Object{}
				mockResourcesMgr.err = expectedErr

				err := statusUpdater.updateStatusFields(ctx, workload, builtResources, nil)

				Expect(err).To(Equal(expectedErr))
			})
		})

		Context("when no LWS resource is found", func() {
			It("should handle gracefully when no build error", func() {
				mockResourcesMgr.resources = distresources.DistributedResources{
					Lws: nil,
				}
				mockResourcesMgr.missingResources = []client.Object{}
				mockResourcesMgr.err = nil

				err := statusUpdater.updateStatusFields(ctx, workload, builtResources, nil)

				Expect(err).NotTo(HaveOccurred())
				Expect(workload.Status.Phase).To(Equal(PhaseCreating))
			})

			It("should set phase to failed when build error contains LeaderWorkerSet", func() {
				mockResourcesMgr.resources = distresources.DistributedResources{
					Lws: nil,
				}
				mockResourcesMgr.missingResources = []client.Object{}
				mockResourcesMgr.err = nil

				buildErr := fmt.Errorf("failed to create %s resource", constants.KindLeaderWorkerSet)

				err := statusUpdater.updateStatusFields(ctx, workload, builtResources, &buildErr)

				Expect(err).NotTo(HaveOccurred())
				Expect(workload.Status.Phase).To(Equal(PhaseFailed))
			})
		})

		Context("when LWS resource exists", func() {
			var lws *lwsv1.LeaderWorkerSet

			BeforeEach(func() {
				lws = &lwsv1.LeaderWorkerSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-lws",
						Namespace: "test-namespace",
					},
				}
				mockResourcesMgr.resources = distresources.DistributedResources{
					Lws: lws,
				}
				mockResourcesMgr.missingResources = []client.Object{}
				mockResourcesMgr.err = nil
			})

			It("should return error when LWS status calculator fails", func() {
				expectedErr := errors.New("lws calculator error")
				mockLwsCalculator.err = expectedErr

				err := statusUpdater.updateStatusFields(ctx, workload, builtResources, nil)

				Expect(err).To(Equal(expectedErr))
			})

			It("should set phase, replicas and ready replicas from LWS status", func() {
				mockLwsCalculator.phase = PhaseReady
				mockLwsCalculator.replicas = 3
				mockLwsCalculator.readyReplicas = 2
				mockLwsCalculator.phaseMessage = "Number of replicas: 3; Ready: 2; Initializing: 1"
				mockLwsCalculator.err = nil

				err := statusUpdater.updateStatusFields(ctx, workload, builtResources, nil)

				Expect(err).NotTo(HaveOccurred())
				Expect(workload.Status.Phase).To(Equal(PhaseReady))
				Expect(workload.Status.Replicas).To(Equal(int32(3)))
				Expect(workload.Status.ReadyReplicas).To(Equal(int32(2)))
				Expect(workload.Status.PhaseMessage).ToNot(BeEmpty())
			})

			It("should set phase, replicas and ready replicas from LWS status when MNNVL is ready", func() {
				mockLwsCalculator.phase = PhaseReady
				mockLwsCalculator.replicas = 1
				mockLwsCalculator.readyReplicas = 1
				mockLwsCalculator.phaseMessage = testPhaseMessageOneReady
				mockLwsCalculator.err = nil

				computeDomain := &nvapi.ComputeDomain{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-compute-domain",
					},
					Status: nvapi.ComputeDomainStatus{
						Status: constants.ComputeDomainStatusReady,
					},
				}
				mockResourcesMgr.resources.ComputeDomains = []*nvapi.ComputeDomain{computeDomain}

				err := statusUpdater.updateStatusFields(ctx, workload, builtResources, nil)

				Expect(err).NotTo(HaveOccurred())
				Expect(workload.Status.Phase).To(Equal(PhaseReady))
				Expect(workload.Status.Replicas).To(Equal(int32(1)))
				Expect(workload.Status.ReadyReplicas).To(Equal(int32(1)))
				Expect(workload.Status.PhaseMessage).ToNot(BeEmpty())
			})

			It("should set phase to degraded when LWS is ready but there's a build error", func() {
				mockLwsCalculator.phase = PhaseReady
				mockLwsCalculator.replicas = 3
				mockLwsCalculator.readyReplicas = 3
				mockLwsCalculator.phaseMessage = testPhaseMessageThreeReady
				mockLwsCalculator.err = nil

				buildErr := errors.New("some build error")

				err := statusUpdater.updateStatusFields(ctx, workload, builtResources, &buildErr)

				Expect(err).NotTo(HaveOccurred())
				Expect(workload.Status.Phase).To(Equal(PhaseDegraded))
				Expect(workload.Status.Replicas).To(Equal(int32(3)))
				Expect(workload.Status.ReadyReplicas).To(Equal(int32(3)))
				Expect(workload.Status.PhaseMessage).To(ContainSubstring("build error"))
			})

			It("should include build error message when LWS is initializing with build error", func() {
				mockLwsCalculator.phase = PhaseInitializing
				mockLwsCalculator.replicas = 3
				mockLwsCalculator.readyReplicas = 1
				mockLwsCalculator.phaseMessage = "Number of replicas: 3; Ready: 1; Initializing: 2"
				mockLwsCalculator.err = nil

				buildErr := errors.New("resource creation failed")

				err := statusUpdater.updateStatusFields(ctx, workload, builtResources, &buildErr)

				Expect(err).NotTo(HaveOccurred())
				Expect(workload.Status.Phase).To(Equal(PhaseInitializing)) // Phase should remain Initializing
				Expect(workload.Status.Replicas).To(Equal(int32(3)))
				Expect(workload.Status.ReadyReplicas).To(Equal(int32(1)))
				Expect(workload.Status.PhaseMessage).To(ContainSubstring("resource creation failed"))
			})

			It("should set phase to degraded when LWS is ready but resources are missing", func() {
				mockLwsCalculator.phase = PhaseReady
				mockLwsCalculator.replicas = 3
				mockLwsCalculator.readyReplicas = 3
				mockLwsCalculator.phaseMessage = testPhaseMessageThreeReady
				mockLwsCalculator.err = nil

				// Create a missing service with proper metadata
				missingService := &corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name: "missing-service",
					},
					TypeMeta: metav1.TypeMeta{
						Kind: constants.KindService,
					},
				}
				mockResourcesMgr.missingResources = []client.Object{missingService}

				err := statusUpdater.updateStatusFields(ctx, workload, builtResources, nil)

				Expect(err).NotTo(HaveOccurred())
				Expect(workload.Status.Phase).To(Equal(PhaseDegraded))
				Expect(workload.Status.PhaseMessage).To(ContainSubstring("Missing resources: Service/missing-service"))
			})

			It("should set phase to degraded when LWS is ready but multiple resources are missing", func() {
				mockLwsCalculator.phase = PhaseReady
				mockLwsCalculator.replicas = 3
				mockLwsCalculator.readyReplicas = 3
				mockLwsCalculator.phaseMessage = testPhaseMessageThreeReady
				mockLwsCalculator.err = nil

				// Create multiple missing resources with proper metadata
				missingService := &corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name: "missing-service",
					},
					TypeMeta: metav1.TypeMeta{
						Kind: constants.KindService,
					},
				}
				missingIngress := &networkingv1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name: "missing-ingress",
					},
					TypeMeta: metav1.TypeMeta{
						Kind: constants.KindIngress,
					},
				}
				mockResourcesMgr.missingResources = []client.Object{missingService, missingIngress}

				err := statusUpdater.updateStatusFields(ctx, workload, builtResources, nil)

				Expect(err).NotTo(HaveOccurred())
				Expect(workload.Status.Phase).To(Equal(PhaseDegraded))
				Expect(workload.Status.PhaseMessage).To(ContainSubstring("Missing resources: Service/missing-service, Ingress/missing-ingress"))
			})

			It("should set phase to degraded when LWS is ready, but MNNVL is not ready", func() {
				mockLwsCalculator.phase = PhaseReady
				mockLwsCalculator.replicas = 1
				mockLwsCalculator.readyReplicas = 1
				mockLwsCalculator.phaseMessage = testPhaseMessageOneReady
				mockLwsCalculator.err = nil

				// Create a compute domain resource that exists but is not ready
				computeDomain := &nvapi.ComputeDomain{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-compute-domain",
					},
					Status: nvapi.ComputeDomainStatus{
						Status: constants.ComputeDomainStatusNotReady,
					},
				}

				mockResourcesMgr.resources = distresources.DistributedResources{
					Lws:            lws,
					ComputeDomains: []*nvapi.ComputeDomain{computeDomain},
				}
				mockResourcesMgr.missingResources = []client.Object{} // Empty since resource exists

				err := statusUpdater.updateStatusFields(ctx, workload, builtResources, nil)

				Expect(err).NotTo(HaveOccurred())
				Expect(workload.Status.Phase).To(Equal(PhaseDegraded))
				Expect(workload.Status.PhaseMessage).To(ContainSubstring(mnnvltypes.UserFacingFeatureName))
				Expect(workload.Status.PhaseMessage).To(ContainSubstring("is not ready"))
			})

			Context("when setting services", func() {
				It("should set service statuses from resources", func() {
					service := &corev1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test-service",
						},
						Spec: corev1.ServiceSpec{
							Type: corev1.ServiceTypeClusterIP,
							Ports: []corev1.ServicePort{
								{
									Name:       "http",
									Protocol:   corev1.ProtocolTCP,
									Port:       80,
									TargetPort: intstr.FromInt(8080),
								},
								{
									Name:       "https",
									Protocol:   corev1.ProtocolTCP,
									Port:       443,
									TargetPort: intstr.FromInt(8443),
									NodePort:   30443,
								},
							},
						},
					}

					mockResourcesMgr.resources = distresources.DistributedResources{
						Lws:      lws,
						Services: []*corev1.Service{service},
					}
					mockLwsCalculator.phase = PhaseReady
					mockLwsCalculator.replicas = 1
					mockLwsCalculator.readyReplicas = 1
					mockLwsCalculator.phaseMessage = testPhaseMessageOneReady
					mockLwsCalculator.err = nil

					err := statusUpdater.updateStatusFields(ctx, workload, builtResources, nil)

					Expect(err).NotTo(HaveOccurred())
					Expect(workload.Status.NetworkStatus.ServiceStatuses).To(HaveLen(1))

					serviceStatus := workload.Status.NetworkStatus.ServiceStatuses[0]
					Expect(serviceStatus.Name).To(Equal("test-service"))
					Expect(serviceStatus.Type).To(Equal(corev1.ServiceTypeClusterIP))
					Expect(serviceStatus.Ports).To(HaveLen(2))

					// Check first port (without NodePort)
					Expect(serviceStatus.Ports[0].Name).To(Equal("http"))
					Expect(serviceStatus.Ports[0].Protocol).To(Equal(corev1.ProtocolTCP))
					Expect(serviceStatus.Ports[0].Port).To(Equal(int32(80)))
					Expect(serviceStatus.Ports[0].TargetPort).To(Equal(intstr.FromInt(8080)))

					// Check second port (with NodePort)
					Expect(serviceStatus.Ports[1].Name).To(Equal("https"))
					Expect(serviceStatus.Ports[1].Protocol).To(Equal(corev1.ProtocolTCP))
					Expect(serviceStatus.Ports[1].Port).To(Equal(int32(30443))) // Should use NodePort
					Expect(serviceStatus.Ports[1].TargetPort).To(Equal(intstr.FromInt(8443)))

					Expect(workload.Status.PhaseMessage).ToNot(BeEmpty())
				})
			})

			Context("when setting URLs", func() {
				It("should set external URLs from ingresses", func() {
					ingress := &networkingv1.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test-ingress",
						},
						Spec: networkingv1.IngressSpec{
							Rules: []networkingv1.IngressRule{
								{
									Host: "example.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path: "/api",
												},
											},
										},
									},
								},
							},
						},
					}

					mockResourcesMgr.resources = distresources.DistributedResources{
						Lws:       lws,
						Ingresses: []*networkingv1.Ingress{ingress},
					}
					mockLwsCalculator.phase = PhaseReady
					mockLwsCalculator.replicas = 1
					mockLwsCalculator.readyReplicas = 1
					mockLwsCalculator.phaseMessage = testPhaseMessageOneReady
					mockLwsCalculator.err = nil

					err := statusUpdater.updateStatusFields(ctx, workload, builtResources, nil)

					Expect(err).NotTo(HaveOccurred())
					Expect(workload.Status.DistributedInferenceStatus.ExternalUrls).To(HaveLen(1))
					// The actual URL format depends on the BuildURLFromIngress implementation
					Expect(workload.Status.DistributedInferenceStatus.ExternalUrls[0]).To(Equal("https://example.com/api"))
					Expect(workload.Status.PhaseMessage).ToNot(BeEmpty())
				})

				It("should set external URLs from routes", func() {
					route := &routev1.Route{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test-route",
						},
						Spec: routev1.RouteSpec{
							Host: "route.example.com",
						},
					}

					mockResourcesMgr.resources = distresources.DistributedResources{
						Lws:    lws,
						Routes: []*routev1.Route{route},
					}
					mockLwsCalculator.phase = PhaseReady
					mockLwsCalculator.replicas = 1
					mockLwsCalculator.readyReplicas = 1
					mockLwsCalculator.phaseMessage = testPhaseMessageOneReady
					mockLwsCalculator.err = nil

					err := statusUpdater.updateStatusFields(ctx, workload, builtResources, nil)

					Expect(err).NotTo(HaveOccurred())
					Expect(workload.Status.DistributedInferenceStatus.ExternalUrls).To(HaveLen(1))
					// The actual URL format depends on the BuildURLFromRoute implementation
					Expect(workload.Status.DistributedInferenceStatus.ExternalUrls[0]).To(ContainSubstring("route.example.com"))
					Expect(workload.Status.PhaseMessage).ToNot(BeEmpty())
				})

				Context("when setting servingUrls", func() {
					It("should set servingUrls from service when no ingress/route exists", func() {
						serviceName := builders.GetServingPortNetworkResourceName(workload.Name)
						service := &corev1.Service{
							ObjectMeta: metav1.ObjectMeta{
								Name:      serviceName,
								Namespace: "test-namespace",
							},
							Spec: corev1.ServiceSpec{
								Ports: []corev1.ServicePort{
									{
										Name: "http",
										Port: 80,
									},
								},
							},
						}

						mockResourcesMgr.resources = distresources.DistributedResources{
							Lws:      lws,
							Services: []*corev1.Service{service},
						}
						mockLwsCalculator.phase = PhaseReady
						mockLwsCalculator.replicas = 1
						mockLwsCalculator.readyReplicas = 1
						mockLwsCalculator.phaseMessage = testPhaseMessageOneReady
						mockLwsCalculator.err = nil

						err := statusUpdater.updateStatusFields(ctx, workload, builtResources, nil)

						Expect(err).NotTo(HaveOccurred())
						Expect(workload.Status.DistributedInferenceStatus.ServingUrls).To(HaveLen(1))
						Expect(workload.Status.DistributedInferenceStatus.ServingUrls[0]).To(Equal(fmt.Sprintf("http://%s.test-namespace.svc.cluster.local", serviceName)))
						Expect(workload.Status.DistributedInferenceStatus.ExternalUrls).To(BeEmpty())
					})

					It("should set servingUrls from ingress when serving ingress exists", func() {
						serviceName := builders.GetServingPortNetworkResourceName(workload.Name)
						service := &corev1.Service{
							ObjectMeta: metav1.ObjectMeta{
								Name:      serviceName,
								Namespace: "test-namespace",
							},
							Spec: corev1.ServiceSpec{
								Ports: []corev1.ServicePort{
									{
										Name: "http",
										Port: 80,
									},
								},
							},
						}

						servingIngress := &networkingv1.Ingress{
							ObjectMeta: metav1.ObjectMeta{
								Name: serviceName, // Same name as service - this makes it a serving ingress
							},
							Spec: networkingv1.IngressSpec{
								Rules: []networkingv1.IngressRule{
									{
										Host: "serving.example.com",
										IngressRuleValue: networkingv1.IngressRuleValue{
											HTTP: &networkingv1.HTTPIngressRuleValue{
												Paths: []networkingv1.HTTPIngressPath{
													{
														Path: "/serve",
													},
												},
											},
										},
									},
								},
							},
						}

						externalIngress := &networkingv1.Ingress{
							ObjectMeta: metav1.ObjectMeta{
								Name: "external-ingress", // Different name - this goes to external URLs
							},
							Spec: networkingv1.IngressSpec{
								Rules: []networkingv1.IngressRule{
									{
										Host: "external.example.com",
										IngressRuleValue: networkingv1.IngressRuleValue{
											HTTP: &networkingv1.HTTPIngressRuleValue{
												Paths: []networkingv1.HTTPIngressPath{
													{
														Path: "/external",
													},
												},
											},
										},
									},
								},
							},
						}

						mockResourcesMgr.resources = distresources.DistributedResources{
							Lws:       lws,
							Services:  []*corev1.Service{service},
							Ingresses: []*networkingv1.Ingress{servingIngress, externalIngress},
						}
						mockLwsCalculator.phase = PhaseReady
						mockLwsCalculator.replicas = 1
						mockLwsCalculator.readyReplicas = 1
						mockLwsCalculator.phaseMessage = testPhaseMessageOneReady
						mockLwsCalculator.err = nil

						err := statusUpdater.updateStatusFields(ctx, workload, builtResources, nil)

						Expect(err).NotTo(HaveOccurred())
						Expect(workload.Status.DistributedInferenceStatus.ServingUrls).To(HaveLen(1))
						Expect(workload.Status.DistributedInferenceStatus.ServingUrls[0]).To(Equal("https://serving.example.com/serve"))
						Expect(workload.Status.DistributedInferenceStatus.ExternalUrls).To(HaveLen(1))
						Expect(workload.Status.DistributedInferenceStatus.ExternalUrls[0]).To(Equal("https://external.example.com/external"))
					})

					It("should set servingUrls from route when serving route exists", func() {
						serviceName := builders.GetServingPortNetworkResourceName(workload.Name)
						service := &corev1.Service{
							ObjectMeta: metav1.ObjectMeta{
								Name:      serviceName,
								Namespace: "test-namespace",
							},
							Spec: corev1.ServiceSpec{
								Ports: []corev1.ServicePort{
									{
										Name: "http",
										Port: 80,
									},
								},
							},
						}

						servingRoute := &routev1.Route{
							ObjectMeta: metav1.ObjectMeta{
								Name: serviceName, // Same name as service - this makes it a serving route
							},
							Spec: routev1.RouteSpec{
								Host: "serving-route.example.com",
							},
						}

						externalRoute := &routev1.Route{
							ObjectMeta: metav1.ObjectMeta{
								Name: "external-route", // Different name - this goes to external URLs
							},
							Spec: routev1.RouteSpec{
								Host: "external-route.example.com",
							},
						}

						mockResourcesMgr.resources = distresources.DistributedResources{
							Lws:      lws,
							Services: []*corev1.Service{service},
							Routes:   []*routev1.Route{servingRoute, externalRoute},
						}
						mockLwsCalculator.phase = PhaseReady
						mockLwsCalculator.replicas = 1
						mockLwsCalculator.readyReplicas = 1
						mockLwsCalculator.phaseMessage = testPhaseMessageOneReady
						mockLwsCalculator.err = nil

						err := statusUpdater.updateStatusFields(ctx, workload, builtResources, nil)

						Expect(err).NotTo(HaveOccurred())
						Expect(workload.Status.DistributedInferenceStatus.ServingUrls).To(HaveLen(1))
						Expect(workload.Status.DistributedInferenceStatus.ServingUrls[0]).To(ContainSubstring("serving-route.example.com"))
						Expect(workload.Status.DistributedInferenceStatus.ExternalUrls).To(HaveLen(1))
						Expect(workload.Status.DistributedInferenceStatus.ExternalUrls[0]).To(ContainSubstring("external-route.example.com"))
					})
				})

				It("should set external URLs from HTTPRoutes", func() {
					httpRoute := &gatewayv1.HTTPRoute{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test-httproute",
						},
						Spec: gatewayv1.HTTPRouteSpec{
							Hostnames: []gatewayv1.Hostname{"gateway.example.com"},
							Rules: []gatewayv1.HTTPRouteRule{
								{
									Matches: []gatewayv1.HTTPRouteMatch{
										{
											Path: &gatewayv1.HTTPPathMatch{
												Type:  ptrTo(gatewayv1.PathMatchPathPrefix),
												Value: ptrTo("/api"),
											},
										},
									},
								},
							},
						},
					}

					mockResourcesMgr.resources = distresources.DistributedResources{
						Lws:        lws,
						HTTPRoutes: []*gatewayv1.HTTPRoute{httpRoute},
					}
					mockLwsCalculator.phase = PhaseReady
					mockLwsCalculator.replicas = 1
					mockLwsCalculator.readyReplicas = 1
					mockLwsCalculator.phaseMessage = testPhaseMessageOneReady
					mockLwsCalculator.err = nil

					err := statusUpdater.updateStatusFields(ctx, workload, builtResources, nil)

					Expect(err).NotTo(HaveOccurred())
					Expect(workload.Status.Phase).To(Equal(PhaseReady))
					Expect(workload.Status.DistributedInferenceStatus.ExternalUrls).To(HaveLen(1))
					Expect(workload.Status.DistributedInferenceStatus.ExternalUrls[0]).To(ContainSubstring("gateway.example.com"))
					Expect(workload.Status.DistributedInferenceStatus.ExternalUrls[0]).To(ContainSubstring("/api"))
				})

				It("should set servingUrls from HTTPRoute when serving HTTPRoute exists", func() {
					serviceName := builders.GetServingPortNetworkResourceName(workload.Name)
					service := &corev1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Name:      serviceName,
							Namespace: "test-namespace",
						},
						Spec: corev1.ServiceSpec{
							Ports: []corev1.ServicePort{
								{
									Name: "http",
									Port: 80,
								},
							},
						},
					}

					servingHTTPRoute := &gatewayv1.HTTPRoute{
						ObjectMeta: metav1.ObjectMeta{
							Name: serviceName,
						},
						Spec: gatewayv1.HTTPRouteSpec{
							Hostnames: []gatewayv1.Hostname{"serving-gw.example.com"},
							Rules: []gatewayv1.HTTPRouteRule{
								{
									Matches: []gatewayv1.HTTPRouteMatch{
										{
											Path: &gatewayv1.HTTPPathMatch{
												Type:  ptrTo(gatewayv1.PathMatchPathPrefix),
												Value: ptrTo("/serve"),
											},
										},
									},
								},
							},
						},
					}

					externalHTTPRoute := &gatewayv1.HTTPRoute{
						ObjectMeta: metav1.ObjectMeta{
							Name: "external-httproute",
						},
						Spec: gatewayv1.HTTPRouteSpec{
							Hostnames: []gatewayv1.Hostname{"external-gw.example.com"},
							Rules: []gatewayv1.HTTPRouteRule{
								{
									Matches: []gatewayv1.HTTPRouteMatch{
										{
											Path: &gatewayv1.HTTPPathMatch{
												Type:  ptrTo(gatewayv1.PathMatchPathPrefix),
												Value: ptrTo("/external"),
											},
										},
									},
								},
							},
						},
					}

					mockResourcesMgr.resources = distresources.DistributedResources{
						Lws:        lws,
						Services:   []*corev1.Service{service},
						HTTPRoutes: []*gatewayv1.HTTPRoute{servingHTTPRoute, externalHTTPRoute},
					}
					mockLwsCalculator.phase = PhaseReady
					mockLwsCalculator.replicas = 1
					mockLwsCalculator.readyReplicas = 1
					mockLwsCalculator.phaseMessage = testPhaseMessageOneReady
					mockLwsCalculator.err = nil

					err := statusUpdater.updateStatusFields(ctx, workload, builtResources, nil)

					Expect(err).NotTo(HaveOccurred())
					Expect(workload.Status.DistributedInferenceStatus.ServingUrls).To(HaveLen(1))
					Expect(workload.Status.DistributedInferenceStatus.ServingUrls[0]).To(ContainSubstring("serving-gw.example.com"))
					Expect(workload.Status.DistributedInferenceStatus.ExternalUrls).To(HaveLen(1))
					Expect(workload.Status.DistributedInferenceStatus.ExternalUrls[0]).To(ContainSubstring("external-gw.example.com"))
				})

				It("should deduplicate URLs in dual-mode with Ingress and HTTPRoute", func() {
					ingress := &networkingv1.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test-ingress",
						},
						Spec: networkingv1.IngressSpec{
							Rules: []networkingv1.IngressRule{
								{
									Host: "shared.example.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path: "/api",
												},
											},
										},
									},
								},
							},
						},
					}

					httpRoute := &gatewayv1.HTTPRoute{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test-httproute",
						},
						Spec: gatewayv1.HTTPRouteSpec{
							Hostnames: []gatewayv1.Hostname{"shared.example.com"},
							Rules: []gatewayv1.HTTPRouteRule{
								{
									Matches: []gatewayv1.HTTPRouteMatch{
										{
											Path: &gatewayv1.HTTPPathMatch{
												Type:  ptrTo(gatewayv1.PathMatchPathPrefix),
												Value: ptrTo("/api"),
											},
										},
									},
								},
							},
						},
					}

					mockResourcesMgr.resources = distresources.DistributedResources{
						Lws:        lws,
						Ingresses:  []*networkingv1.Ingress{ingress},
						HTTPRoutes: []*gatewayv1.HTTPRoute{httpRoute},
					}
					mockLwsCalculator.phase = PhaseReady
					mockLwsCalculator.replicas = 1
					mockLwsCalculator.readyReplicas = 1
					mockLwsCalculator.phaseMessage = testPhaseMessageOneReady
					mockLwsCalculator.err = nil

					err := statusUpdater.updateStatusFields(ctx, workload, builtResources, nil)

					Expect(err).NotTo(HaveOccurred())
					Expect(workload.Status.DistributedInferenceStatus.ExternalUrls).To(HaveLen(1))
					Expect(workload.Status.DistributedInferenceStatus.ExternalUrls[0]).To(ContainSubstring("shared.example.com"))
				})

				It("should not set HTTPRoute URLs when readyReplicas = 0", func() {
					httpRoute := &gatewayv1.HTTPRoute{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test-httproute",
						},
						Spec: gatewayv1.HTTPRouteSpec{
							Hostnames: []gatewayv1.Hostname{"gateway.example.com"},
							Rules: []gatewayv1.HTTPRouteRule{
								{
									Matches: []gatewayv1.HTTPRouteMatch{
										{
											Path: &gatewayv1.HTTPPathMatch{
												Type:  ptrTo(gatewayv1.PathMatchPathPrefix),
												Value: ptrTo("/api"),
											},
										},
									},
								},
							},
						},
					}

					mockResourcesMgr.resources = distresources.DistributedResources{
						Lws:        lws,
						HTTPRoutes: []*gatewayv1.HTTPRoute{httpRoute},
					}
					mockLwsCalculator.phase = PhaseInitializing
					mockLwsCalculator.replicas = 3
					mockLwsCalculator.readyReplicas = 0
					mockLwsCalculator.phaseMessage = "Number of replicas: 3; Ready: 0; Initializing: 3"
					mockLwsCalculator.err = nil

					err := statusUpdater.updateStatusFields(ctx, workload, builtResources, nil)

					Expect(err).NotTo(HaveOccurred())
					Expect(workload.Status.DistributedInferenceStatus.ExternalUrls).To(BeEmpty())
					Expect(workload.Status.DistributedInferenceStatus.ServingUrls).To(BeEmpty())
				})

				It("should set phase to degraded when missing HTTPRoute resource", func() {
					mockLwsCalculator.phase = PhaseReady
					mockLwsCalculator.replicas = 1
					mockLwsCalculator.readyReplicas = 1
					mockLwsCalculator.phaseMessage = testPhaseMessageOneReady
					mockLwsCalculator.err = nil

					missingHTTPRoute := &gatewayv1.HTTPRoute{
						ObjectMeta: metav1.ObjectMeta{
							Name: "missing-httproute",
						},
						TypeMeta: metav1.TypeMeta{
							Kind: constants.KindHTTPRoute,
						},
					}
					mockResourcesMgr.missingResources = []client.Object{missingHTTPRoute}

					err := statusUpdater.updateStatusFields(ctx, workload, builtResources, nil)

					Expect(err).NotTo(HaveOccurred())
					Expect(workload.Status.Phase).To(Equal(PhaseDegraded))
					Expect(workload.Status.PhaseMessage).To(ContainSubstring("Missing resources: HTTPRoute/missing-httproute"))
				})

				It("should set phase to degraded in dual-mode when HTTPRoute is missing but Ingress exists", func() {
					mockLwsCalculator.phase = PhaseReady
					mockLwsCalculator.replicas = 1
					mockLwsCalculator.readyReplicas = 1
					mockLwsCalculator.phaseMessage = testPhaseMessageOneReady
					mockLwsCalculator.err = nil

					ingress := &networkingv1.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name: "existing-ingress",
						},
						Spec: networkingv1.IngressSpec{
							Rules: []networkingv1.IngressRule{
								{
									Host: "example.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{Path: "/api"},
											},
										},
									},
								},
							},
						},
					}
					mockResourcesMgr.resources = distresources.DistributedResources{
						Lws:       lws,
						Ingresses: []*networkingv1.Ingress{ingress},
					}

					missingHTTPRoute := &gatewayv1.HTTPRoute{
						ObjectMeta: metav1.ObjectMeta{
							Name: "missing-httproute",
						},
						TypeMeta: metav1.TypeMeta{
							Kind: constants.KindHTTPRoute,
						},
					}
					mockResourcesMgr.missingResources = []client.Object{missingHTTPRoute}

					err := statusUpdater.updateStatusFields(ctx, workload, builtResources, nil)

					Expect(err).NotTo(HaveOccurred())
					Expect(workload.Status.Phase).To(Equal(PhaseDegraded))
					Expect(workload.Status.PhaseMessage).To(ContainSubstring("HTTPRoute/missing-httproute"))
					Expect(workload.Status.DistributedInferenceStatus.ExternalUrls).To(HaveLen(1))
					Expect(workload.Status.DistributedInferenceStatus.ExternalUrls[0]).To(ContainSubstring("example.com"))
				})

				It("should handle services with no ports gracefully", func() {
					service := &corev1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-service-no-ports",
							Namespace: "test-namespace",
						},
						Spec: corev1.ServiceSpec{
							Ports: []corev1.ServicePort{}, // empty ports
						},
					}

					mockResourcesMgr.resources = distresources.DistributedResources{
						Lws:      lws,
						Services: []*corev1.Service{service},
					}
					mockLwsCalculator.phase = PhaseReady
					mockLwsCalculator.replicas = 1
					mockLwsCalculator.readyReplicas = 1
					mockLwsCalculator.phaseMessage = testPhaseMessageOneReady
					mockLwsCalculator.err = nil

					err := statusUpdater.updateStatusFields(ctx, workload, builtResources, nil)

					Expect(err).NotTo(HaveOccurred())
					// Should not have any URLs since service has no ports
					Expect(workload.Status.DistributedInferenceStatus.ExternalUrls).To(BeEmpty())
					Expect(workload.Status.PhaseMessage).ToNot(BeEmpty())
				})

				Context("when readyReplicas = 0", func() {
					It("should not set URLs even when network resources exist", func() {
						service := &corev1.Service{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test-service",
								Namespace: "test-namespace",
							},
							Spec: corev1.ServiceSpec{
								Ports: []corev1.ServicePort{
									{
										Name: "http",
										Port: 8080,
									},
								},
							},
						}

						ingress := &networkingv1.Ingress{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-ingress",
							},
							Spec: networkingv1.IngressSpec{
								Rules: []networkingv1.IngressRule{
									{
										Host: "example.com",
										IngressRuleValue: networkingv1.IngressRuleValue{
											HTTP: &networkingv1.HTTPIngressRuleValue{
												Paths: []networkingv1.HTTPIngressPath{
													{
														Path: "/api",
													},
												},
											},
										},
									},
								},
							},
						}

						route := &routev1.Route{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-route",
							},
							Spec: routev1.RouteSpec{
								Host: "route.example.com",
							},
						}

						mockResourcesMgr.resources = distresources.DistributedResources{
							Lws:       lws,
							Services:  []*corev1.Service{service},
							Ingresses: []*networkingv1.Ingress{ingress},
							Routes:    []*routev1.Route{route},
						}
						mockLwsCalculator.phase = PhaseInitializing
						mockLwsCalculator.replicas = 5
						mockLwsCalculator.readyReplicas = 0 // No ready replicas
						mockLwsCalculator.phaseMessage = "Number of replicas: 5; Ready: 0; Initializing: 5"
						mockLwsCalculator.err = nil

						err := statusUpdater.updateStatusFields(ctx, workload, builtResources, nil)

						Expect(err).NotTo(HaveOccurred())
						// Should not have any URLs since readyReplicas = 0
						Expect(workload.Status.DistributedInferenceStatus.ServingUrls).To(BeEmpty())
						Expect(workload.Status.DistributedInferenceStatus.ExternalUrls).To(BeEmpty())
					})
				})

				Context("when readyReplicas > 0", func() {
					It("should set URLs when network resources exist", func() {
						service := &corev1.Service{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test-service",
								Namespace: "test-namespace",
							},
							Spec: corev1.ServiceSpec{
								Ports: []corev1.ServicePort{
									{
										Name: "http",
										Port: 8080,
									},
								},
							},
						}

						ingress := &networkingv1.Ingress{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-ingress",
							},
							Spec: networkingv1.IngressSpec{
								Rules: []networkingv1.IngressRule{
									{
										Host: "example.com",
										IngressRuleValue: networkingv1.IngressRuleValue{
											HTTP: &networkingv1.HTTPIngressRuleValue{
												Paths: []networkingv1.HTTPIngressPath{
													{
														Path: "/api",
													},
												},
											},
										},
									},
								},
							},
						}

						route := &routev1.Route{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-route",
							},
							Spec: routev1.RouteSpec{
								Host: "route.example.com",
							},
						}

						mockResourcesMgr.resources = distresources.DistributedResources{
							Lws:       lws,
							Services:  []*corev1.Service{service},
							Ingresses: []*networkingv1.Ingress{ingress},
							Routes:    []*routev1.Route{route},
						}
						mockLwsCalculator.phase = PhaseReady
						mockLwsCalculator.replicas = 2
						mockLwsCalculator.readyReplicas = 1 // Some ready replicas
						mockLwsCalculator.phaseMessage = "Number of replicas: 2; Ready: 1; Initializing: 1"
						mockLwsCalculator.err = nil

						err := statusUpdater.updateStatusFields(ctx, workload, builtResources, nil)

						Expect(err).NotTo(HaveOccurred())
						// Should have both internal and external URLs since readyReplicas > 0
						Expect(workload.Status.DistributedInferenceStatus.ExternalUrls).To(HaveLen(2))
						Expect(workload.Status.DistributedInferenceStatus.ExternalUrls).To(ContainElement("https://example.com/api"))
						Expect(workload.Status.DistributedInferenceStatus.ExternalUrls).To(ContainElement(ContainSubstring("route.example.com")))
					})
				})
			})
		})
	})

	Describe("statusChanged", func() {
		It("should return false when statuses are identical", func() {
			previous := &dinfw.DistributedInferenceWorkloadStatus{
				DistributedInferenceStatus: wccommon.DistributedInferenceStatus{
					Phase:    PhaseReady,
					Replicas: 3,
				},
			}
			current := &dinfw.DistributedInferenceWorkloadStatus{
				DistributedInferenceStatus: wccommon.DistributedInferenceStatus{
					Phase:    PhaseReady,
					Replicas: 3,
				},
			}

			changed := statusUpdater.statusChanged(previous, current)

			Expect(changed).To(BeFalse())
		})

		It("should return true when phases are different", func() {
			previous := &dinfw.DistributedInferenceWorkloadStatus{
				DistributedInferenceStatus: wccommon.DistributedInferenceStatus{
					Phase:    PhaseCreating,
					Replicas: 3,
				},
			}
			current := &dinfw.DistributedInferenceWorkloadStatus{
				DistributedInferenceStatus: wccommon.DistributedInferenceStatus{
					Phase:    PhaseReady,
					Replicas: 3,
				},
			}

			changed := statusUpdater.statusChanged(previous, current)

			Expect(changed).To(BeTrue())
		})

		It("should return true when replica counts are different", func() {
			previous := &dinfw.DistributedInferenceWorkloadStatus{
				DistributedInferenceStatus: wccommon.DistributedInferenceStatus{
					Phase:    PhaseReady,
					Replicas: 2,
				},
			}
			current := &dinfw.DistributedInferenceWorkloadStatus{
				DistributedInferenceStatus: wccommon.DistributedInferenceStatus{
					Phase:    PhaseReady,
					Replicas: 3,
				},
			}

			changed := statusUpdater.statusChanged(previous, current)

			Expect(changed).To(BeTrue())
		})

		It("should ignore LastStatusChangeTime differences", func() {
			now := metav1.Now()
			later := metav1.NewTime(now.Add(1 * time.Hour))

			previous := &dinfw.DistributedInferenceWorkloadStatus{
				DistributedInferenceStatus: wccommon.DistributedInferenceStatus{
					Phase:    PhaseReady,
					Replicas: 3,
				},
				WorkloadDebugStatus: wccommon.WorkloadDebugStatus{
					LastStatusChangeTime: &now,
				},
			}
			current := &dinfw.DistributedInferenceWorkloadStatus{
				DistributedInferenceStatus: wccommon.DistributedInferenceStatus{
					Phase:    PhaseReady,
					Replicas: 3,
				},
				WorkloadDebugStatus: wccommon.WorkloadDebugStatus{
					LastStatusChangeTime: &later,
				},
			}

			changed := statusUpdater.statusChanged(previous, current)

			Expect(changed).To(BeFalse())
		})

		It("should ignore SyncStatus differences", func() {
			now := metav1.Now()
			later := metav1.NewTime(now.Add(1 * time.Hour))

			previous := &dinfw.DistributedInferenceWorkloadStatus{
				DistributedInferenceStatus: wccommon.DistributedInferenceStatus{
					Phase:    PhaseReady,
					Replicas: 3,
				},
				SyncStatus: &wccommon.WorkloadSyncStatus{
					LastSyncTimestamp:  now,
					LastSyncGeneration: 1,
				},
			}
			current := &dinfw.DistributedInferenceWorkloadStatus{
				DistributedInferenceStatus: wccommon.DistributedInferenceStatus{
					Phase:    PhaseReady,
					Replicas: 3,
				},
				SyncStatus: &wccommon.WorkloadSyncStatus{
					LastSyncTimestamp:  later,
					LastSyncGeneration: 2,
				},
			}

			changed := statusUpdater.statusChanged(previous, current)

			Expect(changed).To(BeFalse())
		})
	})

	Describe("shared status input adapters", func() {
		It("converts a build error to a message", func() {
			buildErr := errors.New("build failed with error")

			message := buildErrorMessage(&buildErr)

			Expect(message).NotTo(BeNil())
			Expect(*message).To(Equal("build failed with error"))
		})

		It("formats missing resource identities", func() {
			missingService := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-service",
				},
				TypeMeta: metav1.TypeMeta{
					Kind: constants.KindService,
				},
			}
			missingResources := []client.Object{missingService}

			names := missingResourceNames(missingResources)

			Expect(names).To(Equal([]string{"Service/test-service"}))
		})

		It("preserves legacy missing resource order", func() {
			missingService := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-service",
				},
				TypeMeta: metav1.TypeMeta{
					Kind: constants.KindService,
				},
			}
			missingIngress := &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-ingress",
				},
				TypeMeta: metav1.TypeMeta{
					Kind: constants.KindIngress,
				},
			}
			missingResources := []client.Object{missingService, missingIngress}

			names := missingResourceNames(missingResources)

			Expect(names).To(Equal([]string{"Service/test-service", "Ingress/test-ingress"}))
		})

		It("returns nil for no build error", func() {
			Expect(buildErrorMessage(nil)).To(BeNil())
		})
	})

	Describe("Update", func() {
		BeforeEach(func() {
			// Create workload in k8s client
			err := k8sClient.Create(ctx, workload)
			Expect(err).NotTo(HaveOccurred())

			// Setup default mocks
			mockResourcesMgr.resources = distresources.DistributedResources{}
			mockResourcesMgr.missingResources = []client.Object{}
			mockResourcesMgr.err = nil
		})

		Context("when updateStatusFields succeeds", func() {
			It("should patch status when status has changed", func() {
				// Make status different from original
				originalWorkload.Status.Phase = ""
				workload.Status.Phase = PhaseCreating

				err := statusUpdater.Update(ctx, workload, originalWorkload, builtResources, nil)

				Expect(err).NotTo(HaveOccurred())

				// Verify workload was patched by fetching it
				fetchedWorkload := &dinfw.DistributedInferenceWorkload{}
				err = k8sClient.Get(ctx, types.NamespacedName{
					Name:      workload.Name,
					Namespace: workload.Namespace,
				}, fetchedWorkload)
				Expect(err).NotTo(HaveOccurred())
				Expect(fetchedWorkload.Status.Phase).To(Equal(PhaseCreating))
			})

			It("should skip patch when status hasn't changed", func() {
				// Make status identical to original
				originalWorkload.Status.Phase = PhaseCreating
				originalWorkload.Status.ObservedGeneration = 1

				err := statusUpdater.Update(ctx, workload, originalWorkload, builtResources, nil)

				Expect(err).NotTo(HaveOccurred())
			})

			It("should return error when status patch fails", func() {
				// Create a workload with invalid namespace to cause patch to fail
				invalidWorkload := workload.DeepCopy()
				invalidWorkload.Namespace = "non-existent-namespace"

				err := statusUpdater.Update(ctx, invalidWorkload, originalWorkload, builtResources, nil)

				Expect(err).To(HaveOccurred())
			})
		})

		Context("when updateStatusFields fails", func() {
			It("should return the error", func() {
				expectedErr := errors.New("update status fields error")
				mockResourcesMgr.err = expectedErr

				err := statusUpdater.Update(ctx, workload, originalWorkload, builtResources, nil)

				Expect(err).To(Equal(expectedErr))
			})
		})

		Context("integration scenarios", func() {
			var lws *lwsv1.LeaderWorkerSet

			BeforeEach(func() {
				lws = &lwsv1.LeaderWorkerSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-lws",
						Namespace: "test-namespace",
					},
					Status: lwsv1.LeaderWorkerSetStatus{
						Replicas:      3,
						ReadyReplicas: 2,
					},
				}

				mockResourcesMgr.resources = distresources.DistributedResources{
					Lws: lws,
				}
				mockLwsCalculator.phase = PhaseReady
				mockLwsCalculator.replicas = 3
				mockLwsCalculator.readyReplicas = 2
				mockLwsCalculator.err = nil
			})

			It("should successfully update workload status with all fields", func() {
				mockLwsCalculator.phaseMessage = "Number of replicas: 3; Ready: 2"

				err := statusUpdater.Update(ctx, workload, originalWorkload, builtResources, nil)

				Expect(err).NotTo(HaveOccurred())

				// Verify all status fields are set correctly
				Expect(workload.Status.Phase).To(Equal(PhaseReady))
				Expect(workload.Status.Replicas).To(Equal(int32(3)))
				Expect(workload.Status.ReadyReplicas).To(Equal(int32(2)))
				Expect(workload.Status.ObservedGeneration).To(Equal(workload.Generation))
				Expect(workload.Status.FirstAcknowledgedTime).NotTo(BeNil())
				Expect(workload.Status.LastStatusChangeTime).NotTo(BeNil())
				Expect(workload.Status.PhaseMessage).ToNot(BeEmpty())
			})

			It("should handle workload with build error", func() {
				mockLwsCalculator.phaseMessage = "Number of replicas: 3; Ready: 2"
				buildErr := errors.New("build failed")

				err := statusUpdater.Update(ctx, workload, originalWorkload, builtResources, &buildErr)

				Expect(err).NotTo(HaveOccurred())
				Expect(workload.Status.Phase).To(Equal(PhaseDegraded)) // Ready + build error = Degraded
				Expect(workload.Status.PhaseMessage).To(ContainSubstring("build failed"))
			})
		})
	})
})

func ptrTo[T any](v T) *T {
	return &v
}
