import { describe, test, expect, vi } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import RunaiSystemPage from "./runai-system-page.vue";
import { RunaiDistributionLogo } from "@/components/common/runai-distribution-logo";

describe("runai-system-page.vue", () => {
  let wrapper: VueWrapper<InstanceType<typeof RunaiSystemPage>>;

  function mountComponent(slots = {}) {
    wrapper = mount(RunaiSystemPage, {
      global: {
        plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
        stubs: {
          RunaiDistributionLogo: true,
        },
      },
      slots,
    });
  }

  test("renders the brand logo and the default slot content", () => {
    mountComponent({ default: "<div class='page-body'>body</div>" });

    const distributionLogo = wrapper.findComponent(RunaiDistributionLogo);
    expect(distributionLogo.exists()).toBe(true);
    expect(wrapper.find(".page-body").text()).toBe("body");
  });

  test("renders the header-end slot in the header", () => {
    mountComponent({ "header-end": "<span class='header-extra'>401</span>" });

    const headerExtra = wrapper.find("header .header-extra");
    expect(headerExtra.exists()).toBe(true);
    expect(headerExtra.text()).toBe("401");
  });
});
