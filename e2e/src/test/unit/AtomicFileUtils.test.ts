/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect, test } from '@playwright/test';
import { spawnSync } from 'child_process';
import fs from 'fs';
import os from 'os';
import path from 'path';
import { quoteShellArgument, writeExecutableAtomically } from '../../modules/helpers/utils/AtomicFileUtils';

test.describe('writeExecutableAtomically', () => {
    let testDirectory: string;
    let destinationPath: string;

    test.beforeEach(() => {
        testDirectory = fs.mkdtempSync(path.join(os.tmpdir(), 'atomic-executable-test-'));
        destinationPath = path.join(testDirectory, 'installer');
    });

    test.afterEach(() => {
        fs.rmSync(testDirectory, { recursive: true, force: true });
    });

    test('keeps the current executable available while its replacement is open', async () => {
        await writeExecutableAtomically(destinationPath, async (temporaryPath) => {
            await fs.promises.writeFile(temporaryPath, '#!/bin/sh\nprintf old');
        });

        let markWriterReady: () => void;
        const writerReady = new Promise<void>((resolve) => {
            markWriterReady = resolve;
        });
        let releaseWriter: () => void;
        const writerRelease = new Promise<void>((resolve) => {
            releaseWriter = resolve;
        });

        const replacement = writeExecutableAtomically(destinationPath, async (temporaryPath) => {
            const fileHandle = await fs.promises.open(temporaryPath, 'w');
            try {
                await fileHandle.writeFile('#!/bin/sh\n');
                markWriterReady();
                await writerRelease;
                await fileHandle.writeFile('printf new');
            } finally {
                await fileHandle.close();
            }
        });

        await writerReady;
        try {
            const outputDuringWrite = spawnSync(destinationPath, { encoding: 'utf8' });
            expect(outputDuringWrite.status).toBe(0);
            expect(outputDuringWrite.stdout).toBe('old');
        } finally {
            releaseWriter();
        }

        await replacement;
        const outputAfterPublish = spawnSync(destinationPath, { encoding: 'utf8' });
        expect(outputAfterPublish.status).toBe(0);
        expect(outputAfterPublish.stdout).toBe('new');
    });

    test('uses distinct temporary paths for concurrent publishers', async () => {
        const temporaryPaths: string[] = [];
        const scripts = ['#!/bin/sh\nprintf first', '#!/bin/sh\nprintf second'];

        await Promise.all(
            scripts.map((script) =>
                writeExecutableAtomically(destinationPath, async (temporaryPath) => {
                    temporaryPaths.push(temporaryPath);
                    await fs.promises.writeFile(temporaryPath, script);
                })
            )
        );

        expect(new Set(temporaryPaths).size).toBe(scripts.length);
        const output = spawnSync(destinationPath, { encoding: 'utf8' });
        expect(output.status).toBe(0);
        expect(['first', 'second']).toContain(output.stdout);
    });

    test('preserves the current executable when a replacement fails', async () => {
        await writeExecutableAtomically(destinationPath, async (temporaryPath) => {
            await fs.promises.writeFile(temporaryPath, '#!/bin/sh\nprintf current');
        });

        await expect(
            writeExecutableAtomically(destinationPath, async (temporaryPath) => {
                await fs.promises.writeFile(temporaryPath, '#!/bin/sh\nprintf incomplete');
                throw new Error('download failed');
            })
        ).rejects.toThrow('download failed');

        const output = spawnSync(destinationPath, { encoding: 'utf8' });
        expect(output.status).toBe(0);
        expect(output.stdout).toBe('current');
    });
});

test.describe('quoteShellArgument', () => {
    test('preserves spaces, quotes, and shell expansion characters', () => {
        const argument = `/tmp/installer${String.fromCharCode(39)}s $file`;
        const output = spawnSync('/bin/sh', ['-c', `printf %s ${quoteShellArgument(argument)}`], {
            encoding: 'utf8'
        });

        expect(output.status).toBe(0);
        expect(output.stdout).toBe(argument);
    });
});
