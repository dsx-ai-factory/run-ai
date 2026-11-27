/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BaseWorkload } from '../models/baseWorkload';
import { crds, logger, workloadsUtilsV2, workloadUtils } from '../index';
import { ExternalWorkloadCrd } from '../models/externalWorkload';
import { get } from 'lodash-es';
import { expect } from 'chai';
import { WorkloadTypePreemptibility } from '../../api/apis/workloadsApi/generated/model/workloadTypePreemptibility';
import { kubeCli } from '../../cli';

/**
 * Generate a RayCluster YAML manifest as a string, with configurable fields.
 *
 * @param options - Configuration for the RayCluster
 * @returns YAML string for the RayCluster
 */
export class ExternalWorkloadUtils {
    async submitWorkloadAndGetWorkloadId(
        externalWorkload: BaseWorkload,
        projectId: string | number
    ): Promise<string> {
        const workloadCrd = externalWorkload.getCrd();
        logger.debug(`Applying workload: ${JSON.stringify(workloadCrd)}`);
        await workloadUtils.submit(externalWorkload);
        const workloadId = await workloadsUtilsV2.waitForWorkloadIdByWorkloadNameAndProjectId(
            externalWorkload.name,
            projectId as string
        );
        const workload = (await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId)).body;
        if (!workload) throw new Error('Workload not found');
        return workloadId;
    }

    // Helper function to validate ExternalWorkload CRD
    async validateExternalWorkload(
        dgdName: string,
        namespace: string,
        expectedPriority: string,
        expectedPreemptibility: WorkloadTypePreemptibility
    ): Promise<void> {
        const expectedExternalWorkloadName = `ew-dynamographdeployment-${dgdName}`;
        const externalWorkloadCrd = await kubeCli.getCrd<ExternalWorkloadCrd>(
            crds.externalWorkloadCrd,
            expectedExternalWorkloadName,
            namespace
        );

        if (!externalWorkloadCrd) {
            throw new Error(`ExternalWorkload CRD ${expectedExternalWorkloadName} not found`);
        }

        const phasePath = crds.externalWorkloadCrd.getPhasePath();
        const phase = get(externalWorkloadCrd, phasePath);
        logger.info(`ExternalWorkload CRD ${expectedExternalWorkloadName} found with phase: ${phase}`);

        expect(phase).to.equal('Running', 'ExternalWorkload should be in Running phase');
        expect(externalWorkloadCrd.spec.preemptibility).to.equal(
            expectedPreemptibility,
            `Preemptibility should be ${expectedPreemptibility}`
        );

        expect(externalWorkloadCrd.spec.priorityClassName).to.equal(
            expectedPriority,
            `Priority class should be ${expectedPriority}`
        );
    }

    async hasStructureElements(ewName: string, namespace: string): Promise<boolean> {
        const ew = await kubeCli.getCrd<ExternalWorkloadCrd>(crds.externalWorkloadCrd, ewName, namespace);
        const elements = (ew.status as any)?.structureElements?.elements;
        return Array.isArray(elements) && elements.length > 0;
    }

    async getStructureElements(ewName: string, namespace: string): Promise<any[]> {
        const ew = await kubeCli.getCrd<ExternalWorkloadCrd>(crds.externalWorkloadCrd, ewName, namespace);
        return (ew.status as any)?.structureElements?.elements || [];
    }

    async allPodsHaveElementLabels(
        namespace: string,
        deploymentName: string,
        expectedCount: number
    ): Promise<boolean> {
        const pods = await kubeCli.listPodsByNamespaceAndPrefix(namespace, deploymentName);
        if (pods.length < expectedCount) return false;
        return pods.every((pod) => pod.metadata?.labels?.['run.ai/element-1-id']);
    }

    async hasExpectedPodCount(
        namespace: string,
        deploymentName: string,
        expectedCount: number
    ): Promise<boolean> {
        const pods = await kubeCli.listPodsByNamespaceAndPrefix(namespace, deploymentName);
        return pods.length >= expectedCount;
    }
}
