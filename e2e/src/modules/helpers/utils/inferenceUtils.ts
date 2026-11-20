/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { flowRootState } from '../config/config';
import { WorkloadsCrudApiUtils } from './WorkloadsCrudUtils';
import { expect } from 'chai';
import { StatusCodes } from 'http-status-codes';
import { config } from '../config/config';
import { logger } from '../logger/logger';
import { ObjectGenerator } from '../ObjectGenerator';
import { WaitUtils } from './WaitUtils';
import { TrainingWorkload } from '../models/trainingWorkload';
import { PodStatus } from '../models/pods';
import { Operations } from '../Operations';
import { bashCli, kubeCli } from '../../cli';
import { exec } from 'child_process';
import { promisify } from 'util';
import { WorkloadDetailed } from '../../api/apis/workloadsApi/generated/model/workloadDetailed';
import { Connection } from '../../api/apis/workloadsApi/generated/model/connection';
import { WorkloadsUtilsV2 } from './WorkloadsUtilsV2';

export class InferenceUtils {
    private workloadsCrudUtils: WorkloadsCrudApiUtils;

    private waitUtils: WaitUtils;

    private workloadsUtilsV2: WorkloadsUtilsV2;

    private objectGenerator: ObjectGenerator;

    private operations: Operations;

    constructor() {
        this.workloadsCrudUtils = new WorkloadsCrudApiUtils();
        this.waitUtils = new WaitUtils();
        this.workloadsUtilsV2 = new WorkloadsUtilsV2();
        this.objectGenerator = new ObjectGenerator();
        this.operations = new Operations();
    }

    // ==================== Basic Inference Operations ====================

    /**
     * Creates an inference workload and returns [workloadId, workloadName]
     */
    async createInference(
        workloadName: string,
        workloadSpec: any,
        projectId: string,
        addSuffix = false,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<string[]> {
        const result = await this.workloadsCrudUtils.createInference(
            workloadName,
            addSuffix,
            flowRootState.cluster.uuid,
            projectId,
            workloadSpec,
            accessToken
        );

        expect(result.response.statusCode).to.equal(StatusCodes.ACCEPTED);

        const workloadIdReturn = result.body.workloadId;
        const workloadNameReturn = result.body.name;

        return [workloadIdReturn, workloadNameReturn];
    }

    /**
     * Updates an inference workload specification
     */
    async updateInferenceSpec(workloadId: string, workloadSpec: any): Promise<void> {
        const result = await this.workloadsCrudUtils.updateInferenceSpec(workloadId, workloadSpec);

        expect(result.response.statusCode).to.equal(StatusCodes.ACCEPTED);
    }

    // ==================== Knative/Serving Operations ====================

    /**
     * Sets the Knative progress deadline in seconds
     */
    setKnativeProgressDeadlineSeconds(timeoutSeconds: number): void {
        const cmd = `kubectl patch cm config-deployment -n knative-serving --type='merge' -p '{"data": {"progress-deadline": "${timeoutSeconds}s"}}'`;
        bashCli.exec(cmd);
    }

    /**
     * Gets the Knative URL for an inference workload from the ServingPort connection
     */
    getKnativeUrlForInferenceWorkload(workload: WorkloadDetailed): string {
        const servingPortPredicate = function (connection: Connection): boolean {
            return connection.connectionType === 'ServingPort';
        };

        if (
            workload.environments &&
            workload.environments.length > 0 &&
            workload.environments[0].connections.some(servingPortPredicate)
        ) {
            return workload.environments[0].connections.find(servingPortPredicate)?.url;
        }

        return `http://${workload.name}.${workload.namespace}.svc.cluster.local`;
    }

    /**
     * Gets the internal cluster-local URL for an inference workload
     * Use this for sending requests from inside the cluster
     */
    getInternalUrlForInferenceWorkload(workload: WorkloadDetailed): string {
        return `http://${workload.name}.${workload.namespace}.svc.cluster.local`;
    }

    /**
     * Gets the serving URL for a workload by ID
     */
    async getServingUrl(workloadId: string): Promise<string> {
        const workload = await this.workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);
        return workload.body.urls.length > 0 ? workload.body.urls[0] : '';
    }

    /**
     * Builds a service URL from workload name, namespace and port
     */
    buildServiceUrl(workloadName: string, namespace: string, port: number): string {
        return `${workloadName}.${namespace}.svc.cluster.local:${port}`;
    }

    // ==================== Request/Testing Operations ====================

    async deleteInferenceClientJob(name: string, namespace: string): Promise<void> {
        try {
            await kubeCli.BatchV1Api.deleteNamespacedJob({
                name,
                namespace,
                propagationPolicy: 'Background'
            });
        } catch (error) {
            const statusCode = error?.statusCode ?? error?.body?.code;
            if (statusCode !== StatusCodes.NOT_FOUND) {
                throw error;
            }
        }
        await this.waitUtils.waitForBatchJobToExist(name, namespace, false, 60);
        await this.waitUtils.waitForWorkloadPodNotExist(namespace, name, 60);
    }

    /**
     * Sends an inference request to an internal URL using a client workload
     */
    async sendInferenceRequest(
        namespace: string,
        url: string,
        token = '',
        expectSuccess = true,
        options?: {
            method?: 'GET' | 'POST';
            headers?: Record<string, string>;
            data?: string;
        }
    ): Promise<string> {
        // Build the curl command with provided options
        let curlCommand = 'curl --fail';

        // Add HTTP method if specified
        if (options?.method) {
            curlCommand += ` -X ${options.method}`;
        }

        // Add URL
        curlCommand += ` ${url}`;

        // Add authorization header if token provided
        if (token !== '') {
            curlCommand += ` -H "Authorization: Bearer ${token}"`;
        }

        // Add custom headers
        if (options?.headers) {
            Object.entries(options.headers).forEach(([key, value]) => {
                curlCommand += ` -H "${key}: ${value}"`;
            });
        }

        // Add data payload
        if (options?.data) {
            curlCommand += ` -d '${options.data}'`;
        }

        // Base64 encode the curl command to avoid quote escaping issues
        const curlCommandBase64 = Buffer.from(curlCommand).toString('base64');
        const command = `echo '${curlCommandBase64}' | base64 -d | bash`;

        const clientWorkloadName = this.objectGenerator.generateNameByPrefix('inference-client');
        logger.info(
            `creating client job ${clientWorkloadName} in namespace ${namespace} with command ${command}. expected status is ${expectSuccess ? PodStatus.SUCCEEDED : PodStatus.FAILED}`
        );
        await kubeCli.BatchV1Api.createNamespacedJob({
            namespace,
            body: {
                apiVersion: 'batch/v1',
                kind: 'Job',
                metadata: { name: clientWorkloadName, namespace },
                spec: {
                    backoffLimit: 0,
                    template: {
                        metadata: { labels: { app: clientWorkloadName } },
                        spec: {
                            restartPolicy: 'Never',
                            containers: [
                                {
                                    name: 'client',
                                    image: config.quickStartImage.networkTools,
                                    command: ['/bin/bash', '-c', command]
                                }
                            ]
                        }
                    }
                }
            }
        });
        if (expectSuccess) {
            await this.waitUtils.waitForBatchJobToBeCompleted(clientWorkloadName, namespace, 90);
        } else {
            await this.waitUtils.waitForBatchJobToBeFailed(clientWorkloadName, namespace, 90);
        }
        logger.info('client job completed as expected');
        return clientWorkloadName;
    }

    /**
     * Sends a chat completion request to an internal URL using a client workload
     * Sends a POST request to /v1/chat/completions endpoint
     */
    async sendModelPromptRequest(
        namespace: string,
        url: string,
        model: string,
        prompt = 'Hello, how are you?',
        token = '',
        expectSuccess = true
    ): Promise<string> {
        // Ensure URL ends without trailing slash for proper path concatenation
        const baseUrl = url.endsWith('/') ? url.slice(0, -1) : url;
        const chatUrl = `${baseUrl}/v1/chat/completions`;

        // Build the JSON payload
        const payloadString = `{"model": "${model}", "messages": [{"role": "user", "content": "${prompt}"}], "max_tokens": 50}`;

        logger.info(`sending chat completion request to ${chatUrl} with model ${model}`);
        return await this.sendInferenceRequest(namespace, chatUrl, token, expectSuccess, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            data: payloadString
        });
    }

    /**
     * Applies Ingress URL rewrite fix for NIM services with path-based routing
     */
    private async applyIngressUrlRewriteFix(workloadName: string, namespace: string): Promise<void> {
        try {
            const rewriteAnnotationCmd = `kubectl get ingress ${workloadName} -n ${namespace} -o jsonpath='{.metadata.annotations.nginx\\.ingress\\.kubernetes\\.io/rewrite-target}'`;
            const rewriteAnnotation = bashCli.exec(rewriteAnnotationCmd);

            const currentPathCmd = `kubectl get ingress ${workloadName} -n ${namespace} -o jsonpath='{.spec.rules[0].http.paths[0].path}'`;
            const currentPath = bashCli.exec(currentPathCmd);

            if (!rewriteAnnotation && currentPath !== '' && currentPath !== '/') {
                logger.info('Applying Ingress URL rewrite fix for path-based routing...');
                const newPath = `${currentPath}(/|$)(.*)`;
                const patchCmd = `kubectl patch ingress ${workloadName} -n ${namespace} --type=json -p='[
                    {"op": "add", "path": "/metadata/annotations/nginx.ingress.kubernetes.io~1rewrite-target", "value": "/$2"},
                    {"op": "replace", "path": "/spec/rules/0/http/paths/0/pathType", "value": "ImplementationSpecific"},
                    {"op": "replace", "path": "/spec/rules/0/http/paths/0/path", "value": "${newPath}"}
                ]'`;
                bashCli.exec(patchCmd);
                await this.waitUtils.sleep(15000);
            }
        } catch (e) {
            logger.warn(`Failed to apply Ingress fix: ${e.message}`);
        }
    }

    /**
     * Sends an inference request to an external URL
     */
    async sendInferenceRequestExternalUrl(
        url: string,
        token = '',
        expectSuccess = true,
        treat404AsSuccess = false,
        workloadName?: string,
        namespace?: string
    ): Promise<void> {
        if (workloadName && namespace) {
            await this.applyIngressUrlRewriteFix(workloadName, namespace);
        }

        let curlCommand = `curl --fail ${url}`;
        if (token !== '') {
            curlCommand += ` -H "Authorization: Bearer ${token}"`;
        }

        const command = `/bin/bash -c '${curlCommand}'`;
        const maxRetries = 4;

        for (let attempt = 1; attempt <= maxRetries; attempt++) {
            try {
                const execAsync = promisify(exec);

                await this.waitUtils.sleep(1000);
                logger.info(
                    `executing external curl command (attempt ${attempt}/${maxRetries}): ${command}. expected to ${expectSuccess ? 'succeed' : 'fail with 401'}`
                );
                await execAsync(command);

                if (!expectSuccess) {
                    throw new Error('expected request to fail with 401 but it succeeded');
                }

                logger.info('external curl command completed successfully');
                return;
            } catch (error) {
                const errorMessage = error.message || '';
                const httpErrorMatch = errorMessage.match(/The requested URL returned error: (\d+)/);
                const httpErrorCode = httpErrorMatch ? parseInt(httpErrorMatch[1]) : null;

                logger.debug(
                    `Curl failed with error: ${errorMessage}, extracted HTTP code: ${httpErrorCode}`
                );

                // 503 = backend not ready; 404 (when not expected as success) = nginx ingress rule
                // not yet propagated. Both are transient startup races — retry the same way.
                if (
                    attempt < maxRetries &&
                    (httpErrorCode === 503 || (httpErrorCode === 404 && !treat404AsSuccess))
                ) {
                    logger.warn(`Got ${httpErrorCode}, retrying (attempt ${attempt}/${maxRetries})`);
                    continue;
                }

                if (expectSuccess) {
                    if (httpErrorCode === 404 && treat404AsSuccess) {
                        logger.info('external curl command returned 404 but treating as success');
                        return;
                    }
                    logger.error(`external curl command failed unexpectedly: ${error.message}`);
                    throw error;
                } else {
                    if (httpErrorCode === 401) {
                        logger.info('external curl command failed as expected with 401');
                        return;
                    } else {
                        logger.error(
                            `external curl command failed with unexpected error code ${httpErrorCode}. Expected 401 but got: ${error.message}`
                        );
                        throw new Error(`Expected 401 error but got ${httpErrorCode}: ${error.message}`);
                    }
                }
            }
        }

        throw new Error(`Failed after ${maxRetries} attempts`);
    }

    /**
     * Sends a gRPC request to an internal URL using a client workload
     */
    async sendGrpcRequest(namespace: string, grpcUrl: string, expectSuccess = true): Promise<string> {
        const urlWithoutProtocol = grpcUrl.replace(/^(grpc:\/\/|https:\/\/)/, '');

        const grpcCommand = `grpcurl -plaintext ${urlWithoutProtocol} list`;
        const command = `/bin/bash -c '${grpcCommand}'`;

        const clientWorkloadName = this.objectGenerator.generateNameByPrefix('grpc-client');
        const clientWorkload = new TrainingWorkload(
            config.quickStartImage.networkTools,
            namespace,
            clientWorkloadName
        );
        clientWorkload.command = command;
        clientWorkload.backoffLimit = 0;
        logger.info(
            `creating gRPC client workload ${clientWorkloadName} in namespace ${namespace} to test gRPC endpoint: ${urlWithoutProtocol}. expected status: ${expectSuccess ? PodStatus.SUCCEEDED : PodStatus.FAILED}`
        );
        const expectedStatus = expectSuccess ? PodStatus.SUCCEEDED : PodStatus.FAILED;
        await this.operations.submitWorkloads([clientWorkload], expectedStatus, 90);
        logger.info('gRPC client workload completed as expected');
        return clientWorkloadName;
    }

    /**
     * Sends a gRPC request to an external URL
     */
    async sendGrpcRequestExternalUrl(
        grpcUrl: string,
        workloadName: string,
        namespace: string,
        expectSuccess = true
    ): Promise<void> {
        const urlWithoutProtocol = grpcUrl.replace(/^(grpc:\/\/|https:\/\/)/, '');

        const [hostPort, ...pathParts] = urlWithoutProtocol.split('/');
        const path = pathParts.length > 0 ? '/' + pathParts.join('/') : '';

        const target = hostPort.includes(':') ? hostPort : `${hostPort}:443`;

        logger.info(`Parsed gRPC external URL: target=${target}, path=${path}`);

        const execAsync = promisify(exec);

        try {
            await this.waitUtils.sleep(1000);

            let grpcurlCmd = 'grpcurl -insecure';
            if (path) {
                grpcurlCmd += ` -authority "${hostPort}${path}"`;
            }
            grpcurlCmd += ` ${target} list`;

            logger.info(
                `executing external grpcurl command: ${grpcurlCmd}. expected to ${expectSuccess ? 'succeed' : 'fail'}`
            );

            const result = await execAsync(grpcurlCmd);

            if (!expectSuccess) {
                throw new Error('expected gRPC request to fail but it succeeded');
            }

            logger.info('external grpcurl command completed successfully');
            if (result.stdout) {
                logger.debug(`gRPC services found: ${result.stdout}`);
            }
        } catch (error) {
            const errorMessage = error.message || '';
            logger.debug(`grpcurl failed with error: ${errorMessage}`);

            if (expectSuccess) {
                logger.error(`external grpcurl command failed: ${error.message}`);
                throw error;
            } else {
                logger.info('external grpcurl command failed as expected');
            }
        }
    }
}
