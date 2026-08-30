import { mount } from "@vue/test-utils";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { defineComponent } from "vue";
import { ERROR_ROUTE_NAMES } from "@/router/error.routes/error.routes.names";

const LOADING_TIMEOUT_MS = 90 * 1000;

vi.mock("vue3-lottie", () => ({
  Vue3Lottie: defineComponent({
    name: "Vue3Lottie",
    template: "<div></div>",
  }),
}));

import RunaiAppLoader from "./runai-app-loader.vue";

const mockRouterReplace = vi.fn();

describe("RunaiAppLoader", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  const mountComponent = () => {
    return mount(RunaiAppLoader, {
      global: {
        mocks: {
          $router: {
            replace: mockRouterReplace,
          },
        },
      },
    });
  };

  it("renders correctly", () => {
    const wrapper = mountComponent();
    expect(wrapper.exists()).toBe(true);
    expect(wrapper.find(".runai-app-loader").exists()).toBe(true);
  });

  it("should not redirect before timeout", () => {
    mountComponent();
    vi.advanceTimersByTime(LOADING_TIMEOUT_MS - 1000);
    expect(mockRouterReplace).not.toHaveBeenCalled();
  });

  it("should redirect to unknown-timeout page after timeout", () => {
    mountComponent();
    vi.advanceTimersByTime(LOADING_TIMEOUT_MS);
    expect(mockRouterReplace).toHaveBeenCalledWith({
      name: ERROR_ROUTE_NAMES.UNKNOWN_TIMEOUT,
    });
  });

  it("should clear timeout when component is unmounted before timeout", () => {
    const wrapper = mountComponent();
    vi.advanceTimersByTime(LOADING_TIMEOUT_MS / 2);
    wrapper.unmount();
    vi.advanceTimersByTime(LOADING_TIMEOUT_MS);
    expect(mockRouterReplace).not.toHaveBeenCalled();
  });
});
