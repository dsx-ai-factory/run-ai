package screens

import (
	"testing"

	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/config"
	prereqsteps "github.com/run-ai/runai/runai-installer/installer/internal/prereq/components"
	preReqMocks "github.com/run-ai/runai/runai-installer/installer/internal/prereq/components/mocks"
	"github.com/stretchr/testify/assert"
)

func TestInitPrerequisitesScreen(t *testing.T) {
	runaiData := &model.RunAIInstallData{
		Config: &config.InstallConfig{},
	}
	runaiData.Prerequisites = []string{"prometheus"}

	prereq1 := preReqMocks.NewMockPrerequisite(t)
	prereq1.On("GetName").Return("GPU Operator").Maybe()
	prereq1.On("IsMandatory").Return(true).Maybe()
	prereq1.On("InstalledOnCluster", runaiData).Return(model.NotInstalled).Maybe()
	prereq1.On("IsDefaultSelected").Return(false).Maybe()

	prereq2 := preReqMocks.NewMockPrerequisite(t)
	prereq2.On("GetName").Return("Network Operator").Maybe()
	prereq2.On("IsMandatory").Return(false).Maybe()
	prereq2.On("InstalledOnCluster", runaiData).Return(model.Installed).Maybe()
	prereq2.On("IsDefaultSelected").Return(false).Maybe()

	prereq3 := preReqMocks.NewMockPrerequisite(t)
	prereq3.On("GetName").Return("prometheus").Maybe()
	prereq3.On("IsMandatory").Return(false).Maybe()
	prereq3.On("InstalledOnCluster", runaiData).Return(model.NotInstalled).Maybe()
	prereq3.On("IsDefaultSelected").Return(false).Maybe()

	prereq4 := preReqMocks.NewMockPrerequisite(t)
	prereq4.On("GetName").Return("knative Serving").Maybe()
	prereq4.On("IsMandatory").Return(false).Maybe()
	prereq4.On("InstalledOnCluster", runaiData).Return(model.NotInstalled).Maybe()
	prereq4.On("IsDefaultSelected").Return(true).Maybe()

	prereqMap := map[string]prereqsteps.Prerequisite{
		"GPU Operator":     prereq1,
		"Network Operator": prereq2,
		"prometheus":       prereq3,
		"knative Serving":  prereq4,
	}

	prereqs := initPrerequisitesScreen(runaiData, prereqMap)
	assert.Equal(t, 4, len(prereqs))

	var foundPrometheus, foundGPU, foundNetwork, foundKnativeServing bool
	for _, p := range prereqs {
		switch p.prerequisite.GetName() {
		case "prometheus":
			assert.True(t, p.selected, "prometheus should be selected from runaiData")
			foundPrometheus = true
		case "GPU Operator":
			assert.True(t, p.selected, "GPU Operator should be selected because it's mandatory")
			foundGPU = true
		case "Network Operator":
			assert.False(t, p.selected, "Network Operator should not be selected (not mandatory, just installed)")
			foundNetwork = true
		case "knative Serving":
			assert.True(t, p.selected, "knative Serving should be selected because it's default selected")
			foundKnativeServing = true
		}
	}
	assert.True(t, foundPrometheus && foundGPU && foundNetwork && foundKnativeServing)
}

func TestUpdateSelection(t *testing.T) {
	runaiData := &model.RunAIInstallData{
		Config: &config.InstallConfig{},
	}

	prereq1 := preReqMocks.NewMockPrerequisite(t)
	prereq1.On("GetName").Return("NVIDIA GPU Operator").Maybe()
	prereq1.On("IsMandatory").Return(true).Maybe()
	prereq1.On("InstalledOnCluster", runaiData).Return(model.NotInstalled).Maybe()
	prereq1.On("IsDefaultSelected").Return(false).Maybe()

	prereq2 := preReqMocks.NewMockPrerequisite(t)
	prereq2.On("GetName").Return("NVIDIA Network Operator").Maybe()
	prereq2.On("IsMandatory").Return(false).Maybe()
	prereq2.On("InstalledOnCluster", runaiData).Return(model.Installed).Maybe()
	prereq2.On("IsDefaultSelected").Return(false).Maybe()

	prereqMap := map[string]prereqsteps.Prerequisite{
		"NVIDIA GPU Operator":     prereq1,
		"NVIDIA Network Operator": prereq2,
	}

	screen := NewPrerequisitesScreen(prereqMap, nil)
	screen.Init(runaiData)

	// Simulate user selection
	for _, opt := range screen.OptionsModel.GetOptions() {
		if opt.(*PrerequisiteInstallnData).prerequisite.GetName() == "NVIDIA GPU Operator" {
			opt.SetSelected(true)
		}
	}

	screen.UpdateSelection(runaiData)
	assert.Contains(t, runaiData.Prerequisites, "NVIDIA GPU Operator")
}

func TestPrerequisiteInstallnData_GetUiLabel(t *testing.T) {
	t.Run("installed prerequisite shows installed label", func(t *testing.T) {
		prereq := preReqMocks.NewMockPrerequisite(t)
		prereq.On("GetName").Return("GPU Operator")
		prereq.On("GetVersion").Return("v23.9.0")
		prereq.On("IsDefaultSelected").Return(false).Maybe()

		data := &PrerequisiteInstallnData{
			prerequisite: prereq,
			installPhase: model.Installed,
		}

		expected := "GPU Operator - v23.9.0 - already installed"
		assert.Equal(t, expected, data.GetUiLabel())
	})

	t.Run("not installed prerequisite shows regular label", func(t *testing.T) {
		prereq := preReqMocks.NewMockPrerequisite(t)
		prereq.On("GetName").Return("prometheus")
		prereq.On("GetVersion").Return("v1.0.0")
		prereq.On("IsDefaultSelected").Return(false).Maybe()

		data := &PrerequisiteInstallnData{
			prerequisite: prereq,
			installPhase: model.NotInstalled,
		}

		expected := "prometheus - v1.0.0"
		assert.Equal(t, expected, data.GetUiLabel())
	})
}

func TestPrerequisiteInstallnData_Methods(t *testing.T) {
	prereq := preReqMocks.NewMockPrerequisite(t)
	prereq.On("IsMandatory").Return(true)
	prereq.On("IsDefaultSelected").Return(false).Maybe()

	data := &PrerequisiteInstallnData{
		prerequisite: prereq,
		selected:     false,
		id:           5,
	}

	t.Run("GetOrder returns id", func(t *testing.T) {
		assert.Equal(t, 5, data.GetOrder())
	})

	t.Run("IsMandatory delegates to prerequisite", func(t *testing.T) {
		assert.True(t, data.IsMandatory())
	})

	t.Run("IsSelected returns selection state", func(t *testing.T) {
		assert.False(t, data.IsSelected())
	})

	t.Run("SetSelected updates selection state", func(t *testing.T) {
		data.SetSelected(true)
		assert.True(t, data.IsSelected())
	})
}

func TestInitPrerequisitesScreenData(t *testing.T) {
	runaiData := &model.RunAIInstallData{
		Config: &config.InstallConfig{},
	}

	t.Run("mandatory prerequisite is selected by default", func(t *testing.T) {
		prereq := preReqMocks.NewMockPrerequisite(t)
		prereq.On("IsMandatory").Return(true)
		prereq.On("IsDefaultSelected").Return(false).Maybe()
		prereq.On("InstalledOnCluster", runaiData).Return(model.NotInstalled)

		data := initPrerequisitesScreenData(runaiData, prereq, 1)
		assert.True(t, data.selected)
		assert.Equal(t, 1, data.id)
		assert.Equal(t, model.NotInstalled, data.installPhase)
	})

	t.Run("installed prerequisite is not selected by default", func(t *testing.T) {
		prereq := preReqMocks.NewMockPrerequisite(t)
		prereq.On("IsMandatory").Return(false)
		prereq.On("IsDefaultSelected").Return(false).Maybe()
		prereq.On("InstalledOnCluster", runaiData).Return(model.Installed)

		data := initPrerequisitesScreenData(runaiData, prereq, 2)
		assert.False(t, data.selected)
		assert.Equal(t, 2, data.id)
		assert.Equal(t, model.Installed, data.installPhase)
	})

	t.Run("optional not installed but selected by default prerequisite is selected by default", func(t *testing.T) {
		prereq := preReqMocks.NewMockPrerequisite(t)
		prereq.On("IsMandatory").Return(false).Maybe()
		prereq.On("IsDefaultSelected").Return(true).Maybe()
		prereq.On("InstalledOnCluster", runaiData).Return(model.NotInstalled).Maybe()
		prereq.On("GetName").Return("non mandatory prerequisite selected by default").Maybe()

		data := initPrerequisitesScreenData(runaiData, prereq, 3)
		assert.True(t, data.selected)
		assert.Equal(t, 3, data.id)
		assert.Equal(t, model.NotInstalled, data.installPhase)
	})

}

func TestNewPrerequisitesScreen(t *testing.T) {
	prereqMap := map[string]prereqsteps.Prerequisite{}
	screen := NewPrerequisitesScreen(prereqMap, nil)

	assert.NotNil(t, screen)
	assert.Equal(t, prereqMap, screen.prerequisitesMap)
}

func TestPrerequisitesScreen_GetHeader(t *testing.T) {
	screen := &PrerequisitesScreen{}
	assert.Equal(t, "Select software dependencies to install", screen.GetHeader())
}

func TestPrerequisitesScreen_NextScreen(t *testing.T) {
	mockNextScreen := &PrerequisitesScreen{} // Use any screen as mock next screen
	screen := NewPrerequisitesScreen(map[string]prereqsteps.Prerequisite{}, mockNextScreen)
	nextScreen := screen.NextScreen()

	// Verify it returns the configuration screen
	assert.NotNil(t, nextScreen)
}

func TestInitPrerequisitesScreen_EmptyMap(t *testing.T) {
	runaiData := &model.RunAIInstallData{
		Config: &config.InstallConfig{},
	}
	prereqMap := map[string]prereqsteps.Prerequisite{}

	prereqs := initPrerequisitesScreen(runaiData, prereqMap)
	assert.Equal(t, 0, len(prereqs))
}

func TestInitPrerequisitesScreen_WithPreviousSelections(t *testing.T) {
	runaiData := &model.RunAIInstallData{
		Config: &config.InstallConfig{},
	}
	runaiData.Prerequisites = []string{"GPU Operator", "prometheus"}

	prereq1 := preReqMocks.NewMockPrerequisite(t)
	prereq1.On("GetName").Return("GPU Operator").Maybe()
	prereq1.On("IsMandatory").Return(false).Maybe()
	prereq1.On("InstalledOnCluster", runaiData).Return(model.NotInstalled).Maybe()
	prereq1.On("IsDefaultSelected").Return(false).Maybe()

	prereq2 := preReqMocks.NewMockPrerequisite(t)
	prereq2.On("GetName").Return("prometheus").Maybe()
	prereq2.On("IsMandatory").Return(false).Maybe()
	prereq2.On("InstalledOnCluster", runaiData).Return(model.NotInstalled).Maybe()
	prereq2.On("IsDefaultSelected").Return(false).Maybe()

	prereqMap := map[string]prereqsteps.Prerequisite{
		"GPU Operator": prereq1,
		"prometheus":   prereq2,
	}

	prereqs := initPrerequisitesScreen(runaiData, prereqMap)
	assert.Equal(t, 2, len(prereqs))

	// Both should be selected due to previous selection
	for _, p := range prereqs {
		assert.True(t, p.selected, "Prerequisites from runaiData should be selected")
	}
}

func TestUpdateSelection_MultipleSelections(t *testing.T) {
	runaiData := &model.RunAIInstallData{
		Config: &config.InstallConfig{},
	}

	prereq1 := preReqMocks.NewMockPrerequisite(t)
	prereq1.On("GetName").Return("GPU Operator").Maybe()
	prereq1.On("IsMandatory").Return(false).Maybe()
	prereq1.On("InstalledOnCluster", runaiData).Return(model.NotInstalled).Maybe()
	prereq1.On("IsDefaultSelected").Return(false).Maybe()

	prereq2 := preReqMocks.NewMockPrerequisite(t)
	prereq2.On("GetName").Return("prometheus").Maybe()
	prereq2.On("IsMandatory").Return(false).Maybe()
	prereq2.On("InstalledOnCluster", runaiData).Return(model.NotInstalled).Maybe()
	prereq2.On("IsDefaultSelected").Return(false).Maybe()

	prereq3 := preReqMocks.NewMockPrerequisite(t)
	prereq3.On("GetName").Return("Network Operator").Maybe()
	prereq3.On("IsMandatory").Return(false).Maybe()
	prereq3.On("InstalledOnCluster", runaiData).Return(model.NotInstalled).Maybe()
	prereq3.On("IsDefaultSelected").Return(false).Maybe()

	prereqMap := map[string]prereqsteps.Prerequisite{
		"GPU Operator":     prereq1,
		"prometheus":       prereq2,
		"Network Operator": prereq3,
	}

	screen := NewPrerequisitesScreen(prereqMap, nil)
	screen.Init(runaiData)

	// Simulate selecting multiple prerequisites
	for _, opt := range screen.OptionsModel.GetOptions() {
		label := opt.(*PrerequisiteInstallnData).prerequisite.GetName()
		if label == "GPU Operator" || label == "prometheus" {
			opt.SetSelected(true)
		} else {
			opt.SetSelected(false)
		}
	}

	screen.UpdateSelection(runaiData)

	assert.Len(t, runaiData.Prerequisites, 2)
	assert.Contains(t, runaiData.Prerequisites, "GPU Operator")
	assert.Contains(t, runaiData.Prerequisites, "prometheus")
	assert.NotContains(t, runaiData.Prerequisites, "Network Operator")
}

func TestUpdateSelection_NoSelections(t *testing.T) {
	runaiData := &model.RunAIInstallData{
		Config: &config.InstallConfig{},
	}

	prereq1 := preReqMocks.NewMockPrerequisite(t)
	prereq1.On("GetName").Return("GPU Operator").Maybe()
	prereq1.On("IsMandatory").Return(false).Maybe()
	prereq1.On("InstalledOnCluster", runaiData).Return(model.NotInstalled).Maybe()
	prereq1.On("IsDefaultSelected").Return(false).Maybe()

	prereqMap := map[string]prereqsteps.Prerequisite{
		"GPU Operator": prereq1,
	}

	screen := NewPrerequisitesScreen(prereqMap, nil)
	screen.Init(runaiData)

	// Simulate no selections
	for _, opt := range screen.OptionsModel.GetOptions() {
		opt.SetSelected(false)
	}

	screen.UpdateSelection(runaiData)

	assert.Len(t, runaiData.Prerequisites, 0)
}
