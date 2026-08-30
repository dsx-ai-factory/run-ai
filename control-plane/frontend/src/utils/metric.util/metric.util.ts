import {
  type MetricsResponse,
  type GroupedMeasurementResponse,
  type GroupedMeasurementResponseValuesInner,
  MetricsType,
  NodeMetricType,
} from "@/swagger-models/cluster-service-client";
import { PodMetricType } from "@/swagger-models/workloads-client";
import type { TMeasurementsTimeAndValue } from "@/models/chart.model";

import { dateUtil } from "@/utils/date.util";
import { DATE_FORMATS } from "@/constants/date-formats";

const PERCENTAGE_METRIC_TYPES: ReadonlySet<string> = new Set<string>([
  MetricsType.GpuUtilization,
  MetricsType.GpuMemoryUtilization,
  MetricsType.CpuUtilization,
  MetricsType.CpuMemoryUtilization,
  NodeMetricType.GpuUtilizationPerGpu,
  NodeMetricType.GpuMemoryUtilizationPerGpu,
  NodeMetricType.GpuGraphicsEngineActivityPerGpu,
  NodeMetricType.GpuSmActivityPerGpu,
  NodeMetricType.GpuSmOccupancyPerGpu,
  NodeMetricType.GpuTensorActivityPerGpu,
  NodeMetricType.GpuFp64EngineActivityPerGpu,
  NodeMetricType.GpuFp32EngineActivityPerGpu,
  NodeMetricType.GpuFp16EngineActivityPerGpu,
  NodeMetricType.GpuMemoryBandwidthUtilizationPerGpu,
  PodMetricType.NimGpuCacheUsagePerc,
]);
const MAX_PERCENTAGE_VALUE = 100;

export const metricUtil = {
  mapMeasurementsToMap,
  getMeasurementTimestamp,
  getMeasurementValue,
  insertConstantValueIntoTimeframe,
};

function mapMeasurementsToMap(
  metricsResponse: MetricsResponse,
): Record<string, Array<GroupedMeasurementResponseValuesInner>> {
  const measurementMap: Record<string, Array<GroupedMeasurementResponseValuesInner>> = {};

  return metricsResponse.measurements.reduce(
    (
      measurementMap: Record<string, Array<GroupedMeasurementResponseValuesInner>>,
      measurement: GroupedMeasurementResponse,
    ) => {
      const rawValues = measurement.values || [];
      const values = PERCENTAGE_METRIC_TYPES.has(measurement.type)
        ? capValuesAt(rawValues, MAX_PERCENTAGE_VALUE)
        : rawValues;
      measurementMap[measurement.type] = [...(measurementMap[measurement.type] || []), ...values];
      return measurementMap;
    },
    measurementMap,
  );
}

function capValuesAt(
  values: Array<GroupedMeasurementResponseValuesInner>,
  maxValue: number,
): Array<GroupedMeasurementResponseValuesInner> {
  return values.map((measurement: GroupedMeasurementResponseValuesInner) => {
    if (measurement.value === null || measurement.value === undefined) return measurement;
    const numValue = Number(measurement.value);
    if (Number.isNaN(numValue) || numValue <= maxValue) return measurement;
    return { ...measurement, value: String(maxValue) };
  });
}

function getMeasurementValue(measurements: Array<GroupedMeasurementResponseValuesInner>): Array<number> {
  return measurements.map((measurement: GroupedMeasurementResponseValuesInner) => Number(measurement.value));
}

function getMeasurementTimestamp(measurements: Array<GroupedMeasurementResponseValuesInner>): Array<string> {
  return measurements.map((measurement: GroupedMeasurementResponseValuesInner) =>
    measurement.timestamp ? dateUtil.dateFormat(new Date(measurement.timestamp), DATE_FORMATS.SHORT_DATE_TIME) : "",
  );
}

function insertConstantValueIntoTimeframe(
  constantValue: number,
  timeframes: Array<number>,
  startTime?: number,
): Array<TMeasurementsTimeAndValue> {
  const startTimestamp = startTime || 0;

  return timeframes.map((timeframe: number) => {
    const value = timeframe >= startTimestamp ? constantValue : null;
    return [timeframe, value];
  });
}
