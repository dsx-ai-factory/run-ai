/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { bashCli, kubeCli } from '../../cli';
import { logger, flowRootState, clusterService, crds } from '..';
import { DisplayedCluster } from '../../api/apis/clusterApi';
import { CustomResourceDefinition } from '../models/customResourceDefinition';

export async function isKubeflowInstalled(): Promise<boolean> {
    // the indication for kubeflow installation is the existence of a deployment called 'training-operator'.
    // note that we cannot rely on the namespace as it can be installed under kubeflow, runai-kubeflow or some other namespace
    const deps = await kubeCli.AppsV1Api.listDeploymentForAllNamespaces();
    for (const dep of deps.items) {
        if (dep.metadata.name === 'training-operator') {
            return true;
        }
    }
    return false;
}

export async function isDynamoInstalled(): Promise<boolean> {
    // The CRD itself rather than a served version, since which versions Dynamo serves
    // changes between its releases while the CRD stays.
    return kubeCli.isCrdInstalled(crds.dynamoGraphDeploymentCrd);
}

const isDgdVersionServed = async (crd: CustomResourceDefinition): Promise<boolean> => {
    try {
        await kubeCli.CustomObjectsApi.listClusterCustomObject({
            group: crd.group,
            version: crd.version,
            plural: crd.plural
        });
        return true;
    } catch (e) {
        return false;
    }
};

export function isDynamoV1Alpha1Served(): Promise<boolean> {
    return isDgdVersionServed(crds.dynamoGraphDeploymentCrd);
}

// Dynamo serves v1beta1 only from 1.2.0 onwards; older installations serve v1alpha1 alone.
export function isDynamoV1Beta1Served(): Promise<boolean> {
    return isDgdVersionServed(crds.dynamoGraphDeploymentV1Beta1Crd);
}

export async function isKnativeInstalled(): Promise<boolean> {
    // Primary check: Query the cluster service for Knative availability from the dependency tree
    let knativeAvailableInClusterService = false;
    let cluster: DisplayedCluster;
    try {
        cluster = await clusterService.getCluster(
            flowRootState.users.system_administrator.accessToken,
            flowRootState.cluster.uuid
        );

        if (cluster?.status?.dependencies?.optional) {
            const knativeStatus = cluster.status.dependencies.optional['knative'];
            if (knativeStatus) {
                knativeAvailableInClusterService = knativeStatus.available;

                if (!knativeAvailableInClusterService && knativeStatus.reason) {
                    logger.warn(`Knative not available in cluster service: ${knativeStatus.reason}`);
                }
            }
        }
    } catch (error) {
        logger.warn(`Failed to check Knative availability from cluster service: ${error?.message}`);
    }

    // Secondary check: Verify the actual cluster state by checking for deployments
    let knativeDeploymentFound = false;

    try {
        const deps = await kubeCli.AppsV1Api.listDeploymentForAllNamespaces();
        for (const dep of deps.items) {
            if (dep.metadata.name === 'knative-operator' || dep.metadata.name === 'knative-openshift') {
                knativeDeploymentFound = true;
            }
        }
    } catch (error) {
        logger.warn(`Failed to check Knative deployments in cluster: ${error?.message}`);
    }

    // Check for misalignment between cluster service status and actual cluster state
    if (knativeAvailableInClusterService && !knativeDeploymentFound) {
        logger.warn(
            `MISALIGNMENT: Knative is marked as available in cluster service but deployment was not found in cluster. cluster data: ${JSON.stringify(cluster)}`
        );
    } else if (!knativeAvailableInClusterService && knativeDeploymentFound) {
        logger.warn(
            `MISALIGNMENT: Knative deployment found in cluster but not marked as available in cluster service. cluster data: ${JSON.stringify(cluster)}`
        );
    }

    // Return based on cluster service status (primary source of truth)
    return knativeAvailableInClusterService;
}

export async function isRayOperatorInstalled(): Promise<boolean> {
    // The indication for Ray operator installation is the existence of a deployment called 'kuberay-operator'.
    try {
        const deps = await kubeCli.AppsV1Api.listDeploymentForAllNamespaces();
        for (const dep of deps.items) {
            if (dep.metadata.name === 'kuberay-operator') {
                return true;
            }
        }
        return false;
    } catch (e) {
        // If an error occurs (e.g., namespace or pod not found), assume not installed
        return false;
    }
}

export async function isKGatewayInstalled(): Promise<boolean> {
    try {
        const deps = await kubeCli.AppsV1Api.listDeploymentForAllNamespaces();
        for (const dep of deps.items) {
            if (dep.metadata.name === 'kgateway') {
                return true;
            }
        }
        return false;
    } catch (e) {
        return false;
    }
}

export async function isInferencePoolCRDInstalled(): Promise<boolean> {
    try {
        await kubeCli.CustomObjectsApi.listClusterCustomObject({
            group: 'inference.networking.k8s.io',
            version: 'v1',
            plural: 'inferencepools'
        });
        return true;
    } catch (e) {
        return false;
    }
}

export async function isLwsInstalled(): Promise<boolean> {
    // Primary check: Query the cluster service for LWS availability from the dependency tree.
    let lwsAvailableInClusterService = false;
    let cluster: DisplayedCluster;
    try {
        cluster = await clusterService.getCluster(
            flowRootState.users.system_administrator.accessToken,
            flowRootState.cluster.uuid
        );

        if (cluster?.status?.dependencies?.optional) {
            const lwsStatus = cluster.status.dependencies.optional['lws'];
            if (lwsStatus) {
                lwsAvailableInClusterService = lwsStatus.available;

                if (!lwsAvailableInClusterService && lwsStatus.reason) {
                    logger.debug(`LWS not available in cluster service: ${lwsStatus.reason}`);
                }
            }
        }
    } catch (error) {
        logger.warn(`Failed to check LWS availability from cluster service: ${error?.message}`);
    }

    // Secondary check: Verify the actual cluster state by checking for deployments
    let lwsDeploymentFound = false;

    try {
        const deps = await kubeCli.AppsV1Api.listDeploymentForAllNamespaces();
        for (const dep of deps.items) {
            if (dep.metadata.name === 'lws-controller-manager') {
                lwsDeploymentFound = true;
                break;
            }
        }
    } catch (error) {
        logger.warn(`Failed to check LWS deployments in cluster: ${error?.message}`);
    }

    // Check for misalignment between cluster service status and actual cluster state
    if (lwsAvailableInClusterService && !lwsDeploymentFound) {
        logger.warn(
            `MISALIGNMENT: LWS is marked as available in cluster service but deployment was not found in cluster. cluster data: ${JSON.stringify(cluster)}`
        );
    } else if (!lwsAvailableInClusterService && lwsDeploymentFound) {
        logger.warn(
            `MISALIGNMENT: LWS deployment found in cluster but not marked as available in cluster service. cluster data: ${JSON.stringify(cluster)}`
        );
    }

    // Return based on cluster service status (primary source of truth)
    // but only if deployment is found when marked as available
    if (lwsAvailableInClusterService) {
        return lwsDeploymentFound;
    }

    return false;
}

export async function isNimOperatorInstalled(): Promise<boolean> {
    // The indication for NIM operator installation is the existence of a deployment called 'k8s-nim-operator'.
    const maxRetries = 3;
    const initialDelayMs = 1000;

    for (let attempt = 0; attempt <= maxRetries; attempt++) {
        try {
            const deps = await kubeCli.AppsV1Api.listDeploymentForAllNamespaces();
            for (const dep of deps.items) {
                if (dep.metadata.name.includes('k8s-nim-operator')) {
                    return true;
                }
            }
            return false;
        } catch (error) {
            if (attempt < maxRetries) {
                const delayMs = initialDelayMs * Math.pow(2, attempt);
                logger.warn(
                    `Failed to check NIM operator installation (attempt ${attempt + 1}/${maxRetries + 1}): ${error?.message}. Retrying in ${delayMs}ms...`
                );
                await new Promise((resolve) => setTimeout(resolve, delayMs));
            } else {
                logger.warn(
                    `Failed to check NIM operator installation after ${maxRetries + 1} attempts: ${error?.message}`
                );
                return false;
            }
        }
    }
    return false;
}

export async function isPrometheusAdapterInstalled(): Promise<boolean> {
    // The indication for Prometheus Adapter installation is the existence of a deployment called 'prometheus-adapter'.
    // This is required for HPA to use custom metrics.
    try {
        const deps = await kubeCli.AppsV1Api.listDeploymentForAllNamespaces();
        for (const dep of deps.items) {
            if (dep.metadata.name.includes('prometheus-adapter')) {
                return true;
            }
        }
        return false;
    } catch (error) {
        logger.warn(`Failed to check Prometheus Adapter installation: ${error?.message}`);
        return false;
    }
}

export async function isNvidiaDraDriverInstalled(): Promise<boolean> {
    // The indication for Nvidia DRA Driver installation is the existence of a deployment called 'nvidia-dra-driver-gpu'.
    try {
        const deps = await kubeCli.AppsV1Api.listDeploymentForAllNamespaces();
        for (const dep of deps.items) {
            if (dep.metadata.name.includes('nvidia-dra-driver-gpu')) {
                return true;
            }
        }
        return false;
    } catch (error) {
        logger.warn(`Failed to check Nvidia DRA Driver installation: ${error?.message}`);
        return false;
    }
}

export async function installRayOperator(): Promise<void> {
    // Check if Ray operator is already installed
    if (await isRayOperatorInstalled()) {
        logger.info('Ray operator is already installed.');
        return;
    }
    // Install Ray operator using the Run:ai CLI
    try {
        await bashCli.asyncExec('helm repo add kuberay https://ray-project.github.io/kuberay-helm/');
        await bashCli.asyncExec('helm repo update');
        await bashCli.asyncExec(
            'helm install kuberay-operator kuberay/kuberay-operator -n default --version 1.3.2'
        );
    } catch (error) {
        throw new Error(`Failed to install Ray operator: ${error.message}`);
    }
}
