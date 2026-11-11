/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import yaml from 'js-yaml';
import fs from 'fs';
import { assert } from 'chai';
import { config, kubeConfigUtils, waitUtils } from '../index';

export class KubeConfigUtils {
    private validateParseKubeConfigToYaml(config: string): Record<string, any> {
        let kubeConfigYaml: Record<string, any>;
        try {
            kubeConfigYaml = yaml.load(config) as Record<string, any>;
        } catch (error) {
            reportError(`Error parsing kubeConfig to YAML:\n${error.message}`);
        }
        return kubeConfigYaml;
    }

    private parseKubeConfigToString(config: Record<string, any>): string {
        let kubeConfig: string;
        try {
            kubeConfig = yaml.dump(config);
        } catch (error) {
            reportError(`Error parsing kubeConfig to string:\n${error.message}`);
        }
        return kubeConfig;
    }

    addUser(kubeConfig: string, name = 'runai-authenticated-user'): string {
        const kubeConfigYaml = this.validateParseKubeConfigToYaml(kubeConfig);
        const authFlowValue = process.env.ENABLE_AUTH === 'true' ? 'cli' : 'browser';
        const user = {
            name: name,
            user: {
                'auth-provider': {
                    config: {
                        airgapped: 'true',
                        'auth-flow': authFlowValue,
                        realm: 'runai',
                        'idp-issuer-url': `${process.env.APP_URL}/auth/realms/runai`,
                        'client-id': 'runai-cli'
                    },
                    name: 'oidc'
                }
            }
        };
        kubeConfigYaml.users.push(user);
        return this.parseKubeConfigToString(kubeConfigYaml);
    }

    setUserToken(
        kubeConfig: string,
        idToken: string,
        refreshToken: string,
        name = 'runai-authenticated-user'
    ): string {
        const kubeConfigYaml = this.validateParseKubeConfigToYaml(kubeConfig);
        const users = kubeConfigYaml.users;
        const runaiUserIndex = users.findIndex((user) => user['name'] === name);
        assert(runaiUserIndex !== -1, 'The user name you provided is not part of your kubeConfig');
        const runaiUserConfig = users[runaiUserIndex]['user']['auth-provider']['config'];
        runaiUserConfig['id-token'] = idToken;
        runaiUserConfig['refresh-token'] = refreshToken;
        return this.parseKubeConfigToString(kubeConfigYaml);
    }

    addContext(kubeConfig: string, contextName: string, userName = 'runai-authenticated-user'): string {
        const kubeConfigYaml = this.validateParseKubeConfigToYaml(kubeConfig);
        const cliLoginContext = {
            context: {
                cluster: process.env.CLUSTER_NAME_PREFIX,
                user: userName
            },
            name: contextName
        };
        kubeConfigYaml.contexts.push(cliLoginContext);
        return this.parseKubeConfigToString(kubeConfigYaml);
    }

    setContext(kubeConfig: string, contextName: string): string {
        const kubeConfigYaml = this.validateParseKubeConfigToYaml(kubeConfig);
        kubeConfigYaml['current-context'] = contextName;
        return this.parseKubeConfigToString(kubeConfigYaml);
    }

    async setContextAndSave(contextName: string): Promise<void> {
        const kubeConfig = await kubeConfigUtils.getKubeConfig();
        const kubeConfigEdited = this.setContext(kubeConfig, contextName);
        await this.setKubeConfig(kubeConfigEdited);
    }

    async getKubeConfig(): Promise<string> {
        for (let i = 0; i < 10; i++) {
            try {
                const kubeConfig = fs.readFileSync(config.kubeconfigPath, 'utf8');
                return kubeConfig;
            } catch (error) {
                console.log(`Error getting kubeConfig:\n${error.message}`);
                await waitUtils.sleep(2000);
            }
        }
        throw new Error('unable to get kubeConfig');
    }

    //Make sure you are setting the kube config to default value, after you finish altering it for your test.
    async setKubeConfig(kubeConfig: string): Promise<void> {
        for (let i = 0; i < 10; i++) {
            try {
                fs.writeFileSync(config.kubeconfigPath, kubeConfig, 'utf8');
                return;
            } catch (error) {
                console.log(`Error saving kubeConfig:\n${error.message}`);
                await waitUtils.sleep(2000);
            }
        }
        throw new Error('unable to save kubeConfig');
    }
}
