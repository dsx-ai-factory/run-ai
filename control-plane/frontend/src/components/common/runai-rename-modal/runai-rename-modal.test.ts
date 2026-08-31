/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// tests/unit/RenameModal.spec.ts
import type { VueWrapper } from "@vue/test-utils";
import { flushPromises, mount } from "@vue/test-utils";
import { RunaiRenameModal } from ".";
import { Quasar, QForm, QInput, QBtn } from "quasar";
import { AssetKind, Scope } from "@/swagger-models/assets-service-client";
import * as formValidators from "@/common/form.validators";
import type { IScopeModel } from "@/models/global.model";

describe("runai-rename-modal.vue", () => {
  const oldName = "old-name";
  const oldDescription = "old description";
  const newName = "new-name";
  const newDescription = "new description";

  vi.spyOn(formValidators, "isAssetNameUnique").mockReturnValue(() => Promise.resolve(true));

  function factory(props = {}): VueWrapper<InstanceType<typeof RunaiRenameModal>> {
    return mount(RunaiRenameModal, {
      props: {
        modelValue: true,
        existingName: oldName,
        existingDescription: oldDescription,
        renaming: false,
        assetType: AssetKind.Pvc,
        scopeModel: {
          scope: Scope.Cluster,
          clusterId: "cluster-1",
          departmentId: "department-1",
          projectId: 1,
        } as IScopeModel,
        ...props,
      },
      global: {
        plugins: [Quasar],
        components: {
          QForm,
          QInput,
          QBtn,
        },
        mocks: {
          $route: {
            query: {},
          },
        },
        stubs: {
          RunaiBaseDialog: {
            template: '<div><slot name="header"></slot><slot name="body"></slot><slot name="footer"></slot></div>',
          },
        },
      },
    });
  }

  describe("Html elements check", () => {
    it("Component exists", () => {
      const wrapper = factory();
      expect(wrapper.exists()).toBe(true);
      expect(wrapper.classes()).toContain("runai-rename-modal");
    });

    it("Component contains 2 inputs", () => {
      const wrapper = factory();
      const inputs = wrapper.findAllComponents(QInput);
      expect(inputs.length).toBe(2);
    });
  });

  describe("Props", () => {
    it("name input has the old name value", () => {
      const wrapper = factory();
      expect(wrapper.findAllComponents(QInput)[0].props().modelValue).toBe(oldName);
    });

    it("description input has the old description value", () => {
      const wrapper = factory();
      expect(wrapper.findAllComponents(QInput)[1].props().modelValue).toBe(oldDescription);
    });

    it("description input is hidden if description is not provided", () => {
      const wrapper = factory({ existingDescription: undefined });
      expect(wrapper.findAllComponents(QInput).length).toBe(1);
    });
  });

  describe("Emits", () => {
    it("emits 'close' event on cancel button click", async () => {
      const wrapper = factory();
      await wrapper.find('[aid="rename-modal-cancel-button"]').trigger("click");
      expect(wrapper.emitted("close")).toBeTruthy();
    });

    it("emits 'close' event when name and description aren't changed", async () => {
      const wrapper = factory();
      await wrapper.find('[aid="rename-modal-save-button"]').trigger("click");
      expect(wrapper.emitted("close")).toBeTruthy();
    });

    it("emits 'rename' event with new name and description", async () => {
      const wrapper = factory();
      const nameInputCmp = wrapper.findAllComponents(QInput)[0];
      const descriptionInputCmp = wrapper.findAllComponents(QInput)[1];

      nameInputCmp.setValue(newName);
      descriptionInputCmp.setValue(newDescription);

      await wrapper.find('[aid="rename-modal-save-button"]').trigger("click");
      await flushPromises(); // Validating form is async
      expect(wrapper.emitted("rename")).toBeTruthy();
    });
  });
});
//
