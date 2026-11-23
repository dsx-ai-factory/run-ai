/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    config,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    PodStatus,
    TrainingWorkload,
    workloadsCrudUtils,
    workloadsUtilsV2
} from '../../../../../../modules/helpers';
import { readFileSync } from 'node:fs';
import { deflateSync } from 'node:zlib';
import { InferenceSpecSpec } from '../../../../../../modules/api/apis/workloadsApi/generated/model/inferenceSpecSpec';
import { expect } from 'chai';
import { StatusCodes } from 'http-status-codes';

export function createNimMetricsFixtureSpec(
    image = process.env.NIM_METRICS_IMAGE
): Pick<InferenceSpecSpec, 'image' | 'command' | 'args' | 'compute'> {
    if (image) {
        if (!/^ghcr\.io\/run-ai\/e2e-nim-metrics@sha256:[a-f0-9]{64}$/.test(image)) {
            throw new Error('NIM_METRICS_IMAGE must be a public e2e-nim-metrics digest reference');
        }
        return { image };
    }

    // Run the checked-in simulator without requiring a separately published fixture image.
    const files = Object.fromEntries(
        ['nim_llm.py', 'nim_service.proto', 'requirements.txt'].map((name) => [
            name,
            readFileSync(new URL(`./nimservice/nim-cpu-image/${name}`, import.meta.url), 'utf8')
        ])
    );
    const payload = deflateSync(Buffer.from(JSON.stringify(files)), { level: 9 }).toString('base64');
    const bootstrap = [
        'import base64, json, os, pathlib, subprocess, sys, tempfile, zlib',
        `files = json.loads(zlib.decompress(base64.b64decode("${payload}")))`,
        'root = pathlib.Path(tempfile.mkdtemp(prefix="runai-nim-metrics-"))',
        'for name, source in files.items():',
        '    (root / name).write_text(source)',
        'packages = root / "packages"',
        'subprocess.run([sys.executable, "-m", "pip", "install", "--disable-pip-version-check", "--no-cache-dir", "--no-compile", "--target", str(packages), "-r", str(root / "requirements.txt")], check=True)',
        'env = dict(os.environ, PYTHONPATH=str(packages) + os.pathsep + str(root), PYTHONUNBUFFERED="1")',
        'subprocess.run([sys.executable, "-m", "grpc_tools.protoc", "--proto_path=.", "--python_out=.", "--grpc_python_out=.", "nim_service.proto"], cwd=root, env=env, check=True)',
        'os.chdir(root)',
        'os.execve(sys.executable, [sys.executable, "-u", "nim_llm.py"], env)'
    ].join('\n');
    const encoded = deflateSync(Buffer.from(bootstrap), { level: 9 }).toString('base64');
    return {
        image: 'docker.io/library/python:3.11-slim@sha256:b7a9e5f6a93d6d4a31f129bbd4bdee69efccb6ded0fb230d78aafaf37af98377',
        compute: { cpuCoreRequest: 0.5, cpuMemoryRequest: '512M' },
        command: 'python',
        args: `-u -c "exec(__import__('zlib').decompress(__import__('base64').b64decode('${encoded}')))"`
    };
}

export async function createInference(
    workloadName,
    workloadSpec,
    project,
    addSuffix = false
): Promise<string[]> {
    const result = await workloadsCrudUtils.createInference(
        workloadName,
        addSuffix,
        flowRootState.cluster.uuid,
        String(project.id),
        workloadSpec
    );

    expect(result.response.statusCode).to.equal(StatusCodes.ACCEPTED);

    const workloadIdReturn = result.body.workloadId;
    const workloadNameReturn = result.body.name;

    return [workloadIdReturn, workloadNameReturn];
}

export async function updateInferenceSpec(workloadId, workloadSpec): Promise<void> {
    logger.debug('update inference workload');
    const result = await workloadsCrudUtils.updateInferenceSpec(workloadId, workloadSpec);

    expect(result.response.statusCode).to.equal(StatusCodes.ACCEPTED);
}

export async function createDistributedInference(workloadName, workloadSpec, project): Promise<string[]> {
    logger.debug(`creating distributed inference workload ${workloadName} in namespace ${project.namespace}`);
    const result = await workloadsCrudUtils.createDistributedInference(
        workloadName,
        true /*add suffix*/,
        flowRootState.cluster.uuid,
        String(project.id),
        workloadSpec
    );

    expect(result.response.statusCode).to.equal(StatusCodes.ACCEPTED);

    const workloadIdReturn = result.body.workloadId;
    const workloadNameReturn = result.body.name;

    return [workloadIdReturn, workloadNameReturn];
}

export async function getInternalServingUrl(workloadId: string): Promise<string> {
    const workload = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);
    const internalUrl = workload.body.urls?.find((url: string) => url.includes('cluster.local'));

    if (!internalUrl) {
        throw new Error(`Internal service URL not found for workload ${workloadId}`);
    }

    return internalUrl;
}

export async function sendInferenceRequest(
    namespace: string,
    url: string,
    token = '',
    expectSuccess = true
): Promise<string> {
    // the command installs and runs curl. curl --fail will return a non-zero exit code on failure
    let curlCommand = `curl --fail ${url}`;
    if (token !== '') {
        curlCommand += ` -H "Authorization: Bearer ${token}"`;
    }

    const command = `/bin/bash -c '${curlCommand}'`;

    const clientWorkloadName = objectGenerator.generateNameByPrefix('inference-client');
    const clientWorkload = new TrainingWorkload(
        config.quickStartImage.networkTools,
        namespace,
        clientWorkloadName
    );
    clientWorkload.command = command;
    clientWorkload.backoffLimit = 0; // we want no retries on failure
    logger.info(
        `creating client workload ${clientWorkloadName} in namespace ${namespace} with command ${command}. expected status is ${expectSuccess ? PodStatus.SUCCEEDED : PodStatus.FAILED}`
    );
    const expectedStatus = expectSuccess ? PodStatus.SUCCEEDED : PodStatus.FAILED;
    await operations.submitWorkloads([clientWorkload], expectedStatus);
    logger.info('client workload completed as expected');
    return clientWorkloadName;
}

export interface GenerateHttpTrafficOptions {
    url: string;
    namespace: string;
    durationSeconds: number;
    concurrentRequests: number;
    qps?: number;
    method?: 'GET' | 'POST' | 'PUT' | 'DELETE';
    headers?: Record<string, string>;
    body?: string;
    urlPath?: string;
}

/**
 * Generate HTTP traffic to a service using hey load testing tool.
 */
export async function generateHttpTraffic(options: GenerateHttpTrafficOptions): Promise<void> {
    const {
        url,
        namespace,
        durationSeconds,
        concurrentRequests,
        qps = 1,
        method = 'GET',
        headers = {},
        body,
        urlPath = ''
    } = options;

    const fullUrl = `${url}${urlPath}`;

    let command = `hey -z ${durationSeconds}s -c ${concurrentRequests} -q ${qps}`;

    if (method !== 'GET') {
        command += ` -m ${method}`;
    }

    for (const [key, value] of Object.entries(headers)) {
        command += ` -H "${key}: ${value}"`;
    }

    if (body) {
        // eslint-disable-next-line quotes
        const escapedBody = body.replace(/'/g, "'\\''");
        command += ` -d '${escapedBody}'`;
    }

    command += ` ${fullUrl}`;

    logger.info(
        `Generating ${concurrentRequests} concurrent ${method} requests for ${durationSeconds}s to ${fullUrl}`
    );

    const clientWorkloadName = objectGenerator.generateNameByPrefix('load-client');
    const clientWorkload = new TrainingWorkload(
        config.quickStartImage.inferenceAutoScaleClient,
        namespace,
        clientWorkloadName
    );
    clientWorkload.command = command;

    logger.info(`Creating load generator workload ${clientWorkloadName} in namespace ${namespace}`);
    logger.debug(`Load generator command: ${command}`);

    await operations.submitWorkloads([clientWorkload], PodStatus.SUCCEEDED);

    logger.info('Load generation completed successfully');
}
