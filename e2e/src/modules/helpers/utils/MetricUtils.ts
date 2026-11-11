/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { GroupedMetricsResponse, MetricsResponse, MetricsType } from '../../api/apis/clusterApi';
import http from 'http';
import { clusterApi, workloadsService } from '../../api';
import { flowRootState } from '../config/config';
import { WorkloadMetricType } from '../../api/apis/workloadsApi/generated/model/workloadMetricType';
import { PodMetricType } from '../../api/apis/workloadsApi/generated/model/podMetricType';
export class MetricUtils {
    public async getClusterMetrics(
        start: Date,
        end: Date,
        metricType: Array<MetricsType>,
        numberOfSamples?: number,
        groupBy?: 'Category' | 'Nodepool' | 'Project' | 'Department'
    ): Promise<{ response: http.IncomingMessage; body: GroupedMetricsResponse }> {
        // metricType - a workaround to overcome a bug in openapi-generator
        // https://github.com/OpenAPITools/openapi-generator/issues/9003
        return await clusterApi.clustersApi.getClusterMetrics(
            flowRootState.cluster.uuid,
            start,
            end,
            [metricType.join(',') as unknown as MetricsType],
            numberOfSamples,
            groupBy
        );
    }

    public async getNodepoolMetrics(
        nodepoolName: string,
        start: Date,
        end: Date,
        metricType: Array<MetricsType>,
        numberOfSamples?: number
    ): Promise<{ response: http.IncomingMessage; body: MetricsResponse }> {
        // metricType - a workaround to overcome a bug in openapi-generator
        // https://github.com/OpenAPITools/openapi-generator/issues/9003
        return await clusterApi.nodepoolsApi.getNodepoolMetrics(
            flowRootState.cluster.uuid,
            nodepoolName,
            start,
            end,
            [metricType.join(',') as unknown as MetricsType],
            numberOfSamples
        );
    }

    public async getWorkloadMetrics(
        workloadId: string,
        metricType: Array<WorkloadMetricType>,
        start: Date,
        end: Date,
        numberOfSamples?: number
    ): Promise<{ response: http.IncomingMessage; body: MetricsResponse }> {
        // metricType - a workaround to overcome a bug in openapi-generator
        // https://github.com/OpenAPITools/openapi-generator/issues/9003
        return await workloadsService.workloadsApi.getWorkloadMetrics(
            workloadId,
            [metricType.join(',') as unknown as WorkloadMetricType],
            start,
            end,
            numberOfSamples
        );
    }

    public async getWorkloadsScrapeableMetrics(
        accessToken: string
    ): Promise<{ response: http.IncomingMessage; body: string }> {
        workloadsService.metricsApi.accessToken = accessToken;
        return await workloadsService.metricsApi.getWorkloadsScrapeableMetrics();
    }

    public async getPodMetrics(
        workloadId: string,
        podId: string,
        metricType: Array<PodMetricType>,
        start: Date,
        end: Date,
        numberOfSamples?: number
    ): Promise<{ response: http.IncomingMessage; body: MetricsResponse }> {
        // metricType - a workaround to overcome a bug in openapi-generator
        // https://github.com/OpenAPITools/openapi-generator/issues/9003
        return await workloadsService.podsApi.getWorkloadPodMetrics(
            workloadId,
            podId,
            [metricType.join(',') as unknown as PodMetricType],
            start,
            end,
            numberOfSamples
        );
    }

    public getRecommendedStartEndForMetrics(): { start: Date; end: Date } {
        const start = new Date(Date.now());
        const end = new Date(start);
        start.setMinutes(start.getMinutes() - 1);
        start.setSeconds(start.getSeconds() - 30);
        end.setSeconds(end.getSeconds() - 5);

        return { start, end };
    }

    public validateMetrics(measurements: any[]): boolean {
        if (!measurements?.length) return false;
        return measurements.every((m) => {
            const values = m.values;
            const metricType = m.type;
            if (values?.length > 0) {
                return values.every((v) => {
                    const numValue = parseFloat(v.value);
                    const isGpuMetric = metricType.includes('GpuUtilization');
                    return isGpuMetric ? numValue > 0 && numValue <= 100 : numValue > 0;
                });
            } else {
                return false;
            }
        });
    }

    public validateNIMMetrics(response: any): boolean {
        const measurements = response.measurements || response;
        const histogram = response.histogram || [];
        const allMetrics = [...measurements, ...histogram];

        if (!allMetrics.length) return false;

        return allMetrics.every((metric: any) => {
            if (!metric.type || !metric.values?.length) return false;

            const isHistogramMetric = [
                'NIM_TIME_TO_FIRST_TOKEN_SECONDS',
                'NIM_E2E_REQUEST_LATENCY_SECONDS',
                'NIM_TIME_TO_FIRST_TOKEN_SECONDS_PERCENTILES',
                'NIM_E2E_REQUEST_LATENCY_SECONDS_PERCENTILES'
            ].includes(metric.type);

            return metric.values.every((value: any) => {
                if (!value.timestamp) return false;

                // For histogram metrics
                if (isHistogramMetric) {
                    if (!value.data) return false;
                    if (metric.type.includes('PERCENTILES')) {
                        if (!value.data.p50 || !value.data.p90 || !value.data.p99) return false;
                    }
                    return Object.values(value.data).every((val: any) => {
                        const num = parseFloat(val);
                        return !isNaN(num) && num >= 0;
                    });
                }

                // For regular metrics
                if (!value.value) return false;
                const num = parseFloat(value.value);
                if (isNaN(num) || num < 0) return false;
                if (metric.type === 'NIM_GPU_CACHE_USAGE_PERC' && num > 100) return false;

                return true;
            });
        });
    }
}
