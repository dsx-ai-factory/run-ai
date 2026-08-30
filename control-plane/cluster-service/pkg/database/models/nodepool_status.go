package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/lib/pq"
	"github.com/run-ai/runai/control-plane/cluster-service/internal/generated/pkg/api"
)

const (
	StatusLabelKeyColumn                   = "status_label_key"
	StatusLabelValueColumn                 = "status_label_value"
	StatusGpuPlacementStrategyColumn       = "status_gpu_placement_strategy"
	StatusGpuDevicePlacementStrategyColumn = "status_gpu_device_placement_strategy"
	StatusCpuPlacementStrategyColumn       = "status_cpu_placement_strategy"
	StatusNodesColumn                      = "status_nodes"
	StatusGPUNetworkAccelerationDetected   = "status_gpu_network_acceleration_detected"
	StatusMinGuaranteedRuntimeColumn       = "status_min_guaranteed_runtime"
	StatusConditions                       = "status_conditions"
)

type NodepoolConditionsDetails []api.NodepoolConditionDetails

// Scan implements the sql.Scanner interface
func (c *NodepoolConditionsDetails) Scan(value any) error {
	if value == nil {
		*c = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan NodepoolConditionsDetails: invalid type")
	}

	// Try to unmarshal as an array first
	var conditions []api.NodepoolConditionDetails
	if err := json.Unmarshal(bytes, &conditions); err == nil {
		*c = conditions
		return nil
	}

	// If that fails, try to unmarshal as a single object and wrap it in an array
	var singleCondition api.NodepoolConditionDetails
	if err := json.Unmarshal(bytes, &singleCondition); err == nil {
		*c = []api.NodepoolConditionDetails{singleCondition}
		return nil
	}

	return fmt.Errorf("failed to unmarshal NodepoolConditionsDetails")
}

// Value implements the driver.Valuer interface
func (c NodepoolConditionsDetails) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	bytes, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

type NodepoolStatus struct {
	LabelKey                       *string
	LabelValue                     *string
	GpuPlacementStrategy           api.PlacementStrategy
	GpuDevicePlacementStrategy     api.PlacementStrategy
	CpuPlacementStrategy           api.PlacementStrategy
	Nodes                          pq.StringArray `gorm:"type:text[]"`
	GPUNetworkAccelerationDetected bool           `gorm:"not null"`

	MinGuaranteedRuntime *string
	Conditions           NodepoolConditionsDetails `gorm:"serializer:json;type:jsonb"`
}

func (n NodepoolStatus) DeepCopy() NodepoolStatus {
	return NodepoolStatus{
		LabelKey:                       n.LabelKey,
		LabelValue:                     n.LabelValue,
		GpuPlacementStrategy:           n.GpuPlacementStrategy,
		GpuDevicePlacementStrategy:     n.GpuDevicePlacementStrategy,
		CpuPlacementStrategy:           n.CpuPlacementStrategy,
		Nodes:                          n.Nodes,
		GPUNetworkAccelerationDetected: n.GPUNetworkAccelerationDetected,
		MinGuaranteedRuntime:           n.MinGuaranteedRuntime,
		Conditions:                     n.Conditions,
	}
}
