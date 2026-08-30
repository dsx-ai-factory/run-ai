import { mount, type VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import { beforeEach, describe, expect, it } from "vitest";

import { ClaimInfoSection } from "./";
import { VolumeSizeSection } from "@/components/section/volume-section/volume-box/volume-size-section";
import type WrapperLike from "node_modules/@vue/test-utils/dist/interfaces/wrapperLike";
import type { Component } from "vue";

const selectors = {
  storageClassSelect: ".storage-class-select",
  accessModeSelect: ".access-mode-select",
  volumeModeSelect: ".volume-mode-select",
  volumeSizeSectionCmp: ".volume-size-section",
};

function getStorageClassSelectCmp(wrapper: VueWrapper<InstanceType<typeof ClaimInfoSection>>): WrapperLike {
  return wrapper.findComponent(selectors.storageClassSelect);
}

function getAccessModeSelectCmp(wrapper: VueWrapper<InstanceType<typeof ClaimInfoSection>>): WrapperLike {
  return wrapper.findComponent(selectors.accessModeSelect);
}

function getVolumeModeSelectCmp(wrapper: VueWrapper<InstanceType<typeof ClaimInfoSection>>): WrapperLike {
  return wrapper.findComponent(selectors.volumeModeSelect);
}

function getVolumeSizeSectionCmp(wrapper: VueWrapper<InstanceType<typeof ClaimInfoSection>>): WrapperLike {
  return wrapper.findComponent(selectors.volumeSizeSectionCmp);
}

describe("claim-info-section.vue", () => {
  let wrapper: VueWrapper<InstanceType<typeof ClaimInfoSection>>;
  function factory(props = {}): VueWrapper<InstanceType<typeof ClaimInfoSection>> {
    return mount(ClaimInfoSection, {
      props: {
        // define default props here
        ...props,
        modelValue: {
          storageClass: null,
        },
        storageClasses: [],
      },
      emits: ["update:model-value"],
      global: {
        plugins: [Quasar],
        components: {
          VolumeSizeSection,
        },
      },
    });
  }

  beforeEach(() => {
    wrapper = factory();
  });

  describe("Html elements check", () => {
    it("Component exists", () => {
      expect(wrapper.exists()).toBe(true);
    });

    it("All inputs exists", () => {
      const storageClassSelect = getStorageClassSelectCmp(wrapper);
      const accessModeSelect = getAccessModeSelectCmp(wrapper);
      const volumeSizeSectionCmp = getVolumeSizeSectionCmp(wrapper);
      const volumeModeSelect = getVolumeModeSelectCmp(wrapper);

      expect(storageClassSelect.exists()).toBe(true);
      expect(accessModeSelect.exists()).toBe(true);
      expect(volumeSizeSectionCmp.exists()).toBe(true);
      expect(volumeModeSelect.exists()).toBe(true);
    });
  });

  describe("Props", () => {
    describe("storageClasses", () => {
      it("should render storage class options", async () => {
        const storageClasses = [
          { storageClassName: "standard", clusterId: "1", spec: { is_default: true } },
          { storageClassName: "premium", clusterId: "1", spec: { is_default: false } },
        ];
        await wrapper.setProps({ storageClasses });

        const storageClassSelectCmp: Component = getStorageClassSelectCmp(wrapper);

        expect(storageClassSelectCmp).toBeTruthy();
        expect(storageClassSelectCmp.props("options").length).toEqual(wrapper.vm.storageClasses.length);
      });
    });

    describe("modelValue", () => {
      it("should set the storage class value", async () => {
        const storageClass = "standard";
        await wrapper.setProps({ modelValue: { storageClass } });

        const storageClassSelectCmp: Component = getStorageClassSelectCmp(wrapper);
        expect(storageClassSelectCmp.props("modelValue")).toEqual(storageClass);
      });

      it("should set the access mode value", async () => {
        const accessModeOptions = wrapper.vm.accessModeOptions;

        await wrapper.setProps({
          modelValue: { accessModes: { [accessModeOptions[0].value as string]: true }, storageClass: null },
        });

        const accessModeSelectCmp: Component = getAccessModeSelectCmp(wrapper);
        expect(accessModeSelectCmp.props("modelValue")).toEqual([accessModeOptions[0].value]);
      });

      it("should set the volume mode value", async () => {
        const volumeMode = "Filesystem";
        await wrapper.setProps({ modelValue: { volumeMode, storageClass: null } });

        const volumeModeSelectCmp: Component = getVolumeModeSelectCmp(wrapper);
        expect(volumeModeSelectCmp.props("modelValue")).toEqual(volumeMode);
      });

      it("should set the volume size value", async () => {
        const volumeSize = "10G";
        await wrapper.setProps({ modelValue: { size: volumeSize, storageClass: null } });

        const volumeSizeSectionCmp: Component = getVolumeSizeSectionCmp(wrapper);
        expect(volumeSizeSectionCmp.props("modelValue")).toEqual(volumeSize);
      });
    });

    describe("loading", () => {
      it("should pass loading prop to inner components", async () => {
        await wrapper.setProps({ loading: true });

        const accessModeSelectCmp: Component = getAccessModeSelectCmp(wrapper);
        const volumeSizeSectionCmp: Component = getVolumeSizeSectionCmp(wrapper);
        const volumeModeSelectCmp: Component = getVolumeModeSelectCmp(wrapper);

        expect(accessModeSelectCmp.props("loading")).toBe(true);
        expect(volumeSizeSectionCmp.props("loading")).toBe(true);
        expect(volumeModeSelectCmp.props("loading")).toBe(true);
      });

      it("should pass loading = false to inner components by default", () => {
        const accessModeSelectCmp: Component = getAccessModeSelectCmp(wrapper);
        const volumeSizeSectionCmp: Component = getVolumeSizeSectionCmp(wrapper);
        const volumeModeSelectCmp: Component = getVolumeModeSelectCmp(wrapper);

        expect(accessModeSelectCmp.props("loading")).toBe(false);
        expect(volumeSizeSectionCmp.props("loading")).toBe(false);
        expect(volumeModeSelectCmp.props("loading")).toBe(false);
      });
    });
  });

  // describe("Emits", () => {
  //   describe("storage class changes", () => {
  //     it("should emit the correct event when the storage class is changed", async () => {
  //       const storageClassSelect: Component = getStorageClassSelectCmp(wrapper);

  //       await storageClassSelect.vm.$emit("update:modelValue", "standard");
  //       expect(wrapper.emitted("update:model-value")).toEqual([[{ storageClass: "standard" }]]);
  //     });

  //     it("should emit the runai storage class properly", async () => {
  //       const storageClassSelect: Component = getStorageClassSelectCmp(wrapper);

  //       await storageClassSelect.vm.$emit("update:modelValue", "None");
  //       expect(wrapper.emitted("update:model-value")).toEqual([[{ storageClass: "runai-none" }]]);
  //     });
  //   });

  //   describe("access mode changes", () => {
  //     it("should emit the correct event when the access mode is changed", async () => {
  //       const accessModeSelect: Component = getAccessModeSelectCmp(wrapper);

  //       await accessModeSelect.vm.$emit("update:modelValue", ["readWriteOnce", "readWriteMany"]);
  //       expect(wrapper.emitted("update:model-value")).toEqual([
  //         [{ accessModes: { readWriteOnce: true, readWriteMany: true }, storageClass: null }],
  //       ]);
  //     });
  //   });
  // });
});
