import type { TTestEmit } from "@/models/unit-test.model";
import type { IUIWorkloadEnvSectionModel } from "./environment-section.model";
import {
  InternalConnectionType,
  type EnvironmentAsset,
  type SpecificRunConnectionInfo,
  type Connection,
} from "@/swagger-models/assets-service-client";
import { createTestingPinia } from "@pinia/testing";
import { mount, type VueWrapper } from "@vue/test-utils";
import { setActivePinia } from "pinia";
import { Quasar } from "quasar";
import { beforeEach, afterEach, describe, expect, test, vi } from "vitest";
import { EnvironmentSection } from ".";
import {
  customToolsEnvTest,
  environments,
  environmentUidGidSourceTest,
  environmentVariablesEnvTest,
} from "./environment-section.mock";
import { EWorkloadFormType } from "@/models/workload.model";
import { useAuthStore } from "@/stores/auth.store";
const pinia = createTestingPinia({ createSpy: vi.fn });
setActivePinia(pinia);

describe("environment-section.vue", () => {
  let environmentSection: VueWrapper<InstanceType<typeof EnvironmentSection>>;
  const specificEnv: IUIWorkloadEnvSectionModel = {
    connections: [],
    command: "",
    args: "",
    runAsUid: null,
    runAsGid: null,
    environmentVariables: [],
  };

  beforeEach(() => {
    environmentSection = mount(EnvironmentSection, {
      global: {
        plugins: [Quasar, pinia],
        directives: {
          clickOutside: vi.fn(),
        },
      },
      props: {
        entityType: EWorkloadFormType.Workspace,
        specificEnv,
        environmentId: "",
        loading: false,
        environments,
        clusterId: "123",
        credentials: [],
      },
      provide: {
        policy: {},
      },
      data() {
        return {
          authStore: useAuthStore(pinia),
        };
      },
    });
  });
  afterEach(() => {
    vi.clearAllMocks();
    vi.resetAllMocks();
  });

  describe("Default rendering behavior", () => {
    test("component is loading properly", () => {
      expect(environmentSection.html()).toContain("Environment");
      expect(environmentSection.html()).toContain("new environment");
    });
  });

  describe("uid, gid and supplemental groups inputs behavior", () => {
    describe("'uidGidSource' = 'fromIdPtoken'", () => {
      let uidInput: HTMLInputElement;
      let gidInput: HTMLInputElement;
      let supplementalGroupsInput: HTMLInputElement;

      beforeEach(async () => {
        const selectedEnvironment = environmentUidGidSourceTest;
        await environmentSection.setProps({
          environmentId: selectedEnvironment.meta.id,
          environments: [selectedEnvironment],
        });
        uidInput = environmentSection.find("[aid=uid-input]").element as HTMLInputElement;
        gidInput = environmentSection.find("[aid=gid-input]").element as HTMLInputElement;
        supplementalGroupsInput = environmentSection.find("[aid=supplemental-groups-input]").element as HTMLInputElement;
      });

      test("the component will render the uid-gid-section", async () => {
        expect(environmentSection.find("[aid=uid-gid-section]").exists()).toBe(true);
      });

      test("the component will render the inputs disabled when the fields exists in the token", async () => {
        await environmentSection.setData({
          tokenUid: 1000,
          tokenGid: 2000,
          tokenSupplementalGroups: [3000, 4000],
        });
        expect(uidInput.disabled).toBe(true);
        expect(gidInput.disabled).toBe(true);
        expect(supplementalGroupsInput.disabled).toBe(true);
      });

      test("the component will render the uid,gid inputs enabled and supplemental disabled when the fields do not exists in the token", async () => {
        await environmentSection.setData({
          tokenUid: null,
          tokenGid: null,
          tokenSupplementalGroups: null,
        });

        expect(uidInput.disabled).toBe(false);
        expect(gidInput.disabled).toBe(false);
        expect(supplementalGroupsInput.disabled).toBe(true);
      });

      test("the component will render the supplemental input enabled when the gid field exists in the token", async () => {
        await environmentSection.setData({
          tokenUid: null,
          tokenGid: 322,
          tokenSupplementalGroups: null,
        });

        await environmentSection.setProps({
          specificEnv: {
            runAsGid: 322,
            args: "",
            environmentVariables: [],
            connections: [],
          },
        });

        expect(gidInput.disabled).toBe(true);
        expect(supplementalGroupsInput.disabled).toBe(false);
      });
    });

    describe("'uidGidSource' = 'custom'", () => {
      let selectedEnvironment: EnvironmentAsset;
      beforeEach(() => {
        selectedEnvironment = environmentUidGidSourceTest;
        selectedEnvironment.spec.uidGidSource = "custom";
        selectedEnvironment.spec.overrideUidGidInWorkspace = true;
      });

      test("the component will render the uid-gid-section when override is enabled", async () => {
        await environmentSection.setProps({
          environmentId: selectedEnvironment.meta.id,
          environments: [selectedEnvironment],
        });
        expect(environmentSection.find("[aid=uid-gid-section]").exists()).toBe(true);
      });

      test("the component will not render the uid-gid-section when override is disabled", async () => {
        selectedEnvironment.spec.overrideUidGidInWorkspace = false;
        await environmentSection.setProps({
          environmentId: selectedEnvironment.meta.id,
          environments: [selectedEnvironment],
        });
        expect(environmentSection.find("[aid=uid-gid-section]").exists()).toBe(false);
      });

      test("the inputs will have the values accroding to the specificEnv prop", async () => {
        await environmentSection.setProps({
          environmentId: selectedEnvironment.meta.id,
          environments: [selectedEnvironment],
          specificEnv: {
            args: "",
            environmentVariables: [],
            connections: [],
            runAsUid: 1,
            runAsGid: 2,
            supplementalGroups: "3;4",
          },
        });
        const uidInput: HTMLInputElement = environmentSection.find("[aid=uid-input]").element as HTMLInputElement;
        const gidInput: HTMLInputElement = environmentSection.find("[aid=gid-input]").element as HTMLInputElement;
        const supplementalGroupsInput: HTMLInputElement = environmentSection.find("[aid=supplemental-groups-input]")
          .element as HTMLInputElement;

        expect(+uidInput.value).toBe(environmentSection.vm.specificEnv?.runAsUid);
        expect(+gidInput.value).toBe(environmentSection.vm.specificEnv?.runAsGid);
        expect(supplementalGroupsInput.value).toBe(environmentSection.vm.specificEnv?.supplementalGroups);
      });

      test("the supplemental groups input will be disabled if no gid is provided", async () => {
        await environmentSection.setProps({
          environmentId: selectedEnvironment.meta.id,
          environments: [selectedEnvironment],
        });

        const supplementalGroupsInput: HTMLInputElement = environmentSection.find("[aid=supplemental-groups-input]")
          .element as HTMLInputElement;
        expect(supplementalGroupsInput.disabled).toBe(true);
      });

      test("the supplemental groups input will be enabled if gid is provided", async () => {
        await environmentSection.setProps({
          environmentId: selectedEnvironment.meta.id,
          environments: [selectedEnvironment],
          specificEnv: {
            runAsGid: 21,
            args: "",
            environmentVariables: [],
            connections: [],
          },
        });

        const supplementalGroupsInput: HTMLInputElement = environmentSection.find("[aid=supplemental-groups-input]")
          .element as HTMLInputElement;
        expect(supplementalGroupsInput.disabled).toBe(false);
      });
    });

    describe("'uidGidSource' = 'fromTheImage'", () => {
      test("the component will not render the uid-gid-section", async () => {
        const selectedEnvironment = environmentUidGidSourceTest;
        selectedEnvironment.spec.uidGidSource = "fromTheImage";
        await environmentSection.setProps({
          environmentId: selectedEnvironment.meta.id,
          environments: [selectedEnvironment],
        });
        expect(environmentSection.find("[aid=uid-gid-section]").exists()).toBe(false);
      });
    });
  });

  describe("Environment selection behavior", () => {
    test("should emit 'environment-changed' with the environmentId that was clicked and the environment specifcEnv", async () => {
      const selectedEnvironment: EnvironmentAsset = environments[0];
      await environmentSection.find(`[aid=${selectedEnvironment.meta.name}-card]`).trigger("click");
      const emittedEnvironmentId: string = (
        environmentSection.emitted("environment-changed") as TTestEmit<{ environmentId: string }>
      )[0][0].environmentId;
      expect(emittedEnvironmentId).toBe(selectedEnvironment.meta.id);
    });

    test("should emit 'environment-changed' with the environmentId valued null when clicking the selected environment card", async () => {
      const selectedEnvironment: EnvironmentAsset = environments[0];
      await environmentSection.setProps({ environmentId: selectedEnvironment.meta.id });
      await environmentSection.find(`[aid=${selectedEnvironment.meta.name}-card]`).trigger("click");
      const emittedEnvironmentId: string = (
        environmentSection.emitted("environment-changed") as TTestEmit<{ environmentId: string }>
      )[0][0].environmentId;
      expect(emittedEnvironmentId).toBe(null);
    });

    test("should emit 'environment-changed' with the new environmentId when clicking another environment card", async () => {
      const selectedEnvironment: EnvironmentAsset = environments[0];
      const newSelectedEnvironment: EnvironmentAsset = environments[1];
      await environmentSection.setProps({ environmentId: selectedEnvironment.meta.id });
      await environmentSection.find(`[aid=${newSelectedEnvironment.meta.name}-card]`).trigger("click");
      const emittedEnvironmentId: string = (
        environmentSection.emitted("environment-changed") as TTestEmit<{ environmentId: string }>
      )[0][0].environmentId;
      expect(emittedEnvironmentId).toBe(newSelectedEnvironment.meta.id);
    });
  });

  describe("Summary behavior", () => {
    test("section summary should be 'None' when selectedDataSources is empty", () => {
      expect(environmentSection.vm.environmentSummary).toBe("None");
    });

    test("section summary should be the selected environment name when environmentId is not empty", async () => {
      const environmentId = environments[0].meta.id;
      const environmentName = environments[0].meta.name;
      await environmentSection.setProps({ environmentId });
      expect(environmentSection.vm.environmentSummary).toBe(environmentName);
    });
  });

  describe("Environment variables behavior", () => {
    test("should render the correct amount of environment variables from the specificEnv", async () => {
      const selectedEnvironment: EnvironmentAsset = environmentVariablesEnvTest;

      const environmentSection = mount(EnvironmentSection, {
        global: {
          plugins: [Quasar, pinia],
          directives: {
            clickOutside: vi.fn(),
          },
        },
        props: {
          entityType: EWorkloadFormType.Workspace,
          specificEnv: {
            ...specificEnv,
            environmentVariables: selectedEnvironment.spec.environmentVariables || [],
            args: "",
            connections: [],
          },
          environmentId: selectedEnvironment.meta.id,
          loading: false,
          environments: [selectedEnvironment],
          clusterId: "123",
        },
        provide: {
          policy: {},
        },
        data() {
          return {
            authStore: useAuthStore(pinia),
          };
        },
      });

      expect(environmentSection.findAll(".environment-variable-box").length).toBe(
        selectedEnvironment.spec.environmentVariables?.length,
      );
    });
  });

  describe("Command and argumets behavior", () => {
    test("should render the '+ command & arguments' button when the specificEnv has no command or arguments", async () => {
      expect(environmentSection.find("[aid=add-command-and-args-button]").exists()).toBe(true);
    });

    test("should not render the '+ command & arguments' button when the specificEnv has command", async () => {
      expect(environmentSection.find("[aid=add-command-and-args-button]").exists()).toBe(true);
      await environmentSection.setProps({
        specificEnv: {
          command: "command",
          args: "",
          environmentVariables: [],
          connections: [],
        },
      });
      expect(environmentSection.find("[aid=add-command-and-args-button]").exists()).toBe(false);
    });

    test("should render the 'env-var-section' when the specificEnv has command", async () => {
      expect(environmentSection.find(".env-var-section").exists()).toBe(false);
      await environmentSection.setProps({
        specificEnv: {
          command: "command",
          args: "",
          environmentVariables: [],
          connections: [],
        },
      });
      expect(environmentSection.find(".env-var-section").exists()).toBe(true);
    });

    test("should not render the '+ command & arguments' button when the specificEnv has args", async () => {
      expect(environmentSection.find("[aid=add-command-and-args-button]").exists()).toBe(true);
      await environmentSection.setProps({
        specificEnv: {
          args: "--arg1",
          environmentVariables: [],
          connections: [],
        },
      });
      expect(environmentSection.find("[aid=add-command-and-args-button]").exists()).toBe(false);
    });

    test("should render the 'env-var-section' when the specificEnv has args", async () => {
      expect(environmentSection.find(".env-var-section").exists()).toBe(false);
      await environmentSection.setProps({
        specificEnv: {
          args: "--arg1",
          environmentVariables: [],
          connections: [],
        },
      });
      expect(environmentSection.find(".env-var-section").exists()).toBe(true);
    });

    test("should render the 'env-var-section' when the specificEnv has args or command with the command input disabled , and args input enabled", async () => {
      // This test is here to provide context to previous product requirements
      await environmentSection.setProps({
        specificEnv: {
          args: "--arg1",
          environmentVariables: [],
          connections: [],
        },
      });
      expect((environmentSection.find("[aid=command-input]").element as HTMLInputElement).disabled).toBe(false);
      expect((environmentSection.find("[aid=args-input]").element as HTMLInputElement).disabled).toBe(false);
    });
  });

  describe("Emits behavior", () => {
    let selectedEnvironmentSpecificEnv: IUIWorkloadEnvSectionModel;
    beforeEach(() => {
      environmentSection.setProps({
        environments,
        environmentId: "",
      });
      selectedEnvironmentSpecificEnv = {
        environmentVariables: [],
        command: customToolsEnvTest.spec.command || "",
        args: customToolsEnvTest.spec.args || "",
        connections: [
          {
            externalUrl: undefined,
            nodePort: undefined,
            name: "TensorBoard",
          },
          {
            externalUrl: undefined,
            nodePort: undefined,
            name: "Jupyter - NodePort",
          },
        ],
        runAsGid: customToolsEnvTest.spec.runAsGid,
        runAsUid: customToolsEnvTest.spec.runAsUid,
        supplementalGroups: customToolsEnvTest.spec.supplementalGroups,
      };
    });

    describe("environment changes", () => {
      test("should emit 'environment-changed' with the correct specificEnv of the selected environment when selecting environment", async () => {
        const selectedEnv = environments[0];
        await environmentSection.setProps({
          environments: [customToolsEnvTest],
        });
        await environmentSection.find(`[aid=${selectedEnv.meta.name}-card]`).trigger("click");
        const expectedSpecificEnv: IUIWorkloadEnvSectionModel = selectedEnvironmentSpecificEnv;
        const expectedEmit: TTestEmit<{ environmentId: string; specificEnv: IUIWorkloadEnvSectionModel }> = [
          [{ environmentId: customToolsEnvTest.meta.id, specificEnv: expectedSpecificEnv }],
        ];
        expect(environmentSection.emitted("environment-changed")).toEqual(expectedEmit);
      });
    });

    describe("connections changes", () => {
      test("should emit 'environment-changed' with the correct specificEnv of the selected environment when changing the tool custom url", async () => {
        await environmentSection.setProps({
          environments: [customToolsEnvTest],
          environmentId: customToolsEnvTest.meta.id,
          specificEnv: selectedEnvironmentSpecificEnv,
        });

        const newUrl = "http://custom-url.com/tom";
        await environmentSection.find("[aid=custom-url-input]").setValue(newUrl);

        const expectedSpecificEnv: IUIWorkloadEnvSectionModel = selectedEnvironmentSpecificEnv;
        // finding external connections and setting thier authorizedUsers
        customToolsEnvTest.spec.connections?.forEach((envTool: Connection) => {
          if (envTool.internalToolInfo?.connectionType === InternalConnectionType.ExternalUrl) {
            const connection: SpecificRunConnectionInfo | undefined = expectedSpecificEnv.connections.find(
              (con: SpecificRunConnectionInfo) => con.name === envTool.name,
            );

            if (connection) {
              connection.authorizedUsers = null;
              connection.authorizedGroups = null;
            }
          }
        });
        expectedSpecificEnv.connections[0].externalUrl = newUrl;

        const expectedEmit: TTestEmit<{ environmentId: string; specificEnv: IUIWorkloadEnvSectionModel }> = [
          [{ environmentId: customToolsEnvTest.meta.id, specificEnv: expectedSpecificEnv }],
        ];
        expect(environmentSection.emitted("environment-changed")).toEqual(expectedEmit);
      });

      test("should emit 'environment-changed' with the correct specificEnv of the selected environment when changing the tool custom port", async () => {
        await environmentSection.setProps({
          environments: [customToolsEnvTest],
          environmentId: customToolsEnvTest.meta.id,
          specificEnv: selectedEnvironmentSpecificEnv,
        });

        const newPort = 30200;
        await environmentSection.find("[aid=custom-port-input]").setValue(newPort);

        const expectedSpecificEnv: IUIWorkloadEnvSectionModel = selectedEnvironmentSpecificEnv;
        // finding external connections and setting thier authorizedUsers
        customToolsEnvTest.spec.connections?.forEach((envTool: Connection) => {
          if (envTool.internalToolInfo?.connectionType === InternalConnectionType.ExternalUrl) {
            const connection: SpecificRunConnectionInfo | undefined = expectedSpecificEnv.connections.find(
              (con: SpecificRunConnectionInfo) => con.name === envTool.name,
            );

            if (connection) {
              connection.authorizedUsers = null;
              connection.authorizedGroups = null;
            }
          }
        });

        expectedSpecificEnv.connections[1].nodePort = newPort;

        const expectedEmit: TTestEmit<{ environmentId: string; specificEnv: IUIWorkloadEnvSectionModel }> = [
          [{ environmentId: customToolsEnvTest.meta.id, specificEnv: expectedSpecificEnv }],
        ];
        expect(environmentSection.emitted("environment-changed")).toEqual(expectedEmit);
      });
    });

    describe("environment variables changes", () => {
      test("should emit 'environment-changed' with the new environment variable", async () => {
        const currSpecificEnv = {
          ...specificEnv,
          environmentVariables: [],
        };

        environmentSection = mount(EnvironmentSection, {
          global: {
            plugins: [Quasar, pinia],
            directives: {
              clickOutside: vi.fn(),
            },
          },
          props: {
            entityType: EWorkloadFormType.Workspace,
            specificEnv: currSpecificEnv,
            environmentId: "",
            loading: false,
            environments,
            clusterId: "123",
          },
          provide: {
            policy: {},
          },
          data() {
            return {
              authStore: useAuthStore(pinia),
            };
          },
        });

        environmentSection.vm.environmentVariablesChanged([{ name: "", value: "" }]);
        const emitted: TTestEmit<{ environmentId: string; specificEnv: IUIWorkloadEnvSectionModel }> | undefined =
          environmentSection.emitted("environment-changed");

        const emitVal = emitted ? emitted[0][0] : undefined;

        expect(emitted).toBeDefined();
        expect(emitVal?.specificEnv).toBeDefined();
        expect(emitVal?.specificEnv?.environmentVariables.length).toBe(currSpecificEnv.environmentVariables.length + 1);
      });
    });

    describe("arguments changes", () => {
      test("should emit 'environment-changed' with the new arguments", async () => {
        selectedEnvironmentSpecificEnv.args = "--args1";
        await environmentSection.setProps({
          environments: [customToolsEnvTest],
          environmentId: customToolsEnvTest.meta.id,
          specificEnv: selectedEnvironmentSpecificEnv,
        });

        const newArgs = "--args2";
        await environmentSection.find("[aid=args-input]").setValue(newArgs);

        const expectedSpecificEnv: IUIWorkloadEnvSectionModel = selectedEnvironmentSpecificEnv;
        expectedSpecificEnv.args = newArgs;
        const expectedEmit: TTestEmit<{ environmentId: string; specificEnv: IUIWorkloadEnvSectionModel }> = [
          [{ environmentId: customToolsEnvTest.meta.id, specificEnv: expectedSpecificEnv }],
        ];
        expect(environmentSection.emitted("environment-changed")).toEqual(expectedEmit);
      });
    });
  });
});
