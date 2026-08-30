import { afterEach, beforeEach, describe, expect, it, vi, type MockInstance } from "vitest";
import { mount, type VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import LegacyRoleCol from "../legacy-role-col.vue";
import type { ILegacyRoleTransitionRow } from "@/models/legacy-access-rule-transition.model";
import type { ITableColumn } from "@/models/table.model";
import { ROLE_ROUTE_NAMES } from "@/router/role.routes/role.routes.names";

const { mockResolve } = vi.hoisted(() => ({
  mockResolve: vi.fn().mockReturnValue({ href: "/roles/10" }),
}));
vi.mock("vue-router", () => ({
  useRouter: vi.fn().mockReturnValue({ resolve: mockResolve }),
  useRoute: vi.fn().mockReturnValue({ name: "", query: {}, matched: [], params: {}, path: "", fullPath: "", hash: "" }),
}));

const mockRow = (overrides: Partial<ILegacyRoleTransitionRow> = {}): ILegacyRoleTransitionRow => ({
  id: "L1 researcher",
  legacyRole: "L1 researcher",
  roleId: 10,
  accessRuleCount: 3,
  newRole: "AI practitioner",
  ...overrides,
});

const mockColumn: ITableColumn = {
  name: "legacyRole",
  label: "Legacy role",
  field: (r: ILegacyRoleTransitionRow) => r.legacyRole,
  align: "left",
  style: "width: 206px; max-width: 206px",
};

interface ILegacyRoleColVm {
  openReview: () => void;
}

describe("legacy-role-col.vue", () => {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  let wrapper: VueWrapper<any>;
  let mockOpen: MockInstance;

  const createComponent = (row: ILegacyRoleTransitionRow = mockRow()) => {
    wrapper = mount(LegacyRoleCol, {
      global: {
        plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
        stubs: { RunaiSvgIcon: { template: "<span />" } },
      },
      props: { row, column: mockColumn },
    });
  };

  const vm = (): ILegacyRoleColVm => wrapper.vm as unknown as ILegacyRoleColVm;

  beforeEach(() => {
    mockOpen = vi.spyOn(window, "open").mockReturnValue(null);
    createComponent();
  });

  afterEach(() => {
    mockOpen.mockRestore();
    wrapper.unmount();
    vi.clearAllMocks();
  });

  describe("rendering", () => {
    it("displays the legacyRole name", () => {
      expect(wrapper.find(".legacy-role-text").text()).toBe("L1 researcher");
    });

    it("renders a details button", () => {
      expect(wrapper.find(".q-btn").exists()).toBe(true);
    });

    it("uses the cataloged Phosphor details icon", () => {
      const detailsButton = wrapper.find("[aid='view-legacy-role']");
      expect(detailsButton.findComponent({ name: "QIcon" }).props("name")).toBe("ph-fill ph-file-text");
    });
  });

  describe("openReview", () => {
    it("resolves the route with the row's roleId", () => {
      vm().openReview();
      expect(mockResolve).toHaveBeenCalledWith({
        name: ROLE_ROUTE_NAMES.ACCESS_RULES_ROLE_VIEW,
        params: { id: 10 },
      });
    });

    it("opens the resolved href in a new tab", () => {
      vm().openReview();
      expect(mockOpen).toHaveBeenCalledWith("/roles/10", "_blank", "noopener,noreferrer");
    });

    it("uses the correct roleId for different rows", () => {
      createComponent(mockRow({ roleId: 99 }));
      vm().openReview();
      expect(mockResolve).toHaveBeenCalledWith(expect.objectContaining({ params: { id: 99 } }));
    });
  });
});
