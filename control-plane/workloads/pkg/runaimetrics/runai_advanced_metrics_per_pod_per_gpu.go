/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

// runai_gpu_gr_engine_active_per_pod_per_gpu

type RunaiGpuGREngineActivePerPodPerGpuFilters struct {
	PodId []string `mapstructure:"pod_uuid,omitempty"`
}

type RunaiGpuGREngineActivePerPodPerGpuResponseLabels struct {
	PodId bool `mapstructure:"pod_uuid,omitempty"`
	Gpu   bool `mapstructure:"gpu,omitempty"`
}

type RunaiGpuGREngineActivePerPodPerGpuMetric struct {
	Filters        RunaiGpuGREngineActivePerPodPerGpuFilters
	ResponseLabels RunaiGpuGREngineActivePerPodPerGpuResponseLabels
	Quantile       *float32
}

func (m RunaiGpuGREngineActivePerPodPerGpuMetric) GetName() string {
	return "runai_gpu_gr_engine_active_per_pod_per_gpu"
}

func (m RunaiGpuGREngineActivePerPodPerGpuMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuGREngineActivePerPodPerGpuMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuGREngineActivePerPodPerGpuMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuGREngineActivePerPodPerGpuMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuGREngineActivePerPodPerGpuMetric) IsQuantile() bool {
	return false
}

// runai_gpu_sm_active_per_pod_per_gpu

type RunaiGpuSmActivePerPodPerGpuFilters struct {
	PodId []string `mapstructure:"pod_uuid,omitempty"`
}

type RunaiGpuSmActivePerPodPerGpuResponseLabels struct {
	PodId bool `mapstructure:"pod_uuid,omitempty"`
	Gpu   bool `mapstructure:"gpu,omitempty"`
}

type RunaiGpuSmActivePerPodPerGpuMetric struct {
	Filters        RunaiGpuSmActivePerPodPerGpuFilters
	ResponseLabels RunaiGpuSmActivePerPodPerGpuResponseLabels
	Quantile       *float32
}

func (m RunaiGpuSmActivePerPodPerGpuMetric) GetName() string {
	return "runai_gpu_sm_active_per_pod_per_gpu"
}

func (m RunaiGpuSmActivePerPodPerGpuMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuSmActivePerPodPerGpuMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuSmActivePerPodPerGpuMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuSmActivePerPodPerGpuMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuSmActivePerPodPerGpuMetric) IsQuantile() bool {
	return false
}

// runai_gpu_sm_occupancy_per_pod_per_gpu

type RunaiGpuSmOccupancyPerPodPerGpuFilters struct {
	PodId []string `mapstructure:"pod_uuid,omitempty"`
}

type RunaiGpuSmOccupancyPerPodPerGpuResponseLabels struct {
	PodId bool `mapstructure:"pod_uuid,omitempty"`
	Gpu   bool `mapstructure:"gpu,omitempty"`
}

type RunaiGpuSmOccupancyPerPodPerGpuMetric struct {
	Filters        RunaiGpuSmOccupancyPerPodPerGpuFilters
	ResponseLabels RunaiGpuSmOccupancyPerPodPerGpuResponseLabels
	Quantile       *float32
}

func (m RunaiGpuSmOccupancyPerPodPerGpuMetric) GetName() string {
	return "runai_gpu_sm_occupancy_per_pod_per_gpu"
}

func (m RunaiGpuSmOccupancyPerPodPerGpuMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuSmOccupancyPerPodPerGpuMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuSmOccupancyPerPodPerGpuMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuSmOccupancyPerPodPerGpuMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuSmOccupancyPerPodPerGpuMetric) IsQuantile() bool {
	return false
}

// runai_gpu_pipe_tensor_active_per_pod_per_gpu

type RunaiGpuPipeTensorActivePerPodPerGpuFilters struct {
	PodId []string `mapstructure:"pod_uuid,omitempty"`
}

type RunaiGpuPipeTensorActivePerPodPerGpuResponseLabels struct {
	PodId bool `mapstructure:"pod_uuid,omitempty"`
	Gpu   bool `mapstructure:"gpu,omitempty"`
}

type RunaiGpuPipeTensorActivePerPodPerGpuMetric struct {
	Filters        RunaiGpuPipeTensorActivePerPodPerGpuFilters
	ResponseLabels RunaiGpuPipeTensorActivePerPodPerGpuResponseLabels
	Quantile       *float32
}

func (m RunaiGpuPipeTensorActivePerPodPerGpuMetric) GetName() string {
	return "runai_gpu_pipe_tensor_active_per_pod_per_gpu"
}

func (m RunaiGpuPipeTensorActivePerPodPerGpuMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuPipeTensorActivePerPodPerGpuMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuPipeTensorActivePerPodPerGpuMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuPipeTensorActivePerPodPerGpuMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuPipeTensorActivePerPodPerGpuMetric) IsQuantile() bool {
	return false
}

// runai_gpu_pipe_fp64_active_per_pod_per_gpu

type RunaiGpuPipeFp64ActivePerPodPerGpuFilters struct {
	PodId []string `mapstructure:"pod_uuid,omitempty"`
}

type RunaiGpuPipeFp64ActivePerPodPerGpuResponseLabels struct {
	PodId bool `mapstructure:"pod_uuid,omitempty"`
	Gpu   bool `mapstructure:"gpu,omitempty"`
}

type RunaiGpuPipeFp64ActivePerPodPerGpuMetric struct {
	Filters        RunaiGpuPipeFp64ActivePerPodPerGpuFilters
	ResponseLabels RunaiGpuPipeFp64ActivePerPodPerGpuResponseLabels
	Quantile       *float32
}

func (m RunaiGpuPipeFp64ActivePerPodPerGpuMetric) GetName() string {
	return "runai_gpu_pipe_fp64_active_per_pod_per_gpu"
}

func (m RunaiGpuPipeFp64ActivePerPodPerGpuMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuPipeFp64ActivePerPodPerGpuMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuPipeFp64ActivePerPodPerGpuMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuPipeFp64ActivePerPodPerGpuMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuPipeFp64ActivePerPodPerGpuMetric) IsQuantile() bool {
	return false
}

// runai_gpu_pipe_fp32_active_per_pod_per_gpu

type RunaiGpuPipeFp32ActivePerPodPerGpuFilters struct {
	PodId []string `mapstructure:"pod_uuid,omitempty"`
}

type RunaiGpuPipeFp32ActivePerPodPerGpuResponseLabels struct {
	PodId bool `mapstructure:"pod_uuid,omitempty"`
	Gpu   bool `mapstructure:"gpu,omitempty"`
}

type RunaiGpuPipeFp32ActivePerPodPerGpuMetric struct {
	Filters        RunaiGpuPipeFp32ActivePerPodPerGpuFilters
	ResponseLabels RunaiGpuPipeFp32ActivePerPodPerGpuResponseLabels
	Quantile       *float32
}

func (m RunaiGpuPipeFp32ActivePerPodPerGpuMetric) GetName() string {
	return "runai_gpu_pipe_fp32_active_per_pod_per_gpu"
}

func (m RunaiGpuPipeFp32ActivePerPodPerGpuMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuPipeFp32ActivePerPodPerGpuMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuPipeFp32ActivePerPodPerGpuMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuPipeFp32ActivePerPodPerGpuMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuPipeFp32ActivePerPodPerGpuMetric) IsQuantile() bool {
	return false
}

// runai_gpu_pipe_fp16_active_per_pod_per_gpu

type RunaiGpuPipeFp16ActivePerPodPerGpuFilters struct {
	PodId []string `mapstructure:"pod_uuid,omitempty"`
}

type RunaiGpuPipeFp16ActivePerPodPerGpuResponseLabels struct {
	PodId bool `mapstructure:"pod_uuid,omitempty"`
	Gpu   bool `mapstructure:"gpu,omitempty"`
}

type RunaiGpuPipeFp16ActivePerPodPerGpuMetric struct {
	Filters        RunaiGpuPipeFp16ActivePerPodPerGpuFilters
	ResponseLabels RunaiGpuPipeFp16ActivePerPodPerGpuResponseLabels
	Quantile       *float32
}

func (m RunaiGpuPipeFp16ActivePerPodPerGpuMetric) GetName() string {
	return "runai_gpu_pipe_fp16_active_per_pod_per_gpu"
}

func (m RunaiGpuPipeFp16ActivePerPodPerGpuMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuPipeFp16ActivePerPodPerGpuMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuPipeFp16ActivePerPodPerGpuMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuPipeFp16ActivePerPodPerGpuMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuPipeFp16ActivePerPodPerGpuMetric) IsQuantile() bool {
	return false
}

// runai_gpu_dram_active_per_pod_per_gpu

type RunaiGpuDramActivePerPodPerGpuFilters struct {
	PodId []string `mapstructure:"pod_uuid,omitempty"`
}

type RunaiGpuDramActivePerPodPerGpuResponseLabels struct {
	PodId bool `mapstructure:"pod_uuid,omitempty"`
	Gpu   bool `mapstructure:"gpu,omitempty"`
}

type RunaiGpuDramActivePerPodPerGpuMetric struct {
	Filters        RunaiGpuDramActivePerPodPerGpuFilters
	ResponseLabels RunaiGpuDramActivePerPodPerGpuResponseLabels
	Quantile       *float32
}

func (m RunaiGpuDramActivePerPodPerGpuMetric) GetName() string {
	return "runai_gpu_dram_active_per_pod_per_gpu"
}

func (m RunaiGpuDramActivePerPodPerGpuMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuDramActivePerPodPerGpuMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuDramActivePerPodPerGpuMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuDramActivePerPodPerGpuMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuDramActivePerPodPerGpuMetric) IsQuantile() bool {
	return false
}

// runai_gpu_nvlink_tx_bytes_per_pod_per_gpu

type RunaiGpuNvlinkTxBytesPerPodPerGpuFilters struct {
	PodId []string `mapstructure:"pod_uuid,omitempty"`
}

type RunaiGpuNvlinkTxBytesPerPodPerGpuResponseLabels struct {
	PodId bool `mapstructure:"pod_uuid,omitempty"`
	Gpu   bool `mapstructure:"gpu,omitempty"`
}

type RunaiGpuNvlinkTxBytesPerPodPerGpuMetric struct {
	Filters        RunaiGpuNvlinkTxBytesPerPodPerGpuFilters
	ResponseLabels RunaiGpuNvlinkTxBytesPerPodPerGpuResponseLabels
	Quantile       *float32
}

func (m RunaiGpuNvlinkTxBytesPerPodPerGpuMetric) GetName() string {
	return "runai_gpu_nvlink_tx_bytes_per_pod_per_gpu"
}

func (m RunaiGpuNvlinkTxBytesPerPodPerGpuMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuNvlinkTxBytesPerPodPerGpuMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuNvlinkTxBytesPerPodPerGpuMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuNvlinkTxBytesPerPodPerGpuMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuNvlinkTxBytesPerPodPerGpuMetric) IsQuantile() bool {
	return false
}

// runai_gpu_nvlink_rx_bytes_per_pod_per_gpu

type RunaiGpuNvlinkRxBytesPerPodPerGpuFilters struct {
	PodId []string `mapstructure:"pod_uuid,omitempty"`
}

type RunaiGpuNvlinkRxBytesPerPodPerGpuResponseLabels struct {
	PodId bool `mapstructure:"pod_uuid,omitempty"`
	Gpu   bool `mapstructure:"gpu,omitempty"`
}

type RunaiGpuNvlinkRxBytesPerPodPerGpuMetric struct {
	Filters        RunaiGpuNvlinkRxBytesPerPodPerGpuFilters
	ResponseLabels RunaiGpuNvlinkRxBytesPerPodPerGpuResponseLabels
	Quantile       *float32
}

func (m RunaiGpuNvlinkRxBytesPerPodPerGpuMetric) GetName() string {
	return "runai_gpu_nvlink_rx_bytes_per_pod_per_gpu"
}

func (m RunaiGpuNvlinkRxBytesPerPodPerGpuMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuNvlinkRxBytesPerPodPerGpuMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuNvlinkRxBytesPerPodPerGpuMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuNvlinkRxBytesPerPodPerGpuMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuNvlinkRxBytesPerPodPerGpuMetric) IsQuantile() bool {
	return false
}

// runai_gpu_pcie_tx_bytes_per_pod_per_gpu

type RunaiGpuPcieTxBytesPerPodPerGpuFilters struct {
	PodId []string `mapstructure:"pod_uuid,omitempty"`
}

type RunaiGpuPcieTxBytesPerPodPerGpuResponseLabels struct {
	PodId bool `mapstructure:"pod_uuid,omitempty"`
	Gpu   bool `mapstructure:"gpu,omitempty"`
}

type RunaiGpuPcieTxBytesPerPodPerGpuMetric struct {
	Filters        RunaiGpuPcieTxBytesPerPodPerGpuFilters
	ResponseLabels RunaiGpuPcieTxBytesPerPodPerGpuResponseLabels
	Quantile       *float32
}

func (m RunaiGpuPcieTxBytesPerPodPerGpuMetric) GetName() string {
	return "runai_gpu_pcie_tx_bytes_per_pod_per_gpu"
}

func (m RunaiGpuPcieTxBytesPerPodPerGpuMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuPcieTxBytesPerPodPerGpuMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuPcieTxBytesPerPodPerGpuMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuPcieTxBytesPerPodPerGpuMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuPcieTxBytesPerPodPerGpuMetric) IsQuantile() bool {
	return false
}

// runai_gpu_pcie_rx_bytes_per_pod_per_gpu

type RunaiGpuPcieRxBytesPerPodPerGpuFilters struct {
	PodId []string `mapstructure:"pod_uuid,omitempty"`
}

type RunaiGpuPcieRxBytesPerPodPerGpuResponseLabels struct {
	PodId bool `mapstructure:"pod_uuid,omitempty"`
	Gpu   bool `mapstructure:"gpu,omitempty"`
}

type RunaiGpuPcieRxBytesPerPodPerGpuMetric struct {
	Filters        RunaiGpuPcieRxBytesPerPodPerGpuFilters
	ResponseLabels RunaiGpuPcieRxBytesPerPodPerGpuResponseLabels
	Quantile       *float32
}

func (m RunaiGpuPcieRxBytesPerPodPerGpuMetric) GetName() string {
	return "runai_gpu_pcie_rx_bytes_per_pod_per_gpu"
}

func (m RunaiGpuPcieRxBytesPerPodPerGpuMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuPcieRxBytesPerPodPerGpuMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuPcieRxBytesPerPodPerGpuMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuPcieRxBytesPerPodPerGpuMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuPcieRxBytesPerPodPerGpuMetric) IsQuantile() bool {
	return false
}
