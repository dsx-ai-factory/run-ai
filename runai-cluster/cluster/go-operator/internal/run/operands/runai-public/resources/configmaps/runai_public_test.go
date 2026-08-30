package configmaps

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kc "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/kube-client"
	kcfake "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/kube-client/fake"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/global"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"

	fakediscovery "k8s.io/client-go/discovery/fake"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestLWSStatus(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "LWS Status Suite")
}

var _ = Describe("getLWSStatus", func() {
	var (
		ctx         context.Context
		kbClient    client.Client
		scheme      *runtime.Scheme
		runaiConfig *runv1.RunaiConfig
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme = runtime.NewScheme()
		Expect(appsv1.AddToScheme(scheme)).To(Succeed())
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(runv1.AddToScheme(scheme)).To(Succeed())
		kbClient = fake.NewClientBuilder().WithScheme(scheme).Build()

		// Create a default RunaiConfig with default LWS namespace
		runaiConfig = &runv1.RunaiConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "runai",
				Namespace: "runai",
			},
		}
		runaiConfig.SetDefaultsWhereNeeded()
	})

	Context("when LWS CRD is not installed", func() {
		It("should return unavailable status", func() {
			status := getLWSStatus(ctx, kbClient, runaiConfig)
			Expect(*status.Available).To(BeFalse())
			// The function checks deployment first, so we get deployment error, not CRD error
			Expect(status.Reason).To(ContainSubstring("Failed to find LeaderWorkerSet operator deployment"))
		})
	})

	Context("when LWS operator namespace is not found", func() {
		BeforeEach(func() {
			// Create a mock client that returns error when trying to find the namespace
			kbClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects().Build()
		})

		It("should return unavailable status", func() {
			status := getLWSStatus(ctx, kbClient, runaiConfig)
			Expect(*status.Available).To(BeFalse())
			Expect(status.Reason).To(ContainSubstring("Failed to find LeaderWorkerSet operator deployment"))
		})
	})

	Context("when LWS operator deployment is not found", func() {
		BeforeEach(func() {
			// Create a mock client with namespace but no deployment
			namespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "lws-system",
				},
			}
			kbClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(namespace).Build()
		})

		It("should return unavailable status", func() {
			status := getLWSStatus(ctx, kbClient, runaiConfig)
			Expect(*status.Available).To(BeFalse())
			Expect(status.Reason).To(ContainSubstring("Failed to find LeaderWorkerSet operator deployment"))
		})
	})

	Context("when LWS operator has no available replicas", func() {
		BeforeEach(func() {
			// Create a mock client with namespace and deployment but no available replicas
			namespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "lws-system",
				},
			}
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "lws-controller-manager",
					Namespace: "lws-system",
				},
				Status: appsv1.DeploymentStatus{
					AvailableReplicas: 0,
				},
			}
			kbClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(namespace, deployment).Build()
		})

		It("should return unavailable status", func() {
			status := getLWSStatus(ctx, kbClient, runaiConfig)
			Expect(*status.Available).To(BeFalse())
			Expect(status.Reason).To(Equal("LeaderWorkerSet operator has no available replicas"))
		})
	})

	Context("when LWS is properly installed", func() {
		BeforeEach(func() {
			// Create a mock client with namespace and deployment with available replicas
			namespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "lws-system",
				},
			}
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "lws-controller-manager",
					Namespace: "lws-system",
				},
				Status: appsv1.DeploymentStatus{
					AvailableReplicas: 1,
				},
			}
			kbClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(namespace, deployment).Build()
		})

		It("should return available status", func() {
			status := getLWSStatus(ctx, kbClient, runaiConfig)
			Expect(*status.Available).To(BeTrue())
			Expect(status.Reason).To(BeEmpty())
		})
	})

	Context("when LWS is installed in a custom namespace", func() {
		BeforeEach(func() {
			// Override the runaiConfig to use a custom LWS namespace
			customNamespace := "custom-lws-ns"
			runaiConfig.Spec.LWS.Namespace = ptr.To(customNamespace)

			// Create a mock client with custom namespace and deployment with available replicas
			namespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: customNamespace,
				},
			}
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "lws-controller-manager",
					Namespace: customNamespace,
				},
				Status: appsv1.DeploymentStatus{
					AvailableReplicas: 1,
				},
			}
			kbClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(namespace, deployment).Build()
		})

		It("should return available status from custom namespace", func() {
			status := getLWSStatus(ctx, kbClient, runaiConfig)
			Expect(*status.Available).To(BeTrue())
			Expect(status.Reason).To(BeEmpty())
		})
	})

	Context("when service account lacks permissions to create LWS resources", func() {
		BeforeEach(func() {
			// Create a mock client that returns a forbidden error when trying to create a resource
			// and returns the correct namespace when looking it up
			namespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "lws-system",
				},
			}
			kbClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(namespace).Build()
			// Mock the dryRunCreate to return a forbidden error
			kbClient = &mockClient{
				Client: kbClient,
				createFunc: func(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
					return &kerrors.StatusError{
						ErrStatus: metav1.Status{
							Status:  metav1.StatusFailure,
							Code:    http.StatusForbidden,
							Reason:  metav1.StatusReasonForbidden,
							Message: "leaderworkersets.leaderworkerset.x-k8s.io is forbidden: User \"system:serviceaccount:runai:runai-operator\" cannot create resource \"leaderworkersets\" in API group \"leaderworkerset.x-k8s.io\" in the namespace \"default\"",
						},
					}
				},
			}
		})

		It("should continue checking operator deployment when CRD exists but no permissions", func() {
			status := getLWSStatus(ctx, kbClient, runaiConfig)
			// Since we're mocking a forbidden error for CRD check and have the namespace but no operator deployment,
			// we should get an error about not finding the operator deployment
			Expect(*status.Available).To(BeFalse())
			Expect(status.Reason).To(ContainSubstring("Failed to find LeaderWorkerSet operator deployment"))
		})
	})
})

// mockClient is a custom client that allows mocking the Create method
type mockClient struct {
	client.Client
	createFunc func(ctx context.Context, obj client.Object, opts ...client.CreateOption) error
}

func (m *mockClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	return m.createFunc(ctx, obj, opts...)
}

var _ = Describe("getGatewayAPIStatus", func() {
	var (
		ctx             context.Context
		kbClient        kc.HybridClient
		scheme          *runtime.Scheme
		runaiConfig     *runv1.RunaiConfig
		discoveryClient *fakediscovery.FakeDiscovery
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme = runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(runv1.AddToScheme(scheme)).To(Succeed())
		Expect(gwv1.Install(scheme)).To(Succeed())

		fakeClientSet := fakeclientset.NewSimpleClientset()
		discoveryClient = &fakediscovery.FakeDiscovery{Fake: &fakeClientSet.Fake}

		runaiConfig = &runv1.RunaiConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "runai",
				Namespace: "runai",
			},
		}
		runaiConfig.SetDefaultsWhereNeeded()
	})

	Context("when gatewayAPI is disabled", func() {
		It("should return nil", func() {
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
			kbClient = kcfake.New(fakeClient, fakeClient, discoveryClient)
			result := getGatewayAPIStatus(ctx, kbClient, runaiConfig, discoveryClient)
			Expect(result).To(BeNil())
		})
	})

	Context("when gatewayAPI is enabled but HTTPRoute CRD not installed", func() {
		BeforeEach(func() {
			runaiConfig.Spec.Global.GatewayAPI = &global.GatewayAPI{
				Enabled:   ptr.To(true),
				Name:      "my-gateway",
				Namespace: "gateway-system",
			}
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
			kbClient = kcfake.New(fakeClient, fakeClient, discoveryClient)

			// API group exists but HTTPRoute kind is not listed
			discoveryClient.Resources = []*metav1.APIResourceList{{
				GroupVersion: "gateway.networking.k8s.io/v1",
				APIResources: []metav1.APIResource{
					{Kind: "Gateway"},
				},
			}}
		})

		It("should return unavailable with CRD reason", func() {
			result := getGatewayAPIStatus(ctx, kbClient, runaiConfig, discoveryClient)
			Expect(result).NotTo(BeNil())
			Expect(*result.Available).To(BeFalse())

			crdStatus := result.Components["httproute-crd"]
			Expect(*crdStatus.Available).To(BeFalse())
			Expect(crdStatus.Reason).To(ContainSubstring("Gateway API CRDs not installed"))

			gwStatus := result.Components["gateway"]
			Expect(*gwStatus.Available).To(BeFalse())
			Expect(gwStatus.Reason).To(ContainSubstring("Gateway API CRDs not installed"))
		})
	})

	Context("when CRD is installed but Gateway resource not found", func() {
		BeforeEach(func() {
			runaiConfig.Spec.Global.GatewayAPI = &global.GatewayAPI{
				Enabled:   ptr.To(true),
				Name:      "my-gateway",
				Namespace: "gateway-system",
			}
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
			kbClient = kcfake.New(fakeClient, fakeClient, discoveryClient)

			discoveryClient.Resources = []*metav1.APIResourceList{{
				GroupVersion: "gateway.networking.k8s.io/v1",
				APIResources: []metav1.APIResource{
					{Kind: "Gateway"},
					{Kind: "HTTPRoute"},
				},
			}}
		})

		It("should return CRD available but gateway unavailable", func() {
			result := getGatewayAPIStatus(ctx, kbClient, runaiConfig, discoveryClient)
			Expect(result).NotTo(BeNil())
			Expect(*result.Available).To(BeFalse())

			crdStatus := result.Components["httproute-crd"]
			Expect(*crdStatus.Available).To(BeTrue())

			gwStatus := result.Components["gateway"]
			Expect(*gwStatus.Available).To(BeFalse())
			Expect(gwStatus.Reason).To(ContainSubstring("Gateway gateway-system/my-gateway not found"))
		})
	})

	Context("when CRD is installed and Gateway resource exists", func() {
		BeforeEach(func() {
			runaiConfig.Spec.Global.GatewayAPI = &global.GatewayAPI{
				Enabled:   ptr.To(true),
				Name:      "my-gateway",
				Namespace: "gateway-system",
			}

			gateway := &gwv1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-gateway",
					Namespace: "gateway-system",
				},
			}
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(gateway).Build()
			kbClient = kcfake.New(fakeClient, fakeClient, discoveryClient)

			discoveryClient.Resources = []*metav1.APIResourceList{{
				GroupVersion: "gateway.networking.k8s.io/v1",
				APIResources: []metav1.APIResource{
					{Kind: "Gateway"},
					{Kind: "HTTPRoute"},
				},
			}}
		})

		It("should return fully available", func() {
			result := getGatewayAPIStatus(ctx, kbClient, runaiConfig, discoveryClient)
			Expect(result).NotTo(BeNil())
			Expect(*result.Available).To(BeTrue())

			crdStatus := result.Components["httproute-crd"]
			Expect(*crdStatus.Available).To(BeTrue())

			gwStatus := result.Components["gateway"]
			Expect(*gwStatus.Available).To(BeTrue())
		})
	})

	Context("when discovery returns an error", func() {
		BeforeEach(func() {
			runaiConfig.Spec.Global.GatewayAPI = &global.GatewayAPI{
				Enabled:   ptr.To(true),
				Name:      "my-gateway",
				Namespace: "gateway-system",
			}
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
			kbClient = kcfake.New(fakeClient, fakeClient, discoveryClient)

			// Inject a non-NotFound error via reactor to simulate discovery failure
			discoveryClient.PrependReactor("*", "*", func(action k8stesting.Action) (bool, runtime.Object, error) {
				return true, nil, fmt.Errorf("discovery unavailable")
			})
		})

		It("should return unavailable with CRD check failure reason", func() {
			result := getGatewayAPIStatus(ctx, kbClient, runaiConfig, discoveryClient)
			Expect(result).NotTo(BeNil())
			Expect(*result.Available).To(BeFalse())

			crdStatus := result.Components["httproute-crd"]
			Expect(*crdStatus.Available).To(BeFalse())
			Expect(crdStatus.Reason).To(ContainSubstring("Gateway API CRD check failed"))
		})
	})

	Context("when Gateway Get returns a non-NotFound error", func() {
		BeforeEach(func() {
			runaiConfig.Spec.Global.GatewayAPI = &global.GatewayAPI{
				Enabled:   ptr.To(true),
				Name:      "my-gateway",
				Namespace: "gateway-system",
			}

			fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithInterceptorFuncs(interceptor.Funcs{
				Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
					if _, ok := obj.(*gwv1.Gateway); ok {
						return kerrors.NewForbidden(schema.GroupResource{Group: "gateway.networking.k8s.io", Resource: "gateways"}, key.Name, fmt.Errorf("access denied"))
					}
					return c.Get(ctx, key, obj, opts...)
				},
			}).Build()
			kbClient = kcfake.New(fakeClient, fakeClient, discoveryClient)

			discoveryClient.Resources = []*metav1.APIResourceList{{
				GroupVersion: "gateway.networking.k8s.io/v1",
				APIResources: []metav1.APIResource{
					{Kind: "Gateway"},
					{Kind: "HTTPRoute"},
				},
			}}
		})

		It("should return unavailable with error details", func() {
			result := getGatewayAPIStatus(ctx, kbClient, runaiConfig, discoveryClient)
			Expect(result).NotTo(BeNil())
			Expect(*result.Available).To(BeFalse())

			crdStatus := result.Components["httproute-crd"]
			Expect(*crdStatus.Available).To(BeTrue())

			gwStatus := result.Components["gateway"]
			Expect(*gwStatus.Available).To(BeFalse())
			Expect(gwStatus.Reason).To(ContainSubstring("failed to get Gateway"))
			Expect(gwStatus.Reason).To(ContainSubstring("forbidden"))
		})
	})
})
