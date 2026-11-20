/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { flowRootState, ngcCatalogUtils, tags } from '../../../../../../modules/helpers';
import { test } from '@playwright/test';
import { expect } from 'chai';
import { workloadsCrudService } from '../../../../../../modules/api';

test.describe(`inference workload ngc catalog api tests: ${tags.teams.viper}`, function () {
    test.describe.configure({ retries: 1 });
    test.beforeEach(function () {
        workloadsCrudService.inferencesApi.accessToken = flowRootState.users.ml_engineer.accessToken;
    });

    test('get NGC catalog with default parameters', async function () {
        const response = await ngcCatalogUtils.callNgcCatalog();
        const catalog = ngcCatalogUtils.validateBasicResponse(response);

        expect(catalog.resources.length).to.be.greaterThan(0);
        const firstResource = catalog.resources[0];

        expect(firstResource.resourceId).to.be.a('string');
        expect(firstResource.orgName).to.be.a('string');

        // Check first resource structure
        ngcCatalogUtils.validateResourceStructure(firstResource);

        // Log which data source is being used
        ngcCatalogUtils.logDataSource(catalog);
    });

    test('get NGC catalog with pagination parameters', async function () {
        const response = await ngcCatalogUtils.callNgcCatalog({
            page: 0,
            pageSize: 10
        });

        const catalog = ngcCatalogUtils.validateBasicResponse(response);

        // Check that we get resources (may be limited by pageSize)
        expect(catalog.resources).to.be.an('array');
        expect(catalog.resources.length).to.be.lessThanOrEqual(10);
    });

    test('get NGC catalog with orderBy parameter', async function () {
        const response = await ngcCatalogUtils.callNgcCatalog({
            orderBy: 'name,ASC',
            pageSize: 20
        });

        const catalog = ngcCatalogUtils.validateBasicResponse(response);

        // Check that we get resources
        expect(catalog.resources).to.be.an('array');
        expect(catalog.resultTotal).to.be.a('number');
    });

    test('get NGC catalog with complex query and pagination', async function () {
        const response = await ngcCatalogUtils.callNgcCatalog({
            query: 'name:*llama*',
            orderBy: 'dateCreated,DESC',
            page: 1,
            pageSize: 5
        });

        const catalog = ngcCatalogUtils.validateBasicResponse(response);

        // Check that we get resources
        expect(catalog.resources).to.be.an('array');
        expect(catalog.resultTotal).to.be.a('number');

        ngcCatalogUtils.validateSearchResults(catalog, 'llama');
    });

    test('get NGC catalog with page beyond available data', async function () {
        const response = await ngcCatalogUtils.callNgcCatalog({
            page: 1000,
            pageSize: 10
        });

        ngcCatalogUtils.validateBasicResponse(response, true);
    });

    test('verify catalog resource structure', async function () {
        const response = await ngcCatalogUtils.callNgcCatalog({
            pageSize: 1
        });

        const catalog = ngcCatalogUtils.validateBasicResponse(response);
        const firstResource = ngcCatalogUtils.validateCatalogStructure(catalog);

        // Verify the structure of the first resource
        ngcCatalogUtils.validateResourceStructure(firstResource);
    });

    test('should handle real NGC API pagination correctly', async function () {
        // Test first page
        const firstPageResponse = await ngcCatalogUtils.callNgcCatalog({
            page: 0,
            pageSize: 5
        });

        const firstPage = ngcCatalogUtils.validateBasicResponse(firstPageResponse);

        // Test second page
        const secondPageResponse = await ngcCatalogUtils.callNgcCatalog({
            page: 1,
            pageSize: 5
        });

        const secondPage = ngcCatalogUtils.validateBasicResponse(secondPageResponse);

        // Validate pagination behavior
        ngcCatalogUtils.validatePagination(firstPage, secondPage);
    });

    test('should handle real NGC API filtering correctly', async function () {
        const response = await ngcCatalogUtils.callNgcCatalog({
            filter: ['name:nvidia'],
            pageSize: 10
        });

        ngcCatalogUtils.validateBasicResponse(response, true);
    });

    test('should handle real NGC API search queries correctly', async function () {
        const response = await ngcCatalogUtils.callNgcCatalog({
            query: 'name:*llama*',
            pageSize: 10
        });

        const catalog = ngcCatalogUtils.validateBasicResponse(response);

        // Validate search behavior
        ngcCatalogUtils.validateSearchResults(catalog, 'llama');
    });

    test('should provide consistent API response structure regardless of data source', async function () {
        const response = await ngcCatalogUtils.callNgcCatalog();
        const catalog = ngcCatalogUtils.validateBasicResponse(response);

        // The API should always return the same structure whether using NGC API or fallback
        const firstResource = ngcCatalogUtils.validateCatalogStructure(catalog);

        // Resources should have consistent structure
        ngcCatalogUtils.validateResourceStructure(firstResource);

        console.log(`✅ API response structure is consistent (isStaticData: ${catalog.isStaticData})`);
    });
});
