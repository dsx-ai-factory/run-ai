package steps

import (
	"strings"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/actions"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
)

const (
	serverlessNamespace = "openshift-serverless"
	serverlessPackage   = "serverless-operator"
	// serverlessSentinel is a controller Deployment the OpenShift Serverless
	// operator runs in its namespace, used for idempotency detection.
	serverlessSentinel = "knative-operator"

	knativeServingAPIVersion = "operator.knative.dev/v1beta1"
	knativeServingKind       = "KnativeServing"
)

// KnativeOpenShiftParams configures the OpenShift Knative flow. Channels and the
// KnativeServing override are resolved by the planner.
type KnativeOpenShiftParams struct {
	ServerlessChannel string
	// ServerlessVersion is the resolved Serverless operator version; it pins the
	// Subscription's startingCSV (serverless-operator.<version>) so the
	// OperatorHub install lands on a known version rather than the channel head.
	ServerlessVersion string
	// InstallPlanApproval is the OLM upgrade strategy for the Serverless operator.
	InstallPlanApproval string
	Cluster             *spec.Cluster
	CRSpec              map[string]any
	SubscriptionEnv     map[string]string
	Catalog             actions.OLMCatalog
}

// NewKnativeOpenShift installs Knative on OpenShift via the Red Hat OpenShift
// Serverless operator (OperatorHub): the operator (cluster-wide), the
// knative-serving namespace, and the KnativeServing CR built from the operator's
// CSV default with the inference wildcard domain injected and any user override
// merged on top. Unlike the upstream Helm flow, no fronting Ingress/Kourier
// service is created — OpenShift Serverless exposes serving through OpenShift
// Routes; custom certificates/exposure are configured through customResource.
func NewKnativeOpenShift(p KnativeOpenShiftParams, deps orchestrator2.Deps) []orchestrator2.Step {
	operator := NewOLMOperator(OLMOperator{
		StepName:            "install-serverless-operator",
		Namespace:           serverlessNamespace,
		Package:             serverlessPackage,
		Channel:             p.ServerlessChannel,
		StartingCSV:         startingCSV(serverlessPackage, p.ServerlessVersion),
		InstallPlanApproval: p.InstallPlanApproval,
		Catalog:             p.Catalog,
		Sentinel:            deploymentSentinel(serverlessSentinel, serverlessNamespace),
		AllNamespaces:       true,
		SubscriptionEnv:     p.SubscriptionEnv,
	}, deps)

	namespace := NewNamespace("create-knative-serving-namespace", knativeServingNamespace, deps)

	knativeServing := NewOperatorDefaultCR(OperatorCR{
		StepName:         "install-knative-serving",
		CRD:              knativeServingCRD,
		CRAPIVersion:     knativeServingAPIVersion,
		CRKind:           knativeServingKind,
		CRName:           knativeServingName,
		CRNamespace:      knativeServingNamespace,
		CSVNamespace:     serverlessNamespace,
		SubscriptionName: serverlessPackage,
		Override:         knativeServingOverride(p.Cluster, p.CRSpec),
	}, deps)

	return []orchestrator2.Step{operator, namespace, knativeServing}
}

// knativeServingOverride injects the inference wildcard serving domain
// (config.domain) and the Run:ai-required Serving feature flags (config.features,
// keyed by the bare ConfigMap name per the KnativeServing operator convention)
// into the KnativeServing spec, then merges the user's customResource on top so
// the user override wins over these, which win over the operator's CSV default.
//
// The feature flags are essential: Run:ai's cluster operator marks Knative
// available (and the UI offers "standard" inference) only when they are enabled.
// Without them OpenShift Serverless keeps its defaults (affinity/schedulername
// disabled) and Run:ai treats Knative as not installed.
func knativeServingOverride(c *spec.Cluster, userOverride map[string]any) map[string]any {
	base := map[string]any{
		"config": map[string]any{
			"features": knativeServingFeatures(),
		},
	}
	if servingDomain := strings.TrimPrefix(inferenceWildcardDomain(c), "*."); servingDomain != "" {
		base["config"].(map[string]any)["domain"] = map[string]any{servingDomain: ""}
	}

	return actions.MergeCRSpec(base, userOverride)
}
