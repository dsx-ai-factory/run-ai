package page

import (
	"fmt"
	"time"

	ousc "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
	wsc "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

type EntryAdapter interface {
	GetId() string
	GetCreatedAt() time.Time
}

func GetEntryAdapter(entry any) (EntryAdapter, error) {
	switch castedEntry := entry.(type) {
	case wsc.Workload:
		return NewWorkloadEntryAdapter(&castedEntry), nil
	case ousc.Project:
		return NewProjectEntryAdapter(&castedEntry), nil
	case wsc.AIApplication:
		return NewAIApplicationEntryAdapter(&castedEntry), nil
	default:
		return nil, fmt.Errorf("unknown pager entry type: %T", entry)
	}
}
