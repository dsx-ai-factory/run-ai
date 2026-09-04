/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
  ResearcherCommandLineInterfaceApi,
  Configuration,
  type ConfigurationParameters,
} from "@/swagger-models/cli-exposer-client";
import { authenticatedAxios } from "@/services/infra/https.service/auth-axios.service";

class CliApi {
  configuration: Configuration;

  cliApi: ResearcherCommandLineInterfaceApi;

  constructor() {
    this.configuration = new Configuration();

    this.cliApi = new ResearcherCommandLineInterfaceApi(this.configuration, this.baseUrl, authenticatedAxios);
  }

  get baseUrl() {
    return `${window.location.origin}`;
  }

  public set setAccessToken(accessToken: ConfigurationParameters["accessToken"]) {
    this.configuration.accessToken = accessToken;
  }
}

export const cliApi = new CliApi();
