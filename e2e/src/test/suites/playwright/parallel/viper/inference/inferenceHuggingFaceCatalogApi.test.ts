/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { flowRootState, tags } from '../../../../../../modules/helpers';
import { test } from '@playwright/test';
import { expect } from 'chai';
import { catalogApi } from '../../../../../../modules/api';
import type {
    HuggingFaceModelItem,
    HuggingFaceModelSearchResults
} from '../../../../../../modules/api/apis/catalogApi';
import type { IncomingMessage } from 'http';

test.describe(
    `inference workload huggingface catalog api tests: ${tags.resource.inferenceWorkload} ${tags.teams.viper}`,
    { tag: [tags.resource.inferenceWorkload] },
    function () {
        test.describe.configure({ retries: 1 });
        // Test constants
        const EXPECTED_DEFAULT_CATALOG_SIZE = 100; // Default page size returned by HuggingFace catalog API
        const TEST_MODEL_SENTENCE_TRANSFORMER = 'sentence-transformers/all-MiniLM-L6-v2';
        const TEST_MODEL_PYANNOTE = 'pyannote/segmentation-3.0';

        // Test case types
        interface GatedStatusTestCase {
            modelName: string;
            expectedGated: boolean;
        }

        /**
         * Validates the catalog API response structure and returns the catalog data
         */
        function validateCatalogResponse(response: {
            response: IncomingMessage;
            body: HuggingFaceModelSearchResults;
        }): HuggingFaceModelSearchResults {
            expect(response).to.not.be.undefined;
            expect(response.response.statusCode).to.equal(200);
            expect(response.body).to.not.be.undefined;
            expect(response.body).to.have.property('resources');
            expect(response.body.resources).to.be.an('array');
            return response.body;
        }

        /**
         * Validates the structure of a HuggingFace model item
         * Checks required properties: id (string), downloads (number)
         * Checks optional properties: gated (boolean if present)
         */
        function validateModelStructure(model: HuggingFaceModelItem): void {
            expect(model).to.not.be.undefined;

            // Required properties
            expect(model).to.have.property('id');
            expect(model.id).to.be.a('string');
            expect(model.id.length).to.be.greaterThan(0);

            expect(model).to.have.property('downloads');
            expect(model.downloads).to.be.a('number');
            expect(model.downloads).to.be.greaterThanOrEqual(0);

            // Optional properties
            if ('gated' in model && model.gated !== undefined) {
                expect(model.gated).to.be.a('boolean');
            }
        }

        /**
         * Validates the gated status of a model
         */
        function validateGatedStatus(model: HuggingFaceModelItem, expectedGated: boolean): void {
            if (expectedGated) {
                expect(model.gated).to.equal(expectedGated, `Model ${model.id} gated status mismatch`);
            } else {
                expect(model.gated).to.be.oneOf([false, undefined]);
            }
        }
        test.beforeEach(function () {
            catalogApi.inferencesApi.accessToken = flowRootState.users.ml_engineer.accessToken;
        });

        test('should return default catalog with 100 models', async function () {
            const response = await catalogApi.inferencesApi.searchHuggingFaceCatalog();
            const catalog = validateCatalogResponse(response);

            expect(catalog.resources.length).to.equal(EXPECTED_DEFAULT_CATALOG_SIZE);
            validateModelStructure(catalog.resources[0]);
        });

        test('should filter catalog by query parameter', async function () {
            const response = await catalogApi.inferencesApi.searchHuggingFaceCatalog(
                TEST_MODEL_SENTENCE_TRANSFORMER
            );
            const catalog = validateCatalogResponse(response);

            expect(catalog.resources.length).to.be.greaterThan(0);
            validateModelStructure(catalog.resources[0]);

            // Verify search results contain the search term
            const matchingModels = catalog.resources.filter((model) =>
                model.id.toLowerCase().includes(TEST_MODEL_SENTENCE_TRANSFORMER.toLowerCase())
            );
            expect(matchingModels.length).to.be.greaterThan(
                0,
                `No models found matching search term: ${TEST_MODEL_SENTENCE_TRANSFORMER}`
            );
        });

        test('should correctly identify gated and non-gated models', async function () {
            const testCases: GatedStatusTestCase[] = [
                { modelName: TEST_MODEL_SENTENCE_TRANSFORMER, expectedGated: false },
                { modelName: TEST_MODEL_PYANNOTE, expectedGated: true }
            ];

            for (const testCase of testCases) {
                const response = await catalogApi.inferencesApi.searchHuggingFaceCatalog(testCase.modelName);
                const catalog = validateCatalogResponse(response);

                expect(catalog.resources.length).to.be.greaterThan(0);

                const model = catalog.resources.find((m) => m.id === testCase.modelName);
                expect(model, `Should find model: ${testCase.modelName}`).to.not.be.undefined;

                validateModelStructure(model!);
                validateGatedStatus(model!, testCase.expectedGated);
            }
        });
    }
);
