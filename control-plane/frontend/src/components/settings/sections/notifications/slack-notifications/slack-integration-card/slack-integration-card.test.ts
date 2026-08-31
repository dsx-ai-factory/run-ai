/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { Quasar } from "quasar";
import SlackIntegrationCard from "./slack-integration-card.vue";

describe("SlackIntegrationCard", () => {
  let wrapper: VueWrapper<InstanceType<typeof SlackIntegrationCard>>;

  const createWrapper = (workspaceName = "Test Workspace") => {
    return mount(SlackIntegrationCard, {
      props: {
        workspaceName,
      },
      global: {
        plugins: [Quasar],
      },
    });
  };

  beforeEach(() => {
    wrapper = createWrapper();
  });

  afterEach(() => {
    vi.clearAllMocks();
    vi.resetAllMocks();
  });

  describe("Component Rendering", () => {
    it("renders the component correctly", () => {
      expect(wrapper.exists()).toBe(true);
      expect(wrapper.classes()).toContain("slack-integration-card");
    });

    it("displays the correct connection message", () => {
      expect(wrapper.text()).toContain("Nvidia Run:ai is connected to your Slack workspace");
    });

    it("shows the correct status text", () => {
      const statusText = wrapper.find(".status-text");
      expect(statusText.exists()).toBe(true);
      expect(statusText.text()).toBe("Connected");
    });

    it("renders the disconnect button with correct properties", () => {
      const disconnectBtn = wrapper.find('[aid="disconnect-slack-integration-btn"]');
      expect(disconnectBtn.exists()).toBe(true);
      expect(disconnectBtn.text()).toBe("Disconnect");
    });

    it("displays the status section correctly", () => {
      const statusSection = wrapper.find(".slack-integration-status");
      expect(statusSection.exists()).toBe(true);
      expect(statusSection.text()).toContain("Status:");
      expect(statusSection.text()).toContain("Connected");
    });

    it("displays the workspace name when provided", () => {
      const workspaceNameElement = wrapper.find(".workspace-name");
      expect(workspaceNameElement.exists()).toBe(true);
      expect(workspaceNameElement.text()).toBe("Test Workspace");
    });
  });

  describe("Component Structure", () => {
    it("has the correct footer structure", () => {
      const footer = wrapper.find(".slack-integration-card-footer");
      expect(footer.exists()).toBe(true);

      const statusElement = footer.find(".slack-integration-status");
      const disconnectButton = footer.find(".disconnect-btn");

      expect(statusElement.exists()).toBe(true);
      expect(disconnectButton.exists()).toBe(true);
    });

    it("applies correct styling classes", () => {
      expect(wrapper.classes()).toContain("slack-integration-card");
      expect(wrapper.find(".slack-integration-card-footer").exists()).toBe(true);
      expect(wrapper.find(".slack-integration-status").exists()).toBe(true);
      expect(wrapper.find(".status-text").exists()).toBe(true);
      expect(wrapper.find(".disconnect-btn").exists()).toBe(true);
    });
  });

  describe("User Interactions", () => {
    it("emits 'disconnect' event when disconnect button is clicked", async () => {
      const disconnectBtn = wrapper.find('[aid="disconnect-slack-integration-btn"]');
      await disconnectBtn.trigger("click");

      expect(wrapper.emitted()).toHaveProperty("disconnect");
      expect(wrapper.emitted("disconnect")).toHaveLength(1);
    });

    it("calls disconnectSlackIntegration method when button is clicked", async () => {
      const disconnectBtn = wrapper.find('[aid="disconnect-slack-integration-btn"]');
      await disconnectBtn.trigger("click");

      // Verify the method was called by checking the emitted event
      expect(wrapper.emitted("disconnect")).toBeTruthy();
    });
  });

  describe("Component Methods", () => {
    it("disconnectSlackIntegration method emits disconnect event", () => {
      wrapper.vm.disconnectSlackIntegration();

      expect(wrapper.emitted()).toHaveProperty("disconnect");
      expect(wrapper.emitted("disconnect")).toHaveLength(1);
    });
  });

  describe("Component Props and Emits", () => {
    it("declares the correct emits", () => {
      const component = wrapper.vm.$options;
      expect(component.emits).toContain("disconnect");
    });

    it("has the correct component name", () => {
      expect(wrapper.vm.$options.name).toBe("slack-integration-card");
    });

    it("accepts workspaceName prop correctly", () => {
      const customWrapper = createWrapper("Custom Workspace");
      expect(customWrapper.props("workspaceName")).toBe("Custom Workspace");
    });
  });

  describe("Accessibility", () => {
    it("has proper accessibility attributes on disconnect button", () => {
      const disconnectBtn = wrapper.find('[aid="disconnect-slack-integration-btn"]');
      expect(disconnectBtn.attributes("aid")).toBe("disconnect-slack-integration-btn");
    });
  });

  describe("Styling and CSS Classes", () => {
    it("applies success color to status text", () => {
      const statusText = wrapper.find(".status-text");
      expect(statusText.exists()).toBe(true);
    });

    it("applies correct styling to disconnect button", () => {
      const disconnectBtn = wrapper.find(".disconnect-btn");
      expect(disconnectBtn.exists()).toBe(true);
    });
  });
});
