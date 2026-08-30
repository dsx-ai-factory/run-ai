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
		deployment  *appsv1.Deployment
		envVars     []v1.EnvVar
		ctx         context.Context
		fakeReader  = fake.NewClientBuilder().WithScheme(scheme_init.Scheme()).Build()
		operandName = "inference-workload-controller"
	)

	buildDeploymentFromConfig := func(runaiConfig *runaiv1.RunaiConfig) {
		resource := NewDeploymentResource(operandName, fakeReader, runaiConfig)
		obj, err := resource.Object(ctx)
		Expect(err).NotTo(HaveOccurred())

		deployment = obj.(*appsv1.Deployment)
		envVars = deployment.Spec.Template.Spec.Containers[0].Env
	}
	buildDeployment := func(gatewayAPI *global.GatewayAPI) {
		runaiConfig := utils.GetDefaultRunaiConfig()
		runaiConfig.Spec.Global.GatewayAPI = gatewayAPI
		buildDeploymentFromConfig(runaiConfig)
	}

	BeforeEach(func() {
		ctx = context.Background()
	})

	Context("GATEWAY_API_ENABLED env var", func() {
		It("should set GATEWAY_API_ENABLED=true when GatewayAPI.Enabled is true", func() {
			buildDeployment(&global.GatewayAPI{
				Enabled:   ptr.To(true),
				Name:      "my-gateway",
				Namespace: "gateway-ns",
			})

			var found *v1.EnvVar
			for i := range envVars {
				if envVars[i].Name == "GATEWAY_API_ENABLED" {
					found = &envVars[i]
					break
				}
			}
			Expect(found).NotTo(BeNil(), "GATEWAY_API_ENABLED env var should be present")
			Expect(found.Value).To(Equal("true"))
		})

		It("should set GATEWAY_API_ENABLED=false when GatewayAPI.Enabled is false", func() {
			buildDeployment(&global.GatewayAPI{
				Enabled: ptr.To(false),
			})

			var found *v1.EnvVar
			for i := range envVars {
				if envVars[i].Name == "GATEWAY_API_ENABLED" {
					found = &envVars[i]
					break
				}
			}
			Expect(found).NotTo(BeNil(), "GATEWAY_API_ENABLED env var should be present")
			Expect(found.Value).To(Equal("false"))
		})

		It("should not include GATEWAY_API_ENABLED when GatewayAPI is nil", func() {
			buildDeployment(nil)

			for _, env := range envVars {
				Expect(env.Name).NotTo(Equal("GATEWAY_API_ENABLED"),
					"GATEWAY_API_ENABLED should not be present when GatewayAPI is nil")
			}
		})
	})

	Context("workload-injected container resource env vars", func() {
		It("should serialize all configured resources", func() {
			runaiConfig := utils.GetDefaultRunaiConfig()
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

			buildDeploymentFromConfig(runaiConfig)

			Expect(findEnvVar(envVars, constants.OAuth2ProxyResourcesEnvVar).Value).To(MatchJSON(`{"requests":{"cpu":"100m"},"limits":{"memory":"128Mi"}}`))
			for _, name := range []string{
				constants.UsernameLoaderResourcesEnvVar,
				constants.GitSyncResourcesEnvVar,
				constants.S3FileSystemResourcesEnvVar,
			} {
				Expect(findEnvVar(envVars, name)).NotTo(BeNil())
				Expect(findEnvVar(envVars, name).Value).To(MatchJSON(`{}`))
			}
		})

		It("should omit all env vars when resources are not configured", func() {
			buildDeployment(nil)

			for _, name := range []string{
				constants.OAuth2ProxyResourcesEnvVar,
				constants.UsernameLoaderResourcesEnvVar,
				constants.GitSyncResourcesEnvVar,
				constants.S3FileSystemResourcesEnvVar,
			} {
				Expect(findEnvVar(envVars, name)).To(BeNil())
			}
		})
	})
})

func findEnvVar(envVars []v1.EnvVar, name string) *v1.EnvVar {
	for i := range envVars {
		if envVars[i].Name == name {
			return &envVars[i]
		}
	}
	return nil
}
