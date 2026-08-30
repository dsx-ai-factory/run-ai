package common

import (
	"fmt"
	"strings"
	"time"

	"github.com/run-ai/runai/runai-installer/installer/internal/flow"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"

	spinner "github.com/charmbracelet/bubbles/spinner"
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	styles "github.com/run-ai/runai/runai-installer/installer/internal/tui/styles"
)

type uiFlowManager struct {
	flowManager *flow.FlowManager
	flows       []*flow.InstallFlow
}

func NewUIFlowManager() *uiFlowManager {
	return &uiFlowManager{
		flowManager: flow.NewFlowManager(nil),
	}
}

func (s *uiFlowManager) SetFlows(flows []*flow.InstallFlow) {
	s.flowManager.SetFlows(flows)
	s.flows = flows
}

func (s *uiFlowManager) getMessages() []string {
	var messages []string
	for _, flow := range s.flows {
		if flow.GetStatus() == model.NotInstalled {
			continue
		}
		flowErrors := flow.GetStepErrors()
		for _, err := range flowErrors {
			if err != nil {
				title := styles.AppStyles.Normal.Render(fmt.Sprintf("Error in %s: ", flow.GetName()))
				content := styles.AppStyles.ErrorRed.Render(fmt.Sprintf("%v", err))
				messages = append(messages, title+content)
			}
		}
	}

	if !s.flowManager.IsDone() {
		currentFlow := s.flowManager.GetCurrentFlow()
		if currentFlow.GetStatus() == model.Failed {
			if lastErr := currentFlow.GetCurrentStepError(); lastErr != nil {
				title := styles.AppStyles.Normal.Render(fmt.Sprintf("Error in %s: ", currentFlow.GetName()))
				content := styles.AppStyles.ErrorRed.Render(fmt.Sprintf("%v", lastErr))
				messages = append(messages, title+content)
			}
		}
	}
	return messages
}

func (s *uiFlowManager) Render(spinner spinner.Model, runaiData *model.RunAIInstallData, timeout time.Duration) string {
	var content string
	allFlowsAreInstalled := true
	for i, flow := range s.flows {
		flowName := flow.GetName()
		if !flow.IsTerminalStatus() {
			allFlowsAreInstalled = false
			if i != s.flowManager.GetCurrentFlowIndex() {
				stepLine := fmt.Sprintf("%s %s \n", "[ ]", flowName)
				content += styles.AppStyles.Normal.Render(stepLine) + "\n"
			} else {
				stepStatusStr := s.getInstallationStatusDesc(flow, runaiData, spinner)
				content += stepStatusStr
			}
		} else {
			allFlowsAreInstalled = allFlowsAreInstalled && flow.IsDone()
			flowStatusStr := s.getInstallationStatusDesc(flow, runaiData, spinner)
			content += flowStatusStr
		}
	}
	content += "\n"
	statusLabel := styles.AppStyles.Normal.Render("Status: ")
	if s.flowManager.IsDone() {
		if allFlowsAreInstalled {
			content += statusLabel + styles.AppStyles.SuccessGreen.Render("All components ready") + "\n\n"
		} else {
			content += statusLabel + styles.AppStyles.ErrorRed.Render("Some components failed") + "\n\n"
		}
	} else {
		timeoutInfo := fmt.Sprintf(" (timeout: %v)", timeout)
		content += statusLabel + styles.AppStyles.Normal.Render("Installing..."+timeoutInfo) + "\n\n"
	}
	return content
}

func (s *uiFlowManager) getInstallationStatusDesc(flow *flow.InstallFlow, runaiData *model.RunAIInstallData, spinner spinner.Model) string {
	var status model.InstallPhase
	var name string
	var currentStep step.InstallStep
	if flow.IsTerminalStatus() {
		status = flow.GetStatus()
		name = flow.GetName()
	} else {
		currentStep = flow.GetCurrentStep()
		// Use cached status during rendering to avoid I/O on every frame
		status = currentStep.GetCachedStatus()
		name = fmt.Sprintf("%s - %s", flow.GetName(), currentStep.GetName())
	}

	style := styles.AppStyles.Normal
	stepStatusStr := " "
	timeInfo := ""
	switch status {
	case model.NotInstalled:
		stepStatusStr = "[ ]"
	case model.Installing:
		stepStatusStr = "[" + strings.TrimSpace(spinner.View()) + "]"
		// Show elapsed time and timeout for installing steps
		if currentStep != nil {
			elapsed := currentStep.GetElapsedTime()
			timeout := currentStep.GetTimeout()
			if elapsed > 0 {
				timeInfo = fmt.Sprintf(" (%v / %v)", elapsed, timeout)
			}
		}
	case model.Installed:
		style = styles.AppStyles.SuccessGreen
		stepStatusStr = "[✓]"
	case model.Failed, model.Timeout:
		style = styles.AppStyles.ErrorRed
		stepStatusStr = "[✗]"
	}
	flowLine := fmt.Sprintf("%s %s%s\n", stepStatusStr, name, timeInfo)
	return style.Render(flowLine) + "\n"
}
