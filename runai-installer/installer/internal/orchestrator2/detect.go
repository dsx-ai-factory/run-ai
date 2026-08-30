package orchestrator2

type DetectPhase string

const (
	DetectNotInstalled DetectPhase = "not_installed"
	DetectInstalled    DetectPhase = "installed"
)

type DetectResult struct {
	Phase   DetectPhase
	Version string
}
