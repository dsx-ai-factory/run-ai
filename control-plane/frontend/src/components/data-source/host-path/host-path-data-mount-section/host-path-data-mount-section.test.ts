import { createTestingPinia } from "@pinia/testing";
import { mount, type VueWrapper } from "@vue/test-utils";
import { setActivePinia } from "pinia";
import { Quasar } from "quasar";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { HostPathDataMountSection } from ".";

const pinia = createTestingPinia({ createSpy: vi.fn });
setActivePinia(pinia);

describe("HostPathDataMountSection", () => {
  let dataMountSection: VueWrapper<InstanceType<typeof HostPathDataMountSection>>;
  beforeEach(() => {
    dataMountSection = mount(HostPathDataMountSection, {
      global: {
        plugins: [Quasar, pinia],
      },
      props: {
        modelValue: {
          hostPath: "",
          mountPath: "",
          readOnly: false,
        },
      },
    });
  });

  it("should render component with 2 text inputs", () => {
    expect(dataMountSection.findAll("input[type=text]").length).toBe(2);
  });

  it("should render component with 1 checkbox input", () => {
    expect(dataMountSection.find("input[type=checkbox]").exists()).toBe(true);
  });
});
