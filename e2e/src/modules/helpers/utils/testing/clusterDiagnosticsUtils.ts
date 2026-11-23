/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { bashCli } from '../../../cli';
import { logger } from '../../logger/logger';

function run(command: string): string {
    const output = bashCli.execWithOutput(command, false);
    return `${output.stdout ?? ''}${output.stderr ?? ''}`.trim();
}

export function logClusterDiagnostics(namespace: string, workloadName: string): void {
    const sections = [
        ['Nodes', run('kubectl --request-timeout=15s get nodes -o wide')],
        ['Pods', run('kubectl --request-timeout=15s get pods -A -o wide')],
        [
            'Deployments in runai namespaces',
            run(
                'kubectl --request-timeout=15s get deployments -n runai -o wide; kubectl --request-timeout=15s get deployments -n runai-backend -o wide'
            )
        ],
        [`Pods in ${namespace}`, run(`kubectl --request-timeout=15s describe pods -n ${namespace}`)],
        [
            `Recent events in ${namespace}`,
            run(
                `kubectl --request-timeout=15s get events -n ${namespace} --sort-by=.lastTimestamp | tail -40`
            )
        ],
        [
            `Workload resources for ${workloadName}`,
            run(
                `kubectl --request-timeout=15s get trainingworkload,runaiJob,job ${workloadName} -n ${namespace} -o yaml`
            )
        ],
        [
            'Workload controller logs',
            run(
                'kubectl --request-timeout=15s logs -n runai deployment/workload-controller --tail=100; kubectl --request-timeout=15s logs -n runai deployment/anyworkload-controller --tail=100'
            )
        ]
    ];
    const diagnostics = sections.map(([title, output]) => `=== ${title} ===\n${output}`).join('\n\n');

    logger.error(`Cluster diagnostics after pod startup failure:\n${diagnostics}`);
}
