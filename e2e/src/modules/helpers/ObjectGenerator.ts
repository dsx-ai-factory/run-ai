/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { flowRootState, logger } from '.';
import { Utils } from './utils/utils';
import { randomBytes, randomInt } from 'crypto';

export class ObjectGenerator {
    private _generateRandomString = (length: number): string => {
        const characters = 'abcdefghijklmnopqrstuvwxyz';
        let result = '';
        for (let i = 0; i < length; i++) {
            const randomIndex = Math.floor(Math.random() * characters.length);
            result += characters[randomIndex];
        }
        return result;
    };

    // Generates a random password that always satisfies the complexity
    // requirements (one upper, one lower, one digit, one special, >= 12 chars).
    generatePassword = (): string => {
        const upper = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ'[randomInt(26)];
        const lower = 'abcdefghijklmnopqrstuvwxyz'[randomInt(26)];
        const digit = '0123456789'[randomInt(10)];
        const special = '!@#$%^&*'[randomInt(8)];
        const filler = randomBytes(8).toString('base64url').slice(0, 8);
        return `${upper}${lower}${digit}${special}${filler}`;
    };

    generateNameByPrefix = (prefix: string): string => {
        const generated = `${prefix}-${flowRootState.currentFlowUuid}-${Utils.generateShortUuid().slice(
            0,
            6
        )}`;
        logger.debug(`Generating name by prefix: ${prefix}, generated: ${generated}`);
        return generated;
    };

    generateShortNameByPrefix = (prefix: string, uuidLength = 4): string => {
        return `${prefix}-${Utils.generateShortUuid().slice(0, uuidLength)}`;
    };

    generateEmailByPrefix = (prefix: string, clusterSuffix?: string): string => {
        logger.debug(`Generating email by prefix: ${prefix}`);
        const suffix = clusterSuffix
            ? clusterSuffix + flowRootState.currentFlowUuid
            : flowRootState.currentFlowUuid;
        return `${prefix}-${suffix}@example.com`;
    };

    generateEmailByPrefixWithRandom = (
        prefix: string,
        uuidLength?: number,
        clusterSuffix?: string
    ): string => {
        const randomPrefix = this.generateShortNameByPrefix(prefix, uuidLength);
        logger.debug(`Generating email by prefix: ${randomPrefix}`);
        const suffix = clusterSuffix
            ? clusterSuffix + flowRootState.currentFlowUuid
            : flowRootState.currentFlowUuid;
        return `${randomPrefix}-${suffix}@example.com`;
    };

    generateLettersOnlyByPrefix = (prefix: string): string => {
        return `${prefix}${this._generateRandomString(6)}`;
    };
}
