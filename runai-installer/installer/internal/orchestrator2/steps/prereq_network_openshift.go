package steps

import (
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/actions"
)

const (
	sriovNamespace = "openshift-sriov-network-operator"
	sriovPackage   = "sriov-network-operator"
	sriovSentinel  = "sriov-network-operator"

	networkOperatorOCPNamespace = "nvidia-network-operator"
	networkOperatorOCPPackage   = "nvidia-network-operator"
	networkOperatorOCPSentinel  = "nvidia-network-operator-controller-manager"
	nicClusterPolicyCRD         = "nicclusterpolicies.mellanox.com"
	nicClusterPolicyName        = "nic-cluster-policy"
	nicClusterPolicyAPIVersion  = "mellanox.com/v1alpha1"
	nicClusterPolicyKind        = "NicClusterPolicy"
)

// NetworkOpenShiftParams configures the OpenShift network operator flow. Channels
// and the NicClusterPolicy override are resolved by the planner.
type NetworkOpenShiftParams struct {
	NetworkChannel string
	// NetworkVersion is the resolved network operator version; it pins the
	// Subscription's startingCSV (nvidia-network-operator.<version>) so the
	// OperatorHub install lands on the same version the Helm path would install.
	NetworkVersion string
	NFDChannel     string
	SRIOVChannel   string
	// InstallPlanApproval is the OLM upgrade strategy applied to every operator
	// in the flow (NFD + SR-IOV + network).
	InstallPlanApproval  string
	NicClusterPolicySpec map[string]any
	SubscriptionEnv      map[string]string
	CertifiedCatalog     actions.OLMCatalog
	RedHatCatalog        actions.OLMCatalog
}

// NewNetworkOperatorOpenShift installs the NVIDIA network operator via
// OperatorHub: Node Feature Discovery and the SR-IOV Network Operator (its
// OpenShift dependencies) first, then the certified network operator and its
// NicClusterPolicy CR. The hardware-specific SriovNetworkNodePolicy is left to
// the user, as it depends on the cluster's NICs.
func NewNetworkOperatorOpenShift(p NetworkOpenShiftParams, deps orchestrator2.Deps) []orchestrator2.Step {
	out := newNFDSteps(p.NFDChannel, p.InstallPlanApproval, p.RedHatCatalog, deps)

	sriov := NewOLMOperator(OLMOperator{
		StepName:            "install-sriov-network-operator",
		Namespace:           sriovNamespace,
		Package:             sriovPackage,
		Channel:             p.SRIOVChannel,
		InstallPlanApproval: p.InstallPlanApproval,
		Catalog:             p.RedHatCatalog,
		Sentinel:            deploymentSentinel(sriovSentinel, sriovNamespace),
	}, deps)

	networkOperator := NewOLMOperator(OLMOperator{
		StepName:            "install-network-operator",
		Namespace:           networkOperatorOCPNamespace,
		Package:             networkOperatorOCPPackage,
		Channel:             p.NetworkChannel,
		StartingCSV:         startingCSV(networkOperatorOCPPackage, p.NetworkVersion),
		InstallPlanApproval: p.InstallPlanApproval,
		Catalog:             p.CertifiedCatalog,
		Sentinel:            deploymentSentinel(networkOperatorOCPSentinel, networkOperatorOCPNamespace),
		SubscriptionEnv:     p.SubscriptionEnv,
	}, deps)

	nicPolicy := NewOperatorDefaultCR(OperatorCR{
		StepName:         "create-nic-cluster-policy",
		CRD:              nicClusterPolicyCRD,
		CRAPIVersion:     nicClusterPolicyAPIVersion,
		CRKind:           nicClusterPolicyKind,
		CRName:           nicClusterPolicyName,
		CSVNamespace:     networkOperatorOCPNamespace,
		SubscriptionName: networkOperatorOCPPackage,
		Override:         p.NicClusterPolicySpec,
	}, deps)

	return append(out, sriov, networkOperator, nicPolicy)
}
