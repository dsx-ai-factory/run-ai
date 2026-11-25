/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    authUtils,
    flowRootState,
    logger,
    objectGenerator,
    serviceAccountUtils,
    tags
} from '../../../../../../modules/helpers';
import { ServiceAccountPostResponse } from '../../../../../../modules/api/apis/identityManagerApi/generated/api';
import { IncomingMessage } from 'http';
import { expect } from 'chai';
import { StatusCodes } from 'http-status-codes';
import { test } from '@playwright/test';

test.describe(`Service Accounts Advanced API tests tags: ${tags.runOn.k8sFlavCertKit} ${tags.runOn.k8sFlavCertKitTemp} ${tags.teams.phoenix} ${tags.suites.legacySphinx}`, function () {
    let serviceAccount: { response: IncomingMessage; body: ServiceAccountPostResponse };
    const createdServiceAccounts: string[] = [];

    test.afterEach(async function () {
        for (const id of createdServiceAccounts) {
            try {
                await serviceAccountUtils.deleteServiceAccount(
                    flowRootState.users.system_administrator.accessToken,
                    id
                );
                logger.debug(`Cleaned up service account: ${id}`);
            } catch (e) {
                logger.error(`Failed to cleanup service account ${id}: ${e}`);
            }
        }
        createdServiceAccounts.length = 0;
    });

    test('should authenticate using service account clientId and clientSecret', async function () {
        serviceAccount = await serviceAccountUtils.createServiceAccount(
            flowRootState.users.system_administrator.accessToken,
            { name: objectGenerator.generateShortNameByPrefix(`auth-sa-${Date.now()}`) }
        );
        createdServiceAccounts.push(serviceAccount.body.id);

        expect(serviceAccount.body.clientId, 'service account should have a non-empty clientId').to.not.be.empty;
        expect(serviceAccount.body.clientSecret, 'service account should have a non-empty clientSecret').to.not.be.empty;
        const tokenResponse = await authUtils.authenticateAppAndGetToken(
            serviceAccount.body.clientId,
            serviceAccount.body.clientSecret
        );

        expect(tokenResponse.accessToken, 'authentication with valid credentials should return an accessToken').to.not.be.null;
        expect(tokenResponse.accessToken, 'accessToken should not be undefined after successful authentication').to.not.be.undefined;
    });

    test('should fail authentication with invalid clientSecret', async function () {
        serviceAccount = await serviceAccountUtils.createServiceAccount(
            flowRootState.users.system_administrator.accessToken,
            { name: objectGenerator.generateShortNameByPrefix(`invalid-sa-${Date.now()}`) }
        );
        createdServiceAccounts.push(serviceAccount.body.id);

        try {
            await authUtils.authenticateAppAndGetToken(serviceAccount.body.clientId, 'invalid-secret-123');
            expect.fail('Authentication with invalid secret should fail');
        } catch (e) {
            expect(e.message, 'error message should indicate token failure when using invalid clientSecret').to.include('Failed to obtain token');
        }
    });

    test('should invalidate old secret after regeneration', async function () {
        serviceAccount = await serviceAccountUtils.createServiceAccount(
            flowRootState.users.system_administrator.accessToken,
            { name: objectGenerator.generateShortNameByPrefix(`regen-sa-${Date.now()}`) }
        );
        createdServiceAccounts.push(serviceAccount.body.id);

        const originalSecret = serviceAccount.body.clientSecret;
        const clientId = serviceAccount.body.clientId;

        const firstAuth = await authUtils.authenticateAppAndGetToken(clientId, originalSecret);
        expect(firstAuth.accessToken, 'authentication with original secret should succeed before regeneration').to.not.be.null;

        const regenerateResponse = await serviceAccountUtils.regenerateServiceAccountSecret(
            flowRootState.users.system_administrator.accessToken,
            serviceAccount.body.id
        );

        const newSecret = regenerateResponse.body.clientSecret;
        expect(newSecret, 'regenerated secret should differ from the original secret').to.not.equal(originalSecret);

        try {
            await authUtils.authenticateAppAndGetToken(clientId, originalSecret);
            expect.fail('Old secret should not work after regeneration');
        } catch (e) {
            expect(e.message, 'error message should indicate token failure when using invalidated original secret').to.include('Failed to obtain token');
        }

        const newAuth = await authUtils.authenticateAppAndGetToken(clientId, newSecret);
        expect(newAuth.accessToken, 'authentication with new regenerated secret should succeed').to.not.be.null;
    });

    test('should generate unique secrets on each regeneration', async function () {
        serviceAccount = await serviceAccountUtils.createServiceAccount(
            flowRootState.users.system_administrator.accessToken,
            { name: objectGenerator.generateShortNameByPrefix(`unique-sa-${Date.now()}`) }
        );
        createdServiceAccounts.push(serviceAccount.body.id);

        const secrets = new Set<string>();
        secrets.add(serviceAccount.body.clientSecret);

        for (let i = 0; i < 3; i++) {
            const regenerated = await serviceAccountUtils.regenerateServiceAccountSecret(
                flowRootState.users.system_administrator.accessToken,
                serviceAccount.body.id
            );
            secrets.add(regenerated.body.clientSecret);
        }

        expect(secrets.size, 'each regeneration should produce a unique secret (4 unique secrets after initial + 3 regenerations)').to.equal(4);
    });

    test('should not expose clientSecret in GET operations', async function () {
        serviceAccount = await serviceAccountUtils.createServiceAccount(
            flowRootState.users.system_administrator.accessToken,
            { name: objectGenerator.generateShortNameByPrefix(`nosecret-sa-${Date.now()}`) }
        );
        createdServiceAccounts.push(serviceAccount.body.id);

        expect(serviceAccount.body.clientSecret, 'clientSecret should be present in the POST creation response').to.not.be.undefined;

        const retrieved = await serviceAccountUtils.getServiceAccountById(
            flowRootState.users.system_administrator.accessToken,
            serviceAccount.body.id
        );

        expect((retrieved.body as any).clientSecret, 'clientSecret should not be returned in GET response').to.be.undefined;
        expect((retrieved.body as any).secret, 'secret field should not be returned in GET response').to.be.undefined;
    });

    test('should update service account enabled state to false', async function () {
        serviceAccount = await serviceAccountUtils.createServiceAccount(
            flowRootState.users.system_administrator.accessToken,
            { name: objectGenerator.generateShortNameByPrefix(`disable-sa-${Date.now()}`) }
        );
        createdServiceAccounts.push(serviceAccount.body.id);

        let retrieved = await serviceAccountUtils.getServiceAccountById(
            flowRootState.users.system_administrator.accessToken,
            serviceAccount.body.id
        );
        expect(retrieved.body.enabled, 'service account should be enabled by default after creation').to.be.true;

        await serviceAccountUtils.updateServiceAccountById(
            flowRootState.users.system_administrator.accessToken,
            serviceAccount.body.id,
            { enabled: false }
        );

        retrieved = await serviceAccountUtils.getServiceAccountById(
            flowRootState.users.system_administrator.accessToken,
            serviceAccount.body.id
        );
        expect(retrieved.body.enabled, 'service account should be disabled after updating enabled to false').to.be.false;
    });

    test('should re-enable disabled service account', async function () {
        serviceAccount = await serviceAccountUtils.createServiceAccount(
            flowRootState.users.system_administrator.accessToken,
            { name: objectGenerator.generateShortNameByPrefix(`reenable-sa-${Date.now()}`) }
        );
        createdServiceAccounts.push(serviceAccount.body.id);

        await serviceAccountUtils.updateServiceAccountById(
            flowRootState.users.system_administrator.accessToken,
            serviceAccount.body.id,
            { enabled: false }
        );

        await serviceAccountUtils.updateServiceAccountById(
            flowRootState.users.system_administrator.accessToken,
            serviceAccount.body.id,
            { enabled: true }
        );

        const retrieved = await serviceAccountUtils.getServiceAccountById(
            flowRootState.users.system_administrator.accessToken,
            serviceAccount.body.id
        );
        expect(retrieved.body.enabled, 'service account should be enabled again after re-enabling').to.be.true;
    });

    test('should prevent authentication when service account is disabled', async function () {
        serviceAccount = await serviceAccountUtils.createServiceAccount(
            flowRootState.users.system_administrator.accessToken,
            { name: objectGenerator.generateShortNameByPrefix(`disabled-auth-sa-${Date.now()}`) }
        );
        createdServiceAccounts.push(serviceAccount.body.id);

        const clientId = serviceAccount.body.clientId;
        const clientSecret = serviceAccount.body.clientSecret;

        const authEnabled = await authUtils.authenticateAppAndGetToken(clientId, clientSecret);
        expect(authEnabled.accessToken, 'authentication should succeed before the service account is disabled').to.not.be.null;

        await serviceAccountUtils.updateServiceAccountById(
            flowRootState.users.system_administrator.accessToken,
            serviceAccount.body.id,
            { enabled: false }
        );

        try {
            await authUtils.authenticateAppAndGetToken(clientId, clientSecret);
            expect.fail('Disabled service account should not authenticate');
        } catch (e) {
            expect(e.message, 'error message should indicate token failure when authenticating with a disabled service account').to.include('Failed to obtain token');
        }
    });

    test('should handle duplicate service account names', async function () {
        const duplicateName = objectGenerator.generateShortNameByPrefix(`duplicate-sa-${Date.now()}`);

        const first = await serviceAccountUtils.createServiceAccount(
            flowRootState.users.system_administrator.accessToken,
            { name: duplicateName }
        );
        createdServiceAccounts.push(first.body.id);

        try {
            const second = await serviceAccountUtils.createServiceAccount(
                flowRootState.users.system_administrator.accessToken,
                { name: duplicateName }
            );
            createdServiceAccounts.push(second.body.id);

            expect(second.body.id, 'duplicate-named service accounts should receive different IDs').to.not.equal(first.body.id);
            logger.info('Duplicate names are allowed with different IDs');
        } catch (e) {
            expect(e.response.statusCode, 'duplicate service account name should return 409 Conflict or 400 Bad Request').to.be.oneOf([StatusCodes.CONFLICT, StatusCodes.BAD_REQUEST]);
            logger.info('Duplicate names are not allowed');
        }
    });

    test('should track createdBy field correctly', async function () {
        serviceAccount = await serviceAccountUtils.createServiceAccount(
            flowRootState.users.system_administrator.accessToken,
            { name: objectGenerator.generateShortNameByPrefix(`createdby-sa-${Date.now()}`) }
        );
        createdServiceAccounts.push(serviceAccount.body.id);

        const retrieved = await serviceAccountUtils.getServiceAccountById(
            flowRootState.users.system_administrator.accessToken,
            serviceAccount.body.id
        );

        expect(retrieved.body.createdBy, 'createdBy field should be set after service account creation').to.not.be.undefined;
        expect(retrieved.body.createdBy, 'createdBy field should not be null').to.not.be.null;
        logger.info(`Service account created by: ${retrieved.body.createdBy}`);
    });

    test('should initialize lastLogin as null for new service account', async function () {
        serviceAccount = await serviceAccountUtils.createServiceAccount(
            flowRootState.users.system_administrator.accessToken,
            { name: objectGenerator.generateShortNameByPrefix(`lastlogin-sa-${Date.now()}`) }
        );
        createdServiceAccounts.push(serviceAccount.body.id);

        const retrieved = await serviceAccountUtils.getServiceAccountById(
            flowRootState.users.system_administrator.accessToken,
            serviceAccount.body.id
        );

        expect(retrieved.body.lastLogin, 'lastLogin should be null for a newly created service account that has never authenticated').to.be.null;
    });

    test('should have valid timestamps for createdAt and updatedAt', async function () {
        serviceAccount = await serviceAccountUtils.createServiceAccount(
            flowRootState.users.system_administrator.accessToken,
            { name: objectGenerator.generateShortNameByPrefix(`timestamps-sa-${Date.now()}`) }
        );
        createdServiceAccounts.push(serviceAccount.body.id);

        const retrieved = await serviceAccountUtils.getServiceAccountById(
            flowRootState.users.system_administrator.accessToken,
            serviceAccount.body.id
        );

        expect(retrieved.body.createdAt, 'createdAt should be set after service account creation').to.not.be.null;
        expect(retrieved.body.updatedAt, 'updatedAt should be set after service account creation').to.not.be.null;

        const createdAt = new Date(retrieved.body.createdAt);
        const updatedAt = new Date(retrieved.body.updatedAt);

        expect(createdAt.getTime(), 'createdAt should not be a future timestamp').to.be.lessThanOrEqual(Date.now());
        expect(updatedAt.getTime(), 'updatedAt should not be a future timestamp').to.be.lessThanOrEqual(Date.now());
    });

    test('should update updatedAt timestamp when service account is modified', async function () {
        serviceAccount = await serviceAccountUtils.createServiceAccount(
            flowRootState.users.system_administrator.accessToken,
            { name: objectGenerator.generateShortNameByPrefix(`update-ts-sa-${Date.now()}`) }
        );
        createdServiceAccounts.push(serviceAccount.body.id);

        const original = await serviceAccountUtils.getServiceAccountById(
            flowRootState.users.system_administrator.accessToken,
            serviceAccount.body.id
        );
        const originalUpdatedAt = new Date(original.body.updatedAt);

        await new Promise((resolve) => setTimeout(resolve, 1000));

        await serviceAccountUtils.updateServiceAccountById(
            flowRootState.users.system_administrator.accessToken,
            serviceAccount.body.id,
            { enabled: false }
        );

        const updated = await serviceAccountUtils.getServiceAccountById(
            flowRootState.users.system_administrator.accessToken,
            serviceAccount.body.id
        );
        const newUpdatedAt = new Date(updated.body.updatedAt);

        expect(newUpdatedAt.getTime(), 'updatedAt should advance after modifying the service account').to.be.greaterThan(originalUpdatedAt.getTime());
    });

    test('should list multiple service accounts', async function () {
        const prefix = `list-test-${Date.now()}`;

        for (let i = 0; i < 5; i++) {
            const sa = await serviceAccountUtils.createServiceAccount(
                flowRootState.users.system_administrator.accessToken,
                { name: objectGenerator.generateShortNameByPrefix(`${prefix}-${i}`) }
            );
            createdServiceAccounts.push(sa.body.id);
        }

        const allServiceAccounts = await serviceAccountUtils.getServiceAccounts(
            flowRootState.users.system_administrator.accessToken
        );

        expect(allServiceAccounts.body, 'list service accounts response body should be an array').to.be.an('array');
        expect(allServiceAccounts.body.length, 'list should contain at least the 5 service accounts just created').to.be.at.least(5);

        const ourIds = new Set(createdServiceAccounts);
        const foundIds = allServiceAccounts.body.filter((sa) => ourIds.has(sa.id));
        expect(foundIds.length, 'all 5 created service accounts should appear in the list response').to.equal(5);
    });

    test('should handle service account name with maximum allowed length', async function () {
        const maxLengthName = 'a'.repeat(63);

        try {
            serviceAccount = await serviceAccountUtils.createServiceAccount(
                flowRootState.users.system_administrator.accessToken,
                { name: maxLengthName }
            );
            createdServiceAccounts.push(serviceAccount.body.id);

            expect(serviceAccount.body.name, 'service account name should match the submitted 63-character name').to.equal(maxLengthName);
            logger.info('Maximum length name (63 chars) is accepted');
        } catch (e) {
            logger.info(`Name length 63 is too long: ${e.message}`);
            expect(e.response.statusCode, 'server should reject a name exceeding max length with 400 Bad Request').to.equal(StatusCodes.BAD_REQUEST);
        }
    });

    test('should handle service account name with valid special characters', async function () {
        const nameWithHyphens = objectGenerator.generateShortNameByPrefix(
            `test-sa-with-hyphens-${Date.now()}`
        );

        serviceAccount = await serviceAccountUtils.createServiceAccount(
            flowRootState.users.system_administrator.accessToken,
            { name: nameWithHyphens }
        );
        createdServiceAccounts.push(serviceAccount.body.id);

        expect(serviceAccount.body.name, 'service account name with hyphens should be stored exactly as submitted').to.equal(nameWithHyphens);
    });

    test('should invalidate tokens after service account deletion', async function () {
        serviceAccount = await serviceAccountUtils.createServiceAccount(
            flowRootState.users.system_administrator.accessToken,
            { name: objectGenerator.generateShortNameByPrefix(`delete-token-sa-${Date.now()}`) }
        );

        const clientId = serviceAccount.body.clientId;
        const clientSecret = serviceAccount.body.clientSecret;

        const authResponse = await authUtils.authenticateAppAndGetToken(clientId, clientSecret);
        const token = authResponse.accessToken;
        expect(token, 'authentication should succeed before the service account is deleted').to.not.be.null;

        await serviceAccountUtils.deleteServiceAccount(
            flowRootState.users.system_administrator.accessToken,
            serviceAccount.body.id
        );

        try {
            await authUtils.authenticateAppAndGetToken(clientId, clientSecret);
            expect.fail('Deleted service account should not authenticate');
        } catch (e) {
            expect(e.message, 'error message should indicate token failure after service account deletion').to.include('Failed to obtain token');
        }
    });

    test('should handle concurrent service account creations', async function () {
        const prefix = `concurrent-${Date.now()}`;
        const createPromises = [];

        for (let i = 0; i < 5; i++) {
            createPromises.push(
                serviceAccountUtils.createServiceAccount(
                    flowRootState.users.system_administrator.accessToken,
                    { name: objectGenerator.generateShortNameByPrefix(`${prefix}-${i}`) }
                )
            );
        }

        const results = await Promise.all(createPromises);
        expect(results.length, 'all 5 concurrent creation requests should resolve').to.equal(5);
        results.forEach((result) => {
            expect(result.body.id, 'each concurrently created service account should have an id').to.not.be.undefined;
            createdServiceAccounts.push(result.body.id);
        });

        const ids = new Set(results.map((r) => r.body.id));
        expect(ids.size, 'each concurrent creation should produce a unique service account id').to.equal(5);
    });
});
