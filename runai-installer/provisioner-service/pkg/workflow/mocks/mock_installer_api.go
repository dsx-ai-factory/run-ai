package mocks

import (
	"time"

	"github.com/run-ai/runai/runai-installer/installer/pkg/api"
	"github.com/run-ai/runai/runai-installer/provisioner-service/pkg/workflow"
)

type MockInstallerAPI struct {
	InitFunc     func(config *api.Config) error
	GetFunc      func() workflow.InstallerInterface
	EventsToSend []MockEvent
}

type MockEvent struct {
	FlowName   string
	FlowStatus api.InstallPhase
	Delay      time.Duration
}

func NewMockInstallerAPI() *MockInstallerAPI {
	return &MockInstallerAPI{
		EventsToSend: []MockEvent{},
	}
}

func (m *MockInstallerAPI) Init(config *api.Config) error {
	if m.InitFunc != nil {
		return m.InitFunc(config)
	}
	return nil
}

func (m *MockInstallerAPI) Get() workflow.InstallerInterface {
	if m.GetFunc != nil {
		return m.GetFunc()
	}
	return &MockInstaller{events: m.EventsToSend}
}

type MockInstaller struct {
	events []MockEvent
}

func (m *MockInstaller) RunInstallationWithEventListner() (<-chan *api.FlowStatusReportingMsg, error) {
	eventsChan := make(chan *api.FlowStatusReportingMsg, len(m.events))

	go func() {
		defer close(eventsChan)
		for _, event := range m.events {
			if event.Delay > 0 {
				time.Sleep(event.Delay)
			}
			eventsChan <- &api.FlowStatusReportingMsg{
				FlowName:   event.FlowName,
				FlowStatus: event.FlowStatus,
			}
		}
	}()

	return eventsChan, nil
}

func (m *MockInstallerAPI) WithEvents(events []MockEvent) *MockInstallerAPI {
	m.EventsToSend = events
	return m
}

func (m *MockInstallerAPI) WithSuccessfulInstallation() *MockInstallerAPI {
	m.EventsToSend = []MockEvent{
		{FlowName: "prerequisites", FlowStatus: api.Installing, Delay: 200 * time.Millisecond},
		{FlowName: "prerequisites", FlowStatus: api.Installed, Delay: 200 * time.Millisecond},
		{FlowName: "control-plane", FlowStatus: api.Installing, Delay: 200 * time.Millisecond},
		{FlowName: "control-plane", FlowStatus: api.Installed, Delay: 200 * time.Millisecond},
	}
	return m
}

func (m *MockInstallerAPI) WithFailedInstallation(failedFlow string) *MockInstallerAPI {
	m.EventsToSend = []MockEvent{
		{FlowName: "prerequisites", FlowStatus: api.Installing, Delay: 200 * time.Millisecond},
		{FlowName: "prerequisites", FlowStatus: api.Installed, Delay: 200 * time.Millisecond},
		{FlowName: failedFlow, FlowStatus: api.Installing, Delay: 200 * time.Millisecond},
		{FlowName: failedFlow, FlowStatus: api.Failed, Delay: 200 * time.Millisecond},
	}
	return m
}
