/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { UnleashClient } from "@unleash/proxy-client-vue";
import { getBootConfig } from "@/services/infra/boot-config.service/boot-config.service";

const ONE_HOUR_IN_SECONDS = 3600;

class UnleashConfigService {
  private client: UnleashClient | undefined;

  async init(tenantId: string): Promise<void> {
    const unleashClient = this.getClient();
    if (unleashClient) {
      await this.getClient()?.updateContext({ properties: { tenantId: tenantId } });
    }
  }
  getClient(): UnleashClient | undefined {
    if (!this.client) {
      const deploymentConfig = getBootConfig().unleashConfig;
      const config = deploymentConfig ? { ...deploymentConfig, refreshInterval: ONE_HOUR_IN_SECONDS } : undefined;
      if (!config) {
        return;
      }
      try {
        this.client = new UnleashClient(config);
      } catch {
        return;
      }
    }
    return this.client;
  }

  isFeatureEnabled(featureName: string): boolean {
    return this.getClient()?.isEnabled(featureName) ?? true;
  }
}

export const unleashConfigService = new UnleashConfigService();
