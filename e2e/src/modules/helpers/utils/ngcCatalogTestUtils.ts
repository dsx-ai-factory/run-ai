/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import { workloadsCrudService } from '../../api';
import { CatalogResourcesInner } from '../../api/apis/workloadsApi/generated/model/catalogResourcesInner';
import { Catalog } from '../../api/apis/workloadsApi/generated/model/catalog';

export interface NgcCatalogTestOptions {
    filter?: string[];
    query?: string;
    orderBy?: string;
    page?: number;
    pageSize?: number;
}

export interface NgcCatalogResponse {
    response: any;
    body: Catalog;
}

export class NgcCatalogTestUtils {
    async callNgcCatalog(options: NgcCatalogTestOptions = {}): Promise<NgcCatalogResponse> {
        const {
            filter = undefined,
            query = undefined,
            orderBy = undefined,
            page = undefined,
            pageSize = undefined
        } = options;

        return await workloadsCrudService.inferencesApi.getNimCatalog(filter, query, orderBy, page, pageSize);
    }

    validateBasicResponse(response: NgcCatalogResponse, expectEmpty = false): Catalog {
        expect(response.response.statusCode).to.equal(200);
        expect(response.body).to.not.be.undefined;

        const catalog = response.body;

        if (!expectEmpty) {
            expect(catalog.resultTotal).to.be.a('number');
            expect(catalog.resultPageTotal).to.be.a('number');
            expect(catalog.resources).to.be.an('array');
            expect(catalog.resources.length).to.be.greaterThan(0);
        }

        return catalog;
    }

    validateResourceStructure(resource: CatalogResourcesInner): void {
        expect(resource).to.have.property('orgName');
        expect(resource).to.have.property('resourceId');
        expect(resource).to.have.property('displayName');
        expect(resource).to.have.property('name');
        expect(resource).to.have.property('resourceType');
        expect(resource).to.have.property('description');
        expect(resource).to.have.property('isPublic');

        expect(resource.orgName).to.be.a('string');
        expect(resource.resourceId).to.be.a('string');
        expect(resource.displayName).to.be.a('string');
        expect(resource.name).to.be.a('string');
        expect(resource.resourceType).to.equal('CONTAINER');
        expect(resource.description).to.be.a('string');
        if (resource.isPublic !== undefined) {
            expect(resource.isPublic).to.be.a('boolean');
        }

        // Verify resourceId format (should be org/name)
        expect(resource.resourceId).to.include('/');
        const parts = resource.resourceId.split('/');
        expect(parts.length).to.equal(2);
    }

    validateCatalogStructure(catalog: Catalog): CatalogResourcesInner {
        expect(catalog).to.have.property('resultTotal');
        expect(catalog).to.have.property('resultPageTotal');
        expect(catalog).to.have.property('resources');
        expect(catalog).to.have.property('isStaticData');

        expect(catalog.resultTotal).to.be.a('number');
        expect(catalog.resultPageTotal).to.be.a('number');
        expect(catalog.resources).to.be.an('array');
        expect(catalog.isStaticData).to.be.a('boolean');

        expect(catalog.resources.length).to.be.greaterThan(0);

        // Find the first resource
        expect(catalog.resources.length).to.be.greaterThan(0);
        const firstResource = catalog.resources[0];

        expect(firstResource).to.have.property('resourceId');
        expect(firstResource).to.have.property('orgName');

        expect(firstResource.resourceId).to.be.a('string');
        expect(firstResource.orgName).to.be.a('string');

        return firstResource;
    }

    logDataSource(catalog: Catalog): void {
        if (catalog.isStaticData) {
            console.log('Using fallback catalog data (NGC API may be unavailable)');
        } else {
            console.log('Using real NGC API data');
        }
    }

    validatePagination(firstPage: Catalog, secondPage: Catalog): void {
        expect(firstPage.resources).to.be.an('array');
        expect(secondPage.resources).to.be.an('array');
        expect(firstPage.resources.length).to.be.greaterThan(0);
        expect(secondPage.resources.length).to.be.greaterThan(0);

        if (!firstPage.isStaticData && !secondPage.isStaticData) {
            console.log('✅ Testing real NGC API pagination');

            const firstPageResources = firstPage.resources;
            const secondPageResources = secondPage.resources;

            if (firstPage.resultTotal > 5) {
                expect(firstPageResources.length).to.equal(5);
                expect(secondPageResources.length).to.equal(5);

                // Resources should be different between pages
                const firstPageIds = firstPageResources.map((r: CatalogResourcesInner) => r.resourceId);
                const secondPageIds = secondPageResources.map((r: CatalogResourcesInner) => r.resourceId);

                const overlap = firstPageIds.filter((id: string) => secondPageIds.includes(id));
                expect(overlap.length).to.equal(0);
            }
        } else {
            console.log('⚠️  Using fallback data - pagination test limited');
        }
    }

    validateSearchResults(catalog: Catalog, searchTerm: string): void {
        if (!catalog.isStaticData) {
            console.log('✅ Testing real NGC API search');

            expect(catalog.resources).to.be.an('array');

            const matchingResources = catalog.resources.filter(
                (resource: CatalogResourcesInner) =>
                    resource.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                    resource.description.toLowerCase().includes(searchTerm.toLowerCase())
            );

            console.log(`Found ${matchingResources.length} resources containing '${searchTerm}'`);
        } else {
            console.log('⚠️  Using fallback data - search test limited');
        }
    }
}
