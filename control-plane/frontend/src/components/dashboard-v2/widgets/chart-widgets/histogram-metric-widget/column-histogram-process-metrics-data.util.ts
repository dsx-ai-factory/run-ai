import { type IColumnHistogramDataParser } from "@/models/chart.model";
import type {
  HistogramValue,
  MeasurementResponseValuesInner,
  MetricsCompositeResponse,
} from "@/swagger-models/workloads-client";

interface IProcessedMetricsData {
  timestamp: string;
  data: Array<[number, number]>;
  total: number;
}

/**
 * Converts data object to array of [key, value] pairs with numeric conversion
 */
function convertDataToArray(data: Record<string, string>): Array<[number, number]> {
  return Object.entries(data).map(
    ([key, value]) =>
      [
        parseFloat(key), // Convert key to number
        parseInt(value, 10), // Convert value to number
      ] as [number, number],
  );
}

/**
 * Sorts array by key values from min to max, with NaN values at the end
 */
function sortDataByKey(dataArray: Array<[number, number]>): Array<[number, number]> {
  return dataArray.sort((a, b) => {
    const aKey = a[0];
    const bKey = b[0];

    // If both are NaN, maintain original order
    if (isNaN(aKey) && isNaN(bKey)) return 0;
    // If only a is NaN, move it to the end
    if (isNaN(aKey)) return 1;
    // If only b is NaN, move it to the end
    if (isNaN(bKey)) return -1;
    // Normal numeric comparison
    return aKey - bKey;
  });
}

/**
 * Calculates differences between consecutive values: value[i] = value[i] - value[i-1]
 */
function calculateDifferences(sortedData: Array<[number, number]>): Array<[number, number]> {
  return sortedData.map(([key, value], index) => {
    if (index === 0) {
      // First value remains the same
      return [key, value] as [number, number];
    }
    // Calculate difference from previous value
    const previousValue = sortedData[index - 1][1];
    return [key, value - previousValue] as [number, number];
  });
}

/**
 * Calculates the total sum of all values
 */
function calculateTotal(data: Array<[number, number]>): number {
  return data.reduce((sum, [, value]) => sum + value, 0);
}

/**
 * Converts an array of processed metrics data to time series format
 * where each bucket becomes a series with historical data points
 *
 * @param processedDataArray - Array of processed metrics data with timestamps
 * @returns Array of time series data where each bucket becomes a series
 *
 * @example
 * ```typescript
 * const processedDataArray = [
 *   {
 *     timestamp: "2025-09-28T18:02:37.5Z",
 *     data: [[0.001, 0], [0.005, 0], [0.1, 8]],
 *     total: 8
 *   },
 *   {
 *     timestamp: "2025-09-28T18:03:37.5Z",
 *     data: [[0.001, 1], [0.005, 1], [0.1, 8]],
 *     total: 10
 *   }
 * ];
 *
 * const result: IColumnHistogramDataParser[] = [
 *   {
 *     name: "0.001",
 *     data: [
 *       {timestamp: "2025-09-28T18:02:37.5Z", value: 0},
 *       {timestamp: "2025-09-28T18:03:37.5Z", value: 1}
 *     ]
 *   },
 *   {
 *     name: "0.005",
 *     data: [
 *       {timestamp: "2025-09-28T18:02:37.5Z", value: 0},
 *       {timestamp: "2025-09-28T18:03:37.5Z", value: 1}
 *     ]
 *   },
 *   {
 *     name: "0.1",
 *     data: [
 *       {timestamp: "2025-09-28T18:02:37.5Z", value: 8},
 *       {timestamp: "2025-09-28T18:03:37.5Z", value: 8}
 *     ]
 *   }
 * ]
 * ```
 */
export function convertToTimeSeries(processedDataArray: IProcessedMetricsData[]): IColumnHistogramDataParser[] {
  if (processedDataArray.length === 0) {
    return [];
  }

  // Get all unique bucket keys from the first data point
  const bucketKeys = processedDataArray[0].data.map(([key]) => key);

  // Create a series for each bucket
  return bucketKeys.map((bucketKey) => {
    const seriesData: MeasurementResponseValuesInner[] = processedDataArray.map((dataPoint) => {
      // Find the value for this bucket key in the current data point
      const bucketData = dataPoint.data.find(([key]) => key === bucketKey);
      const value = bucketData ? bucketData[1] : 0;

      return {
        timestamp: dataPoint.timestamp || "",
        value: value.toString(),
      };
    });

    return {
      name: bucketKey.toString(),
      data: seriesData,
    };
  });
}

/**
 * Processes metrics data by extracting timestamp and converting data object
 * to a sorted array of [key, value] pairs where keys are sorted from min to max
 *
 * @param input - The input object containing timestamp and data
 * @returns Object with timestamp, sorted data array, and total
 *
 * @example
 * ```typescript
 * const input: HistogramValue = {
 *   "timestamp": "2025-09-28T18:02:37.5Z",
 *   "data": {
 *     "+Inf": "605",
 *     "0.001": "0",
 *     "0.005": "0",
 *     "0.01": "8",
 *     "0.02": "26",
 *     "0.04": "374",
 *     "0.06": "534",
 *     "0.08": "605",
 *     "0.1": "605",
 *     "0.25": "605",
 *     "0.5": "605",
 *     "0.75": "605",
 *     "1.0": "605",
 *     "10.0": "605",
 *     "2.5": "605",
 *     "5.0": "605",
 *     "7.5": "605"
 *   }
 * };
 *
 * const result: IProcessedMetricsData = {
 *   timestamp: "2025-09-28T18:02:37.5Z",
 *   data: [
 *     [0.001, 0],
 *     [0.005, 0],
 *     [0.01, 8],
 *     [0.02, 18],
 *     [0.04, 348],
 *     [0.06, 160],
 *     [0.08, 71],
 *     [0.1, 0],
 *     [0.25, 0],
 *     [0.5, 0],
 *     [0.75, 0],
 *     [1.0, 0],
 *     [2.5, 0],
 *     [5.0, 0],
 *     [7.5, 0],
 *     [10.0, 0],
 *     [Infinity, 0]
 *   ],
 *   total: 605
 * }
 * ```
 */
export function processMetricsData(input: HistogramValue): IProcessedMetricsData {
  const timestamp = input.timestamp || "";

  const dataArray = convertDataToArray(input.data);
  const sortedData = sortDataByKey(dataArray);
  const dataWithDifferences = calculateDifferences(sortedData);
  const total = calculateTotal(dataWithDifferences);

  return {
    timestamp,
    data: dataWithDifferences,
    total,
  };
}

/**
 * Parses histogram metrics response and converts it to column histogram format
 * Handles NaN values by replacing them with "+Inf" label
 *
 * @param metricsResponse - The metrics composite response containing histogram data
 * @returns Array of column histogram data parsers with time series data
 *
 * @example
 * ```typescript
 * const metricsResponse: MetricsCompositeResponse = {
 *   histogram: [{
 *     values: [
 *       { timestamp: "2025-09-28T18:02:37.5Z", data: { "0.001": "0", "0.005": "0", "0.1": "8" } }
 *     ]
 *   }]
 * };
 *
 * const result = parseHistogramToColumnSeries(metricsResponse);
 * // Returns array of series with time series data, NaN values converted to "+Inf"
 * ```
 */
export function parseHistogramToColumnSeries(metricsResponse: MetricsCompositeResponse): IColumnHistogramDataParser[] {
  const values = metricsResponse.histogram?.[0]?.values;
  const processedDataArray = values?.map((item) => processMetricsData(item));
  let columnSeries: IColumnHistogramDataParser[] = processedDataArray ? convertToTimeSeries(processedDataArray) : [];

  columnSeries = columnSeries.map((series: IColumnHistogramDataParser) => {
    if (series.name === "NaN") {
      series.name = "+Inf";
    }
    return series;
  });

  return columnSeries;
}

/**
 * Maps histogram column series data into custom ranges and sums the values
 * @param columnSeries - Array of column histogram data from parseHistogramToColumnSeries
 * @param getRangeKey - Function that determines which range a bucket value belongs to
 * @param rangeKeys - Array of range keys in the desired order
 * @param defaultRangeForNaN - Optional range key to use for NaN/Inf values (defaults to last range)
 * @returns Array of aggregated data grouped by ranges
 *
 * @example
 * ```typescript
 * const getRangeKey = (value: number) => {
 *   if (value <= 0.25) return "0-0.25";
 *   if (value <= 0.5) return "0.25-0.5";
 *   return "5+";
 * };
 * const rangeKeys = ["0-0.25", "0.25-0.5", "5+"];
 * const result = mapHistogramToRanges(columnSeries, getRangeKey, rangeKeys);
 * ```
 */
export function mapHistogramToRanges(
  columnSeries: IColumnHistogramDataParser[],
  getRangeKey: (bucketValue: number) => string,
  rangeKeys: string[],
  defaultRangeForNaN?: string,
  rangeSuffix = "",
): IColumnHistogramDataParser[] {
  // Use the last range as default for NaN values if not specified
  const nanRange = defaultRangeForNaN || (rangeKeys.length > 0 ? rangeKeys[rangeKeys.length - 1] : "");

  // Initialize range map with empty data arrays
  const rangeMap = new Map<string, Map<string, number>>();
  rangeKeys.forEach((key) => rangeMap.set(key, new Map()));

  // Process each series
  columnSeries.forEach((series) => {
    const bucketValue = parseFloat(series.name);

    // Handle if the name is not a valid number (e.g., "+Inf" or "NaN")
    if (isNaN(bucketValue)) {
      const rangeData = rangeMap.get(nanRange);
      if (rangeData) {
        series.data.forEach((dataPoint) => {
          const timestamp = dataPoint.timestamp || "";
          const value = parseFloat(dataPoint.value || "0");
          rangeData.set(timestamp, (rangeData.get(timestamp) || 0) + value);
        });
      }
      return;
    }

    // Determine the range and aggregate the data
    const rangeKey = getRangeKey(bucketValue);
    const rangeData = rangeMap.get(rangeKey);
    if (rangeData) {
      series.data.forEach((dataPoint) => {
        const timestamp = dataPoint.timestamp || "";
        const value = parseFloat(dataPoint.value || "0");
        rangeData.set(timestamp, (rangeData.get(timestamp) || 0) + value);
      });
    }
  });

  // Convert the map back to IColumnHistogramDataParser format
  const result: IColumnHistogramDataParser[] = rangeKeys.map((rangeKey) => {
    const rangeData = rangeMap.get(rangeKey);
    const dataArray = rangeData
      ? Array.from(rangeData.entries()).map(([timestamp, value]) => ({
          timestamp,
          value: value.toString(),
        }))
      : [];

    return {
      name: `${rangeKey} ${rangeSuffix}`.trim(),
      data: dataArray,
    };
  });

  return result;
}
