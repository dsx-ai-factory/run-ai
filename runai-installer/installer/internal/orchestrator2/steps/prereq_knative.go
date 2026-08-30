package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/actions"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	knativeOperatorRelease   = "knative-operator"
	knativeOperatorChart     = "knative-operator"
	knativeOperatorNamespace = "knative-operator"
	knativeOperatorRepo      = "https://knative.github.io/operator"
	knativeOperatorSentinel  = "knative-operator"

	knativeServingNamespace = "knative-serving"
	// knativeServingName is the name of both the KnativeServing CR and the
	// fronting Ingress the operator reconciles into the serving deployments.
	knativeServingName = "knative-serving"
	// knativeInferenceSecret is the wildcard TLS secret the fronting Ingress
	// terminates inference traffic with (matching the legacy installer).
	knativeInferenceSecret = "runai-cluster-inference-tls"
	// knativeServingCRD is the operator CRD that must be established before the
	// KnativeServing CR can be applied.
	knativeServingCRD = "knativeservings.operator.knative.dev"
	// knativeServingHPADeployment is the serving-hpa autoscaler the HPA manifest
	// adds; used as the step sentinel.
	knativeServingHPADeployment = "autoscaler-hpa"

	// knativeServingHPAManifest is the upstream serving-hpa overlay for a
	// release tag. The tag is the bare semver (e.g. "1.18.2"); the upstream
	// release is prefixed "knative-v".
	knativeServingHPAManifest = "https://github.com/knative/serving/releases/download/knative-v%s/serving-hpa.yaml"
)

// NewKnativeSteps builds the ordered steps that install Knative Serving as a
// Run:ai inference prerequisite, reusing the shared namespace/TLS-secret steps
// and the prereq-step sentinel detection. The flow is:
//
//  1. install the Knative operator (Helm)
//  2. create the knative-serving namespace
//  3. apply the wildcard TLS secret the fronting Ingress terminates with
//  4. apply the KnativeServing CR (after the operator CRD is established) and
//     the fronting Ingress
//  5. apply the serving-HPA autoscaler manifest
//
// The wildcard domain and certificate come from the cluster's Inference block
// (input validation enforces they are present when Knative is enabled).
//
// gateway selects the front-door: when non-nil the serving step attaches an
// HTTPRoute to that existing Gateway (TLS terminated by the Gateway listener, so
// no wildcard secret is created); otherwise it applies a fronting Ingress of the
// given class that terminates TLS with the wildcard secret.
func NewKnativeSteps(version, ingressClass string, gateway *spec.GatewayAPIConnection, c *spec.Cluster, tls *spec.TLS, deps orchestrator2.Deps) []orchestrator2.Step {
	operator := actions.HelmInstall{
		RepoURL:     knativeOperatorRepo,
		Version:     version,
		ReleaseName: knativeOperatorRelease,
		ChartName:   knativeOperatorChart,
		Namespace:   knativeOperatorNamespace,
	}

	wildcardDomain := inferenceWildcardDomain(c)

	out := []orchestrator2.Step{
		newPrereqStep("install-knative-operator", deploymentSentinel(knativeOperatorSentinel, knativeOperatorNamespace), deps, operator),
		NewNamespace("create-knative-serving-namespace", knativeServingNamespace, deps),
	}

	// The wildcard TLS secret is only the Ingress front-door's concern. Under
	// Gateway API the Gateway listener terminates TLS, so no secret is created.
	if gateway == nil {
		if tls != nil && tls.CertPath != "" {
			out = append(out, NewTLSSecret("install-knative-serving-tls", knativeInferenceSecret, knativeServingNamespace, tls.CertPath, tls.KeyPath, deps))
		}
	}

	out = append(out,
		newKnativeServingStep(wildcardDomain, ingressClass, gateway, deps),
		newPrereqStep("install-knative-serving-hpa", deploymentSentinel(knativeServingHPADeployment, knativeServingNamespace), deps,
			actions.ApplyRemoteManifest{URL: fmt.Sprintf(knativeServingHPAManifest, version)}),
	)
	return out
}

// knativeServingStep applies the KnativeServing CR and its fronting Ingress.
// Detect keys off the KnativeServing CR (a user who configured serving by other
// means is recognized and skipped). BuildActions waits for the operator CRD to
// be established and the operator Deployment to be available before applying
// the CR, so the apply races neither the operator's asynchronous CRD
// registration nor its admission webhook coming up.
type knativeServingStep struct {
	*orchestrator2.BaseStep
	deps           orchestrator2.Deps
	wildcardDomain string
	ingressClass   string
	gateway        *spec.GatewayAPIConnection
}

func newKnativeServingStep(wildcardDomain, ingressClass string, gateway *spec.GatewayAPIConnection, deps orchestrator2.Deps) *knativeServingStep {
	return &knativeServingStep{
		BaseStep:       orchestrator2.NewBaseStep("install-knative-serving"),
		deps:           deps,
		wildcardDomain: wildcardDomain,
		ingressClass:   ingressClass,
		gateway:        gateway,
	}
}

func (s *knativeServingStep) Detect(ctx context.Context, _ *spec.Spec) (orchestrator2.DetectResult, error) {
	cr := knativeServingObject()
	err := s.deps.K8sExecutor.Get(ctx, client.ObjectKey{Namespace: knativeServingNamespace, Name: knativeServingName}, cr)
	if err == nil {
		return orchestrator2.DetectResult{Phase: orchestrator2.DetectInstalled}, nil
	}
	if apierrors.IsNotFound(err) {
		return orchestrator2.DetectResult{Phase: orchestrator2.DetectNotInstalled}, nil
	}
	return orchestrator2.DetectResult{}, err
}

func (s *knativeServingStep) BuildActions(*spec.Spec) ([]orchestrator2.PerformAction, error) {
	frontDoor := knativeServingIngress(s.wildcardDomain, s.ingressClass)
	if s.gateway != nil {
		frontDoor = knativeServingHTTPRoute(s.wildcardDomain, s.gateway)
	}
	return []orchestrator2.PerformAction{
		actions.WaitForCRDEstablished{Name: knativeServingCRD},
		actions.WaitForDeploymentAvailable{Namespace: knativeOperatorNamespace, Name: knativeOperatorSentinel},
		actions.K8sApply{Object: knativeServingCR(s.wildcardDomain)},
		actions.K8sApply{Object: frontDoor},
	}, nil
}

// knativeServingObject is the empty CR used for the detection Get.
func knativeServingObject() *unstructured.Unstructured {
	cr := &unstructured.Unstructured{}
	cr.SetAPIVersion("operator.knative.dev/v1beta1")
	cr.SetKind("KnativeServing")
	cr.SetNamespace(knativeServingNamespace)
	cr.SetName(knativeServingName)
	return cr
}

// knativeServingCR builds the KnativeServing custom resource the operator
// reconciles into the serving deployments. The serving domain is the wildcard
// domain with its leading "*." stripped, and Kourier is the ingress the
// operator wires up internally.
func knativeServingCR(wildcardDomain string) *unstructured.Unstructured {
	servingDomain := strings.TrimPrefix(wildcardDomain, "*.")
	cr := knativeServingObject()
	cr.Object["spec"] = map[string]any{
		"config": map[string]any{
			"autoscaler": map[string]any{
				"enable-scale-to-zero": "true",
			},
			"features": knativeServingFeatures(),
			"domain": map[string]any{
				servingDomain: "",
			},
			"network": map[string]any{
				"domainTemplate":          "{{.Name}}{{.Namespace}}.{{.Domain}}",
				"ingress-class":           "kourier.ingress.networking.knative.dev",
				"default-external-scheme": "https",
			},
		},
		"high-availability": map[string]any{
			"replicas": int64(1),
		},
		"ingress": map[string]any{
			"kourier": map[string]any{
				"enabled":      true,
				"service-type": "ClusterIP",
			},
		},
	}
	return cr
}

// knativeServingFeatures are the Knative Serving feature flags the KnativeServing
// CR enables (written into the config-features ConfigMap by the operator). Run:ai's
// cluster operator gates "standard" (Knative-backed) inference on these: it marks
// Knative available only when the required flags are enabled (notably
// kubernetes.podspec-affinity and kubernetes.podspec-schedulername). Shared by the
// upstream and OpenShift Serverless install paths so the two cannot drift and leave
// Knative undetected on one of them.
func knativeServingFeatures() map[string]any {
	return map[string]any{
		"kubernetes.podspec-affinity":                "enabled",
		"kubernetes.podspec-init-containers":         "enabled",
		"kubernetes.podspec-persistent-volume-claim": "enabled",
		"kubernetes.podspec-persistent-volume-write": "enabled",
		"kubernetes.podspec-schedulername":           "enabled",
		"kubernetes.podspec-securitycontext":         "enabled",
		"kubernetes.podspec-tolerations":             "enabled",
		"kubernetes.podspec-volumes-emptydir":        "enabled",
		"multi-container":                            "enabled",
	}
}

// knativeServingIngress builds the cluster Ingress that fronts the Kourier
// gateway and terminates wildcard TLS for inference traffic.
func knativeServingIngress(wildcardDomain, ingressClass string) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "networking.k8s.io/v1",
		"kind":       "Ingress",
		"metadata": map[string]any{
			"name":      knativeServingName,
			"namespace": knativeServingNamespace,
		},
		"spec": map[string]any{
			"ingressClassName": ingressClass,
			"rules": []any{
				map[string]any{
					"host": wildcardDomain,
					"http": map[string]any{
						"paths": []any{
							map[string]any{
								"path":     "/",
								"pathType": "Prefix",
								"backend": map[string]any{
									"service": map[string]any{
										"name": "kourier",
										"port": map[string]any{"number": int64(80)},
									},
								},
							},
						},
					},
				},
			},
			"tls": []any{
				map[string]any{
					"hosts":      []any{wildcardDomain},
					"secretName": knativeInferenceSecret,
				},
			},
		},
	}}
}

// knativeServingHTTPRoute builds the HTTPRoute that attaches to the existing
// platform Gateway and routes the inference wildcard host to the Kourier
// gateway service. TLS is terminated by the Gateway listener (which must cover
// the wildcard host), so unlike the Ingress front-door this carries no TLS
// block. The backend Kourier service is in the same namespace as the route, so
// no ReferenceGrant is needed.
func knativeServingHTTPRoute(wildcardDomain string, gw *spec.GatewayAPIConnection) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "gateway.networking.k8s.io/v1",
		"kind":       "HTTPRoute",
		"metadata": map[string]any{
			"name":      knativeServingName,
			"namespace": knativeServingNamespace,
		},
		"spec": map[string]any{
			"parentRefs": []any{
				map[string]any{
					"name":      gw.Name,
					"namespace": gw.Namespace,
				},
			},
			"hostnames": []any{wildcardDomain},
			"rules": []any{
				map[string]any{
					"backendRefs": []any{
						map[string]any{
							"name": "kourier",
							"port": int64(80),
						},
					},
				},
			},
		},
	}}
}

// inferenceWildcardDomain returns the wildcard domain configured for inference,
// or empty when unset.
func inferenceWildcardDomain(c *spec.Cluster) string {
	if c == nil || c.Inference == nil {
		return ""
	}
	return c.Inference.WildcardDomain
}
