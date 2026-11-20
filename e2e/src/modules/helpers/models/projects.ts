/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Project } from '../../api/apis/backendUndocumentedApi';
import { logger, objectGenerator, projectResourcesUtils, Utils } from '../index';

export async function createProject(projectNameSuffix: string, projectOptions: object): Promise<Project> {
    const projectName = objectGenerator.generateNameByPrefix(`p-${projectNameSuffix}`);
    const project = await projectResourcesUtils.createProject(Utils.namespaceFromProjectName(projectName), {
        name: projectName,
        ...projectOptions
    });
    logger.info(`created project with id ${project.id} and name ${project.name}`);
    return project;
}
