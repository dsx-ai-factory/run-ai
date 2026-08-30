package orchestrator

import (
	"context"
	"sync"
	"time"

	"github.com/run-ai/runai/runai-installer/installer/internal/flow"
	"github.com/run-ai/runai/runai-installer/installer/internal/logging"
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
)

// InstallationExecutor provides an event-driven interface to the installation process.
//
// Architecture Overview:
//
//	Goroutine: executeFlows()
//	  ├─ Runs FlowManager.ExecuteFlow() for each flow group sequentially
//	  ├─ FlowManager sends FlowStatusReportingMsg events to externalEvents channel
//	  └─ Closes externalEvents when all flows complete
//
// The executor guarantees:
//   - Flow groups are executed sequentially
//   - Events are sent to externalEvents channel in order
//   - Clean cancellation via Context
//   - No goroutine leaks
//
// FlowGroup represents a group of installation flows with their manager
type FlowGroup struct {
	Flows   []*flow.InstallFlow
	Manager *flow.FlowManager
}

// NewFlowGroup creates a new FlowGroup with the given flows
func NewFlowGroup(flows []*flow.InstallFlow, externalEvents chan *flow.FlowStatusReportingMsg) *FlowGroup {
	return &FlowGroup{
		Flows:   flows,
		Manager: flow.NewFlowManagerWithExternalEvents(flows, externalEvents, false),
	}
}

type InstallationExecutor struct {
	installer          *Installer
	runaiData          *model.RunAIInstallData
	ctx                context.Context
	cancel             context.CancelFunc
	wg                 sync.WaitGroup
	flowGroups         []*FlowGroup
	statusPollInterval time.Duration
	externalEvents     chan *flow.FlowStatusReportingMsg
}

func NewInstallationExecutor(installer *Installer, runaiData *model.RunAIInstallData) *InstallationExecutor {
	executionTimeout := runaiData.Config.ExecutionTimeout
	if executionTimeout == 0 {
		executionTimeout = 50 * time.Minute
	}
	ctx, cancel := context.WithTimeout(context.Background(), executionTimeout)

	executor := &InstallationExecutor{
		installer:          installer,
		runaiData:          runaiData,
		ctx:                ctx,
		cancel:             cancel,
		statusPollInterval: 2 * time.Second,
		externalEvents:     make(chan *flow.FlowStatusReportingMsg, 100),
	}

	return executor
}

// Start begins the installation process and starts emitting events
// This method returns immediately. Use Wait() to block until completion.
func (e *InstallationExecutor) Start() <-chan *flow.FlowStatusReportingMsg {
	logging.Infof("================================================================================")
	logging.Infof("=           Starting Run:AI Installation with Executor (Library Mode)         =")
	logging.Infof("================================================================================")

	e.flowGroups = BuildFlowGroups(e.installer.deps, e.runaiData, e.externalEvents)

	e.wg.Add(1)
	go e.executeFlows()

	return e.externalEvents
}

// Wait blocks until the installation completes or the context is cancelled
func (e *InstallationExecutor) Wait() error {
	e.wg.Wait()
	return nil
}

// Cancel stops the installation process
func (e *InstallationExecutor) Cancel() {
	e.cancel()
}

// executeFlows runs the installation flows sequentially
func (e *InstallationExecutor) executeFlows() {
	defer e.wg.Done()
	defer close(e.externalEvents)

	// Execute each flow group sequentially
	for i, flowGroup := range e.flowGroups {
		if flowGroup == nil || len(flowGroup.Flows) == 0 {
			logging.Infof("Skipping flow group %d/%d: no flows to execute", i+1, len(e.flowGroups))
			continue
		}

		// Check if context was cancelled before starting new flow group
		select {
		case <-e.ctx.Done():
			logging.Errorf("Installation cancelled via context timeout before flow group %d/%d", i+1, len(e.flowGroups))
			e.externalEvents <- &flow.FlowStatusReportingMsg{
				FlowStatus: model.Timeout,
				FlowName:   flowGroup.Flows[0].GetName(),
			}
			return
		default:
		}

		logging.Infof("Starting flow group %d/%d...", i+1, len(e.flowGroups))

		// Create internal events channel for flow manager communication
		internalEvents := make(chan interface{}, 100)

		// Start the flow execution with context
		flowGroup.Manager.ExecuteFlow(e.ctx, e.runaiData, internalEvents)

		// Wait for the manager to complete
		// The manager will send FlowStatusReportingMsg events to e.externalEvents as flows complete
		flowGroup.Manager.Wait()

		close(internalEvents)

		if flowGroup.Manager.HasFailed() {
			logging.Errorf("Flow group %d/%d failed, stopping installation", i+1, len(e.flowGroups))
			return
		}

		logging.Infof("Flow group %d/%d completed", i+1, len(e.flowGroups))
	}

	logging.Infof("All installation flows completed")
}
