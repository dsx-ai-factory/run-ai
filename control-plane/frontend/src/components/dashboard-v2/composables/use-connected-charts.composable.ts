import { connect, disconnect, getInstanceByDom } from "echarts/core";
import type { ECharts } from "echarts/core";
import { chartUtil } from "@/utils/chart.util";

export interface IConnectedChartsOptions {
  chartIds: Array<string>;
}

const MAX_REGISTRATION_RETRY_FRAMES = 60;
let connectedChartGroupCounter = 0;

export function useConnectedCharts(options: IConnectedChartsOptions) {
  const groupId = `chart-group-${Date.now()}-${connectedChartGroupCounter++}`;
  const chartInstances = new Map<string, ECharts>();
  const registrationRetryHandles = new Map<string, number>();
  const chartIds = options.chartIds;

  function registerChartToCrosshair(chartId: string): void {
    registerChart(chartId);
  }

  function unregisterChartFromCrosshair(chartId: string): void {
    unregisterChart(chartId);
  }

  function registerChart(chartId: string, instance?: ECharts): void {
    const chart = instance || getChartInstanceById(chartId);

    if (!chart) {
      requestChartRegistrationRetry(chartId);
      return;
    }

    chartInstances.set(chartId, chart);
    chart.group = groupId;
    connect(groupId);
  }

  function getChartInstanceById(chartId: string): ECharts | undefined {
    if (typeof document === "undefined") return undefined;

    const chartDom = document.getElementById(chartId);
    if (!chartDom) return undefined;

    return getInstanceByDom(chartDom) as ECharts | undefined;
  }

  function requestChartRegistrationRetry(chartId: string, retryCount = 0): void {
    if (typeof window === "undefined") return;
    if (registrationRetryHandles.has(chartId)) return;

    const handle = window.requestAnimationFrame(() => {
      registrationRetryHandles.delete(chartId);
      const chart = getChartInstanceById(chartId);
      if (chart) {
        registerChart(chartId, chart);
        return;
      }

      if (retryCount < MAX_REGISTRATION_RETRY_FRAMES) {
        requestChartRegistrationRetry(chartId, retryCount + 1);
      }
    });

    registrationRetryHandles.set(chartId, handle);
  }

  function unregisterChart(chartId: string): void {
    cancelChartRegistrationRetry(chartId);

    const instance = chartInstances.get(chartId);
    if (instance) {
      instance.group = "";
      chartInstances.delete(chartId);
    }

    if (chartInstances.size === 0) {
      disconnect(groupId);
    } else {
      connect(groupId);
    }
  }

  function cancelChartRegistrationRetry(chartId: string): void {
    if (typeof window === "undefined") return;

    const handle = registrationRetryHandles.get(chartId);
    if (handle === undefined) return;

    window.cancelAnimationFrame(handle);
    registrationRetryHandles.delete(chartId);
  }

  function zoomIn(xMin: number, xMax: number, currentChartId: string): void {
    chartInstances.forEach((chart, id) => {
      if (id !== currentChartId) {
        chart.dispatchAction({ type: "dataZoom", startValue: xMin, endValue: xMax }, { silent: true });
      }
    });
  }

  function resetChartZoom(chart: ECharts): void {
    chartUtil.resetChartDataZoomInstance(chart);
  }

  function resetZoom(_currentChartId: string): void {
    chartInstances.forEach(resetChartZoom);
  }

  function setZoomMode(active: boolean): void {
    chartInstances.forEach((chart) => {
      try {
        chart.dispatchAction(
          { type: "takeGlobalCursor", key: "dataZoomSelect", dataZoomSelectActive: active },
          { silent: true },
        );
        if (!active) chart.getZr().setCursorStyle("default");
      } catch {
        // chart may be disposed during transitions
      }
    });
  }

  function resizeConnectedCharts(): void {
    chartInstances.forEach((chart) => chart.resize());
  }

  return {
    zoomIn,
    resetZoom,
    setZoomMode,
    resizeConnectedCharts,
    registerChartToCrosshair,
    unregisterChartFromCrosshair,
    registerChart,
    unregisterChart,
    chartIds,
  };
}
