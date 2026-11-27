/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { kubeCli } from '../../cli';
import { config, settings } from '../config/config';
import { logger } from '../logger/logger';
import { namespaceCP, helmRepoName } from '../config/const';
import { HelmReleaseUtils, resolveHelmSecretRef, SecretRef } from './HelmReleaseUtils';

const helmReleaseUtils = new HelmReleaseUtils();

const TENANTS_MANAGER_CONFIGMAP = 'runai-backend-tenants-manager';
const GEN_TENANTS_MANAGER_SECRET = 'runai-backend-gen-tenants-manager';

const CONSOLE_ADMIN_USERNAME = 'admin@run.ai';

function decodeSecretValue(valueB64?: string): string | undefined {
    if (!valueB64) {
        return undefined;
    }
    return Buffer.from(valueB64, 'base64').toString('utf-8');
}

async function readSecretValue(secretName: string, key: string): Promise<string | undefined> {
    try {
        const secret = await kubeCli.getSecret(secretName, namespaceCP);
        const value = decodeSecretValue(secret.data?.[key]);
        if (value) {
            logger.debug(`Fetched credential value from ${secretName}/${key}`);
        }
        return value;
    } catch (error) {
        logger.warn(`Failed to fetch secret ${secretName} from cluster: ${error.message}`);
        return undefined;
    }
}

function usesConsoleTenantPasswordFromEnv(): boolean {
    if (settings.isMultiTenancyCluster) {
        return true;
    }
    return process.env.MULTI_TENANCY_ENABLED === 'true';
}

async function usesConsoleTenantPassword(): Promise<boolean> {
    if (usesConsoleTenantPasswordFromEnv()) {
        return true;
    }

    try {
        const cm = await kubeCli.getConfigMap(TENANTS_MANAGER_CONFIGMAP, namespaceCP);
        if (cm.data?.TENANT_MNG_NAME) {
            return true;
        }
        if (cm.data?.ADMIN_USERNAME) {
            return false;
        }
    } catch (error) {
        logger.debug(`Could not read ${TENANTS_MANAGER_CONFIGMAP} for MT detection: ${error.message}`);
    }

    return false;
}

function resolvePasswordRefFromHelm(values: Record<string, any>, consoleTenantMode: boolean): SecretRef {
    if (consoleTenantMode) {
        return resolveHelmSecretRef(
            values.global?.management,
            GEN_TENANTS_MANAGER_SECRET,
            'TENANT_MNG_PASSWORD'
        );
    }
    return resolveHelmSecretRef(
        values.tenantsManager?.config,
        GEN_TENANTS_MANAGER_SECRET,
        'ADMIN_PASSWORD'
    );
}

async function resolveUsernameFromHelm(values: Record<string, any>, consoleTenantMode: boolean): Promise<string> {
    if (!consoleTenantMode) {
        try {
            const cm = await kubeCli.getConfigMap(TENANTS_MANAGER_CONFIGMAP, namespaceCP);
            if (cm.data?.ADMIN_USERNAME) {
                return cm.data.ADMIN_USERNAME;
            }
        } catch {
            // fall through to the env-supplied username
        }
        if (!config.controlPlaneAdminUsername) {
            throw new Error(
                `No admin username found in ${TENANTS_MANAGER_CONFIGMAP} (namespace ${namespaceCP}) ` +
                    'and no CONTROL_PLANE_ADMIN_USERNAME was provided.'
            );
        }
        return config.controlPlaneAdminUsername;
    }

    const userSecretName = values.global?.management?.existingSecret || GEN_TENANTS_MANAGER_SECRET;
    const username = await readSecretValue(userSecretName, 'TENANT_MNG_USER');
    return username || CONSOLE_ADMIN_USERNAME;
}

/** Resolves CP admin credentials from helm values and cluster secrets. */
let credentialsResolved = false;

export async function resolveControlPlaneCredentials(): Promise<void> {
    if (!settings.isSelfHosted) {
        throw new Error(
            'Control plane credentials are resolved from cluster secrets; only self-hosted environments are supported'
        );
    }

    const values = helmReleaseUtils.helmGetValues(helmRepoName, namespaceCP);
    const consoleTenantMode = await usesConsoleTenantPassword();
    const passwordRef = resolvePasswordRefFromHelm(values, consoleTenantMode);
    const password = await readSecretValue(passwordRef.secretName, passwordRef.key);
    if (password) {
        config.controlPlaneAdminPassword = password;
    } else if (config.controlPlaneAdminPassword) {
        // Released CP versions may name the secret differently (e.g. without the
        // "gen-" prefix), so the cluster scrape can miss it. Fall back to the
        // operator-supplied CONTROL_PLANE_ADMIN_PASSWORD instead of failing the run.
        logger.warn(
            `Admin password not found in ${passwordRef.secretName}/${passwordRef.key} (namespace ${namespaceCP}); ` +
                'falling back to the supplied CONTROL_PLANE_ADMIN_PASSWORD.'
        );
    } else {
        throw new Error(
            `No control plane admin password found in ${passwordRef.secretName}/${passwordRef.key} (namespace ${namespaceCP}) ` +
                'and no CONTROL_PLANE_ADMIN_PASSWORD was provided.'
        );
    }

    config.controlPlaneAdminUsername = await resolveUsernameFromHelm(values, consoleTenantMode);

    // Propagate to env so Playwright workers (spawned after globalSetup) inherit
    // the resolved creds via config.ts; the in-memory config object set here does
    // not cross process boundaries.
    // Use RESOLVED_ vars to avoid overwriting the original CI-supplied env vars
    // (e.g. CONTROL_PLANE_ADMIN_PASSWORD is read at config load time).
    process.env.RESOLVED_CONTROL_PLANE_ADMIN_PASSWORD = config.controlPlaneAdminPassword;
    process.env.RESOLVED_CONTROL_PLANE_ADMIN_USERNAME = config.controlPlaneAdminUsername;

    logger.debug(
        `Resolved control plane admin credentials from ${passwordRef.secretName}/${passwordRef.key} for ${config.controlPlaneAdminUsername}`
    );
}

let cachedKeycloakPassword: string | null = null;

export async function resolveKeycloakPassword(): Promise<string> {
    if (cachedKeycloakPassword) {
        return cachedKeycloakPassword;
    }
    const values = helmReleaseUtils.helmGetValues(helmRepoName, namespaceCP);
    const { secretName, key } = resolveHelmSecretRef(
        values.keycloakx,
        'runai-backend-gen-keycloakx',
        'KEYCLOAK_ADMIN_PASSWORD'
    );
    const password = await readSecretValue(secretName, key);
    if (!password) {
        throw new Error(`No Keycloak admin password found in ${secretName}/${key}`);
    }
    cachedKeycloakPassword = password;
    return password;
}

/** Idempotent wrapper — safe to call from hooks, AuthUtils, and setup entry points. */
export async function ensureControlPlaneCredentials(): Promise<void> {
    if (!settings.isSelfHosted) {
        return;
    }
    if (credentialsResolved && config.controlPlaneAdminPassword) {
        return;
    }
    // Worker processes inherit RESOLVED_CONTROL_PLANE_ADMIN_PASSWORD from globalSetup
    // via env inheritance. Accept it directly instead of re-running helm+kubectl in each worker.
    if (process.env.RESOLVED_CONTROL_PLANE_ADMIN_PASSWORD) {
        config.controlPlaneAdminPassword = process.env.RESOLVED_CONTROL_PLANE_ADMIN_PASSWORD;
        config.controlPlaneAdminUsername =
            process.env.RESOLVED_CONTROL_PLANE_ADMIN_USERNAME || config.controlPlaneAdminUsername;
        credentialsResolved = true;
        return;
    }
    await resolveControlPlaneCredentials();
    credentialsResolved = true;
}
