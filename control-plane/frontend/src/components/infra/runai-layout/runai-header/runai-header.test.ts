import { shallowMount } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";

const storeState = vi.hoisted(() => ({
  isWhiteLabel: false,
}));

vi.mock("@/stores/app.store", () => ({
  useAppStore: () => ({
    config: { tenantName: "Runai" },
    pageTitleOverride: "New cluster",
    isWhiteLabel: storeState.isWhiteLabel,
    mainLogoDark: { src: "/distribution-assets/cc/vendor-logo-on-light.svg", height: "17px" },
    theme: { logo: { sub: "/distribution-assets/product-logo.svg" } },
  }),
}));

vi.mock("@/stores/auth.store", () => ({
  useAuthStore: () => ({ logo: undefined, tenant: { displayName: "Runai" } }),
}));

vi.mock("@/stores/permissions.store", () => ({
  usePermissionStore: () => ({ hasSomePermittedActions: () => true }),
}));

vi.mock("@/stores/cluster.store", () => ({
  useClusterStore: () => ({ clusters: [{}], isClustersLoaded: true }),
}));

import { RunaiDistributionLogo } from "@/components/common/runai-distribution-logo";
import { RunaiBrandLogo } from "@/components/common/runai-brand-logo";
import RunaiHeader from "./runai-header.vue";

function mountHeader(fullWidth = true, breadcrumbsFullPath = true) {
  return shallowMount(RunaiHeader, {
    props: { breadcrumbsFullPath },
    global: {
      mocks: {
        $route: {
          name: "cluster-new",
          meta: { fullWidth, pageTitle: "New cluster" },
        },
        $router: { currentRoute: { value: { name: "cluster-new" } } },
      },
      stubs: {
        QHeader: { template: "<header><slot /></header>" },
        QToolbar: { template: "<div class='header-toolbar'><slot /></div>" },
        QSpace: true,
        QSeparator: true,
      },
    },
  });
}

describe("runai header branding", () => {
  beforeEach(() => {
    storeState.isWhiteLabel = false;
    vi.clearAllMocks();
  });

  it("renders the distribution product logo in a standard full-path header", () => {
    const wrapper = mountHeader(false);

    const logo = wrapper.findComponent(RunaiDistributionLogo);
    expect(logo.exists()).toBe(true);
    expect(logo.props("composition")).toBe("product");
  });

  it("renders one combined distribution logo in a full-page header", () => {
    const wrapper = mountHeader(true);

    expect(wrapper.findAllComponents(RunaiDistributionLogo)).toHaveLength(1);
    expect(wrapper.findComponent(RunaiDistributionLogo).props("composition")).toBe("combined");
  });

  it("does not add a product logo when breadcrumbs do not show the full path", () => {
    const wrapper = mountHeader(false, false);

    expect(wrapper.findComponent(RunaiDistributionLogo).exists()).toBe(false);
  });

  it("preserves the white-label full-page branch with a profile-based theme URL", () => {
    storeState.isWhiteLabel = true;
    const wrapper = mountHeader(true);

    expect(wrapper.findComponent(RunaiDistributionLogo).exists()).toBe(false);
    expect(wrapper.findComponent(RunaiBrandLogo).props("img")).toBe("/distribution-assets/cc/vendor-logo-on-light.svg");
    expect(wrapper.find(".tenant-name").exists()).toBe(false);
  });

  it("preserves the white-label sub-logo branch with the public product asset", () => {
    storeState.isWhiteLabel = true;
    const wrapper = mountHeader(false);

    expect(wrapper.findComponent(RunaiDistributionLogo).exists()).toBe(false);
    expect(wrapper.find("img.sub-logo").attributes("src")).toBe("/distribution-assets/product-logo.svg");
  });
});
