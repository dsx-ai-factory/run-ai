/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { V1CustomResourceDefinition, V1JSONSchemaProps } from '@kubernetes/client-node';

/** Older Dynamo CRDs reject this field instead of accepting a runtime override. */
export function supportsDynamoRuntimeVersionOverride(
    definition: V1CustomResourceDefinition,
    version: string
): boolean {
    const schema = definition.spec.versions.find((entry) => entry.name === version && entry.served)?.schema
        ?.openAPIV3Schema?.properties?.spec?.properties;
    const serviceSchema = schema?.services?.additionalProperties as V1JSONSchemaProps | undefined;
    const componentSchema = schema?.components?.items as V1JSONSchemaProps | undefined;
    const runtimeSchema = version === 'v1alpha1' ? serviceSchema : componentSchema;
    return runtimeSchema?.properties?.runtimeVersionOverride?.type === 'string';
}

/**
 * Python health servers exercise scheduling and metadata, not model inference. New operators
 * still require a Dynamo runtime version for them, because their image has no Dynamo tag.
 */
export function resolveDynamoMockRuntimeVersion(
    definition: V1CustomResourceDefinition,
    version: string,
    operatorImages: string[]
): string | undefined {
    if (!supportsDynamoRuntimeVersionOverride(definition, version)) {
        return undefined;
    }
    const versions = new Set(
        operatorImages.map(
            (image) => image.match(/:(?:v)?(\d+\.\d+\.\d+(?:[-+][\w.-]+)?)(?:@sha256:[a-f0-9]+)?$/)?.[1]
        )
    );
    if (versions.size !== 1 || versions.has(undefined)) {
        throw new Error('Cannot resolve one Dynamo operator release for the mock runtime override');
    }
    return [...versions][0];
}

/** Wait for both Grove status and the KAI topology consumed by Run:ai admission. */
export function isDynamoTopologyBindingInSync(binding: any, topology: any): boolean {
    const synced = binding.status?.schedulerTopologyStatuses?.find(
        (status) =>
            status.schedulerName === 'kai-scheduler' && status.topologyReference === topology.metadata.name
    );
    const drift = binding.status?.conditions?.find(
        (condition) => condition.type === 'SchedulerTopologyDrift'
    );
    return (
        binding.status?.observedGeneration >= binding.metadata.generation &&
        synced?.inSync === true &&
        synced.schedulerBackendTopologyObservedGeneration === topology.metadata.generation &&
        drift?.status === 'False' &&
        topology.metadata.annotations?.['run.ai/grove-topology-name'] === binding.metadata.name
    );
}
