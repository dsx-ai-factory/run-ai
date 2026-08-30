package manipulate

import (
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

type BriefWorkloadType struct {
	Name           string `json:"name"`
	CategoryName   string `json:"categoryName,omitempty"`
	PriorityName   string `json:"priorityName,omitempty"`
	Preemptibility string `json:"preemptibility"`
	Group          string `json:"group"`
	Source         string `json:"source,omitempty"`
}

type WorkloadTypeAlignedRow struct {
	Name           string `json:"name,omitempty"`
	CategoryName   string `json:"categoryName,omitempty"`
	PriorityName   string `json:"priorityName,omitempty"`
	Preemptibility string `json:"preemptibility,omitempty"`
	Group          string `json:"group,omitempty"`
	Source         string `json:"source,omitempty"`
}

type WorkloadTypesAlignedOutput struct {
	WorkloadTypes []WorkloadTypeAlignedRow `json:"workloadTypes" yaml:"workloadTypes"`
}

type Data struct {
	WorkloadTypes []openapi.WorkloadTypeConfig `json:"workloadTypes,omitempty" yaml:"workloadTypes"`
}

func newBriefFromWorkloadType(wt openapi.WorkloadTypeConfig) BriefWorkloadType {
	brief := BriefWorkloadType{
		Name:           wt.Name,
		Group:          wt.Group,
		Preemptibility: string(wt.Preemptibility),
	}

	if wt.CategoryName != nil {
		brief.CategoryName = *wt.CategoryName
	}
	if wt.PriorityName != nil {
		brief.PriorityName = *wt.PriorityName
	}
	if wt.Source != nil {
		brief.Source = *wt.Source
	}

	return brief
}

func BriefListFromResponse(workloadTypes []openapi.WorkloadTypeConfig) ([]BriefWorkloadType, error) {
	var briefList = make([]BriefWorkloadType, 0)

	for _, wt := range workloadTypes {
		briefList = append(briefList, newBriefFromWorkloadType(wt))
	}

	return briefList, nil
}
