package steps

import (
	"context"

	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	k8s "github.com/run-ai/runai/runai-installer/installer/internal/utils/k8sclient"
	"github.com/run-ai/runai/runai-installer/installer/internal/logging"
)

const KnativeOperatorStepName = "Installing Knative Operator"

type KnativeOperatorInstallStep struct {
	*step.RunAIInstallStepImpl
	k8sClient     k8s.K8sClients
	crdChecker    CRDChecker
	version       string
	helmInstaller step.HelmInstaller
}

func NewKnativeOperatorInstallStep(k8sClient k8s.K8sClients, version string, helmInstaller step.HelmInstaller) step.InstallStep {
	k := &KnativeOperatorInstallStep{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl(KnativeOperatorStepName),
		crdChecker:           &DefaultCRDChecker{},
		k8sClient:            k8sClient,
		version:              version,
		helmInstaller:        helmInstaller,
	}
	k.SetStatusEvaluator(k.EvaluateStatus)
	return k
}

func (k *KnativeOperatorInstallStep) EvaluateStatus(runaiData *model.RunAIInstallData) {
	// Early return if already in terminal state (prevents race with timeout detection)
	if k.ShouldSkipStatusEvaluation() {
		return
	}

	logging.Infof("[Prereq] Checking if Knative operator is installed...")
	err := ValidateComponentNamespace(context.Background(), k.k8sClient.GetClient(), "knative", "operator", "knative-operator")
	if err != nil {
		elapsed := k.GetElapsedTime()
		timeout := k.GetTimeout()
		if elapsed > 0 {
			logging.Infof("[Prereq] Knative operator not yet ready (elapsed: %v, timeout: %v)", elapsed, timeout)
		} else {
			logging.Infof("[Prereq] Knative operator not found in cluster")
		}
		return
	}
	logging.Infof("[Prereq] Knative Operator is installed")
	k.SetStatus(model.Installed, nil)
}

func (k *KnativeOperatorInstallStep) Execute(ctx context.Context, runaiData *model.RunAIInstallData) error {
	logging.Infof("[Prereq] Installing Knative Operator %s...", k.version)
	k.SetStatus(model.Installing, nil)

	repo := k.getChartRepo()
	logging.Infof("[Prereq] Installing Knative Operator from repo: %s, version: %s", repo, k.version)
	err := k.helmInstaller.InstallOrUpgradeRelease(ctx,
		repo,
		k.version, "knative-operator", "knative-operator", "knative-operator", false, []string{})
	if err != nil {
		logging.Errorf("[Prereq - %s] %v", KnativeOperatorStepName, err)
		k.SetStatus(model.Failed, err)
		return err
	}
	logging.Infof("[Prereq - %s] Helm chart deployed, waiting for readiness", KnativeOperatorStepName)
	return nil
}

func (i *KnativeOperatorInstallStep) getChartRepo() string {
	return "https://knative.github.io/operator"
}
