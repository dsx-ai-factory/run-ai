package composer_v2

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer"

	clustersApi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
)

var _ = Describe("buildExternalAccess", func() {
	const (
		svcName     = "kf-py1-url-master"
		urlStr      = "https://ofer-py1.demotest.runailabs.com/"
		servicePort = int32(80)
	)

	Context("vanilla Kubernetes (no OpenShift, no Gateway)", func() {
		It("produces an Ingress and nothing else", func() {
			cfg := composer.StaticClusterStorageConfig{}

			ea, err := buildExternalAccess(cfg, svcName, urlStr, servicePort)
			Expect(err).To(BeNil())
			Expect(ea).NotTo(BeNil())
			Expect(ea.HasIngress()).To(BeTrue())
			Expect(ea.HasRoute()).To(BeFalse())
			Expect(ea.HasHttpRoute()).To(BeFalse())

			ingress := ea.GetIngress()
			Expect(ingress["kind"]).To(Equal("Ingress"))
			Expect(ingress["apiVersion"]).To(Equal("networking.k8s.io/v1"))
		})
	})

	Context("OpenShift (no Gateway)", func() {
		It("produces a Route (no Ingress) with edge TLS", func() {
			cfg := composer.StaticClusterStorageConfig{}
			cfg.SetIsOpenshift(true)

			ea, err := buildExternalAccess(cfg, svcName, urlStr, servicePort)
			Expect(err).To(BeNil())
			Expect(ea.HasIngress()).To(BeFalse())
			Expect(ea.HasRoute()).To(BeTrue())
			Expect(ea.HasHttpRoute()).To(BeFalse())

			route := ea.GetRoute()
			Expect(route["kind"]).To(Equal("Route"))
			Expect(route["apiVersion"]).To(Equal("route.openshift.io/v1"))

			meta := route["metadata"].(map[string]interface{})
			Expect(meta["name"]).To(Equal(svcName))

			spec := route["spec"].(map[string]interface{})
			Expect(spec["host"]).To(Equal("ofer-py1.demotest.runailabs.com"))
			Expect(spec["path"]).To(Equal("/"))

			to := spec["to"].(map[string]interface{})
			Expect(to["kind"]).To(Equal("Service"))
			Expect(to["name"]).To(Equal(svcName))

			tls := spec["tls"].(map[string]interface{})
			Expect(tls["termination"]).To(Equal("edge"))
			Expect(tls["insecureEdgeTerminationPolicy"]).To(Equal("Redirect"))
		})

		It("strips the runai. host prefix when subdomain support is enabled", func() {
			cfg := composer.StaticClusterStorageConfig{}
			cfg.SetIsOpenshift(true)
			cfg.SetSubdomainSupportEnabled(true)

			subdomainURL := "https://runai.ofer-py1.demotest.runailabs.com/"

			ea, err := buildExternalAccess(cfg, svcName, subdomainURL, servicePort)
			Expect(err).To(BeNil())

			spec := ea.GetRoute()["spec"].(map[string]interface{})
			Expect(spec["host"]).To(Equal("ofer-py1.demotest.runailabs.com"))
		})

		It("uses passthrough termination when the cluster opts in via openshiftRouteConfiguration", func() {
			cfg := composer.StaticClusterStorageConfig{}
			cfg.SetIsOpenshift(true)
			cfg.SetOpenshiftRouteTerminationPolicy(string(clustersApi.OPENSHIFTROUTETERMINATIONPOLICY_PASSTHROUGH))

			ea, err := buildExternalAccess(cfg, svcName, urlStr, servicePort)
			Expect(err).To(BeNil())

			spec := ea.GetRoute()["spec"].(map[string]interface{})
			tls := spec["tls"].(map[string]interface{})
			Expect(tls["termination"]).To(Equal("passthrough"))
			Expect(tls["insecureEdgeTerminationPolicy"]).To(Equal("Redirect"))
			// OpenShift rejects passthrough Routes with spec.path set — the Router
			// cannot do path-based routing on encrypted traffic.
			_, hasPath := spec["path"]
			Expect(hasPath).To(BeFalse())
		})
	})

	Context("Gateway API (additive)", func() {
		It("produces both an Ingress and an HTTPRoute on a vanilla cluster with a Gateway configured", func() {
			cfg := composer.StaticClusterStorageConfig{}
			cfg.SetGatewayAPIEnabled(true)
			cfg.SetGatewayName("runai-gateway")
			cfg.SetGatewayNamespace("runai")

			ea, err := buildExternalAccess(cfg, svcName, urlStr, servicePort)
			Expect(err).To(BeNil())
			Expect(ea.HasIngress()).To(BeTrue())
			Expect(ea.HasRoute()).To(BeFalse())
			Expect(ea.HasHttpRoute()).To(BeTrue())

			httpRoute := ea.GetHttpRoute()
			Expect(httpRoute["kind"]).To(Equal("HTTPRoute"))
			Expect(httpRoute["apiVersion"]).To(Equal("gateway.networking.k8s.io/v1"))

			meta := httpRoute["metadata"].(map[string]interface{})
			Expect(meta["name"]).To(Equal(svcName))

			spec := httpRoute["spec"].(map[string]interface{})
			parents := spec["parentRefs"].([]map[string]interface{})
			Expect(parents).To(HaveLen(1))
			Expect(parents[0]["name"]).To(Equal("runai-gateway"))
			Expect(parents[0]["namespace"]).To(Equal("runai"))

			Expect(spec["hostnames"]).To(Equal([]string{"ofer-py1.demotest.runailabs.com"}))

			rules := spec["rules"].([]map[string]interface{})
			Expect(rules).To(HaveLen(1))

			matches := rules[0]["matches"].([]map[string]interface{})
			Expect(matches).To(HaveLen(1))
			Expect(matches[0]["path"]).To(Equal(map[string]interface{}{"type": "PathPrefix", "value": "/"}))

			backendRefs := rules[0]["backendRefs"].([]map[string]interface{})
			Expect(backendRefs).To(HaveLen(1))
			Expect(backendRefs[0]["name"]).To(Equal(svcName))
			Expect(backendRefs[0]["port"]).To(Equal(servicePort))
		})

		It("produces both a Route and an HTTPRoute on OpenShift with a Gateway configured", func() {
			cfg := composer.StaticClusterStorageConfig{}
			cfg.SetGatewayAPIEnabled(true)
			cfg.SetIsOpenshift(true)
			cfg.SetGatewayName("runai-gateway")
			cfg.SetGatewayNamespace("runai")

			ea, err := buildExternalAccess(cfg, svcName, urlStr, servicePort)
			Expect(err).To(BeNil())
			Expect(ea.HasIngress()).To(BeFalse())
			Expect(ea.HasRoute()).To(BeTrue())
			Expect(ea.HasHttpRoute()).To(BeTrue())
		})

		It("strips the runai. host prefix from the HTTPRoute on OpenShift+subdomain", func() {
			cfg := composer.StaticClusterStorageConfig{}
			cfg.SetGatewayAPIEnabled(true)
			cfg.SetIsOpenshift(true)
			cfg.SetSubdomainSupportEnabled(true)
			cfg.SetGatewayName("runai-gateway")
			cfg.SetGatewayNamespace("runai")

			subdomainURL := "https://runai.ofer-py1.demotest.runailabs.com/"

			ea, err := buildExternalAccess(cfg, svcName, subdomainURL, servicePort)
			Expect(err).To(BeNil())

			spec := ea.GetHttpRoute()["spec"].(map[string]interface{})
			Expect(spec["hostnames"]).To(Equal([]string{"ofer-py1.demotest.runailabs.com"}))
		})

		It("does NOT strip the runai. host prefix from the HTTPRoute on a non-OpenShift cluster", func() {
			cfg := composer.StaticClusterStorageConfig{}
			cfg.SetGatewayAPIEnabled(true)
			cfg.SetSubdomainSupportEnabled(true)
			cfg.SetGatewayName("runai-gateway")
			cfg.SetGatewayNamespace("runai")

			subdomainURL := "https://runai.ofer-py1.demotest.runailabs.com/"

			ea, err := buildExternalAccess(cfg, svcName, subdomainURL, servicePort)
			Expect(err).To(BeNil())

			spec := ea.GetHttpRoute()["spec"].(map[string]interface{})
			Expect(spec["hostnames"]).To(Equal([]string{"runai.ofer-py1.demotest.runailabs.com"}))
		})

		It("omits namespace from parentRefs when GatewayNamespace is empty", func() {
			cfg := composer.StaticClusterStorageConfig{}
			cfg.SetGatewayAPIEnabled(true)
			cfg.SetGatewayName("runai-gateway")

			ea, err := buildExternalAccess(cfg, svcName, urlStr, servicePort)
			Expect(err).To(BeNil())

			spec := ea.GetHttpRoute()["spec"].(map[string]interface{})
			parents := spec["parentRefs"].([]map[string]interface{})
			Expect(parents[0]["name"]).To(Equal("runai-gateway"))
			_, hasNs := parents[0]["namespace"]
			Expect(hasNs).To(BeFalse())
		})

		It("produces only an HTTPRoute when ingress is disabled and Gateway API is enabled", func() {
			cfg := composer.StaticClusterStorageConfig{}
			cfg.SetIngressEnabled(false)
			cfg.SetGatewayAPIEnabled(true)
			cfg.SetGatewayNamespace("runai")
			cfg.SetGatewayName("runai-gateway")

			ea, err := buildExternalAccess(cfg, svcName, urlStr, servicePort)
			Expect(err).To(BeNil())
			Expect(ea.HasIngress()).To(BeFalse())
			Expect(ea.HasRoute()).To(BeFalse())
			Expect(ea.HasHttpRoute()).To(BeTrue())
		})

		It("fails when Gateway API is enabled but gateway name is missing", func() {
			cfg := composer.StaticClusterStorageConfig{}
			cfg.SetIngressEnabled(false)
			cfg.SetGatewayAPIEnabled(true)

			_, err := buildExternalAccess(cfg, svcName, urlStr, servicePort)
			Expect(err).NotTo(BeNil())
			Expect(err.Code).To(Equal(int32(503)))
		})
	})

	Context("no external routing mechanism", func() {
		It("fails when both ingress and Gateway API are disabled", func() {
			cfg := composer.StaticClusterStorageConfig{}
			cfg.SetIngressEnabled(false)
			cfg.SetGatewayAPIEnabled(false)

			_, err := buildExternalAccess(cfg, svcName, urlStr, servicePort)
			Expect(err).NotTo(BeNil())
			Expect(err.Code).To(Equal(int32(503)))
		})
	})

	Context("URL parsing failure", func() {
		It("returns 400 for an invalid URL on the OpenShift Route path", func() {
			cfg := composer.StaticClusterStorageConfig{}
			cfg.SetIsOpenshift(true)

			_, err := buildExternalAccess(cfg, svcName, "://not-a-url", servicePort)
			Expect(err).NotTo(BeNil())
			Expect(err.Code).To(Equal(int32(400)))
		})

		It("returns 400 for an invalid URL on the HTTPRoute path", func() {
			cfg := composer.StaticClusterStorageConfig{}
			cfg.SetIngressEnabled(false)
			cfg.SetGatewayAPIEnabled(true)
			cfg.SetGatewayName("runai-gateway")
			cfg.SetGatewayNamespace("runai")

			_, err := buildExternalAccess(cfg, svcName, "://not-a-url", servicePort)
			Expect(err).NotTo(BeNil())
			Expect(err.Code).To(Equal(int32(400)))
		})
	})
})
