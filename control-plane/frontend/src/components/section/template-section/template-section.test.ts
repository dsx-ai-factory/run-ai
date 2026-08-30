import { mount } from "@vue/test-utils";
import TemplateSection from "./template-section.vue";
import { beforeEach, describe, expect, test } from "vitest";
import { templatesMockData } from "./template-section.mock";
import type { VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import { errorMessages } from "@/common/error-message.constant";
import type { WorkloadTemplate } from "@/swagger-models/assets-service-client";
import type { TTestEmit } from "@/models/unit-test.model";
import { createTestingPinia, type TestingPinia } from "@pinia/testing";

const pinia: TestingPinia = createTestingPinia({ createSpy: vi.fn });
describe("template-section.vue", () => {
  let templateSectionCmp: VueWrapper<InstanceType<typeof TemplateSection>>;
  const props = {
    loading: false as boolean,
    templates: [] as WorkloadTemplate[],
    templateId: null as string | null,
  };

  beforeEach(() => {
    templateSectionCmp = mount(TemplateSection, {
      global: {
        plugins: [Quasar, pinia],
        directives: {
          clickOutside: vi.fn(),
        },
      },
      props,
    });
  });

  test("component is loading properly", () => {
    expect(templateSectionCmp.html()).toContain("Template");
    expect(templateSectionCmp.text()).toContain("Select a preconfigured template to launch a new workload quickly");
  });

  test("the component will render empty text when the templates array prop is empty and vice versa", async () => {
    // note: that the templates array prop is [] by default
    expect(templateSectionCmp.find(".empty-data-display-paragraph").exists()).toBe(true);
    await templateSectionCmp.setProps({ templates: templatesMockData });
    expect(templateSectionCmp.find(".empty-data-display-paragraph").exists()).toBe(false);
  });

  test("isRequiredTemplate method will return true when templateId is truthy and will return the proper text when templateId is falsey ", async () => {
    // note: this is a method specific test
    expect(templateSectionCmp.vm.isRequiredTemplate("")).toBe(errorMessages.SELECT_TEMPLATE);
    expect(templateSectionCmp.vm.isRequiredTemplate("1")).toBe(true);
  });

  test("the component will emit the autoselected 'Start from scartch' id when its loaded with an empty array of templates", async () => {
    await templateSectionCmp.find(`[aid=create-new-card]`).trigger("click");
    const expectedEmit: string[][] = [["-1"]];
    expect(templateSectionCmp.emitted("template-changed")).toEqual(expectedEmit);
  });

  test("the component will emit the updated selected templateId correctly when a new selection is made", async () => {
    await templateSectionCmp.setProps({ templates: templatesMockData });

    // select a card
    await templateSectionCmp.find(`[aid=create-new-card]`).trigger("click");
    const expectedFirstEmit: Array<string> = ["-1"];
    let emitResult: TTestEmit<string[]> = templateSectionCmp.emitted("template-changed") || [];
    expect(emitResult[0]).toEqual(expectedFirstEmit);

    // select a different card
    await templateSectionCmp.find(`[aid=${templatesMockData[0].meta.name}-card]`).trigger("click");
    const expectedSecondEmit: Array<string> = [templatesMockData[0].meta.id];
    emitResult = templateSectionCmp.emitted("template-changed") || [];
    expect(emitResult[1]).toEqual(expectedSecondEmit);
  });

  test("the component will emit 'null' when the selected template is clicked again", async () => {
    const selectedTemplate = templatesMockData[0];
    await templateSectionCmp.setProps({
      templateId: selectedTemplate.meta.id,
      templates: templatesMockData,
    });

    // clicking the selected template card (notice the selector,first card is 'Start from scratch'):
    await templateSectionCmp.find(`[aid=${selectedTemplate.meta.name}-card]`).trigger("click");
    const expectedEmit: Array<null> = [null];
    const emitResult: TTestEmit<string[]> = templateSectionCmp.emitted("template-changed") || [];
    expect(emitResult[0]).toEqual(expectedEmit);
  });
});
