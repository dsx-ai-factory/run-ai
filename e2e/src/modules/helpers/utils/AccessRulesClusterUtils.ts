/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    accessRulesClusterUtils,
    authorizationUtils,
    authV2Utils,
    crds,
    flowRootState,
    logger,
    versionUtils,
    waitUtils
} from '../index';
import { KubernetesObject } from '@kubernetes/client-node/dist/types';
import { kubeCli } from '../../cli';
import { expect } from 'chai';
import { AccessRule } from '../../api/apis/authorizationApi/generated/model/accessRule';
import { namespaceCluster } from '../config/const';
import { V1RoleBinding } from '@kubernetes/client-node';
import { SubjectType } from '../../api/apis/authorizationApi/generated/model/subjectType';

export class AccessRulesClusterUtils {
    isClusterWithAccessRules(): boolean {
        return versionUtils.isVersionSupported(flowRootState.cluster.version, '2.14');
    }

    private async getAccessRuleCrds(): Promise<KubernetesObject[]> {
        return await kubeCli.getCrds(crds.accessRuleCrd);
    }

    private async getAccessRuleCrdBySpec(
        roleName: string,
        subjectId: string,
        subjectType: string,
        scopeType: string,
        scopeName = ''
    ): Promise<KubernetesObject> {
        const accessRules = await this.getAccessRuleCrds();

        return accessRules.find(
            (accessRule) =>
                accessRule['spec'].roleName.toLowerCase() === authV2Utils.getRoleNameInK8s(roleName) &&
                accessRule['spec'].subjectId.toLowerCase() === subjectId.toLowerCase() &&
                accessRule['spec'].subjectType.toLowerCase() === subjectType.toLowerCase() &&
                accessRule['spec'].scopeType.toLowerCase() === scopeType.toLowerCase() &&
                (scopeName ? accessRule['spec'].scopeName.toLowerCase() === scopeName.toLowerCase() : true)
        );
    }

    private async getAccessRuleCrdById(accessRuleId: string): Promise<KubernetesObject> {
        const accessRules = await this.getAccessRuleCrds();

        return accessRules.find(
            (accessRule) => accessRule['metadata'].labels['runai/accessrule-id'] === accessRuleId
        );
    }

    async waitForAccessRuleCrdBySpec(
        roleName: string,
        subjectId: string,
        subjectType: string,
        scopeType: string,
        scopeName = '',
        timeout = 40,
        exceptedStatuses = ['Ready']
    ): Promise<object> {
        if (!this.isClusterWithAccessRules()) {
            return;
        }

        let res: object;
        let status;
        let accessRuleName;
        logger.debug(
            `Waiting for accessrule crd to be updated in cluster - [${roleName}/${subjectType}/${subjectId}/${scopeType}/${scopeName}]`
        );

        const doesAccessRuleStatusMatch = async (): Promise<boolean> => {
            try {
                res = await this.getAccessRuleCrdBySpec(
                    roleName,
                    subjectId,
                    subjectType,
                    scopeType,
                    scopeName
                );
                accessRuleName = res['metadata'].name;
                status = res['status'].phase;
                return exceptedStatuses.includes(status);
            } catch (e) {
                return false;
            }
        };

        try {
            await waitUtils.waitForCondition(
                doesAccessRuleStatusMatch,
                `waitForAccessRuleCrdBySpec [${roleName}/${subjectType}/${subjectId}/${scopeType}/${scopeName}]; exceptedStatuses: ${exceptedStatuses}`,
                true,
                timeout
            );
        } catch (e) {
            logger.debug(
                `Failed waiting for accessrule with statuses [${exceptedStatuses}], status was: ${status} (name was: ${accessRuleName})`
            );
            e.message = `${e.message} - accessrule status was: ${status}`;
            throw e;
        }
        return res;
    }

    private async waitForAccessRuleCrdById(
        accessRuleId: string,
        timeout = 40,
        exceptedStatuses = ['Ready']
    ): Promise<object> {
        let res: object;
        let status;
        let accessRuleName;

        const doesAccessRuleStatusMatch = async (): Promise<boolean> => {
            try {
                res = await this.getAccessRuleCrdById(accessRuleId);
                accessRuleName = res['metadata'].name;
                status = res['status'].phase;
                return exceptedStatuses.includes(status);
            } catch (e) {
                return false;
            }
        };

        try {
            await waitUtils.waitForCondition(
                doesAccessRuleStatusMatch,
                `waitForAccessRuleCrdById [${accessRuleId}]; exceptedStatuses: ${exceptedStatuses}`,
                true,
                timeout
            );
        } catch (e) {
            logger.debug(
                `Failed waiting for accessrule with statuses [${exceptedStatuses}], status was: ${status} (name was: ${accessRuleName})`
            );
            e.message = `${e.message} - accessrule status was: ${status}`;
            throw e;
        }
        return res;
    }

    async waitForAccessRulesCrdsById(accessRuleIds: string[]): Promise<void> {
        if (!this.isClusterWithAccessRules()) {
            return;
        }

        logger.info(`Waiting for accessrules ${accessRuleIds} to be ready in cluster`);
        const promises = [];
        for (const accessRuleId of accessRuleIds) {
            promises.push(accessRulesClusterUtils.waitForAccessRuleCrdById(accessRuleId));
        }
        await Promise.all(promises);
    }

    async waitForAccessRuleToBeDeleted(
        accessRuleId: number,
        roleName: string,
        subjectId: string,
        subjectType: string,
        scopeType: string,
        scopeName = '',
        timeout = 60
    ): Promise<void> {
        if (!this.isClusterWithAccessRules()) {
            await authorizationUtils.waitForAccessRule(accessRuleId, false);
            return;
        }

        const doesAccessRuleExist = async (): Promise<boolean> => {
            try {
                const res = await this.getAccessRuleCrdBySpec(
                    roleName,
                    subjectId,
                    subjectType,
                    scopeType,
                    scopeName
                );
                return !!res;
            } catch (e) {
                return false;
            }
        };

        await waitUtils.waitForCondition(
            doesAccessRuleExist,
            `waitForAccessRuleCrdBySpecToBeDeleted [${roleName}/${subjectType}/${subjectId}/${scopeType}/${scopeName}]`,
            false,
            timeout
        );
    }

    getRoleClusterRoleBindingName(roleName: string): string {
        return `runai-${authV2Utils.getRoleNameInK8s(roleName)}`;
    }

    getRoleRunaiRoleBindingName(roleName: string): string {
        return `${this.getRoleClusterRoleBindingName(roleName)}-runai`;
    }

    getRoleRunaiClusterScopedRoleBindingName(roleName: string): string {
        return `${this.getRoleRunaiRoleBindingName(roleName)}-cs`;
    }

    getRoleProjectRoleBindingName(roleName: string): string {
        return `${this.getRoleClusterRoleBindingName(roleName)}-ns`;
    }

    async validateAccessRulesRoleBindingInCluster(
        createdAccessRules: AccessRule[],
        namespacesOfProjects: string[]
    ): Promise<void> {
        if (!this.isClusterWithAccessRules()) {
            return;
        }

        logger.info(
            `Validating accessrules crds [${createdAccessRules.map((ar) => ar.id)}] RoleBindings in cluster`
        );

        const clusterRoleBindings = await kubeCli.listClusterRoleBindings();
        const runaiRoleBindings = await kubeCli.listRoleBindings(namespaceCluster);
        const namespaceRoleBindingsPromises = namespacesOfProjects.map((namespacesOfTestCreatedProject) =>
            kubeCli.listRoleBindings(namespacesOfTestCreatedProject)
        );
        const namespacesRoleBindings = await Promise.all(namespaceRoleBindingsPromises);

        const namespaceToRoleBindings = new Map<string, V1RoleBinding[]>();
        for (let idx = 0; idx < namespacesOfProjects.length; idx++) {
            namespaceToRoleBindings[namespacesOfProjects[idx]] = namespacesRoleBindings[idx];
        }

        for (const accessRule of createdAccessRules) {
            // validate subject in rolebinding in runai namespace
            const runaiRoleRoleBinding = runaiRoleBindings.find(
                (runaiRoleBinding) =>
                    runaiRoleBinding.metadata.name ===
                    accessRulesClusterUtils.getRoleRunaiRoleBindingName(accessRule.roleName)
            );

            let foundAccessRulesSubject = runaiRoleRoleBinding.subjects.find(
                (subject) => subject.name === accessRule.subjectId
            );
            expect(
                foundAccessRulesSubject,
                `expect to find accessrule's subject ${accessRule.subjectId} in runai role binding ${runaiRoleRoleBinding.metadata.name}`
            ).to.not.be.undefined;

            // validate subject in rolebinding in projects' namespaces
            for (const [namespace, namespaceRoleBindings] of Object.entries(namespaceToRoleBindings)) {
                const namespaceRoleRoleBindingName = accessRulesClusterUtils.getRoleProjectRoleBindingName(
                    accessRule.roleName
                );
                const namespaceRoleRoleBinding = namespaceRoleBindings.find(
                    (namespaceRoleBinding) =>
                        namespaceRoleBinding.metadata.name === namespaceRoleRoleBindingName
                );
                expect(
                    namespaceRoleRoleBinding,
                    `expect to find rolebinding with name '${namespaceRoleRoleBindingName}' in project namespace ${namespace}`
                ).to.not.be.undefined;

                foundAccessRulesSubject = namespaceRoleRoleBinding.subjects.find(
                    (subject) => subject.name === accessRule.subjectId
                );
                expect(
                    foundAccessRulesSubject,
                    `expect to find accessrule's subject ${accessRule.subjectId} in namespace's role binding ${namespaceRoleRoleBinding.metadata.name}`
                ).to.not.be.undefined;
            }

            // validate subject in clusterrolebinding
            const accessRuleClusterRoleBinding = clusterRoleBindings.find(
                (clusterRoleBinding) =>
                    clusterRoleBinding.metadata.name ===
                    accessRulesClusterUtils.getRoleClusterRoleBindingName(accessRule.roleName)
            );

            foundAccessRulesSubject = accessRuleClusterRoleBinding.subjects.find(
                (subject) => subject.name === accessRule.subjectId
            );
            expect(
                foundAccessRulesSubject,
                `expect to find accessrule's subject ${accessRule.subjectId} in cluster role binding ${accessRuleClusterRoleBinding.metadata.name}`
            ).to.not.be.undefined;
        }
    }

    async getAccessRulesByUser(accessToken: string, userEmail: string): Promise<any> {
        const accessRules = await authorizationUtils.getAccessRules(
            {
                subjectType: SubjectType.User,
                filterBy: [`subjectIdFilter=@${userEmail}`]
            },
            accessToken
        );
        return accessRules.accessRules.length;
    }

    async getAccessRulesDetailsByUser(accessToken: string, userEmail: string): Promise<AccessRule[]> {
        const accessRules = await authorizationUtils.getAccessRules(
            {
                subjectType: SubjectType.User,
                filterBy: [`subjectIdFilter=@${userEmail}`]
            },
            accessToken
        );
        return accessRules.accessRules;
    }
}
