/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { flowRootState, tags } from '../../../../../../modules/helpers';
import { HuggingFaceModelInfoRequest } from '../../../../../../modules/api/apis/assetsApi';
import { assetsApi } from '../../../../../../modules/api';
import { expect } from 'chai';
import { test } from '@playwright/test';

// some of the tests are using and environment variable process.env.HUGGING_FACE_TOKEN, which is the hugging face token,
// required to access gated models.
// it is taken as a github repo secret called HUGGING_FACE_TOKEN_FOR_E2E.
// the token allows access to the model: meta-llama/Llama-2-7b-chat-hf
// but doesn't allow access to the model: mistralai/Codestral-22B-v0.1
test.describe(
    `inference workload hugging face model info v2 tests: ${tags.resource.inferenceWorkload} ${tags.teams.viper}`,
    { tag: [tags.resource.inferenceWorkload] },
    function () {
        test.describe.configure({ retries: 1 });
        test.beforeEach(function () {
            assetsApi.inferenceApi.accessToken = flowRootState.users.ml_engineer.accessToken;
        });

        test('get model info for a non-gated model', async function () {
            const request = new HuggingFaceModelInfoRequest();
            request.modelName = 'openai-community/gpt2';

            const response = await assetsApi.inferenceApi.getHuggingFaceModelInfo(request);
            expect(response).to.not.be.undefined;
            expect(response.response.statusCode).to.equal(200);
            expect(response.body).to.not.be.undefined;
            const modelInfo = response.body;
            expect(modelInfo.modelName).to.equal(request.modelName);
            expect(modelInfo.gated).to.be.undefined;
            expect(modelInfo.authorized).to.equal(true);
            expect(modelInfo.language).to.equal(true);
        });

        test('get model info for a gated model without providing a token', async function () {
            const request = new HuggingFaceModelInfoRequest();
            request.modelName = 'meta-llama/Llama-2-7b-chat-hf';

            const response = await assetsApi.inferenceApi.getHuggingFaceModelInfo(request);
            expect(response).to.not.be.undefined;
            expect(response.response.statusCode).to.equal(200);
            expect(response.body).to.not.be.undefined;
            const modelInfo = response.body;
            expect(modelInfo.modelName).to.equal(request.modelName);
            expect(modelInfo.gated).to.equal(true);
            expect(modelInfo.authorized).to.be.undefined;
            expect(modelInfo.language).to.equal(true);
        });

        test('get model info for a gated model with an authorized token', async function () {
            const request = new HuggingFaceModelInfoRequest();
            request.modelName = 'meta-llama/Llama-2-7b-chat-hf';
            request.hfToken = process.env.HUGGING_FACE_TOKEN; // token is authorized for this model

            const response = await assetsApi.inferenceApi.getHuggingFaceModelInfo(request);
            expect(response).to.not.be.undefined;
            expect(response.response.statusCode).to.equal(200);
            expect(response.body).to.not.be.undefined;
            const modelInfo = response.body;
            expect(modelInfo.modelName).to.equal(request.modelName);
            expect(modelInfo.gated).to.equal(true);
            expect(modelInfo.authorized).to.equal(true);
            expect(modelInfo.language).to.equal(true);
        });

        test('get model info for a gated model with an unauthorized token', async function () {
            const request = new HuggingFaceModelInfoRequest();
            request.modelName = 'Sunbird/Sunflower-14B';
            request.hfToken = process.env.HUGGING_FACE_TOKEN; // token is NOT authorized to this model

            const response = await assetsApi.inferenceApi.getHuggingFaceModelInfo(request);
            expect(response).to.not.be.undefined;
            expect(response.response.statusCode).to.equal(200);
            expect(response.body).to.not.be.undefined;
            const modelInfo = response.body;
            expect(modelInfo.modelName).to.equal(request.modelName);
            expect(modelInfo.gated).to.equal(true);
            expect(modelInfo.authorized).to.be.undefined;
            expect(modelInfo.language).to.equal(true);
        });

        test('get model info for a gated model with an invalid token', async function () {
            const request = new HuggingFaceModelInfoRequest();
            request.modelName = 'meta-llama/Llama-2-7b-chat-hf';
            request.hfToken = 'hf_invalid'; // token is not recognized by HF

            const response = await assetsApi.inferenceApi.getHuggingFaceModelInfo(request);
            expect(response).to.not.be.undefined;
            expect(response.response.statusCode).to.equal(200);
            expect(response.body).to.not.be.undefined;
            const modelInfo = response.body;
            expect(modelInfo.modelName).to.equal(request.modelName);
            expect(modelInfo.gated).to.equal(true);
            expect(modelInfo.authorized).to.be.undefined;
            expect(modelInfo.language).to.equal(true);
        });

        test('get model info for a non-existing model', async function () {
            const request = new HuggingFaceModelInfoRequest();
            request.modelName = 'no-such/model';

            try {
                await assetsApi.inferenceApi.getHuggingFaceModelInfo(request);
                expect(0).to.equal(1); // should not reach here
            } catch (e) {
                expect(e.response.statusCode).to.equal(400);
                expect(e.body).to.not.be.undefined;
                expect(e.body.message).to.not.be.undefined;
                expect(e.body.message).to.equal('the provided model was not found on hugging face');
            }
        });

        test('get model info for a non-llm model', async function () {
            const request = new HuggingFaceModelInfoRequest();
            request.modelName = 'microsoft/resnet-50';

            const response = await assetsApi.inferenceApi.getHuggingFaceModelInfo(request);
            expect(response).to.not.be.undefined;
            expect(response.response.statusCode).to.equal(200);
            expect(response.body).to.not.be.undefined;
            const modelInfo = response.body;
            expect(modelInfo.modelName).to.equal(request.modelName);
            expect(modelInfo.language).to.be.undefined;
        });
    }
);
