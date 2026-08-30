import { describe, it, expect } from "vitest";
import { shallowMount } from "@vue/test-utils";
import AiApplicationChartUploadField from "./ai-application-chart-upload-field.vue";
import { errorMessages } from "@/common/error-message.constant";

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type Vm = any;

const mountComponent = (props: Record<string, unknown> = {}) =>
  shallowMount(AiApplicationChartUploadField, {
    props: {
      selectedFile: null,
      chartId: "",
      uploading: false,
      uploadError: "",
      ...props,
    },
    global: {
      stubs: {
        "q-file": {
          template: "<div><slot name='prepend' /><slot name='append' /></div>",
          props: ["modelValue", "clearable", "rules"],
        },
        "q-icon": true,
      },
    },
  });

function makeFile(name: string, size: number): File {
  const file = new File([new Uint8Array(1)], name);
  Object.defineProperty(file, "size", { value: size });
  return file;
}

describe("ai-application-chart-upload-field", () => {
  it("requires a file when none is selected", () => {
    const vm = mountComponent().vm as Vm;

    expect(vm.chartFileRequiredRule()).toBe(errorMessages.CHART_FILE_REQUIRED);
  });

  it("surfaces the upload error via the required rule when no file is selected", () => {
    const vm = mountComponent({ uploadError: errorMessages.CHART_FILE_TOO_LARGE }).vm as Vm;

    expect(vm.chartFileRequiredRule()).toBe(errorMessages.CHART_FILE_TOO_LARGE);
  });

  it("treats the field as valid while an upload is in progress", () => {
    const vm = mountComponent({ selectedFile: makeFile("chart.tgz", 2), uploading: true }).vm as Vm;

    expect(vm.chartFileUploadedRule()).toBe(true);
  });

  it("reports a failure when a selected file has no chart id and is not uploading", () => {
    const vm = mountComponent({ selectedFile: makeFile("chart.tgz", 2) }).vm as Vm;

    expect(vm.chartFileUploadedRule()).toBe(errorMessages.CHART_UPLOAD_FAILED);
  });

  it("passes validation once a chart id is present", () => {
    const vm = mountComponent({ selectedFile: makeFile("chart.tgz", 2), chartId: "chart-123" }).vm as Vm;

    expect(vm.chartFileUploadedRule()).toBe(true);
  });

  it("shows the wait hint and spinner while uploading", () => {
    const wrapper = mountComponent({ selectedFile: makeFile("chart.tgz", 2), uploading: true });
    const vm = wrapper.vm as Vm;

    expect(vm.hintText).toBe(errorMessages.CHART_UPLOAD_IN_PROGRESS);
    expect(wrapper.find("[aid='chart-uploading-spinner']").exists()).toBe(true);
  });

  it("hides the spinner and wait hint when not uploading", () => {
    const wrapper = mountComponent({ selectedFile: makeFile("chart.tgz", 2), uploading: false });
    const vm = wrapper.vm as Vm;

    expect(vm.hintText).not.toBe(errorMessages.CHART_UPLOAD_IN_PROGRESS);
    expect(wrapper.find("[aid='chart-uploading-spinner']").exists()).toBe(false);
  });

  it("forwards input rejections to the parent", () => {
    const wrapper = mountComponent();
    const vm = wrapper.vm as Vm;
    const rejections = [{ failedPropValidation: "max-file-size", file: makeFile("big.tgz", 11 * 1024 * 1024) }];

    vm.onRejected(rejections);

    expect(wrapper.emitted("rejected")?.[0]?.[0]).toBe(rejections);
  });
});
