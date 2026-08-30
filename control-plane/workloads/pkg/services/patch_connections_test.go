package services

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/testutils"
)

const malformedUrl = "txtx:fofj/1!"

func TestPatchConnectionsServingPortWithPreprocessedExternalURL(t *testing.T) {
	const (
		jobURL      = "https://annotated-job.example.com"
		servingURL  = "http://inference.runai-project.inference.example.com"
		externalURL = "https://tool.example.com"
	)
	envs := []api.Environment{{
		Connections: []api.Connection{
			{
				Name:                "tool",
				ConnectionType:      TypeUrl,
				ServiceResourceName: convutils.Ref("infw-inference-url-tool"),
				ExposedUrlPath:      convutils.Ref("/project/inference"),
			},
			{
				Name:           "ServingPort",
				ConnectionType: TypeServingPort,
			},
		},
	}}
	networkStatus := &api.WorkloadNetworkStatus{
		Services: []api.WorkloadNetworkStatusServicesInner{{
			Name: "infw-inference-url-tool",
			Urls: []api.WorkloadNetworkStatusServicesInnerUrlsInner{{Url: externalURL}},
		}},
	}

	method, err := patchConnections(context.Background(), envs, []string{jobURL, servingURL}, networkStatus)
	if err != nil {
		t.Fatalf("patchConnections returned an error: %v", err)
	}
	if method != PatchingMethodNetworkStatPreProcessed {
		t.Fatalf("expected preprocessed patching, got %q", method)
	}
	if got := envs[0].Connections[0].Url; got != externalURL {
		t.Errorf("expected external URL %q, got %q", externalURL, got)
	}
	if got := envs[0].Connections[1].Url; got != servingURL {
		t.Errorf("expected serving URL %q, got %q", servingURL, got)
	}
}

var _ = Describe("patching test", func() {
	var bg = context.Background()

	Context("urlinfo", func() {
		It("empty list - empty info", func() {
			urlsInfo, err := loadUrlsInfo([]string{})
			Expect(err).To(BeNil())
			Expect(urlsInfo.urls).To(BeEmpty())
			Expect(urlsInfo.domain).To(BeNil())
			Expect(urlsInfo.ipAddress).To(BeNil())
		})

		It("only port", func() {
			urlsInfo, err := loadUrlsInfo([]string{
				"1.2.3.4:5",
			})
			Expect(err).To(BeNil())
			Expect(len(urlsInfo.urls)).To(Equal(1))

			Expect(urlsInfo.ipAddress).NotTo(BeNil())
			Expect(*urlsInfo.ipAddress).To(Equal("1.2.3.4"))

			Expect(urlsInfo.domain).To(BeNil())
		})

		It("only exposedUrl - subdomain", func() {
			urlsInfo, err := loadUrlsInfo([]string{
				"https://a.domain.com",
			})
			Expect(err).To(BeNil())
			Expect(len(urlsInfo.urls)).To(Equal(1))

			Expect(urlsInfo.ipAddress).To(BeNil())

			Expect(urlsInfo.domain).NotTo(BeNil())
			Expect(urlsInfo.domain.String()).To(Equal("https://domain.com"))
			Expect(urlsInfo.subdomainSupport).To(Equal(true))
		})

		It("only exposedUrl - subdomain", func() {
			urlsInfo, err := loadUrlsInfo([]string{
				"https://a.domain.com/",
			})
			Expect(err).To(BeNil())
			Expect(len(urlsInfo.urls)).To(Equal(1))

			Expect(urlsInfo.ipAddress).To(BeNil())

			Expect(urlsInfo.domain).NotTo(BeNil())
			Expect(urlsInfo.domain.String()).To(Equal("https://domain.com"))
			Expect(urlsInfo.subdomainSupport).To(Equal(true))
		})

		It("only exposedUrl - no subdomain", func() {
			urlsInfo, err := loadUrlsInfo([]string{
				"https://domain.com/a",
			})
			Expect(err).To(BeNil())
			Expect(len(urlsInfo.urls)).To(Equal(1))

			Expect(urlsInfo.ipAddress).To(BeNil())

			Expect(urlsInfo.domain).NotTo(BeNil())
			Expect(urlsInfo.domain.String()).To(Equal("https://domain.com"))
			Expect(urlsInfo.subdomainSupport).To(Equal(false))
		})

		It("1+1", func() {
			urlsInfo, err := loadUrlsInfo([]string{
				"https://a.domain.com",
				"1.2.3.4:5",
			})
			Expect(err).To(BeNil())
			Expect(len(urlsInfo.urls)).To(Equal(2))

			Expect(urlsInfo.ipAddress).NotTo(BeNil())
			Expect(*urlsInfo.ipAddress).To(Equal("1.2.3.4"))

			Expect(urlsInfo.domain).NotTo(BeNil())
			Expect(urlsInfo.domain.String()).To(Equal("https://domain.com"))
			Expect(urlsInfo.subdomainSupport).To(Equal(true))
		})

		It("n+m", func() {
			urlsInfo, err := loadUrlsInfo([]string{
				"https://a.domain.com",
				"https://b.domain.com",
				"https://c.domain.com",
				"1.2.3.4:5",
				"1.2.3.4:6",
				"1.2.3.4:7",
			})
			Expect(err).To(BeNil())
			Expect(len(urlsInfo.urls)).To(Equal(6))

			Expect(urlsInfo.ipAddress).NotTo(BeNil())
			Expect(*urlsInfo.ipAddress).To(Equal("1.2.3.4"))

			Expect(urlsInfo.domain).NotTo(BeNil())
			Expect(urlsInfo.domain.String()).To(Equal("https://domain.com"))
			Expect(urlsInfo.subdomainSupport).To(Equal(true))
		})
	})

	Context("network status service urls", func() {
		It("patches a URL connection from the correlated service status URL", func() {
			envs := []api.Environment{
				{
					Connections: []api.Connection{
						{
							Name:                "jupyter",
							ConnectionType:      TypeUrl,
							ServiceResourceName: convutils.Ref("svc-jupyter"),
							ExposedUrlPath:      convutils.Ref("/fallback"),
						},
					},
				},
			}
			networkStatus := &api.WorkloadNetworkStatus{
				Services: []api.WorkloadNetworkStatusServicesInner{
					{
						Name: "svc-jupyter",
						Urls: []api.WorkloadNetworkStatusServicesInnerUrlsInner{
							{
								Url:        "https://jupyter.example.com/lab",
								SourceKind: "Ingress",
								SourceName: "jupyter-ingress",
							},
						},
					},
				},
			}

			method, err := patchConnections(bg, envs, nil, networkStatus)

			Expect(err).To(BeNil())
			Expect(method).To(Equal(PatchingMethodNetworkStatPreProcessed))
			Expect(envs[0].Connections).To(HaveLen(1))
			Expect(envs[0].Connections[0].Url).To(Equal("https://jupyter.example.com/lab"))
		})

		It("patches a URL connection from a service status URL with only serviceResourceName", func() {
			envs := []api.Environment{
				{
					Connections: []api.Connection{
						{
							Name:                "jupyter",
							ConnectionType:      TypeUrl,
							ServiceResourceName: convutils.Ref("svc-jupyter"),
						},
					},
				},
			}
			networkStatus := &api.WorkloadNetworkStatus{
				Services: []api.WorkloadNetworkStatusServicesInner{
					{
						Name: "svc-jupyter",
						Urls: []api.WorkloadNetworkStatusServicesInnerUrlsInner{
							{
								Url:        "https://jupyter.example.com/lab",
								SourceKind: "Ingress",
								SourceName: "jupyter-ingress",
							},
						},
					},
				},
			}

			method, err := patchConnections(bg, envs, nil, networkStatus)

			Expect(err).To(BeNil())
			Expect(method).To(Equal(PatchingMethodNetworkStatPreProcessed))
			Expect(envs[0].Connections[0].Url).To(Equal("https://jupyter.example.com/lab"))
		})

		It("does not choose a service status URL when the service has multiple URLs", func() {
			envs := []api.Environment{
				{
					Connections: []api.Connection{
						{
							Name:                "jupyter",
							ConnectionType:      TypeUrl,
							ServiceResourceName: convutils.Ref("svc-jupyter"),
							ExposedUrlPath:      convutils.Ref("/fallback"),
						},
					},
				},
			}
			networkStatus := &api.WorkloadNetworkStatus{
				Services: []api.WorkloadNetworkStatusServicesInner{
					{
						Name: "svc-jupyter",
						Urls: []api.WorkloadNetworkStatusServicesInnerUrlsInner{
							{
								Url:        "https://jupyter-primary.example.com/lab",
								SourceKind: "Ingress",
								SourceName: "jupyter-ingress",
							},
							{
								Url:        "https://jupyter-gateway.example.com/lab",
								SourceKind: "HTTPRoute",
								SourceName: "jupyter-route",
							},
						},
					},
				},
			}

			method, err := patchConnections(bg, envs, []string{"https://apps.example.com/base"}, networkStatus)

			Expect(err).To(BeNil())
			Expect(method).To(Equal(PatchingMethodNetworkStatPreProcessed))
			Expect(envs[0].Connections[0].Url).To(Equal("https://apps.example.com/fallback"))
		})

		It("patches a URL connection when multiple external access resources report the same service URL", func() {
			envs := []api.Environment{
				{
					Connections: []api.Connection{
						{
							Name:                "jupyter",
							ConnectionType:      TypeUrl,
							ServiceResourceName: convutils.Ref("svc-jupyter"),
							ExposedUrlPath:      convutils.Ref("/fallback"),
						},
					},
				},
			}
			networkStatus := &api.WorkloadNetworkStatus{
				Services: []api.WorkloadNetworkStatusServicesInner{
					{
						Name: "svc-jupyter",
						Urls: []api.WorkloadNetworkStatusServicesInnerUrlsInner{
							{
								Url:        "https://jupyter.example.com/lab",
								SourceKind: "Ingress",
								SourceName: "jupyter-ingress",
							},
							{
								Url:        "https://jupyter.example.com/lab",
								SourceKind: "HTTPRoute",
								SourceName: "jupyter-route",
							},
						},
					},
				},
			}

			method, err := patchConnections(bg, envs, []string{"https://apps.example.com/base"}, networkStatus)

			Expect(err).To(BeNil())
			Expect(method).To(Equal(PatchingMethodNetworkStatPreProcessed))
			Expect(envs[0].Connections[0].Url).To(Equal("https://jupyter.example.com/lab"))
		})

		It("patches each URL connection from its correlated service when dual external access resources report duplicate URLs", func() {
			envs := []api.Environment{
				{
					Connections: []api.Connection{
						{
							Name:                "worker",
							ConnectionType:      TypeUrl,
							ServiceResourceName: convutils.Ref("svc-worker"),
							ExposedUrlPath:      convutils.Ref("/fallback"),
						},
						{
							Name:                "master",
							ConnectionType:      TypeUrl,
							ServiceResourceName: convutils.Ref("svc-master"),
							ExposedUrlPath:      convutils.Ref("/fallback"),
						},
					},
				},
			}
			networkStatus := &api.WorkloadNetworkStatus{
				Services: []api.WorkloadNetworkStatusServicesInner{
					{
						Name: "svc-worker",
						Urls: []api.WorkloadNetworkStatusServicesInnerUrlsInner{
							{
								Url:        "https://worker.example.com",
								SourceKind: "Ingress",
								SourceName: "worker-ingress",
							},
							{
								Url:        "https://worker.example.com",
								SourceKind: "HTTPRoute",
								SourceName: "worker-route",
							},
						},
					},
					{
						Name: "svc-master",
						Urls: []api.WorkloadNetworkStatusServicesInnerUrlsInner{
							{
								Url:        "https://master.example.com",
								SourceKind: "Ingress",
								SourceName: "master-ingress",
							},
							{
								Url:        "https://master.example.com",
								SourceKind: "HTTPRoute",
								SourceName: "master-route",
							},
						},
					},
				},
			}

			method, err := patchConnections(bg, envs, []string{"https://apps.example.com/base"}, networkStatus)

			Expect(err).To(BeNil())
			Expect(method).To(Equal(PatchingMethodNetworkStatPreProcessed))
			Expect(envs[0].Connections[0].Url).To(Equal("https://worker.example.com"))
			Expect(envs[0].Connections[1].Url).To(Equal("https://master.example.com"))
		})

		It("patches node port connections from correlated service status URLs", func() {
			envs, networkStatus := nodePortServiceStatusURLFixture()

			method, err := patchConnections(bg, envs, nil, networkStatus)

			Expect(err).To(BeNil())
			Expect(method).To(Equal(PatchingMethodNetworkStatPreProcessed))
			Expect(envs[0].Connections[0].Url).To(Equal("10.0.0.1:32101"))
			Expect(envs[0].Connections[1].Url).To(Equal("10.0.0.1:32102"))
		})

		It("patches load balancer connections from the first valid ingress host", func() {
			envs := []api.Environment{
				{
					Connections: []api.Connection{
						{
							Name:           "tensorboard",
							ConnectionType: TypeLoadBalancer,
							ContainerPort:  6006,
						},
					},
				},
			}
			networkStatus := &api.WorkloadNetworkStatus{
				Services: []api.WorkloadNetworkStatusServicesInner{
					{
						Name: "svc-tensorboard",
						LoadBalancer: api.WorkloadNetworkStatusServicesInnerLoadBalancer{
							Ingress: []api.WorkloadNetworkStatusServicesInnerLoadBalancerIngressInner{
								{},
								{Hostname: "tensorboard-lb.example.com"},
							},
						},
						Ports: []api.WorkloadNetworkStatusServicesInnerPortsInner{
							{
								Port:       80,
								TargetPort: "6006",
								Protocol:   "TCP",
							},
						},
					},
				},
			}

			method, err := patchConnections(bg, envs, nil, networkStatus)

			Expect(err).To(BeNil())
			Expect(method).To(Equal(PatchingMethodNetworkStatNoPreProc))
			Expect(envs[0].Connections[0].Url).To(Equal("tensorboard-lb.example.com:80"))
		})

		It("patches load balancer connections by service and port metadata before target port fallback", func() {
			envs := []api.Environment{
				{
					Connections: []api.Connection{
						{
							Name:                "tensorboard-a",
							ConnectionType:      TypeLoadBalancer,
							ContainerPort:       6006,
							ServiceResourceName: convutils.Ref("svc-tensorboard-a"),
							PortResourceName:    convutils.Ref("port-6006"),
						},
						{
							Name:                "tensorboard-b",
							ConnectionType:      TypeLoadBalancer,
							ContainerPort:       6006,
							ServiceResourceName: convutils.Ref("svc-tensorboard-b"),
							PortResourceName:    convutils.Ref("port-6006"),
						},
					},
				},
			}
			networkStatus := &api.WorkloadNetworkStatus{
				Services: []api.WorkloadNetworkStatusServicesInner{
					{
						Name: "svc-tensorboard-a",
						LoadBalancer: api.WorkloadNetworkStatusServicesInnerLoadBalancer{
							Ingress: []api.WorkloadNetworkStatusServicesInnerLoadBalancerIngressInner{
								{Hostname: "tensorboard-a-lb.example.com"},
							},
						},
						Ports: []api.WorkloadNetworkStatusServicesInnerPortsInner{
							{
								Name:       "port-6006",
								Port:       80,
								TargetPort: "6006",
								Protocol:   "TCP",
							},
						},
					},
					{
						Name: "svc-tensorboard-b",
						LoadBalancer: api.WorkloadNetworkStatusServicesInnerLoadBalancer{
							Ingress: []api.WorkloadNetworkStatusServicesInnerLoadBalancerIngressInner{
								{Hostname: "tensorboard-b-lb.example.com"},
							},
						},
						Ports: []api.WorkloadNetworkStatusServicesInnerPortsInner{
							{
								Name:       "port-6006",
								Port:       81,
								TargetPort: "6006",
								Protocol:   "TCP",
							},
						},
					},
				},
			}

			method, err := patchConnections(bg, envs, nil, networkStatus)

			Expect(err).To(BeNil())
			Expect(method).To(Equal(PatchingMethodNetworkStatPreProcessed))
			Expect(envs[0].Connections[0].Url).To(Equal("tensorboard-a-lb.example.com:80"))
			Expect(envs[0].Connections[1].Url).To(Equal("tensorboard-b-lb.example.com:81"))
		})
	})

	Context("naive method", func() {
		It("no connections, no links -> no patching", func() {
			urls, _, envs, _ := initTest(2)

			patched := duplicateOf(envs)

			method, err := patchConnections(bg, patched, urls, nil)
			Expect(method).To(Equal(PatchingMethodNaive))
			Expect(err).To(BeNil())
			Expect(patched).To(Equal(envs))
		})

		It("two environments x 2 connections, one links one urls", func() {
			urls, networkStatus, envs, envsPrepped := initTest(2)

			addToTest(&envs[0], &envsPrepped[0], &urls, &networkStatus,
				"https://a.exposed.com",
				&api.Connection{
					Name:                "url1",
					ConnectionType:      TypeUrl,
					ContainerPort:       80,
					ExposedUrlPath:      convutils.Ref("/prepped-1"),
					ExposedUrlSubDomain: convutils.Ref("prepped-1"),
				},
				"svc-url1",
				"",
				&api.WorkloadNetworkStatusServicesInnerPortsInner{
					TargetPort: "80",
					Port:       8080,
					Name:       "svc-url1-80-8080",
				},
			)

			addToTest(&envs[0], &envsPrepped[0], &urls, &networkStatus,
				"https://b.exposed.com",
				&api.Connection{
					Name:                "url2",
					ConnectionType:      TypeUrl,
					ContainerPort:       81,
					ExposedUrlPath:      convutils.Ref("/prepped-2"),
					ExposedUrlSubDomain: convutils.Ref("prepped-2"),
				},
				"svc-url2",
				"",
				&api.WorkloadNetworkStatusServicesInnerPortsInner{
					TargetPort: "81",
					Port:       8181,
					Name:       "svc-url2-81-8181",
				},
			)

			addToTest(&envs[1], &envsPrepped[1], &urls, &networkStatus,
				"10.10.10.10:1111",
				&api.Connection{
					Name:                "np1",
					ConnectionType:      TypeNodePort,
					ContainerPort:       1,
					ServiceResourceName: convutils.Ref("svc-np"),
					PortResourceName:    convutils.Ref("svc-np-1-1111"),
				},
				"svc-np",
				"",
				&api.WorkloadNetworkStatusServicesInnerPortsInner{
					TargetPort: "1",
					Port:       1111,
					Name:       "svc-np-1-1111",
				},
			)

			addToTest(&envs[1], &envsPrepped[1], &urls, &networkStatus,
				"", // load balancer no url
				&api.Connection{
					Name:           "lb",
					ConnectionType: TypeLoadBalancer,
					ContainerPort:  3,
				},
				"svc-lb",
				"20.20.20.20",
				&api.WorkloadNetworkStatusServicesInnerPortsInner{
					TargetPort: "3",
					Port:       3333,
					Name:       "svc-lb-3-3333",
				},
			)

			addToTest(&envs[1], &envsPrepped[1], &urls, &networkStatus,
				"10.10.10.10:2222",
				&api.Connection{
					Name:                "np2",
					ConnectionType:      TypeNodePort,
					ContainerPort:       2,
					ServiceResourceName: convutils.Ref("svc-np"),
					PortResourceName:    convutils.Ref("svc-np-2-2222"),
				},
				"svc-np",
				"",
				&api.WorkloadNetworkStatusServicesInnerPortsInner{
					TargetPort: "2",
					Port:       2222,
					Name:       "svc-np-2-2222",
				},
			)

			//******************************************************
			//  naive approach
			//******************************************************
			patched := duplicateOf(envs)

			method, err := patchConnections(bg, patched, urls, nil)
			Expect(method).To(Equal(PatchingMethodNaive))
			Expect(err).To(BeNil())

			// mimic the patching
			expected := duplicateOf(envs)

			expected[0].Connections[0].Url = "https://a.exposed.com"
			expected[0].Connections[1].Url = "https://b.exposed.com"
			expected[1].Connections[0].Url = "10.10.10.10:1111"
			expected[1].Connections[1].Url = "" // load balancer
			expected[1].Connections[2].Url = "10.10.10.10:2222"

			testutils.Expect(patched).ToEqual(expected)

			//******************************************************
			// networkStatus/no-prep approach does not support
			// multiple environments so result should be the same
			//******************************************************
			patched = duplicateOf(envs)

			method, err = patchConnections(bg, patched, urls, &networkStatus)
			Expect(method).To(Equal(PatchingMethodNaive))
			Expect(err).To(BeNil())

			testutils.Expect(patched).ToEqual(expected)

			//******************************************************
			// networkStatus/with-prep
			//******************************************************
			patched = duplicateOf(envsPrepped)

			method, err = patchConnections(bg, patched, urls, &networkStatus)
			Expect(method).To(Equal(PatchingMethodNetworkStatPreProcessed))
			Expect(err).To(BeNil())

			expected = duplicateOf(envsPrepped)

			expected[0].Connections[0].Url = "https://prepped-1.exposed.com"
			expected[0].Connections[1].Url = "https://prepped-2.exposed.com"
			expected[1].Connections[0].Url = "10.10.10.10:1111"
			expected[1].Connections[1].Url = "20.20.20.20:3333"
			expected[1].Connections[2].Url = "10.10.10.10:2222"

			testutils.Expect(patched).ToEqual(expected)
		})

		It("1 env x 4 connections, links and urls", func() {
			urls, networkStatus, envs, envsPrepped := initTest(1)

			addToTest(&envs[0], &envsPrepped[0], &urls, &networkStatus,
				"https://exposed.com/1",
				&api.Connection{
					Name:                "url1",
					ConnectionType:      TypeUrl,
					ContainerPort:       80,
					ExposedUrlPath:      convutils.Ref("/prepped-1"),
					ExposedUrlSubDomain: convutils.Ref("prepped-1"),
				},
				"svc-url1",
				"",
				&api.WorkloadNetworkStatusServicesInnerPortsInner{
					TargetPort: "80",
					Port:       8080,
					Name:       "svc-url1-80-8080",
				},
			)

			addToTest(&envs[0], &envsPrepped[0], &urls, &networkStatus,
				"https://exposed.com/2",
				&api.Connection{
					Name:                "url2",
					ConnectionType:      TypeUrl,
					ContainerPort:       81,
					ExposedUrlPath:      convutils.Ref("/prepped-2"),
					ExposedUrlSubDomain: convutils.Ref("prepped-2"),
				},
				"svc-url2",
				"",
				&api.WorkloadNetworkStatusServicesInnerPortsInner{
					TargetPort: "81",
					Port:       8181,
					Name:       "svc-url2-81-8181",
				},
			)

			addToTest(&envs[0], &envsPrepped[0], &urls, &networkStatus,
				"10.10.10.10:2222", // reversed order
				&api.Connection{
					Name:                "np1",
					ConnectionType:      TypeNodePort,
					ContainerPort:       1,
					ServiceResourceName: convutils.Ref("svc-np"),
					PortResourceName:    convutils.Ref("svc-np-1-1111"),
				},
				"svc-np",
				"",
				&api.WorkloadNetworkStatusServicesInnerPortsInner{
					TargetPort: "1",
					Port:       1111,
					Name:       "svc-np-1-1111",
				},
			)

			addToTest(&envs[0], &envsPrepped[0], &urls, &networkStatus,
				"", // load balancer no url
				&api.Connection{
					Name:                "lb1",
					ConnectionType:      TypeLoadBalancer,
					ContainerPort:       3,
					ServiceResourceName: convutils.Ref("svc-lb"),
					PortResourceName:    convutils.Ref("svc-np-81-8181"),
				},
				"svc-lb",
				"20.20.20.20",
				&api.WorkloadNetworkStatusServicesInnerPortsInner{
					TargetPort: "3",
					Port:       3333,
					Name:       "svc-lb-3-3333",
				},
			)

			addToTest(&envs[0], &envsPrepped[0], &urls, &networkStatus,
				"10.10.10.10:1111", // reversed order
				&api.Connection{
					Name:                "np2",
					ConnectionType:      TypeNodePort,
					ContainerPort:       2,
					ServiceResourceName: convutils.Ref("svc-np"),
					PortResourceName:    convutils.Ref("svc-np-2-2222"),
				},
				"svc-np",
				"",
				&api.WorkloadNetworkStatusServicesInnerPortsInner{
					TargetPort: "2",
					Port:       2222,
					Name:       "svc-np-2-2222",
				},
			)

			//******************************************************
			//  naive approach
			//******************************************************
			patched := duplicateOf(envs)

			method, err := patchConnections(bg, patched, urls, nil)
			Expect(method).To(Equal(PatchingMethodNaive))
			Expect(err).To(BeNil())

			// mimic the patching
			expected := duplicateOf(envs)

			expected[0].Connections[0].Url = "https://exposed.com/1"
			expected[0].Connections[1].Url = "https://exposed.com/2"
			expected[0].Connections[2].Url = "10.10.10.10:2222" // this demos the bug, wrong url for this conn
			expected[0].Connections[3].Url = ""                 // load balancer
			expected[0].Connections[4].Url = "10.10.10.10:1111" // also here, wrong url for this conn

			testutils.Expect(patched).ToEqual(expected)

			//******************************************************
			// networkStatus/no-prep the ports will be ordered
			// correctly due to the ports matching
			//******************************************************
			patched = duplicateOf(envs)

			method, err = patchConnections(bg, patched, urls, &networkStatus)
			Expect(method).To(Equal(PatchingMethodNetworkStatNoPreProc))
			Expect(err).To(BeNil())

			expected = duplicateOf(envs)

			expected[0].Connections[0].Url = "https://exposed.com/1"
			expected[0].Connections[1].Url = "https://exposed.com/2"
			expected[0].Connections[2].Url = "10.10.10.10:1111" // this is ok!
			expected[0].Connections[3].Url = "20.20.20.20:3333" // load balancer
			expected[0].Connections[4].Url = "10.10.10.10:2222"

			testutils.Expect(patched).ToEqual(expected)

			//******************************************************
			// networkStatus/prepped the ports will be ordered
			// correctly based on the pre-processed info
			//******************************************************
			patched = duplicateOf(envsPrepped)

			method, err = patchConnections(bg, patched, urls, &networkStatus)
			Expect(method).To(Equal(PatchingMethodNetworkStatPreProcessed))
			Expect(err).To(BeNil())

			expected = duplicateOf(envsPrepped)

			expected[0].Connections[0].Url = "https://exposed.com/prepped-1"
			expected[0].Connections[1].Url = "https://exposed.com/prepped-2"
			expected[0].Connections[2].Url = "10.10.10.10:1111" // this is ok!
			expected[0].Connections[3].Url = "20.20.20.20:3333"
			expected[0].Connections[4].Url = "10.10.10.10:2222"

			testutils.Expect(patched).ToEqual(expected)
		})

		It("all connections in the 2nd environment", func() {
			urls, networkStatus, envs, envsPrepped := initTest(2)

			addToTest(&envs[1], &envsPrepped[1], &urls, &networkStatus,
				"https://exposed.com/1/1",
				&api.Connection{
					Name:                "url1",
					ConnectionType:      TypeUrl,
					ContainerPort:       80,
					ExposedUrlPath:      convutils.Ref("/prepped/1/1"),
					ExposedUrlSubDomain: convutils.Ref("prepped-1-1"),
				},
				"svc-url1",
				"",
				&api.WorkloadNetworkStatusServicesInnerPortsInner{
					TargetPort: "80",
					Port:       8080,
					Name:       "svc-url1-80-8080",
				},
			)

			addToTest(&envs[1], &envsPrepped[1], &urls, &networkStatus,
				"https://exposed.com/2/2",
				&api.Connection{
					Name:                "url2",
					ConnectionType:      TypeUrl,
					ContainerPort:       81,
					ExposedUrlPath:      convutils.Ref("/prepped/2/2"),
					ExposedUrlSubDomain: convutils.Ref("prepped-2-2"),
				},
				"svc-url2",
				"",
				&api.WorkloadNetworkStatusServicesInnerPortsInner{
					TargetPort: "81",
					Port:       8181,
					Name:       "svc-url2-81-8181",
				},
			)

			addToTest(&envs[1], &envsPrepped[1], &urls, &networkStatus,
				"10.10.10.10:1111",
				&api.Connection{
					Name:                "np1",
					ConnectionType:      TypeNodePort,
					ContainerPort:       1,
					ServiceResourceName: convutils.Ref("svc-np"),
					PortResourceName:    convutils.Ref("svc-np-1-1111"),
				},
				"svc-np",
				"",
				&api.WorkloadNetworkStatusServicesInnerPortsInner{
					TargetPort: "1",
					Port:       1111,
					Name:       "svc-np-1-1111",
				},
			)

			addToTest(&envs[1], &envsPrepped[1], &urls, &networkStatus,
				"", // load balancer no url
				&api.Connection{
					Name:                "lb1",
					ConnectionType:      TypeLoadBalancer,
					ContainerPort:       3,
					ServiceResourceName: convutils.Ref("svc-lb"),
					PortResourceName:    convutils.Ref("svc-np-81-8181"),
				},
				"svc-lb",
				"20.20.20.20",
				&api.WorkloadNetworkStatusServicesInnerPortsInner{
					TargetPort: "3",
					Port:       3333,
					Name:       "svc-lb-3-3333",
				},
			)

			addToTest(&envs[1], &envsPrepped[1], &urls, &networkStatus,
				"10.10.10.10:2222",
				&api.Connection{
					Name:                "np2",
					ConnectionType:      TypeNodePort,
					ContainerPort:       2,
					ServiceResourceName: convutils.Ref("svc-np"),
					PortResourceName:    convutils.Ref("svc-np-2-2222"),
				},
				"svc-np",
				"",
				&api.WorkloadNetworkStatusServicesInnerPortsInner{
					TargetPort: "2",
					Port:       2222,
					Name:       "svc-np-2-2222",
				},
			)

			//******************************************************
			//  naive approach
			//******************************************************
			patched := duplicateOf(envs)

			method, err := patchConnections(bg, patched, urls, nil)
			Expect(method).To(Equal(PatchingMethodNaive))
			Expect(err).To(BeNil())

			// mimic the patching
			expected := duplicateOf(envs)

			expected[1].Connections[0].Url = "https://exposed.com/1/1"
			expected[1].Connections[1].Url = "https://exposed.com/2/2"
			expected[1].Connections[2].Url = "10.10.10.10:1111"
			expected[1].Connections[3].Url = "" // load balancer
			expected[1].Connections[4].Url = "10.10.10.10:2222"

			testutils.Expect(patched).ToEqual(expected)

			//******************************************************
			// networkStatus/no-prep approach does not support
			// multiple environments so result should be the same
			//******************************************************
			patched = duplicateOf(envs)

			method, err = patchConnections(bg, patched, urls, &networkStatus)
			Expect(method).To(Equal(PatchingMethodNaive))
			Expect(err).To(BeNil())

			testutils.Expect(patched).ToEqual(expected)

			//******************************************************
			// networkStatus/pre-processed
			//******************************************************
			patched = duplicateOf(envsPrepped)

			method, err = patchConnections(bg, patched, urls, &networkStatus)
			Expect(method).To(Equal(PatchingMethodNetworkStatPreProcessed))
			Expect(err).To(BeNil())

			// mimic the patching
			expected = duplicateOf(envsPrepped)

			expected[1].Connections[0].Url = "https://exposed.com/prepped/1/1"
			expected[1].Connections[1].Url = "https://exposed.com/prepped/2/2"
			expected[1].Connections[2].Url = "10.10.10.10:1111"
			expected[1].Connections[3].Url = "20.20.20.20:3333"
			expected[1].Connections[4].Url = "10.10.10.10:2222"

			testutils.Expect(patched).ToEqual(expected)
		})

		It("all connections in the 1st environment", func() {
			urls, networkStatus, envs, envsPrepped := initTest(2)

			addToTest(&envs[0], &envsPrepped[0], &urls, &networkStatus,
				"https://exposed.com/1/1",
				&api.Connection{
					Name:                "url1",
					ConnectionType:      TypeUrl,
					ContainerPort:       80,
					ExposedUrlPath:      convutils.Ref("/prepped/1/1"),
					ExposedUrlSubDomain: convutils.Ref("prepped-1-1"),
				},
				"svc-url1",
				"",
				&api.WorkloadNetworkStatusServicesInnerPortsInner{
					TargetPort: "80",
					Port:       8080,
					Name:       "svc-url1-80-8080",
				},
			)

			addToTest(&envs[0], &envsPrepped[0], &urls, &networkStatus,
				"https://exposed.com/2/2",
				&api.Connection{
					Name:                "url2",
					ConnectionType:      TypeUrl,
					ContainerPort:       81,
					ExposedUrlPath:      convutils.Ref("/prepped/2/2"),
					ExposedUrlSubDomain: convutils.Ref("prepped-2-2"),
				},
				"svc-url2",
				"",
				&api.WorkloadNetworkStatusServicesInnerPortsInner{
					TargetPort: "81",
					Port:       8181,
					Name:       "svc-url2-81-8181",
				},
			)

			addToTest(&envs[0], &envsPrepped[0], &urls, &networkStatus,
				"10.10.10.10:1111",
				&api.Connection{
					Name:                "np1",
					ConnectionType:      TypeNodePort,
					ContainerPort:       1,
					ServiceResourceName: convutils.Ref("svc-np"),
					PortResourceName:    convutils.Ref("svc-np-1-1111"),
				},
				"svc-np",
				"",
				&api.WorkloadNetworkStatusServicesInnerPortsInner{
					TargetPort: "1",
					Port:       1111,
					Name:       "svc-np-1-1111",
				},
			)

			addToTest(&envs[0], &envsPrepped[0], &urls, &networkStatus,
				"", // load balancer no url
				&api.Connection{
					Name:                "lb1",
					ConnectionType:      TypeLoadBalancer,
					ContainerPort:       3,
					ServiceResourceName: convutils.Ref("svc-lb"),
					PortResourceName:    convutils.Ref("svc-np-81-8181"),
				},
				"svc-lb",
				"20.20.20.20",
				&api.WorkloadNetworkStatusServicesInnerPortsInner{
					TargetPort: "3",
					Port:       3333,
					Name:       "svc-lb-3-3333",
				},
			)

			addToTest(&envs[0], &envsPrepped[0], &urls, &networkStatus,
				"10.10.10.10:2222",
				&api.Connection{
					Name:                "np2",
					ConnectionType:      TypeNodePort,
					ContainerPort:       2,
					ServiceResourceName: convutils.Ref("svc-np"),
					PortResourceName:    convutils.Ref("svc-np-2-2222"),
				},
				"svc-np",
				"",
				&api.WorkloadNetworkStatusServicesInnerPortsInner{
					TargetPort: "2",
					Port:       2222,
					Name:       "svc-np-2-2222",
				},
			)

			//******************************************************
			//  naive approach
			//******************************************************
			patched := duplicateOf(envs)

			method, err := patchConnections(bg, patched, urls, nil)
			Expect(method).To(Equal(PatchingMethodNaive))
			Expect(err).To(BeNil())

			// mimic the patching
			expected := duplicateOf(envs)

			expected[0].Connections[0].Url = "https://exposed.com/1/1"
			expected[0].Connections[1].Url = "https://exposed.com/2/2"
			expected[0].Connections[2].Url = "10.10.10.10:1111"
			expected[0].Connections[3].Url = "" // load balancer
			expected[0].Connections[4].Url = "10.10.10.10:2222"

			testutils.Expect(patched).ToEqual(expected)

			//******************************************************
			// networkStatus/no-prep approach does not support
			// multiple environments so result should be the same
			//******************************************************
			patched = duplicateOf(envs)

			method, err = patchConnections(bg, patched, urls, &networkStatus)
			Expect(method).To(Equal(PatchingMethodNaive))
			Expect(err).To(BeNil())

			testutils.Expect(patched).ToEqual(expected)

			//******************************************************
			// networkStatus/pre-processed
			//******************************************************
			patched = duplicateOf(envsPrepped)

			method, err = patchConnections(bg, patched, urls, &networkStatus)
			Expect(method).To(Equal(PatchingMethodNetworkStatPreProcessed))
			Expect(err).To(BeNil())

			// mimic the patching
			expected = duplicateOf(envsPrepped)

			expected[0].Connections[0].Url = "https://exposed.com/prepped/1/1"
			expected[0].Connections[1].Url = "https://exposed.com/prepped/2/2"
			expected[0].Connections[2].Url = "10.10.10.10:1111"
			expected[0].Connections[3].Url = "20.20.20.20:3333"
			expected[0].Connections[4].Url = "10.10.10.10:2222"

			testutils.Expect(patched).ToEqual(expected)
		})

		It("mixed", func() {
			urls, networkStatus, envs, envsPrepped := initTest(2)

			// env #1

			addToTest(&envs[0], &envsPrepped[0], &urls, &networkStatus,
				"https://exposed.com/1/1",
				&api.Connection{
					Name:                "url1",
					ConnectionType:      TypeUrl,
					ContainerPort:       80,
					ExposedUrlPath:      convutils.Ref("/prepped/1/1"),
					ExposedUrlSubDomain: convutils.Ref("prepped-1-1"),
				},
				"svc-url1",
				"",
				&api.WorkloadNetworkStatusServicesInnerPortsInner{
					TargetPort: "80",
					Port:       8080,
					Name:       "svc-url1-80-8080",
				},
			)

			addToTest(&envs[0], &envsPrepped[0], &urls, &networkStatus,
				"10.10.10.10:1111",
				&api.Connection{
					Name:                "np1",
					ConnectionType:      TypeNodePort,
					ContainerPort:       1,
					ServiceResourceName: convutils.Ref("svc-np"),
					PortResourceName:    convutils.Ref("svc-np-1-1111"),
				},
				"svc-np",
				"",
				&api.WorkloadNetworkStatusServicesInnerPortsInner{
					TargetPort: "1",
					Port:       1111,
					Name:       "svc-np-1-1111",
				},
			)

			addToTest(&envs[0], &envsPrepped[0], &urls, &networkStatus,
				"", // load balancer no url
				&api.Connection{
					Name:                "lb1",
					ConnectionType:      TypeLoadBalancer,
					ContainerPort:       3,
					ServiceResourceName: convutils.Ref("svc-lb"),
					PortResourceName:    convutils.Ref("svc-np-81-8181"),
				},
				"svc-lb",
				"20.20.20.20",
				&api.WorkloadNetworkStatusServicesInnerPortsInner{
					TargetPort: "3",
					Port:       3333,
					Name:       "svc-lb-3-3333",
				},
			)

			// env #2

			addToTest(&envs[1], &envsPrepped[1], &urls, &networkStatus,
				"https://exposed.com/2/2",
				&api.Connection{
					Name:                "url2",
					ConnectionType:      TypeUrl,
					ContainerPort:       81,
					ExposedUrlPath:      convutils.Ref("/prepped/2/2"),
					ExposedUrlSubDomain: convutils.Ref("prepped-2-2"),
				},
				"svc-url2",
				"",
				&api.WorkloadNetworkStatusServicesInnerPortsInner{
					TargetPort: "81",
					Port:       8181,
					Name:       "svc-url2-81-8181",
				},
			)

			addToTest(&envs[1], &envsPrepped[1], &urls, &networkStatus,
				"10.10.10.10:2222",
				&api.Connection{
					Name:                "np2",
					ConnectionType:      TypeNodePort,
					ContainerPort:       2,
					ServiceResourceName: convutils.Ref("svc-np"),
					PortResourceName:    convutils.Ref("svc-np-2-2222"),
				},
				"svc-np",
				"",
				&api.WorkloadNetworkStatusServicesInnerPortsInner{
					TargetPort: "2",
					Port:       2222,
					Name:       "svc-np-2-2222",
				},
			)

			//******************************************************
			//  naive approach
			//******************************************************
			patched := duplicateOf(envs)

			method, err := patchConnections(bg, patched, urls, nil)
			Expect(method).To(Equal(PatchingMethodNaive))
			Expect(err).To(BeNil())

			// mimic the patching
			expected := duplicateOf(envs)

			expected[0].Connections[0].Url = "https://exposed.com/1/1"
			expected[0].Connections[1].Url = "10.10.10.10:1111"
			expected[0].Connections[2].Url = "" // load balancer

			expected[1].Connections[0].Url = "https://exposed.com/2/2"
			expected[1].Connections[1].Url = "10.10.10.10:2222"

			testutils.Expect(patched).ToEqual(expected)

			//******************************************************
			// networkStatus/no-prep approach does not support
			// multiple environments so result should be the same
			//******************************************************
			patched = duplicateOf(envs)

			method, err = patchConnections(bg, patched, urls, &networkStatus)
			Expect(method).To(Equal(PatchingMethodNaive))
			Expect(err).To(BeNil())

			testutils.Expect(patched).ToEqual(expected)

			//******************************************************
			//  networkStatus/pre-processed
			//******************************************************
			patched = duplicateOf(envsPrepped)

			method, err = patchConnections(bg, patched, urls, &networkStatus)
			Expect(method).To(Equal(PatchingMethodNetworkStatPreProcessed))
			Expect(err).To(BeNil())

			// mimic the patching
			expected = duplicateOf(envsPrepped)

			expected[0].Connections[0].Url = "https://exposed.com/prepped/1/1"
			expected[0].Connections[1].Url = "10.10.10.10:1111"
			expected[0].Connections[2].Url = "20.20.20.20:3333"

			expected[1].Connections[0].Url = "https://exposed.com/prepped/2/2"
			expected[1].Connections[1].Url = "10.10.10.10:2222"

			testutils.Expect(patched).ToEqual(expected)
		})

		It("mismatch", func() {
			urls, networkStatus, envs, envsPrepped := initTest(1)

			addToTest(&envs[0], &envsPrepped[0], &urls, &networkStatus,
				"",
				&api.Connection{
					Name:                "url1",
					ConnectionType:      TypeUrl,
					ContainerPort:       80,
					ExposedUrlPath:      convutils.Ref("/prepped/1"),
					ExposedUrlSubDomain: convutils.Ref("prepped-1"),
				},
				"svc-url1",
				"",
				&api.WorkloadNetworkStatusServicesInnerPortsInner{
					TargetPort: "80",
					Port:       8080,
					Name:       "svc-url1-80-8080",
				},
			)

			addToTest(&envs[0], &envsPrepped[0], &urls, &networkStatus,
				"https://exposed.com/1", // this one should go to the first conn
				&api.Connection{
					Name:           "url2",
					ConnectionType: TypeUrl,
					ContainerPort:  81,
				},
				"svc-url2",
				"",
				&api.WorkloadNetworkStatusServicesInnerPortsInner{
					TargetPort: "81",
					Port:       8181,
					Name:       "svc-url2-80-8080",
				},
			)

			//******************************************************
			//  naive approach
			//******************************************************
			patched := duplicateOf(envs)

			method, err := patchConnections(bg, patched, urls, nil)
			Expect(method).To(Equal(PatchingMethodNaive))
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("urls count is lower than total count of connections"))

			//******************************************************
			// networkStatus/no-prep approach
			//******************************************************
			patched = duplicateOf(envs)

			method, err = patchConnections(bg, patched, urls, &networkStatus)
			Expect(method).To(Equal(PatchingMethodNetworkStatNoPreProc))
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("urls count is lower than total count of exposedUrl connections"))

			//******************************************************
			// networkStatus/prepped approach - we match the first
			//******************************************************
			patched = duplicateOf(envsPrepped)

			method, err = patchConnections(bg, patched, urls, &networkStatus)
			Expect(method).To(Equal(PatchingMethodNetworkStatPreProcessed))
			Expect(err).To(BeNil())

			// mimic the patching
			expected := duplicateOf(envsPrepped)

			expected[0].Connections[0].Url = "https://exposed.com/prepped/1"

			testutils.Expect(patched).ToEqual(expected)
		})

		It("error parsing url", func() {
			urls, networkStatus, envs, envsPrepped := initTest(1)

			addToTest(&envs[0], &envsPrepped[0], &urls, &networkStatus,
				"https://exposed.com/1",
				&api.Connection{
					Name:                "url1",
					ConnectionType:      TypeUrl,
					ContainerPort:       80,
					ExposedUrlPath:      convutils.Ref("/prepped/1"),
					ExposedUrlSubDomain: convutils.Ref("prepped/1"),
				},
				"svc-url1",
				"",
				&api.WorkloadNetworkStatusServicesInnerPortsInner{
					TargetPort: "80",
					Port:       8080,
					Name:       "svc-url1-80-8080",
				},
			)

			addToTest(&envs[0], &envsPrepped[0], &urls, &networkStatus,
				malformedUrl,
				&api.Connection{
					Name:                "url2",
					ConnectionType:      TypeUrl,
					ContainerPort:       81,
					ExposedUrlPath:      convutils.Ref("/prepped/2"),
					ExposedUrlSubDomain: convutils.Ref("prepped/2"),
				},
				"svc-url2",
				"",
				&api.WorkloadNetworkStatusServicesInnerPortsInner{
					TargetPort: "81",
					Port:       8181,
					Name:       "svc-url2-81-8181",
				},
			)

			addToTest(&envs[0], &envsPrepped[0], &urls, &networkStatus,
				"10.10.10.10:1111",
				&api.Connection{
					Name:                "np1",
					ConnectionType:      TypeNodePort,
					ContainerPort:       1,
					ServiceResourceName: convutils.Ref("svc-np"),
					PortResourceName:    convutils.Ref("svc-np-1-1111"),
				},
				"svc-np",
				"",
				&api.WorkloadNetworkStatusServicesInnerPortsInner{
					TargetPort: "1",
					Port:       1111,
					Name:       "svc-np-1-1111",
				},
			)

			addToTest(&envs[0], &envsPrepped[0], &urls, &networkStatus,
				"", // load balancer no url
				&api.Connection{
					Name:                "lb1",
					ConnectionType:      TypeLoadBalancer,
					ContainerPort:       3,
					ServiceResourceName: convutils.Ref("svc-lb"),
					PortResourceName:    convutils.Ref("svc-np-3-3333"),
				},
				"svc-lb",
				"20.20.20.20",
				nil,
			)

			addToTest(&envs[0], &envsPrepped[0], &urls, &networkStatus,
				"", // load balancer no url
				&api.Connection{
					Name:                "lb2",
					ConnectionType:      TypeLoadBalancer,
					ContainerPort:       4,
					ServiceResourceName: convutils.Ref("svc-lb"),
					PortResourceName:    convutils.Ref("svc-np-4-4444"),
				},
				"svc-lb",
				"20.20.20.20",
				&api.WorkloadNetworkStatusServicesInnerPortsInner{
					TargetPort: "4",
					Port:       4444,
					Name:       "svc-lb-4-4444",
				},
			)

			addToTest(&envs[0], &envsPrepped[0], &urls, &networkStatus,
				"10.10.10.10:5555", // non related url
				&api.Connection{
					Name:                "np2",
					ConnectionType:      TypeNodePort,
					ContainerPort:       2,
					ServiceResourceName: convutils.Ref("svc-np"),
					PortResourceName:    convutils.Ref("svc-np-2-2222"),
				},
				"svc-np",
				"",
				&api.WorkloadNetworkStatusServicesInnerPortsInner{
					TargetPort: "2",
					Port:       2222,
					Name:       "svc-np-2-2222",
				},
			)

			//******************************************************
			// networkStatus/no-prep
			//******************************************************
			patched := duplicateOf(envs)

			method, err := patchConnections(bg, patched, urls, &networkStatus)
			Expect(method).To(Equal(PatchingMethodNetworkStatNoPreProc))
			Expect(err).To(BeNil())

			expected := duplicateOf(envs)

			expected[0].Connections[0].Url = "https://exposed.com/1"
			expected[0].Connections[1].Url = malformedUrl       // in this method we do not check the urls, just copy them
			expected[0].Connections[2].Url = "10.10.10.10:1111" // node-port for target port 1
			expected[0].Connections[3].Url = ""                 // the load balancer 3 that is missing
			expected[0].Connections[4].Url = "20.20.20.20:4444" // the load balancer that had all the info
			expected[0].Connections[5].Url = ""                 // there was no url for this node-port (target port 2)
			testutils.Expect(patched).ToEqual(expected)

			//******************************************************
			// networkStatus/prep
			//******************************************************
			patched = duplicateOf(envsPrepped)

			method, err = patchConnections(bg, patched, urls, &networkStatus)
			Expect(method).To(Equal(PatchingMethodNetworkStatPreProcessed))
			Expect(err).To(BeNil())

			expected = duplicateOf(envsPrepped)

			expected[0].Connections[0].Url = "https://exposed.com/prepped/1"
			expected[0].Connections[1].Url = "https://exposed.com/prepped/2" // we don't need the 2nd url, we know the domain from the 1st
			expected[0].Connections[2].Url = "10.10.10.10:1111"              // node-port for target port 1
			expected[0].Connections[3].Url = ""                              // load balancer that is missing
			expected[0].Connections[4].Url = "20.20.20.20:4444"              // the load balancer that had all the info
			expected[0].Connections[5].Url = "10.10.10.10:2222"              // we don't need the specific IP address.

			testutils.Expect(patched).ToEqual(expected)
		})
	})

	It("error parsing url - first url corrupter", func() {
		urls, networkStatus, envs, envsPrepped := initTest(1)

		addToTest(&envs[0], &envsPrepped[0], &urls, &networkStatus,
			malformedUrl,
			&api.Connection{
				Name:                "url1",
				ConnectionType:      TypeUrl,
				ContainerPort:       80,
				ExposedUrlPath:      convutils.Ref("/prepped/1"),
				ExposedUrlSubDomain: convutils.Ref("prepped/1"),
			},
			"svc-url1",
			"",
			&api.WorkloadNetworkStatusServicesInnerPortsInner{
				TargetPort: "80",
				Port:       8080,
				Name:       "svc-url1-80-8080",
			},
		)

		addToTest(&envs[0], &envsPrepped[0], &urls, &networkStatus,
			"https://exposed.com/2",
			&api.Connection{
				Name:                "url2",
				ConnectionType:      TypeUrl,
				ContainerPort:       81,
				ExposedUrlPath:      convutils.Ref("/prepped/2"),
				ExposedUrlSubDomain: convutils.Ref("prepped/2"),
			},
			"svc-url2",
			"",
			&api.WorkloadNetworkStatusServicesInnerPortsInner{
				TargetPort: "81",
				Port:       8181,
				Name:       "svc-url2-81-8181",
			},
		)

		addToTest(&envs[0], &envsPrepped[0], &urls, &networkStatus,
			"10.10.10.10:1111",
			&api.Connection{
				Name:                "np1",
				ConnectionType:      TypeNodePort,
				ContainerPort:       1,
				ServiceResourceName: convutils.Ref("svc-np"),
				PortResourceName:    convutils.Ref("svc-np-1-1111"),
			},
			"svc-np",
			"",
			&api.WorkloadNetworkStatusServicesInnerPortsInner{
				TargetPort: "1",
				Port:       1111,
				Name:       "svc-np-1-1111",
			},
		)

		addToTest(&envs[0], &envsPrepped[0], &urls, &networkStatus,
			"10.10.10.10:5555", // non related url
			&api.Connection{
				Name:                "np2",
				ConnectionType:      TypeNodePort,
				ContainerPort:       2,
				ServiceResourceName: convutils.Ref("svc-np"),
				PortResourceName:    convutils.Ref("svc-np-2-2222"),
			},
			"svc-np",
			"",
			&api.WorkloadNetworkStatusServicesInnerPortsInner{
				TargetPort: "2",
				Port:       2222,
				Name:       "svc-np-2-2222",
			},
		)

		//******************************************************
		// any method
		//******************************************************
		patched := duplicateOf(envs)

		method, err := patchConnections(bg, patched, urls, &networkStatus)
		Expect(method).To(Equal(PatchingMethodUndefined))
		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(ContainSubstring("is malformed for cluster supporting subdomain"))
	})
})

func nodePortServiceStatusURLFixture() ([]api.Environment, *api.WorkloadNetworkStatus) {
	envs := []api.Environment{
		{
			Connections: []api.Connection{
				{
					Name:                "jupyter",
					ConnectionType:      TypeNodePort,
					ServiceResourceName: convutils.Ref("svc-nodeport"),
					PortResourceName:    convutils.Ref("jupyter-port"),
					ContainerPort:       8888,
				},
				{
					Name:                "tensorboard",
					ConnectionType:      TypeNodePort,
					ServiceResourceName: convutils.Ref("svc-nodeport"),
					PortResourceName:    convutils.Ref("tensorboard-port"),
					ContainerPort:       6006,
				},
			},
		},
	}
	networkStatus := &api.WorkloadNetworkStatus{
		Services: []api.WorkloadNetworkStatusServicesInner{
			{
				Name: "svc-nodeport",
				Type: "NodePort",
				Ports: []api.WorkloadNetworkStatusServicesInnerPortsInner{
					{
						Name:       "jupyter-port",
						Port:       32101,
						TargetPort: "8888",
						Protocol:   "TCP",
					},
					{
						Name:       "tensorboard-port",
						Port:       32102,
						TargetPort: "6006",
						Protocol:   "TCP",
					},
				},
				Urls: []api.WorkloadNetworkStatusServicesInnerUrlsInner{
					{
						Url:        "10.0.0.1:32101",
						SourceKind: "Service",
						SourceName: "svc-nodeport",
					},
					{
						Url:        "10.0.0.1:32102",
						SourceKind: "Service",
						SourceName: "svc-nodeport",
					},
				},
			},
		},
	}

	return envs, networkStatus
}

func duplicateOf(envs []api.Environment) []api.Environment {
	result := make([]api.Environment, len(envs))

	for idx := range envs {
		result[idx].Connections = []api.Connection{}

		err := convutils.CopySpecs(&result[idx], envs[idx])
		Expect(err).To(BeNil())
	}

	return result
}

func initTest(numEnvironments int) ([]string, api.WorkloadNetworkStatus, []api.Environment, []api.Environment) {
	urls := make([]string, 0)

	networkStatus := api.WorkloadNetworkStatus{
		Services: []api.WorkloadNetworkStatusServicesInner{},
	}

	envs := make([]api.Environment, numEnvironments)
	envsPrepped := make([]api.Environment, numEnvironments)

	for i := 0; i < numEnvironments; i++ {
		envs[i].Connections = []api.Connection{}
		envsPrepped[i].Connections = []api.Connection{}
	}

	return urls, networkStatus, envs, envsPrepped
}

func addToTest(env *api.Environment, envPrepped *api.Environment, urls *[]string, networkStatus *api.WorkloadNetworkStatus,
	url string,
	preppedConn *api.Connection,
	serviceName string,
	loadBalancerIp string,
	port *api.WorkloadNetworkStatusServicesInnerPortsInner,
) {
	if url != "" {
		*urls = append(*urls, url)
	}

	if preppedConn != nil {
		envPrepped.Connections = append(envPrepped.Connections, *preppedConn)

		nonPrepped := *preppedConn
		nonPrepped.ExposedUrlPath = nil
		nonPrepped.ExposedUrlSubDomain = nil
		nonPrepped.PortResourceName = nil
		nonPrepped.ServiceResourceName = nil

		env.Connections = append(env.Connections, nonPrepped)
	}

	if serviceName == "" {
		return
	}

	var service *api.WorkloadNetworkStatusServicesInner

	for idx, entry := range networkStatus.Services {
		if entry.Name == serviceName {
			service = &networkStatus.Services[idx]
			break
		}
	}

	if service == nil {
		networkStatus.Services = append(networkStatus.Services, api.WorkloadNetworkStatusServicesInner{
			Name:  serviceName,
			Ports: []api.WorkloadNetworkStatusServicesInnerPortsInner{},
		})

		service = &networkStatus.Services[len(networkStatus.Services)-1]
	}

	if loadBalancerIp != "" {
		service.LoadBalancer = api.WorkloadNetworkStatusServicesInnerLoadBalancer{
			Ingress: []api.WorkloadNetworkStatusServicesInnerLoadBalancerIngressInner{
				{
					Ip: loadBalancerIp,
				},
			},
		}
	}

	if port != nil {
		service.Ports = append(service.Ports, *port)
	}
}
