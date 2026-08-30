package displayer

import "time"

// Type definitions
type Phase string

type General struct {
	Name                 string             `json:"name"`
	ID                   string             `json:"id"`
	DistributedFramework *string            `json:"distributedFramework,omitempty"`
	Type                 string             `json:"type"`
	Phase                Phase              `json:"phase"`
	PhaseMessage         *string            `json:"phaseMessage,omitempty"`
	Project              string             `json:"project"`
	Department           string             `json:"department"`
	CreatedAt            time.Time          `json:"createdAt"`
	UpdatedAt            time.Time          `json:"updatedAt"`
	SubmittedBy          *string            `json:"submittedBy,omitempty"`
	Priority             *string            `json:"priority,omitempty"`
	Category             string             `json:"category"`
	Topology             *string            `json:"topology,omitempty"`
	EnvironmentVariables *map[string]string `json:"environmentVariables,omitempty"`
	Arguments            *string            `json:"arguments,omitempty"`
	Command              *string            `json:"command,omitempty"`
	RunaiCliCommand      *string            `json:"cliCommand,omitempty"`
}
