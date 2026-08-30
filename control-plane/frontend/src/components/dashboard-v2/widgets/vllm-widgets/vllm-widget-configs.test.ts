import { beforeEach, describe, expect, it, vi } from "vitest";
import { PodMetricType, WorkloadMetricType } from "@/swagger-models/workloads-client";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { podService } from "@/services/control-plane/pod.service/pod.service";
import { workloadWidgetConfig as nimConcurrencyConfig } from "@/components/dashboard-v2/widgets/nim-widgets/requests-concurrency-by-status-widget/requests-concurrency-by-status-widget-config";
import { workloadThroughputByStatusWidgetConfig as nimRequestCountConfig } from "@/components/dashboard-v2/widgets/nim-widgets/throughput-by-status-widget/throughput-by-status-widget-config";
import { workloadGpuKvCacheUtilizationWidgetConfig as nimKvCacheConfig } from "@/components/dashboard-v2/widgets/nim-widgets/gpu-kv-cache-utilization-widget/gpu-kv-cache-utilization-widget-config";
import { workloadTimeToFirstTokenWidgetConfig as nimTtftConfig } from "@/components/dashboard-v2/widgets/nim-widgets/time-to-first-token-widget/time-to-first-token-widget-config";
import { workloadTimeToFirstTokenByPercentilesWidgetConfig as nimTtftPercentilesConfig } from "@/components/dashboard-v2/widgets/nim-widgets/time-to-first-token-by-percentiles-widget/time-to-first-token-by-percentiles-widget-config";
import { workloadEndToEndRequestLatencyWidgetConfig as nimE2ELatencyConfig } from "@/components/dashboard-v2/widgets/nim-widgets/end-to-end-request-latency-widget/end-to-end-request-latency-widget-config";
import { workloadEndToEndLatencyByPercentilesWidgetConfig as nimE2EPercentilesConfig } from "@/components/dashboard-v2/widgets/nim-widgets/end-to-end-latency-by-percentiles-widget/end-to-end-latency-by-percentiles-widget-config";
import { vllmHistogramWidgets, vllmLinearWidgets } from "./vllm-widget-configs";

vi.mock("@/services/cluster/workload.service/workload.service", () => ({
  workloadService: { getWorkloadMetrics: vi.fn() },
}));
vi.mock("@/services/control-plane/pod.service/pod.service", () => ({
  podService: { getPodMetrics: vi.fn() },
}));

const allWorkloadMetrics = [...vllmLinearWidgets, ...vllmHistogramWidgets].flatMap(
  ({ workloadConfig }) => workloadConfig.metricTypes,
);
const allPodMetrics = [...vllmLinearWidgets, ...vllmHistogramWidgets].flatMap(({ podConfig }) => podConfig.metricTypes);

describe("vLLM widget configs", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(workloadService.getWorkloadMetrics).mockResolvedValue({ measurements: [] });
    vi.mocked(podService.getPodMetrics).mockResolvedValue({ measurements: [] });
  });

  it("uses every planned workload and pod vLLM metric enum", () => {
    const expectedWorkloadMetrics = Object.values(WorkloadMetricType).filter((metric) => metric.startsWith("VLLM_"));
    const expectedPodMetrics = Object.values(PodMetricType).filter((metric) => metric.startsWith("VLLM_"));

    expect(new Set(allWorkloadMetrics)).toEqual(new Set(expectedWorkloadMetrics));
    expect(new Set(allPodMetrics)).toEqual(new Set(expectedPodMetrics));
  });

  it("uses the NIM chart titles for equivalent metrics", () => {
    const vllmTitlesById = Object.fromEntries(
      [...vllmLinearWidgets, ...vllmHistogramWidgets].map(({ id, workloadConfig }) => [id, workloadConfig.chartTitle]),
    );

    expect(vllmTitlesById).toMatchObject({
      "vllm-concurrency": nimConcurrencyConfig.chartTitle,
      "vllm-throughput": nimRequestCountConfig.chartTitle,
      "vllm-kv-cache": nimKvCacheConfig.chartTitle,
      "vllm-ttft-distribution": nimTtftConfig.chartTitle,
      "vllm-ttft-percentiles": nimTtftPercentilesConfig.chartTitle,
      "vllm-e2e-distribution": nimE2ELatencyConfig.chartTitle,
      "vllm-e2e-percentiles": nimE2EPercentilesConfig.chartTitle,
    });
  });

  it("reserves enough space for the latency phase legend", () => {
    const widget = vllmHistogramWidgets.find(({ id }) => id === "vllm-latency-phase");
    expect(widget).toBeDefined();
    if (!widget) throw new Error("vLLM latency phase widget is missing");

    expect(widget.workloadConfig.chartOptions).toMatchObject({
      height: 300,
      adaptiveLegendLayout: true,
      grid: { bottom: 110 },
    });
  });

  it("routes workload and pod configs to the corresponding APIs and arguments", async () => {
    const widget = vllmLinearWidgets[0];
    await widget.workloadConfig.getData(widget.workloadConfig.metricTypes, "2026-01-01", "2026-01-02", 100, [
      "workload-id",
    ]);
    await widget.podConfig.getData(widget.podConfig.metricTypes, "2026-01-01", "2026-01-02", 100, [
      "workload-id",
      "pod-id",
    ]);

    expect(workloadService.getWorkloadMetrics).toHaveBeenCalledWith(
      "workload-id",
      expect.any(String),
      expect.any(String),
      widget.workloadConfig.metricTypes,
      100,
    );
    expect(podService.getPodMetrics).toHaveBeenCalledWith(
      "workload-id",
      "pod-id",
      expect.any(String),
      expect.any(String),
      widget.podConfig.metricTypes,
      100,
    );
  });

  it("normalizes KV cache ratios to UI percentages", async () => {
    vi.mocked(workloadService.getWorkloadMetrics).mockResolvedValue({
      measurements: [{ type: WorkloadMetricType.VllmKvCacheUsagePerc, values: [{ timestamp: null, value: "0.42" }] }],
    });
    const widget = vllmLinearWidgets.find(({ id }) => id === "vllm-kv-cache");
    expect(widget).toBeDefined();
    if (!widget) throw new Error("vLLM KV cache widget is missing");
    const config = widget.workloadConfig;

    const response = await config.getData(config.metricTypes, "2026-01-01", "2026-01-02", 100, ["workload-id"]);

    expect(response.measurements[0].values?.[0].value).toBe("42");
  });

  it("parses p50, p95, and p99 percentile series", () => {
    const widget = vllmHistogramWidgets.find(({ id }) => id === "vllm-ttft-percentiles");
    expect(widget).toBeDefined();
    if (!widget) throw new Error("vLLM TTFT percentile widget is missing");
    const parser = widget.workloadConfig.histogramParser;
    expect(parser).toBeDefined();
    if (!parser) throw new Error("vLLM TTFT percentile parser is missing");
    const parsed = parser({
      measurements: [],
      histogram: [
        {
          type: WorkloadMetricType.VllmTimeToFirstTokenSecondsPercentiles,
          values: [{ timestamp: "2026-01-01T00:00:00Z", data: { p50: "1", p95: "2", p99: "3" } }],
        },
      ],
    });

    expect(Object.keys(parsed)).toEqual(["P50", "P95", "P99"]);
  });
});
