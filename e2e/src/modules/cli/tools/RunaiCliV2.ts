/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BashCli } from './BashCli';
import { config, flowRootState, logger, settings, User, waitUtils } from '../../helpers';
import os from 'os';
import path from 'path';
import { randomUUID } from 'crypto';
import child_process from 'node:child_process';
import { cliExposerApi } from '../../api';
import { ChildProcess, spawn } from 'child_process';
import * as pty from 'node-pty';
import {
    quoteShellArgument,
    writeExecutableAtomically
} from '../../helpers/utils/AtomicFileUtils';
import {
    childProcessOutput,
    cliV2Actions,
    cliV2TrainingType,
    cliV2workloadNoun,
    cliV2WorkloadType,
    collectDiagnosticLogsFlags,
    ConfigFlags,
    DeleteFlagsMerged,
    describeFlags,
    ensureCommandLast,
    execFlags,
    bashFlags,
    ArrayFlags,
    listFlags,
    loginWithUser,
    LogsFlags,
    portForwardFlags,
    ResumeFlagsMerged,
    SubmitFlagsMerged,
    SuspendFlagsMerged,
    workloadLog,
    NimSubmitFlagsMerged,
    NimScaleFlags,
    applicationSubmitFlags
} from './RunaiCliV2.constants';
import { BrowserEngine } from '../../ui/runaiV2/pages/BrowserEngine';
import { CliOidcPage } from '../../ui/runaiV2/pages/CliOidcPage';
import { stringUtil } from '../../ui/utils/stringUtils';
import { AuthPage } from '../../ui/runaiV2/pages/AuthPage';
import { runaiCliV2 } from '../index';
import { UserRole } from '../../helpers/config/roleMapping';

export class RunaiCliV2 {
    cliPath: string;

    installerPath: string;

    bashCli: BashCli;

    constructor(installerPath?: string) {
        const defaultPath = path.join(os.tmpdir(), `runaiCliV2-${process.pid}-${randomUUID()}`);
        this.installerPath = installerPath ? installerPath : defaultPath;
        this.bashCli = new BashCli();
        const runaiDefaultPath = '/.runai/bin/runai';
        this.cliPath = process.env.RUNAI_CLI_V2_PATH || (os.homedir() + runaiDefaultPath);
    }

    runKubectlCommand(command: string): child_process.SpawnSyncReturns<string> {
        const kubectlCommand = `kubectl ${command}`;
        return this.bashCli.execWithOutput(kubectlCommand);
    }

    version(): child_process.SpawnSyncReturns<string> {
        const runaiCommand = `${this.cliPath} version`;
        const cmdOutput = this.bashCli.execWithOutput(runaiCommand);
        logger.debug('successfully executed runai CLI help command');
        return cmdOutput;
    }

    list(
        assetType: string,
        flags?: Partial<typeof listFlags>,
        removeDepMsg = true
    ): child_process.SpawnSyncReturns<string> {
        const FlagsParsed = this.parseFlag(flags);
        const runaiCommand = `${this.cliPath} ${assetType} list ${FlagsParsed}`;
        const cmdOutput = this.bashCli.execWithOutput(runaiCommand, false);
        this.validateCommand(runaiCommand, cmdOutput);
        logger.debug(`successfully executed list ${assetType} runai CLI command`);
        const result = { ...cmdOutput };
        if (removeDepMsg && result.stdout) {
            const jsonStartIndex = result.stdout.indexOf('{');
            if (jsonStartIndex > 0) {
                result.stdout = result.stdout.substring(jsonStartIndex);
            }
        }
        return result;
    }

    myCredentialList(flags?: Partial<typeof listFlags>): child_process.SpawnSyncReturns<string> {
        return this.list('my-credential', flags, false);
    }

    myCredentialCreate(name: string, typeFlag: string, items: string[]): string {
        const itemFlags = items.map((item) => `--item ${item}`).join(' ');
        const runaiCommand = `${this.cliPath} my-credential create ${name} --type ${typeFlag} ${itemFlags}`;
        logger.info(`Executing CLI command: ${runaiCommand}`);
        return this.bashCli.exec(runaiCommand);
    }

    myCredentialDelete(name: string, typeFlag: string, throwOnErr = true): string {
        const runaiCommand = `${this.cliPath} my-credential delete ${name} --type ${typeFlag}`;
        logger.info(`Executing CLI command: ${runaiCommand}`);
        const result = this.bashCli.execWithOutput(runaiCommand, throwOnErr);
        if (throwOnErr) {
            this.validateCommand(runaiCommand, result);
        }
        return result.stdout;
    }

    help(): string {
        const runaiCommand = `${this.cliPath} help`;
        const stdout = this.bashCli.exec(runaiCommand);
        logger.debug('successfully executed runai CLI help command');
        return stdout;
    }

    async workloadLogsFollow(
        workloadName: string,
        workloadType: cliV2WorkloadType | cliV2TrainingType | workloadLog,
        sleepTimeSec: number
    ): Promise<string[]> {
        const cmdNoun = this.parseTrainingType(workloadType);
        const runaiCommand = `${this.cliPath} ${cmdNoun} logs ${workloadName} -f`;
        const childOutput = await this.childProcess(runaiCommand, sleepTimeSec * 1000 + 100);
        childOutput[1].kill();
        return childOutput[0].stdout;
    }

    async portForward(
        flags: Partial<typeof portForwardFlags>,
        workloadName: string,
        workloadType: cliV2WorkloadType | cliV2TrainingType,
        isDistributed?: boolean
    ): Promise<[childProcessOutput, ChildProcess]> {
        const FlagsParsed = this.parseFlag(flags);
        const cmdNoun = this.parseTrainingType(workloadType, isDistributed);
        const runaiCommand = `${this.cliPath} ${cmdNoun} port-forward ${workloadName} ${FlagsParsed}`;
        logger.info(`Executing CLI command: ${runaiCommand}`);
        return await this.childProcess(runaiCommand);
    }

    async childProcess(runaiCommand: string, sleepTime = 3000): Promise<[childProcessOutput, ChildProcess]> {
        const commandParts = runaiCommand.split(' ');
        const command = commandParts.shift();
        const childProcess = spawn(command, commandParts, { env: process.env });

        // Listen for new stdout data
        const childOutput = { stdout: [], stderr: '' };
        childProcess.stdout.on('data', function (data) {
            childOutput.stdout.push(data);
            console.log(`Received data: ${data}`);
        });
        // Listen for error data
        childProcess.stderr.on('data', (data) => {
            childOutput.stderr += data;
            console.error(`Received error: ${data}`);
        });
        // Listen for error event
        childProcess.on('error', (error) => {
            console.error(`Error from child process: ${error}`);
        });
        await waitUtils.sleep(sleepTime);
        return [childOutput, childProcess];
    }

    private logsBase(
        workloadName: string,
        commandBody: string,
        flags: { [k in LogsFlags]?: string }
    ): child_process.SpawnSyncReturns<string> {
        const logsFlagsParsed = this.parseFlag(flags);
        const runaiCommand = `${this.cliPath} ${commandBody} ${logsFlagsParsed}`;
        const cmdOutput = this.bashCli.execWithOutput(runaiCommand);
        logger.debug(
            `successfully executed runai CLI workload logs command for workload ${workloadName} flags: ${logsFlagsParsed}`
        );
        return cmdOutput;
    }

    workspaceLogs(
        workspaceName: string,
        flags: { [k in LogsFlags]?: string }
    ): child_process.SpawnSyncReturns<string> {
        return this.logsBase(workspaceName, `workspace logs ${workspaceName}`, flags);
    }

    inferenceLogs(
        inferenceName: string,
        flags: { [k in LogsFlags]?: string }
    ): child_process.SpawnSyncReturns<string> {
        return this.logsBase(inferenceName, `inference logs ${inferenceName}`, flags);
    }

    trainingLogs(
        trainingName: string,
        flags: { [k in LogsFlags]?: string }
    ): child_process.SpawnSyncReturns<string> {
        return this.logsBase(trainingName, `training logs ${trainingName}`, flags);
    }

    unifyLogs(
        workloadName: string,
        workloadType: cliV2TrainingType | cliV2WorkloadType | workloadLog,
        flags: { [k in LogsFlags]?: string }
    ): child_process.SpawnSyncReturns<string> {
        const cmdNoun = this.parseTrainingType(workloadType);
        return this.logsBase(workloadName, `${cmdNoun} logs ${workloadName}`, flags);
    }

    updateRawFlags(workloadName: string, workloadType: cliV2WorkloadType, flags: string): string {
        return runaiCliV2.cliRaw(workloadType, workloadName, cliV2Actions.update, flags);
    }

    describeRawFlags(workloadName: string, workloadType: cliV2WorkloadType, flags: string): string {
        return runaiCliV2.cliRaw(workloadType, workloadName, cliV2Actions.describe, flags);
    }

    submitRawFlags(workloadName: string, workloadType: cliV2WorkloadType, flags: string): string {
        return runaiCliV2.cliRaw(workloadType, workloadName, cliV2Actions.submit, flags);
    }

    cliRaw(
        workloadType: cliV2WorkloadType,
        workloadName: string,
        action: cliV2Actions,
        actionsFlags: string
    ): string {
        const runaiCommand = `${this.cliPath} ${workloadType} ${action} ${workloadName} ${actionsFlags}`;
        const stdout = this.bashCli.exec(runaiCommand);
        logger.debug(
            `successfully executed runai CLI - type: ${workloadType}, action: ${action}, name: ${workloadName}, flags: ${actionsFlags}`
        );
        return stdout;
    }

    exec(
        flags: Partial<typeof execFlags>,
        workloadName: string,
        workloadType: cliV2WorkloadType | cliV2TrainingType,
        isDistributed?: boolean
    ): child_process.SpawnSyncReturns<string> {
        const cmdNoun = this.parseTrainingType(workloadType, isDistributed);
        const submitFlagsParsed = this.parseFlag(flags);
        const baseCommand = `${this.cliPath} ${cmdNoun} exec ${workloadName} ${submitFlagsParsed}`;
        const stdout = this.bashCli.execWithOutput(baseCommand);
        logger.debug(`successfully executed runai CLI exec ${workloadType} command`);
        return stdout;
    }

    async bashInteractive(
        flags: Partial<typeof bashFlags>,
        workloadName: string,
        workloadType: cliV2WorkloadType | cliV2TrainingType,
        isDistributed?: boolean,
        sleepTime = 2000
    ): Promise<[childProcessOutput, any]> {
        const cmdNoun = this.parseTrainingType(workloadType, isDistributed);
        const bashFlagsParsed = this.parseFlag(flags);
        const runaiCommand = `${this.cliPath} ${cmdNoun} bash ${workloadName} ${bashFlagsParsed}`;
        logger.info(`Executing CLI command (interactive): ${runaiCommand}`);
        return await this.childProcessWithPty(runaiCommand, sleepTime);
    }

    async childProcessWithPty(command: string, sleepTime = 3000): Promise<[childProcessOutput, any]> {
        const childOutput = { stdout: [], stderr: '' };
        const term = pty.spawn('bash', ['-c', command], {
            name: 'xterm-color',
            cols: 80,
            rows: 30,
            cwd: process.cwd(),
            env: process.env
        });
        term.onData((data) => {
            childOutput.stdout.push(data);
            logger.info(`Received data: ${data}`);
        });
        term.onExit(({ exitCode, signal }) => {
            logger.info(`Process exited with code ${exitCode}, signal ${signal}`);
        });
        await waitUtils.sleep(sleepTime);
        const ptyWrapper = {
            stdin: {
                write: (data: string): void => {
                    term.write(data);
                }
            },
            kill: (): void => {
                term.kill();
            },
            _ptyTerm: term
        };
        return [childOutput, ptyWrapper];
    }

    suspend(
        flags: Partial<typeof SuspendFlagsMerged>,
        workloadName: string,
        workloadType: cliV2WorkloadType | cliV2TrainingType
    ): string {
        const suspendFlagsParsed = this.parseFlag(flags);
        const runaiCommand = `${this.cliPath} ${workloadType} suspend ${workloadName} ${suspendFlagsParsed}`;
        const stdout = this.bashCli.exec(runaiCommand);
        logger.debug(`successfully executed runai CLI suspend ${workloadType} command`);
        return stdout;
    }

    resume(
        flags: Partial<typeof ResumeFlagsMerged>,
        workloadName: string,
        workloadType: cliV2WorkloadType | cliV2TrainingType
    ): string {
        const resumeFlagsParsed = this.parseFlag(flags);
        const runaiCommand = `${this.cliPath} ${workloadType} resume ${workloadName} ${resumeFlagsParsed}`;
        const stdout = this.bashCli.exec(runaiCommand);
        logger.debug(`successfully executed runai CLI resume ${workloadType} command`);
        return stdout;
    }

    delete(
        flags: Partial<typeof DeleteFlagsMerged>,
        workloadName: string,
        workloadType: cliV2WorkloadType | cliV2TrainingType | typeof cliV2workloadNoun,
        throwOnErr = true,
        isDistributed = false
    ): string {
        const deleteFlagsParsed = this.parseFlag(flags);
        const cmdNoun = this.parseTrainingType(workloadType, isDistributed);
        const runaiCommand = `${this.cliPath} ${cmdNoun} delete ${workloadName} ${deleteFlagsParsed}`;
        const stdout = this.bashCli.execWithOutput(runaiCommand, throwOnErr).stdout;
        logger.debug(`successfully executed runai CLI delete ${workloadType} command`);
        return stdout;
    }

    projectSet(projectName: string): string {
        const runaiCommand = `${this.cliPath} project set ${projectName}`;
        const stdout = this.bashCli.exec(runaiCommand);
        logger.debug(`successfully set project to ${projectName}, using runai CLI command`);
        return stdout;
    }

    clusterSet(clusterName: string): string {
        const runaiCommand = `${this.cliPath} cluster set ${clusterName}`;
        const stdout = this.bashCli.exec(runaiCommand);
        logger.debug(`successfully set cluster to ${clusterName}, using runai CLI command`);
        return stdout;
    }

    kubeconfigSet(): string {
        const runaiCommand = `${this.cliPath} kubeconfig set`;
        const stdout = this.bashCli.exec(runaiCommand);
        logger.debug('successfully executed runai CLI kubeconfig set command');
        return stdout;
    }

    templateDescribe(templateName: string, flags?: Partial<typeof describeFlags>): string {
        const FlagsParsed = this.parseFlag(flags);
        const runaiCommand = `${this.cliPath} template describe ${templateName} ${FlagsParsed}`;
        const stdout = this.bashCli.exec(runaiCommand);
        return stdout;
    }

    datasourceDescribe(datasourceName: string, type: string, flags?: Partial<typeof describeFlags>): string {
        const FlagsParsed = this.parseFlag(flags);
        const runaiCommand = `${this.cliPath} datasource describe ${datasourceName} --type ${type} ${FlagsParsed}`;
        const stdout = this.bashCli.exec(runaiCommand);
        logger.debug(
            `successfully executed datasource describe command for ${datasourceName} (type: ${type})`
        );
        return stdout;
    }

    datasourceHelp(subcommand: string = 'help'): string {
        const runaiCommand = `${this.cliPath} datasource ${subcommand}`;
        const stdout = this.bashCli.exec(runaiCommand);
        logger.debug(`successfully executed datasource help command with: ${subcommand}`);
        return stdout;
    }

    computeDescribe(computeName: string, flags?: Partial<typeof describeFlags>): string {
        const FlagsParsed = this.parseFlag(flags);
        const runaiCommand = `${this.cliPath} compute describe ${computeName} ${FlagsParsed}`;
        logger.debug(`Executing command: ${runaiCommand}`);
        const stdout = this.bashCli.exec(runaiCommand);
        logger.debug(`successfully executed compute describe command for ${computeName}`);
        return stdout;
    }

    computeHelp(subcommand: string = 'help'): string {
        const runaiCommand = `${this.cliPath} compute ${subcommand}`;
        const stdout = this.bashCli.exec(runaiCommand);
        logger.debug(`successfully executed compute help command with: ${subcommand}`);
        return stdout;
    }

    environmentHelp(subcommand: string = 'help'): string {
        const runaiCommand = `${this.cliPath} environment ${subcommand}`;
        const stdout = this.bashCli.exec(runaiCommand);
        logger.debug(`successfully executed environment help command with: ${subcommand}`);
        return stdout;
    }

    environmentDescribe(environmentName: string, flags?: Partial<typeof describeFlags>): string {
        const FlagsParsed = this.parseFlag(flags);
        const runaiCommand = `${this.cliPath} environment describe ${environmentName} ${FlagsParsed}`;
        const stdout = this.bashCli.exec(runaiCommand);
        logger.debug(`successfully executed environment describe command for ${environmentName}`);
        return stdout;
    }

    applicationDescribe(appName: string, flags?: Partial<typeof describeFlags>): string {
        const FlagsParsed = this.parseFlag(flags);
        const runaiCommand = `${this.cliPath} application describe ${appName} ${FlagsParsed}`;
        const result = this.bashCli.execWithOutput(runaiCommand);
        logger.debug(`successfully executed application describe command for ${appName}`);
        return result.stdout;
    }

    applicationSubmit(appName: string, flags: Partial<typeof applicationSubmitFlags>): string {
        const FlagsParsed = this.parseFlag(flags);
        const runaiCommand = `${this.cliPath} application submit ${appName} ${FlagsParsed}`;
        const result = this.bashCli.execWithOutput(runaiCommand);
        logger.debug(`successfully executed application submit command for ${appName}`);
        return result.stdout;
    }

    collectDiagnosticLogs(flags: { [k in collectDiagnosticLogsFlags]?: string }): string {
        const parsedFlags = this.parseFlag(flags);
        const runaiCommand = `${this.cliPath} diagnostics collect-logs ${parsedFlags}`;
        const stdout = this.bashCli.exec(runaiCommand);
        logger.debug(
            `successfully executed runai CLI diagnostics collect-logs command with flags: ${parsedFlags}`
        );
        return stdout;
    }

    loginWithAccessKey(
        clientId: string,
        secret: string,
        throwOnErr = true
    ): child_process.SpawnSyncReturns<string> {
        const loginCommand = `${this.cliPath} login access-key --client-id ${clientId} --secret ${secret} --interactive=disabled`;
        const cmdOutput = this.bashCli.execWithOutput(loginCommand, throwOnErr);
        logger.debug(`successfully logged in to runai CLI, with client-id ${clientId}`);
        return cmdOutput;
    }

    loginAsUser(
        username: string,
        password: string,
        throwOnErr = true
    ): child_process.SpawnSyncReturns<string> {
        const loginCommand = `${this.cliPath} login user --user ${username} --password ${password}`;
        const cmdOutput = this.bashCli.execWithOutput(loginCommand, throwOnErr);
        logger.debug(`successfully logged in to runai CLI, with username ${username}`);
        return cmdOutput;
    }

    logout(throwOnErr = true): child_process.SpawnSyncReturns<string> {
        const loginCommand = `${this.cliPath} logout`;
        const cmdOutput = this.bashCli.execWithOutput(loginCommand, throwOnErr);
        logger.debug('successfully logged out.');
        return cmdOutput;
    }

    private async executeLoginCommand(
        command: string,
        browserEngine: BrowserEngine,
        processFirstData?: (data: string, childProcess: ChildProcess) => Promise<void>
    ): Promise<{ stdout: any[]; stderr: string }> {
        const commandParts = command.split(' ');
        const mainCommand = commandParts.shift();
        const childOutput = { stdout: [], stderr: '' };

        return new Promise((resolve) => {
            const childProcess = spawn(mainCommand, commandParts, { env: process.env });
            let dataCounter = 0;

            childProcess.stdout.on('data', async (data) => {
                console.log(`Data from child process: ${data}`);
                childOutput.stdout.push(data.toString());

                if (dataCounter === 0 && processFirstData) {
                    await processFirstData(data.toString(), childProcess);
                    dataCounter++;
                }
            });

            childProcess.stderr.on('data', (data) => {
                console.error(`Received error: ${data}`);
                childOutput.stderr += data.toString();
            });

            childProcess.on('error', (error) => {
                console.error(`Error from child process: ${error}`);
            });

            childProcess.on('close', (code) => {
                console.log(`Child process exited with code ${code}`);
                resolve(childOutput);
            });

            // Fallback timeout
            setTimeout(() => {
                childProcess.kill();
                resolve(childOutput);
            }, 20000);
        });
    }

    async loginBrowser(browserEngine: BrowserEngine, user: User): Promise<{ stdout: any[]; stderr: string }> {
        const loginCommand = `${this.cliPath} login --print-only`;
        return this.executeLoginCommand(loginCommand, browserEngine, async (data) => {
            const auth = new AuthPage(browserEngine.currentPage);
            try {
                const navUrl = stringUtil.extractUrl(data);
                await browserEngine.page.goto(navUrl);
                await auth.signInOnly(user.email, user.password);
            }
            catch (error) {
                console.error(error);
            }
        });
    }

    async loginSSO(
        browserEngine: BrowserEngine,
        role: UserRole,
        token?: string
    ): Promise<{ stdout: any[]; stderr: string }> {
        const loginCommand = `${this.cliPath} login sso`;
        return this.executeLoginCommand(loginCommand, browserEngine, async (data, childProcess) => {
            const navUrl = stringUtil.extractUrl(data);
            await browserEngine.page.goto(navUrl);
            const auth = new AuthPage(browserEngine.currentPage);
            await auth.signInOnly(flowRootState.users[role].email, flowRootState.users[role].password);
            const cliOidcPage = new CliOidcPage(browserEngine.currentPage);
            await waitUtils.waitForElementToExist(cliOidcPage, cliOidcPage.codeBlock);
            token = token ? token : (await cliOidcPage.getCodeBlockContent())?.trim();
            console.log(token);
            childProcess.stdin.write(`${token}\n`);
        });
    }

    async loginSSOv2(
        browserEngine: BrowserEngine,
        user: User,
        token?: string
    ): Promise<{ stdout: any[]; stderr: string }> {
        const loginCommand = `${this.cliPath} login sso`;
        return this.executeLoginCommand(loginCommand, browserEngine, async (data, childProcess) => {
            const navUrl = stringUtil.extractUrl(data);
            await browserEngine.page.goto(navUrl);
            const auth = new AuthPage(browserEngine.currentPage);
            await auth.signInOnly(user.email, user.password);
            const cliOidcPage = new CliOidcPage(browserEngine.currentPage);
            await waitUtils.waitForElementToExist(cliOidcPage, cliOidcPage.codeBlock);
            token = token ? token : (await cliOidcPage.getCodeBlockContent())?.trim();
            console.log(token);
            childProcess.stdin.write(`${token}\n`);
        });
    }

    configSet(flag: ConfigFlags, value: string): string {
        const runaiCommand = `${this.cliPath} config set ${flag.toString()} ${value}`;
        const cmdOutput = this.bashCli.execWithOutput(runaiCommand, false);
        this.validateCommand(runaiCommand, cmdOutput);
        logger.debug(`successfully configured runai CLI, ${flag.toString()} = ${value}`);
        return cmdOutput.stdout;
    }

    async downloadAndInstall(
        accessToken: string = flowRootState.users.system_administrator.accessToken
    ): Promise<void> {
        const runaiDefaultPath = '/.runai/bin/runai';
        this.cliPath = os.homedir() + runaiDefaultPath;
        try {
            loginWithUser(flowRootState.users.system_administrator);
            const clusterList = this.list('cluster').stdout;
            if (clusterList.includes(flowRootState.cluster.name)) {
                runaiCliV2.clusterSet(flowRootState.cluster.name);
                flowRootState.cliVersion = this.version().stdout?.trim() || 'unknown';
                return;
            }
        } catch (e) {
            logger.debug(e instanceof Error ? e.message : String(e));
        }
        await this.download(accessToken);
        this.upgrade();
        const configUrl = settings.isSelfHosted
            ? config.appUrl
            : stringUtil.extractUrl(await this.getInstallCommand()).split('api')[0];
        this.configSet(ConfigFlags.setControlPlane, configUrl);
        this.configSet(ConfigFlags.setAuthServer, configUrl);
        loginWithUser(flowRootState.users.system_administrator);
        runaiCliV2.clusterSet(flowRootState.cluster.name);
        flowRootState.cliVersion = this.version().stdout?.trim() || 'unknown';
        logger.debug('successfully installed runai CLI');
    }

    async download(
        accessToken: string = flowRootState.users.system_administrator.accessToken
    ): Promise<void> {
        const installCommand = await this.getInstallCommand(accessToken);
        await writeExecutableAtomically(this.installerPath, (temporaryPath) => {
            const curlCmd = `${installCommand} > ${quoteShellArgument(temporaryPath)}`;
            this.bashCli.exec(curlCmd);
        });
        logger.debug('successfully downloaded runai CLI');
    }

    async getInstallCommand(
        accessToken: string = flowRootState.users.system_administrator.accessToken
    ): Promise<any> {
        try {
            return (await cliExposerApi.cliApi.getInstallerUnixCommands(undefined, accessToken)).body.command
                .split('(')[1]
                .split(')')[0];
        } catch {
            throw new Error('Could not get install command');
        }
    }

    upgrade(): void {
        const command = `${this.installerPath} -u -s`;
        const cmdOutput = this.bashCli.execWithOutput(command, false);
        this.validateCommand(command, cmdOutput);
        logger.debug('successfully upgraded runai CLI');
    }

    submit(
        flags: Partial<typeof SubmitFlagsMerged>,
        workloadName: string,
        workloadType: cliV2WorkloadType,
        trainingType?: cliV2TrainingType,
        isDistributed?: boolean,
        user?: User
    ): string {
        if (flags[SubmitFlagsMerged.attach] !== undefined) {
            throw new Error('Attach flag is only valid in submitAttach function');
        }

        if (user) {
            this.loginAsUser(user.email, user.password);
        }

        const submitCmd = this.getSubmitCommand(
            flags,
            workloadName,
            workloadType,
            trainingType,
            isDistributed
        );
        logger.info(`Executing CLI command: ${submitCmd}`);
        const stdout = this.bashCli.exec(submitCmd);
        logger.debug(`successfully executed runai CLI submit ${workloadType} command`);
        return stdout;
    }

    describe(
        flags: Partial<typeof describeFlags>,
        workloadName: string,
        workloadType: cliV2WorkloadType,
        isDistributed?: boolean
    ): string {
        const FlagsParsed = this.parseFlag(flags);
        const cmdNoun = this.parseTrainingType(workloadType, isDistributed);
        const runaiCommand = `${this.cliPath} ${cmdNoun} describe ${workloadName} ${FlagsParsed}`;
        logger.info(`Executing CLI command: ${runaiCommand}`);
        const stdout = this.bashCli.exec(runaiCommand);
        logger.debug(`successfully executed runai CLI describe ${workloadType} command`);
        return stdout;
    }

    async describeAsync(
        flags: Partial<typeof describeFlags>,
        workloadName: string,
        workloadType: cliV2WorkloadType
    ): Promise<string> {
        const FlagsParsed = this.parseFlag(flags);
        const cmdNoun = this.parseTrainingType(workloadType);
        const runaiCommand = `${this.cliPath} ${cmdNoun} describe ${workloadName} ${FlagsParsed}`;
        const stdout = await this.bashCli.asyncExec(runaiCommand);
        logger.debug(`successfully executed runai CLI describe ${workloadType} command`);
        return stdout;
    }

    getSubmitCommand(
        flags: Partial<typeof SubmitFlagsMerged>,
        workloadName: string,
        workloadType: cliV2WorkloadType,
        trainingType?: cliV2TrainingType,
        isDistributed?: boolean
    ): string {
        if (workloadType === cliV2WorkloadType.workspace && trainingType) {
            throw new Error('you cant set trainingType when submitting workspace workload');
        }

        // Handle distributed inference
        if (isDistributed && workloadType === cliV2WorkloadType.inference) {
            flags = ensureCommandLast(flags);
            const submitFlagsParsed = this.parseFlag(flags);
            return `${this.cliPath} inference distributed submit ${workloadName} ${submitFlagsParsed}`;
        }

        const trainingTypeParsed = trainingType ? trainingType : '';
        flags = ensureCommandLast(flags);
        const submitFlagsParsed = this.parseFlag(flags);
        return `${this.cliPath} ${workloadType} ${trainingTypeParsed} submit ${workloadName} ${submitFlagsParsed}`;
    }

    submitAttach(
        flags: Partial<typeof SubmitFlagsMerged>,
        workloadName: string,
        workloadType: cliV2WorkloadType,
        attachInput: string,
        trainingType?: cliV2TrainingType,
        isDistributed?: boolean
    ): Promise<string> {
        return new Promise((resolve, reject) => {
            let resolved = false;
            let output = '';
            const expectedPrompt = "If you don't see a command prompt, try pressing enter."; // eslint-disable-line quotes
            let promptDetected = false;
            let inputSent = false;
            const submitCmd = this.getSubmitCommand(
                flags,
                workloadName,
                workloadType,
                trainingType,
                isDistributed
            );

            console.log(`Starting command: ${submitCmd}`);

            // Use node-pty to spawn a pseudo-terminal
            const term = pty.spawn('bash', ['-c', submitCmd], {
                name: 'xterm-color',
                cols: 80,
                rows: 30,
                cwd: process.cwd(),
                env: process.env
            });

            // Set a max timeout
            const maxTimeout = setTimeout(() => {
                if (!resolved) {
                    term.kill();
                    reject(new Error('Timed out waiting for the prompt message'));
                    resolved = true;
                }
            }, 60000); // 1 minute timeout

            // Handle data from the terminal
            term.onData((data) => {
                output += data;
                console.log(`Received: ${data.trim()}`);

                // Check if we've detected the prompt message
                if (!promptDetected && data.includes(expectedPrompt)) {
                    console.log('Prompt message detected.');
                    promptDetected = true;

                    // Wait a short time before sending the command
                    setTimeout(() => {
                        if (!inputSent) {
                            console.log(`Sending input: ${attachInput}`);
                            term.write(attachInput + '\r');
                            inputSent = true;

                            // Wait for response, then resolve
                            setTimeout(() => {
                                if (!resolved) {
                                    resolved = true;
                                    clearTimeout(maxTimeout);
                                    term.kill();
                                    resolve(output);
                                }
                            }, 5000); // Wait 5 seconds for response
                        }
                    }, 500);
                }
            });

            // Handle terminal exit
            term.onExit(({ exitCode }) => {
                clearTimeout(maxTimeout);
                console.log(`Process exited with code ${exitCode}`);

                if (!resolved) {
                    resolved = true;
                    if (promptDetected) {
                        // If we saw the prompt, consider it a success
                        resolve(output);
                    } else if (exitCode !== 0) {
                        reject(new Error(`Process failed with exit code ${exitCode}`));
                    } else {
                        resolve(output);
                    }
                }
            });
        });
    }

    private parseFlag(flags: { [k: string]: string }): string {
        let flagsParsed = '';
        for (const key in flags) {
            if (ArrayFlags.includes(key) && Array.isArray(flags[key])) {
                for (const value of flags[key]) {
                    flagsParsed += `${key} ${value} `;
                }
            } else {
                flagsParsed += `${key} ${flags[key]} `;
            }
        }
        return flagsParsed.slice(0, -1);
    }

    private validateCommand(command: string, cmdOutput: child_process.SpawnSyncReturns<string>): void {
        if (cmdOutput.status !== 0 && cmdOutput.status !== null) {
            throw new Error(
                `run command ${cmdOutput} failed with code ${cmdOutput.status}, stdout: ${cmdOutput.stdout}, stderr: ${cmdOutput.stderr}`
            );
        }
    }

    private parseTrainingType(
        workloadType: cliV2WorkloadType | cliV2TrainingType | typeof cliV2workloadNoun,
        isDistributed?: boolean
    ): string | cliV2WorkloadType | cliV2TrainingType {
        if (isDistributed && workloadType === cliV2WorkloadType.inference) {
            return 'inference distributed';
        }

        if (workloadType === cliV2WorkloadType.inferenceNim) {
            return 'inference nim';
        }

        return workloadType === cliV2TrainingType.standard
            ? `training ${cliV2TrainingType.standard}`
            : workloadType;
    }

    async waitForWorkloadPhaseWithSampling(
        workloadName: string,
        projectName: string,
        expectedPhase: string,
        workloadType: cliV2WorkloadType,
        timeoutSec = 60,
        invalidPhase?: string
    ): Promise<void> {
        return new Promise((resolve, reject) => {
            const startTime = Date.now();
            const timeoutMs = timeoutSec * 1000;
            const samplingIntervalMs = 500; // 0.5 seconds
            let lastPhase: string | undefined = undefined;
            let sampleCount = 0;
            let resolved = false;

            logger.debug(
                `[waitForWorkloadPhaseWithSampling] Starting concurrent sampling for workload=${workloadName}, expected=${expectedPhase}, invalidPhase=${invalidPhase}, timeout=${timeoutSec}s, interval=0.5s`
            );

            // Use recursive setTimeout instead of setInterval for better control
            const scheduleNextSample = (): void => {
                if (resolved) return;

                // Check timeout
                if (Date.now() - startTime >= timeoutMs) {
                    if (!resolved) {
                        resolved = true;
                        const elapsedSec = Math.round((Date.now() - startTime) / 1000);
                        logger.error(
                            `[waitForWorkloadPhaseWithSampling] TIMEOUT after ${elapsedSec}s and ${sampleCount} samples! workload=${workloadName}, expected=${expectedPhase}, lastPhase=${lastPhase}`
                        );
                        reject(
                            new Error(
                                `[waitForWorkloadPhaseWithSampling] Phase missing: [${expectedPhase}] after ${sampleCount} samples over ${elapsedSec}s. Last observed phase: ${lastPhase}`
                            )
                        );
                    }
                    return;
                }

                sampleCount++;
                const currentSampleId = sampleCount;

                // Execute sample asynchronously
                setTimeout(async () => {
                    if (resolved) return;

                    try {
                        const flags = { [describeFlags.output]: 'yaml', '-p': projectName } as any;
                        const desc = await this.describeAsync(flags, workloadName, workloadType);
                        const match = desc.match(/^[ \t]*phase:[ \t]+(.+)$/im);
                        const phase = match ? match[1].trim() : undefined;
                        lastPhase = phase;

                        const elapsedSec = Math.round((Date.now() - startTime) / 1000);
                        logger.debug(
                            `[waitForWorkloadPhaseWithSampling] Sample #${currentSampleId} (${elapsedSec}s): workload=${workloadName}, current=${phase}, expected=${expectedPhase}`
                        );

                        // Check for invalid phase first - immediate failure
                        if (invalidPhase && phase === invalidPhase) {
                            if (!resolved) {
                                resolved = true;
                                logger.error(
                                    `[waitForWorkloadPhaseWithSampling] Invalid phase detected at sample #${currentSampleId}: ${phase} (after ${elapsedSec}s)`
                                );
                                reject(new Error(`Invalid phase detected: ${phase}`));
                            }
                            return;
                        }

                        // Check for expected phase - immediate success
                        if (phase === expectedPhase) {
                            if (!resolved) {
                                resolved = true;
                                logger.debug(
                                    `[waitForWorkloadPhaseWithSampling] SUCCESS at sample #${currentSampleId}: workload=${workloadName} reached expected phase=${expectedPhase} (after ${elapsedSec}s)`
                                );
                                resolve();
                            }
                            return;
                        }
                    } catch (err) {
                        if (resolved) return;

                        // Handle describe command failures
                        if (err.message && err.message.includes('Invalid phase detected')) {
                            if (!resolved) {
                                resolved = true;
                                reject(err);
                            }
                            return;
                        }

                        logger.debug(
                            `[waitForWorkloadPhaseWithSampling] Sample #${currentSampleId} failed: ${err.message}`
                        );
                        if (err.message && err.message.includes('workload not found')) {
                            // Continue sampling if workload not found yet
                        } else {
                            // For other errors, continue but log them
                            logger.warn(
                                `[waitForWorkloadPhaseWithSampling] Describe command error: ${err.message}`
                            );
                        }
                    }
                }, 0);

                // Schedule next sample
                setTimeout(scheduleNextSample, samplingIntervalMs);
            };

            // Start the sampling process
            scheduleNextSample();
        });
    }

    scaleNim(
        workloadName: string,
        flags: Partial<typeof NimScaleFlags>
    ): string {
        const flagsParsed = this.parseFlag(flags);
        const runaiCommand = `${this.cliPath} inference nim scale ${workloadName} ${flagsParsed}`;
        logger.info(`Executing CLI command: ${runaiCommand}`);
        const stdout = this.bashCli.exec(runaiCommand);
        logger.debug('successfully executed runai CLI scale nim command');
        return stdout;
    }

    updateNim(
        workloadName: string,
        flags: Partial<typeof NimSubmitFlagsMerged>
    ): string {
        const flagsParsed = this.parseFlag(flags);
        const runaiCommand = `${this.cliPath} inference nim update ${workloadName} ${flagsParsed}`;
        logger.info(`Executing CLI command: ${runaiCommand}`);
        const stdout = this.bashCli.exec(runaiCommand);
        logger.debug('successfully executed runai CLI update nim command');
        return stdout;
    }
}
