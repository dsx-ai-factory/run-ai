/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import crypto from 'crypto';
import fs from 'fs';

export class Md5FileHasher {
    static async getFileHash(filePath: string, algorithm: string): Promise<string> {
        return new Promise((resolve, reject) => {
            const hash = crypto.createHash(algorithm);
            const stream = fs.createReadStream(filePath);

            stream.on('data', (data) => {
                hash.update(data);
            });

            stream.on('end', () => {
                const result = hash.digest('hex');
                resolve(result);
            });

            stream.on('error', (err) => {
                reject(err);
            });
        });
    }
}
