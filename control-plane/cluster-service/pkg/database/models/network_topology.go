package models

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"

	"github.com/run-ai/runai/runai-application-packages/datastructs"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/run-ai/runai/runai-application-packages/queryapi"
	"gorm.io/gorm"
)

var NetworkTopologyQueryParser = queryapi.NewParser(networkTopologyColumnMapping())

type GroveTopologyCondition struct {
	Type               string `json:"type"`
	Status             string `json:"status"`
	Reason             string `json:"reason"`
	Message            string `json:"message"`
	ObservedGeneration int64  `json:"observedGeneration,omitempty"`
}

type GroveTopologyConditions []GroveTopologyCondition

type Grove struct {
	Name       string                  `json:"name" gorm:"column:grove_name"`
	Domains    pq.StringArray          `json:"domains" gorm:"column:grove_domains;type:text[]"`
	Conditions GroveTopologyConditions `json:"conditions" gorm:"column:grove_conditions;serializer:json;type:jsonb"`
}

const (
	NetworkTopologyTable     = "network_topologies"
	NetworkTopologyViewTable = "network_topology_view"
)

// NetworkTopology phase constants
const (
	NetworkTopologyPhaseCreating = "Creating"
	NetworkTopologyPhaseDeleted  = "Deleted"
	NetworkTopologyPhaseReady    = "Ready"
)

type NetworkTopology struct {
	CreatedAt time.Time
	ID        uuid.UUID      `gorm:"primaryKey"`
	Name      string         `gorm:"not null"`
	Levels    pq.StringArray `gorm:"type:text[];not null"`
	Source    string         `gorm:"not null"`
	ClusterID uuid.UUID      `gorm:"not null;column:cluster_id"`
	TenantID  int32          `gorm:"not null;column:tenant_id"`
	CreatedBy *string        `gorm:"column:created_by"`
	Phase     string         `gorm:"not null"` // Creating - entry added to DB, (cluster-sync when staus is empty). Ready
	Grove     *Grove         `gorm:"embedded"`
	DeletedAt gorm.DeletedAt
}

type NetworkTopologyView struct {
	NetworkTopology
	NodePools datatypes.JSON `gorm:"column:nodepools"`
}

func (n *NetworkTopologyView) GetNodepools() []string {
	if n == nil || n.NodePools == nil {
		return nil
	}
	if len(n.NodePools) > 0 {
		var names []string
		if err := json.Unmarshal(n.NodePools, &names); err == nil {
			return names
		}
	}
	return nil
}

func (NetworkTopology) TableName() string {
	return NetworkTopologyTable
}

func (NetworkTopology) PrimaryKey() string {
	return "id"
}

func (nt *NetworkTopology) BeforeCreate(tx *gorm.DB) (err error) {
	nt.ID = uuid.New()
	return
}

func (nt *NetworkTopology) AfterFind(tx *gorm.DB) (err error) {
	if nt.Grove != nil && nt.Grove.Name == "" && len(nt.Grove.Domains) == 0 && len(nt.Grove.Conditions) == 0 {
		nt.Grove = nil
	}
	return
}

func networkTopologyColumnMapping() queryapi.APIFieldsMapping {
	return queryapi.APIFieldsMapping{
		"name":      queryapi.NewFieldMapping("name"),
		"clusterId": queryapi.NewFieldMapping("cluster_id", queryapi.UUIDValidator{}),
		"source":    queryapi.NewFieldMapping("source"),
		"phase":     queryapi.NewFieldMapping("phase"),
	}
}

// DeepCopyInto performs a deep copy of the NetworkTopology into the target
func (nt *NetworkTopology) DeepCopyInto(target *NetworkTopology) {
	if target == nil {
		return
	}

	// Copy embedded Common struct
	target.CreatedAt = nt.CreatedAt

	// Copy all fields
	target.ID = nt.ID
	target.Name = nt.Name
	target.Source = nt.Source
	target.ClusterID = nt.ClusterID
	target.TenantID = nt.TenantID
	target.Phase = nt.Phase
	target.DeletedAt = nt.DeletedAt

	// Deep copy the Levels slice
	target.Levels = nil
	if nt.Levels != nil {
		target.Levels = make([]string, len(nt.Levels))
		copy(target.Levels, nt.Levels)
	}

	// Copy pointer fields
	target.CreatedBy = nil
	if nt.CreatedBy != nil {
		createdBy := *nt.CreatedBy
		target.CreatedBy = &createdBy
	}

	target.Grove = nil
	if nt.Grove != nil {
		groveCopy := *nt.Grove
		if nt.Grove.Domains != nil {
			groveCopy.Domains = make([]string, len(nt.Grove.Domains))
			copy(groveCopy.Domains, nt.Grove.Domains)
		}
		if nt.Grove.Conditions != nil {
			groveCopy.Conditions = make(GroveTopologyConditions, len(nt.Grove.Conditions))
			copy(groveCopy.Conditions, nt.Grove.Conditions)
		}
		target.Grove = &groveCopy
	}
}

func (NetworkTopologyView) TableName() string {
	return NetworkTopologyViewTable
}

type NetworkTopologyArrayAdapter []NetworkTopology

func NewNetworkTopologyArrayAdapter(arr []NetworkTopology) NetworkTopologyArrayAdapter {
	return NetworkTopologyArrayAdapter(arr)
}

func (nta NetworkTopologyArrayAdapter) ToIds() []uuid.UUID {
	var ids []uuid.UUID
	for _, topology := range nta {
		ids = append(ids, topology.ID)
	}
	return ids
}

func (nta NetworkTopologyArrayAdapter) ToNames() datastructs.Set[string] {
	names := datastructs.Set[string]{}
	for _, topology := range nta {
		names.Add(topology.Name)
	}
	return names
}
