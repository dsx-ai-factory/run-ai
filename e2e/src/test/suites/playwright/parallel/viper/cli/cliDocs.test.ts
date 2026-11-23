/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { cliExposerApi } from '../../../../../../modules/api';
import { StatusCodes } from 'http-status-codes';
import { tags } from '../../../../../../modules/helpers';

import { test, expect } from '@playwright/test';

interface docsParams {
    docFileName: string;
    expectedString: string;
}

const mainDocFileName: docsParams = {
    docFileName: 'runai.md',
    expectedString: 'Run:ai Command-line Interface'
};

const pfDocFileName: docsParams = {
    docFileName: 'runai_workspace_port-forward.md',
    expectedString: 'runai workspace port-forward'
};

const invalidDocFileName: docsParams = {
    docFileName: 'invalid-file.md',
    expectedString: 'getManualDocument should have failed with invalid doc name.'
};

test.describe(
    `Runai CLI V2 Docs  ${tags.teams.viper}`,
    { tag: [tags.runOn.parallelCI, tags.services.runaiExposer] },
    function () {
        [mainDocFileName, pfDocFileName].forEach(function (doc) {
            test(`verify cli docs ${doc.docFileName}`, async function () {
                const docsResponse = await cliExposerApi.cliApi.getManualDocument(doc.docFileName);
                expect(
                    docsResponse.response.statusCode,
                    `expecting response status code to be ${StatusCodes.OK} instead got ${docsResponse.response.statusCode}`
                ).toBe(200);
                expect(
                    docsResponse.body.includes(doc.expectedString),
                    `expecting the body to contain "${doc.expectedString}"`
                ).toBeTruthy();
            });
        });

        test('validate response on invalid cli doc file', async function () {
            try {
                await cliExposerApi.cliApi.getManualDocument(invalidDocFileName.docFileName);
                throw new Error(invalidDocFileName.expectedString);
            } catch (error) {
                expect(error.message, invalidDocFileName.expectedString).not.toBe(
                    invalidDocFileName.expectedString
                );
                expect(
                    error.statusCode,
                    `expecting response status code to be ${StatusCodes.NOT_FOUND} instead got ${error.statusCode}`
                ).toBe(StatusCodes.NOT_FOUND);
                expect(error.body).toBe(`Manual document does not exist: ${invalidDocFileName.docFileName}`);
            }
        });
    }
);
