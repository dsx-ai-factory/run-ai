/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import http from 'http';
import { auditApi } from '../../api';
import { AuditLog, GetAuditLogs200Response } from '../../api/apis/auditApi';
import { waitUtils } from '../index';
import { grantTypeMetadataKey, loginAuditAction } from '../config/const';

export class AuditLogsUtils {
    async getAuditLogs(
        accessToken: string,
        s: Date,
        e: Date,
        filters: string[]
    ): Promise<{ response: http.IncomingMessage; body: GetAuditLogs200Response }> {
        const { response, body } = await auditApi.auditLogsApi.getAuditLogs(
            s,
            e,
            undefined, // page
            undefined, // limit
            undefined, // sortOrder
            undefined, // sortBy
            filters,
            {
                headers: {
                    Authorization: 'Bearer ' + accessToken
                }
            }
        );

        return { response, body };
    }

    async downloadAuditLogs(
        accessToken: string,
        startTime: Date,
        endTime: Date,
        fileType?: 'CSV' | 'JSON',
        numberOfSamples?: number,
        filterBy?: string[]
    ): Promise<{
        response: http.IncomingMessage;
        body?: any;
    }> {
        return await auditApi.auditLogsApi.downloadAuditLogs(
            startTime,
            endTime,
            fileType,
            undefined,
            numberOfSamples,
            filterBy,
            {
                headers: {
                    Authorization: 'Bearer ' + accessToken
                }
            }
        );
    }

    async findLoginAudit(
        accessToken: string,
        subject: string,
        grantType: string,
        result: AuditLog.ResultEnum = AuditLog.ResultEnum.Succeeded
    ): Promise<AuditLog> {
        const start = new Date(Date.now() - 1000 * 60 * 10);
        const end = new Date(Date.now() + 1000 * 60);
        const filters = [`subject==${subject}`];

        const matches = (log: AuditLog): boolean =>
            log.action === loginAuditAction &&
            log.result === result &&
            log.metadata?.[grantTypeMetadataKey] === grantType;

        let found: AuditLog | undefined;
        await waitUtils.waitForCondition(
            async () => {
                const { auditLogs } = (await this.getAuditLogs(accessToken, start, end, filters)).body;
                found = auditLogs?.find(matches);
                return found !== undefined;
            },
            `waiting for ${grantType} login audit for subject ${subject}`,
            true,
            60,
            5
        );
        return found;
    }
}
