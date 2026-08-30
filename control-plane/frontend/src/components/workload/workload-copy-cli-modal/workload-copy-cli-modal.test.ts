import { describe, test, expect, vi, beforeEach, afterEach } from "vitest";
import { createTestingPinia, type TestingPinia } from "@pinia/testing";
import { Quasar } from "quasar";
import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { WorkloadCopyCliModal } from "@/components/workload/workload-copy-cli-modal/";
import { workload1 } from "@/components/workload/workload-copy-cli-modal/workload-copy-cli-modal.mock";
import { unitTestUtil } from "@/utils/unit-test.util";

const pinia: TestingPinia = createTestingPinia({ createSpy: vi.fn });
vi.mock("@/utils/alert.util");
describe("workload-copy-cli-modal component", () => {
  let wrapper: VueWrapper<InstanceType<typeof WorkloadCopyCliModal>>;

  const createComponent = (props: { cliCommand: string; workloadName: string }) => {
    wrapper = mount(WorkloadCopyCliModal, {
      global: {
        plugins: [Quasar, pinia],
        directives: {
          permission: vi.fn(),
        },
      },
      props,
    });
    wrapper.vm.$q.notify = vi.fn(); // disabling notify to prevent it from failing tests
  };

  beforeEach(() => {
    createComponent({ workloadName: workload1.name, cliCommand: workload1.additionalFields?.cliCommand });
  });

  afterEach(() => {
    wrapper.unmount();
  });

  test("renders workload copy CLI modal correctly", () => {
    const modalEl = unitTestUtil.getModalElement();
    expect(wrapper.exists()).toBeTruthy();
    expect(modalEl.text()).toContain("Make a Copy of");
    expect(modalEl.find(".copy-command-box").exists()).toBeTruthy();
    expect(modalEl.find(".copy-command-box pre").text()).toContain("runai submit-dist pytorch --name myworkload");
  });

  test("emits 'cancel' event when canceled", async () => {
    const modalEl = unitTestUtil.getModalElement();
    await modalEl.find("[aid=close-workload-copy-cli-modal]").trigger("click");
    expect(wrapper.emitted("cancel")).toBeTruthy();
  });
});
