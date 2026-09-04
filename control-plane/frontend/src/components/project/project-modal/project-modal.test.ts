/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { afterEach, beforeEach, describe, expect, type MockedFunction, test, vi } from "vitest";
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils";
import { ProjectModal } from "@/components/project/project-modal/";
import { Quasar } from "quasar";
import { createTestingPinia, type TestingPinia } from "@pinia/testing";
import { EProjectModalEntity, type IProjectsModalOptions } from "@/models/project.model";
import { unitTestUtil } from "@/utils/unit-test.util";
import { setActivePinia } from "pinia";
import { projectsMock } from "@/components/project/project-modal/project-modal.mock";
import { useClusterStore } from "@/stores/cluster.store";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";

const pinia: TestingPinia = createTestingPinia({ createSpy: vi.fn });
setActivePinia(pinia);
vi.mock("@/services/control-plane/org-unit.service/org-unit.service");

describe("project-modal component", () => {
  let wrapper: VueWrapper<InstanceType<typeof ProjectModal>>;
  const clusterStore = useClusterStore();
  clusterStore.$patch({
    clusters: [
      {
        uuid: "1c16fffa-fd84-4bff-a148-798e305b8d60",
        tenantId: 1001,
        name: "barak-1720974608",
        domain: "https://barak.runailabs.com",
        version: "0.0.0-10000136.master-cabf8e4",
      },
    ],
  });
  const mockGetProjects = orgUnitService.getProjects as MockedFunction<typeof orgUnitService.getProjects>;
  const createComponent = (props: { modalOptions: IProjectsModalOptions }): void => {
    wrapper = mount(ProjectModal, {
      global: {
        plugins: [Quasar, pinia],
        directives: {
          permission: vi.fn(),
        },
      },
      props: props,
    });
    wrapper.vm.$q.notify = vi.fn(); // disabling notify to prevent it from failing tests
  };

  beforeEach(async () => {
    mockGetProjects.mockResolvedValue(projectsMock);

    createComponent({
      modalOptions: {
        parentId: "1",
        entityName: "rnd",
        entityType: EProjectModalEntity.Department,
        clusterId: "1c16fffa-fd84-4bff-a148-798e305b8d60",
      },
    });

    await flushPromises();
  });
  afterEach(() => {
    wrapper.unmount();
    vi.clearAllMocks();
  });

  test("Table Rendering Test", () => {
    expect(wrapper.vm.$data.tableFilter.searchTerm).toBe("");
    expect(wrapper.findComponent(".runai-base-dialog-header").text()).eql("Projects Associated with Department rnd");
  });

  test("Close Modal Test", async () => {
    const modalEl = unitTestUtil.getModalElement();
    await modalEl.find("[aid=close-projects-modal-btn]").trigger("click");
    expect(wrapper.emitted("close")).toBeTruthy();
  });

  test("Modal load current entity data correctly", async () => {
    const modalEl = unitTestUtil.getModalElement();
    const projectNameCell = modalEl.find('[aid="name"]');
    const numberOfRows = modalEl.findAll(".runai-table-row");
    expect(projectNameCell.text()).toBe("suricata");
    expect(numberOfRows.length).toEqual(1);
  });

  test("Search Functionality Test", async () => {
    const modalEl = unitTestUtil.getModalElement();
    await modalEl.find("[aid=runai-list-modal-search]").setValue("not-found-project");
    expect(modalEl.find(".empty-state-message").exists()).toBe(true);
  });
});
