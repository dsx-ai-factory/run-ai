package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	k8s "github.com/run-ai/runai/runai-installer/installer/internal/utils/k8sclient"
	"github.com/run-ai/runai/runai-installer/installer/internal/logging"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const TrainingOperatorStepName = "Installing Kubeflow Training Operator"

// TrainingOperatorExecutor defines the operations needed to install Training Operator
type TrainingOperatorExecutor interface {
	ApplyTrainingOperatorKubectl(version string) error
	PatchDeployment(k8sClient k8s.K8sClients) error
	DeleteCRD(k8sClient k8s.K8sClients) error
	ApplyMPIOperator() error
}

// DefaultTrainingOperatorExecutor implements the actual installation logic
type DefaultTrainingOperatorExecutor struct{}

func (e *DefaultTrainingOperatorExecutor) ApplyTrainingOperatorKubectl(version string) error {
	return applyTrainingOperatorKubectl(version)
}

func (e *DefaultTrainingOperatorExecutor) PatchDeployment(k8sClient k8s.K8sClients) error {
	return patchDeployment(k8sClient)
}

func (e *DefaultTrainingOperatorExecutor) DeleteCRD(k8sClient k8s.K8sClients) error {
	return deleteCRD(k8sClient)
}

func (e *DefaultTrainingOperatorExecutor) ApplyMPIOperator() error {
	return applyMPIOperator()
}

type TrainingOperatorInstallStep struct {
	*step.RunAIInstallStepImpl
	k8sClient  k8s.K8sClients
	crdChecker CRDChecker
	executor   TrainingOperatorExecutor
	version    string
}

func NewTrainingOperatorInstallStep(k8sClient k8s.K8sClients, version string) step.InstallStep {
	t := &TrainingOperatorInstallStep{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl(TrainingOperatorStepName),
		crdChecker:           &DefaultCRDChecker{},
		executor:             &DefaultTrainingOperatorExecutor{},
		k8sClient:            k8sClient,
		version:              version,
	}
	t.SetStatusEvaluator(t.EvaluateStatus)
	return t
}

func (t *TrainingOperatorInstallStep) EvaluateStatus(runaiData *model.RunAIInstallData) {
	// Early return if already in terminal state (prevents race with timeout detection)
	if t.ShouldSkipStatusEvaluation() {
		return
	}

	logging.Infof("[Prereq] Checking if Kubeflow Training Operator is installed...")
	t.getInstallationPhaseWithClientAndCRDChecker(
		t.k8sClient.GetClient(),
		t.k8sClient.GetConfig(),
		t.crdChecker.IsCRDInstalled,
		runaiData,
	)
}

// getInstallationPhaseWithClientAndCRDChecker is a test helper to allow injection of fake client and CRD checker.
// It is unexported and can be used in tests.
func (t *TrainingOperatorInstallStep) getInstallationPhaseWithClientAndCRDChecker(
	k8sClient client.Client,
	cfg *rest.Config,
	crdChecker func(cfg *rest.Config, crdName string) bool,
	runaiData *model.RunAIInstallData,
) model.InstallPhase {
	timeout := t.GetTimeout()

	// Validate namespace and deployment
	err := ValidateComponentNamespace(context.Background(), k8sClient, "training", "operator", "kubeflow")
	if err != nil {
		elapsed := t.GetElapsedTime()
		if elapsed > 0 {
			logging.Infof("[Prereq] Training Operator deployment not yet ready (elapsed: %v, timeout: %v)", elapsed, timeout)
		} else {
			logging.Infof("[Prereq] Training Operator deployment not found")
		}
		return t.GetCachedStatus()
	}
	trainingCRDs := []string{
		"tfjobs.kubeflow.org",
		"pytorchjobs.kubeflow.org",
		"mpijobs.kubeflow.org",
		"xgboostjobs.kubeflow.org",
		"mxjobs.kubeflow.org",
		"paddlejobs.kubeflow.org",
	}
	logging.Infof("[Prereq] Checking for Training Operator CRDs...")
	for _, name := range trainingCRDs {
		existsOnCluster := crdChecker(cfg, name)
		if existsOnCluster {
			logging.Infof("[Prereq] Found CRD '%s' - Training Operator is installed", name)
			return t.SetStatus(model.Installed, nil)
		}
	}
	elapsed := t.GetElapsedTime()
	if elapsed > 0 {
		logging.Infof("[Prereq] Training Operator CRDs not yet ready (elapsed: %v, timeout: %v)", elapsed, timeout)
	} else {
		logging.Infof("[Prereq] Training Operator CRDs not found")
	}
	return t.GetCachedStatus()
}

func (t *TrainingOperatorInstallStep) Execute(ctx context.Context, runaiData *model.RunAIInstallData) error {
	logging.Infof("[Prereq] Installing Kubeflow Training Operator %s...", t.version)
	t.SetStatus(model.Installing, nil)

	logging.Infof("[Prereq] Step 1: Applying Training Operator manifests...")
	err := t.executor.ApplyTrainingOperatorKubectl(t.version)
	if err != nil {
		logging.Errorf("[Prereq] Failed to apply Training Operator manifests: %v", err)
		t.SetStatus(model.Failed, err)
		return err
	}

	logging.Infof("[Prereq] Step 2: Patching Training Operator deployment...")
	err = t.executor.PatchDeployment(t.k8sClient)
	if err != nil {
		logging.Errorf("[Prereq] Failed to patch deployment: %v", err)
		t.SetStatus(model.Failed, err)
		return err
	}

	logging.Infof("[Prereq] Step 3: Deleting old MPI CRD...")
	err = t.executor.DeleteCRD(t.k8sClient)
	if err != nil {
		logging.Errorf("[Prereq] Failed to delete CRD: %v", err)
		t.SetStatus(model.Failed, err)
		return err
	}

	logging.Infof("[Prereq] Step 4: Applying MPI Operator...")
	err = t.executor.ApplyMPIOperator()
	if err != nil {
		logging.Errorf("[Prereq] Failed to apply MPI Operator: %v", err)
		t.SetStatus(model.Failed, err)
		return err
	}

	logging.Infof("[Prereq] Kubeflow Training Operator manifests applied, waiting for readiness")
	return nil
}

func applyTrainingOperatorKubectl(version string) error {
	trainingOperator := fmt.Sprintf("github.com/kubeflow/training-operator.git/manifests/overlays/standalone?ref=%s", version)
	logging.Infof("[Prereq] Running: kubectl apply -k %s", trainingOperator)
	cmd := exec.Command("kubectl", "apply", "-k", trainingOperator)
	cmd.Stdout = logging.InfoWriter()
	cmd.Stderr = logging.ErrorWriter()
	err := cmd.Run()
	if err != nil {
		logging.Errorf("[Prereq] kubectl apply failed: %v", err)
	} else {
		logging.Infof("[Prereq] kubectl apply completed successfully")
	}
	return err
}

func patchDeployment(k8sClient k8s.K8sClients) error {
	logging.Infof("[Prereq] Patching deployment 'training-operator' in namespace 'kubeflow'...")
	clientset, err := kubernetes.NewForConfig(k8sClient.GetConfig())
	if err != nil {
		logging.Errorf("[Prereq] Failed to create Kubernetes clientset: %v", err)
		return err
	}

	patch := []map[string]interface{}{
		{
			"op":    "add",
			"path":  "/spec/template/spec/containers/0/args",
			"value": []string{"--enable-scheme=tfjob", "--enable-scheme=pytorchjob", "--enable-scheme=xgboostjob"},
		},
	}
	patchBytes, _ := json.Marshal(patch)

	_, err = clientset.AppsV1().Deployments("kubeflow").Patch(
		context.TODO(),
		"training-operator",
		types.JSONPatchType,
		patchBytes,
		metav1.PatchOptions{},
	)
	if err != nil {
		logging.Errorf("[Prereq] Failed to patch deployment: %v", err)
	} else {
		logging.Infof("[Prereq] Deployment patched successfully")
	}
	return err
}

func deleteCRD(k8sClient k8s.K8sClients) error {
	logging.Infof("[Prereq] Deleting CRD 'mpijobs.kubeflow.org'...")
	apiExtClient, err := apiextensionsclient.NewForConfig(k8sClient.GetConfig())
	if err != nil {
		logging.Errorf("[Prereq] Failed to create API extensions client: %v", err)
		return err
	}

	err = apiExtClient.ApiextensionsV1().CustomResourceDefinitions().Delete(context.Background(), "mpijobs.kubeflow.org", metav1.DeleteOptions{})
	if err != nil {
		logging.Infof("[Prereq] CRD delete result: %v (may be expected if CRD doesn't exist)", err)
	} else {
		logging.Infof("[Prereq] CRD deleted successfully")
	}
	return err
}

func applyMPIOperator() error {
	mpiOperatorURL := "https://raw.githubusercontent.com/kubeflow/mpi-operator/v0.6.0/deploy/v2beta1/mpi-operator.yaml"
	logging.Infof("[Prereq] Running: kubectl apply --server-side -f %s", mpiOperatorURL)
	cmd := exec.Command("kubectl", "apply", "--server-side", "-f", mpiOperatorURL)
	cmd.Stdout = logging.InfoWriter()
	cmd.Stderr = logging.ErrorWriter()
	err := cmd.Run()
	if err != nil {
		logging.Errorf("[Prereq] kubectl apply failed: %v", err)
	} else {
		logging.Infof("[Prereq] kubectl apply completed successfully")
	}
	return err
}
