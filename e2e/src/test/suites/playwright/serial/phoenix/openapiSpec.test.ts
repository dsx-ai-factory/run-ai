/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { config, settings, tags } from '../../../../../modules/helpers';

test.describe(
    `OpenAPI spec endpoint ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.api} ${tags.services.redoc}`,
    () => {
        test('GET /api/docs should return the raw OpenAPI spec as a YAML download on self-hosted', async () => {
            test.skip(!settings.isSelfHosted, 'On SaaS, /api/docs redirects to the external docs site');

            const response = await fetch(`${config.tenantUrl}/api/docs`);

            expect(response.status).toBe(200);

            const contentType = response.headers.get('content-type') ?? '';
            expect(contentType).toContain('application/yaml');

            const contentDisposition = response.headers.get('content-disposition') ?? '';
            expect(contentDisposition).toContain('attachment');
            expect(contentDisposition).toContain('runai-api.yaml');

            const body = await response.text();
            expect(body).toContain('openapi:');
        });

        test('GET /api/docs should redirect to the external docs site on SaaS', async () => {
            test.skip(settings.isSelfHosted, 'On self-hosted, /api/docs serves the raw YAML spec directly');

            const response = await fetch(`${config.tenantUrl}/api/docs`, { redirect: 'manual' });

            expect([301, 302]).toContain(response.status);

            const location = response.headers.get('location') ?? '';
            expect(location).toMatch(/^https?:\/\//);
        });
    }
);
