package common

import "errors"

const Rules = "rules"
const ItemRules = "itemRules"
const KeyRules = "keyRules"
const Items = "items"

const OptionsValue = "value"
const OptionsDisabled = "disabled"

const UsedAsTemplate = "Template"
const UsedForSubmit = "Submit"

// for each value we want to know which level in the stack (policy or workloads) it has been
// merged from, and if it has been deleted in certain level - in which level it has been received

const Value = "value"
const Values = "values"
const ValueFromLevel = "valueFromLevel"
const Deleted = "deleted"
const DeletedLevel = "deletionLevel"
const Level = "level"           // this is for the frontend
const ItemDefaults = "DEFAULTS" // item providing defaults for itemized field

// generation status codes

const StatusFailed = "Failed"
const StatusSubmitted = "Submitted"

const GroupsSeparator = ";"

// Phase is an enum of workload lifecycle phases.
type Phase string

const (
	PhaseCreating     Phase = "Creating"
	PhaseInitializing Phase = "Initializing"
	PhaseRunning      Phase = "Running"
	PhaseDegraded     Phase = "Degraded"
	PhasePending      Phase = "Pending"
	PhaseSuspended    Phase = "Suspended"
	PhaseCompleted    Phase = "Completed"
	PhaseFailed       Phase = "Failed"
	PhaseTerminating  Phase = "Terminating"
	PhaseSuspending   Phase = "Suspending"
	PhaseResuming     Phase = "Resuming"
	PhaseStopped      Phase = "Stopped"    // PhaseStopped is used by inference
	PhaseRestarting   Phase = "Restarting" // PhaseRestarting is used by restart controller
)

var phaseFromString = map[string]Phase{
	"Creating":     PhaseCreating,
	"Initializing": PhaseInitializing,
	"Running":      PhaseRunning,
	"Degraded":     PhaseDegraded,
	"Pending":      PhasePending,
	"Suspended":    PhaseSuspended,
	"Completed":    PhaseCompleted,
	"Failed":       PhaseFailed,
	"Terminating":  PhaseTerminating,
	"Suspending":   PhaseSuspending,
	"Resuming":     PhaseResuming,
	"Stopped":      PhaseStopped,
	"Restarting":   PhaseRestarting,
}

func (p Phase) String() string {
	if _, ok := phaseFromString[string(p)]; ok {
		return string(p)
	}
	return "Unknown"
}

func PhaseStringToEnum(value string) (Phase, error) {
	if v, ok := phaseFromString[value]; ok {
		return v, nil
	}
	return "", errors.New("undefined phase:" + value)
}

// IsFinitePhase reports whether the given phase string represents a finite
// workload phase (Completed or Failed). Status.CompletionTime is deliberately
// sticky (see workload_status_calculator.go), so callers that gate on
// "completed" must also check phase to avoid acting on a workload that has
// since transitioned back to a running phase (e.g. via the restart flow).
func IsFinitePhase(phase string) bool {
	return phase == string(PhaseCompleted) || phase == string(PhaseFailed)
}
