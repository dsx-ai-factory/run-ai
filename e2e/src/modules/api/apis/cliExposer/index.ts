/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { ResearcherCommandLineInterfaceDeprecatedApi, HealthApi } from './generated/api';
import { API_DEFAULT_HEADERS } from '../apiDefaults';
import { flowRootState } from '../../../helpers';

export * from './generated/model/models';

// Authentication previously added inside generated code must survive regeneration.
class CLIApi extends ResearcherCommandLineInterfaceDeprecatedApi {
    getInstallerUnixCommands(
        options: Parameters<ResearcherCommandLineInterfaceDeprecatedApi['getInstallerUnixCommandsDeprecated']>[0] = { headers: {} },
        accessToken = flowRootState.users.system_administrator.accessToken
    ): ReturnType<ResearcherCommandLineInterfaceDeprecatedApi['getInstallerUnixCommandsDeprecated']> {
        this.accessToken = accessToken;
        return super.getInstallerUnixCommandsDeprecated(options);
    }

    getManualDocument(
        documentName: string,
        options: Parameters<ResearcherCommandLineInterfaceDeprecatedApi['getManualDocumentDeprecated']>[1] = { headers: {} },
        accessToken = flowRootState.users.system_administrator.accessToken
    ): ReturnType<ResearcherCommandLineInterfaceDeprecatedApi['getManualDocumentDeprecated']> {
        this.accessToken = accessToken;
        return super.getManualDocumentDeprecated(documentName, options);
    }
}

export class CliExposer {
    cliApi: CLIApi;

    healthApi: HealthApi;

    constructor(url: string) {
        this.cliApi = new CLIApi(url);
        this.cliApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.healthApi = new HealthApi(url);
        this.healthApi.defaultHeaders = API_DEFAULT_HEADERS;
    }
}
