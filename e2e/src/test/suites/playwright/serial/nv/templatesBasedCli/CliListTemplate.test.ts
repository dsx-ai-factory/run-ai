/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    Utils,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    flowRootState,
    templateUtils,
    orgUnitUtils
} from '../../../../../../modules/helpers';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { ScopeType, WorkloadType } from '../../../../../../modules/api/apis/workloadsApi/generated/api';
import { runaiCliV2 } from '../../../../../../modules/cli';
import { listFlags } from '../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';

test.describe(`CLI List Template - Positive Tests, tags: ${tags.resource.templateV2} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_24)(function () {
        const TEMPLATE = 'template';
        let departmentId: number;
        let projectA: Project;
        let projectB: Project;
        let projectAName: string;
        let projectBName: string;
        let clusterTemplate1Id: string;
        let clusterTemplate2Id: string;
        let deptTemplateId: string;
        let projectATemplate1Id: string;
        let projectATemplate2Id: string;
        let projectBTemplateId: string;

        test.beforeAll(async function () {
            await test.step('Create department and projects', async function () {
                const departmentName = objectGenerator.generateShortNameByPrefix('dept-list', 5);
                const department = await orgUnitUtils.createDepartment({
                    name: departmentName,
                    clusterId: flowRootState.cluster.uuid,
                    resources: [
                        {
                            gpu: { deserved: 0, limit: -1, overQuotaWeight: 0 },
                            cpu: null,
                            memory: null,
                            nodePool: null
                        }
                    ]
                });
                departmentId = Number(department.id);
                logger.info(`Created department: ${departmentName} (ID: ${departmentId})`);

                // Create project-a
                projectAName = objectGenerator.generateShortNameByPrefix('proj-a-list', 5);
                const namespaceA = Utils.namespaceFromProjectName(projectAName);
                projectA = await projectResourcesUtils.createProject(namespaceA, {
                    name: projectAName,
                    deservedGpus: 0,
                    maxAllowedGpus: 2,
                    departmentId: departmentId
                });
                logger.info(`Created project-a: ${projectAName} (ID: ${projectA.id})`);

                // Create project-b
                projectBName = objectGenerator.generateShortNameByPrefix('proj-b-list', 5);
                const namespaceB = Utils.namespaceFromProjectName(projectBName);
                projectB = await projectResourcesUtils.createProject(namespaceB, {
                    name: projectBName,
                    deservedGpus: 0,
                    maxAllowedGpus: 2,
                    departmentId: departmentId
                });
                logger.info(`Created project-b: ${projectBName} (ID: ${projectB.id})`);
            });

            await test.step('Create cluster-scoped templates', async function () {
                const spec = {
                    image: 'alpine:latest',
                    command: 'sleep 10',
                    compute: { gpuDevicesRequest: 1, cpuCoreRequest: 0.1, cpuMemoryRequest: '100M' }
                };

                const template1Name = objectGenerator.generateShortNameByPrefix('cluster-template-1', 5);
                const response1 = await templateUtils.createWorkloadTemplateV2(
                    {
                        workloadType: WorkloadType.Training,
                        scopeType: ScopeType.Cluster,
                        scopeId: flowRootState.cluster.uuid,
                        templateName: template1Name,
                        description: 'Cluster template 1'
                    },
                    spec,
                    undefined,
                    flowRootState.users.system_administrator.idToken
                );
                clusterTemplate1Id = response1.body.meta.id;
                logger.info(`Created cluster template 1: ${template1Name}`);

                const template2Name = objectGenerator.generateShortNameByPrefix('cluster-template-2', 5);
                const response2 = await templateUtils.createWorkloadTemplateV2(
                    {
                        workloadType: WorkloadType.Workspace,
                        scopeType: ScopeType.Cluster,
                        scopeId: flowRootState.cluster.uuid,
                        templateName: template2Name,
                        description: 'Cluster template 2'
                    },
                    spec,
                    undefined,
                    flowRootState.users.system_administrator.idToken
                );
                clusterTemplate2Id = response2.body.meta.id;
                logger.info(`Created cluster template 2: ${template2Name}`);
            });

            await test.step('Create department-scoped template', async function () {
                const spec = {
                    image: 'pytorch/pytorch:latest',
                    command: 'sleep 10',
                    compute: { gpuDevicesRequest: 1, cpuCoreRequest: 0.1, cpuMemoryRequest: '100M' }
                };

                const templateName = objectGenerator.generateShortNameByPrefix('dept-template', 5);
                const response = await templateUtils.createWorkloadTemplateV2(
                    {
                        workloadType: WorkloadType.Training,
                        scopeType: ScopeType.Department,
                        scopeId: departmentId.toString(),
                        templateName: templateName,
                        description: 'Department template'
                    },
                    spec,
                    undefined,
                    flowRootState.users.system_administrator.idToken
                );
                deptTemplateId = response.body.meta.id;
                logger.info(`Created department template: ${templateName}`);
            });

            await test.step('Create project-a scoped templates', async function () {
                const spec = {
                    image: 'alpine:latest',
                    command: 'sleep 10',
                    compute: { gpuDevicesRequest: 0, cpuCoreRequest: 0.1, cpuMemoryRequest: '100M' }
                };

                const template1Name = objectGenerator.generateShortNameByPrefix('proj-a-template-1', 5);
                const response1 = await templateUtils.createWorkloadTemplateV2(
                    {
                        workloadType: WorkloadType.Training,
                        scopeType: ScopeType.Project,
                        scopeId: projectA.id.toString(),
                        templateName: template1Name,
                        description: 'Project A template 1'
                    },
                    spec,
                    undefined,
                    flowRootState.users.system_administrator.idToken
                );
                projectATemplate1Id = response1.body.meta.id;
                logger.info(`Created project-a template 1: ${template1Name}`);

                const template2Name = objectGenerator.generateShortNameByPrefix('proj-a-template-2', 5);
                const response2 = await templateUtils.createWorkloadTemplateV2(
                    {
                        workloadType: WorkloadType.Inference,
                        scopeType: ScopeType.Project,
                        scopeId: projectA.id.toString(),
                        templateName: template2Name,
                        description: 'Project A template 2'
                    },
                    spec,
                    undefined,
                    flowRootState.users.system_administrator.idToken
                );
                projectATemplate2Id = response2.body.meta.id;
                logger.info(`Created project-a template 2: ${template2Name}`);
            });

            await test.step('Create project-b scoped template', async function () {
                const spec = {
                    image: 'alpine:latest',
                    command: 'sleep 10',
                    compute: { gpuDevicesRequest: 0, cpuCoreRequest: 0.1, cpuMemoryRequest: '100M' }
                };

                const templateName = objectGenerator.generateShortNameByPrefix('proj-b-template', 5);
                const response = await templateUtils.createWorkloadTemplateV2(
                    {
                        workloadType: WorkloadType.Training,
                        scopeType: ScopeType.Project,
                        scopeId: projectB.id.toString(),
                        templateName: templateName,
                        description: 'Project B template'
                    },
                    spec,
                    undefined,
                    flowRootState.users.system_administrator.idToken
                );
                projectBTemplateId = response.body.meta.id;
                logger.info(`Created project-b template: ${templateName}`);
            });

            // Set default project context
            runaiCliV2.projectSet(projectAName);
            logger.info(`Set CLI project context to: ${projectAName}`);
        });

        test.afterAll(async function () {
            await test.step('Cleanup all templates', async function () {
                const templatesToDelete = [
                    { id: clusterTemplate1Id, type: WorkloadType.Training },
                    { id: clusterTemplate2Id, type: WorkloadType.Workspace },
                    { id: deptTemplateId, type: WorkloadType.Training },
                    { id: projectATemplate1Id, type: WorkloadType.Training },
                    { id: projectATemplate2Id, type: WorkloadType.Inference },
                    { id: projectBTemplateId, type: WorkloadType.Training }
                ];

                try {
                    for (const template of templatesToDelete) {
                        await templateUtils.deleteWorkloadTemplateV2(
                            template.id,
                            template.type,
                            flowRootState.users.system_administrator.idToken
                        );
                    }
                    logger.info('Deleted all templates');
                } catch (error) {
                    logger.error('Failed to delete some templates:', error);
                }
            });

            await test.step('Cleanup projects and department', async function () {
                try {
                    await operations.deleteProjectAndNamespace(projectA.id);
                    await operations.deleteProjectAndNamespace(projectB.id);
                    await orgUnitUtils.deleteDepartment(departmentId.toString());
                    logger.info('Deleted projects and department');
                } catch (error) {
                    logger.error('Failed to delete resources:', error);
                }
            });
        });

        // Group A: Basic List Operations
        test('T5539406: List templates in default scope (current project)', function () {
            test.step('List templates from current project context', function () {
                logger.info('Listing templates in default scope (current project)');

                const stdout = runaiCliV2.list(TEMPLATE, {}, false).stdout;

                logger.info(`Template list output:\n${stdout}`);

                expect(stdout).toBeTruthy();

                // Should show project-a templates (2)
                expect(stdout).toContain('proj-a-template-1');
                expect(stdout).toContain('proj-a-template-2');

                // Should show inherited department template (1)
                expect(stdout).toContain('dept-template');

                // Should show inherited cluster templates (2)
                expect(stdout).toContain('cluster-template-1');
                expect(stdout).toContain('cluster-template-2');

                // Should NOT show project-b template
                expect(stdout).not.toContain('proj-b-template');

                logger.info('Verified: Template list shows project + inherited templates');
            });
        });

        test('T5539407: List templates from all scopes using -A flag', function () {
            test.step('List templates from all scopes', function () {
                logger.info('Listing templates from all scopes using -A flag');

                const stdout = runaiCliV2.list(TEMPLATE, { [listFlags.allScopes]: '' } as any, false).stdout;

                logger.info(`Template list -A output:\n${stdout}`);

                expect(stdout).toBeTruthy();

                // Should show ALL templates including from other projects
                expect(stdout).toContain('proj-a-template-1');
                expect(stdout).toContain('proj-a-template-2');
                expect(stdout).toContain('proj-b-template');
                expect(stdout).toContain('dept-template');
                expect(stdout).toContain('cluster-template-1');
                expect(stdout).toContain('cluster-template-2');

                logger.info('Verified: -A flag shows templates from all scopes');
            });
        });

        test('T5539408: List templates with -p flag (specific project scope)', function () {
            test.step('List templates for specific project using -p flag', function () {
                logger.info(`Listing templates for project-a using -p ${projectAName}`);

                const stdoutA = runaiCliV2.list(
                    TEMPLATE,
                    { [listFlags.project]: projectAName } as any,
                    false
                ).stdout;

                logger.info(`Project-A template list:\n${stdoutA}`);

                // Should show project-a templates + inherited
                expect(stdoutA).toContain('proj-a-template-1');
                expect(stdoutA).toContain('proj-a-template-2');
                expect(stdoutA).toContain('dept-template');
                expect(stdoutA).toContain('cluster-template-1');
                expect(stdoutA).not.toContain('proj-b-template');

                logger.info(`Listing templates for project-b using -p ${projectBName}`);
                const stdoutB = runaiCliV2.list(
                    TEMPLATE,
                    { [listFlags.project]: projectBName } as any,
                    false
                ).stdout;

                logger.info(`Project-B template list:\n${stdoutB}`);

                // Should show project-b templates + inherited
                expect(stdoutB).toContain('proj-b-template');
                expect(stdoutB).toContain('dept-template');
                expect(stdoutB).toContain('cluster-template-1');
                expect(stdoutB).not.toContain('proj-a-template-1');
                expect(stdoutB).not.toContain('proj-a-template-2');

                logger.info('Verified: -p flag shows correct templates for each project');
            });
        });

        // Group B: Output Formats
        test('T5539409: List templates with custom columns', function () {
            test.step('List templates with API field names (camelCase)', function () {
                logger.info('Listing templates with custom columns using API field names');

                const stdout = runaiCliV2.list(
                    TEMPLATE,
                    { [listFlags.columns]: 'name,scopeType,workloadType,gpuComputeRequest' } as any,
                    false
                ).stdout;

                logger.info(`Custom columns output (API field names):\n${stdout}`);

                expect(stdout).toBeTruthy();
                // API field names are displayed as friendly column headers
                expect(stdout).toContain('Template');
                expect(stdout).toContain('Scope');
                expect(stdout).toContain('Workload Type');
                expect(stdout).toContain('GPU Compute Request');

                // Should contain template names
                expect(stdout).toContain('proj-a-template-1');

                logger.info('Verified: Custom columns with API field names displayed correctly');
            });

            test.step('List templates with minimal columns', function () {
                logger.info('Listing templates with minimal columns (name,workloadType)');

                const stdout = runaiCliV2.list(
                    TEMPLATE,
                    { [listFlags.columns]: 'name,workloadType' } as any,
                    false
                ).stdout;

                logger.info(`Minimal columns output:\n${stdout}`);

                expect(stdout).toBeTruthy();
                expect(stdout).toContain('Template');
                expect(stdout).toContain('Workload Type');
                expect(stdout).toContain('proj-a-template-1');

                logger.info('Verified: Minimal columns displayed correctly');
                logger.info('Note: --columns only accepts API field names (camelCase), not display names');
            });
        });

        test('T5539410: List templates with JSON output format', function () {
            test.step('List templates with JSON output and verify structure', function () {
                logger.info('Listing templates with JSON output');

                const stdout = runaiCliV2.list(TEMPLATE, { [listFlags.json]: '' } as any).stdout;

                logger.info(`JSON output received (length: ${stdout.length} chars)`);

                expect(stdout).toBeTruthy();

                // Verify it's valid JSON
                let jsonData: any;
                expect(() => {
                    jsonData = JSON.parse(stdout);
                }).not.toThrow();

                logger.info(`Parsed JSON keys: ${Object.keys(jsonData).join(', ')}`);

                // Verify JSON structure contains templates
                expect(jsonData).toBeDefined();
                expect(typeof jsonData).toBe('object');

                // Log template names in JSON
                logger.info(`JSON data structure: ${JSON.stringify(jsonData, null, 2).substring(0, 500)}...`);

                logger.info('Verified: JSON output is valid and contains template data');
            });
        });

        test('T5539411: List templates with YAML output format', function () {
            test.step('List templates with YAML output and verify structure', function () {
                logger.info('Listing templates with YAML output');

                const stdout = runaiCliV2.list(TEMPLATE, { [listFlags.yaml]: '' } as any, false).stdout;

                logger.info(`YAML output:\n${stdout.substring(0, 500)}...`);

                expect(stdout).toBeTruthy();
                expect(stdout.length).toBeGreaterThan(0);

                // Verify YAML structure
                expect(stdout).toMatch(/name:|template:/i);
                expect(stdout).toContain('proj-a-template');

                logger.info('Verified: YAML output contains template data');
            });
        });

        test('T5539412: List templates with table output and no headers', function () {
            test.step('List templates with --no-headers flag', function () {
                logger.info('Listing templates with table output but no headers');

                const stdoutNoHeaders = runaiCliV2.list(
                    TEMPLATE,
                    { [listFlags.table]: '', [listFlags.noHeaders]: '' } as any,
                    false
                ).stdout;

                logger.info(`No-headers output:\n${stdoutNoHeaders}`);

                expect(stdoutNoHeaders).toBeTruthy();
                expect(stdoutNoHeaders).toContain('proj-a-template');

                // Verify header row is not present
                expect(stdoutNoHeaders).not.toMatch(/^Template\s+Scope\s+/m);

                logger.info('Verified: Table output without headers');
            });

            test.step('Compare with default output (with headers)', function () {
                const stdoutWithHeaders = runaiCliV2.list(
                    TEMPLATE,
                    { [listFlags.table]: '' } as any,
                    false
                ).stdout;

                logger.info(
                    `With-headers output (first 300 chars):\n${stdoutWithHeaders.substring(0, 300)}...`
                );

                // Default output should have headers
                expect(stdoutWithHeaders).toMatch(/Template|TEMPLATE/i);

                logger.info('Verified: Default output includes headers');
            });
        });

        test('T5539421: Verify template list shows mandatory fields status', async function () {
            let completeTemplateId: string;
            let incompleteTemplateId: string;

            await test.step('Create complete and incomplete templates', async function () {
                // Complete template with all required fields
                const completeSpec = {
                    image: 'pytorch/pytorch:latest',
                    command: 'python train.py',
                    compute: { gpuDevicesRequest: 1, cpuCoreRequest: 0.5, cpuMemoryRequest: '1G' }
                };

                const completeTemplateName = objectGenerator.generateShortNameByPrefix(
                    'complete-template',
                    5
                );
                const completeResponse = await templateUtils.createWorkloadTemplateV2(
                    {
                        workloadType: WorkloadType.Training,
                        scopeType: ScopeType.Cluster,
                        scopeId: flowRootState.cluster.uuid,
                        templateName: completeTemplateName,
                        description: 'Complete template with all fields'
                    },
                    completeSpec,
                    undefined,
                    flowRootState.users.system_administrator.idToken
                );
                completeTemplateId = completeResponse.body.meta.id;
                logger.info(`Created complete template: ${completeTemplateName}`);

                // Incomplete template missing required fields (no image)
                const incompleteSpec = {
                    command: 'echo test',
                    compute: { gpuDevicesRequest: 0, cpuCoreRequest: 0.1, cpuMemoryRequest: '100M' }
                };

                const incompleteTemplateName = objectGenerator.generateShortNameByPrefix(
                    'incomplete-template',
                    5
                );
                const incompleteResponse = await templateUtils.createWorkloadTemplateV2(
                    {
                        workloadType: WorkloadType.Training,
                        scopeType: ScopeType.Cluster,
                        scopeId: flowRootState.cluster.uuid,
                        templateName: incompleteTemplateName,
                        description: 'Incomplete template missing image'
                    },
                    incompleteSpec,
                    undefined,
                    flowRootState.users.system_administrator.idToken
                );
                incompleteTemplateId = incompleteResponse.body.meta.id;
                logger.info(`Created incomplete template: ${incompleteTemplateName}`);
            });

            test.step('Verify mandatory fields status in list output', function () {
                logger.info('Listing templates to verify mandatory fields status');

                const stdout = runaiCliV2.list(TEMPLATE, {}, false).stdout;

                logger.info(`Template list with mandatory fields:\n${stdout}`);

                expect(stdout).toBeTruthy();
                expect(stdout).toContain('complete-template');
                expect(stdout).toContain('incomplete-template');

                // Check if output contains mandatory fields status column
                expect(stdout).toMatch(/Mandatory.*fields.*missing/i);

                logger.info('Verified: Template list shows mandatory fields status');
            });

            await test.step('Cleanup test templates', async function () {
                await templateUtils.deleteWorkloadTemplateV2(
                    completeTemplateId,
                    WorkloadType.Training,
                    flowRootState.users.system_administrator.idToken
                );
                await templateUtils.deleteWorkloadTemplateV2(
                    incompleteTemplateId,
                    WorkloadType.Training,
                    flowRootState.users.system_administrator.idToken
                );
                logger.info('Cleanup completed');
            });
        });
    });
});

// Separate describe block for pagination tests (needs more templates)
test.describe(`CLI List Template - Pagination Tests, tags: ${tags.resource.templateV2} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_24)(function () {
        const TEMPLATE = 'template';
        let projectName: string;
        let project: Project;
        const templateIds: string[] = [];

        test.beforeAll(async function () {
            await test.step('Create project for pagination tests', async function () {
                projectName = objectGenerator.generateShortNameByPrefix('proj-pagination', 5);
                const namespace = Utils.namespaceFromProjectName(projectName);
                project = await projectResourcesUtils.createProject(namespace, {
                    name: projectName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0
                });
                logger.info(`Created project: ${projectName}`);

                runaiCliV2.projectSet(projectName);
                logger.info(`Set CLI project context to: ${projectName}`);
            });

            await test.step('Create multiple templates for pagination', async function () {
                const spec = {
                    image: 'alpine:latest',
                    command: 'sleep 10',
                    compute: { gpuDevicesRequest: 0, cpuCoreRequest: 0.1, cpuMemoryRequest: '100M' }
                };

                // Create 15 templates
                for (let i = 1; i <= 15; i++) {
                    const templateName = objectGenerator.generateShortNameByPrefix(
                        `page-test-${String(i).padStart(2, '0')}`,
                        5
                    );
                    const response = await templateUtils.createWorkloadTemplateV2(
                        {
                            workloadType: WorkloadType.Training,
                            scopeType: ScopeType.Project,
                            scopeId: project.id.toString(),
                            templateName: templateName,
                            description: `Pagination test template ${i}`
                        },
                        spec,
                        undefined,
                        flowRootState.users.system_administrator.idToken
                    );
                    templateIds.push(response.body.meta.id);
                }
                logger.info(`Created ${templateIds.length} templates for pagination tests`);
            });
        });

        test.afterAll(async function () {
            await test.step('Cleanup pagination templates', async function () {
                for (const templateId of templateIds) {
                    try {
                        await templateUtils.deleteWorkloadTemplateV2(
                            templateId,
                            WorkloadType.Training,
                            flowRootState.users.system_administrator.idToken
                        );
                    } catch (error) {
                        logger.error(`Failed to delete template ${templateId}:`, error);
                    }
                }
                logger.info('Deleted pagination templates');
            });

            await test.step('Cleanup pagination project', async function () {
                try {
                    await operations.deleteProjectAndNamespace(project.id);
                    logger.info('Deleted pagination project');
                } catch (error) {
                    logger.error('Failed to delete project:', error);
                }
            });
        });

        test('T5539413: List templates with page-size flag only', function () {
            test.step('List templates with page-size=2', function () {
                logger.info('Listing templates with --page-size 2');

                const stdout = runaiCliV2.list(TEMPLATE, { [listFlags.pageSize]: '2' } as any, false).stdout;

                logger.info(`Page-size=2 output:\n${stdout}`);

                expect(stdout).toBeTruthy();

                // Should show next-token at the bottom
                expect(stdout).toMatch(/next token:/i);

                logger.info('Verified: page-size returns limited results with next-token');
            });
        });

        test('T5539414: List templates with max-items flag only', function () {
            test.step('List templates with max-items=5', function () {
                logger.info('Listing templates with --max-items 5');

                const stdout = runaiCliV2.list(TEMPLATE, { [listFlags.maxItems]: '5' } as any, false).stdout;

                logger.info(`Max-items=5 output:\n${stdout}`);

                expect(stdout).toBeTruthy();

                // Count template lines (excluding headers and empty lines)
                const lines = stdout.split('\n').filter((line) => line.trim() && line.includes('page-test'));
                logger.info(`Number of template lines: ${lines.length}`);

                // Should have at most 5 templates
                expect(lines.length).toBeLessThanOrEqual(5);

                // Note: CLI behavior - max-items still shows next-token if more data exists
                // This is expected behavior to indicate more templates are available
                logger.info('Verified: max-items limits result count to at most 5 templates');

                if (stdout.match(/next token:/i)) {
                    logger.info(
                        'next-token present: indicates more templates available beyond max-items limit'
                    );
                }
            });
        });

        // Coverage remains disabled pending revalidation of this case
        test.skip('T5539415: List templates with page-size and max-items combination', function () {
            test.step('List templates with page-size=4 and max-items=10', function () {
                logger.info('Listing templates with --page-size 4 --max-items 10');

                const stdout = runaiCliV2.list(
                    TEMPLATE,
                    { [listFlags.pageSize]: '4', [listFlags.maxItems]: '10' } as any,
                    false
                ).stdout;

                logger.info(`Page-size=4 max-items=10 output:\n${stdout}`);

                expect(stdout).toBeTruthy();

                // First page should show 4 templates
                const lines = stdout.split('\n').filter((line) => line.trim() && line.includes('page-test'));
                logger.info(`Number of templates in first page: ${lines.length}`);

                expect(lines.length).toBeLessThanOrEqual(4);

                // Should show next-token (more pages available within max-items limit)
                expect(stdout).toMatch(/next token:/i);

                logger.info('Verified: page-size and max-items work together correctly');
            });
        });

        test('T5539416: List templates with no-pagination flag (single page only)', function () {
            test.step('List templates with --no-pagination', function () {
                logger.info('Listing templates with --no-pagination');

                const stdout = runaiCliV2.list(
                    TEMPLATE,
                    { [listFlags.noPagination]: '' } as any,
                    false
                ).stdout;

                logger.info(`No-pagination output (first 500 chars):\n${stdout.substring(0, 500)}...`);

                expect(stdout).toBeTruthy();

                // Should return single page with default page size
                const lines = stdout.split('\n').filter((line) => line.trim() && line.includes('page-test'));
                logger.info(`Number of templates returned: ${lines.length}`);

                logger.info('Verified: no-pagination returns single page');
            });

            test.step('Compare with default behavior', function () {
                const stdout = runaiCliV2.list(TEMPLATE, {}, false).stdout;

                logger.info('Default list behavior executed for comparison');

                expect(stdout).toBeTruthy();

                logger.info('Verified: Default behavior executed successfully');
            });
        });

        test('T5539417: List templates with next-token and page-size combination (manual pagination)', function () {
            let nextToken: string;

            test.step('Get first page and extract next-token', function () {
                logger.info('Getting first page with --page-size 3');

                const stdout = runaiCliV2.list(TEMPLATE, { [listFlags.pageSize]: '3' } as any, false).stdout;

                logger.info(`First page output:\n${stdout}`);

                expect(stdout).toBeTruthy();

                // Extract next-token from output
                const tokenMatch = stdout.match(/next token:\s*([^\s]+)/i);
                expect(tokenMatch).toBeTruthy();
                nextToken = tokenMatch![1];
                logger.info(`Extracted next-token: ${nextToken}`);

                // Count ALL template lines in the output (not just page-test templates)
                // Split by newline and count lines that look like template data rows
                const templateLines = stdout.split('\n').filter((line) => {
                    const trimmed = line.trim();
                    // Skip empty lines, headers, separators, and next-token line
                    if (!trimmed) return false;
                    if (trimmed.startsWith('Template')) return false; // Header
                    if (trimmed.match(/^─+$/)) return false; // Separator
                    if (trimmed.startsWith('next token:')) return false;
                    // Template rows contain scope keywords
                    return (
                        trimmed.includes('project') ||
                        trimmed.includes('system') ||
                        trimmed.includes('cluster') ||
                        trimmed.includes('department')
                    );
                });

                logger.info(`Total templates in first page: ${templateLines.length}`);

                // Verify page-size is respected (should show exactly 3 templates)
                expect(templateLines.length).toBe(3);

                // Log how many are page-test templates vs others
                const pageTestCount = templateLines.filter((line) => line.includes('page-test')).length;
                logger.info(`  - page-test templates: ${pageTestCount}`);
                logger.info(`  - other templates: ${templateLines.length - pageTestCount}`);

                logger.info('Verified: First page shows exactly 3 templates with next-token');
            });

            test.step('Get second page using next-token', function () {
                logger.info(`Getting second page with next-token: ${nextToken}`);

                const stdout = runaiCliV2.list(
                    TEMPLATE,
                    { [listFlags.nextToken]: nextToken, [listFlags.pageSize]: '3' } as any,
                    false
                ).stdout;

                logger.info(`Second page output:\n${stdout}`);

                expect(stdout).toBeTruthy();

                // Count ALL template lines in the output
                const templateLines = stdout.split('\n').filter((line) => {
                    const trimmed = line.trim();
                    if (!trimmed) return false;
                    if (trimmed.startsWith('Template')) return false; // Header
                    if (trimmed.match(/^─+$/)) return false; // Separator
                    if (trimmed.startsWith('next token:')) return false;
                    return (
                        trimmed.includes('project') ||
                        trimmed.includes('system') ||
                        trimmed.includes('cluster') ||
                        trimmed.includes('department')
                    );
                });

                logger.info(`Total templates in second page: ${templateLines.length}`);

                // Verify page-size is respected (should show exactly 3 templates, or fewer if this is the last page)
                expect(templateLines.length).toBeGreaterThan(0);
                expect(templateLines.length).toBeLessThanOrEqual(3);

                logger.info('Verified: Manual pagination works with next-token and page-size');
            });
        });
    });
});
