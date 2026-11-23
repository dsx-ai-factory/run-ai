/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import fs from 'fs';
import path from 'path';

type AtomicFileWriter = (temporaryPath: string) => Promise<void> | void;

export function quoteShellArgument(argument: string): string {
    const singleQuote = String.fromCharCode(39);
    const escapedSingleQuote = `${singleQuote}\\${singleQuote}${singleQuote}`;
    return `${singleQuote}${argument.split(singleQuote).join(escapedSingleQuote)}${singleQuote}`;
}

/**
 * Writes an executable away from its destination and publishes it only after
 * the writer has completed, the file is durable, and all local handles close.
 */
export async function writeExecutableAtomically(
    destinationPath: string,
    writer: AtomicFileWriter
): Promise<string> {
    const resolvedDestinationPath = path.resolve(destinationPath);
    const destinationDirectory = path.dirname(resolvedDestinationPath);
    const temporaryDirectory = await fs.promises.mkdtemp(
        path.join(destinationDirectory, `.${path.basename(resolvedDestinationPath)}-`)
    );
    const temporaryPath = path.join(temporaryDirectory, 'download');

    try {
        await writer(temporaryPath);
        await fs.promises.chmod(temporaryPath, 0o755);

        const fileHandle = await fs.promises.open(temporaryPath, 'r');
        try {
            await fileHandle.sync();
        } finally {
            await fileHandle.close();
        }

        await fs.promises.rename(temporaryPath, resolvedDestinationPath);
        return resolvedDestinationPath;
    } finally {
        await fs.promises.rm(temporaryDirectory, { recursive: true, force: true });
    }
}
