package models

type WorkloadNodeAffinity struct {
	SpecName     string       `json:"specName,omitempty"`
	NodeAffinity NodeAffinity `json:"nodeAffinity"`
}

type WorkloadNodeAffinityArray []WorkloadNodeAffinity
