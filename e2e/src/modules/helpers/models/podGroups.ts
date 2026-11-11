/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { KubernetesObject } from '@kubernetes/client-node';
import { logger, waitUtils } from '../index';
import { expect } from 'chai';
import { V1Pod } from '@kubernetes/client-node/';

// PodGroup CRD from Kubernetes (scheduling.run.ai/v2alpha2)
export interface PodGroupSubGroup {
    [key: string]: any;
}

export interface PodGroupTopologyConstraint {
    /** PreferredTopologyLevel defines the preferred level in the topology hierarchy */
    preferredTopologyLevel?: string;
    /** RequiredTopologyLevel defines the maximal level in the topology hierarchy that all pods must be scheduled within */
    requiredTopologyLevel?: string;
    /** Topology specifies the name of the topology CRD that defines the physical layout */
    topology?: string;
}

export interface PodGroupSpec {
    /** The number of failed pods required for this podgroup to move to 'Failed' phase */
    backoffLimit?: number;
    /** The number of successful pods required for this podgroup to move to 'Successful' phase */
    completions?: number;
    /** Should add "Unschedulable" event to the pods or not */
    markUnschedulable?: boolean;
    /** MinMember defines the minimal number of members to run the PodGroup */
    minMember?: number;
    /** The number of pods which will try to run at any instant */
    parallelism?: number;
    /** Preemptibility determines if this PodGroup can be preempted by higher priority workloads */
    preemptibility?: 'preemptible' | 'non-preemptible';
    /** If specified, indicates the PodGroup's priority */
    priorityClassName?: string;
    /** Queue defines the queue to allocate resource for PodGroup */
    queue?: string;
    /** The number of scheduling cycles to try before marking the pod group as UnschedulableOnNodePool */
    schedulingBackoff?: number;
    /** SubGroups defines finer-grained subsets of pods within the PodGroup with individual scheduling constraints */
    subGroups?: PodGroupSubGroup[];
    /** TopologyConstraint defines the topology constraints for this PodGroup */
    topologyConstraint?: PodGroupTopologyConstraint;
}

export interface PodGroupCrd extends KubernetesObject {
    spec?: PodGroupSpec;
    status?: {
        [key: string]: any;
    };
}

export function validateBasicPodGroupData(
    podGroups: KubernetesObject[],
    expectedNumOfPodGroups: number
): void {
    expect(podGroups, 'Failed to get pod groups from the cluster').not.equal(undefined);
    expect(podGroups, 'Unexpected number of pod groups').lengthOf(expectedNumOfPodGroups);
}
