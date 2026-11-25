/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

export interface UnleashConfig {
    url: string;
    apiToken: string;
    env: string;
    projectName: string;
    strategyTitle: string;
}

export interface Constrain {
    values: string[];
    inverted: boolean;
    operator: string;
    contextName: string;
    caseInsensitive: boolean;
}

export interface Strategy {
    id: string;
    name: string;
    constraints: Constrain[];
    title: string;
    disabled: boolean;
}

export interface FeatureToggle {
    name: string;
    description: string;
    enabled: boolean;
    strategies: unknown[];
}
