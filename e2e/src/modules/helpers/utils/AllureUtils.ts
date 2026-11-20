/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { kubeCli } from '../../cli';
import * as yaml from 'js-yaml';
import { CustomResourceDefinition } from '../models/customResourceDefinition';
import { attachment } from 'allure-js-commons';
import { bashCli } from '../../cli';

export class AllureUtils {
    attachPodsLogs(podsNames: string[], namespace = 'runai', tailLines = 700): void {
        podsNames.forEach(async function (pod) {
            const podLogs = await kubeCli.getLogs(pod, namespace, tailLines);
            await attachment(pod + ' Logs', podLogs, 'text/plain');
        });
    }

    async attachWorkloadYaml(
        jobName: string,
        namespace: string,
        crd: CustomResourceDefinition
    ): Promise<void> {
        const wl_yaml = yaml.dump(await kubeCli.getWorkload(jobName, namespace, crd));
        await attachment('wl yaml', wl_yaml, 'text/plain');
    }

    async attachCrdYaml(crd: CustomResourceDefinition): Promise<void> {
        const crd_yaml = yaml.dump(await kubeCli.getCrds(crd));
        const yamlAttachmentName = `${crd.kind
            .toLowerCase()
            .replace(/([A-Z])/g, ' $1')
            .trim()} crd yaml`;
        await attachment(yamlAttachmentName, crd_yaml, 'text/plain');

        const describeOutput = await kubeCli.describeCrd(crd);
        const describeAttachmentName = `${crd.kind
            .toLowerCase()
            .replace(/([A-Z])/g, ' $1')
            .trim()} crd describe`;
        await attachment(describeAttachmentName, describeOutput, 'text/plain');
    }

    async attachPodsYaml(
        labelSelector: string = 'workloadKind',
        attachmentPrefix: string = 'User Workload'
    ): Promise<void> {
        try {
            const podsResponse = await kubeCli.CoreV1Api.listPodForAllNamespaces({
                labelSelector: labelSelector
            });

            if (podsResponse.items && podsResponse.items.length > 0) {
                const podsYaml = yaml.dump({
                    apiVersion: 'v1',
                    kind: 'List',
                    items: podsResponse.items
                });
                await attachment(`${attachmentPrefix} Pods YAML`, podsYaml, 'text/plain');
            }

            const podsDescribeCmd = `kubectl describe pods -A -l "${labelSelector}"`;
            const podsDescribe = bashCli.exec(podsDescribeCmd);
            if (podsDescribe && podsDescribe.trim()) {
                await attachment(`${attachmentPrefix} Pods Describe`, podsDescribe, 'text/plain');
            }
        } catch (error) {
            throw new Error(`Failed to collect pods information: ${error}`);
        }
    }

    async attachSystemComponentLogs(
        component: string,
        displayName?: string,
        namespace: string = 'runai',
        logTailLines: number = 500
    ): Promise<void> {
        try {
            const logs = await kubeCli.getLogs(component, namespace, logTailLines);
            if (logs) {
                const attachmentName = displayName || `${component} Logs`;
                await attachment(attachmentName, logs, 'text/plain');
            }
        } catch (error) {
            throw new Error(`Failed to collect ${component} logs: ${error}`);
        }
    }

    async attachPodsLogsBySelector(
        labelSelector: string = 'workloadKind',
        logTailLines: number = 500
    ): Promise<void> {
        try {
            const podsResponse = await kubeCli.CoreV1Api.listPodForAllNamespaces({
                labelSelector: labelSelector
            });

            if (podsResponse.items && podsResponse.items.length > 0) {
                const collectedPodLogs = new Set<string>();

                for (const pod of podsResponse.items) {
                    if (pod.metadata?.name && pod.metadata?.namespace) {
                        const namespace = pod.metadata.namespace;
                        const podName = pod.metadata.name;
                        const podLogKey = `${namespace}-${podName}`;

                        if (collectedPodLogs.has(podLogKey)) {
                            continue;
                        }

                        try {
                            const logs = await kubeCli.getLogs(podName, namespace, logTailLines);
                            if (logs && logs.trim()) {
                                await attachment(`${namespace}/${podName} Logs`, logs, 'text/plain');
                                collectedPodLogs.add(podLogKey);
                            }
                        } catch {
                            // Skip failed pods silently
                        }
                    }
                }
            }
        } catch (error) {
            throw new Error(`Failed to collect pods logs: ${error}`);
        }
    }
}
