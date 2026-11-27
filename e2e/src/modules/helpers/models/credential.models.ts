/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

export enum ECredentialSource {
    Private = 'private', // "My credentials"
    Shared = 'shared' // "Shared secret"
}

export enum ECredentialType {
    NgcApiKey = 'ngc-api-key',
    GenericSecret = 'generic-secret',
    DockerRegistry = 'docker-registry'
}

export const CREDENTIAL_SOURCE_LABELS = {
    [ECredentialSource.Private]: 'My credentials',
    [ECredentialSource.Shared]: 'Shared secret'
} as const;

export const CREDENTIAL_TYPE_LABELS = {
    [ECredentialType.NgcApiKey]: 'NGC API key',
    [ECredentialType.GenericSecret]: 'Generic secret',
    [ECredentialType.DockerRegistry]: 'Docker registry'
} as const;
