/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount } from "@vue/test-utils";
import { createTestingPinia } from "@pinia/testing";
import ProjectSection from "./project-section.vue";
import { beforeEach, describe, expect, test, vi } from "vitest";
import { projectsMockData } from "./project-section.mock";
import type { VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import { errorMessages } from "@/common/error-message.constant";
import { usePermissionStore } from "@/stores/permissions.store";
import type { ProjectMetadata } from "@/swagger-models/org-unit-service-client";
import { PROJECT_ROUTE_NAMES } from "@/router/project.routes/project.routes.names";

const pinia = createTestingPinia({ createSpy: vi.fn });
const permissionStore = usePermissionStore(pinia);

describe("project-section.vue", () => {
  let projectSectionCmp: VueWrapper<InstanceType<typeof ProjectSection>>;
  const props = {
    loading: false as boolean,
    projects: projectsMockData as ProjectMetadata[],
    selectedProjectId: null as number | null,
    clusterId: "123",
  };

  beforeEach(() => {
    projectSectionCmp = mount(ProjectSection, {
      global: {
        plugins: [Quasar],
        directives: {
          clickOutside: vi.fn(),
        },
      },
      props,
    });

    vi.mocked(permissionStore.hasPermission).mockImplementation(() => true);
  });

  test("component is loading properly", () => {
    expect(projectSectionCmp.html()).toContain("Projects");
    expect(projectSectionCmp.text()).toContain("Select the project in which your workload will run");
  });

  test("uses the Figma project-card text color hierarchy", () => {
    const projectCard = projectSectionCmp.get(".project-card");

    expect(projectCard.get(".text-subtitle1").classes()).toContain("text-b-main");
    expect(projectCard.get(".top-content .text-caption").classes()).toContain("text-b-label");
    expect(projectCard.get(".bottom-content").classes()).toContain("text-b-sub");
  });

  test("isRequiredProject method will return true when projectId is truthy and will return the proper text when projectId is falsey ", async () => {
    // note: this is a method specific test
    expect(projectSectionCmp.vm.isRequiredProject("")).toBe(errorMessages.SELECT_PROJECT);
    expect(projectSectionCmp.vm.isRequiredProject("1")).toBe(true);
  });

  test("the component will emit the updated selected projectId correctly when a new selection is made", async () => {
    const selectedProject = projectsMockData[0];
    await projectSectionCmp.find(`[aid=${selectedProject.name}-card]`).trigger("click");
    const expectedEmit: number[][] = [[+selectedProject.id]];
    expect(projectSectionCmp.emitted("project-changed")).toEqual(expectedEmit);
  });

  test("the component will emit 'null' when the selected project is clicked again", async () => {
    const selectedProject = projectsMockData[0];
    await projectSectionCmp.setProps({
      selectedProjectId: +selectedProject.id,
    });

    await projectSectionCmp.find(`[aid=${selectedProject.name}-card]`).trigger("click");
    const expectedEmit: null[][] = [[null]];
    expect(projectSectionCmp.emitted("project-changed")).toEqual(expectedEmit);
  });

  describe("createProject", () => {
    const mockRouter = { push: vi.fn() };

    beforeEach(() => {
      mockRouter.push.mockClear();
      projectSectionCmp = mount(ProjectSection, {
        global: {
          plugins: [Quasar],
          directives: {
            clickOutside: vi.fn(),
          },
          mocks: {
            $router: mockRouter,
          },
        },
        props,
      });
      vi.mocked(permissionStore.hasPermission).mockImplementation(() => true);
    });

    test("navigates to PROJECT_NEW", () => {
      projectSectionCmp.vm.createProject();

      expect(mockRouter.push).toHaveBeenCalledWith({
        name: PROJECT_ROUTE_NAMES.PROJECT_NEW,
        query: { clusterId: props.clusterId },
      });
    });
  });
});
