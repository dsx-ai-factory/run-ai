/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { logger, objectGenerator, tags, waitUtils, flowRootState } from '../../../../../../modules/helpers';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { crds, CustomResourceDefinition } from '../../../../../../modules/helpers';
import { kubeCli } from '../../../../../../modules/cli';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { ClusterPage } from '../../../../../../modules/ui/runaiV2/pages/ClusterPage';
import { NetworkTopologyModal } from '../../../../../../modules/ui/runaiV2/pages/NetworkTopologyModal';

test.describe(`Network Topology Kubectl Tests, tags: ${tags.tool.kubeCli} ${tags.resource.networkTopologies} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_23)(function () {
        let browserEngine: BrowserEngine;
        let authPage: AuthPage;
        let clusterPage: ClusterPage;
        let topoModal: NetworkTopologyModal;
        let videoName: string;
        let deletedFalg: boolean = false;
        let topoName: string;
        let topoCrd: CustomResourceDefinition;
        let topoBody: object;

        test.beforeAll(async function () {
            browserEngine = new BrowserEngine();
            await browserEngine.start();
            authPage = new AuthPage(browserEngine.currentPage);
            clusterPage = new ClusterPage(browserEngine.currentPage);
            topoModal = new NetworkTopologyModal(browserEngine.currentPage);

            topoName = objectGenerator.generateShortNameByPrefix('kubectl-topo', 5);
            topoCrd = crds.topologyCrd;
            topoBody = {
                apiVersion: 'kai.scheduler/v1alpha1',
                kind: 'Topology',
                metadata: {
                    name: topoName
                },
                spec: {
                    levels: [
                        { nodeLabel: 'topology.kubernetes.io/rack' },
                        { nodeLabel: 'topology.kubernetes.io/region' }
                    ]
                }
            };

            await authPage.signIn(
                flowRootState.users.system_administrator.email,
                flowRootState.users.system_administrator.password
            );
        });

        test.afterAll('stop browser', async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, videoName);
            if (!deletedFalg) {
                try {
                    await kubeCli.deleteCrd(topoCrd, topoName);
                    await kubeCli.waitForCrdToBeDeleted(topoCrd, topoName, undefined, 60);
                } catch (error) {
                    logger.error(`failed to delete topo ${topoName}`);
                }
            }
        });

        test('T5404988: test submitting a network topolog CR directly to the cluster', async function () {
            videoName = 'sumbit-topo-crd-check-on-ui';

            await test.step('Submit a network topology CR to the cluster', async function () {
                await kubeCli.createCrd(topoCrd, topoBody);
                await waitUtils.waitForCrd(topoCrd, topoName, undefined, 60);
            });

            await test.step('Open cluster details page and verify the topo can be seen', async function () {
                await clusterPage.goToPage();
                await clusterPage.openNetworkTopoCreationModal(flowRootState.cluster.name);
                const topoList = await topoModal.get_topo_name_list();
                logger.info(`topo list: ${topoList}`);
                expect(topoList.some((item) => item === topoName)).toBeTruthy();
            });

            await test.step('Delete the topology CR', async function () {
                await kubeCli.deleteCrd(topoCrd, topoName);
                await kubeCli.waitForCrdToBeDeleted(topoCrd, topoName, undefined, 60);
                deletedFalg = true;
            });
        });
    });
});
