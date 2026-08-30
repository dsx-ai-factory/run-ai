import type { ITableColumn } from "@/models/table.model";
import { idleGpuDurationFormat, memoryFormat } from "@/utils/format.util";
import { percentFormat, tableNumberFormat } from "@/utils/table-format.util";

export const allNodeGpuColumnsMap: Record<string, ITableColumn> = {
  index: {
    name: "index",
    label: "Index",
    field: "index",
    sortable: true,
    align: "left",
    display: true,
  },
  usedMemory: {
    name: "usedMemory",
    label: "Used memory",
    field: "usedMemory",
    sortable: true,
    align: "left",
    display: true,
    format: memoryFormat,
  },
  utilization: {
    name: "utilization",
    label: "Compute utilization",
    field: "utilization",
    sortable: true,
    align: "left",
    display: true,
    format: tableNumberFormat("%"),
  },
  memoryUtilization: {
    name: "memoryUtilization",
    label: "Memory utilization",
    field: "usedMemory",
    sortable: true,
    align: "left",
    display: true,
    format: percentFormat("totalMemory"),
  },
  idleTime: {
    name: "idleTime",
    label: "Idle time",
    field: "idleTime",
    sortable: true,
    align: "left",
    display: true,
    format: idleGpuDurationFormat,
  },
  graphicsEngineActivity: {
    name: "graphics-engine-activity",
    label: "Graphics engine activity",
    field: "graphicsEngineActivity",
    sortable: true,
    align: "left",
    display: false,
    format: tableNumberFormat("%"),
  },
  smActivity: {
    name: "sm-activity",
    label: "SM activity",
    field: "smActivity",
    sortable: true,
    align: "left",
    display: false,
    format: tableNumberFormat("%"),
  },
  smOccupancy: {
    name: "sm-occupancy",
    label: "SM occupancy",
    field: "smOccupancy",
    sortable: true,
    align: "left",
    display: false,
    format: tableNumberFormat("%"),
  },
  tensorActivity: {
    name: "tensor-activity",
    label: "Tensor activity",
    field: "tensorActivity",
    sortable: true,
    align: "left",
    display: false,
    format: tableNumberFormat("%"),
  },
  fp64EngineActivity: {
    name: "fp64-engine-activity",
    label: "FP64 engine activity",
    field: "fp64EngineActivity",
    sortable: true,
    align: "left",
    display: false,
    format: tableNumberFormat("%"),
  },
  fp32EngineActivity: {
    name: "fp32-engine-activity",
    label: "FP32 engine activity",
    field: "fp32EngineActivity",
    sortable: true,
    align: "left",
    display: false,
    format: tableNumberFormat("%"),
  },
  fp16EngineActivity: {
    name: "fp16-engine-activity",
    label: "FP16 engine activity",
    field: "fp16EngineActivity",
    sortable: true,
    align: "left",
    display: false,
    format: tableNumberFormat("%"),
  },
  memoryBWUtilization: {
    name: "memory-bw-utilization",
    label: "Memory bandwidth utilization",
    field: "memoryBWUtilization",
    sortable: true,
    align: "left",
    display: false,
    format: tableNumberFormat("%"),
  },
  nvLinkBandwidthRead: {
    name: "nv-link-bandwidth-read",
    label: "NVLink received bandwidth",
    field: "nvLinkBandwidthRead",
    sortable: true,
    align: "left",
    display: false,
    format: (val: string) => {
      const formattedVal: string = memoryFormat(val);
      return formattedVal !== "-" ? formattedVal + "/s" : formattedVal;
    },
  },
  nvLinkBandwidthTransmit: {
    name: "nv-link-bandwidth-transmit",
    label: "NVLink transmitted bandwidth",
    field: "nvLinkBandwidthTransmit",
    sortable: true,
    align: "left",
    display: false,
    format: (val: string) => {
      const formattedVal: string = memoryFormat(val);
      return formattedVal !== "-" ? formattedVal + "/s" : formattedVal;
    },
  },
  pcieBandwidthRead: {
    name: "pcie-bandwidth-read",
    label: "PCIe received bandwidth",
    field: "pcieBandwidthRead",
    sortable: true,
    align: "left",
    display: false,
    format: (val: string) => {
      const formattedVal: string = memoryFormat(val);
      return formattedVal !== "-" ? formattedVal + "/s" : formattedVal;
    },
  },
  pcieBandwidthTransmit: {
    name: "pcie-bandwidth-transmit",
    label: "PCIe transmitted bandwidth",
    field: "pcieBandwidthTransmit",
    sortable: true,
    align: "left",
    display: false,
    format: (val: string) => {
      const formattedVal: string = memoryFormat(val);
      return formattedVal !== "-" ? formattedVal + "/s" : formattedVal;
    },
  },
};
export const nodeDefaultGpuColumns = [
  { ...allNodeGpuColumnsMap.index },
  { ...allNodeGpuColumnsMap.usedMemory },
  { ...allNodeGpuColumnsMap.utilization },
  { ...allNodeGpuColumnsMap.memoryUtilization },
  { ...allNodeGpuColumnsMap.idleTime },
] as Array<ITableColumn>;

export const nodeAdvancedGpuColumns = [
  { ...allNodeGpuColumnsMap.graphicsEngineActivity },
  { ...allNodeGpuColumnsMap.smActivity },
  { ...allNodeGpuColumnsMap.smOccupancy },
  { ...allNodeGpuColumnsMap.tensorActivity },
  { ...allNodeGpuColumnsMap.fp64EngineActivity },
  { ...allNodeGpuColumnsMap.fp32EngineActivity },
  { ...allNodeGpuColumnsMap.fp16EngineActivity },
  { ...allNodeGpuColumnsMap.memoryBWUtilization },
  { ...allNodeGpuColumnsMap.nvLinkBandwidthRead },
  { ...allNodeGpuColumnsMap.nvLinkBandwidthTransmit },
  { ...allNodeGpuColumnsMap.pcieBandwidthRead },
  { ...allNodeGpuColumnsMap.pcieBandwidthTransmit },
] as Array<ITableColumn>;
