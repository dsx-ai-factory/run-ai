/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import * as yaml from 'js-yaml';

const runaiRemoteWriteName = 'runai';

const isObject = (value: unknown): value is Record<string, unknown> => {
    return typeof value === 'object' && value !== null && !Array.isArray(value);
};

const parseMonitoringConfig = (configYaml: string): Record<string, unknown> => {
    const parsed = yaml.load(configYaml);
    return isObject(parsed) ? parsed : {};
};

export function getRunaiRemoteWrites(configYaml: string): Record<string, unknown>[] {
    const prometheusK8s = parseMonitoringConfig(configYaml).prometheusK8s;
    if (!isObject(prometheusK8s) || !Array.isArray(prometheusK8s.remoteWrite)) {
        return [];
    }

    return prometheusK8s.remoteWrite.filter(
        (remoteWrite): remoteWrite is Record<string, unknown> =>
            isObject(remoteWrite) && remoteWrite.name === runaiRemoteWriteName
    );
}

export function monitoringConfigWithoutRunaiRemoteWrites(configYaml: string): Record<string, unknown> {
    const config = parseMonitoringConfig(configYaml);
    const prometheusK8s = config.prometheusK8s;

    if (!isObject(prometheusK8s) || !Array.isArray(prometheusK8s.remoteWrite)) {
        return config;
    }

    const remainingRemoteWrites = prometheusK8s.remoteWrite.filter(
        (remoteWrite) => !isObject(remoteWrite) || remoteWrite.name !== runaiRemoteWriteName
    );

    if (remainingRemoteWrites.length === 0) {
        delete prometheusK8s.remoteWrite;
    } else {
        prometheusK8s.remoteWrite = remainingRemoteWrites;
    }

    if (Object.keys(prometheusK8s).length === 0) {
        delete config.prometheusK8s;
    }

    return config;
}
