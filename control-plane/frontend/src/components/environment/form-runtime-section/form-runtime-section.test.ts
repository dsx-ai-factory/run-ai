import type { EnvironmentVariableOfAsset } from "@/swagger-models/assets-service-client";
import { mount, type VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import { describe } from "vitest";
import { FormRuntimeSection } from ".";

describe("form-runtime-section.vue", () => {
  let formRuntimeSection: VueWrapper<InstanceType<typeof FormRuntimeSection>>;

  beforeEach(() => {
    formRuntimeSection = mount(FormRuntimeSection, {
      global: {
        plugins: [Quasar],
      },
      props: {
        runtimeSettings: {
          command: null as string | null,
          args: null as string | null,
          workingDir: null as string | null,
          assetEnvironmentVariables: [] as Array<EnvironmentVariableOfAsset>,
        },
      },
    });
  });

  test("component is loading properly", () => {
    expect(formRuntimeSection.html()).toContain("Runtime settings");
  });

  test("summary should be 'None' when no field are filled", () => {
    expect(formRuntimeSection.vm.summary).toBe("None");
  });

  test("summary should be 'Command / 1 Argument(s)' when provided command and 1 argument", () => {
    if (!formRuntimeSection.vm.runtimeSettings) return;
    formRuntimeSection.vm.runtimeSettings.command = "command";
    formRuntimeSection.vm.runtimeSettings.args = "--arg1";
    expect(formRuntimeSection.vm.summary).toBe("Command / 1 Argument(s)");
  });

  test("summary should count arguments seperated by space and have 2 args", () => {
    if (!formRuntimeSection.vm.runtimeSettings) return;
    formRuntimeSection.vm.runtimeSettings.args = "--arg1 --arg2";
    expect(formRuntimeSection.vm.summary).toContain("2 Argument(s)");
  });

  test("summary should count environment variables and display an appropriate number in the summary", () => {
    if (!formRuntimeSection.vm.runtimeSettings) return;
    formRuntimeSection.vm.runtimeSettings.assetEnvironmentVariables = [{}, {}, {}];
    expect(formRuntimeSection.vm.summary).toContain("3 Environment Variable(s)");
  });

  test("section should be invalid when there are environment variables with no name", () => {
    if (!formRuntimeSection.vm.runtimeSettings) return;
    formRuntimeSection.vm.runtimeSettings.assetEnvironmentVariables = [{ name: "" }, { name: "" }];
    expect(formRuntimeSection.vm.sectionValid).toBe(false);
  });

  test("section should be invalid when the workingDir is invalid", () => {
    if (!formRuntimeSection.vm.runtimeSettings) return;
    formRuntimeSection.vm.runtimeSettings.workingDir = "invalid";
    expect(formRuntimeSection.vm.sectionValid).toBe(false);
  });

  test("section should be valid when there are environment variables with name filled", () => {
    if (!formRuntimeSection.vm.runtimeSettings) return;
    formRuntimeSection.vm.runtimeSettings.assetEnvironmentVariables = [{ name: "env1" }, { name: "env2" }];
    expect(formRuntimeSection.vm.sectionValid).toBe(true);
  });

  test("should not emit 'update' when nothing changed", async () => {
    expect(formRuntimeSection.emitted("update")).toBeFalsy();
  });
});
