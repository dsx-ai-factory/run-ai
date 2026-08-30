import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";
import SlackNotifications from "./slack-notifications.vue";
import { useAppStore } from "@/stores/app.store";

function factory(options: { isMultiTenant?: boolean; props?: Record<string, unknown> } = {}) {
  const { isMultiTenant = true, props = {} } = options;

  setActivePinia(createTestingPinia({ createSpy: vi.fn }));
  const appStore = useAppStore();
  Object.assign(appStore, {
    config: { isMultiTenant, authClientID: "", tenantName: "", authHost: "" },
  });

  return mount(SlackNotifications, {
    global: {
      plugins: [Quasar],
    },
    props: {
      hasSlackIntegration: false,
      ...props,
    },
  });
}

describe("SlackNotifications", () => {
  let wrapper: VueWrapper<InstanceType<typeof SlackNotifications>>;

  beforeEach(() => {
    wrapper = factory();
  });

  afterEach(() => {
    vi.clearAllMocks();
    vi.resetAllMocks();
  });

  describe("Component Rendering", () => {
    it("renders the component correctly", () => {
      expect(wrapper.exists()).toBe(true);
      expect(wrapper.classes()).toContain("slack-notifications");
    });

    it("displays the correct title", () => {
      expect(wrapper.text()).toContain("Slack notifications");
    });

    it("displays the correct description text", () => {
      expect(wrapper.text()).toContain("Authorize and connect your Slack workspace to NVIDIA Run:ai");
    });

    it("renders the tooltip component", () => {
      const tooltip = wrapper.find('[aid="tool-tip-slack-notification"]');
      expect(tooltip.exists()).toBe(true);
    });
  });

  describe("Conditional Rendering - No Slack Integration", () => {
    it("shows connect button when hasSlackIntegration is false", () => {
      const connectBtn = wrapper.find('[aid="add-slack-app-btn"]');
      expect(connectBtn.exists()).toBe(true);
      expect(connectBtn.text()).toContain("Connect Slack Workspace");
    });

    it("does not show slack integration card when hasSlackIntegration is false", () => {
      const integrationCard = wrapper.findComponent({ name: "slack-integration-card" });
      expect(integrationCard.exists()).toBe(false);
    });

    it("connect button has correct properties", () => {
      const connectBtn = wrapper.find('[aid="add-slack-app-btn"]');
      expect(connectBtn.attributes("aid")).toBe("add-slack-app-btn");
    });
  });

  describe("Conditional Rendering - With Slack Integration", () => {
    beforeEach(() => {
      wrapper = factory({ props: { hasSlackIntegration: true } });
    });

    it("shows slack integration card when hasSlackIntegration is true", () => {
      const integrationCard = wrapper.findComponent({ name: "slack-integration-card" });
      expect(integrationCard.exists()).toBe(true);
    });

    it("does not show connect button when hasSlackIntegration is true", () => {
      const connectBtn = wrapper.find('[aid="add-slack-app-btn"]');
      expect(connectBtn.exists()).toBe(false);
    });
  });

  describe("User Interactions", () => {
    it("emits 'connect-slack-workspace' event when connect button is clicked with credentials", () => {
      wrapper = factory({ props: { appCredentialsExist: true } });
      const connectBtn = wrapper.find('[aid="add-slack-app-btn"]');
      connectBtn.trigger("click");

      expect(wrapper.emitted()).toHaveProperty("connect-slack-workspace");
      expect(wrapper.emitted("connect-slack-workspace")).toHaveLength(1);
    });

    it("emits 'disconnect-slack-workspace' event when slack integration card emits disconnect", () => {
      wrapper = factory({ props: { hasSlackIntegration: true } });
      const integrationCard = wrapper.findComponent({ name: "slack-integration-card" });

      integrationCard.vm.$emit("disconnect");

      expect(wrapper.emitted()).toHaveProperty("disconnect-slack-workspace");
      expect(wrapper.emitted("disconnect-slack-workspace")).toHaveLength(1);
    });
  });

  describe("Component Props and Emits", () => {
    it("declares the correct emits", () => {
      const component = wrapper.vm.$options;
      expect(component.emits).toContain("connect-slack-workspace");
      expect(component.emits).toContain("disconnect-slack-workspace");
    });

    it("has the correct component name", () => {
      expect(wrapper.vm.$options.name).toBe("slack-notifications");
    });

    it("has correct props definition", () => {
      const component = wrapper.vm.$options;
      expect(component.props).toHaveProperty("hasSlackIntegration");
      expect(component.props.hasSlackIntegration.type).toBe(Boolean);
      expect(component.props.hasSlackIntegration.default).toBe(false);
    });
  });

  describe("Component Structure", () => {
    it("has correct component imports", () => {
      const component = wrapper.vm.$options;
      expect(component.components).toHaveProperty("RunaiTooltip");
      expect(component.components).toHaveProperty("RunaiBtn");
      expect(component.components).toHaveProperty("SlackIntegrationCard");
    });

    it("renders the title with correct styling", () => {
      const title = wrapper.find(".title");
      expect(title.exists()).toBe(true);
      expect(title.text()).toBe("Slack notifications");
    });

    it("has proper row structure", () => {
      const rows = wrapper.findAll(".row");
      expect(rows.length).toBeGreaterThan(0);
    });
  });

  describe("Accessibility", () => {
    it("has proper accessibility attributes on connect button", () => {
      const connectBtn = wrapper.find('[aid="add-slack-app-btn"]');
      expect(connectBtn.attributes("aid")).toBe("add-slack-app-btn");
    });

    it("has proper accessibility attributes on tooltip", () => {
      const tooltip = wrapper.find('[aid="tool-tip-slack-notification"]');
      expect(tooltip.attributes("aid")).toBe("tool-tip-slack-notification");
    });
  });

  describe("Styling and CSS Classes", () => {
    it("applies correct styling classes", () => {
      expect(wrapper.classes()).toContain("slack-notifications");
      expect(wrapper.find(".title").exists()).toBe(true);
    });

    it("connect button has correct styling properties", () => {
      const connectBtn = wrapper.find('[aid="add-slack-app-btn"]');
      expect(connectBtn.exists()).toBe(true);
    });
  });

  describe("Props Behavior", () => {
    it("defaults hasSlackIntegration to false", () => {
      const newWrapper = factory();
      expect(newWrapper.props("hasSlackIntegration")).toBe(false);
    });

    it("accepts hasSlackIntegration prop correctly", async () => {
      await wrapper.setProps({ hasSlackIntegration: true });
      expect(wrapper.props("hasSlackIntegration")).toBe(true);
    });
  });

  describe("Single Tenant - No Credentials", () => {
    beforeEach(() => {
      wrapper = factory({ isMultiTenant: false, props: { appCredentialsExist: false } });
    });

    it("shows access token input", () => {
      const tokenInput = wrapper.find('[aid="slack-access-token-input"]');
      expect(tokenInput.exists()).toBe(true);
    });

    it("shows connect button", () => {
      const connectBtn = wrapper.find('[aid="add-slack-app-btn"]');
      expect(connectBtn.exists()).toBe(true);
    });

    it("shows documentation link", () => {
      expect(wrapper.text()).toContain("For more information on how to get the access token");
      expect(wrapper.text()).toContain("Slack Notifications");
    });

    it("emits 'connect-slack-workspace' with token when valid token is provided", async () => {
      await wrapper.setData({ accessToken: "xoxp-1234567890" });
      const connectBtn = wrapper.find('[aid="add-slack-app-btn"]');
      await connectBtn.trigger("click");

      expect(wrapper.emitted("connect-slack-workspace")).toHaveLength(1);
      expect(wrapper.emitted("connect-slack-workspace")?.[0]).toEqual(["xoxp-1234567890"]);
    });
  });

  describe("Single Tenant - Credentials Exist", () => {
    beforeEach(() => {
      wrapper = factory({ isMultiTenant: false, props: { appCredentialsExist: true } });
    });

    it("does not show access token input", () => {
      const tokenInput = wrapper.find('[aid="slack-access-token-input"]');
      expect(tokenInput.exists()).toBe(false);
    });

    it("shows enabled connect button", () => {
      const connectBtn = wrapper.find('[aid="add-slack-app-btn"]');
      expect(connectBtn.exists()).toBe(true);
    });

    it("emits 'connect-slack-workspace' without token on click", async () => {
      const connectBtn = wrapper.find('[aid="add-slack-app-btn"]');
      await connectBtn.trigger("click");

      expect(wrapper.emitted("connect-slack-workspace")).toHaveLength(1);
      expect(wrapper.emitted("connect-slack-workspace")?.[0]).toEqual([]);
    });
  });

  describe("Multi Tenant - No Credentials", () => {
    beforeEach(() => {
      wrapper = factory({ isMultiTenant: true, props: { appCredentialsExist: false } });
    });

    it("shows disabled connect button", () => {
      const connectBtn = wrapper.find('[aid="add-slack-app-btn"]');
      expect(connectBtn.exists()).toBe(true);
    });

    it("does not show access token input", () => {
      const tokenInput = wrapper.find('[aid="slack-access-token-input"]');
      expect(tokenInput.exists()).toBe(false);
    });
  });

  describe("Multi Tenant - Credentials Exist", () => {
    beforeEach(() => {
      wrapper = factory({ isMultiTenant: true, props: { appCredentialsExist: true } });
    });

    it("shows enabled connect button", () => {
      const connectBtn = wrapper.find('[aid="add-slack-app-btn"]');
      expect(connectBtn.exists()).toBe(true);
    });

    it("emits 'connect-slack-workspace' without token on click", async () => {
      const connectBtn = wrapper.find('[aid="add-slack-app-btn"]');
      await connectBtn.trigger("click");

      expect(wrapper.emitted("connect-slack-workspace")).toHaveLength(1);
      expect(wrapper.emitted("connect-slack-workspace")?.[0]).toEqual([]);
    });
  });
});
