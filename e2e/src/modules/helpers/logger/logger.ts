/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { settings } from '..';
import { createLogger, format, transports } from 'winston';

const { combine, splat, timestamp, printf, colorize } = format;

// Safe stringify to avoid circular structure errors in logs
function safeStringify(value: unknown): string {
    const seen = new WeakSet();
    try {
        return JSON.stringify(value, (_key, val) => {
            if (typeof val === 'object' && val !== null) {
                if (seen.has(val as object)) {
                    return '[Circular]';
                }
                seen.add(val as object);
            }
            if (val instanceof Error) {
                return {
                    name: val.name,
                    message: val.message,
                    stack: val.stack
                };
            }
            return val;
        });
    } catch (_e) {
        return '[Unserializable]';
    }
}

const textFormat = printf(({ level, message, timestamp, ...metadata }) => {
    let metadataLogString = '';
    if (metadata && Object.keys(metadata).length > 0) {
        metadataLogString = safeStringify(metadata);
    }
    const msg = `${timestamp} [${level}] : ${message} ${metadataLogString}`;
    return msg;
});

const defaultFormat = combine(
    timestamp({
        format: 'YYYY-MM-DD HH:mm:ss'
    }),
    textFormat,
    splat()
);
export const logger = createLogger({
    level: settings.logLevel,
    transports: [
        new transports.File({
            filename: 'logs/testLogger.log',
            format: defaultFormat,
            handleExceptions: true
        }),
        new transports.Console({
            format: combine(colorize(), defaultFormat),
            silent: !settings.enableConsoleLogs,
            handleExceptions: true
        })
    ]
});
