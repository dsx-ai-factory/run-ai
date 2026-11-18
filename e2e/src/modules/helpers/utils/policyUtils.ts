/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    CustomResourceDefinition,
    flowRootState,
    logger,
    policyUtils,
    versionUtils,
    waitUtils
} from '../index';
import { bashCli, kubeCli } from '../../cli';
import {
    DistributedInferencePolicyChangeRequestV2,
    DistributedInferencePolicyOverwriteRequestV2,
    DistributedInferencePolicyV2,
    DistributedPolicyChangeRequestV2,
    DistributedPolicyOverwriteRequestV2,
    DistributedPolicyV2,
    InferencePolicyChangeRequestV2,
    InferencePolicyOverwriteRequestV2,
    InferencePolicyV2,
    MatchExpressionOperator,
    ScopeType,
    TrainingPolicyChangeRequestV2,
    TrainingPolicyOverwriteRequestV2,
    TrainingPolicyV2,
    WorkspacePolicyChangeRequestV2,
    WorkspacePolicyOverwriteRequestV2,
    WorkspacePolicyV2
} from '../../api/apis/policyApi';
import { policyApi as policyApiHandler } from '../../api';
import * as yaml from 'js-yaml';
import { NodeSelector, PolicyDefaultsType, PolicyRuleType } from '../config/types';
import { hostnameLabel } from '../config/const';
import http from 'http';
import { StatusCodes } from 'http-status-codes';
import { expect } from '@playwright/test';
import _ from 'lodash';

export enum WorkloadType {
    Training = 'trainingpolicies.run.ai',
    Interactive = 'interactivepolicies.run.ai',
    Distributed = 'distributedpolicies.run.ai',
    Inference = 'inferencepolicies.run.ai',
    DistributedInference = 'distributedinferencepolicies.run.ai'
}

export class PolicyUtils {
    async waitForPolicyToBeSynced(
        workloadType: WorkloadType,
        namespace: string,
        policyName: string
    ): Promise<void> {
        await waitUtils.waitForCondition(
            () => {
                try {
                    /*
                    const policyYaml = bashCli.exec(
                        `kubectl get ${workloadType} -n ${namespace} ${policyName} -oyaml`
                    );
                    return policyYaml.includes('syncStatus');
                    */
                    return true; // OFER: we are no longer syncing policies to the cluster,
                    // if all tests are ok we can remove this function
                } catch (e) {
                    return false;
                }
            },
            'wait for policy to be synced (should not happen)',
            true,
            90
        );
    }

    async createPolicy(
        isInteractive: boolean,
        name: string,
        namespace: string,
        scopeId: number,
        kind: WorkloadType,
        defaults: object,
        rules?: object,
        scope = ScopeType.Project
    ): Promise<void> {
        const policyRequest = {
            meta: {
                scope: scope,
                name: name
            },
            policy: {
                defaults: defaults,
                rules: rules
            }
        };
        logger.info(`policy payload: ${JSON.stringify(policyRequest)}`);
        if (scope === ScopeType.Project) {
            policyRequest['meta']['projectId'] = scopeId;
        } else if (scope === ScopeType.Department) {
            policyRequest['meta']['departmentId'] = scopeId.toString();
        }
        if (isInteractive) {
            await policyApiHandler.policyApi.overwriteWorkspacePolicyV2(
                false,
                false,
                policyRequest as WorkspacePolicyOverwriteRequestV2
            );
        } else {
            await policyApiHandler.policyApi.overwriteTrainingPolicyV2(
                false,
                false,
                policyRequest as TrainingPolicyOverwriteRequestV2
            );
        }
        const policyScope = scope === ScopeType.Project ? 'project' : 'department';
        await policyUtils.waitForPolicyToBeSynced(kind, namespace, `${policyScope}-${scopeId}`);
    }

    /**
     * create policy V2
     * @param scopeName project name | department name | cluster name
     * @param namespace for 'cluster' and 'department' scoped policy, it's 'runai', for 'project' scoped policy it's `runai-projectName`
     * @param scopeId scope ID
     * @param kind type in WorkloadType
     * @param defaults policy's 'default' section
     * @param rules policy's 'rules' section
     * @param scopeType Project | Department | Cluster
     */
    async createPolicyV2(
        scopeName: string,
        scopeId: string,
        kind: WorkloadType,
        defaults: object,
        rules?: object,
        scopeType = ScopeType.Project,
        validationOnly: boolean = false
    ): Promise<{
        response: http.IncomingMessage;
        body:
            | TrainingPolicyV2
            | DistributedPolicyV2
            | WorkspacePolicyV2
            | InferencePolicyV2
            | DistributedInferencePolicyV2;
    }> {
        const policyRequest = {
            meta: {
                scope: scopeType,
                name: scopeName
            },
            policy: {
                defaults: defaults,
                rules: rules
            }
        };
        logger.info(`policy payload: ${JSON.stringify(policyRequest)}`);

        const key = this.getPolicyScopeKeyName(scopeType);
        policyRequest['meta'][key] = scopeType === ScopeType.Project ? parseInt(scopeId) : scopeId;

        policyApiHandler.policyApi.accessToken = flowRootState.users.system_administrator.accessToken;
        switch (kind) {
            case WorkloadType.Training:
                return await policyApiHandler.policyApi.overwriteTrainingPolicyV2(
                    validationOnly,
                    false,
                    policyRequest as TrainingPolicyOverwriteRequestV2
                );
            case WorkloadType.Distributed:
                return await policyApiHandler.policyApi.overwriteDistributedPolicyV2(
                    validationOnly,
                    false,
                    policyRequest as DistributedPolicyOverwriteRequestV2
                );
            case WorkloadType.Interactive:
                return await policyApiHandler.policyApi.overwriteWorkspacePolicyV2(
                    validationOnly,
                    false,
                    policyRequest as WorkspacePolicyOverwriteRequestV2
                );
            case WorkloadType.Inference:
                return await policyApiHandler.policyApi.overwriteInferencePolicyV2(
                    validationOnly,
                    false,
                    policyRequest as InferencePolicyOverwriteRequestV2
                );
            case WorkloadType.DistributedInference:
                return await policyApiHandler.policyApi.overwriteDistributedInferencePolicyV2(
                    validationOnly,
                    false,
                    policyRequest as DistributedInferencePolicyOverwriteRequestV2
                );
            default:
                throw new Error(`Unknown WorkloadType: ${kind}!`);
        }
    }

    async getPolicyV2(
        kind: WorkloadType,
        scopeInfo: { scope: string; scopeId: string } = {
            scope: 'cluster',
            scopeId: flowRootState.cluster.uuid
        }
    ): Promise<{
        response: http.IncomingMessage;
        body:
            | TrainingPolicyV2
            | DistributedPolicyV2
            | WorkspacePolicyV2
            | InferencePolicyV2
            | DistributedInferencePolicyV2;
    }> {
        policyApiHandler.policyApi.accessToken = flowRootState.users.system_administrator.accessToken;

        const getPolicyMap = {
            [WorkloadType.Training]: policyApiHandler.policyApi.getTrainingPolicyV2,
            [WorkloadType.Distributed]: policyApiHandler.policyApi.getDistributedPolicyV2,
            [WorkloadType.Interactive]: policyApiHandler.policyApi.getWorkspacePolicyV2,
            [WorkloadType.Inference]: policyApiHandler.policyApi.getInferencePolicyV2,
            [WorkloadType.DistributedInference]: policyApiHandler.policyApi.getDistributedInferencePolicyV2
        };

        const getFn = getPolicyMap[kind];

        if (!getFn) {
            throw new Error(`Unknown WorkloadType: ${kind}!`);
        }

        const fnArgs = this.getPolicyArgs(scopeInfo);
        const resp = await getFn.call(policyApiHandler.policyApi, ...fnArgs);

        return resp;
    }

    getPolicyName(scopeType: ScopeType, scopeId: string): string {
        const policyNameMap: { [key in ScopeType]: string } = {
            [ScopeType.Project]: `project-${scopeId}`,
            [ScopeType.Department]: `department-${scopeId}`,
            [ScopeType.Cluster]: 'cluster'
        };

        const policyName = policyNameMap[scopeType];
        if (!policyName) {
            throw new Error(`Unknown scopeType: ${scopeType}!`);
        }
        return policyName;
    }

    getPolicyScopeKeyName(scopeType: ScopeType): string {
        const scopeKeyMap = {
            [ScopeType.Project]: 'projectId',
            [ScopeType.Department]: 'departmentId',
            [ScopeType.Cluster]: 'clusterId'
        };

        const keyName = scopeKeyMap[scopeType];
        if (!keyName) {
            throw new Error(`Unknown scopeType: ${scopeType}!`);
        }
        return keyName;
    }

    async deleteWorkloadPolicyV2(
        kind: WorkloadType,
        scopeInfo: { scope: string; scopeId: string } = {
            scope: 'cluster',
            scopeId: flowRootState.cluster.uuid
        }
    ): Promise<number> {
        policyApiHandler.policyApi.accessToken = flowRootState.users.system_administrator.accessToken;

        const deletePolicyMap = {
            [WorkloadType.Training]: policyApiHandler.policyApi.deleteTrainingPolicy,
            [WorkloadType.Distributed]: policyApiHandler.policyApi.deleteDistributedPolicy,
            [WorkloadType.Interactive]: policyApiHandler.policyApi.deleteWorkspacePolicy,
            [WorkloadType.Inference]: policyApiHandler.policyApi.deleteInferencePolicy,
            [WorkloadType.DistributedInference]: policyApiHandler.policyApi.deleteDistributedInferencePolicy
        };

        const getPolicyMap = {
            [WorkloadType.Training]: policyApiHandler.policyApi.getTrainingPolicyV2,
            [WorkloadType.Distributed]: policyApiHandler.policyApi.getDistributedPolicyV2,
            [WorkloadType.Interactive]: policyApiHandler.policyApi.getWorkspacePolicyV2,
            [WorkloadType.Inference]: policyApiHandler.policyApi.getInferencePolicyV2,
            [WorkloadType.DistributedInference]: policyApiHandler.policyApi.getDistributedInferencePolicyV2
        };

        const deleteFn = deletePolicyMap[kind];
        const getFn = getPolicyMap[kind];

        if (!deleteFn || !getFn) {
            throw new Error(`Unknown WorkloadType: ${kind}!`);
        }

        const fnArgs = this.getPolicyArgs(scopeInfo);

        const resposne = await deleteFn.call(policyApiHandler.policyApi, ...fnArgs);

        await waitUtils.waitForCondition(
            async () => {
                const resp = await getFn.call(policyApiHandler.policyApi, ...fnArgs);

                return !resp.body.policy.defaults && !resp.body.policy.rules;
            },
            'wait for policy to be deleted',
            true,
            90
        );

        return resposne.response.statusCode;
    }

    getPolicyArgs(scopeInfo: { scope: string; scopeId: string }): (string | undefined)[] {
        if (scopeInfo.scope === 'cluster') {
            return [scopeInfo.scope, undefined, undefined, scopeInfo.scopeId];
        } else if (scopeInfo.scope === 'department') {
            return [scopeInfo.scope, scopeInfo.scopeId, undefined, undefined];
        } else {
            return [scopeInfo.scope, undefined, scopeInfo.scopeId, undefined];
        }
    }

    getNodeSelectorTerm(
        matechExpressionOperator: MatchExpressionOperator,
        nodeNames: string[]
    ): NodeSelector {
        return {
            nodeAffinityRequired: {
                nodeSelectorTerms: [
                    {
                        matchExpressions: [
                            {
                                key: hostnameLabel,
                                operator: matechExpressionOperator,
                                values: nodeNames
                            }
                        ]
                    }
                ]
            }
        };
    }

    getRules(canEdit: boolean): { nodeAffinityRequired: { canEdit: boolean } } {
        return {
            nodeAffinityRequired: {
                canEdit: canEdit
            }
        };
    }

    async updateProjectScopedPolicy(
        isInteractive: boolean,
        name: string,
        namespace: string,
        projectId: number,
        kind: WorkloadType,
        defaults: object,
        rules?: object
    ): Promise<void> {
        const policyRequest = {
            meta: {
                projectId: projectId,
                scope: ScopeType.Project,
                name: name
            },
            policy: {
                defaults: defaults,
                rules: rules
            }
        };
        if (isInteractive) {
            await policyApiHandler.policyApi.updateWorkspacePolicyV2(
                false,
                false,
                policyRequest as WorkspacePolicyChangeRequestV2
            );
        } else {
            await policyApiHandler.policyApi.updateTrainingPolicyV2(
                false,
                false,
                policyRequest as TrainingPolicyChangeRequestV2
            );
        }
        await policyUtils.waitForPolicyToBeSynced(kind, namespace, `project-${projectId}`);
    }

    async updatePolicyV2(
        scopeName: string,
        scopeId: string,
        kind: WorkloadType,
        defaults: object,
        rules?: object,
        scopeType = ScopeType.Project
    ): Promise<{
        response: http.IncomingMessage;
        body: TrainingPolicyV2 | DistributedPolicyV2 | WorkspacePolicyV2 | InferencePolicyV2;
    }> {
        const policyRequest = {
            meta: {
                scope: scopeType,
                name: scopeName
            },
            policy: {
                defaults: defaults,
                rules: rules
            }
        };
        logger.info(`policy payload: ${JSON.stringify(policyRequest)}`);

        const key = this.getPolicyScopeKeyName(scopeType);
        policyRequest['meta'][key] = scopeType === ScopeType.Project ? parseInt(scopeId) : scopeId;

        policyApiHandler.policyApi.accessToken = flowRootState.users.system_administrator.accessToken;

        switch (kind) {
            case WorkloadType.Training:
                return await policyApiHandler.policyApi.updateTrainingPolicyV2(
                    false,
                    false,
                    policyRequest as TrainingPolicyChangeRequestV2
                );
            case WorkloadType.Distributed:
                return await policyApiHandler.policyApi.updateDistributedPolicyV2(
                    false,
                    false,
                    policyRequest as DistributedPolicyChangeRequestV2
                );
            case WorkloadType.Interactive:
                return await policyApiHandler.policyApi.updateWorkspacePolicyV2(
                    false,
                    false,
                    policyRequest as WorkspacePolicyChangeRequestV2
                );
            case WorkloadType.Inference:
                return await policyApiHandler.policyApi.updateInferencePolicyV2(
                    false,
                    false,
                    policyRequest as InferencePolicyChangeRequestV2
                );
            case WorkloadType.DistributedInference:
                return await policyApiHandler.policyApi.updateDistributedInferencePolicyV2(
                    false,
                    false,
                    policyRequest as DistributedInferencePolicyChangeRequestV2
                );
            default:
                throw new Error(`Unknown WorkloadType: ${kind}!`);
        }
    }

    async waitForPolicyToBeEnabled(namespace = 'runai'): Promise<void> {
        await waitUtils.waitForCondition(
            async () => {
                const policyEnvVar = await kubeCli.getPodEnvVarByPrefix(
                    `${namespace}-agent`,
                    'POLICY_SYNC_ENABLED',
                    namespace
                );
                return policyEnvVar.value === 'true';
            },
            'waiting for policy to be enabled',
            true
        );
    }

    async enableClusterPolicySync(): Promise<void> {
        if (versionUtils.isVersionSupported(flowRootState.cluster.version, '2.16')) {
            try {
                const patchPolicyCmd =
                    'kubectl patch runaiconfigs runai -n runai --type=\'merge\' -p=\'{"spec": {"global": {"controlPlanePoliciesEnabled": true}}}\'';
                bashCli.exec(patchPolicyCmd);
                await this.waitForPolicyToBeEnabled();
            } catch (error) {
                logger.error(`Failed to enable cluster policy sync: ${error}`);
            }
        }
    }

    disableClusterPolicySync(): void {
        const patchPolicyCmd =
            'kubectl patch runaiconfigs runai -n runai --type=\'merge\' -p=\'{"spec": {"global": {"controlPlanePoliciesEnabled": false}}}\'';
        bashCli.exec(patchPolicyCmd);
    }

    async waitForCpuCount(
        workloadType: WorkloadType,
        policyName: string,
        namespace: string,
        crd: CustomResourceDefinition,
        expectedCpuCount: number,
        timeout = 60
    ): Promise<void> {
        const getCpusCount = (): boolean => {
            try {
                const policyYaml = yaml.load(
                    bashCli.exec(`kubectl get ${workloadType} -n ${namespace} ${policyName} -oyaml`)
                );
                return policyYaml['spec'].cpu.value === expectedCpuCount.toString();
            } catch (error: unknown) {
                throw (error as Error).message;
            }
        };

        await waitUtils.waitForCondition(
            getCpusCount,
            `wait for cpu count to be equal to ${expectedCpuCount}`,
            true,
            timeout
        );
    }

    async waitForMinCpuCount(
        workloadType: WorkloadType,
        policyName: string,
        namespace: string,
        crd: CustomResourceDefinition,
        expectedCpuMinCount: number,
        timeout = 60
    ): Promise<void> {
        const getMinCpusCount = (): boolean => {
            try {
                const policyYaml = yaml.load(
                    bashCli.exec(`kubectl get ${workloadType} -n ${namespace} ${policyName} -oyaml`)
                );
                return policyYaml['spec'].cpu.rules.min === expectedCpuMinCount.toString();
            } catch (error: unknown) {
                throw (error as Error).message;
            }
        };

        await waitUtils.waitForCondition(
            getMinCpusCount,
            `wait for min cpu count to be equal to ${expectedCpuMinCount}`,
            true,
            timeout
        );
    }

    async waitForCanAdd(
        workloadType: WorkloadType,
        policyName: string,
        namespace: string,
        crd: CustomResourceDefinition,
        expectedValue: boolean,
        timeout = 60
    ): Promise<void> {
        const getMinCpusCount = (): boolean => {
            try {
                const policyYaml = yaml.load(
                    bashCli.exec(`kubectl get ${workloadType} -n ${namespace} ${policyName} -oyaml`)
                );
                return policyYaml['spec']?.pvcs?.rules?.canAdd === expectedValue;
            } catch (error: unknown) {
                throw (error as Error).message;
            }
        };

        await waitUtils.waitForCondition(
            getMinCpusCount,
            `wait for min cpu count to be equal to ${expectedValue}`,
            true,
            timeout
        );
    }

    async validateNoWorkloadPolicy(
        workloadType: WorkloadType,
        scopeInfo: { scope: string; scopeId: string }
    ): Promise<void> {
        const response = await policyUtils.getPolicyV2(workloadType, scopeInfo);
        expect(response.response.statusCode).toEqual(StatusCodes.OK);

        expect(response.body.policy).toEqual({
            defaults: undefined,
            rules: undefined,
            imposedAssets:
                typeof response.body.policy.imposedAssets === 'object'
                    ? { leader: undefined, worker: undefined }
                    : undefined,
            status: undefined,
            defaultFromOrder: undefined
        });
    }

    compareEffectiveRules(setRules: PolicyRuleType, getEffectiveRules: PolicyRuleType): boolean {
        return Object.keys(setRules).every((key) => {
            if (typeof setRules[key] === 'object' && setRules[key] !== null) {
                return this.compareEffectiveRules(setRules[key], _.get(getEffectiveRules, key));
            }

            return _.isEqual(setRules[key], _.get(getEffectiveRules, key));
        });
    }

    compareEffectiveDefaults(
        setDefaults: PolicyDefaultsType,
        getEffectiveDefaults: PolicyDefaultsType
    ): boolean {
        return Object.keys(setDefaults).every((key) => {
            if (typeof setDefaults[key] === 'object' && setDefaults[key] !== null) {
                return this.compareEffectiveDefaults(setDefaults[key], _.get(getEffectiveDefaults, key));
            }
            return _.isEqual(setDefaults[key], _.get(getEffectiveDefaults, key));
        });
    }

    async validatePolicyBody(
        workloadType: WorkloadType,
        scopeInfo: { scope: string; scopeId: string },
        defaults: PolicyDefaultsType,
        rules: PolicyRuleType
    ): Promise<void> {
        const response = await policyUtils.getPolicyV2(workloadType, scopeInfo);

        expect(response.response.statusCode).toEqual(StatusCodes.OK);
        expect(response.body.policy.defaults as PolicyDefaultsType).toEqual(defaults);
        expect(response.body.policy.rules as PolicyRuleType).toEqual(rules);

        const compareDefaultsResult = this.compareEffectiveDefaults(
            defaults,
            response.body.effective.defaults as PolicyDefaultsType
        );
        expect(compareDefaultsResult).toEqual(true);

        const compareRulesResult = this.compareEffectiveRules(
            rules,
            response.body.effective.rules as PolicyRuleType
        );
        expect(compareRulesResult).toEqual(true);
    }

    async validateSubDomainPolicyBody(
        workloadType: WorkloadType,
        scopeInfo: { scope: string; scopeId: string },
        defaults: PolicyDefaultsType,
        rules: PolicyRuleType
    ): Promise<void> {
        const response = await policyUtils.getPolicyV2(workloadType, scopeInfo);

        expect(response.response.statusCode).toEqual(StatusCodes.OK);
        expect(response.body.policy.defaults as PolicyDefaultsType).toBeUndefined();
        expect(response.body.policy.rules as PolicyRuleType).toBeUndefined();

        const compareDefaultsResult = this.compareEffectiveDefaults(
            defaults,
            response.body.effective.defaults as PolicyDefaultsType
        );
        expect(compareDefaultsResult).toEqual(true);

        const compareRulesResult = this.compareEffectiveRules(
            rules,
            response.body.effective.rules as PolicyRuleType
        );
        expect(compareRulesResult).toEqual(true);
    }

    async disableApprovalRequiredPolicy(): Promise<void> {
        try {
            policyApiHandler.policyApi.accessToken = flowRootState.users.system_administrator.accessToken;
            const response = await policyApiHandler.policyApi.updateSyncConfig({
                clusterId: flowRootState.cluster.uuid,
                avoidDeletionsOfObsoletePolicies: true
            });
            expect(response.response.statusCode).toEqual(StatusCodes.OK);
        } catch (error) {
            logger.error(
                `Failed to disable approval required for policy of cluster ${flowRootState.cluster.uuid}: ${error}`
            );
            throw error;
        }
    }
}
