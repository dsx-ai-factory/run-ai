/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from '@playwright/test';
import { type TarEntry } from '../../../../../../modules/helpers/utils/TarUtils';

export const CLI_OUTPUT_LOG_FILE_NAME = 'cli-output.log';
export const POD_OWNER_PATH_SEGMENTS = 3;
export const NON_POD_OWNER_PATH_SEGMENTS = 4;

export const ALLOWED_NAMESPACES = [
    'runai',
    'runai-reservation',
    'runai-backend',
    'training-operator',
    'knative',
    'gpu-operator',
    'nim-operator'
];

export const ALLOWED_K8S_RESOURCE_KIND_FOLDER_NAMINGS = [
    'daemon-set',
    'replica-set',
    'deployment',
    'stateful-set',
    'job',
    'cron-job'
];

export function extractTarTopLevelFolders(tarEntries: TarEntry[]): string[] {
    const folders = new Set<string>();

    tarEntries.forEach((entry) => {
        if (entry.path.includes('/')) {
            folders.add(entry.path.split('/')[0]);
        }
    });

    return Array.from(folders);
}

export function isResourceKindFolderNamingValid(resourceKind: string): boolean {
    for (const allowedResourceKind of ALLOWED_K8S_RESOURCE_KIND_FOLDER_NAMINGS) {
        if (resourceKind.startsWith(allowedResourceKind)) {
            return true;
        }
    }
    return false;
}

export function assertLogsStructure(tarEntries: TarEntry[]): void {
    const allowedNamespacesSet = new Set(ALLOWED_NAMESPACES);

    tarEntries.forEach((entry) => {
        if (!entry.path.endsWith('.log') || entry.path === CLI_OUTPUT_LOG_FILE_NAME) {
            return;
        }

        const pathSegments = entry.path.split('/');
        expect([POD_OWNER_PATH_SEGMENTS, NON_POD_OWNER_PATH_SEGMENTS]).toContain(pathSegments.length);

        if (pathSegments.length === POD_OWNER_PATH_SEGMENTS) {
            const [namespace, podName, logFileName] = pathSegments;
            expect(allowedNamespacesSet.has(namespace)).toBe(true);
            expect(podName).toBeDefined();
            expect(logFileName).toBeDefined();
        }

        const [namespace, resourceKind, podName, logFileName] = pathSegments;
        expect(allowedNamespacesSet.has(namespace)).toBe(true);
        expect(isResourceKindFolderNamingValid(resourceKind)).toBe(true);
        expect(podName).toBeDefined();
        expect(logFileName).toBeDefined();
    });
}

export function assertNoPreviousLogs(tarEntries: TarEntry[]): void {
    tarEntries.forEach((entry) => {
        if (entry.path.endsWith('.log')) {
            expect(entry.path).not.toContain('/previous_'); // '/' is the path separator, and 'previous_' is file name prefix
        }
    });
}
