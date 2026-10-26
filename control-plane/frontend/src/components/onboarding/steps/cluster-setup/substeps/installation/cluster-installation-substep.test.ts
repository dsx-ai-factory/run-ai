/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { afterEach, describe, expect, it, vi } from "vitest";
import { shallowMount } from "@vue/test-utils";
import { DOCS } from "@/constants/docs";
import ClusterInstallationSubstep from "./cluster-installation-substep.vue";

vi.mock("vue3-lottie", () => ({ Vue3Lottie: { template: "<div />" } }));

const installationDocs = DOCS.INSTALLATION as {
  PREREQUISITES: string | undefined;
  NETWORK_REQUIREMENTS: string | undefined;
};
const originalRequirements = { ...installationDocs };

describe("onboarding installation documentation", () => {
  afterEach(() => {
    Object.assign(installationDocs, originalRequirements);
  });

  it.each([
    [false, false],
    [true, false],
    [false, true],
    [true, true],
  ])("renders only available requirements links (system: %s, network: %s)", (hasSystem, hasNetwork) => {
    installationDocs.PREREQUISITES = hasSystem ? "https://docs.example.test/system" : undefined;
    installationDocs.NETWORK_REQUIREMENTS = hasNetwork ? "https://docs.example.test/network" : undefined;

    const wrapper = shallowMount(ClusterInstallationSubstep, {
      props: {
        clusterData: {
          clusterName: "example-cluster",
          clusterVersion: "3.0.0",
          installationOption: "",
          clusterLocation: "",
          clusterUrl: "",
          hasClusterValues: false,
          isClusterConnected: false,
        },
      },
      global: {
        stubs: { OnboardingContentWrapper: { template: '<div><slot name="subtitle" /></div>' } },
      },
    });

    expect(wrapper.text().replace(/\s+/g, " ")).toContain("Review the system and network requirements.");
    expect(wrapper.findAll("a:not([href])")).toHaveLength(0);
    for (const [label, expectedHref] of [
      ["system", installationDocs.PREREQUISITES],
      ["network", installationDocs.NETWORK_REQUIREMENTS],
    ]) {
      const link = wrapper.findAll("a").find((anchor) => anchor.text() === label);
      if (expectedHref) {
        expect(link?.attributes("href")).toBe(expectedHref);
        expect(link?.attributes("target")).toBe("_blank");
      } else {
        expect(link).toBeUndefined();
        expect(wrapper.findAll("span").some((span) => span.text() === label)).toBe(true);
      }
    }
    expect(
      wrapper
        .findAll("a")
        .find((anchor) => anchor.text() === "install")
        ?.attributes("href"),
    ).toBe(DOCS.INSTALLATION.HELM_INSTALL);
    wrapper.unmount();
  });
});
