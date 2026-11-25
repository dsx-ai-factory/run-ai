/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { logger, UtilsError, CliError, supportedPlatforms } from '..';
import axios from 'axios';
import fs from 'fs';
import { v4 as uuidv4 } from 'uuid';
import { Page } from 'playwright';
import https from 'https';
import { pipeline } from 'stream/promises';
import { config } from '../config/config';
import { writeExecutableAtomically } from './AtomicFileUtils';

const customAgent = new https.Agent({ keepAlive: true, maxSockets: 1 });

const axiosInstance = axios.create({
    httpAgent: customAgent,
    httpsAgent: customAgent
});

export class Utils {
    static strStartsWith = (str: string, prefix: string): boolean => str.indexOf(prefix) === 0;

    static trimLastChar(str: string, charToTrim: string): string {
        if (str && str.charAt(str.length - 1) === charToTrim) {
            return str.substring(0, str.length - 1);
        }
        return str;
    }

    static randomInteger(min, max: number): number {
        return Math.floor(Math.random() * (max - min + 1)) + min;
    }

    static downloadFile = async (fileUrl: string, filePath: string): Promise<string> => {
        logger.debug(`downloading file from url ${fileUrl}`);
        try {
            const fullFilePath = await writeExecutableAtomically(filePath, async (temporaryPath) => {
                const response = await axiosInstance
                    .get(fileUrl, { responseType: 'stream' })
                    .catch((error) => {
                        throw new UtilsError(error);
                    });
                await pipeline(response.data, fs.createWriteStream(temporaryPath));
            });
            logger.debug(filePath + ' execution permissions added');
            return fullFilePath;
        } catch (error) {
            logger.error('unable to download file:  error: ' + error);
            throw new CliError(error);
        }
    };

    static async delay(time: number): Promise<void> {
        // Use this function only if it is your last options
        return new Promise((resolve) => setTimeout(resolve, time));
    }

    static async delayUI(page: Page, time: number): Promise<void> {
        await page.waitForTimeout(time);
    }

    static removeHttpProtocolPrefix(url: string): string {
        return url.split('://')[1];
    }

    static async delayByEnv(platform: string): Promise<void> {
        if (platform === supportedPlatforms.ocp) {
            await Utils.delay(70000);
        } else {
            await Utils.delay(90000);
        }
    }

    static namespaceFromProjectName(projectName: string): string {
        return 'runai-' + projectName;
    }

    static generateShortUuid(): string {
        return uuidv4().split('-').pop();
    }

    static spaceToDash(str: string): string {
        return str.replace(/[ ]/g, '-');
    }

    static extractSaasTenantName(): string | null {
        const regex = /^https:\/\/([^.]+)\./;
        const match = process.env.CONTROL_PLANE_URL.match(regex);
        return match ? match[1] : null;
    }

    static isDGXC(): boolean {
        return config.tenantUrl.includes('dgxc');
    }

    static getRandomElements<T>(arr: T[], m: number): T[] {
        if (m > arr.length) throw new Error('m should not be greater than array length!');
        const shuffled = [...arr];

        for (let i = shuffled.length - 1; i > 0; i--) {
            const j = Math.floor(Math.random() * (i + 1));
            [shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]];
        }
        return shuffled.slice(0, m);
    }

    /**
     * Sometimes, expected event is not triggered after clicking a button.
     * This can happen due to rendering delays or because the event handler has not been attached to the button.
     * this function is used to handle the situation above to make tests more robust.
     * @param fn
     * @param options
     * @returns
     */
    static async retry<T>(fn: () => Promise<T>, options = { retries: 3, delay: 500 }): Promise<T> {
        let lastError;

        for (let i = 0; i < options.retries; i++) {
            try {
                const result = await fn();
                if (i > 0) {
                    logger.warn(`retry succeeded after ${i} failed attempt(s)`);
                }
                return result;
            } catch (e) {
                lastError = e;
                logger.debug(
                    `retry attempt ${i + 1}/${options.retries} failed: ${(e as Error).message?.split('\n')[0]}`
                );
                if (i < options.retries - 1) {
                    await new Promise((res) => setTimeout(res, options.delay));
                }
            }
        }
        throw lastError;
    }

    static readCSVFile(filePath: string): Array<Record<string, string>> {
        try {
            const normalizeNewlines = (str: string): string => str.replace(/\r\n?/g, '\n');

            const parseCsvLine = (line: string): string[] => {
                const values: string[] = [];
                let current = '';
                let inQuotes = false;

                for (let i = 0; i < line.length; i++) {
                    const char = line[i];
                    if (char === '"') {
                        // handle escaped quotes inside quoted fields
                        if (inQuotes && i + 1 < line.length && line[i + 1] === '"') {
                            current += '"';
                            i++; // skip the escaped quote
                        } else {
                            inQuotes = !inQuotes;
                        }
                    } else if (char === ',' && !inQuotes) {
                        values.push(current.trim());
                        current = '';
                    } else {
                        current += char;
                    }
                }
                values.push(current.trim());
                return values;
            };

            const content = normalizeNewlines(fs.readFileSync(filePath, 'utf8'))
                .split('\n')
                .filter((line) => line.trim());
            if (content.length < 1) return [];

            const headers = parseCsvLine(content[0]).map((h) => h.trim());
            // if there is only one row, return column names
            if (content.length === 1) {
                const columnTemplate = headers.reduce((obj, header) => ({ ...obj, [header]: '' }), {});
                return [columnTemplate];
            }

            return content.slice(1).map((line) => {
                const values = parseCsvLine(line);
                return headers.reduce((obj, header, i) => ({ ...obj, [header]: values[i] ?? '' }), {});
            });
        } catch (error) {
            console.error('Failed to read file:', error);
            return [];
        }
    }

    /**
     * Processes an array of items in parallel with limited concurrency.
     */
    static async parallelMap<TItem, TResult>(
        concurrency: number,
        items: TItem[],
        taskFn: (item: TItem, index: number) => Promise<TResult>
    ): Promise<TResult[]> {
        const results: TResult[] = new Array(items.length);

        const queue = items.map((_, i) => i);

        const worker = async (): Promise<void> => {
            while (queue.length > 0) {
                const i = queue.shift();
                if (i === undefined) break;

                try {
                    results[i] = await taskFn(items[i], i);
                } catch (err) {
                    console.error(`Error processing item at index ${i}`, err);
                    throw err;
                }
            }
        };

        // Spin up the workers
        await Promise.all(Array.from({ length: concurrency }).map(worker));

        return results;
    }
}
