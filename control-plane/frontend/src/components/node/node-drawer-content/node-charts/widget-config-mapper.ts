/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { EIntervalLabels } from "@/models/interval.model";
import { NodeMetricType } from "@/swagger-models/cluster-service-client";
import {
  chartColorPaletteArray,
  type TMetricType,
  type IChartWidgetOptions,
  type IGenericMetricWidgetConfig,
} from "@/models/chart.model";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { formatK8sResourceMemory, formatUtil } from "@/utils/format.util/format.util";
import { widgetUtil } from "@/utils/widget.util";
import { DCGM_NVLINK_BANDWIDTH_TOTAL_BYTES_CHART_ID } from "@/components/dashboard-v2/widgets/chart-widgets/dcgm-nvlink-bandwidth-total-bytes-widget";
import { MIN_CLUSTER_VERSION_FOR_DCGM_MNNVL_METRICS } from "@/common/version.constant";

const sharedGetData = (
  metricTypes: Array<TMetricType>,
  start: string,
  end: string,
  numberOfSamples: number,
  args: string[],
) => clusterService.getNodeMetrics(args[0], metricTypes as NodeMetricType[], start, end, numberOfSamples);

const NODE_CHART_X_AXIS_INTERVAL_MS = 10 * 60 * 1000;
const NODE_CHART_MINIMUM_ZOOM_SAMPLE_COUNT = 5;
const NODE_CHART_GRID = { left: 50, right: 20, top: 10, bottom: 85, containLabel: false };

export enum GenericWidgets {
  GpuUtilizationPerNodePerGpu = "GpuUtilizationPerNodePerGpu",
  GpuMemoryUtilizationPerNode = "GpuMemoryUtilizationPerNodePerGpu",
  CpuUtilizationPerNode = "CpuUtilizationPerNode",
  CpuMemoryUtilizationPerNode = "CpuMemoryUtilizationPerNode",
  CpuMemoryUsageBytesPerNode = "CpuMemoryUsageBytesPerNode",
  DcgmFiDevNvlinkBandwidthTotal = "DcgmFiDevNvlinkBandwidthTotal",
}

export const widgetConfigMapper: Record<GenericWidgets, IGenericMetricWidgetConfig> = {
  [GenericWidgets.GpuUtilizationPerNodePerGpu]: {
    metricTypes: [NodeMetricType.GpuUtilization, NodeMetricType.GpuUtilizationPerGpu],
    averageMetricType: NodeMetricType.GpuUtilization,
    chartTitle: "GPU utilization",
    tooltipSuffix: "%",
    intervalLabel: EIntervalLabels.GpuUtilizationPerNodeWidget,
    chartId: "node-gpu-utilization-chart",
    isPercentage: true,
    getData: sharedGetData,
    chartOptions: {
      type: "line",
      height: 250,
      grid: NODE_CHART_GRID,
      adaptiveLegendLayout: true,
      minimumZoomSampleCount: NODE_CHART_MINIMUM_ZOOM_SAMPLE_COUNT,
      xAxisIntervalMs: NODE_CHART_X_AXIS_INTERVAL_MS,
      sharedTooltip: true,
      showSharedCrosshair: true,
    },
    measurementParser: widgetUtil.gpuUtilizationPerNodeParser,
  },
  [GenericWidgets.GpuMemoryUtilizationPerNode]: {
    metricTypes: [NodeMetricType.GpuMemoryUtilization, NodeMetricType.GpuMemoryUtilizationPerGpu],
    averageMetricType: NodeMetricType.GpuMemoryUtilization,
    chartTitle: "GPU memory utilization",
    tooltipSuffix: "%",
    intervalLabel: EIntervalLabels.GpuMemoryUtilizationPerNodeWidget,
    chartId: "node-gpu-memory-utilization-chart",
    isPercentage: true,
    getData: sharedGetData,
    chartOptions: {
      type: "line",
      height: 250,
      grid: NODE_CHART_GRID,
      adaptiveLegendLayout: true,
      minimumZoomSampleCount: NODE_CHART_MINIMUM_ZOOM_SAMPLE_COUNT,
      xAxisIntervalMs: NODE_CHART_X_AXIS_INTERVAL_MS,
      sharedTooltip: true,
      showSharedCrosshair: true,
    },
    measurementParser: widgetUtil.gpuUtilizationPerNodeParser,
  },
  [GenericWidgets.CpuUtilizationPerNode]: {
    metricTypes: [NodeMetricType.CpuUtilization],
    chartTitle: "CPU utilization",
    tooltipSuffix: "%",
    intervalLabel: EIntervalLabels.GpuUtilizationPerNodeWidget,
    chartId: "node-cpu-utilization-chart",
    keyLabelMap: {
      [NodeMetricType.CpuUtilization]: "CPU utilization",
    },
    isPercentage: true,
    getData: sharedGetData,
    chartOptions: {
      type: "line",
      height: 250,
      grid: NODE_CHART_GRID,
      adaptiveLegendLayout: true,
      minimumZoomSampleCount: NODE_CHART_MINIMUM_ZOOM_SAMPLE_COUNT,
      xAxisIntervalMs: NODE_CHART_X_AXIS_INTERVAL_MS,
      sharedTooltip: true,
      showSharedCrosshair: true,
      colors: [chartColorPaletteArray[1]],
    },
  },
  [GenericWidgets.CpuMemoryUtilizationPerNode]: {
    metricTypes: [NodeMetricType.CpuMemoryUtilization],
    chartTitle: "CPU memory utilization",
    tooltipSuffix: "%",
    intervalLabel: EIntervalLabels.GpuUtilizationPerNodeWidget,
    chartId: "node-cpu-memory-utilization-chart",
    keyLabelMap: {
      [NodeMetricType.CpuMemoryUtilization]: "CPU memory utilization",
    },
    getData: sharedGetData,
    chartOptions: {
      type: "line",
      height: 250,
      grid: NODE_CHART_GRID,
      adaptiveLegendLayout: true,
      minimumZoomSampleCount: NODE_CHART_MINIMUM_ZOOM_SAMPLE_COUNT,
      xAxisIntervalMs: NODE_CHART_X_AXIS_INTERVAL_MS,
      sharedTooltip: true,
      showSharedCrosshair: true,
      colors: [chartColorPaletteArray[2]],
    },
  },
  [GenericWidgets.CpuMemoryUsageBytesPerNode]: {
    metricTypes: [NodeMetricType.CpuMemoryUsageBytes],
    chartTitle: "CPU memory usage",
    intervalLabel: EIntervalLabels.GpuUtilizationPerNodeWidget,
    chartId: "node-cpu-memory-usage-chart",
    keyLabelMap: {
      [NodeMetricType.CpuMemoryUsageBytes]: "CPU memory usage",
    },
    getData: sharedGetData,
    chartOptions: {
      type: "line",
      height: 250,
      grid: NODE_CHART_GRID,
      adaptiveLegendLayout: true,
      minimumZoomSampleCount: NODE_CHART_MINIMUM_ZOOM_SAMPLE_COUNT,
      xAxisIntervalMs: NODE_CHART_X_AXIS_INTERVAL_MS,
      showSharedCrosshair: true,
      sharedTooltip: true,
      yAxisFormatFunction: formatK8sResourceMemory,
      colors: [chartColorPaletteArray[3]],
      highchartsLikeYAxis: true,
    },
  },
  [GenericWidgets.DcgmFiDevNvlinkBandwidthTotal]: {
    metricTypes: [NodeMetricType.NvlinkBandwidthTotal],
    chartTitle: "NVLink bandwidth total",
    intervalLabel: EIntervalLabels.DcgmNvlinkBandwidthTotalBytesWidget,
    chartId: DCGM_NVLINK_BANDWIDTH_TOTAL_BYTES_CHART_ID,
    keyLabelMap: {
      [NodeMetricType.NvlinkBandwidthTotal]: "NVLink bandwidth total",
    },
    getData: sharedGetData,
    chartOptions: {
      type: "line",
      height: 250,
      grid: NODE_CHART_GRID,
      adaptiveLegendLayout: true,
      minimumZoomSampleCount: NODE_CHART_MINIMUM_ZOOM_SAMPLE_COUNT,
      xAxisIntervalMs: NODE_CHART_X_AXIS_INTERVAL_MS,
      showSharedCrosshair: true,
      sharedTooltip: true,
      yAxisFormatFunction: formatK8sResourceMemory,
      colors: [chartColorPaletteArray[3]],
      highchartsLikeYAxis: true,
    },
    minSupportedClusterVersion: MIN_CLUSTER_VERSION_FOR_DCGM_MNNVL_METRICS,
  },
};

export enum AdvancedWidgets {
  GpuMemoryBandwidthUtilizationPerGpu = "GpuMemoryBandwidthUtilizationPerGpu",
  GpuNvlinkTransmittedBandwidthPerGpu = "GpuNvlinkTransmittedBandwidthPerGpu",
  GpuNvlinkReceivedBandwidthPerGpu = "GpuNvlinkReceivedBandwidthPerGpu",
  GpuPcieTransmittedBandwidthPerGpu = "GpuPcieTransmittedBandwidthPerGpu",
  GpuPcieReceivedBandwidthPerGpu = "GpuPcieReceivedBandwidthPerGpu",
  GpuOomkillBurstCountPerGpu = "GpuOomkillBurstCountPerGpu",
  GpuOomkillIdleCountPerGpu = "GpuOomkillIdleCountPerGpu",
  GpuOomkillSwapOutOfRamCountPerGpu = "GpuOomkillSwapOutOfRamCountPerGpu",
  GpuGraphicsEngineActivityPerGpu = "GpuGraphicsEngineActivityPerGpu",
  GpuSmActivityPerGpu = "GpuSmActivityPerGpu",
  GpuSmOccupancyPerGpu = "GpuSmOccupancyPerGpu",
  GpuTensorActivityPerGpu = "GpuTensorActivityPerGpu",
  GpuFp64EngineActivityPerGpu = "GpuFp64EngineActivityPerGpu",
  GpuFp32EngineActivityPerGpu = "GpuFp32EngineActivityPerGpu",
  GpuFp16EngineActivityPerGpu = "GpuFp16EngineActivityPerGpu",
}

const sharedChartOptions: IChartWidgetOptions = {
  type: "line",
  height: 250,
  grid: NODE_CHART_GRID,
  adaptiveLegendLayout: true,
  minimumZoomSampleCount: NODE_CHART_MINIMUM_ZOOM_SAMPLE_COUNT,
  xAxisIntervalMs: NODE_CHART_X_AXIS_INTERVAL_MS,
  sharedTooltip: true,
  showSharedCrosshair: true,
};

const formatGBs = (val: string | number) => formatUtil.convertBytesToGB(val) + "GB/s";

export const advancedWidgetConfigMapper: Record<AdvancedWidgets, IGenericMetricWidgetConfig> = {
  [AdvancedWidgets.GpuMemoryBandwidthUtilizationPerGpu]: {
    metricTypes: [NodeMetricType.GpuMemoryBandwidthUtilizationPerGpu],
    chartTitle: "Memory bandwidth utilization",
    tooltipSuffix: "%",
    intervalLabel: EIntervalLabels.GpuUtilizationPerNodeWidget,
    chartId: "node-gpu-memory-bandwidth-utilization-chart",
    isPercentage: true,
    getData: sharedGetData,
    measurementParser: widgetUtil.gpuUtilizationPerNodeParser,
    chartOptions: {
      ...sharedChartOptions,
    },
    keyLabelMap: {
      [NodeMetricType.GpuMemoryBandwidthUtilizationPerGpu]: "GPU Memory Bandwidth Utilization",
    },
  },
  [AdvancedWidgets.GpuNvlinkTransmittedBandwidthPerGpu]: {
    metricTypes: [NodeMetricType.GpuNvlinkTransmittedBandwidthPerGpu],
    chartTitle: "NVLink transmitted bandwidth",
    intervalLabel: EIntervalLabels.GpuUtilizationPerNodeWidget,
    chartId: "node-gpu-nvlink-transmitted-bandwidth-chart",
    getData: sharedGetData,
    measurementParser: widgetUtil.gpuUtilizationPerNodeParser,
    chartOptions: {
      ...sharedChartOptions,
      yAxisFormatFunction: formatGBs,
      highchartsLikeYAxis: true,
    },
    keyLabelMap: {
      [NodeMetricType.GpuNvlinkTransmittedBandwidthPerGpu]: "GPU NVLink Transmitted Bandwidth",
    },
  },
  [AdvancedWidgets.GpuNvlinkReceivedBandwidthPerGpu]: {
    metricTypes: [NodeMetricType.GpuNvlinkReceivedBandwidthPerGpu],
    chartTitle: "NVLink received bandwidth",
    intervalLabel: EIntervalLabels.GpuUtilizationPerNodeWidget,
    chartId: "node-gpu-nvlink-received-bandwidth-chart",
    getData: sharedGetData,
    measurementParser: widgetUtil.gpuUtilizationPerNodeParser,
    chartOptions: {
      ...sharedChartOptions,
      yAxisFormatFunction: formatGBs,
      highchartsLikeYAxis: true,
    },
    keyLabelMap: {
      [NodeMetricType.GpuNvlinkReceivedBandwidthPerGpu]: "GPU NVLink Received Bandwidth",
    },
  },
  [AdvancedWidgets.GpuPcieTransmittedBandwidthPerGpu]: {
    metricTypes: [NodeMetricType.GpuPcieTransmittedBandwidthPerGpu],
    chartTitle: "PCIe transmitted bandwidth",
    intervalLabel: EIntervalLabels.GpuUtilizationPerNodeWidget,
    chartId: "node-gpu-pcie-transmitted-bandwidth-chart",
    getData: sharedGetData,
    measurementParser: widgetUtil.gpuUtilizationPerNodeParser,
    chartOptions: {
      ...sharedChartOptions,
      yAxisFormatFunction: formatGBs,
      highchartsLikeYAxis: true,
    },
    keyLabelMap: {
      [NodeMetricType.GpuPcieTransmittedBandwidthPerGpu]: "GPU PCIe Transmitted Bandwidth",
    },
  },
  [AdvancedWidgets.GpuPcieReceivedBandwidthPerGpu]: {
    metricTypes: [NodeMetricType.GpuPcieReceivedBandwidthPerGpu],
    chartTitle: "PCIe received bandwidth",
    intervalLabel: EIntervalLabels.GpuUtilizationPerNodeWidget,
    chartId: "node-gpu-pcie-received-bandwidth-chart",
    getData: sharedGetData,
    measurementParser: widgetUtil.gpuUtilizationPerNodeParser,
    chartOptions: {
      ...sharedChartOptions,
      yAxisFormatFunction: formatGBs,
      highchartsLikeYAxis: true,
    },
    keyLabelMap: {
      [NodeMetricType.GpuPcieReceivedBandwidthPerGpu]: "GPU PCIe Received Bandwidth",
    },
  },
  [AdvancedWidgets.GpuOomkillBurstCountPerGpu]: {
    metricTypes: [NodeMetricType.GpuOomkillBurstCountPerGpu],
    chartTitle: "OOMKill burst count",
    tooltipSuffix: "",
    intervalLabel: EIntervalLabels.GpuUtilizationPerNodeWidget,
    chartId: "node-gpu-oomkill-burst-count-chart",
    getData: sharedGetData,
    measurementParser: widgetUtil.gpuUtilizationPerNodeParser,
    chartOptions: {
      ...sharedChartOptions,
    },
    keyLabelMap: {
      [NodeMetricType.GpuOomkillBurstCountPerGpu]: "GPU OOMKill Burst Count",
    },
  },
  [AdvancedWidgets.GpuOomkillIdleCountPerGpu]: {
    metricTypes: [NodeMetricType.GpuOomkillIdleCountPerGpu],
    chartTitle: "OOMKill idle count",
    tooltipSuffix: "",
    intervalLabel: EIntervalLabels.GpuUtilizationPerNodeWidget,
    chartId: "node-gpu-oomkill-idle-count-chart",
    getData: sharedGetData,
    measurementParser: widgetUtil.gpuUtilizationPerNodeParser,
    chartOptions: {
      ...sharedChartOptions,
    },
    keyLabelMap: {
      [NodeMetricType.GpuOomkillIdleCountPerGpu]: "GPU OOMKill Idle Count",
    },
  },
  [AdvancedWidgets.GpuOomkillSwapOutOfRamCountPerGpu]: {
    metricTypes: [NodeMetricType.GpuOomkillSwapOutOfRamCountPerGpu],
    chartTitle: "OOMKill swap out of RAM count",
    tooltipSuffix: "",
    intervalLabel: EIntervalLabels.GpuUtilizationPerNodeWidget,
    chartId: "node-gpu-oomkill-swap-out-of-ram-count-chart",
    getData: sharedGetData,
    measurementParser: widgetUtil.gpuUtilizationPerNodeParser,
    chartOptions: {
      ...sharedChartOptions,
    },
    keyLabelMap: {
      [NodeMetricType.GpuOomkillSwapOutOfRamCountPerGpu]: "GPU OOMKill Swap Out of RAM Count",
    },
  },
  [AdvancedWidgets.GpuGraphicsEngineActivityPerGpu]: {
    metricTypes: [NodeMetricType.GpuGraphicsEngineActivityPerGpu],
    chartTitle: "Graphics engine activity",
    tooltipSuffix: "%",
    intervalLabel: EIntervalLabels.GpuUtilizationPerNodeWidget,
    chartId: "node-gpu-graphics-engine-activity-chart",
    isPercentage: true,
    getData: sharedGetData,
    measurementParser: widgetUtil.gpuUtilizationPerNodeParser,
    chartOptions: {
      ...sharedChartOptions,
    },
    keyLabelMap: {
      [NodeMetricType.GpuGraphicsEngineActivityPerGpu]: "GPU Graphics Engine Activity",
    },
  },
  [AdvancedWidgets.GpuSmActivityPerGpu]: {
    metricTypes: [NodeMetricType.GpuSmActivityPerGpu],
    chartTitle: "SM activity",
    tooltipSuffix: "%",
    intervalLabel: EIntervalLabels.GpuUtilizationPerNodeWidget,
    chartId: "node-gpu-sm-activity-chart",
    isPercentage: true,
    getData: sharedGetData,
    measurementParser: widgetUtil.gpuUtilizationPerNodeParser,
    chartOptions: {
      ...sharedChartOptions,
    },
    keyLabelMap: {
      [NodeMetricType.GpuSmActivityPerGpu]: "GPU SM Activity",
    },
  },
  [AdvancedWidgets.GpuSmOccupancyPerGpu]: {
    metricTypes: [NodeMetricType.GpuSmOccupancyPerGpu],
    chartTitle: "SM occupancy",
    tooltipSuffix: "%",
    intervalLabel: EIntervalLabels.GpuUtilizationPerNodeWidget,
    chartId: "node-gpu-sm-occupancy-chart",
    isPercentage: true,
    getData: sharedGetData,
    measurementParser: widgetUtil.gpuUtilizationPerNodeParser,
    chartOptions: {
      ...sharedChartOptions,
    },
    keyLabelMap: {
      [NodeMetricType.GpuSmOccupancyPerGpu]: "GPU SM Occupancy",
    },
  },
  [AdvancedWidgets.GpuTensorActivityPerGpu]: {
    metricTypes: [NodeMetricType.GpuTensorActivityPerGpu],
    chartTitle: "Tensor activity",
    tooltipSuffix: "%",
    intervalLabel: EIntervalLabels.GpuUtilizationPerNodeWidget,
    chartId: "node-gpu-tensor-activity-chart",
    isPercentage: true,
    getData: sharedGetData,
    measurementParser: widgetUtil.gpuUtilizationPerNodeParser,
    chartOptions: {
      ...sharedChartOptions,
    },
    keyLabelMap: {
      [NodeMetricType.GpuTensorActivityPerGpu]: "GPU Tensor Activity",
    },
  },
  [AdvancedWidgets.GpuFp64EngineActivityPerGpu]: {
    metricTypes: [NodeMetricType.GpuFp64EngineActivityPerGpu],
    chartTitle: "FP64 engine activity",
    tooltipSuffix: "%",
    intervalLabel: EIntervalLabels.GpuUtilizationPerNodeWidget,
    chartId: "node-gpu-fp64-engine-activity-chart",
    isPercentage: true,
    getData: sharedGetData,
    measurementParser: widgetUtil.gpuUtilizationPerNodeParser,
    chartOptions: {
      ...sharedChartOptions,
    },
    keyLabelMap: {
      [NodeMetricType.GpuFp64EngineActivityPerGpu]: "GPU FP64 Engine Activity",
    },
  },
  [AdvancedWidgets.GpuFp32EngineActivityPerGpu]: {
    metricTypes: [NodeMetricType.GpuFp32EngineActivityPerGpu],
    chartTitle: "GPU FP32 engine activity",
    tooltipSuffix: "%",
    intervalLabel: EIntervalLabels.GpuUtilizationPerNodeWidget,
    chartId: "node-gpu-fp32-engine-activity-chart",
    isPercentage: true,
    getData: sharedGetData,
    measurementParser: widgetUtil.gpuUtilizationPerNodeParser,
    chartOptions: {
      ...sharedChartOptions,
    },
    keyLabelMap: {
      [NodeMetricType.GpuFp32EngineActivityPerGpu]: "GPU FP32 Engine Activity",
    },
  },
  [AdvancedWidgets.GpuFp16EngineActivityPerGpu]: {
    metricTypes: [NodeMetricType.GpuFp16EngineActivityPerGpu],
    chartTitle: "GPU FP16 engine activity",
    tooltipSuffix: "%",
    intervalLabel: EIntervalLabels.GpuUtilizationPerNodeWidget,
    chartId: "node-gpu-fp16-engine-activity-chart",
    isPercentage: true,
    getData: sharedGetData,
    measurementParser: widgetUtil.gpuUtilizationPerNodeParser,
    chartOptions: {
      ...sharedChartOptions,
    },
    keyLabelMap: {
      [NodeMetricType.GpuFp16EngineActivityPerGpu]: "GPU FP16 Engine Activity",
    },
  },
};

export const defaultWidgets = [
  {
    key: GenericWidgets.GpuUtilizationPerNodePerGpu,
    config: widgetConfigMapper[GenericWidgets.GpuUtilizationPerNodePerGpu],
  },
  {
    key: GenericWidgets.GpuMemoryUtilizationPerNode,
    config: widgetConfigMapper[GenericWidgets.GpuMemoryUtilizationPerNode],
  },
  { key: GenericWidgets.CpuUtilizationPerNode, config: widgetConfigMapper[GenericWidgets.CpuUtilizationPerNode] },
  {
    key: GenericWidgets.CpuMemoryUsageBytesPerNode,
    config: widgetConfigMapper[GenericWidgets.CpuMemoryUsageBytesPerNode],
  },
  {
    key: GenericWidgets.DcgmFiDevNvlinkBandwidthTotal,
    config: widgetConfigMapper[GenericWidgets.DcgmFiDevNvlinkBandwidthTotal],
  },
].map((widget) => ({
  ...widget,
  config: {
    ...widget.config,
    useCanonicalTimeBuckets: true,
    useConnectedTooltipAnchor: true,
  },
}));

export const advancedWidgets = [
  {
    key: AdvancedWidgets.GpuMemoryBandwidthUtilizationPerGpu,
    config: advancedWidgetConfigMapper[AdvancedWidgets.GpuMemoryBandwidthUtilizationPerGpu],
  },
  {
    key: AdvancedWidgets.GpuNvlinkTransmittedBandwidthPerGpu,
    config: advancedWidgetConfigMapper[AdvancedWidgets.GpuNvlinkTransmittedBandwidthPerGpu],
  },
  {
    key: AdvancedWidgets.GpuNvlinkReceivedBandwidthPerGpu,
    config: advancedWidgetConfigMapper[AdvancedWidgets.GpuNvlinkReceivedBandwidthPerGpu],
  },
  {
    key: AdvancedWidgets.GpuPcieTransmittedBandwidthPerGpu,
    config: advancedWidgetConfigMapper[AdvancedWidgets.GpuPcieTransmittedBandwidthPerGpu],
  },
  {
    key: AdvancedWidgets.GpuPcieReceivedBandwidthPerGpu,
    config: advancedWidgetConfigMapper[AdvancedWidgets.GpuPcieReceivedBandwidthPerGpu],
  },
  {
    key: AdvancedWidgets.GpuOomkillBurstCountPerGpu,
    config: advancedWidgetConfigMapper[AdvancedWidgets.GpuOomkillBurstCountPerGpu],
  },
  {
    key: AdvancedWidgets.GpuOomkillIdleCountPerGpu,
    config: advancedWidgetConfigMapper[AdvancedWidgets.GpuOomkillIdleCountPerGpu],
  },
  {
    key: AdvancedWidgets.GpuOomkillSwapOutOfRamCountPerGpu,
    config: advancedWidgetConfigMapper[AdvancedWidgets.GpuOomkillSwapOutOfRamCountPerGpu],
  },
  {
    key: AdvancedWidgets.GpuGraphicsEngineActivityPerGpu,
    config: advancedWidgetConfigMapper[AdvancedWidgets.GpuGraphicsEngineActivityPerGpu],
  },
  {
    key: AdvancedWidgets.GpuSmActivityPerGpu,
    config: advancedWidgetConfigMapper[AdvancedWidgets.GpuSmActivityPerGpu],
  },
  {
    key: AdvancedWidgets.GpuSmOccupancyPerGpu,
    config: advancedWidgetConfigMapper[AdvancedWidgets.GpuSmOccupancyPerGpu],
  },
  {
    key: AdvancedWidgets.GpuTensorActivityPerGpu,
    config: advancedWidgetConfigMapper[AdvancedWidgets.GpuTensorActivityPerGpu],
  },
  {
    key: AdvancedWidgets.GpuFp64EngineActivityPerGpu,
    config: advancedWidgetConfigMapper[AdvancedWidgets.GpuFp64EngineActivityPerGpu],
  },
  {
    key: AdvancedWidgets.GpuFp32EngineActivityPerGpu,
    config: advancedWidgetConfigMapper[AdvancedWidgets.GpuFp32EngineActivityPerGpu],
  },
  {
    key: AdvancedWidgets.GpuFp16EngineActivityPerGpu,
    config: advancedWidgetConfigMapper[AdvancedWidgets.GpuFp16EngineActivityPerGpu],
  },
].map((widget) => ({
  ...widget,
  config: {
    ...widget.config,
    useCanonicalTimeBuckets: true,
    useConnectedTooltipAnchor: true,
  },
}));
