/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { InferencesApi, RegistriesApi, ImagesApi } from './generated/api';
import { API_DEFAULT_HEADERS } from '../apiDefaults';

export * from './generated/model/models';

export class CatalogApi {
    inferencesApi: InferencesApi;

    registriesApi: RegistriesApi;

    imagesApi: ImagesApi;

    constructor(url: string) {
        this.inferencesApi = new InferencesApi(url);
        this.inferencesApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.registriesApi = new RegistriesApi(url);
        this.registriesApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.imagesApi = new ImagesApi(url);
        this.imagesApi.defaultHeaders = API_DEFAULT_HEADERS;
    }
}
