/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { afterEach, describe, expect, it, vi } from "vitest";
import { flushPromises, mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import TemplateNameCol from "../template-name-col.vue";
import { EWorkloadType } from "@/models/workload.model";
import {
  LLM_IMAGE_FOR_HUGGING_FACE,
  LLM_IMAGE_FOR_TGI,
  TGI_DEPRECATION_MESSAGE_EDIT,
  TGI_DEPRECATION_MESSAGE_READ,
} from "@/models/inference.model";
import { Scope } from "@/swagger-models/assets-service-client";
import type { ITableColumn } from "@/models/table.model";
import type { SupersetTemplate } from "@/swagger-models/workloads-client";
import { usePermissionStore } from "@/stores/permissions.store";
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";

vi.mock("@/services/control-plane/rbac/org-tree.service/org-tree.service", () => ({
  orgTreeService: { getExpandedPermittedScopes: vi.fn() },
}));

const INFO_ICON = '[aid="template-info-icon"]';
const column = { name: "name", label: "Template", field: (r: SupersetTemplate) => r.meta.name } as ITableColumn;

function createRow(overrides: { image?: string; workloadType?: EWorkloadType; scopeType?: Scope } = {}) {
  const { image = LLM_IMAGE_FOR_TGI, workloadType = EWorkloadType.Inference, scopeType = Scope.Project } = overrides;
  return {
    meta: { name: "my-template", workloadType, scopeType, scopeId: "p1" },
    spec: { image },
  } as unknown as SupersetTemplate;
}

function mountCell(row: SupersetTemplate, canUpdate = true, permittedProjects: string[] = ["p1"]) {
  const pinia = createTestingPinia({ createSpy: vi.fn });
  vi.mocked(usePermissionStore().hasPermission).mockReturnValue(canUpdate);
  vi.mocked(orgTreeService.getExpandedPermittedScopes).mockReturnValue({ projects: permittedProjects });
  return mount(TemplateNameCol, { global: { plugins: [Quasar, pinia] }, props: { row, column } });
}

describe("template-name-col.vue", () => {
  it("should render the name without an info icon for a vLLM template", () => {
    const wrapper = mountCell(createRow({ image: LLM_IMAGE_FOR_HUGGING_FACE }));

    expect(wrapper.text()).toContain("my-template");
    expect(wrapper.find(INFO_ICON).exists()).toBe(false);
  });

  it("should not evaluate the permissions for a template that is not TGI", () => {
    mountCell(createRow({ image: LLM_IMAGE_FOR_HUGGING_FACE }));

    expect(orgTreeService.getExpandedPermittedScopes).not.toHaveBeenCalled();
  });

  it("should not show the info icon for a non-inference template with the TGI image", () => {
    const wrapper = mountCell(createRow({ workloadType: EWorkloadType.Training }));

    expect(wrapper.find(INFO_ICON).exists()).toBe(false);
  });

  it.each([
    {
      name: "the user can edit the template",
      canUpdate: true,
      projects: ["p1"],
      scope: Scope.Project,
      want: TGI_DEPRECATION_MESSAGE_EDIT,
    },
    {
      name: "the user has no update permission",
      canUpdate: false,
      projects: ["p1"],
      scope: Scope.Project,
      want: TGI_DEPRECATION_MESSAGE_READ,
    },
    {
      name: "the template is outside the permitted scopes",
      canUpdate: true,
      projects: ["other"],
      scope: Scope.Project,
      want: TGI_DEPRECATION_MESSAGE_READ,
    },
    {
      name: "the template is a system template",
      canUpdate: true,
      projects: ["p1"],
      scope: Scope.System,
      want: TGI_DEPRECATION_MESSAGE_READ,
    },
  ])("should show the matching message when $name", ({ canUpdate, projects, scope, want }) => {
    const wrapper = mountCell(createRow({ scopeType: scope }), canUpdate, projects);

    expect(wrapper.find(INFO_ICON).attributes("aria-label")).toBe(`More information: ${want}`);
  });

  describe("truncated names", () => {
    const setWidths = (scrollWidth: number, clientWidth: number) => {
      vi.spyOn(HTMLElement.prototype, "scrollWidth", "get").mockReturnValue(scrollWidth);
      vi.spyOn(HTMLElement.prototype, "clientWidth", "get").mockReturnValue(clientWidth);
    };

    afterEach(() => {
      vi.restoreAllMocks();
    });

    it("should keep the info icon next to a name that is cut with an ellipsis", async () => {
      setWidths(500, 200);
      const wrapper = mountCell(createRow());
      await flushPromises();

      expect(wrapper.find(".template-name-text").exists()).toBe(true);
      expect(wrapper.find(INFO_ICON).exists()).toBe(true);
    });
  });
});
