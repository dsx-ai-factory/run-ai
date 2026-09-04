/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"strconv"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	"k8s.io/utils/ptr"

	"github.com/google/uuid"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"
	"gorm.io/gorm"
)

var (
	NodepoolQueryParser         = queryapi.NewParser(nodepoolColumnMapping())
	NodepoolMetadataQueryParser = queryapi.NewParser(nodepoolMetadataColumnMapping())
)

const (
	NodepoolTable = "nodepools"

	NameColumn                                 = "name"
	CreatedAtColumn                            = "created_at"
	UpdatedAtColumn                            = "updated_at"
	UpdatedByColumn                            = "updated_by"
	DeletedAtColumn                            = "deleted_at"
	DeletedByColumn                            = "deleted_by"
	ClusterIdColumn                            = "cluster_id"
	ClusterNameColumn                          = "cluster_name"
	PhaseColumn                                = "phase"
	PhaseMsgColumn                             = "phase_message"
	GpuNetworkAccelerationLabelKeyColumn       = "gpu_network_acceleration_label_key"
	GpuNetworkAccelerationDetectionColumn      = "gpu_network_acceleration_detection"
	StatusGpuNetworkAccelerationDetectedColumn = "status_gpu_network_acceleration_detected"
	NetworkTopologyNameColumn                  = "network_topology_name"
	AdjustableWeightsColumn                    = "allow_adjustable_weights"
	CpuQuotaColumn                             = "allow_cpu_quota"
	QuotaColumn                                = "allow_quota"
	RanksColumn                                = "allow_ranks"
	NumaAwareSchedulingColumn                  = "numa_aware_scheduling"

	DefaultAllowAdjustableWeights = false
	DefaultAllowQuota             = true
	DefaultAllowRanks             = true
	DefaultAllowCpuQuota          = false
)

type Nodepool struct {
	Common                          `gorm:"embedded"`
	ID                              int32  `gorm:"primaryKey"`
	Name                            string `gorm:"not null"`
	TenantId                        int32  `gorm:"not null"`
	LabelKey                        *string
	LabelValue                      *string
	Phase                           api.NodepoolPhase
	PhaseMessage                    *string
	ClusterId                       uuid.UUID `gorm:"not null"`
	ClusterName                     string    `gorm:"not null"`
	DeletedAt                       gorm.DeletedAt
	CreatedBy                       string `gorm:"not null"`
	UpdatedBy                       string `gorm:"not null"`
	DeletedBy                       *string
	IsDefault                       bool
	NodepoolStatus                  `gorm:"embedded;embeddedPrefix:status_"`
	GPUNetworkAccelerationDetection *api.NodepoolGpuNetworkAccelerationDetection
	GPUNetworkAccelerationLabelKey  string `gorm:"not null"`

	TimeBasedFairShare TimeBasedFairShare `gorm:"serializer:json;type:jsonb"`

	GpuPlacementStrategy       api.PlacementStrategy `gorm:"not null"`
	GpuDevicePlacementStrategy api.PlacementStrategy `gorm:"not null"`
	CpuPlacementStrategy       api.PlacementStrategy `gorm:"not null"`
	MinGuaranteedRuntime       *string

	NetworkTopologyId   *uuid.UUID
	NetworkTopologyName *string

	AllowAdjustableWeights bool
	AllowRanks             bool
	AllowQuota             bool
	AllowCpuQuota          bool

	NumaAwareScheduling *bool

	// NOTE TO DEV: if you add fields here,
	// don't forget to update in funcs "DeepCopyInto" and "clusterFieldsUpdated"
}

func (Nodepool) TableName() string {
	return NodepoolTable
}

func (Nodepool) PrimaryKey() string {
	return "id"
}

func nodepoolColumnMapping() queryapi.APIFieldsMapping {
	return queryapi.APIFieldsMapping{
		"name":                            queryapi.NewFieldMapping(NameColumn),
		"createdAt":                       queryapi.NewFieldMapping(CreatedAtColumn),
		"updatedAt":                       queryapi.NewFieldMapping(UpdatedAtColumn),
		"deletedAt":                       queryapi.NewFieldMapping(DeletedAtColumn),
		"clusterId":                       queryapi.NewFieldMapping(ClusterIdColumn, queryapi.UUIDValidator{}),
		"phase":                           queryapi.NewFieldMapping(PhaseColumn),
		"clusterName":                     queryapi.NewFieldMapping(ClusterNameColumn),
		"gpuNetworkAccelerationDetection": queryapi.NewFieldMapping(GpuNetworkAccelerationDetectionColumn),
		"gpuNetworkAccelerationDetected":  queryapi.NewFieldMapping(StatusGpuNetworkAccelerationDetectedColumn),
		"gpuNetworkAccelerationLabelKey":  queryapi.NewFieldMapping(GpuNetworkAccelerationLabelKeyColumn),
		"networkTopologyName":             queryapi.NewFieldMapping(NetworkTopologyNameColumn),
	}
}

// nodepoolMetadataColumnMapping restricts filtering and sorting to only the fields
// available in the metadata response.
func nodepoolMetadataColumnMapping() queryapi.APIFieldsMapping {
	return queryapi.APIFieldsMapping{
		"clusterId": queryapi.NewFieldMapping(ClusterIdColumn, queryapi.UUIDValidator{}),
	}
}

func (Nodepool) ParseNodepoolPhase(phase string) api.NodepoolPhase {
	return api.NodepoolPhase(phase)
}

func (Nodepool) ParsePlacementStrategy(placementStrategy string) api.PlacementStrategy {
	return api.PlacementStrategy(placementStrategy)
}

func (np *Nodepool) BeforeCreate(tx *gorm.DB) (err error) {
	if np.Phase == api.NODEPOOLPHASE_DELETED || (np.DeletedAt.Valid && np.DeletedAt.Time != time.Time{}) {
		np.alignDeleted()
	}

	return
}

func (np *Nodepool) BeforeUpdate(tx *gorm.DB) (err error) {
	if np.Phase == api.NODEPOOLPHASE_DELETED || (np.DeletedAt.Valid && np.DeletedAt.Time != time.Time{}) {
		np.alignDeleted()
	}

	return
}

func (np *Nodepool) BeforeDelete(tx *gorm.DB) (err error) {
	np.alignDeleted()
	return
}

func (np *Nodepool) alignDeleted() {
	np.Phase = api.NODEPOOLPHASE_DELETED
	np.DeletedAt = gorm.DeletedAt{
		Time:  time.Now(),
		Valid: true,
	}
}

func (np *Nodepool) DeepCopyInto(out *Nodepool) {
	out.CreatedAt = np.CreatedAt
	out.UpdatedAt = np.UpdatedAt
	out.ID = np.ID
	out.Name = np.Name
	out.TenantId = np.TenantId
	out.LabelKey = np.LabelKey
	out.LabelValue = np.LabelValue
	out.Phase = np.Phase
	out.PhaseMessage = np.PhaseMessage
	out.ClusterId = np.ClusterId
	out.ClusterName = np.ClusterName
	out.DeletedAt = np.DeletedAt
	out.CreatedBy = np.CreatedBy
	out.UpdatedBy = np.UpdatedBy
	out.DeletedBy = np.DeletedBy
	out.IsDefault = np.IsDefault
	out.GpuPlacementStrategy = np.GpuPlacementStrategy
	out.GpuDevicePlacementStrategy = np.GpuDevicePlacementStrategy
	out.CpuPlacementStrategy = np.CpuPlacementStrategy
	out.GPUNetworkAccelerationDetection = np.GPUNetworkAccelerationDetection
	out.GPUNetworkAccelerationLabelKey = np.GPUNetworkAccelerationLabelKey
	out.NodepoolStatus = np.DeepCopy()
	out.MinGuaranteedRuntime = np.MinGuaranteedRuntime

	out.TimeBasedFairShare = np.TimeBasedFairShare

	out.NetworkTopologyId = np.NetworkTopologyId
	out.NetworkTopologyName = np.NetworkTopologyName
	out.Conditions = np.Conditions

	out.AllowAdjustableWeights = np.AllowAdjustableWeights
	out.AllowRanks = np.AllowRanks
	out.AllowQuota = np.AllowQuota
	out.AllowCpuQuota = np.AllowCpuQuota

	out.NumaAwareScheduling = np.NumaAwareScheduling
}

func (np *Nodepool) GetSchedulingConfiguration() *api.SchedulingConfiguration {
	return &api.SchedulingConfiguration{
		GpuPodsPlacementStrategy: &api.GpuPodsPlacementStrategy{
			Node:   &np.GpuPlacementStrategy,
			Device: &np.GpuDevicePlacementStrategy,
		},
		CpuPodsPlacementStrategy: &api.CpuPodsPlacementStrategy{
			Node: &np.CpuPlacementStrategy,
		},
		// Deprecated mirror — kept on responses so older clients keep working.
		// Reflects node-level strategies; device-level lives only on the new field.
		PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
			Cpu: &np.CpuPlacementStrategy,
			Gpu: &np.GpuPlacementStrategy,
		},
		MinGuaranteedRuntime: np.MinGuaranteedRuntime,
		TimeBasedFairShare:   ptr.To(np.TimeBasedFairShare.ToApi()),
		NumaAwareScheduling:  np.NumaAwareScheduling,
	}
}

func (np *Nodepool) GetSchedulingCapabilities() api.SchedulingCapabilities {
	return api.SchedulingCapabilities{
		Quota:             np.AllowQuota,
		Ranks:             np.AllowRanks,
		AdjustableWeights: np.AllowAdjustableWeights,
		CpuQuota:          np.AllowCpuQuota,
	}
}

// ToMetadataApiModel converts a DAO NodepoolMinimal model to API NodepoolMinimal model
// This is used for the metadata endpoint which only returns essential fields
func (np *Nodepool) ToMetadataApiModel(gpuTypes []api.NodepoolGpuTypeDetails) api.NodepoolMinimal {
	metadata := api.NodepoolMinimal{
		Id:                             strconv.Itoa(int(np.ID)),
		Name:                           np.Name,
		ClusterId:                      np.ClusterId.String(),
		Phase:                          np.Phase,
		GpuNetworkAccelerationDetected: np.GPUNetworkAccelerationDetected,
		SchedulingCapabilities:         np.GetSchedulingCapabilities(),
		GpuTypes:                       gpuTypes,
	}
	return metadata
}

type TimeBasedFairShare struct {
	Enabled                 *bool                             `json:"enabled" gorm:""`
	HistoricalUsageWeight   *float32                          `json:"historicalUsageWeight,omitempty" gorm:""`
	DecayHalfLife           *string                           `json:"decayHalfLife,omitempty" gorm:"" validate:"regexp=^(0|[+]?((\\\\d+(\\\\.\\\\d+)?)(m|h|d))+$)"`
	WindowType              *api.TimeBasedFairShareWindowType `json:"windowType,omitempty" gorm:""`
	WindowDuration          *string                           `json:"windowDuration,omitempty" gorm:"" validate:"regexp=^(0|[+]?((\\\\d+(\\\\.\\\\d+)?)(m|h|d))+$)"`
	TumblingWindowStartTime *time.Time                        `json:"tumblingWindowStartTime,omitempty" gorm:""`
}

func (tbfs TimeBasedFairShare) ToApi() api.TimeBasedFairShare {
	return api.TimeBasedFairShare{
		Enabled:                 tbfs.Enabled,
		HistoricalUsageWeight:   tbfs.HistoricalUsageWeight,
		DecayHalfLife:           tbfs.DecayHalfLife,
		WindowType:              tbfs.WindowType,
		WindowDuration:          tbfs.WindowDuration,
		TumblingWindowStartTime: tbfs.TumblingWindowStartTime,
	}
}
