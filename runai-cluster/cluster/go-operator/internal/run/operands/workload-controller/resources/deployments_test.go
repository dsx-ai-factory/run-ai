package resources

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/utils"
	scheme_init "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/global"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Deployment", func() {
	var (
		runaiConfig *runaiv1.RunaiConfig
		operandName = "workload-controller"
	)

	BeforeEach(func() {
		runaiConfig = utils.GetDefaultRunaiConfig()
	})

	buildDeployment := func(cfg *runaiv1.RunaiConfig) *appsv1.Deployment {
		fakeClient := fake.NewClientBuilder().WithScheme(scheme_init.Scheme()).Build()
		resource := NewDeploymentResource(operandName, fakeClient, cfg)
		obj, err := resource.Object(context.Background())
		Expect(err).NotTo(HaveOccurred())
		return obj.(*appsv1.Deployment)
	}

	findEnvVar := func(envVars []v1.EnvVar, name string) *v1.EnvVar {
		for i := range envVars {
			if envVars[i].Name == name {
				return &envVars[i]
			}
		}
		return nil
	}

	Context("GATEWAY_API_ENABLED env var", func() {
		It("should set GATEWAY_API_ENABLED=true when GatewayAPI.Enabled is true", func() {
			runaiConfig.Spec.Global.GatewayAPI = &global.GatewayAPI{
				Enabled:   ptr.To(true),
				Name:      "my-gateway",
				Namespace: "gateway-ns",
			}

			deployment := buildDeployment(runaiConfig)
			envVars := deployment.Spec.Template.Spec.Containers[0].Env

			ev := findEnvVar(envVars, "GATEWAY_API_ENABLED")
			Expect(ev).NotTo(BeNil(), "expected GATEWAY_API_ENABLED env var to be present")
			Expect(ev.Value).To(Equal("true"))
		})

		It("should set GATEWAY_API_ENABLED=false when GatewayAPI.Enabled is false", func() {
			runaiConfig.Spec.Global.GatewayAPI = &global.GatewayAPI{
				Enabled: ptr.To(false),
			}

			deployment := buildDeployment(runaiConfig)
			envVars := deployment.Spec.Template.Spec.Containers[0].Env

			ev := findEnvVar(envVars, "GATEWAY_API_ENABLED")
			Expect(ev).NotTo(BeNil(), "expected GATEWAY_API_ENABLED env var to be present")
			Expect(ev.Value).To(Equal("false"))
		})

		It("should not set GATEWAY_API_ENABLED when GatewayAPI is nil", func() {
			runaiConfig.Spec.Global.GatewayAPI = nil

			deployment := buildDeployment(runaiConfig)
			envVars := deployment.Spec.Template.Spec.Containers[0].Env

			ev := findEnvVar(envVars, "GATEWAY_API_ENABLED")
			Expect(ev).To(BeNil(), "expected GATEWAY_API_ENABLED env var to be absent")
		})

		It("should set GATEWAY_NAME and GATEWAY_NAMESPACE when GatewayAPI is enabled", func() {
			runaiConfig.Spec.Global.GatewayAPI = &global.GatewayAPI{
				Enabled:   ptr.To(true),
				Name:      "my-gateway",
				Namespace: "gateway-ns",
			}

			deployment := buildDeployment(runaiConfig)
			envVars := deployment.Spec.Template.Spec.Containers[0].Env

			nameEv := findEnvVar(envVars, "GATEWAY_NAME")
			Expect(nameEv).NotTo(BeNil(), "expected GATEWAY_NAME env var to be present")
			Expect(nameEv.Value).To(Equal("my-gateway"))

			nsEv := findEnvVar(envVars, "GATEWAY_NAMESPACE")
			Expect(nsEv).NotTo(BeNil(), "expected GATEWAY_NAMESPACE env var to be present")
			Expect(nsEv.Value).To(Equal("gateway-ns"))
		})

		It("should not set GATEWAY_NAME or GATEWAY_NAMESPACE when GatewayAPI is disabled", func() {
			runaiConfig.Spec.Global.GatewayAPI = &global.GatewayAPI{
				Enabled: ptr.To(false),
			}

			deployment := buildDeployment(runaiConfig)
			envVars := deployment.Spec.Template.Spec.Containers[0].Env

			nameEv := findEnvVar(envVars, "GATEWAY_NAME")
			Expect(nameEv).To(BeNil(), "expected GATEWAY_NAME env var to be absent")

			nsEv := findEnvVar(envVars, "GATEWAY_NAMESPACE")
			Expect(nsEv).To(BeNil(), "expected GATEWAY_NAMESPACE env var to be absent")
		})
	})

	Context("workload-injected container resource env vars", func() {
		It("should serialize all configured resources", func() {
			runaiConfig.Spec.Global.WorkloadInjectedContainers = &global.WorkloadInjectedContainersConfig{
				OAuth2Proxy: &global.OAuth2ProxyContainerConfig{
					Resources: &global.WorkloadInjectedContainerResourceRequirements{
						Requests: v1.ResourceList{v1.ResourceCPU: resource.MustParse("100m")},
						Limits:   v1.ResourceList{v1.ResourceMemory: resource.MustParse("128Mi")},
					},
				},
				UsernameLoader: &global.UsernameLoaderContainerConfig{Resources: &global.WorkloadInjectedContainerResourceRequirements{}},
				GitSync:        &global.GitSyncContainerConfig{Resources: &global.WorkloadInjectedContainerResourceRequirements{}},
				S3FileSystem:   &global.S3FileSystemContainerConfig{Resources: &global.WorkloadInjectedContainerResourceRequirements{}},
			}

			envs := buildDeployment(runaiConfig).Spec.Template.Spec.Containers[0].Env

			Expect(findEnvVar(envs, constants.OAuth2ProxyResourcesEnvVar).Value).To(MatchJSON(`{"requests":{"cpu":"100m"},"limits":{"memory":"128Mi"}}`))
			for _, name := range []string{
				constants.UsernameLoaderResourcesEnvVar,
				constants.GitSyncResourcesEnvVar,
				constants.S3FileSystemResourcesEnvVar,
			} {
				Expect(findEnvVar(envs, name)).NotTo(BeNil())
				Expect(findEnvVar(envs, name).Value).To(MatchJSON(`{}`))
			}
		})

		It("should omit all env vars when resources are not configured", func() {
			envs := buildDeployment(runaiConfig).Spec.Template.Spec.Containers[0].Env

			for _, name := range []string{
				constants.OAuth2ProxyResourcesEnvVar,
				constants.UsernameLoaderResourcesEnvVar,
				constants.GitSyncResourcesEnvVar,
				constants.S3FileSystemResourcesEnvVar,
			} {
				Expect(findEnvVar(envs, name)).To(BeNil())
			}
		})
	})
})
