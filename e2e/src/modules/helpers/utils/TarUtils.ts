/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import * as tar from 'tar';

export interface TarEntry {
    path: string;
    size: number;
}

export async function getTarEntries(tarFilePath: string): Promise<TarEntry[]> {
    const entries: TarEntry[] = [];

    await tar.t({
        file: tarFilePath,
        onentry: (entry) => {
            entries.push({ path: entry.path, size: entry.size });
        }
    });

    return entries;
}
