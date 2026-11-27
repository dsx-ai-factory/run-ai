/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    config,
    Utils,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    flowRootState,
    workloadsCrudUtils,
    orgUnitUtils,
    nodePoolUtils,
    clusterUtils
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { WorkloadType as WorkloadTypeFromWorkloadsCrudUtils } from '../../../../../../modules/helpers/utils/WorkloadsCrudUtils';
import { workloadsService } from '../../../../../../modules/api';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { WorkloadPage } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { ProjectPage } from '../../../../../../modules/ui/runaiV2/pages/ProjectPage';
import { DepartmentPage } from '../../../../../../modules/ui/runaiV2/pages/DepartmentPage';
import { EnvironmentPage } from '../../../../../../modules/ui/runaiV2/pages/EnvironmentPage';
import { AccessRulesPage } from '../../../../../../modules/ui/runaiV2/pages/AccessRulesPage';
import { UsersV2 } from '../../../../../../modules/ui/runaiV2/pages/UsersV2';

// CSV data validation function
function validateCSVData(csvData: any[], expectedColumns: string[]): void {
    expect(csvData.length).toBeGreaterThan(0);

    const firstRowKeys = Object.keys(csvData[0]);

    for (const expectedColumn of expectedColumns) {
        expect(firstRowKeys).toContain(expectedColumn);
    }

    for (let i = 0; i < csvData.length; i++) {
        const row = csvData[i];
        const rowNumber = i + 1;

        for (const expectedColumn of expectedColumns) {
            expect(Object.prototype.hasOwnProperty.call(row, expectedColumn)).toBe(true);

            const value = row[expectedColumn];
            expect(value).not.toBeNull();
            expect(value).not.toBeUndefined();

            if (value === '-') continue;

            if (typeof value === 'string' && value.trim() === '') {
                throw new Error(`Row ${rowNumber}, Column '${expectedColumn}' has empty string value`);
            }
        }
    }
}

// Date fields validation function
function validateDateFields(csvData: any[], dateFields: string[]): void {
    for (const row of csvData) {
        for (const field of dateFields) {
            if (row[field] && row[field] !== '-') {
                expect(row[field]).toMatch(
                    /^(0?[1-9]|[12][0-9]|3[01])\/(0?[1-9]|1[0-2])\/\d{4}\s([01]?[0-9]|2[0-3]):[0-5][0-9]$/
                );
            }
        }
    }
}

// Email fields validation function
function validateEmailFields(csvData: any[], emailFields: string[]): void {
    for (const row of csvData) {
        for (const field of emailFields) {
            if (row[field] && row[field] !== '-') {
                expect(row[field]).toMatch(/^[A-Za-z0-9._-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$/);
            }
        }
    }
}

test.describe(`Export to CSV ${tags.tool.ui} ${tags.teams.loong}`, function () {
    supportedFrom(SupportedVersions.V2_20)(function () {
        test.setTimeout(300 * 1000);
        let projectName: string;
        let namespace: string;
        let project: Project;
        let browserEngine: BrowserEngine;
        let authPage: AuthPage;
        let workloadPage: WorkloadPage;
        let videoName: string;
        let projectPage: ProjectPage;
        let environmentPage: EnvironmentPage;
        const workloadIds: string[] = [];
        const projects: any[] = [];

        test.beforeAll(async function () {
            projectName = objectGenerator.generateShortNameByPrefix('project-workload-info');
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 8,
                maxAllowedGpus: 8
            });

            workloadsCrudUtils.accessToken = flowRootState.users.l1_researcher.accessToken;
            workloadsService.workloadsApi.accessToken =
                flowRootState.users.l1_researcher.accessToken;

            browserEngine = new BrowserEngine();
        });

        test.afterAll('deleting project and namespace', async () => {
            // Clean up workloads
            if (workloadIds.length > 0) {
                try {
                    for (const workloadId of workloadIds) {
                        await workloadsCrudUtils.deleteTraining(workloadId);
                    }
                    logger.info('Successfully cleaned up workloads');
                } catch (error) {
                    logger.error('Failed to clean up workloads:', error);
                }
            }

            // Clean up projects
            if (projects.length > 0) {
                try {
                    for (const project of projects) {
                        await operations.deleteProjectAndNamespace(project.id);
                    }
                    logger.info('Successfully cleaned up projects');
                } catch (error) {
                    logger.error('Failed to clean up projects:', error);
                }
            }

            // Clean up main project and namespace
            try {
                await operations.deleteProjectAndNamespace(project.id);
                logger.info(`deleted project ${projectName} and namespace ${namespace}`);
            } catch (error) {
                logger.error(`failed to delete project ${projectName} and namespace ${namespace}`);
                throw error;
            }
        });

        test.beforeEach(async function () {
            await browserEngine.start();
            authPage = new AuthPage(browserEngine.currentPage);
            workloadPage = new WorkloadPage(browserEngine.currentPage);
            projectPage = new ProjectPage(browserEngine.currentPage);
            environmentPage = new EnvironmentPage(browserEngine.currentPage);
            await authPage.signIn(
                flowRootState.users.l1_researcher.email,
                flowRootState.users.l1_researcher.password
            );
        });

        test.afterEach('stop browser', async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, videoName);
        });

        test('T5307795: Workloads - export to csv', async function () {
            videoName = 'workloads-export-to-csv';
            const workloadPrefix = 'workload-export-to-csv';
            const workloadNames = [];
            let csvData: any[];

            await test.step('Create three workloads', async () => {
                for (let i = 0; i < 3; i++) {
                    const { body: workload } = await workloadsCrudUtils.createTraining(
                        workloadPrefix,
                        true,
                        flowRootState.cluster.uuid,
                        project.id.toString(),
                        { command: 'sleep infinity', image: config.quickStartImage.ubuntu },
                        flowRootState.users.l1_researcher.accessToken
                    );
                    workloadNames.push(workload.name);
                    workloadIds.push(workload.workloadId);

                    await workloadsCrudUtils.waitForWorkloadPhase(
                        workload.workloadId,
                        Phase.Running,
                        WorkloadTypeFromWorkloadsCrudUtils.TRAINING,
                        undefined,
                        120
                    );
                }
            });

            await test.step('Navigate to Workloads page and set rows per page', async () => {
                await workloadPage.goToPage();
                await workloadPage.setRowsPerPage(1);
                const workloadNamesFiltered = await workloadPage.getWorkloadNames();
                expect(workloadNamesFiltered.length).toBe(1);
                await workloadPage.showAllColumns();
            });

            await test.step('Download CSV file', async () => {
                const csvFilePath = await workloadPage.downloadCSV('workloads-export-to-csv.csv');
                csvData = Utils.readCSVFile(csvFilePath);
                logger.info(JSON.stringify(csvData, null, 2));
            });

            await test.step('Check if workloadNames are in csvData', () => {
                expect(csvData.length).toBeGreaterThanOrEqual(workloadNames.length);
                for (const workloadName of workloadNames) {
                    expect(csvData.some((row) => row['Workload'] === workloadName)).toBe(true);
                }
            });

            await test.step('Validate CSV columns and data integrity', () => {
                const columns = [
                    'Workload',
                    'Type',
                    'Category',
                    'Priority',
                    'Preemptibility',
                    'Status',
                    'Project',
                    'Department',
                    'Created by',
                    'AI application',
                    'Running / requested pods',
                    'Requested runs',
                    'Creation time',
                    'Completion time',
                    'Total pending time',
                    'Total runtime',
                    'Guaranteed runtime left',
                    'Connection(s)',
                    'Image',
                    'Workload architecture',
                    'GPU compute request',
                    'GPU compute allocation',
                    'GPU memory request',
                    'GPU memory allocation',
                    'Idle GPU devices',
                    'CPU compute request',
                    'CPU compute allocation',
                    'CPU memory request',
                    'CPU memory allocation'
                ];

                validateCSVData(csvData, columns);
            });

            await test.step('Validate format for date and email fields', () => {
                const dateFields = ['Creation time', 'Completion time'];
                validateDateFields(csvData, dateFields);
                const emailFields = ['Created by'];
                validateEmailFields(csvData, emailFields);
            });

            await test.step('Validate specific workload data in CSV', () => {
                const randomIndex = Math.floor(Math.random() * workloadNames.length);
                const selectedWorkloadName = workloadNames[randomIndex];
                logger.info(`Selected workload for validation: ${selectedWorkloadName}`);
                const csvRow = csvData.find((row) => row['Workload'] === selectedWorkloadName);
                expect(csvRow).toBeDefined();

                expect(csvRow['Type']).toBe('Training');
                expect(csvRow['Image']).toBe('ubuntu');
                expect(csvRow['Created by']).toBe(flowRootState.users.l1_researcher.email);
                expect(csvRow['Status']).toBe('Running');
                expect(csvRow['Project']).toBe(projectName);
                expect(csvRow['Department']).toBe('default');
                expect(csvRow['Category']).toBe('Train');
                expect(csvRow['Running / requested pods']).toBe('1/1');
            });
        });

        test('T5307796: Project - export to CSV', async function () {
            videoName = 'project-export-to-csv';
            const projectPrefix = 'project-export-to-csv';
            const projectNames = [];
            let csvData: any[];

            await test.step('Create three projects', async () => {
                for (let i = 0; i < 3; i++) {
                    const projectName = objectGenerator.generateShortNameByPrefix(projectPrefix);
                    const namespace = Utils.namespaceFromProjectName(projectName);
                    const project = await projectResourcesUtils.createProject(namespace, {
                        name: projectName,
                        deservedGpus: 4,
                        maxAllowedGpus: 4
                    });
                    projectNames.push(projectName);
                    projects.push(project);
                }
            });

            await test.step('Login UI and navigate to Organization -> Project page', async () => {
                await projectPage.goToPage();
                await projectPage.setRowsPerPage(1);
                const projectNamesInUI = await projectPage.getProjectNames();
                expect(projectNamesInUI.length).toBe(1);
            });

            await test.step('Download CSV file', async () => {
                const csvFilePath = await workloadPage.downloadCSV('projects-export-to-csv.csv');
                csvData = Utils.readCSVFile(csvFilePath);
                logger.info(JSON.stringify(csvData, null, 2));
            });

            await test.step('Check if created projects are included in CSV', () => {
                expect(csvData.length).toBeGreaterThanOrEqual(projectNames.length);
                for (const projectName of projectNames) {
                    expect(csvData.some((row) => row['Project'] === projectName)).toBe(true);
                }
            });

            await test.step('Validate project CSV columns and data integrity', () => {
                const projectColumns = [
                    'Project',
                    'Department',
                    'Status',
                    'GPU quota',
                    'Allocated GPUs',
                    'Avg. GPU allocation',
                    'Avg. GPU utilization',
                    'Avg. GPU memory utilization',
                    'Creation time'
                ];
                validateCSVData(csvData, projectColumns);
            });

            await test.step('Validate date format for date fields', () => {
                const dateFields = ['Creation time'];
                validateDateFields(csvData, dateFields);
            });

            await test.step('Validate specific project data in CSV', () => {
                const randomIndex = Math.floor(Math.random() * projectNames.length);
                const selectedProjectName = projectNames[randomIndex];
                logger.info(`Selected project for validation: ${selectedProjectName}`);
                const csvRow = csvData.find((row) => row['Project'] === selectedProjectName);
                expect(csvRow).toBeDefined();
                expect(csvRow['Status']).toBe('Ready');
                expect(csvRow['Department']).toBe('default');
                expect(csvRow['GPU quota']).toBe('4');
                expect(csvRow['Allocated GPUs']).toBe('-');
                logger.info(`Project ${selectedProjectName} data validation passed`);
            });
        });

        test('T5307798: Environment - export to csv', async function () {
            videoName = 'environments-export-to-csv';
            let csvData: any[];
            const clusterConfig = await clusterUtils.getClusterReportedConfig();
            const isHostBased = clusterConfig?.subdomainEnabled === true;

            await test.step('Login UI and navigate to Workloads Manager -> Environments page', async () => {
                await environmentPage.goToPage();
                await environmentPage.setRowsPerPage(5);
                const environmentNamesInUI = await environmentPage.getEnvironmentNames();
                expect(environmentNamesInUI.length).toBe(5);
            });

            await test.step('Download CSV file', async () => {
                const csvFilePath = await workloadPage.downloadCSV('environments-export-to-csv.csv');
                csvData = Utils.readCSVFile(csvFilePath);
                logger.info(JSON.stringify(csvData, null, 2));
                expect(csvData.length).toBeGreaterThan(5);
            });

            await test.step('Validate environment CSV columns and data integrity', () => {
                expect(csvData.length).toBeGreaterThan(5);
                const environmentColumns = [
                    'Environment',
                    'Status',
                    'Scope',
                    'Image',
                    'Tool(s)',
                    'Created by',
                    'Creation time',
                    'Cluster'
                ];
                validateCSVData(csvData, environmentColumns);

                const dateFields = ['Creation time'];
                validateDateFields(csvData, dateFields);
            });

            await test.step('Check default environments are included in CSV', () => {
                logger.info(`Cluster routing: ${isHostBased ? 'host-based' : 'path-based'}`);

                const expectedEnvironments = isHostBased
                    ? [
                          'gpt2',
                          'tensorboard-tensorflow',
                          'jupyter-scipy',
                          'rstudio',
                          'vscode',
                          'llm-server',
                          'huggingface-chat-ui',
                          'openwebui'
                      ]
                    : [
                          'tensorboard',
                          'nemo',
                          'jupyter-lab',
                          'bio-nemo',
                          'pytorch',
                          'llm-server',
                          'huggingface-chat-ui-path'
                      ];

                for (const expectedEnv of expectedEnvironments) {
                    const csvRow = csvData.find((row) => row['Environment'] === expectedEnv);
                    expect(csvRow).toBeDefined();
                    logger.info(`Environment ${expectedEnv} found in CSV`);
                }
            });

            await test.step('Validate specific environment data in CSV', () => {
                const tensorboardEnvName = isHostBased ? 'tensorboard-tensorflow' : 'tensorboard';
                const tensorboardRow = csvData.find((row) => row['Environment'] === tensorboardEnvName);
                expect(tensorboardRow['Image']).toMatch(/^tensorflow\/tensorflow:/);
                expect(tensorboardRow['Tool(s)']).toBe('TensorBoard');
                expect(tensorboardRow['Scope']).toBe('tenant');
            });
        });
    });
});

test.describe(`Export to CSV - Admin Tests ${tags.tool.ui} ${tags.teams.loong}`, function () {
    supportedFrom(SupportedVersions.V2_20)(function () {
        let projectName: string;
        let namespace: string;
        let project: Project;
        let browserEngine: BrowserEngine;
        let authPage: AuthPage;
        let workloadPage: WorkloadPage;
        let videoName: string;
        let departmentPage: DepartmentPage;
        let accessRulesPage: AccessRulesPage;
        let usersPage: UsersV2;
        const departments: any[] = [];

        test.beforeAll(async function () {
            projectName = objectGenerator.generateShortNameByPrefix('project-workload-info');
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 8,
                maxAllowedGpus: 8
            });

            workloadsCrudUtils.accessToken = flowRootState.users.system_administrator.accessToken;
            workloadsService.workloadsApi.accessToken = flowRootState.users.system_administrator.accessToken;

            browserEngine = new BrowserEngine();
        });

        test.afterAll('deleting project and namespace', async () => {
            // Clean up departments
            if (departments.length > 0) {
                try {
                    for (const department of departments) {
                        await orgUnitUtils.deleteDepartment(department.id);
                    }
                    logger.info('Successfully cleaned up departments');
                } catch (error) {
                    logger.error('Failed to clean up departments:', error);
                }
            }

            // Clean up main project and namespace
            try {
                await operations.deleteProjectAndNamespace(project.id);
                logger.info(`deleted project ${projectName} and namespace ${namespace}`);
            } catch (error) {
                logger.error(`failed to delete project ${projectName} and namespace ${namespace}`);
                throw error;
            }
        });

        test.beforeEach(async function () {
            await browserEngine.start();
            authPage = new AuthPage(browserEngine.currentPage);
            departmentPage = new DepartmentPage(browserEngine.currentPage);
            usersPage = new UsersV2(browserEngine.currentPage);
            accessRulesPage = new AccessRulesPage(browserEngine.currentPage);
            await authPage.signIn(
                flowRootState.users.system_administrator.email,
                flowRootState.users.system_administrator.password
            );
        });

        test.afterEach('stop browser', async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, videoName);
        });

        test('T5307797: Departments - export to csv', async function () {
            videoName = 'departments-export-to-csv';
            const departmentPrefix = 'department-export';
            const departmentNames = [];
            let defaultNodePoolId: string;
            let csvData: any[];

            await test.step('Create three departments', async () => {
                defaultNodePoolId = String((await nodePoolUtils.getDefaultNodePoolFromBackend()).id);
                for (let i = 1; i <= 3; i++) {
                    const departmentName = objectGenerator.generateShortNameByPrefix(departmentPrefix);
                    const department = await orgUnitUtils.createDepartment({
                        name: departmentName,
                        clusterId: flowRootState.cluster.uuid,
                        resources: [
                            {
                                nodePool: { id: defaultNodePoolId, name: 'default' },
                                gpu: { deserved: 16 }
                            }
                        ]
                    });
                    departmentNames.push(departmentName);
                    departments.push(department);
                }
            });

            await test.step('Login UI and navigate to Organization -> Departments page', async () => {
                await departmentPage.goToPage();
                await departmentPage.setRowsPerPage(1);
                await departmentPage.addFilterByDepartment(departmentNames[0]);
                const departmentNamesInUI = await departmentPage.qTable.getNames();
                expect(departmentNamesInUI.length).toBe(1);
            });

            await test.step('Download CSV file', async () => {
                const csvFilePath = await departmentPage.downloadCSV('departments-export-to-csv.csv');
                csvData = Utils.readCSVFile(csvFilePath);
                logger.info(JSON.stringify(csvData, null, 2));
            });

            await test.step('Check if created departments are included in CSV', () => {
                expect(csvData.length).toBeGreaterThan(departmentNames.length);
                for (const departmentName of departmentNames) {
                    expect(csvData.some((row) => row['Department'] === departmentName)).toBe(true);
                }
            });

            await test.step('Validate department CSV columns and data integrity', () => {
                const departmentColumns = ['Department', 'GPU quota', 'Allocated GPUs', 'Creation time'];
                validateCSVData(csvData, departmentColumns);
            });

            await test.step('Validate date format for date fields', () => {
                const dateFields = ['Creation time'];
                validateDateFields(csvData, dateFields);
            });

            await test.step('Validate specific department data in CSV', () => {
                const randomIndex = Math.floor(Math.random() * departmentNames.length);
                const selectedDepartmentName = departmentNames[randomIndex];
                logger.info(`Selected department for validation: ${selectedDepartmentName}`);
                const csvRow = csvData.find((row) => row['Department'] === selectedDepartmentName);
                expect(csvRow['GPU quota']).toBe('16');
                expect(csvRow['Allocated GPUs']).toBe('-');
                logger.info(`Department ${selectedDepartmentName} data validation passed`);
            });
        });

        test('T5307799: Access rules - export to csv', async function () {
            videoName = 'access-rules-export-to-csv';
            let csvData: any[];
            await test.step('Login UI and navigate to Access -> Access rules page', async () => {
                await accessRulesPage.goToPage();
                await accessRulesPage.setRowsPerPage(5);
                const accessRulesInUI = await accessRulesPage.getAccessRulesCount();
                expect(accessRulesInUI).toBe(5);
            });

            await test.step('Download CSV file', async () => {
                const csvFilePath = await accessRulesPage.downloadCSV('access-rules-export-to-csv.csv');
                csvData = Utils.readCSVFile(csvFilePath);
                logger.info(JSON.stringify(csvData, null, 2));
            });

            await test.step('Validate access rules CSV columns and data integrity', () => {
                const accessRulesColumns = [
                    'Type',
                    'Subject',
                    'Role',
                    'Scope',
                    'Status',
                    'Authorized by',
                    'Creation time',
                    'Last updated'
                ];
                validateCSVData(csvData, accessRulesColumns);
            });

            await test.step('Validate date format for date fields', () => {
                const dateFields = ['Creation time', 'Last updated'];
                validateDateFields(csvData, dateFields);
            });

            await test.step('Check specific access rule record exists with correct values', () => {
                const specificRecord = csvData.find(
                    (row) => row['Subject'] === config.controlPlaneAdminUsername && row['Role'] === 'System administrator'
                );
                expect(specificRecord).toBeDefined();
                logger.info('Specific access rule record validation passed');
            });
        });

        test('T5307800: Users - export to csv', async function () {
            videoName = 'users-export-to-csv';
            let csvData: any[];

            await test.step('Login UI and navigate to Access -> Users page', async () => {
                await usersPage.navigateToUsers();
                await usersPage.setRowsPerPage(5);
                const usersInUI = await usersPage.getUsersCount();
                expect(usersInUI).toBe(5);
            });

            await test.step('Download CSV file', async () => {
                const csvFilePath = await usersPage.downloadCSV('users-export-to-csv.csv');
                csvData = Utils.readCSVFile(csvFilePath);
                logger.info(JSON.stringify(csvData, null, 2));
            });

            await test.step('Validate CSV columns and data integrity', () => {
                const usersColumns = [
                    'User',
                    'Full name',
                    'Type',
                    'Last login',
                    'Created by',
                    'Creation time',
                    'Last updated'
                ];
                validateCSVData(csvData, usersColumns);
            });

            await test.step('Validate date format for date fields', () => {
                const dateFields = ['Last login', 'Creation time', 'Last updated'];
                validateDateFields(csvData, dateFields);
                const emailFields = ['User', 'Created by'];
                validateEmailFields(csvData, emailFields);
            });
        });
    });
});
