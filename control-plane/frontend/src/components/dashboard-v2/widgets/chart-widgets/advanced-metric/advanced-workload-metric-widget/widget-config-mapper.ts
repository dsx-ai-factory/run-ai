import { EIntervalLabels } from "@/models/interval.model";
import type { IChartWidgetOptions, IGenericMetricWidgetConfig, TMetricType } from "@/models/chart.model";
import { formatK8sResourceMemory, formatUtil } from "@/utils/format.util/format.util";
import { PodMetricType } from "@/swagger-models/workloads-client";
import type { MetricsResponse } from "@/swagger-models/workloads-client";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { widgetUtil } from "@/utils/widget.util";

export enum AdvancedWidgets {
  GpuMemoryBandwidthUtilizationPerGpu = "GpuMemoryBandwidthUtilizationPerGpu",
  GpuNvlinkTransmittedBandwidthPerGpu = "GpuNvlinkTransmittedBandwidthPerGpu",
  GpuNvlinkReceivedBandwidthPerGpu = "GpuNvlinkReceivedBandwidthPerGpu",
  GpuPcieTransmittedBandwidthPerGpu = "GpuPcieTransmittedBandwidthPerGpu",
  GpuPcieReceivedBandwidthPerGpu = "GpuPcieReceivedBandwidthPerGpu",
  GpuGraphicsEngineActivityPerGpu = "GpuGraphicsEngineActivityPerGpu",
  GpuSmActivityPerGpu = "GpuSmActivityPerGpu",
  GpuSmOccupancyPerGpu = "GpuSmOccupancyPerGpu",
  GpuTensorActivityPerGpu = "GpuTensorActivityPerGpu",
  GpuFp64EngineActivityPerGpu = "GpuFp64EngineActivityPerGpu",
  GpuFp32EngineActivityPerGpu = "GpuFp32EngineActivityPerGpu",
  GpuFp16EngineActivityPerGpu = "GpuFp16EngineActivityPerGpu",
  GpuSwapMemoryBytesPerGpu = "GpuSwapMemoryBytesPerGpu",
}

const sharedGetData = async (
  metricTypes: Array<TMetricType>,
  start: string,
  end: string,
  numberOfSamples: number,
  args: string[],
): Promise<MetricsResponse> =>
  await workloadService.getWorkloadPodMetrics(
    args[0],
    args[1],
    metricTypes as PodMetricType[],
    start,
    end,
    numberOfSamples,
  );

const formatGBs = (val: string | number) => formatUtil.convertBytesToGB(val) + "GB/s";

const sharedChartOptions: IChartWidgetOptions = {
  type: "line",
  height: 250,
  sharedTooltip: true,
  showSharedCrosshair: true,
};

export const advancedWidgetConfigMapper: Record<AdvancedWidgets, IGenericMetricWidgetConfig> = {
  [AdvancedWidgets.GpuMemoryBandwidthUtilizationPerGpu]: {
    metricTypes: [PodMetricType.GpuMemoryBandwidthUtilizationPerGpu],
    chartTitle: "Memory bandwidth utilization",
    tooltipSuffix: "%",
    intervalLabel: EIntervalLabels.GpuUtilizationPerNodeWidget,
    chartId: "pod-gpu-memory-bandwidth-utilization-chart",
    isPercentage: true,
    getData: sharedGetData,
    measurementParser: widgetUtil.gpuUtilizationPerNodeParser,
    chartOptions: {
      ...sharedChartOptions,
    },
    keyLabelMap: {
      [PodMetricType.GpuMemoryBandwidthUtilizationPerGpu]: "GPU Memory Bandwidth Utilization",
    },
  },
  [AdvancedWidgets.GpuNvlinkTransmittedBandwidthPerGpu]: {
    metricTypes: [PodMetricType.GpuNvlinkTransmittedBandwidthPerGpu],
    chartTitle: "NVLink transmitted bandwidth",
    intervalLabel: EIntervalLabels.GpuUtilizationPerNodeWidget,
    chartId: "pod-gpu-nvlink-transmitted-bandwidth-chart",
    getData: sharedGetData,
    measurementParser: widgetUtil.gpuUtilizationPerNodeParser,
    chartOptions: {
      ...sharedChartOptions,
      yAxisFormatFunction: formatGBs,
      highchartsLikeYAxis: true,
    },
    keyLabelMap: {
      [PodMetricType.GpuNvlinkTransmittedBandwidthPerGpu]: "GPU NVLink Transmitted Bandwidth",
    },
  },
  [AdvancedWidgets.GpuNvlinkReceivedBandwidthPerGpu]: {
    metricTypes: [PodMetricType.GpuNvlinkReceivedBandwidthPerGpu],
    chartTitle: "NVLink received bandwidth",
    intervalLabel: EIntervalLabels.GpuUtilizationPerNodeWidget,
    chartId: "pod-gpu-nvlink-received-bandwidth-chart",
    getData: sharedGetData,
    measurementParser: widgetUtil.gpuUtilizationPerNodeParser,
    chartOptions: {
      ...sharedChartOptions,
      yAxisFormatFunction: formatGBs,
      highchartsLikeYAxis: true,
    },
    keyLabelMap: {
      [PodMetricType.GpuNvlinkReceivedBandwidthPerGpu]: "GPU NVLink Received Bandwidth",
    },
  },
  [AdvancedWidgets.GpuPcieTransmittedBandwidthPerGpu]: {
    metricTypes: [PodMetricType.GpuPcieTransmittedBandwidthPerGpu],
    chartTitle: "PCIe transmitted bandwidth",
    intervalLabel: EIntervalLabels.GpuUtilizationPerNodeWidget,
    chartId: "pod-gpu-pcie-transmitted-bandwidth-chart",
    getData: sharedGetData,
    measurementParser: widgetUtil.gpuUtilizationPerNodeParser,
    chartOptions: {
      ...sharedChartOptions,
      yAxisFormatFunction: formatGBs,
      highchartsLikeYAxis: true,
    },
    keyLabelMap: {
      [PodMetricType.GpuPcieTransmittedBandwidthPerGpu]: "GPU PCIe Transmitted Bandwidth",
    },
  },
  [AdvancedWidgets.GpuPcieReceivedBandwidthPerGpu]: {
    metricTypes: [PodMetricType.GpuPcieReceivedBandwidthPerGpu],
    chartTitle: "PCIe received bandwidth",
    intervalLabel: EIntervalLabels.GpuUtilizationPerNodeWidget,
    chartId: "pod-gpu-pcie-received-bandwidth-chart",
    getData: sharedGetData,
    measurementParser: widgetUtil.gpuUtilizationPerNodeParser,
    chartOptions: {
      ...sharedChartOptions,
      yAxisFormatFunction: formatGBs,
      highchartsLikeYAxis: true,
    },
    keyLabelMap: {
      [PodMetricType.GpuPcieReceivedBandwidthPerGpu]: "GPU PCIe Received Bandwidth",
    },
  },
  [AdvancedWidgets.GpuGraphicsEngineActivityPerGpu]: {
    metricTypes: [PodMetricType.GpuGraphicsEngineActivityPerGpu],
    chartTitle: "Graphics engine activity",
    tooltipSuffix: "%",
    intervalLabel: EIntervalLabels.GpuUtilizationPerNodeWidget,
    chartId: "pod-gpu-graphics-engine-activity-chart",
    isPercentage: true,
    getData: sharedGetData,
    measurementParser: widgetUtil.gpuUtilizationPerNodeParser,
    chartOptions: {
      ...sharedChartOptions,
    },
    keyLabelMap: {
      [PodMetricType.GpuGraphicsEngineActivityPerGpu]: "GPU Graphics Engine Activity",
    },
  },
  [AdvancedWidgets.GpuSmActivityPerGpu]: {
    metricTypes: [PodMetricType.GpuSmActivityPerGpu],
    chartTitle: "SM activity",
    tooltipSuffix: "%",
    intervalLabel: EIntervalLabels.GpuUtilizationPerNodeWidget,
    chartId: "pod-gpu-sm-activity-chart",
    isPercentage: true,
    getData: sharedGetData,
    measurementParser: widgetUtil.gpuUtilizationPerNodeParser,
    chartOptions: {
      ...sharedChartOptions,
    },
    keyLabelMap: {
      [PodMetricType.GpuSmActivityPerGpu]: "GPU SM Activity",
    },
  },
  [AdvancedWidgets.GpuSmOccupancyPerGpu]: {
    metricTypes: [PodMetricType.GpuSmOccupancyPerGpu],
    chartTitle: "SM occupancy",
    tooltipSuffix: "%",
    intervalLabel: EIntervalLabels.GpuUtilizationPerNodeWidget,
    chartId: "pod-gpu-sm-occupancy-chart",
    isPercentage: true,
    getData: sharedGetData,
    measurementParser: widgetUtil.gpuUtilizationPerNodeParser,
    chartOptions: {
      ...sharedChartOptions,
    },
    keyLabelMap: {
      [PodMetricType.GpuSmOccupancyPerGpu]: "GPU SM Occupancy",
    },
  },
  [AdvancedWidgets.GpuTensorActivityPerGpu]: {
    metricTypes: [PodMetricType.GpuTensorActivityPerGpu],
    chartTitle: "Tensor activity",
    tooltipSuffix: "%",
    intervalLabel: EIntervalLabels.GpuUtilizationPerNodeWidget,
    chartId: "pod-gpu-tensor-activity-chart",
    isPercentage: true,
    getData: sharedGetData,
    measurementParser: widgetUtil.gpuUtilizationPerNodeParser,
    chartOptions: {
      ...sharedChartOptions,
    },
    keyLabelMap: {
      [PodMetricType.GpuTensorActivityPerGpu]: "GPU Tensor Activity",
    },
  },
  [AdvancedWidgets.GpuFp64EngineActivityPerGpu]: {
    metricTypes: [PodMetricType.GpuFp64EngineActivityPerGpu],
    chartTitle: "FP64 engine activity",
    tooltipSuffix: "%",
    intervalLabel: EIntervalLabels.GpuUtilizationPerNodeWidget,
    chartId: "pod-gpu-fp64-engine-activity-chart",
    isPercentage: true,
    getData: sharedGetData,
    measurementParser: widgetUtil.gpuUtilizationPerNodeParser,
    chartOptions: {
      ...sharedChartOptions,
    },
    keyLabelMap: {
      [PodMetricType.GpuFp64EngineActivityPerGpu]: "GPU FP64 Engine Activity",
    },
  },
  [AdvancedWidgets.GpuFp32EngineActivityPerGpu]: {
    metricTypes: [PodMetricType.GpuFp32EngineActivityPerGpu],
    chartTitle: "GPU FP32 engine activity",
    tooltipSuffix: "%",
    intervalLabel: EIntervalLabels.GpuUtilizationPerNodeWidget,
    chartId: "pod-gpu-fp32-engine-activity-chart",
    isPercentage: true,
    getData: sharedGetData,
    measurementParser: widgetUtil.gpuUtilizationPerNodeParser,
    chartOptions: {
      ...sharedChartOptions,
    },
    keyLabelMap: {
      [PodMetricType.GpuFp32EngineActivityPerGpu]: "GPU FP32 Engine Activity",
    },
  },
  [AdvancedWidgets.GpuFp16EngineActivityPerGpu]: {
    metricTypes: [PodMetricType.GpuFp16EngineActivityPerGpu],
    chartTitle: "GPU FP16 engine activity",
    tooltipSuffix: "%",
    intervalLabel: EIntervalLabels.GpuUtilizationPerNodeWidget,
    chartId: "pod-gpu-fp16-engine-activity-chart",
    isPercentage: true,
    getData: sharedGetData,
    measurementParser: widgetUtil.gpuUtilizationPerNodeParser,
    chartOptions: {
      ...sharedChartOptions,
    },
    keyLabelMap: {
      [PodMetricType.GpuFp16EngineActivityPerGpu]: "GPU FP16 Engine Activity",
    },
  },
  [AdvancedWidgets.GpuSwapMemoryBytesPerGpu]: {
    metricTypes: [PodMetricType.GpuSwapMemoryBytesPerGpu],
    chartTitle: "GPU swap memory",
    tooltipSuffix: "bytes",
    intervalLabel: EIntervalLabels.GpuUtilizationPerNodeWidget,
    chartId: "pod-gpu-swap-memory-bytes-chart",
    getData: sharedGetData,
    measurementParser: widgetUtil.gpuUtilizationPerNodeParser,
    chartOptions: {
      ...sharedChartOptions,
      yAxisFormatFunction: formatK8sResourceMemory,
      highchartsLikeYAxis: true,
    },
    keyLabelMap: {
      [PodMetricType.GpuSwapMemoryBytesPerGpu]: "GPU Swap Memory",
    },
  },
};

export const podMetricAdvancedWidgets = [
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
  {
    key: AdvancedWidgets.GpuSwapMemoryBytesPerGpu,
    config: advancedWidgetConfigMapper[AdvancedWidgets.GpuSwapMemoryBytesPerGpu],
  },
].map((widget) => ({
  ...widget,
  config: {
    ...widget.config,
    useCanonicalTimeBuckets: true,
    useConnectedTooltipAnchor: true,
  },
}));
