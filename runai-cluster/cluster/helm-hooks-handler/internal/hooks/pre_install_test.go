package hooks_test

import (
	"context"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/common"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	ocpv1 "github.com/openshift/api/operator/v1"
	scheme_init "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	"github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/clients"
	"github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/hooks"
	pre_hooks "github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/hooks/pre-hooks"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type Version struct {
	Version string `json:"version"`
}

var _ = Describe("PreInstall Hook", func() {
	var (
		resources *clients.Clients
		server    *ghttp.Server
	)

	BeforeEach(func() {
		resources = &clients.Clients{
			LocalClusterClient: fake.NewClientBuilder().WithScheme(scheme_init.Scheme()).Build(),
		}
		resources.RemoteClusterClient = resources.LocalClusterClient
		server = ghttp.NewServer()
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/_version"),
				ghttp.RespondWithJSONEncoded(http.StatusOK, Version{
					Version: "2.22.5",
				}),
			),
		)
	})

	It("should not perform openshift changes if not openshift", func() {
		ns := &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:   runv1.DefaultNamespace,
				Labels: map[string]string{},
			},
		}
		Expect(resources.LocalClusterClient.Create(context.Background(), ns)).To(Succeed())
		hook := pre_hooks.NewPreInstallHook(&hooks.ClusterHookParams{
			ControlPlaneDetails: clients.ControlPlaneDetails{
				URL: server.URL(),
			},
			ClusterDetails: hooks.ClusterDetails{
				Version: "2.21.8",
			},
		}, &common.CustomCA{})
		Expect(hook.Call(context.Background(), resources)).To(Succeed())

		Expect(resources.LocalClusterClient.Get(context.Background(), rclient.ObjectKeyFromObject(ns), ns)).NotTo(HaveOccurred())
		Expect(ns.Labels["openshift.io/cluster-monitoring"]).To(Equal(""))
	})

	It("should peform openshift changes", func() {
		OcpIngressOperatorNS := &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "openshift-ingress-operator",
			},
		}
		Expect(resources.LocalClusterClient.Create(context.Background(), OcpIngressOperatorNS)).To(Succeed())

		ocpIngressDefaultController := &ocpv1.IngressController{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "default",
				Namespace: "openshift-ingress-operator",
			},
		}
		Expect(resources.LocalClusterClient.Create(context.Background(), ocpIngressDefaultController)).To(Succeed())

		ns := &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: runv1.DefaultNamespace,
			},
		}
		Expect(resources.LocalClusterClient.Create(context.Background(), ns)).To(Succeed())
		hook := pre_hooks.NewPreInstallHook(&hooks.ClusterHookParams{
			ControlPlaneDetails: clients.ControlPlaneDetails{
				URL: server.URL(),
			},
			ClusterDetails: hooks.ClusterDetails{
				Version: "2.21.8",
			},
		}, &common.CustomCA{})
		Expect(hook.Call(context.Background(), resources)).To(Succeed())
		err := resources.LocalClusterClient.Get(context.Background(), rclient.ObjectKeyFromObject(ocpIngressDefaultController),
			ocpIngressDefaultController)
		Expect(err).NotTo(HaveOccurred())
		Expect(ocpIngressDefaultController.Spec.RouteAdmission.NamespaceOwnership).
			To(Equal(ocpv1.InterNamespaceAllowedOwnershipCheck))

		err = resources.LocalClusterClient.Get(context.Background(), rclient.ObjectKeyFromObject(ns), ns)
		Expect(err).NotTo(HaveOccurred())
		Expect(ns.Labels["openshift.io/cluster-monitoring"]).To(Equal("true"))

	})
})
