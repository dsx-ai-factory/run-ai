import type { EnvironmentVariableOfAsset } from "@/swagger-models/assets-service-client";
import { mount, type VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import { describe } from "vitest";
import { FormRuntimeSettings } from ".";
import type { IRuntimeSettings } from "@/models/environment.model";

describe("form-runtime-settings.vue", () => {
  let formRuntimeSettings: VueWrapper<InstanceType<typeof FormRuntimeSettings>>;

  beforeEach(() => {
    formRuntimeSettings = mount(FormRuntimeSettings, {
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
        environmentVariableSourceValueOptions: [],
      },
    });
  });

  test("should emit 'update' when environment variables are updated", async () => {
    const environmentVariables: Array<EnvironmentVariableOfAsset> = [
      { name: "env1", value: "" },
      { name: "env2", value: "" },
    ];
    formRuntimeSettings.vm.updateAssetEnvironmentVariables(environmentVariables);
    const expectedEmits: Array<Array<IRuntimeSettings>> = [
      [{ assetEnvironmentVariables: environmentVariables, args: null, command: null, workingDir: null }],
    ];
    expect(formRuntimeSettings.emitted("update")).toEqual(expectedEmits);
  });

  test("should emit 'update' when arguments are updated", async () => {
    const args = "--arg1 --arg2";
    formRuntimeSettings.vm.updateCommandAndArgsModel({ args, command: "" });
    const expectedEmits: Array<Array<IRuntimeSettings>> = [
      [{ assetEnvironmentVariables: [], args, command: null, workingDir: null }],
    ];
    expect(formRuntimeSettings.emitted("update")).toEqual(expectedEmits);
  });

  test("should emit 'update' when command updated", async () => {
    const command = "command";
    formRuntimeSettings.vm.updateCommandAndArgsModel({ args: "", command });
    const expectedEmits: Array<Array<IRuntimeSettings>> = [
      [{ assetEnvironmentVariables: [], args: null, command, workingDir: null }],
    ];
    expect(formRuntimeSettings.emitted("update")).toEqual(expectedEmits);
  });

  test("should emit 'update' when workingDir updated", async () => {
    const workingDir = "/path/to/working/dir";
    formRuntimeSettings.vm.updateWorkingDir(workingDir);
    const expectedEmits: Array<Array<IRuntimeSettings>> = [
      [{ assetEnvironmentVariables: [], args: null, command: null, workingDir }],
    ];
    expect(formRuntimeSettings.emitted("update")).toEqual(expectedEmits);
  });

  test("should not emit 'update' when nothing changed", async () => {
    expect(formRuntimeSettings.emitted("update")).toBeFalsy();
  });
});
