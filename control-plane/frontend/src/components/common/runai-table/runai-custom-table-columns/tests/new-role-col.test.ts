import { afterEach, beforeEach, describe, expect, it, vi, type MockInstance } from "vitest";
import { mount, type VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import NewRoleCol from "../new-role-col.vue";
import { ECustomCell } from "@/models/table.model";
import type { ILegacyRoleTransitionRow } from "@/models/legacy-access-rule-transition.model";
import type { ITableColumn } from "@/models/table.model";

const { mockResolve } = vi.hoisted(() => ({
  mockResolve: vi.fn().mockReturnValue({ href: "/roles/42" }),
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

const makeColumn = (
  roleOptions = ["AI practitioner", "Project administrator"],
  getReviewRoleId: (_: string) => number | undefined = (_: string) => 42,
): ITableColumn => ({
  name: "newRole",
  label: "New role",
  field: (r: ILegacyRoleTransitionRow) => r.newRole,
  align: "left",
  customCell: ECustomCell.NEW_ROLE_COL,
  customCellConfig: {
    [ECustomCell.NEW_ROLE_COL]: { roleOptions, getReviewRoleId },
  },
});

interface INewRoleColVm {
  roleOptions: string[];
  onRoleSelect: (role: string) => void;
  openReview: () => void;
}

describe("new-role-col.vue", () => {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  let wrapper: VueWrapper<any>;
  let mockOpen: MockInstance;

  const createComponent = (row: ILegacyRoleTransitionRow = mockRow(), column: ITableColumn = makeColumn()) => {
    wrapper = mount(NewRoleCol, {
      global: {
        plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
        stubs: { RunaiSvgIcon: { template: "<span />" } },
      },
      props: { row, column },
    });
  };

  const vm = (): INewRoleColVm => wrapper.vm as unknown as INewRoleColVm;

  beforeEach(() => {
    mockOpen = vi.spyOn(window, "open").mockReturnValue(null);
    createComponent();
  });

  afterEach(() => {
    mockOpen.mockRestore();
    wrapper.unmount();
    vi.clearAllMocks();
  });

  describe("roleOptions", () => {
    it("reads options from customCellConfig", () => {
      expect(vm().roleOptions).toEqual(["AI practitioner", "Project administrator"]);
    });

    it("falls back to empty array when config is missing", () => {
      createComponent(mockRow(), { ...makeColumn(), customCellConfig: undefined });
      expect(vm().roleOptions).toEqual([]);
    });
  });

  describe("onRoleSelect", () => {
    it("mutates row.newRole with the selected value", () => {
      const row = mockRow({ newRole: undefined });
      createComponent(row);
      vm().onRoleSelect("Project administrator");
      expect(row.newRole).toBe("Project administrator");
    });

    it("emits custom-cell-event after selection", () => {
      vm().onRoleSelect("AI practitioner");
      expect(wrapper.emitted("custom-cell-event")).toHaveLength(1);
    });
  });

  describe("openReview", () => {
    it("opens the role page in a new tab using the resolved href", () => {
      vm().openReview();
      expect(mockResolve).toHaveBeenCalled();
      expect(mockOpen).toHaveBeenCalledWith("/roles/42", "_blank", "noopener,noreferrer");
    });

    it("does nothing when getReviewRoleId returns undefined", () => {
      createComponent(
        mockRow(),
        makeColumn(["AI practitioner"], () => undefined),
      );
      vm().openReview();
      expect(mockOpen).not.toHaveBeenCalled();
    });

    it("does nothing when getReviewRoleId returns undefined for the empty-string fallback", () => {
      createComponent(
        mockRow({ newRole: undefined }),
        makeColumn(["AI practitioner"], () => undefined),
      );
      vm().openReview();
      expect(mockOpen).not.toHaveBeenCalled();
    });
  });

  describe("details button visibility", () => {
    it("is hidden when row.newRole is undefined", () => {
      createComponent(mockRow({ newRole: undefined }));
      expect(wrapper.find("[aid='view-new-role']").exists()).toBe(false);
    });

    it("is visible when row.newRole is set", () => {
      createComponent(mockRow({ newRole: "AI practitioner" }));
      expect(wrapper.find("[aid='view-new-role']").exists()).toBe(true);
    });

    it("uses the cataloged Phosphor details icon", () => {
      const detailsButton = wrapper.find("[aid='view-new-role']");
      expect(detailsButton.findComponent({ name: "QIcon" }).props("name")).toBe("ph-fill ph-file-text");
    });
  });
});
