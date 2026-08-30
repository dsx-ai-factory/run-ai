import type { DOMWrapper, VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { beforeEach, describe, expect, test } from "vitest";
import { VolumeSection } from ".";
import type { TTestEmit } from "@/models/unit-test.model";
import type { IUIVolume } from "@/models/data-source.model";
import type { Component, DefineComponent } from "vue";

describe("volume-section.vue", () => {
  let volumeSectionCmp: VueWrapper<InstanceType<typeof VolumeSection>>;

  beforeEach(() => {
    volumeSectionCmp = mount(VolumeSection, {
      global: {
        plugins: [Quasar],
      },
      props: {
        volumes: [],
        storageClasses: [],
      },
    });
  });

  test("component is loading properly", () => {
    expect(volumeSectionCmp.html()).toContain("Volume");
  });

  test("component should not render volume boxes if volumes array is empty", () => {
    expect(volumeSectionCmp.findAll(".volume-box").length).toBe(0);
  });

  test("component should render volume boxes according to the volumes array length", async () => {
    await volumeSectionCmp.setProps({
      volumes: [
        {
          ephemeral: false,
          claimName: "claim-2413",
          path: "/data",
          claimInfo: { size: "3G", storageClass: "standard", accessModes: { readWriteOnce: true } },
        },
      ],
    });
    expect(volumeSectionCmp.findAll(".volume-box").length).toBe(volumeSectionCmp.props("volumes").length);
  });

  test("component should not render add button if volumes array is not empty", async () => {
    await volumeSectionCmp.setProps({
      volumes: [
        {
          ephemeral: false,
          claimName: "claim-2413",
          path: "/data",
          claimInfo: { size: "3G", storageClass: "standard", accessModes: { readWriteOnce: true } },
        },
      ],
    });
    expect(volumeSectionCmp.find("[aid=add-volume-btn]").exists()).toBe(false);
  });

  test("component should render add button if volumes array is empty", async () => {
    expect(volumeSectionCmp.find("[aid=add-volume-btn]").exists()).toBe(true);
  });

  // test that checks that when the add button is clicked, the update-volumes emit is emitted
  test("component should emit update-volumes event with the updated volumes when add button is clicked", async () => {
    await volumeSectionCmp.find("[aid=add-volume-btn]").trigger("click");
    expect((volumeSectionCmp.emitted("update-volumes") as TTestEmit<Array<IUIVolume>>)[0][0].length).toBe(1);
  });

  test("component should emit update-volumes event with the updated volumes when volume is deleted", async () => {
    await volumeSectionCmp.setProps({
      volumes: [
        {
          ephemeral: false,
          claimName: "claim-2413",
          path: "/data",
          claimInfo: { size: "3G", storageClass: "standard", accessModes: { readWriteOnce: true } },
        },
      ],
    });
    const removeButton: DOMWrapper<HTMLButtonElement> = volumeSectionCmp.find('[aid="wrapper-x-btn"]');
    await removeButton.trigger("click");
    expect((volumeSectionCmp.emitted("update-volumes") as TTestEmit<Array<IUIVolume>>)[0][0].length).toBe(0);
  });

  test("component should emit update-volumes event with the updated volumes when volume is updated", async () => {
    const volume: IUIVolume = {
      claimName: "claim-2413",
      path: "/data",
      ephemeral: false,
      claimInfo: { size: "3G", storageClass: "standard", accessModes: { readWriteOnce: true } },
    };
    await volumeSectionCmp.setProps({
      volumes: [volume],
    });
    const volumeBoxCmp: Component = volumeSectionCmp.findComponent<DefineComponent>(".volume-box");
    await volumeBoxCmp.vm.$emit("update-volume", { ...volume, path: "/data2" });
    expect((volumeSectionCmp.emitted("update-volumes") as TTestEmit<Array<IUIVolume>>)[0][0]).toEqual([
      { ...volume, path: "/data2" },
    ]);
  });

  test("component should hide persistence options", async () => {
    const volume: IUIVolume = {
      claimName: "claim-2413",
      path: "/data",
      ephemeral: false,
      claimInfo: { size: "3G", storageClass: "standard", accessModes: { readWriteOnce: true } },
    };
    await volumeSectionCmp.setProps({
      volumes: [volume],
    });

    let volumeBoxCmp: Component = volumeSectionCmp.findComponent<DefineComponent>(".volume-box");
    expect(volumeBoxCmp.html()).toContain("Volume persistency");

    await volumeSectionCmp.setProps({
      showVolumePersistencySection: false,
    });

    volumeBoxCmp = volumeSectionCmp.findComponent<DefineComponent>(".volume-box");
    expect(volumeBoxCmp.html()).not.toContain("Volume persistency");
  });
});
