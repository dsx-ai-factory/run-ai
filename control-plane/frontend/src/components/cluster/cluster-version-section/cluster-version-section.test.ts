import { mount } from "@vue/test-utils";
import type { VueWrapper } from "@vue/test-utils";
import ClusterVersionSection from "./cluster-version-section.vue";
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { Quasar } from "quasar";

describe("cluster-version-section.vue", () => {
  let clusterVersionSectionCmp: VueWrapper<InstanceType<typeof ClusterVersionSection>>;

  beforeEach(() => {
    clusterVersionSectionCmp = mount(ClusterVersionSection, {
      global: {
        plugins: [Quasar],
      },
      props: {
        allowedVersions: ["2.8", "2.9", "2.10", "2.11"],
        modelValue: "",
      },
    });
  });
  afterEach(() => {
    vi.clearAllMocks();
    vi.resetAllMocks();
  });

  test("component is loading properly", () => {
    expect(clusterVersionSectionCmp.html()).toContain("Run:ai version");
    expect(clusterVersionSectionCmp.classes()).toContain("runai-version");
  });

  test("emits an update the update version whena new version is select changes", async () => {
    const versionSelect = clusterVersionSectionCmp.findComponent(".cluster-version-select");
    const version = "2.10";
    await (versionSelect as VueWrapper).vm.$emit("update:model-value", version);

    const expectedEmit: string[][] = [[version]];
    expect(clusterVersionSectionCmp.emitted("update:model-value")).toBeTruthy();
    expect(clusterVersionSectionCmp.emitted("update:model-value")).toEqual(expectedEmit);
  });

  test("displays the selected version in the select", async () => {
    const version = "2.8";
    await clusterVersionSectionCmp.setProps({ modelValue: version });
    const versionSelect = clusterVersionSectionCmp.findComponent(".cluster-version-select");
    expect(versionSelect.text()).toContain(version);
  });

  test('displays "None" in the subheader when modelValue is null', async () => {
    await clusterVersionSectionCmp.find(".runai-expansion-header-content").trigger("click"); // Clicking the section header to close it.
    expect(clusterVersionSectionCmp.text()).toContain("None");
  });

  test("sets sectionInvalid to true when modelValue is empty", () => {
    expect(clusterVersionSectionCmp.vm.sectionInvalid).toBe(true);
  });

  test("sets sectionInvalid to false when modelValue is not empty", async () => {
    await clusterVersionSectionCmp.setProps({ modelValue: "2.8" });
    expect(clusterVersionSectionCmp.vm.sectionInvalid).toBe(false);
  });
});
