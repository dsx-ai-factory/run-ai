package builders

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	openshiftv1 "github.com/openshift/api/route/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-common/config"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-common/project"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"knative.dev/pkg/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

var _ = Describe("ExternalUrl tests", func() {
	var (
		scheme                       *runtime.Scheme
		fakeClient                   client.Client
		eb                           *ExternalUrlBuilder
		ebOpenshift                  *ExternalUrlBuilder
		ctx                          context.Context
		workloadKind                 string
		jobType                      string
		workloadInitials             string
		workloadBaseName             string
		objectMeta                   *metav1.ObjectMeta
		owner                        *corev1.Pod
		exposedUrls                  *flds.ExposedUrlsField
		ingressUrl                   *flds.StringField
		calculatedPriorityClassName  string
		isJupyter                    bool
		isTensorboard                bool
		isAuthDisabled               bool
		projectName                  string

		priorityClassNames = []string{constants.InteractivePreemptiblePriorityClass, constants.InteractivePriorityClass, "train"}
	)

	BeforeEach(func() {
		// Setup common test variables
		scheme = runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(networkingv1.AddToScheme(scheme)).To(Succeed())
		Expect(openshiftv1.AddToScheme(scheme)).To(Succeed())
		Expect(schedulingv1.AddToScheme(scheme)).To(Succeed())
		Expect(gatewayv1.Install(scheme)).To(Succeed())

		fakeClient = fake.NewClientBuilder().WithScheme(scheme).Build()

		eb = NewExternalUrlBuilder(fakeClient, false)
		ebOpenshift = NewExternalUrlBuilder(fakeClient, true)

		ctx = context.Background()
		projectName = "test-project"
		ctx = project.SetProjectInContext(ctx, projectName)

		workloadKind = constants.KindInteractiveWorkload
		jobType = "interactive"
		workloadInitials = "iw"
		workloadBaseName = "test-workload"

		objectMeta = &metav1.ObjectMeta{
			Name:      workloadBaseName,
			Namespace: "default",
			Labels:    map[string]string{},
		}

		owner = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "default",
				UID:       "test-uid",
			},
		}

		exposedUrls = &flds.ExposedUrlsField{
			Value: &map[string]flds.ExposedUrlItemAndRules{},
		}

		ingressUrl = &flds.StringField{
			Value: ptr.To("http://example.com"),
		}

		calculatedPriorityClassName = ""
		isJupyter = false
		isTensorboard = false
		isAuthDisabled = true

		// Set default cluster domain for URL composition
		_ = os.Setenv(config.EnvVarClusterDomain, "http://cluster.example.com")
		_ = os.Setenv(config.EnvVarSubdomainSupport, "true")

		createPriorityClasses(fakeClient, priorityClassNames)
	})

	AfterEach(func() {
		_ = os.Unsetenv(config.EnvVarClusterDomain)
		_ = os.Unsetenv(config.EnvVarSubdomainSupport)
		_ = os.Unsetenv(config.EnvVarExternalAuthUrlEnabled)
		_ = os.Unsetenv(config.EnvVarOpenshift)
		_ = os.Unsetenv(config.EnvVarIngressClass)
		_ = os.Unsetenv(config.EnvVarIngressEnabled)
		_ = os.Unsetenv(config.EnvVarGatewayAPIEnabled)
		_ = os.Unsetenv(config.EnvVarGatewayName)
		_ = os.Unsetenv(config.EnvVarGatewayNamespace)

		deletePriorityClasses(fakeClient, priorityClassNames)
	})

	Context("GetPreview and Build", func() {
		It("should return empty result for no exposed URLs, no Jupyter/Tensorboard", func() {
			// Empty exposedUrls
			exposedUrls = &flds.ExposedUrlsField{
				Value: &map[string]flds.ExposedUrlItemAndRules{},
			}

			// Disable Jupyter and Tensorboard
			isJupyter = false
			isTensorboard = false

			// Verify GetPreview
			preview, err := eb.GetPreview(ctx, workloadInitials, objectMeta, exposedUrls, isJupyter, isTensorboard)
			Expect(err).NotTo(HaveOccurred())
			Expect(preview).To(BeEmpty())

			// Verify Build
			services, ingresses, routes, _, err := eb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadBaseName, owner, exposedUrls, ingressUrl, calculatedPriorityClassName, isJupyter, isTensorboard, isAuthDisabled)
			Expect(err).NotTo(HaveOccurred())
			Expect(services).To(BeEmpty())
			Expect(ingresses).To(BeEmpty())
			Expect(routes).To(BeEmpty())
		})

		Context("when handling exposed URLs", func() {
			It("should handle a single auto-composed exposed URL", func() {
				// Setup exposed URL with container port
				(*exposedUrls.Value)["main"] = flds.ExposedUrlItemAndRules{
					Value: &flds.ExposedUrlItemValue{
						ContainerPort: ptr.To(int32(8080)),
					},
				}

				// Verify GetPreview
				preview, err := eb.GetPreview(ctx, workloadInitials, objectMeta, exposedUrls, isJupyter, isTensorboard)
				Expect(err).NotTo(HaveOccurred())
				Expect(preview).To(HaveLen(2)) // service + ingress

				// Check expected service in preview
				foundService := false
				foundIngress := false
				for _, res := range preview {
					if res.Kind.Kind == constants.KindService {
						foundService = true
						Expect(res.Name).To(MatchRegexp("iw-test-workload-\\d+"))
					} else if res.Kind.Kind == constants.KindIngress {
						foundIngress = true
						Expect(res.Name).To(MatchRegexp("iw-test-workload-\\d+"))
					}
				}
				Expect(foundService).To(BeTrue())
				Expect(foundIngress).To(BeTrue())

				// Verify Build
				services, ingresses, routes, _, err := eb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadBaseName, owner, exposedUrls, ingressUrl, calculatedPriorityClassName, isJupyter, isTensorboard, isAuthDisabled)
				Expect(err).NotTo(HaveOccurred())
				Expect(services).To(HaveLen(1))
				Expect(ingresses).To(HaveLen(1))
				Expect(routes).To(BeEmpty())

				// Check service properties
				service := services[0]
				Expect(service.Name).To(MatchRegexp("iw-test-workload-\\d+"))
				Expect(service.Spec.Ports).To(HaveLen(1))
				Expect(service.Spec.Ports[0].TargetPort.IntVal).To(Equal(int32(8080)))

				// Check ingress properties
				ingress := ingresses[0]
				Expect(ingress.Name).To(Equal(service.Name))
				Expect(ingress.Spec.Rules).To(HaveLen(1))
				Expect(ingress.Spec.Rules[0].HTTP.Paths).To(HaveLen(1))

				// Check annotation for job URL was added
				Expect(objectMeta.Annotations).ToNot(BeNil())
				Expect(objectMeta.Annotations[constants.AnnotationJobUrl]).ToNot(BeEmpty())
			})

			It("should handle a single auto-composed exposed URL on OpenShift", func() {
				// Setup exposed URL with container port
				(*exposedUrls.Value)["main"] = flds.ExposedUrlItemAndRules{
					Value: &flds.ExposedUrlItemValue{
						ContainerPort: ptr.To(int32(8080)),
					},
				}

				// Verify GetPreview
				preview, err := ebOpenshift.GetPreview(ctx, workloadInitials, objectMeta, exposedUrls, isJupyter, isTensorboard)
				Expect(err).NotTo(HaveOccurred())
				Expect(preview).To(HaveLen(2)) // service + route

				// Check expected service and route in preview
				foundService := false
				foundRoute := false
				for _, res := range preview {
					if res.Kind.Kind == constants.KindService {
						foundService = true
						Expect(res.Name).To(MatchRegexp("iw-test-workload-\\d+"))
					} else if res.Kind.Kind == RouteGroupVersion.Kind {
						foundRoute = true
						Expect(res.Name).To(MatchRegexp("iw-test-workload-\\d+"))
					}
				}
				Expect(foundService).To(BeTrue())
				Expect(foundRoute).To(BeTrue())

				// Verify Build
				services, ingresses, routes, _, err := ebOpenshift.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadBaseName, owner, exposedUrls, ingressUrl, calculatedPriorityClassName, isJupyter, isTensorboard, isAuthDisabled)
				Expect(err).NotTo(HaveOccurred())
				Expect(services).To(HaveLen(1))
				Expect(ingresses).To(BeEmpty())
				Expect(routes).To(HaveLen(1))

				// Check service properties
				service := services[0]
				Expect(service.Name).To(MatchRegexp("iw-test-workload-\\d+"))
				Expect(service.Spec.Ports).To(HaveLen(1))
				Expect(service.Spec.Ports[0].TargetPort.IntVal).To(Equal(int32(8080)))

				// Check route properties
				route := routes[0]
				Expect(route.Name).To(Equal(service.Name))
				Expect(route.Spec.To.Kind).To(Equal(constants.KindService))
				Expect(route.Spec.To.Name).To(Equal(service.Name))

				// Check annotation for job URL was added
				Expect(objectMeta.Annotations).ToNot(BeNil())
				Expect(objectMeta.Annotations[constants.AnnotationJobUrl]).ToNot(BeEmpty())
			})

			It("should handle multiple auto-composed exposed URLs (first no suffix, then with nick)", func() {
				// Setup multiple exposed URLs with container ports
				(*exposedUrls.Value)["first"] = flds.ExposedUrlItemAndRules{
					Value: &flds.ExposedUrlItemValue{
						ContainerPort: ptr.To(int32(8080)),
					},
				}
				(*exposedUrls.Value)["second"] = flds.ExposedUrlItemAndRules{
					Value: &flds.ExposedUrlItemValue{
						ContainerPort: ptr.To(int32(9090)),
					},
				}

				// Verify GetPreview
				preview, err := eb.GetPreview(ctx, workloadInitials, objectMeta, exposedUrls, isJupyter, isTensorboard)
				Expect(err).NotTo(HaveOccurred())
				Expect(preview).To(HaveLen(4)) // 2 services + 2 ingresses

				// Verify Build
				services, ingresses, routes, _, err := eb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadBaseName, owner, exposedUrls, ingressUrl, calculatedPriorityClassName, isJupyter, isTensorboard, isAuthDisabled)
				Expect(err).NotTo(HaveOccurred())
				Expect(services).To(HaveLen(2))
				Expect(ingresses).To(HaveLen(2))
				Expect(routes).To(BeEmpty())

				// First service should have no suffix in URL
				firstURL := objectMeta.Annotations[constants.AnnotationJobUrl]
				Expect(firstURL).To(ContainSubstring("test-project-test-workload.cluster.example.com"))

				// Second should have suffix
				Expect(objectMeta.Annotations[constants.AnnotationJobUrl]).To(ContainSubstring("test-project-test-workload-second.cluster.example.com"))
			})

			It("should handle a single exposed URL with a valid custom URL", func() {
				customUrl := "http://custom.example.com/path"
				// Setup exposed URL with custom URL
				(*exposedUrls.Value)["custom"] = flds.ExposedUrlItemAndRules{
					Value: &flds.ExposedUrlItemValue{
						ContainerPort: ptr.To(int32(8080)),
						CustomUrl:     ptr.To(true),
						Url:           ptr.To(customUrl),
					},
				}

				// Verify GetPreview
				preview, err := eb.GetPreview(ctx, workloadInitials, objectMeta, exposedUrls, isJupyter, isTensorboard)
				Expect(err).NotTo(HaveOccurred())
				Expect(preview).To(HaveLen(2)) // service + ingress

				// Verify Build
				services, ingresses, _, _, err := eb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadBaseName, owner, exposedUrls, ingressUrl, calculatedPriorityClassName, isJupyter, isTensorboard, isAuthDisabled)
				Expect(err).NotTo(HaveOccurred())
				Expect(services).To(HaveLen(1))
				Expect(ingresses).To(HaveLen(1))

				// Check service properties
				service := services[0]
				Expect(service.Spec.Ports).To(HaveLen(1))
				Expect(service.Spec.Ports[0].TargetPort.IntVal).To(Equal(int32(8080)))

				// Check ingress properties - should use custom host
				ingress := ingresses[0]
				Expect(ingress.Spec.Rules[0].Host).To(Equal("custom.example.com"))
				Expect(ingress.Spec.Rules[0].HTTP.Paths[0].Path).To(Equal("/path"))

				// Check annotation for job URL was added
				Expect(objectMeta.Annotations[constants.AnnotationJobUrl]).To(ContainSubstring(customUrl))
			})

			It("should skip exposed URL if its value is nil", func() {
				// Setup an exposed URL with nil value
				(*exposedUrls.Value)["nil-value"] = flds.ExposedUrlItemAndRules{
					Value: nil,
				}

				// Add a valid URL for comparison
				(*exposedUrls.Value)["valid"] = flds.ExposedUrlItemAndRules{
					Value: &flds.ExposedUrlItemValue{
						ContainerPort: ptr.To(int32(8080)),
					},
				}

				// Verify GetPreview - should only show resources for the valid URL
				preview, err := eb.GetPreview(ctx, workloadInitials, objectMeta, exposedUrls, isJupyter, isTensorboard)
				Expect(err).NotTo(HaveOccurred())
				Expect(preview).To(HaveLen(2)) // 1 service + 1 ingress for valid URL

				// Verify Build - should only create resources for the valid URL
				services, ingresses, _, _, err := eb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadBaseName, owner, exposedUrls, ingressUrl, calculatedPriorityClassName, isJupyter, isTensorboard, isAuthDisabled)
				Expect(err).NotTo(HaveOccurred())
				Expect(services).To(HaveLen(1))
				Expect(ingresses).To(HaveLen(1))
			})

			It("should handle exposed URL with CustomUrl=true and no URL (error for Build, skip for Preview)", func() {
				// Setup exposed URL with CustomUrl=true but no URL
				(*exposedUrls.Value)["invalid"] = flds.ExposedUrlItemAndRules{
					Value: &flds.ExposedUrlItemValue{
						ContainerPort: ptr.To(int32(8080)),
						CustomUrl:     ptr.To(true),
						// No URL provided
					},
				}

				// GetPreview should skip invalid URLs without error
				preview, err := eb.GetPreview(ctx, workloadInitials, objectMeta, exposedUrls, isJupyter, isTensorboard)
				Expect(err).NotTo(HaveOccurred())
				Expect(preview).To(BeEmpty())

				// Build should return an error
				_, _, _, _, err = eb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadBaseName, owner, exposedUrls, ingressUrl, calculatedPriorityClassName, isJupyter, isTensorboard, isAuthDisabled)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("url is mandatory when customUrl is enabled"))
			})

			It("should handle exposed URL with CustomUrl=false and URL provided (error for Build, skip for Preview)", func() {
				// Setup exposed URL with CustomUrl=false but URL provided
				(*exposedUrls.Value)["invalid"] = flds.ExposedUrlItemAndRules{
					Value: &flds.ExposedUrlItemValue{
						ContainerPort: ptr.To(int32(8080)),
						CustomUrl:     ptr.To(false),
						Url:           ptr.To("http://example.com"),
					},
				}

				// GetPreview should skip invalid URLs without error
				preview, err := eb.GetPreview(ctx, workloadInitials, objectMeta, exposedUrls, isJupyter, isTensorboard)
				Expect(err).NotTo(HaveOccurred())
				Expect(preview).To(BeEmpty())

				// Build should return an error
				_, _, _, _, err = eb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadBaseName, owner, exposedUrls, ingressUrl, calculatedPriorityClassName, isJupyter, isTensorboard, isAuthDisabled)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("url should only be provided for customUrl"))
			})

			It("should return error if exposed URL has no container port", func() {
				// Setup exposed URL without container port
				(*exposedUrls.Value)["missing-port"] = flds.ExposedUrlItemAndRules{
					Value: &flds.ExposedUrlItemValue{
						// No container port
					},
				}

				// GetPreview should work (it doesn't do full validation)
				preview, err := eb.GetPreview(ctx, workloadInitials, objectMeta, exposedUrls, isJupyter, isTensorboard)
				Expect(err).NotTo(HaveOccurred())
				Expect(preview).To(HaveLen(2)) // service + ingress would be created

				// Build should return an error
				_, _, _, _, err = eb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadBaseName, owner, exposedUrls, ingressUrl, calculatedPriorityClassName, isJupyter, isTensorboard, isAuthDisabled)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("container port is mandatory for exposed url"))
			})

			It("should handle error during exposed URL composition", func() {
				// Cause URL composition error by setting invalid cluster domain
				_ = os.Setenv(config.EnvVarClusterDomain, "://invalid-domain")

				// Setup a valid exposed URL
				(*exposedUrls.Value)["main"] = flds.ExposedUrlItemAndRules{
					Value: &flds.ExposedUrlItemValue{
						ContainerPort: ptr.To(int32(8080)),
					},
				}

				// Build should fail
				_, _, _, _, err := eb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadBaseName, owner, exposedUrls, ingressUrl, calculatedPriorityClassName, isJupyter, isTensorboard, isAuthDisabled)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid cluster domain"))
			})

			It("should handle error during service creation for exposed URL", func() {
				// This is hard to test directly since service creation doesn't have many paths that return errors
				// We can test invalid URLs or port parsing errors
				(*exposedUrls.Value)["invalid-port"] = flds.ExposedUrlItemAndRules{
					Value: &flds.ExposedUrlItemValue{
						ContainerPort: ptr.To(int32(8080)),
						CustomUrl:     ptr.To(true),
						Url:           ptr.To("http://invalid:port"),
					},
				}

				// Build should fail with port parsing error
				_, _, _, _, err := eb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadBaseName, owner, exposedUrls, ingressUrl, calculatedPriorityClassName, isJupyter, isTensorboard, isAuthDisabled)
				Expect(err).To(HaveOccurred())
				// Note: This actually fails at URL parsing, not service creation directly
			})

			It("should handle error during ingress/route creation for exposed URL", func() {
				// Similar to the service creation error, this is hard to test directly
				// We can use an invalid URL that passes initial validation but fails during ingress/route creation
				(*exposedUrls.Value)["invalid-host"] = flds.ExposedUrlItemAndRules{
					Value: &flds.ExposedUrlItemValue{
						ContainerPort: ptr.To(int32(8080)),
						CustomUrl:     ptr.To(true),
						Url:           ptr.To("http://   invalid-host"),
					},
				}

				// Build should fail with URL parsing/validation error
				_, _, _, _, err := eb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadBaseName, owner, exposedUrls, ingressUrl, calculatedPriorityClassName, isJupyter, isTensorboard, isAuthDisabled)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when handling Jupyter or Tensorboard specific logic", func() {
			It("should correctly create resources for Jupyter when enabled and ingressUrl is provided (Kubernetes)", func() {
				// Setup Jupyter
				isJupyter = true
				isTensorboard = false

				// Verify GetPreview
				preview, err := eb.GetPreview(ctx, workloadInitials, objectMeta, exposedUrls, isJupyter, isTensorboard)
				Expect(err).NotTo(HaveOccurred())
				Expect(preview).To(HaveLen(2)) // service + ingress

				// Check expected service in preview
				foundService := false
				foundIngress := false
				for _, res := range preview {
					if res.Kind.Kind == constants.KindService {
						foundService = true
						Expect(res.Name).To(Equal("iw-test-workload-jupyter"))
					} else if res.Kind.Kind == constants.KindIngress {
						foundIngress = true
						Expect(res.Name).To(Equal("iw-test-workload-jupyter"))
					}
				}
				Expect(foundService).To(BeTrue())
				Expect(foundIngress).To(BeTrue())

				// Verify Build
				services, ingresses, routes, _, err := eb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadBaseName, owner, exposedUrls, ingressUrl, calculatedPriorityClassName, isJupyter, isTensorboard, isAuthDisabled)
				Expect(err).NotTo(HaveOccurred())
				Expect(services).To(HaveLen(1))
				Expect(ingresses).To(HaveLen(1))
				Expect(routes).To(BeEmpty())

				// Check service properties
				service := services[0]
				Expect(service.Name).To(Equal("iw-test-workload-jupyter"))
				Expect(service.Spec.Ports).To(HaveLen(1))
				Expect(service.Spec.Ports[0].TargetPort.IntVal).To(Equal(int32(DefaultJupyterAndTensorContainerPort)))

				// Check ingress properties
				ingress := ingresses[0]
				Expect(ingress.Name).To(Equal(service.Name))

				// Check the Jupyter label was added
				Expect(objectMeta.Labels[constants.LabelJupyter]).To(Equal("true"))
			})

			It("should correctly create resources for Tensorboard when enabled and ingressUrl is provided (OpenShift)", func() {
				// Setup Tensorboard
				isJupyter = false
				isTensorboard = true

				// Verify GetPreview
				preview, err := ebOpenshift.GetPreview(ctx, workloadInitials, objectMeta, exposedUrls, isJupyter, isTensorboard)
				Expect(err).NotTo(HaveOccurred())
				Expect(preview).To(HaveLen(2)) // service + route

				// Check expected service in preview
				foundService := false
				foundRoute := false
				for _, res := range preview {
					if res.Kind.Kind == constants.KindService {
						foundService = true
						Expect(res.Name).To(Equal("iw-test-workload-tensorboard"))
					} else if res.Kind.Kind == RouteGroupVersion.Kind {
						foundRoute = true
						Expect(res.Name).To(Equal("iw-test-workload-tensorboard"))
					}
				}
				Expect(foundService).To(BeTrue())
				Expect(foundRoute).To(BeTrue())

				// Verify Build
				services, ingresses, routes, _, err := ebOpenshift.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadBaseName, owner, exposedUrls, ingressUrl, calculatedPriorityClassName, isJupyter, isTensorboard, isAuthDisabled)
				Expect(err).NotTo(HaveOccurred())
				Expect(services).To(HaveLen(1))
				Expect(ingresses).To(BeEmpty())
				Expect(routes).To(HaveLen(1))

				// Check service properties
				service := services[0]
				Expect(service.Name).To(Equal("iw-test-workload-tensorboard"))
				Expect(service.Spec.Ports).To(HaveLen(1))
				Expect(service.Spec.Ports[0].TargetPort.IntVal).To(Equal(int32(DefaultJupyterAndTensorContainerPort)))

				// Check route properties
				route := routes[0]
				Expect(route.Name).To(Equal(service.Name))

				// Check the Tensorboard label was added
				Expect(objectMeta.Labels[constants.LabelTensorboard]).To(Equal("true"))
			})

			It("should not create Jupyter/Tensorboard specific resources if ingressUrl is nil", func() {
				// Setup Jupyter but with nil ingressUrl
				isJupyter = true
				ingressUrl = nil

				// Verify GetPreview - should still show resources (it doesn't check ingressUrl)
				preview, err := eb.GetPreview(ctx, workloadInitials, objectMeta, exposedUrls, isJupyter, isTensorboard)
				Expect(err).NotTo(HaveOccurred())
				Expect(preview).To(HaveLen(2)) // service + ingress are in preview

				// Verify Build - should not create resources
				services, ingresses, routes, _, err := eb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadBaseName, owner, exposedUrls, ingressUrl, calculatedPriorityClassName, isJupyter, isTensorboard, isAuthDisabled)
				Expect(err).NotTo(HaveOccurred())
				Expect(services).To(BeEmpty())
				Expect(ingresses).To(BeEmpty())
				Expect(routes).To(BeEmpty())

				// The Jupyter label should still be set
				Expect(objectMeta.Labels[constants.LabelJupyter]).To(Equal("true"))
			})

			It("should set Jupyter/Tensorboard labels correctly", func() {
				// Test setting Jupyter label
				isJupyter = true
				isTensorboard = false

				_, _, _, _, err := eb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadBaseName, owner, exposedUrls, ingressUrl, calculatedPriorityClassName, isJupyter, isTensorboard, isAuthDisabled)
				Expect(err).NotTo(HaveOccurred())
				Expect(objectMeta.Labels[constants.LabelJupyter]).To(Equal("true"))
				Expect(objectMeta.Labels).NotTo(HaveKey(constants.LabelTensorboard))

				// Reset labels and test Tensorboard
				objectMeta.Labels = map[string]string{}
				isJupyter = false
				isTensorboard = true

				_, _, _, _, err = eb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadBaseName, owner, exposedUrls, ingressUrl, calculatedPriorityClassName, isJupyter, isTensorboard, isAuthDisabled)
				Expect(err).NotTo(HaveOccurred())
				Expect(objectMeta.Labels[constants.LabelTensorboard]).To(Equal("true"))
				Expect(objectMeta.Labels).NotTo(HaveKey(constants.LabelJupyter))
			})

			It("should handle error during URL composition for Jupyter/Tensorboard service", func() {
				// Cause URL composition error by setting invalid cluster domain
				_ = os.Setenv(config.EnvVarClusterDomain, "://invalid-domain")

				isJupyter = true

				// Build should fail with URL composition error
				_, _, _, _, err := eb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadBaseName, owner, exposedUrls, ingressUrl, calculatedPriorityClassName, isJupyter, isTensorboard, isAuthDisabled)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to compose job extneral url: invalid cluster domain"))
			})
		})

		Context("when handling combined scenarios", func() {
			It("should handle both exposed URLs and Jupyter enabled", func() {
				// Setup Jupyter
				isJupyter = true

				// Setup exposed URL
				(*exposedUrls.Value)["main"] = flds.ExposedUrlItemAndRules{
					Value: &flds.ExposedUrlItemValue{
						ContainerPort: ptr.To(int32(8080)),
					},
				}

				// Verify GetPreview - should have services and ingresses for both
				preview, err := eb.GetPreview(ctx, workloadInitials, objectMeta, exposedUrls, isJupyter, isTensorboard)
				Expect(err).NotTo(HaveOccurred())
				Expect(preview).To(HaveLen(4)) // 2 services + 2 ingresses

				// Verify Build - should create resources for both
				services, ingresses, routes, _, err := eb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadBaseName, owner, exposedUrls, ingressUrl, calculatedPriorityClassName, isJupyter, isTensorboard, isAuthDisabled)
				Expect(err).NotTo(HaveOccurred())
				Expect(services).To(HaveLen(2))  // Jupyter service + exposed URL service
				Expect(ingresses).To(HaveLen(2)) // Jupyter ingress + exposed URL ingress
				Expect(routes).To(BeEmpty())

				// Check that Jupyter service exists
				foundJupyter := false
				foundExposed := false
				for _, svc := range services {
					if svc.Name == "iw-test-workload-jupyter" {
						foundJupyter = true
					} else {
						foundExposed = true
					}
				}
				Expect(foundJupyter).To(BeTrue())
				Expect(foundExposed).To(BeTrue())

				// Verify URLs in annotations
				Expect(objectMeta.Annotations[constants.AnnotationJobUrl]).To(ContainSubstring("test-project-test-workload.cluster.example.com"))
				// If Jupyter is enabled, exposed URLs should have suffixes
				Expect(objectMeta.Annotations[constants.AnnotationJobUrl]).To(ContainSubstring("test-project-test-workload-main.cluster.example.com"))
			})
		})

		Context("when handling owner references and labels", func() {
			It("should set controller reference if owner is provided (for Build)", func() {
				// Setup basic exposed URL for testing
				(*exposedUrls.Value)["main"] = flds.ExposedUrlItemAndRules{
					Value: &flds.ExposedUrlItemValue{
						ContainerPort: ptr.To(int32(8080)),
					},
				}

				// Verify Build with owner reference
				services, ingresses, _, _, err := eb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadBaseName, owner, exposedUrls, ingressUrl, calculatedPriorityClassName, isJupyter, isTensorboard, isAuthDisabled)
				Expect(err).NotTo(HaveOccurred())
				Expect(services).To(HaveLen(1))
				Expect(ingresses).To(HaveLen(1))

				// Check owner references were set
				service := services[0]
				Expect(service.OwnerReferences).To(HaveLen(1))
				Expect(service.OwnerReferences[0].Name).To(Equal(owner.Name))
				Expect(service.OwnerReferences[0].UID).To(Equal(owner.UID))

				ingress := ingresses[0]
				Expect(ingress.OwnerReferences).To(HaveLen(1))
				Expect(ingress.OwnerReferences[0].Name).To(Equal(owner.Name))
				Expect(ingress.OwnerReferences[0].UID).To(Equal(owner.UID))
			})

			It("should set priority class label if setPriorityClassLabel is true", func() {
				// Enable priority class label
				calculatedPriorityClassName = constants.InteractivePreemptiblePriorityClass

				// Setup basic exposed URL
				(*exposedUrls.Value)["main"] = flds.ExposedUrlItemAndRules{
					Value: &flds.ExposedUrlItemValue{
						ContainerPort: ptr.To(int32(8080)),
					},
				}

				// Verify Build sets priority class label
				_, _, _, _, err := eb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadBaseName, owner, exposedUrls, ingressUrl, calculatedPriorityClassName, isJupyter, isTensorboard, isAuthDisabled)
				Expect(err).NotTo(HaveOccurred())

				// Check priority class label was set
				Expect(objectMeta.Labels[constants.LabelPriorityClassName]).To(Equal(constants.InteractivePreemptiblePriorityClass))

				// Reset and test non-preemptible
				objectMeta.Labels = map[string]string{}
				calculatedPriorityClassName = constants.InteractivePriorityClass

				_, _, _, _, err = eb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadBaseName, owner, exposedUrls, ingressUrl, calculatedPriorityClassName, isJupyter, isTensorboard, isAuthDisabled)
				Expect(err).NotTo(HaveOccurred())

				Expect(objectMeta.Labels[constants.LabelPriorityClassName]).To(Equal(constants.InteractivePriorityClass))

				// Reset and test with priority ClassName set
				objectMeta.Labels = map[string]string{}
				calculatedPriorityClassName = "train"

				_, _, _, _, err = eb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadBaseName, owner, exposedUrls, ingressUrl, calculatedPriorityClassName, isJupyter, isTensorboard, isAuthDisabled)
				Expect(err).NotTo(HaveOccurred())

				Expect(objectMeta.Labels[constants.LabelPriorityClassName]).To(Equal("train"))

				// reset
				calculatedPriorityClassName = ""
			})
		})

		Context("when handling auth configurations", func() {
			It("should handle isAuthDisabled=true for an exposed URL", func() {
				// Set isAuthDisabled=true (default)
				isAuthDisabled = true

				// Setup exposed URL
				(*exposedUrls.Value)["main"] = flds.ExposedUrlItemAndRules{
					Value: &flds.ExposedUrlItemValue{
						ContainerPort: ptr.To(int32(8080)),
					},
				}

				// Verify Build
				services, _, _, _, err := eb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadBaseName, owner, exposedUrls, ingressUrl, calculatedPriorityClassName, isJupyter, isTensorboard, isAuthDisabled)
				Expect(err).NotTo(HaveOccurred())
				Expect(services).To(HaveLen(1))

				// With auth disabled, should use the provided container port directly
				service := services[0]
				Expect(service.Spec.Ports[0].TargetPort.IntVal).To(Equal(int32(8080)))
			})

			It("should adjust container port if isAuthDisabled=false, private URL, and external auth enabled", func() {
				// Enable external auth
				_ = os.Setenv(config.EnvVarExternalAuthUrlEnabled, "true")

				// Set isAuthDisabled=false
				isAuthDisabled = false

				// Setup exposed URL
				(*exposedUrls.Value)["main"] = flds.ExposedUrlItemAndRules{
					Value: &flds.ExposedUrlItemValue{
						ContainerPort: ptr.To(int32(8080)),
					},
				}

				// Verify Build
				services, _, _, _, err := eb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadBaseName, owner, exposedUrls, ingressUrl, calculatedPriorityClassName, isJupyter, isTensorboard, isAuthDisabled)
				Expect(err).NotTo(HaveOccurred())
				Expect(services).To(HaveLen(1))

				// With auth enabled and private URL, should use a different container port
				service := services[0]
				Expect(service.Spec.Ports[0].TargetPort.IntVal).To(Equal(constants.ExternalUrlAuthContainerDefaultPort))
			})
		})

		Context("when handling OpenShift specific configurations", func() {
			It("should trim RunaiTrimStr from host if OpenShift and subdomain support are enabled for Routes", func() {
				// Enable OpenShift and subdomain support
				_ = os.Setenv(config.EnvVarOpenshift, "true")
				_ = os.Setenv(config.EnvVarSubdomainSupport, "true")

				// Setup exposed URL
				(*exposedUrls.Value)["main"] = flds.ExposedUrlItemAndRules{
					Value: &flds.ExposedUrlItemValue{
						ContainerPort: ptr.To(int32(8080)),
					},
				}

				// Verify Build
				_, _, routes, _, err := ebOpenshift.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadBaseName, owner, exposedUrls, ingressUrl, calculatedPriorityClassName, isJupyter, isTensorboard, isAuthDisabled)
				Expect(err).NotTo(HaveOccurred())
				Expect(routes).To(HaveLen(1))

				// Host should have RunaiTrimStr removed
				route := routes[0]
				// The host in our test won't actually have RunaiTrimStr since our URL composition doesn't add it in tests
				// But we can verify that the host doesn't contain RunaiTrimStr
				Expect(route.Spec.Host).NotTo(ContainSubstring(config.RunaiTrimStr))
			})

			It("should not trim RunaiTrimStr from host if OpenShift and subdomain support are disabled for Routes", func() {
				// Enable OpenShift and disable subdomain support
				_ = os.Setenv(config.EnvVarClusterDomain, "http://runai.cluster.example.com")
				_ = os.Setenv(config.EnvVarOpenshift, "true")
				_ = os.Unsetenv(config.EnvVarSubdomainSupport)

				// Setup exposed URL
				(*exposedUrls.Value)["main"] = flds.ExposedUrlItemAndRules{
					Value: &flds.ExposedUrlItemValue{
						ContainerPort: ptr.To(int32(8080)),
					},
				}

				// Verify Build
				_, _, routes, _, err := ebOpenshift.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadBaseName, owner, exposedUrls, ingressUrl, calculatedPriorityClassName, isJupyter, isTensorboard, isAuthDisabled)
				Expect(err).NotTo(HaveOccurred())
				Expect(routes).To(HaveLen(1))

				// Host should contain RunaiTrimStr
				route := routes[0]
				Expect(route.Spec.Host).To(ContainSubstring(config.RunaiTrimStr))
			})

			It("should correctly add JobURL to annotations considering OpenShift and subdomain support", func() {
				// Enable OpenShift and subdomain support
				_ = os.Setenv(config.EnvVarOpenshift, "true")
				_ = os.Setenv(config.EnvVarSubdomainSupport, "true")

				// Create a URL with RunaiTrimStr for testing
				urlWithRunai := "http://" + config.RunaiTrimStr + "example.com/path"
				jobUrl, err := apis.ParseURL(urlWithRunai)
				Expect(err).NotTo(HaveOccurred())

				// Create a service to use for buildIngressOrRoute
				service := corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-service",
					},
				}

				// Call buildIngressOrRoute directly to test URL handling
				err = ebOpenshift.buildIngressOrRoute(objectMeta, *jobUrl, &service)
				Expect(err).NotTo(HaveOccurred())

				// Check annotation - should have RunaiTrimStr removed
				Expect(objectMeta.Annotations[constants.AnnotationJobUrl]).To(Equal("http://example.com/path"))
			})
		})

		Context("when handling Ingress specific configurations", func() {
			It("should set IngressClassName if configured", func() {
				_ = os.Setenv(config.EnvVarIngressClass, "nginx")

				(*exposedUrls.Value)["main"] = flds.ExposedUrlItemAndRules{
					Value: &flds.ExposedUrlItemValue{
						ContainerPort: ptr.To(int32(8080)),
					},
				}

				_, ingresses, _, _, err := eb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadBaseName, owner, exposedUrls, ingressUrl, calculatedPriorityClassName, isJupyter, isTensorboard, isAuthDisabled)
				Expect(err).NotTo(HaveOccurred())
				Expect(ingresses).To(HaveLen(1))

				ingress := ingresses[0]
				Expect(ingress.Spec.IngressClassName).NotTo(BeNil())
				Expect(*ingress.Spec.IngressClassName).To(Equal("nginx"))
			})
		})

		Context("when Gateway API is enabled", func() {
			BeforeEach(func() {
				_ = os.Setenv(config.EnvVarGatewayAPIEnabled, "true")
				_ = os.Setenv(config.EnvVarGatewayName, "test-gateway")
				_ = os.Setenv(config.EnvVarGatewayNamespace, "runai")
			})

			It("should create HTTPRoute when gateway-only mode", func() {
				_ = os.Setenv(config.EnvVarIngressEnabled, "false")

				(*exposedUrls.Value)["main"] = flds.ExposedUrlItemAndRules{
					Value: &flds.ExposedUrlItemValue{
						ContainerPort: ptr.To(int32(8080)),
					},
				}

				services, ingresses, routes, httpRoutes, err := eb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadBaseName, owner, exposedUrls, ingressUrl, calculatedPriorityClassName, isJupyter, isTensorboard, isAuthDisabled)
				Expect(err).NotTo(HaveOccurred())
				Expect(services).To(HaveLen(1))
				Expect(ingresses).To(BeEmpty())
				Expect(routes).To(BeEmpty())
				Expect(httpRoutes).To(HaveLen(1))

				httpRoute := httpRoutes[0]
				Expect(httpRoute.Name).To(Equal(services[0].Name))
				Expect(httpRoute.Spec.ParentRefs).To(HaveLen(1))
				Expect(string(httpRoute.Spec.ParentRefs[0].Name)).To(Equal("test-gateway"))
				Expect(string(*httpRoute.Spec.ParentRefs[0].Namespace)).To(Equal("runai"))
				Expect(httpRoute.Spec.Hostnames).To(HaveLen(1))
				Expect(httpRoute.Spec.Rules).To(HaveLen(1))
				Expect(httpRoute.Spec.Rules[0].BackendRefs).To(HaveLen(1))
				Expect(string(httpRoute.Spec.Rules[0].BackendRefs[0].Name)).To(Equal(services[0].Name))
			})

			It("should create both Ingress and HTTPRoute in dual-mode", func() {
				(*exposedUrls.Value)["main"] = flds.ExposedUrlItemAndRules{
					Value: &flds.ExposedUrlItemValue{
						ContainerPort: ptr.To(int32(8080)),
					},
				}

				services, ingresses, routes, httpRoutes, err := eb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadBaseName, owner, exposedUrls, ingressUrl, calculatedPriorityClassName, isJupyter, isTensorboard, isAuthDisabled)
				Expect(err).NotTo(HaveOccurred())
				Expect(services).To(HaveLen(1))
				Expect(ingresses).To(HaveLen(1))
				Expect(routes).To(BeEmpty())
				Expect(httpRoutes).To(HaveLen(1))

				Expect(ingresses[0].Name).To(Equal(services[0].Name))
				Expect(httpRoutes[0].Name).To(Equal(services[0].Name))
			})

			It("should create both Route and HTTPRoute in dual-mode on OpenShift", func() {
				(*exposedUrls.Value)["main"] = flds.ExposedUrlItemAndRules{
					Value: &flds.ExposedUrlItemValue{
						ContainerPort: ptr.To(int32(8080)),
					},
				}

				services, ingresses, routes, httpRoutes, err := ebOpenshift.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadBaseName, owner, exposedUrls, ingressUrl, calculatedPriorityClassName, isJupyter, isTensorboard, isAuthDisabled)
				Expect(err).NotTo(HaveOccurred())
				Expect(services).To(HaveLen(1))
				Expect(ingresses).To(BeEmpty())
				Expect(routes).To(HaveLen(1))
				Expect(httpRoutes).To(HaveLen(1))

				Expect(routes[0].Name).To(Equal(services[0].Name))
				Expect(httpRoutes[0].Name).To(Equal(services[0].Name))
			})

			It("should include HTTPRoute in GetPreview", func() {
				(*exposedUrls.Value)["main"] = flds.ExposedUrlItemAndRules{
					Value: &flds.ExposedUrlItemValue{
						ContainerPort: ptr.To(int32(8080)),
					},
				}

				preview, err := eb.GetPreview(ctx, workloadInitials, objectMeta, exposedUrls, isJupyter, isTensorboard)
				Expect(err).NotTo(HaveOccurred())

				foundHTTPRoute := false
				for _, res := range preview {
					if res.Kind.Kind == constants.KindHTTPRoute {
						foundHTTPRoute = true
					}
				}
				Expect(foundHTTPRoute).To(BeTrue())
			})

			It("should create HTTPRoute with multiple exposed URLs", func() {
				_ = os.Setenv(config.EnvVarIngressEnabled, "false")

				(*exposedUrls.Value)["web"] = flds.ExposedUrlItemAndRules{
					Value: &flds.ExposedUrlItemValue{
						ContainerPort: ptr.To(int32(8080)),
					},
				}
				(*exposedUrls.Value)["api"] = flds.ExposedUrlItemAndRules{
					Value: &flds.ExposedUrlItemValue{
						ContainerPort: ptr.To(int32(9090)),
					},
				}

				services, ingresses, routes, httpRoutes, err := eb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadBaseName, owner, exposedUrls, ingressUrl, calculatedPriorityClassName, isJupyter, isTensorboard, isAuthDisabled)
				Expect(err).NotTo(HaveOccurred())
				Expect(services).To(HaveLen(2))
				Expect(ingresses).To(BeEmpty())
				Expect(routes).To(BeEmpty())
				Expect(httpRoutes).To(HaveLen(2))
			})

			It("should not create HTTPRoute when gateway is not enabled", func() {
				_ = os.Unsetenv(config.EnvVarGatewayAPIEnabled)
				_ = os.Unsetenv(config.EnvVarGatewayName)
				_ = os.Unsetenv(config.EnvVarGatewayNamespace)

				(*exposedUrls.Value)["main"] = flds.ExposedUrlItemAndRules{
					Value: &flds.ExposedUrlItemValue{
						ContainerPort: ptr.To(int32(8080)),
					},
				}

				_, _, _, httpRoutes, err := eb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadBaseName, owner, exposedUrls, ingressUrl, calculatedPriorityClassName, isJupyter, isTensorboard, isAuthDisabled)
				Expect(err).NotTo(HaveOccurred())
				Expect(httpRoutes).To(BeEmpty())
			})
		})
	})
})
