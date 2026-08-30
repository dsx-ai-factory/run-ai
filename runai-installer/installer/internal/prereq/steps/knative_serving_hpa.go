package steps

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	k8s "github.com/run-ai/runai/runai-installer/installer/internal/utils/k8sclient"
	"github.com/run-ai/runai/runai-installer/installer/internal/logging"
)

const (
	KnativeServingHpaStepName = "Installing Knative Serving HPA"
)

type KnativeServingHPAInstallStep struct {
	k8sClient k8s.K8sClients
	version   string
	*step.RunAIInstallStepImpl
}

func NewKnativeServingHPAInstallStep(k8sClient k8s.K8sClients, version string) step.InstallStep {
	return &KnativeServingHPAInstallStep{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl(KnativeServingHpaStepName),
		k8sClient:            k8sClient,
		version:              version,
	}
}

func (k *KnativeServingHPAInstallStep) GetStatus(runaiData *model.RunAIInstallData) model.InstallPhase {
	return k.GetCachedStatus()
}

func (k *KnativeServingHPAInstallStep) Execute(ctx context.Context, runaiData *model.RunAIInstallData) error {
	logging.Infof("[Prereq] Installing Knative Serving HPA %s...", k.version)
	k.SetStatus(model.Installing, nil)
	if err := k.applyKnativeServingHPAKubectl(); err != nil {
		fullError := fmt.Errorf("failed to apply Knative Serving HPA: %v", err)
		logging.Errorf("[Prereq - %s ]  %v", KnativeServingHpaStepName, fullError)
		k.SetStatus(model.Failed, fullError)
		return fullError
	}
	logging.Infof("[Prereq - %s] Knative Serving HPA manifests applied successfully", KnativeServingHpaStepName)
	k.SetStatus(model.Installed, nil)
	return nil
}

func (k *KnativeServingHPAInstallStep) applyKnativeServingHPAKubectl() error {
	hpaUrl := fmt.Sprintf("https://github.com/knative/serving/releases/download/knative-v%s/serving-hpa.yaml", k.version)
	cmd := exec.Command("kubectl", "apply", "-f", hpaUrl)
	logging.Infof("[Prereq - %s ] Executing command: %s", KnativeServingHpaStepName, cmd.String())

	// Use logging writers to stream output through logging system (like other kubectl commands)
	// This prevents breaking the TUI while allowing real-time line-by-line logging
	cmd.Stdout = logging.InfoWriter()
	cmd.Stderr = logging.ErrorWriter()

	err := cmd.Run()
	if err != nil {
		logging.Errorf("[Prereq - %s] kubectl apply failed: %v", KnativeServingHpaStepName, err)
	} else {
		logging.Infof("[Prereq - %s] kubectl apply completed successfully", KnativeServingHpaStepName)
	}
	return err
}
