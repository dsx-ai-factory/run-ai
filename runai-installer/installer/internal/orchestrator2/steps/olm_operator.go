package steps

import (
	"context"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/actions"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const operatorGroupSuffix = "-og"

type OLMOperator struct {
	StepName        string
	Namespace       string
	NamespaceLabels map[string]string
	Package     string
	Channel     string
	StartingCSV string
	// InstallPlanApproval is the OLM Subscription upgrade strategy ("Manual" or
	// "Automatic"). Empty defaults to Automatic (OLM's own default).
	InstallPlanApproval string
	Catalog             actions.OLMCatalog
	Sentinel            client.Object
	// AllNamespaces installs the operator cluster-wide via an OperatorGroup with
	// no targetNamespaces. Operators whose only install mode is AllNamespaces
	// (e.g. cert-manager) require it; the default is an own-namespace group.
	AllNamespaces bool
	// SubscriptionEnv sets environment variables on the operator pod via the
	// Subscription's spec.config (operator-level settings with no custom
	// resource, e.g. the NIM operator's ENABLE_WEBHOOKS).
	SubscriptionEnv map[string]string
}

type olmOperatorStep struct {
	*orchestrator2.BaseStep
	deps     orchestrator2.Deps
	cfg      OLMOperator
	sentinel client.Object
}

func NewOLMOperator(cfg OLMOperator, deps orchestrator2.Deps) orchestrator2.Step {
	return &olmOperatorStep{
		BaseStep: orchestrator2.NewBaseStep(cfg.StepName),
		deps:     deps,
		cfg:      cfg,
		sentinel: cfg.Sentinel,
	}
}

func (s *olmOperatorStep) Detect(ctx context.Context, _ *spec.Spec) (orchestrator2.DetectResult, error) {
	obj := s.sentinel.DeepCopyObject().(client.Object)
	err := s.deps.K8sExecutor.Get(ctx, client.ObjectKeyFromObject(s.sentinel), obj)
	if err == nil {
		return orchestrator2.DetectResult{Phase: orchestrator2.DetectInstalled}, nil
	}
	if apierrors.IsNotFound(err) {
		return orchestrator2.DetectResult{Phase: orchestrator2.DetectNotInstalled}, nil
	}

	return orchestrator2.DetectResult{}, err
}

func (s *olmOperatorStep) BuildActions(*spec.Spec) ([]orchestrator2.PerformAction, error) {
	namespace := &corev1.Namespace{
		TypeMeta:   metav1.TypeMeta{Kind: "Namespace", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{Name: s.cfg.Namespace, Labels: s.cfg.NamespaceLabels},
	}

	targetNamespaces := []string{s.cfg.Namespace}
	if s.cfg.AllNamespaces {
		targetNamespaces = nil
	}

	operatorGroup := actions.OperatorGroupObject(s.cfg.Namespace+operatorGroupSuffix, s.cfg.Namespace, targetNamespaces)
	subscription := actions.SubscriptionObject(s.cfg.Package, s.cfg.Namespace, s.cfg.Package, s.cfg.Channel, s.cfg.StartingCSV, s.cfg.InstallPlanApproval, s.cfg.Catalog, s.cfg.SubscriptionEnv)

	return []orchestrator2.PerformAction{
		actions.K8sApply{Object: namespace},
		actions.K8sApply{Object: operatorGroup},
		actions.K8sApply{Object: subscription},
		actions.WaitForCSV{Namespace: s.cfg.Namespace, SubscriptionName: s.cfg.Package},
	}, nil
}

// OperatorCR describes the operator's primary custom resource that is applied
// after its OLM install, using the default the operator ships in its CSV
// (alm-examples) merged with the user's Override.
type OperatorCR struct {
	StepName     string
	CRD          string
	CRAPIVersion string
	CRKind       string
	CRName       string
	// CRNamespace is empty for cluster-scoped resources (e.g. ClusterPolicy).
	CRNamespace      string
	CSVNamespace     string
	SubscriptionName string
	Override         map[string]any
}

type operatorCRStep struct {
	*orchestrator2.BaseStep
	deps   orchestrator2.Deps
	crd    string
	probe  *unstructured.Unstructured
	action actions.ApplyCSVExampleCR
}

// NewOperatorDefaultCR builds the step that applies an OLM operator's primary CR
// from its CSV default, gated on the CR's CRD being established. Detect keys off
// the CR's presence, so an existing CR (configured by any means) is skipped.
func NewOperatorDefaultCR(cr OperatorCR, deps orchestrator2.Deps) orchestrator2.Step {
	probe := &unstructured.Unstructured{}
	probe.SetAPIVersion(cr.CRAPIVersion)
	probe.SetKind(cr.CRKind)
	probe.SetName(cr.CRName)
	if cr.CRNamespace != "" {
		probe.SetNamespace(cr.CRNamespace)
	}

	return &operatorCRStep{
		BaseStep: orchestrator2.NewBaseStep(cr.StepName),
		deps:     deps,
		crd:      cr.CRD,
		probe:    probe,
		action: actions.ApplyCSVExampleCR{
			CSVNamespace:     cr.CSVNamespace,
			SubscriptionName: cr.SubscriptionName,
			CRAPIVersion:     cr.CRAPIVersion,
			CRKind:           cr.CRKind,
			CRName:           cr.CRName,
			CRNamespace:      cr.CRNamespace,
			SpecOverride:     cr.Override,
		},
	}
}

func (s *operatorCRStep) Detect(ctx context.Context, _ *spec.Spec) (orchestrator2.DetectResult, error) {
	probe := s.probe.DeepCopy()
	err := s.deps.K8sExecutor.Get(ctx, client.ObjectKeyFromObject(probe), probe)
	if err == nil {
		return orchestrator2.DetectResult{Phase: orchestrator2.DetectInstalled}, nil
	}
	if apierrors.IsNotFound(err) {
		return orchestrator2.DetectResult{Phase: orchestrator2.DetectNotInstalled}, nil
	}

	return orchestrator2.DetectResult{}, err
}

func (s *operatorCRStep) BuildActions(*spec.Spec) ([]orchestrator2.PerformAction, error) {
	return []orchestrator2.PerformAction{
		actions.WaitForCRDEstablished{Name: s.crd},
		s.action,
	}, nil
}
