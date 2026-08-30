import type { DOMWrapper, VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { beforeEach, describe, expect, test } from "vitest";
import { VolumeBox } from ".";
import type { IUIVolume } from "@/models/data-source.model";
import type { TTestEmit } from "@/models/unit-test.model";
import type { Component } from "vue";
describe("volume-box.vue", () => {
  let volumeBoxCmp: VueWrapper<InstanceType<typeof VolumeBox>>;
  // Common selectors
  const pathInputSelector = '[aid="path-input"]';
  const removeButtonSelector = '[aid="wrapper-x-btn"]';

  beforeEach(() => {
    volumeBoxCmp = mount(VolumeBox, {
      global: {
        plugins: [Quasar],
      },
      props: {
        volume: {
          ephemeral: false,
          claimName: "claim-2413",
          path: "/data",
          claimInfo: {
            size: "3G",
            storageClass: "standard",
            accessModes: {
              readWriteOnce: true,
            },
          },
        },
        storageClasses: [{ storageClassName: "standard", clusterId: "cluster-1", spec: { is_default: true } }],
      },
    });
  });

  test("component is loading properly", () => {
    expect(volumeBoxCmp.html()).toContain("volume");
  });

  test("component emits the correct event when the path is changed", async () => {
    const pathInput: DOMWrapper<HTMLInputElement> = volumeBoxCmp.find(pathInputSelector);
    const newPathValue = "/data/local";
    await pathInput.setValue(newPathValue);
    const expectedEmit: TTestEmit<IUIVolume> = [
      [
        {
          ...volumeBoxCmp.props("volume"),
          path: newPathValue,
        },
      ],
    ];
    expect(volumeBoxCmp.emitted("update-volume")).toEqual(expectedEmit);
  });

  test("component emits the correct event when the storage class is changed", async () => {
    const storageClassSelect: Component = volumeBoxCmp.findComponent(".storage-class-select");
    await storageClassSelect.vm.$emit("update:model-value", "");
    const expectedEmit: TTestEmit<IUIVolume> = [
      [
        {
          ...volumeBoxCmp.props("volume"),
          claimInfo: {
            ...volumeBoxCmp.props("volume").claimInfo,
            storageClass: "",
          },
        },
      ],
    ];
    expect(volumeBoxCmp.emitted("update-volume")).toEqual(expectedEmit);
  });

  test("component emits the correct event when the access mode is changed", async () => {
    const accessModeSelect: Component = volumeBoxCmp.findComponent(".access-mode-select");
    await accessModeSelect.vm.$emit("update:model-value", ["readWriteOnce", "readOnlyMany"]);
    const expectedEmit: TTestEmit<IUIVolume> = [
      [
        {
          ...volumeBoxCmp.props("volume"),
          claimInfo: {
            ...volumeBoxCmp.props("volume").claimInfo,
            accessModes: {
              readWriteOnce: true,
              readOnlyMany: true,
            },
          },
        },
      ],
    ];
    expect(volumeBoxCmp.emitted("update-volume")).toEqual(expectedEmit);
  });

  test("component emits the correct event when the size is changed", async () => {
    const sizeSectionCmp: Component = volumeBoxCmp.findComponent(".volume-size-section");
    await sizeSectionCmp.vm.$emit("update:model-value", "5G");
    const expectedEmit: TTestEmit<IUIVolume> = [
      [
        {
          ...volumeBoxCmp.props("volume"),
          claimInfo: {
            ...volumeBoxCmp.props("volume").claimInfo,
            size: "5G",
          },
        },
      ],
    ];
    expect(volumeBoxCmp.emitted("update-volume")).toEqual(expectedEmit);
  });

  test("component should emit 'remove-volume' when the user click the 'x' button", async () => {
    const removeButton: DOMWrapper<HTMLButtonElement> = volumeBoxCmp.find(removeButtonSelector);
    await removeButton.trigger("click");
    expect(volumeBoxCmp.emitted("remove-volume")).toBeTruthy();
  });

  test("component should hide persistence options", async () => {
    expect(volumeBoxCmp.html()).toContain("Volume persistency");
    await volumeBoxCmp.setProps({
      showVolumePersistencySection: false,
    });

    expect(volumeBoxCmp.html()).not.toContain("Volume persistency");
  });
});
