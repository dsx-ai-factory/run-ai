import { afterEach, describe, expect, it, vi } from "vitest";
import { shallowMount } from "@vue/test-utils";
import { createTestingPinia } from "@pinia/testing";
import DOCS from "@/constants/docs";
import SidebarFooter from "@/components/onboarding/common/sidebar-footer/sidebar-footer.vue";
import DelegatedTeamSetup from "@/components/onboarding/steps/team-setup/delegated-team-setup/delegated-team-setup.vue";

const support = DOCS.SUPPORT as { URL: string };
const originalSupportUrl = support.URL;

const mountDelegatedTeamSetup = () =>
  shallowMount(DelegatedTeamSetup, {
    props: { clusterId: "cluster-1" },
    global: {
      plugins: [createTestingPinia({ createSpy: vi.fn, stubActions: true })],
      stubs: {
        QForm: { template: "<form><slot /></form>" },
      },
    },
  });

describe("onboarding support layout", () => {
  afterEach(() => {
    support.URL = originalSupportUrl;
  });

  it("keeps the sidebar footer layout slot without rendering support copy", () => {
    support.URL = "";
    const wrapper = shallowMount(SidebarFooter);

    expect(wrapper.classes()).toContain("support-unavailable");
    expect(wrapper.text()).toBe("");
    expect(wrapper.find("a").exists()).toBe(false);
  });

  it("keeps the delegated footer layout slot without rendering support copy", () => {
    support.URL = "";
    const footer = mountDelegatedTeamSetup().find(".help-section");

    expect(footer.classes()).toContain("support-unavailable");
    expect(footer.text()).toBe("");
    expect(footer.find("a").exists()).toBe(false);
  });

  it("renders both onboarding support links when the URL is configured", () => {
    support.URL = "https://support.example.test/contact";
    const sidebarFooter = shallowMount(SidebarFooter);
    const delegatedFooter = mountDelegatedTeamSetup().find(".help-section");

    expect(sidebarFooter.text().replace(/\s+/g, " ")).toContain("Having trouble? Get help");
    expect(sidebarFooter.get("a").attributes("href")).toBe(support.URL);
    expect(delegatedFooter.text().replace(/\s+/g, " ")).toContain("Having trouble? Get help");
    expect(delegatedFooter.get("a").attributes("href")).toBe(support.URL);
  });
});
