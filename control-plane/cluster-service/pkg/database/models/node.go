/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	NodeTable = "node"

	primaryKey = "id"
)

var NodeQueryParser = queryapi.NewParser(nodeColumnMapping())

type (
	NodeStatusConditionsDetails  []api.NodeStatusConditionDetails
	NodeTaints                   []api.NodeTaint
	NetworkTopologyLabels        []NetworkTopologyLabel
	NetworkTopologyLabelsAdapter NetworkTopologyLabels
)

// NetworkTopologyLabel represents a single topology level and its corresponding label value
type NetworkTopologyLabel struct {
	LabelKey   string `json:"LabelKey"`
	LabelValue string `json:"LabelValue"`
}

func (t *NetworkTopologyLabel) ToApiModel() api.NetworkTopologyLabel {
	return api.NetworkTopologyLabel{
		LabelKey:   t.LabelKey,
		LabelValue: t.LabelValue,
	}
}

func (ta NetworkTopologyLabelsAdapter) ToApiModels() *[]api.NetworkTopologyLabel {
	if ta == nil {
		return nil
	}
	result := make([]api.NetworkTopologyLabel, len(ta))
	for i, tEntry := range ta {
		result[i] = tEntry.ToApiModel()
	}
	return &result
}

type Node struct {
	Common                     `gorm:"embedded"`
	ID                         uuid.UUID `gorm:"primaryKey"`
	TenantId                   int32
	Status                     string                      `gorm:"not null"`
	Conditions                 NodeStatusConditionsDetails `gorm:"serializer:json;type:jsonb"`
	Taints                     NodeTaints                  `gorm:"serializer:json;type:jsonb"`
	NodePool                   string                      `gorm:"not null"`
	GpuType                    string                      `gorm:"not null"`
	GpuCount                   int32                       `gorm:"not null"`
	GpuMemoryBytes             int64                       `gorm:"not null"`
	Name                       string                      `gorm:"uniqueIndex:unique_node_name_per_cluster;not null"`
	ClusterUuid                string                      `gorm:"uniqueIndex:unique_node_name_per_cluster;index:idx_cluster;not null"`
	NvLinkDomainUid            *string                     `gorm:""`
	NvLinkCliqueId             *string                     `gorm:""`
	NetworkTopologyLabelValues *pq.StringArray             `gorm:"type:text[]"`
	NetworkTopologyLabels      *NetworkTopologyLabels      `gorm:"serializer:json;type:jsonb"`
	TopologyManagerPolicy      *string                     `gorm:""`
}

func (n *Node) BeforeCreate(tx *gorm.DB) error {
	n.ID = uuid.New()
	return nil
}

func (Node) TableName() string {
	return NodeTable
}

func (Node) PrimaryKey() string {
	return primaryKey
}

func (Node) ConflictColumns() []clause.Column {
	return []clause.Column{{Name: "name"}, {Name: "cluster_uuid"}}
}

func (Node) ColumnsToUpdate() []string {
	// TODO tenant id should not be updatable, it should be part of the conflict columns and add to the unique index
	return []string{
		"updated_at", "status", "conditions", "taints", "node_pool", "gpu_type", "gpu_count", "gpu_memory_bytes", "tenant_id",
		"nv_link_clique_id", "nv_link_domain_uid", "network_topology_label_values", "network_topology_labels",
		"topology_manager_policy",
	}
}

// InsertColumns returns the list of column names for INSERT statements
func (n Node) InsertColumns() []string {
	return []string{
		"id", "tenant_id", "cluster_uuid", "name", "status", "conditions", "taints",
		"node_pool", "gpu_type", "gpu_count", "gpu_memory_bytes", "created_at", "updated_at",
		"nv_link_domain_uid", "nv_link_clique_id",
		"network_topology_label_values", "network_topology_labels",
		"topology_manager_policy",
	}
}

// InsertColumnsSQL returns comma-separated column names for INSERT INTO clause
func (n Node) InsertColumnsSQL() string {
	return strings.Join(n.InsertColumns(), ", ")
}

// InsertPlaceholdersSQL returns comma-separated placeholders (?) for VALUES clause
func (n Node) InsertPlaceholdersSQL() string {
	cols := n.InsertColumns()
	placeholders := make([]string, len(cols))
	for i := range cols {
		placeholders[i] = "?"
	}
	return strings.Join(placeholders, ", ")
}

// UpdateSetSQL returns the SET clause for ON CONFLICT DO UPDATE
func (n Node) UpdateSetSQL() string {
	cols := n.ColumnsToUpdate()
	setClauses := make([]string, len(cols))
	for i, col := range cols {
		setClauses[i] = fmt.Sprintf("%s = EXCLUDED.%s", col, col)
	}
	return strings.Join(setClauses, ", ")
}

func nodeColumnMapping() queryapi.APIFieldsMapping {
	return queryapi.APIFieldsMapping{
		"name":            queryapi.NewFieldMapping("name"),
		"status":          queryapi.NewFieldMapping("status"),
		"nodePool":        queryapi.NewFieldMapping("node_pool"),
		"gpuType":         queryapi.NewFieldMapping("gpu_type"),
		"nvLinkDomainUid": queryapi.NewFieldMapping("nv_link_domain_uid"),
		"nvLinkCliqueId":  queryapi.NewFieldMapping("nv_link_clique_id"),
		"clusterUuid":     queryapi.NewFieldMapping("cluster_uuid"),
	}
}

func (c *NodeStatusConditionsDetails) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan NodeStatusConditionsDetails: invalid type")
	}
	return json.Unmarshal(bytes, c)
}

func (c NodeStatusConditionsDetails) Value() (driver.Value, error) {
	bytes, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (t *NodeTaints) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan NodeTaints: invalid type")
	}
	return json.Unmarshal(bytes, t)
}

func (t NodeTaints) Value() (driver.Value, error) {
	bytes, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (ta *NetworkTopologyLabels) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan NetworkTopologyLabels: invalid type")
	}
	return json.Unmarshal(bytes, ta)
}

func (ta NetworkTopologyLabels) Value() (driver.Value, error) {
	bytes, err := json.Marshal(ta)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

type NodesAdapter struct {
	nodes *[]Node
}

func NewNodesAdapter(arr *[]Node) *NodesAdapter {
	return &NodesAdapter{
		nodes: arr,
	}
}

func (na *NodesAdapter) ToApiModels() api.Nodes {
	nodes := api.Nodes{
		Nodes: []api.Node{},
	}
	if na.nodes == nil {
		return nodes
	}

	for _, node := range *na.nodes {
		nodes.Nodes = append(nodes.Nodes, node.ToApiModel())
	}
	return nodes
}

func (n *Node) ToApiModel() api.Node {
	nodeInfo := api.Node{
		Id:          n.ID.String(),
		Status:      n.Status,
		Conditions:  n.Conditions,
		Taints:      n.Taints,
		NodePool:    n.NodePool,
		CreatedAt:   n.CreatedAt,
		Name:        n.Name,
		ClusterUuid: n.ClusterUuid,
		UpdatedAt:   n.UpdatedAt,

		NvLinkDomainUid: n.NvLinkDomainUid,
		NvLinkCliqueId:  n.NvLinkCliqueId,
	}

	if n.TopologyManagerPolicy != nil {
		nodeInfo.Numa = &api.NumaInfo{
			TopologyManagerPolicy: (*api.TopologyManagerPolicy)(n.TopologyManagerPolicy),
		}
	}

	if n.NetworkTopologyLabels != nil {
		nodeInfo.NetworkTopologyLabels = NetworkTopologyLabelsAdapter(*n.NetworkTopologyLabels).ToApiModels()
	}

	if n.GpuType != "" || n.GpuCount != 0 {
		nodeInfo.GpuInfo = &api.GpuInfo{
			GpuType:        n.GpuType,
			GpuCount:       n.GpuCount,
			GpuMemoryBytes: n.GpuMemoryBytes,
		}
	}

	// NOTE TO DEV: when adding a field here, don't forget to add it to "ToExtendedApiModel" below as well

	return nodeInfo
}

func (n *Node) ToExtendedApiModel(numaNodes []NodeNuma) api.NodeExtended {
	node := n.ToApiModel()
	apiNumaNodes := make([]api.NumaNodeDetails, 0, len(numaNodes))
	for i := range numaNodes {
		apiNumaNodes = append(apiNumaNodes, numaNodes[i].ToApiModel())
	}

	numaInfo := api.NumaInfoExtended{
		Nodes: apiNumaNodes,
	}
	if node.Numa != nil {
		numaInfo.TopologyManagerPolicy = node.Numa.TopologyManagerPolicy
	}

	return api.NodeExtended{
		Id:                    node.Id,
		Status:                node.Status,
		Conditions:            node.Conditions,
		Taints:                node.Taints,
		NodePool:              node.NodePool,
		CreatedAt:             node.CreatedAt,
		Name:                  node.Name,
		ClusterUuid:           node.ClusterUuid,
		UpdatedAt:             node.UpdatedAt,
		NvLinkDomainUid:       node.NvLinkDomainUid,
		NvLinkCliqueId:        node.NvLinkCliqueId,
		Numa:                  numaInfo,
		NetworkTopologyLabels: node.NetworkTopologyLabels,
		GpuInfo:               node.GpuInfo,
	}
}

func (n *Node) ToInternalApiModel() api.NodeInternal {
	var networkTopologyLabelValues *[]string
	if n.NetworkTopologyLabelValues != nil {
		labelsValues := []string(*n.NetworkTopologyLabelValues)
		networkTopologyLabelValues = &labelsValues
	}

	return api.NodeInternal{
		Name:                       n.Name,
		ClusterUuid:                n.ClusterUuid,
		NetworkTopologyLabelValues: networkTopologyLabelValues,
	}
}

func (Node) ToNetworkTopologyLabelsAndValues(ntl *[]api.NetworkTopologyLabel) (*NetworkTopologyLabels, *pq.StringArray) {
	if ntl == nil {
		return nil, nil
	}
	networkTopologyLabels := make(NetworkTopologyLabels, len(*ntl))
	networkTopologyLabelValues := make(pq.StringArray, len(*ntl))
	for i, label := range *ntl {
		networkTopologyLabels[i] = NetworkTopologyLabel{
			LabelKey:   label.LabelKey,
			LabelValue: label.LabelValue,
		}
		networkTopologyLabelValues[i] = label.LabelValue
	}
	return &networkTopologyLabels, &networkTopologyLabelValues
}
