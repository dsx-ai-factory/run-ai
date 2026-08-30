package install

import (
	"github.com/run-ai/runai/runai-installer/installer/internal/inputs"
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	prereqcomponents "github.com/run-ai/runai/runai-installer/installer/internal/prereq/components"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
)

// GetRelevantInputTypes returns the set of input types relevant to the current configuration
// based on selected prerequisites. This is shared between ConfigurationScreen and InputValidationScreen.
func GetRelevantInputTypes(runaiData *model.RunAIInstallData, prerequisitesMap map[string]prereqcomponents.Prerequisite) map[inputs.InstallInputType]bool {
	relevantInputs := make(map[inputs.InstallInputType]bool)

	// Collect input getters from components
	inputGetters := []step.InstallationComponent{}

	runaiInstallFormat := GetRunaiInstallFormatFromIdentifier(runaiData.Config.RunaiInstallFormat)
	for _, component := range runaiInstallFormat.GetComponents() {
		if runaiData.Config.SkipControlPlane && component.ComponentType() == step.ComponentTypeControlPlane {
			continue
		}
		if runaiData.Config.SkipCluster && component.ComponentType() == step.ComponentTypeCluster {
			continue
		}
		inputGetters = append(inputGetters, component)
	}

	// Add selected prerequisites
	for _, selectedPrerequisite := range runaiData.Prerequisites {
		if prereq, ok := prerequisitesMap[selectedPrerequisite]; ok {
			inputGetters = append(inputGetters, prereq)
		}
	}

	// Collect all relevant inputs from components
	for _, inputGetter := range inputGetters {
		for _, input := range inputGetter.GetMandatoryInputs(runaiData) {
			relevantInputs[input] = true
		}
		for _, input := range inputGetter.GetOptionalInputs(runaiData) {
			relevantInputs[input] = false
		}
	}
	return relevantInputs
}
