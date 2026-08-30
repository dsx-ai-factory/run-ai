package steps

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/config"
	"github.com/run-ai/runai/runai-installer/installer/internal/utils/helmclient"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	k8s "github.com/run-ai/runai/runai-installer/installer/internal/utils/k8sclient"
	"github.com/run-ai/runai/runai-installer/installer/internal/logging"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
)

const GpuOperatorStepName = "Installing NVIDIA GPU Operator"

type GpuOperatorInstallStep struct {
	*step.RunAIInstallStepImpl
	k8sClient  k8s.K8sClients
	crdChecker CRDChecker
	version    string
}

func NewGpuOperatorInstallStep(k8sClient k8s.K8sClients, version string) step.InstallStep {
	g := &GpuOperatorInstallStep{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl(GpuOperatorStepName),
		k8sClient:            k8sClient,
		crdChecker:           &DefaultCRDChecker{},
		version:              version,
	}
	g.SetStatusEvaluator(g.EvaluateStatus)
	return g
}

func (g *GpuOperatorInstallStep) EvaluateStatus(runaiData *model.RunAIInstallData) {
	// Early return if already in terminal state (prevents race with timeout detection)
	if g.ShouldSkipStatusEvaluation() {
		return
	}

	timeout := g.GetTimeout()
	logging.Infof("[Prereq] Checking if NVIDIA GPU Operator is installed...")
	gvk := schema.GroupVersionKind{
		Group:   "nvidia.com",
		Version: "v1",
		Kind:    "ClusterPolicyList", // plural of the CRD kind
	}
	crdInstalledOnCluster := g.crdChecker.IsCRDInstalled(g.k8sClient.GetConfig(), "clusterpolicies.nvidia.com")
	if !crdInstalledOnCluster {
		elapsed := g.GetElapsedTime()
		if elapsed > 0 {
			logging.Infof("[Prereq] GPU Operator CRD not yet ready (elapsed: %v, timeout: %v)", elapsed, timeout)
		} else {
			logging.Infof("[Prereq] GPU Operator CRD not found")
		}
		return
	}
	_, err := HasCustomResources(g.k8sClient.GetClient(), gvk)
	if err != nil {
		elapsed := g.GetElapsedTime()
		if elapsed > 0 {
			logging.Infof("[Prereq] ClusterPolicy resources not yet ready (elapsed: %v, timeout: %v)", elapsed, timeout)
		} else {
			logging.Infof("[Prereq] No ClusterPolicy resources found")
		}
		return
	}
	err = ValidateComponentNamespace(context.Background(), g.k8sClient.GetClient(), "gpu", "operator", "gpu-operator")
	if err != nil {
		elapsed := g.GetElapsedTime()
		if elapsed > 0 {
			logging.Infof("[Prereq] GPU Operator deployment not yet ready (elapsed: %v, timeout: %v)", elapsed, timeout)
		} else {
			logging.Infof("[Prereq] GPU Operator deployment not found")
		}
		return
	}
	logging.Infof("[Prereq] NVIDIA GPU Operator is installed")
	g.SetStatus(model.Installed, nil)
}

func (g *GpuOperatorInstallStep) Execute(ctx context.Context, runaiData *model.RunAIInstallData) error {
	g.SetStatus(model.Installing, nil)
	if runaiData.GetGPUType() == config.GPUTypeReal {
		logging.Infof("[Prereq] Installing real NVIDIA GPU Operator v%s...", g.version)
		return g.installRealGpuOperator()
	} else {
		logging.Infof("[Prereq] Installing fake GPU Operator (simulation mode)...")
		return g.installFakeGpuOperator()
	}
}

func (g *GpuOperatorInstallStep) installRealGpuOperator() error {
	logging.Infof("[Prereq] Installing real NVIDIA GPU operator from NGC repository...")
	// Install real NVIDIA GPU operator with Helm
	err := helmclient.InstallOrUpgradeRelease(context.Background(),
		"https://helmclient.ngc.nvidia.com/nvidia",
		g.version,
		"gpu-operator",
		"gpu-operator",
		"gpu-operator",
		false,
		[]string{})
	if err != nil {
		fullErr := fmt.Errorf("failed to install real NVIDIA GPU operator with helm: %v", err)
		logging.Errorf("[Prereq] %v", fullErr)
		g.SetStatus(model.Failed, fullErr)
		return fullErr
	}
	logging.Infof("[Prereq] NVIDIA GPU Operator Helm chart deployed, waiting for readiness")
	return nil
}

func (g *GpuOperatorInstallStep) installFakeGpuOperator() error {
	// Step 1: Label non-master nodes
	logging.Infof("[Prereq] Labeling non-master nodes for fake GPU simulation...")
	if err := g.labelNonMasterNodes(); err != nil {
		logging.Errorf("[Prereq] Failed to label non-master nodes: %v", err)
		return fmt.Errorf("failed to label non-master nodes: %v", err)
	}

	// Step 2: Install fake GPU operator with Helm
	logging.Infof("[Prereq] Installing fake GPU operator Helm chart...")
	if err := g.installFakeGpuOperatorHelm(); err != nil {
		fullErr := fmt.Errorf("failed to install fake GPU operator: %v", err)
		logging.Errorf("[Prereq] %v", fullErr)
		g.SetStatus(model.Failed, fullErr)
		return fullErr
	}

	logging.Infof("[Prereq] Fake GPU Operator Helm chart deployed, waiting for readiness")
	return nil
}

func (g *GpuOperatorInstallStep) labelNonMasterNodes() error {
	// Create Kubernetes clientset
	logging.Infof("[Prereq] Creating Kubernetes clientset for node labeling...")
	clientset, err := kubernetes.NewForConfig(g.k8sClient.GetConfig())
	if err != nil {
		logging.Errorf("[Prereq] Failed to create Kubernetes clientset: %v", err)
		return fmt.Errorf("failed to create kubernetes clientset: %v", err)
	}

	// Get all nodes
	logging.Infof("[Prereq] Listing cluster nodes...")
	nodes, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		logging.Errorf("[Prereq] Failed to list nodes: %v", err)
		return fmt.Errorf("failed to list nodes: %v", err)
	}
	logging.Infof("[Prereq] Found %d nodes in cluster", len(nodes.Items))

	// Label non-master nodes
	labeledCount := 0
	for _, node := range nodes.Items {
		// Check if node is a master node
		isMaster := false
		if node.Labels != nil {
			if _, exists := node.Labels["node-role.kubernetes.io/master"]; exists {
				isMaster = true
			}
		}

		// Skip master nodes
		if isMaster {
			logging.Infof("[Prereq] Skipping master node: %s", node.Name)
			continue
		}

		// Label the non-master node
		logging.Infof("[Prereq] Labeling node '%s' with 'run.ai/simulated-gpu-node-pool=default'...", node.Name)
		cmd := exec.Command("kubectl", "label", "node", node.Name, "run.ai/simulated-gpu-node-pool=default", "--overwrite")
		cmd.Stdout = logging.InfoWriter()
		cmd.Stderr = logging.ErrorWriter()
		if err := cmd.Run(); err != nil {
			logging.Errorf("[Prereq] Failed to label node %s: %v", node.Name, err)
			return fmt.Errorf("failed to label node %s: %v", node.Name, err)
		}
		labeledCount++
	}
	logging.Infof("[Prereq] Successfully labeled %d worker nodes", labeledCount)

	return nil
}

func (g *GpuOperatorInstallStep) installFakeGpuOperatorHelm() error {
	logging.Infof("[Prereq] Installing fake GPU operator from ghcr.io...")
	err := helmclient.InstallOrUpgradeRelease(context.Background(),
		"oci://ghcr.io/run-ai/fake-gpu-operator/fake-gpu-operator",
		"0.0.62", "gpu-operator", "gpu-operator", "gpu-operator", true, []string{"topology.nodePools.default.gpuCount=8"})
	if err != nil {
		logging.Errorf("[Prereq] Failed to install fake GPU operator: %v", err)
		return fmt.Errorf("failed to install fake GPU operator with helm: %v", err)
	}
	return nil
}
