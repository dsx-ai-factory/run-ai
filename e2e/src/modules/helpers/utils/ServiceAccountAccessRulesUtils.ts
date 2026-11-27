/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { IncomingMessage } from 'http';
import { authorizationUtils, flowRootState, logger, serviceAccountUtils } from '../index';
import { AccessRule } from '../../api/apis/authorizationApi/generated/model/accessRule';
import { ServiceAccountPostResponse } from '../../api/apis/identityManagerApi/generated/api';

export async function cleanupAccessRuleAndServiceAccount(
    accessRule: AccessRule | undefined,
    serviceAccount: { response: IncomingMessage; body: ServiceAccountPostResponse } | undefined
): Promise<void> {
    if (accessRule) {
        try {
            await authorizationUtils.deleteAccessRule(accessRule);
        } catch (e) {
            logger.error(`Failed to delete access rule id=${accessRule.id}: ${JSON.stringify(e)}`);
        }
    }
    if (serviceAccount) {
        try {
            await serviceAccountUtils.deleteServiceAccount(
                flowRootState.users.system_administrator.accessToken,
                serviceAccount.body.id
            );
        } catch (e) {
            logger.error(
                `Failed to delete Service Account with id: ${serviceAccount.body.id} ${JSON.stringify(e)}`
            );
        }
    }
}
