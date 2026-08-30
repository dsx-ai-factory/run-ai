package screens

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	prereqsteps "github.com/run-ai/runai/runai-installer/installer/internal/prereq/components"
	viewscommon "github.com/run-ai/runai/runai-installer/installer/internal/tui/screens/common"
)

const (
	prereqTitle = "Select software dependencies to install"
)

type PrerequisiteInstallnData struct {
	prerequisite prereqsteps.Prerequisite
	installPhase model.InstallPhase
	selected     bool
	disabled     bool
	id           int
	errorMessage string
	errorCause   string
}

func normalizeVersion(version string) string {
	v := strings.TrimSpace(version)
	if v == "" {
		return v
	}
	// Strip any existing leading 'v' or 'V' and re-add lowercase 'v' for consistency
	if len(v) > 0 && (v[0] == 'v' || v[0] == 'V') {
		v = v[1:]
	}
	return "v" + v
}

func getPrereqStatusWithTimeout(pr prereqsteps.Prerequisite, runaiData *model.RunAIInstallData, d time.Duration) model.InstallPhase {
	resultCh := make(chan model.InstallPhase, 1)
	go func() {
		// Use InstalledOnCluster for initial "already installed" check (blocking call)
		resultCh <- pr.InstalledOnCluster(runaiData)
	}()
	select {
	case phase := <-resultCh:
		return phase
	case <-time.After(d):
		return model.Timeout
	}
}

func (p *PrerequisiteInstallnData) GetUiLabel() string {
	uiLabel := fmt.Sprintf("%s - %s", p.prerequisite.GetName(), normalizeVersion(p.prerequisite.GetVersion()))
	switch p.installPhase {
	case model.Installed:
		return fmt.Sprintf("%s - already installed", uiLabel)
	case model.Timeout:
		return fmt.Sprintf("%s - installation status check timed out", uiLabel)
	default:
		return uiLabel
	}
}

func (p *PrerequisiteInstallnData) GetOrder() int {
	return p.id
}

func (p *PrerequisiteInstallnData) IsMandatory() bool {
	return p.prerequisite.IsMandatory()
}

func (p *PrerequisiteInstallnData) IsSelected() bool {
	return p.selected
}

func (p *PrerequisiteInstallnData) SetSelected(selected bool) {
	p.selected = selected
}

func (p *PrerequisiteInstallnData) IsDisabled() bool {
	return p.disabled
}

type PrerequisitesScreen struct {
	*viewscommon.OptionsModel
	runaiData        *model.RunAIInstallData
	prerequisitesMap map[string]prereqsteps.Prerequisite
	nextScreen       viewscommon.Screen
}

func NewPrerequisitesScreen(prerequisitesMap map[string]prereqsteps.Prerequisite, nextScreen viewscommon.Screen) *PrerequisitesScreen {
	return &PrerequisitesScreen{
		prerequisitesMap: prerequisitesMap,
		nextScreen:       nextScreen,
	}
}

func (s *PrerequisitesScreen) Init(runaiData *model.RunAIInstallData) tea.Cmd {
	s.runaiData = runaiData
	prerequisites := initPrerequisitesScreen(runaiData, s.prerequisitesMap)
	opts := make([]viewscommon.Options, len(prerequisites))
	for i, p := range prerequisites {
		opts[i] = p
	}
	optionsScreen := viewscommon.NewOptionsModel(opts, true, runaiData.Config.DevMode)
	s.OptionsModel = optionsScreen
	return nil
}

func (s *PrerequisitesScreen) GetHeader() string {
	return prereqTitle
}

func (s *PrerequisitesScreen) GetMessages() []string {
	return nil
}

func (s *PrerequisitesScreen) NextScreen() viewscommon.Screen {
	return s.nextScreen
}

func (s *PrerequisitesScreen) UpdateSelection(runaiData *model.RunAIInstallData) {
	runaiData.Prerequisites = []string{}
	for _, option := range s.OptionsModel.GetOptions() {
		runaiPrerequisite := option.(*PrerequisiteInstallnData)
		if option.IsSelected() && runaiPrerequisite.installPhase == model.NotInstalled {
			runaiData.Prerequisites = append(runaiData.Prerequisites, runaiPrerequisite.prerequisite.GetName())
		}
	}
}

func (s *PrerequisitesScreen) Render() string {
	if s.OptionsModel == nil {
		return ""
	}
	return s.OptionsModel.Render()
}

func (s *PrerequisitesScreen) CanProceed() bool {
	if s.OptionsModel == nil {
		return true // Allow proceeding if not initialized (though unusual)
	}
	return s.OptionsModel.CanProceed()
}

func (s *PrerequisitesScreen) OnFocus() {
	if s.OptionsModel != nil {
		s.OptionsModel.OnFocus()
	}
}

func (s *PrerequisitesScreen) OnBlur() {
	if s.OptionsModel != nil {
		s.OptionsModel.OnBlur()
	}
}

func initPrerequisitesScreen(runaiData *model.RunAIInstallData, prerequisitesMap map[string]prereqsteps.Prerequisite) []*PrerequisiteInstallnData {
	// Convert map to slice
	prerequisites := make([]prereqsteps.Prerequisite, 0, len(prerequisitesMap))
	for _, prereq := range prerequisitesMap {
		prerequisites = append(prerequisites, prereq)
	}

	prerequisitesScreenArr := make([]*PrerequisiteInstallnData, len(prerequisites))
	var wg sync.WaitGroup
	wg.Add(len(prerequisites))

	for i, prereq := range prerequisites {
		go func(idx int, p prereqsteps.Prerequisite) {
			defer wg.Done()

			uiOrder, ok := GetPrerequisiteViewOrder()[p.GetName()]
			if !ok {
				uiOrder = len(prerequisites) + 1
			}

			preReqData := initPrerequisitesScreenData(runaiData, p, uiOrder)
			prerequisitesScreenArr[idx] = preReqData
		}(i, prereq)
	}

	wg.Wait()

	sort.Slice(prerequisitesScreenArr, func(i, j int) bool {
		iOrder, ok := GetPrerequisiteViewOrder()[prerequisitesScreenArr[i].prerequisite.GetName()]
		if !ok {
			iOrder = len(prerequisites)
		}
		jOrder, ok := GetPrerequisiteViewOrder()[prerequisitesScreenArr[j].prerequisite.GetName()]
		if !ok {
			jOrder = len(prerequisites)
		}
		return iOrder < jOrder
	})

	return prerequisitesScreenArr
}

func initPrerequisitesScreenData(runaiData *model.RunAIInstallData, p prereqsteps.Prerequisite, uiOrder int) *PrerequisiteInstallnData {
	installPhase := getPrereqStatusWithTimeout(p, runaiData, 4*time.Second)

	selected := p.IsMandatory() || p.IsDefaultSelected()
	for _, name := range runaiData.Prerequisites {
		if name == p.GetName() {
			selected = true
			break
		}
	}

	return &PrerequisiteInstallnData{
		installPhase: installPhase,
		selected:     selected,
		disabled:     installPhase == model.Installed,
		prerequisite: p,
		id:           uiOrder,
		errorMessage: "",
		errorCause:   "",
	}
}
