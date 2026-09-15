/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { beforeEach, describe, expect, it, vi } from "vitest";
import type { BootConfig } from "@/swagger-models/bff-service-client";

const apiState = vi.hoisted(() => ({
  getBootConfig: (): Promise<unknown> => Promise.reject(new Error("not configured")),
}));

vi.mock("@/swagger-models/bff-service-client", () => ({
  Configuration: class Configuration {},
  BootConfigApi: class BootConfigApi {
    getBootConfig(): Promise<unknown> {
      return apiState.getBootConfig();
    }
  },
}));

import { FALLBACK_BOOT_CONFIG, getBootConfig, loadBootConfig, showsClusterLocationStep } from "./boot-config.service";

function createDeploymentConfig(): BootConfig {
  return {
    distributionAssets: { profile: "cc" },
    version: { display: "release-42" },
    docs: {
      catalogVersion: "v1",
      baseUrl: "https://docs.example.test/product",
      whatsNew: {
        label: "Product updates",
        url: "https://docs.example.test/product/updates",
      },
    },
    support: {
      url: "https://support.example.test/contact",
      label: "Get help",
    },
    capabilities: {
      showsClusterLocationStep: true,
    },
    unleashConfig: null,
  };
}

describe("bootConfigService", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("uses deployment-provided frontend values without inferring a profile", async () => {
    const deploymentConfig = {
      ...createDeploymentConfig(),
      capabilities: { showsClusterLocationStep: false },
    };
    apiState.getBootConfig = () => Promise.resolve({ data: deploymentConfig });

    await expect(loadBootConfig()).resolves.toEqual(deploymentConfig);
    expect(getBootConfig()).toEqual(deploymentConfig);
    expect(getBootConfig().docs).toEqual(deploymentConfig.docs);
    expect(showsClusterLocationStep()).toBe(false);
  });

  it("accepts a known distribution asset profile", async () => {
    const deploymentConfig = {
      ...createDeploymentConfig(),
      distributionAssets: { profile: "ce" as const },
    };
    apiState.getBootConfig = () => Promise.resolve({ data: deploymentConfig });

    await expect(loadBootConfig()).resolves.toEqual(deploymentConfig);
  });

  it.each(["cc", "ce"] as const)("accepts v2 independently of asset profile %s", async (profile) => {
    const deploymentConfig = createDeploymentConfig();
    deploymentConfig.distributionAssets.profile = profile;
    deploymentConfig.docs.catalogVersion = "v2";
    apiState.getBootConfig = () => Promise.resolve({ data: deploymentConfig });

    await expect(loadBootConfig()).resolves.toEqual(deploymentConfig);
  });

  it("preserves deployment URLs when an older response omits the catalog", async () => {
    const deploymentConfig = createDeploymentConfig();
    delete deploymentConfig.docs.catalogVersion;
    apiState.getBootConfig = () => Promise.resolve({ data: deploymentConfig });

    await expect(loadBootConfig()).resolves.toEqual(deploymentConfig);
  });

  it.each(["v3", 2, null])("rejects unsupported catalog %s", async (catalogVersion) => {
    const deploymentConfig = createDeploymentConfig();
    apiState.getBootConfig = () =>
      Promise.resolve({
        data: { ...deploymentConfig, docs: { ...deploymentConfig.docs, catalogVersion } },
      });

    await expect(loadBootConfig()).resolves.toEqual(FALLBACK_BOOT_CONFIG);
  });

  it("uses fallback defaults for an unknown distribution asset profile", async () => {
    const deploymentConfig = {
      ...createDeploymentConfig(),
      distributionAssets: { profile: "custom" },
    };
    apiState.getBootConfig = () => Promise.resolve({ data: deploymentConfig });

    await expect(loadBootConfig()).resolves.toEqual(FALLBACK_BOOT_CONFIG);
  });

  it("uses fallback defaults when the request fails", async () => {
    apiState.getBootConfig = () => Promise.reject(new Error("network unavailable"));

    await expect(loadBootConfig()).resolves.toEqual(FALLBACK_BOOT_CONFIG);
    expect(getBootConfig()).toEqual(FALLBACK_BOOT_CONFIG);
  });

  it("uses the fallback when the endpoint returns an incomplete payload", async () => {
    apiState.getBootConfig = () => Promise.resolve({ data: { version: { display: "2.26.0" } } });

    await expect(loadBootConfig()).resolves.toEqual(FALLBACK_BOOT_CONFIG);
  });

  it("uses the fallback when an optional Unleash field has the wrong type", async () => {
    const invalidConfig = {
      ...createDeploymentConfig(),
      unleashConfig: {
        url: "https://unleash.example.com",
        clientKey: "public-client-key",
        appName: "browser-app",
        environment: "validation",
        project: 42,
      },
    };
    apiState.getBootConfig = () => Promise.resolve({ data: invalidConfig });

    await expect(loadBootConfig()).resolves.toEqual(FALLBACK_BOOT_CONFIG);
  });

  it("accepts a deployment config without an optional support URL", async () => {
    const deploymentConfig = {
      ...createDeploymentConfig(),
      support: { label: "Get help" },
    };
    apiState.getBootConfig = () => Promise.resolve({ data: deploymentConfig });

    await expect(loadBootConfig()).resolves.toEqual(deploymentConfig);
  });

  it("accepts root-relative deployment URLs", async () => {
    const deploymentConfig = {
      ...createDeploymentConfig(),
      docs: {
        baseUrl: "/product-docs",
        whatsNew: { label: "Product updates", url: "/product-docs/updates" },
      },
      support: { url: "/help", label: "Get help" },
    };
    apiState.getBootConfig = () => Promise.resolve({ data: deploymentConfig });

    await expect(loadBootConfig()).resolves.toEqual(deploymentConfig);
  });

  it.each(["javascript:alert(1)", "//evil.example", "/\\evil.example"])("rejects unsafe browser URL %s", async (url) => {
    const deploymentConfig = {
      ...createDeploymentConfig(),
      support: { ...createDeploymentConfig().support, url },
    };
    apiState.getBootConfig = () => Promise.resolve({ data: deploymentConfig });

    await expect(loadBootConfig()).resolves.toEqual(FALLBACK_BOOT_CONFIG);
  });
});
