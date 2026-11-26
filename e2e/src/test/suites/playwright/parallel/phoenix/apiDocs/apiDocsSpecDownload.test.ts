/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import axios from 'axios';
import * as yaml from 'js-yaml';
import { test, expect } from '@playwright/test';
import { config, tags } from '../../../../../../modules/helpers';

interface OpenApiSpec {
    openapi?: string;
    paths?: Record<string, unknown>;
}

// /api/docs serves the OpenAPI yaml spec download in every environment (it does not redirect to an external docs site).
test.describe(`API docs spec download, tags: ${tags.tool.api} ${tags.teams.phoenix} ${tags.suites.legacySphinx}`, function () {
    test('Verify /api/docs serves the OpenAPI yaml spec download', async () => {
        const response = await axios.get(`${config.appUrl}/api/docs`, {
            timeout: 30000,
            maxRedirects: 0,
            responseType: 'text',
            transformResponse: [(data: unknown): unknown => data],
            validateStatus: (): boolean => true
        });

        expect(response.status, '/api/docs should serve the spec directly, not redirect').toBe(200);
        expect(String(response.headers['content-type']), '/api/docs should serve yaml').toContain('yaml');

        const spec = yaml.load(response.data) as OpenApiSpec;
        expect(spec.openapi, 'the download should be an OpenAPI 3 document').toMatch(/^3\./);
        expect(Object.keys(spec.paths ?? {}).length, 'the spec should list API paths').toBeGreaterThan(0);
    });
});
