import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { FormImageSection } from ".";
import { beforeEach, describe, expect, test, vi } from "vitest";
import { Quasar } from "quasar";

vi.mock("@/services/control-plane/catalog.service/catalog.service", () => ({
  catalogService: {
    getRegistries: vi.fn().mockResolvedValue([]),
  },
}));

import { ImagePullPolicy } from "@/swagger-models/assets-service-client";
import type { IImage } from "@/models/environment.model";
import { createTestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";

const pinia = createTestingPinia({ createSpy: vi.fn });

setActivePinia(pinia);

describe("form-image-section.vue", () => {
  let formImageSectionCmp: VueWrapper<InstanceType<typeof FormImageSection>>;

  beforeEach(() => {
    formImageSectionCmp = mount(FormImageSection, {
      global: {
        plugins: [Quasar],
      },
      props: {
        imageModel: {
          image: "",
          imagePullPolicy: ImagePullPolicy.IfNotPresent,
        },
      },
    });
    formImageSectionCmp.vm.$q.notify = vi.fn();
  });

  test("component is loading properly", () => {
    expect(formImageSectionCmp.html()).toContain("Image");
  });

  test("section should be invalid if image is empty", () => {
    expect(formImageSectionCmp.vm.sectionValid).toBe(false);
  });

  test("section should be valid if image is filled", () => {
    formImageSectionCmp.vm.imageModel.image = "ubuntu";
    expect(formImageSectionCmp.vm.sectionValid).toBe(true);
  });

  test("section should emit update when image changed", () => {
    formImageSectionCmp.find("input[type=text]").setValue("ubuntu");
    const expectedEmits: Array<Array<IImage>> = [[{ image: "ubuntu", imagePullPolicy: ImagePullPolicy.IfNotPresent }]];
    expect(formImageSectionCmp.emitted("update")).toEqual(expectedEmits);
  });

  test("section should emit update when pull policy changed", () => {
    formImageSectionCmp.findAll("input[type=radio]")[1].trigger("click");
    expect(formImageSectionCmp.emitted("update")).toEqual([[{ image: "", imagePullPolicy: ImagePullPolicy.Always }]]);
  });

  test("section summary should be 'None' when image wasn't filled", () => {
    expect(formImageSectionCmp.vm.summary).toBe("None");
  });

  test("section summary should be 'ubuntu' when image was filled with 'ubuntu'", () => {
    const imageValue = "ubuntu";
    formImageSectionCmp.vm.imageModel.image = imageValue;
    expect(formImageSectionCmp.vm.summary).toBe(imageValue);
  });
});
