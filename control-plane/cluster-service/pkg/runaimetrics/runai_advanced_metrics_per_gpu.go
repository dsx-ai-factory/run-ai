package runaimetrics

import (
	"github.com/google/uuid"
)

// runai_gpu_gr_engine_active_per_gpu

type RunaiGpuGREngineActivePerGpuFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
	Node      []string    `mapstructure:"node,omitempty"`
}

type RunaiGpuGREngineActivePerGpuResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
	Node      bool `mapstructure:"node,omitempty"`
	Gpu       bool `mapstructure:"gpu,omitempty"`
}

type RunaiGpuGREngineActivePerGpuMetric struct {
	Filters        RunaiGpuGREngineActivePerGpuFilters
	ResponseLabels RunaiGpuGREngineActivePerGpuResponseLabels
	Quantile       *float32
}

func (m RunaiGpuGREngineActivePerGpuMetric) GetName() string {
	return "runai_gpu_gr_engine_active_per_gpu"
}

func (m RunaiGpuGREngineActivePerGpuMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuGREngineActivePerGpuMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuGREngineActivePerGpuMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuGREngineActivePerGpuMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuGREngineActivePerGpuMetric) IsQuantile() bool {
	return false
}

// runai_gpu_sm_active_per_gpu

type RunaiGpuSmActivePerGpuFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
	Node      []string    `mapstructure:"node,omitempty"`
}

type RunaiGpuSmActivePerGpuResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
	Node      bool `mapstructure:"node,omitempty"`
	Gpu       bool `mapstructure:"gpu,omitempty"`
}

type RunaiGpuSmActivePerGpuMetric struct {
	Filters        RunaiGpuSmActivePerGpuFilters
	ResponseLabels RunaiGpuSmActivePerGpuResponseLabels
	Quantile       *float32
}

func (m RunaiGpuSmActivePerGpuMetric) GetName() string {
	return "runai_gpu_sm_active_per_gpu"
}

func (m RunaiGpuSmActivePerGpuMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuSmActivePerGpuMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuSmActivePerGpuMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuSmActivePerGpuMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuSmActivePerGpuMetric) IsQuantile() bool {
	return false
}

// runai_gpu_sm_occupancy_per_gpu

type RunaiGpuSmOccupancyPerGpuFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
	Node      []string    `mapstructure:"node,omitempty"`
}

type RunaiGpuSmOccupancyPerGpuResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
	Node      bool `mapstructure:"node,omitempty"`
	Gpu       bool `mapstructure:"gpu,omitempty"`
}

type RunaiGpuSmOccupancyPerGpuMetric struct {
	Filters        RunaiGpuSmOccupancyPerGpuFilters
	ResponseLabels RunaiGpuSmOccupancyPerGpuResponseLabels
	Quantile       *float32
}

func (m RunaiGpuSmOccupancyPerGpuMetric) GetName() string {
	return "runai_gpu_sm_occupancy_per_gpu"
}

func (m RunaiGpuSmOccupancyPerGpuMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuSmOccupancyPerGpuMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuSmOccupancyPerGpuMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuSmOccupancyPerGpuMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuSmOccupancyPerGpuMetric) IsQuantile() bool {
	return false
}

// runai_gpu_pipe_tensor_active_per_gpu

type RunaiGpuPipeTensorActivePerGpuFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
	Node      []string    `mapstructure:"node,omitempty"`
}

type RunaiGpuPipeTensorActivePerGpuResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
	Node      bool `mapstructure:"node,omitempty"`
	Gpu       bool `mapstructure:"gpu,omitempty"`
}

type RunaiGpuPipeTensorActivePerGpuMetric struct {
	Filters        RunaiGpuPipeTensorActivePerGpuFilters
	ResponseLabels RunaiGpuPipeTensorActivePerGpuResponseLabels
	Quantile       *float32
}

func (m RunaiGpuPipeTensorActivePerGpuMetric) GetName() string {
	return "runai_gpu_pipe_tensor_active_per_gpu"
}

func (m RunaiGpuPipeTensorActivePerGpuMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuPipeTensorActivePerGpuMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuPipeTensorActivePerGpuMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuPipeTensorActivePerGpuMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuPipeTensorActivePerGpuMetric) IsQuantile() bool {
	return false
}

// runai_gpu_pipe_fp64_active_per_gpu

type RunaiGpuPipeFp64ActivePerGpuFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
	Node      []string    `mapstructure:"node,omitempty"`
}

type RunaiGpuPipeFp64ActivePerGpuResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
	Node      bool `mapstructure:"node,omitempty"`
	Gpu       bool `mapstructure:"gpu,omitempty"`
}

type RunaiGpuPipeFp64ActivePerGpuMetric struct {
	Filters        RunaiGpuPipeFp64ActivePerGpuFilters
	ResponseLabels RunaiGpuPipeFp64ActivePerGpuResponseLabels
	Quantile       *float32
}

func (m RunaiGpuPipeFp64ActivePerGpuMetric) GetName() string {
	return "runai_gpu_pipe_fp64_active_per_gpu"
}

func (m RunaiGpuPipeFp64ActivePerGpuMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuPipeFp64ActivePerGpuMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuPipeFp64ActivePerGpuMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuPipeFp64ActivePerGpuMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuPipeFp64ActivePerGpuMetric) IsQuantile() bool {
	return false
}

// runai_gpu_pipe_fp32_active_per_gpu

type RunaiGpuPipeFp32ActivePerGpuFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
	Node      []string    `mapstructure:"node,omitempty"`
}

type RunaiGpuPipeFp32ActivePerGpuResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
	Node      bool `mapstructure:"node,omitempty"`
	Gpu       bool `mapstructure:"gpu,omitempty"`
}

type RunaiGpuPipeFp32ActivePerGpuMetric struct {
	Filters        RunaiGpuPipeFp32ActivePerGpuFilters
	ResponseLabels RunaiGpuPipeFp32ActivePerGpuResponseLabels
	Quantile       *float32
}

func (m RunaiGpuPipeFp32ActivePerGpuMetric) GetName() string {
	return "runai_gpu_pipe_fp32_active_per_gpu"
}

func (m RunaiGpuPipeFp32ActivePerGpuMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuPipeFp32ActivePerGpuMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuPipeFp32ActivePerGpuMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuPipeFp32ActivePerGpuMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuPipeFp32ActivePerGpuMetric) IsQuantile() bool {
	return false
}

// runai_gpu_pipe_fp16_active_per_gpu

type RunaiGpuPipeFp16ActivePerGpuFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
	Node      []string    `mapstructure:"node,omitempty"`
}

type RunaiGpuPipeFp16ActivePerGpuResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
	Node      bool `mapstructure:"node,omitempty"`
	Gpu       bool `mapstructure:"gpu,omitempty"`
}

type RunaiGpuPipeFp16ActivePerGpuMetric struct {
	Filters        RunaiGpuPipeFp16ActivePerGpuFilters
	ResponseLabels RunaiGpuPipeFp16ActivePerGpuResponseLabels
	Quantile       *float32
}

func (m RunaiGpuPipeFp16ActivePerGpuMetric) GetName() string {
	return "runai_gpu_pipe_fp16_active_per_gpu"
}

func (m RunaiGpuPipeFp16ActivePerGpuMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuPipeFp16ActivePerGpuMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuPipeFp16ActivePerGpuMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuPipeFp16ActivePerGpuMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuPipeFp16ActivePerGpuMetric) IsQuantile() bool {
	return false
}

// runai_gpu_dram_active_per_gpu

type RunaiGpuDramActivePerGpuFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
	Node      []string    `mapstructure:"node,omitempty"`
}

type RunaiGpuDramActivePerGpuResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
	Node      bool `mapstructure:"node,omitempty"`
	Gpu       bool `mapstructure:"gpu,omitempty"`
}

type RunaiGpuDramActivePerGpuMetric struct {
	Filters        RunaiGpuDramActivePerGpuFilters
	ResponseLabels RunaiGpuDramActivePerGpuResponseLabels
	Quantile       *float32
}

func (m RunaiGpuDramActivePerGpuMetric) GetName() string {
	return "runai_gpu_dram_active_per_gpu"
}

func (m RunaiGpuDramActivePerGpuMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuDramActivePerGpuMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuDramActivePerGpuMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuDramActivePerGpuMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuDramActivePerGpuMetric) IsQuantile() bool {
	return false
}

// runai_gpu_nvlink_tx_bytes_per_gpu

type RunaiGpuNvlinkTxBytesPerGpuFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
	Node      []string    `mapstructure:"node,omitempty"`
}

type RunaiGpuNvlinkTxBytesPerGpuResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
	Node      bool `mapstructure:"node,omitempty"`
	Gpu       bool `mapstructure:"gpu,omitempty"`
}

type RunaiGpuNvlinkTxBytesPerGpuMetric struct {
	Filters        RunaiGpuNvlinkTxBytesPerGpuFilters
	ResponseLabels RunaiGpuNvlinkTxBytesPerGpuResponseLabels
	Quantile       *float32
}

func (m RunaiGpuNvlinkTxBytesPerGpuMetric) GetName() string {
	return "runai_gpu_nvlink_tx_bytes_per_gpu"
}

func (m RunaiGpuNvlinkTxBytesPerGpuMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuNvlinkTxBytesPerGpuMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuNvlinkTxBytesPerGpuMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuNvlinkTxBytesPerGpuMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuNvlinkTxBytesPerGpuMetric) IsQuantile() bool {
	return false
}

// runai_gpu_nvlink_rx_bytes_per_gpu

type RunaiGpuNvlinkRxBytesPerGpuFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
	Node      []string    `mapstructure:"node,omitempty"`
}

type RunaiGpuNvlinkRxBytesPerGpuResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
	Node      bool `mapstructure:"node,omitempty"`
	Gpu       bool `mapstructure:"gpu,omitempty"`
}

type RunaiGpuNvlinkRxBytesPerGpuMetric struct {
	Filters        RunaiGpuNvlinkRxBytesPerGpuFilters
	ResponseLabels RunaiGpuNvlinkRxBytesPerGpuResponseLabels
	Quantile       *float32
}

func (m RunaiGpuNvlinkRxBytesPerGpuMetric) GetName() string {
	return "runai_gpu_nvlink_rx_bytes_per_gpu"
}

func (m RunaiGpuNvlinkRxBytesPerGpuMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuNvlinkRxBytesPerGpuMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuNvlinkRxBytesPerGpuMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuNvlinkRxBytesPerGpuMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuNvlinkRxBytesPerGpuMetric) IsQuantile() bool {
	return false
}

// runai_gpu_pcie_tx_bytes_per_gpu

type RunaiGpuPcieTxBytesPerGpuFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
	Node      []string    `mapstructure:"node,omitempty"`
}

type RunaiGpuPcieTxBytesPerGpuResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
	Node      bool `mapstructure:"node,omitempty"`
	Gpu       bool `mapstructure:"gpu,omitempty"`
}

type RunaiGpuPcieTxBytesPerGpuMetric struct {
	Filters        RunaiGpuPcieTxBytesPerGpuFilters
	ResponseLabels RunaiGpuPcieTxBytesPerGpuResponseLabels
	Quantile       *float32
}

func (m RunaiGpuPcieTxBytesPerGpuMetric) GetName() string {
	return "runai_gpu_pcie_tx_bytes_per_gpu"
}

func (m RunaiGpuPcieTxBytesPerGpuMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuPcieTxBytesPerGpuMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuPcieTxBytesPerGpuMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuPcieTxBytesPerGpuMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuPcieTxBytesPerGpuMetric) IsQuantile() bool {
	return false
}

// runai_gpu_pcie_rx_bytes_per_gpu

type RunaiGpuPcieRxBytesPerGpuFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
	Node      []string    `mapstructure:"node,omitempty"`
}

type RunaiGpuPcieRxBytesPerGpuResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
	Node      bool `mapstructure:"node,omitempty"`
	Gpu       bool `mapstructure:"gpu,omitempty"`
}

type RunaiGpuPcieRxBytesPerGpuMetric struct {
	Filters        RunaiGpuPcieRxBytesPerGpuFilters
	ResponseLabels RunaiGpuPcieRxBytesPerGpuResponseLabels
	Quantile       *float32
}

func (m RunaiGpuPcieRxBytesPerGpuMetric) GetName() string {
	return "runai_gpu_pcie_rx_bytes_per_gpu"
}

func (m RunaiGpuPcieRxBytesPerGpuMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuPcieRxBytesPerGpuMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuPcieRxBytesPerGpuMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuPcieRxBytesPerGpuMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuPcieRxBytesPerGpuMetric) IsQuantile() bool {
	return false
}
