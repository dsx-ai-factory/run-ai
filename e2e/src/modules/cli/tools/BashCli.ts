/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import * as spawn from 'cross-spawn';
import { config, logger, settings } from '../../helpers';
import fs from 'fs';
import * as awaitSpawn from 'await-spawn';
import * as child_process from 'node:child_process';

export class BashCli {
    commandTimeoutMS: number;

    constructor() {
        this.commandTimeoutMS = settings.taskTimeoutMS;
    }

    private separateCommandAndArgs(commandInput: string): Array<string | Array<string>> {
        const cmdArgs = commandInput.split(' ');
        if (cmdArgs.length > 0) {
            const cmd = cmdArgs.splice(0, 1)[0];
            return [cmd, cmdArgs];
        }
    }

    execFile(scriptData: string): string {
        const scriptPath = '/tmp/runai-cli-script.sh';
        fs.writeFileSync(scriptPath, scriptData);
        return this.exec(`bash ${scriptPath}`);
    }

    execWithOutput(commandInput: string, throwOnErr = true): child_process.SpawnSyncReturns<string> {
        const [cmd, cmdArgs] = this.separateCommandAndArgs(commandInput);
        logger.debug(`run command ${commandInput}`);
        logger.debug(`command ${cmd}, with args ${cmdArgs}`);
        const output = spawn.sync(cmd, cmdArgs, {
            timeout: settings.taskTimeoutMS,
            encoding: 'utf8',
            shell: true,
            env: {
                ...process.env,
                KUBECONFIG: config.kubeconfigPath
            }
        });
        if (!throwOnErr || (output.status === 0 && output.output)) return output;
        throw new Error(
            `run command ${commandInput} failed with code ${output.status}, stdout: ${output.stdout}, stderr: ${output.stderr}`
        );
    }

    exec(commandInput: string): string {
        const output = this.execWithOutput(commandInput);
        if (output.status === 0 && output.output) return output.stdout.toString();
        throw new Error(
            `run command ${commandInput} failed with code ${output.status}, stdout: ${output.stdout}, stderr: ${output.stderr}`
        );
    }

    async asyncExec(commandInput: string): Promise<string> {
        const [cmd, cmdArgs] = this.separateCommandAndArgs(commandInput);
        logger.debug(`run command ${commandInput}`);
        logger.debug(`command ${cmd}, with args ${cmdArgs}`);

        try {
            const output = await awaitSpawn.default(cmd, cmdArgs);
            if (output) {
                return output.toString();
            } else {
                throw new Error(`Command ${commandInput} returned empty output.`);
            }
        } catch (error) {
            if (error && error.code && error.stderr) {
                throw new Error(
                    `run command ${commandInput} failed with code ${error.code}, stderr: ${error.stderr}`
                );
            } else {
                throw new Error(`Failed to run command ${commandInput}: ${error.message}`);
            }
        }
    }
}
