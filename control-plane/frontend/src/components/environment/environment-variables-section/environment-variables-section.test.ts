import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { EnvironmentVariablesSection } from ".";
import EnvironmentVariableBox from "./environment-variable-box/environment-variable-box.vue";

describe("environment-variables-section.vue", () => {
  function factory(props = {}): VueWrapper<InstanceType<typeof EnvironmentVariablesSection>> {
    return mount(EnvironmentVariablesSection, {
      props: {
        environmentVariables: [],
        sourceValueOptions: [],
        ...props,
      },
      global: {
        plugins: [Quasar],
        components: { EnvironmentVariableBox },
      },
    });
  }

  describe("Html elements check", () => {
    it("Component exists", () => {
      const wrapper = factory();
      expect(wrapper.exists()).toBe(true);
    });

    it("Check that the add environment variable button exists", () => {
      const wrapper = factory();
      expect(wrapper.find("[aid=add-environment-variable-button]").exists()).toBe(true);
    });
  });

  describe("Props", () => {
    describe("environmentVariables", () => {
      it("should not render environment variable box when environmentVariables is empty", () => {
        const wrapper = factory();
        expect(wrapper.findAllComponents(EnvironmentVariableBox).length).toBe(0);
      });

      it("should render 2 environment variable box when environmentVariables contains 2", async () => {
        const wrapper = factory({
          environmentVariables: [
            { name: "name", value: "value" },
            { name: "name2", value: "value2" },
          ],
        });
        expect(wrapper.findAllComponents(EnvironmentVariableBox).length).toBe(2);
      });
    });
  });

  describe("Emits", () => {
    it("should emit update-environment-variables when add environment variable button is clicked", async () => {
      const wrapper = factory();
      await wrapper.find("[aid=add-environment-variable-button]").trigger("click");
      const expectedEmit = [{ name: "", value: "" }];
      expect(wrapper.emitted("update-environment-variables")).toBeTruthy();
      expect(wrapper.emitted("update-environment-variables")?.[0][0]).toEqual(expectedEmit);
    });
  });
});
