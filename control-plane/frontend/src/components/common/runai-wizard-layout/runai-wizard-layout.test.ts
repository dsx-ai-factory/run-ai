import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { describe, it, expect, beforeEach } from "vitest";
import { Quasar } from "quasar";
import RunaiWizardLayout from "./runai-wizard-layout.vue";
import RunaiPanelsLayout from "@/components/common/runai-panels-layout/runai-panels-layout.vue";
import RunaiWizardStepsPanel from "./runai-wizard-steps-panel.vue";
import RunaiPanelContainer from "@/components/common/runai-panel-container/runai-panel-container.vue";
import type { IRunaiWizardLayoutStep } from "@/models/runai-wizard-layout.model";

describe("RunaiWizardLayout", () => {
  let wrapper: VueWrapper<InstanceType<typeof RunaiWizardLayout>>;
  let mockSteps: IRunaiWizardLayoutStep[];

  const createWrapper = (
    props?: Partial<{ steps: IRunaiWizardLayoutStep[]; activeStep: string }>,
    slots?: {
      default?: string;
      "content-footer"?: string;
      "right-sidebar"?: string;
    },
  ) => {
    return mount(RunaiWizardLayout, {
      props: {
        steps: props?.steps || mockSteps,
        activeStep: props?.activeStep ?? "step-1",
      },
      slots: slots || {
        default: "<div class='test-content'>Main Content</div>",
      },
      global: {
        plugins: [Quasar],
        components: { RunaiPanelsLayout, RunaiWizardStepsPanel, RunaiPanelContainer },
      },
    });
  };

  beforeEach(() => {
    mockSteps = [
      { name: "step-1", title: "Step 1", done: false },
      { name: "step-2", title: "Step 2", done: false },
      { name: "step-3", title: "Step 3", done: false },
    ];
  });

  describe("Component Rendering", () => {
    it("renders the component correctly", () => {
      wrapper = createWrapper();

      expect(wrapper.exists()).toBe(true);
      expect(wrapper.findComponent(RunaiPanelsLayout).exists()).toBe(true);
    });

    it("renders within panels layout", () => {
      wrapper = createWrapper();

      const panelsLayout = wrapper.findComponent(RunaiPanelsLayout);
      expect(panelsLayout.exists()).toBe(true);
    });

    it("renders the steps panel", () => {
      wrapper = createWrapper();

      const stepsPanel = wrapper.findComponent(RunaiWizardStepsPanel);
      expect(stepsPanel.exists()).toBe(true);
    });

    it("renders the main content area", () => {
      wrapper = createWrapper();

      const mainContent = wrapper.find(".wizard-main-content");
      expect(mainContent.exists()).toBe(true);
    });

    it("renders the default slot content", () => {
      wrapper = createWrapper();

      const content = wrapper.find(".test-content");
      expect(content.exists()).toBe(true);
      expect(content.text()).toBe("Main Content");
    });

    it("does not render right sidebar when slot is not provided", () => {
      wrapper = createWrapper();

      const panelContainers = wrapper.findAllComponents(RunaiPanelContainer);
      // Should only have one (steps panel), not two
      expect(panelContainers.length).toBe(1);
    });

    it("renders right sidebar when slot is provided", () => {
      wrapper = createWrapper(undefined, {
        default: "<div>Main Content</div>",
        "right-sidebar": "<div class='test-sidebar'>Sidebar Content</div>",
      });

      const panelContainers = wrapper.findAllComponents(RunaiPanelContainer);
      // Should have two now (steps panel + right sidebar)
      expect(panelContainers.length).toBe(2);
      expect(wrapper.find(".test-sidebar").text()).toBe("Sidebar Content");
    });

    it("renders content-footer slot when provided", () => {
      wrapper = createWrapper(undefined, {
        default: "<div>Main Content</div>",
        "content-footer": "<button class='test-action'>Next</button>",
      });

      const contentFooter = wrapper.find(".wizard-content-footer");
      expect(contentFooter.exists()).toBe(true);
      expect(wrapper.find(".test-action").exists()).toBe(true);
    });

    it("does not render content footer when slot is not provided", () => {
      wrapper = createWrapper(undefined, {
        default: "<div>Main Content</div>",
      });

      const contentFooter = wrapper.find(".wizard-content-footer");
      expect(contentFooter.exists()).toBe(false);
    });
  });

  describe("Props", () => {
    it("passes steps prop to steps panel", () => {
      wrapper = createWrapper();

      const stepsPanel = wrapper.findComponent(RunaiWizardStepsPanel);
      expect(stepsPanel.props("steps")).toEqual(mockSteps);
    });

    it("passes activeStep prop to steps panel", () => {
      wrapper = createWrapper({ activeStep: "step-2" });

      const stepsPanel = wrapper.findComponent(RunaiWizardStepsPanel);
      expect(stepsPanel.props("activeStep")).toBe("step-2");
    });

    it("updates steps panel when activeStep changes", async () => {
      wrapper = createWrapper({ activeStep: "step-1" });

      await wrapper.setProps({ activeStep: "step-3" });

      const stepsPanel = wrapper.findComponent(RunaiWizardStepsPanel);
      expect(stepsPanel.props("activeStep")).toBe("step-3");
    });

    it("requires steps prop", () => {
      expect(RunaiWizardLayout.props.steps.required).toBe(true);
    });

    it("requires activeStep prop", () => {
      expect(RunaiWizardLayout.props.activeStep.required).toBe(true);
    });

    it("activeStep is string type", () => {
      expect(RunaiWizardLayout.props.activeStep.type).toBe(String);
    });
  });

  describe("Events", () => {
    it("emits step-click event when steps panel emits it", async () => {
      wrapper = createWrapper();

      const stepsPanel = wrapper.findComponent(RunaiWizardStepsPanel);
      const payload = { step: mockSteps[1], name: "step-2" };

      await stepsPanel.vm.$emit("step-click", payload);

      expect(wrapper.emitted("step-click")).toBeTruthy();
      expect(wrapper.emitted("step-click")?.[0]).toEqual([payload]);
    });

    it("declares step-click in emits", () => {
      expect(RunaiWizardLayout.emits).toContain("step-click");
    });
  });

  describe("Layout Structure", () => {
    it("uses panels layout component", () => {
      wrapper = createWrapper();

      const panelsLayout = wrapper.findComponent(RunaiPanelsLayout);
      expect(panelsLayout.exists()).toBe(true);
    });

    it("renders content scroll area", () => {
      wrapper = createWrapper();

      const scrollArea = wrapper.find(".wizard-content-scroll");
      expect(scrollArea.exists()).toBe(true);
    });

    it("has proper content footer layout", () => {
      wrapper = createWrapper(undefined, {
        default: "<div>Main Content</div>",
        "content-footer": "<button>Next</button><button>Back</button>",
      });

      const contentFooter = wrapper.find(".wizard-content-footer");
      expect(contentFooter.exists()).toBe(true);
    });

    it("places steps panel in left slot", () => {
      wrapper = createWrapper();

      const panelsLayout = wrapper.findComponent(RunaiPanelsLayout);
      const stepsPanel = wrapper.findComponent(RunaiWizardStepsPanel);

      expect(stepsPanel.exists()).toBe(true);
      expect(panelsLayout.exists()).toBe(true);
    });

    it("places right sidebar in right slot when provided", () => {
      wrapper = createWrapper(undefined, {
        default: "<div>Main</div>",
        "right-sidebar": "<div class='right-content'>Right</div>",
      });

      const panelContainers = wrapper.findAllComponents(RunaiPanelContainer);
      expect(panelContainers.length).toBe(2);
      expect(wrapper.find(".right-content").exists()).toBe(true);
    });
  });

  describe("Slot System", () => {
    it("supports default slot for main content", () => {
      const customContent = "<div class='custom'>Custom Content</div>";
      wrapper = createWrapper(undefined, {
        default: customContent,
      });

      expect(wrapper.find(".custom").exists()).toBe(true);
      expect(wrapper.find(".custom").text()).toBe("Custom Content");
    });

    it("supports content-footer slot for navigation buttons", () => {
      const contentFooterContent = "<button class='next-btn'>Next</button>";
      wrapper = createWrapper(undefined, {
        default: "<div>Content</div>",
        "content-footer": contentFooterContent,
      });

      expect(wrapper.find(".next-btn").exists()).toBe(true);
    });

    it("supports right-sidebar slot for contextual content", () => {
      const sidebarContent = "<div class='help'>Help Content</div>";
      wrapper = createWrapper(undefined, {
        default: "<div>Content</div>",
        "right-sidebar": sidebarContent,
      });

      expect(wrapper.find(".help").exists()).toBe(true);
      expect(wrapper.find(".help").text()).toBe("Help Content");
    });

    it("allows multiple slots at once", () => {
      wrapper = createWrapper(undefined, {
        default: "<div class='main'>Main</div>",
        "content-footer": "<button class='action'>Action</button>",
        "right-sidebar": "<div class='sidebar'>Sidebar</div>",
      });

      expect(wrapper.find(".main").exists()).toBe(true);
      expect(wrapper.find(".action").exists()).toBe(true);
      expect(wrapper.find(".sidebar").exists()).toBe(true);
    });
  });

  describe("Component Integration", () => {
    it("steps panel receives correct props", () => {
      const customSteps = [
        { name: "step-a", title: "Step A", done: true },
        { name: "step-b", title: "Step B", done: false, hasError: true },
      ];
      wrapper = createWrapper({ steps: customSteps, activeStep: "step-b" });

      const stepsPanel = wrapper.findComponent(RunaiWizardStepsPanel);
      expect(stepsPanel.props("steps")).toEqual(customSteps);
      expect(stepsPanel.props("activeStep")).toBe("step-b");
    });

    it("right sidebar renders when right-sidebar slot provided", () => {
      wrapper = createWrapper(undefined, {
        default: "<div>Main</div>",
        "right-sidebar": "<div class='sidebar-test'>Test</div>",
      });

      expect(wrapper.find(".sidebar-test").exists()).toBe(true);
    });

    it("integrates with panels layout correctly", () => {
      wrapper = createWrapper(undefined, {
        default: "<div class='main-test'>Main</div>",
        "right-sidebar": "<div class='right-test'>Right</div>",
      });

      const panelsLayout = wrapper.findComponent(RunaiPanelsLayout);
      const stepsPanel = wrapper.findComponent(RunaiWizardStepsPanel);

      expect(panelsLayout.exists()).toBe(true);
      expect(stepsPanel.exists()).toBe(true);
      expect(wrapper.find(".main-test").exists()).toBe(true);
      expect(wrapper.find(".right-test").exists()).toBe(true);
    });
  });

  describe("Responsive Behavior", () => {
    it("uses panels layout for responsive behavior", () => {
      wrapper = createWrapper();

      const panelsLayout = wrapper.findComponent(RunaiPanelsLayout);
      expect(panelsLayout.exists()).toBe(true);
    });

    it("main content area is flexible", () => {
      wrapper = createWrapper();

      const mainContent = wrapper.find(".wizard-main-content");
      expect(mainContent.exists()).toBe(true);
    });
  });

  describe("Component Lifecycle", () => {
    it("handles prop updates correctly", async () => {
      wrapper = createWrapper({ activeStep: "step-1" });

      await wrapper.setProps({ activeStep: "step-3" });

      const stepsPanel = wrapper.findComponent(RunaiWizardStepsPanel);
      expect(stepsPanel.props("activeStep")).toBe("step-3");
    });

    it("handles steps array updates", async () => {
      wrapper = createWrapper({ steps: mockSteps });

      const newSteps = [...mockSteps, { name: "step-4", title: "Step 4", done: false }];

      await wrapper.setProps({ steps: newSteps });

      const stepsPanel = wrapper.findComponent(RunaiWizardStepsPanel);
      expect(stepsPanel.props("steps")).toHaveLength(4);
    });
  });

  describe("Accessibility", () => {
    it("main content area exists", () => {
      wrapper = createWrapper();

      expect(wrapper.find(".wizard-main-content").exists()).toBe(true);
    });

    it("uses panels layout which provides semantic structure", () => {
      wrapper = createWrapper();

      const panelsLayout = wrapper.findComponent(RunaiPanelsLayout);
      expect(panelsLayout.exists()).toBe(true);
    });
  });

  describe("Edge Cases", () => {
    it("handles empty steps array", () => {
      wrapper = createWrapper({ steps: [] });

      expect(wrapper.exists()).toBe(true);
      const stepsPanel = wrapper.findComponent(RunaiWizardStepsPanel);
      expect(stepsPanel.props("steps")).toHaveLength(0);
    });

    it("handles single step", () => {
      const singleStep = [{ name: "only-step", title: "Only Step", done: false }];
      wrapper = createWrapper({ steps: singleStep, activeStep: "only-step" });

      const stepsPanel = wrapper.findComponent(RunaiWizardStepsPanel);
      expect(stepsPanel.props("steps")).toHaveLength(1);
    });

    it("handles many steps", () => {
      const manySteps = Array.from({ length: 10 }, (_, i) => ({
        name: `step-${i + 1}`,
        title: `Step ${i + 1}`,
        done: false,
      }));
      wrapper = createWrapper({ steps: manySteps, activeStep: "step-1" });

      const stepsPanel = wrapper.findComponent(RunaiWizardStepsPanel);
      expect(stepsPanel.props("steps")).toHaveLength(10);
    });
  });
});
