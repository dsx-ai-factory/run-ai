/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi } from "vitest";
import { mount, type VueWrapper } from "@vue/test-utils";
import { createTestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";
import { Quasar } from "quasar";

// cmps
import ServingEndpointAccess from "./serving-endpoint-access.vue";

// models
import { EAccessOptions } from "@/components/environment/connection-access/connection-access-modal/connection-access-modal.model";
import {
  EServingEndpointAccessEvents,
  ProtocolOptions,
  type IServingEndpointAccessData,
  ENDPOINT_TYPE_EXTERNAL,
  ENDPOINT_TYPE_INTERNAL,
} from "./serving-endpoint-access.model";
import type { TTestEmit } from "@/models/unit-test.model";
import type { SupersetServingPortRules } from "@/swagger-models/policy-service-client";

// stores
import { useAuthStore } from "@/stores/auth.store";
import { useClusterStore } from "@/stores/cluster.store";
import { errorMessages } from "@/common/error-message.constant";
import { ServingPortProtocol } from "@/swagger-models/workloads-client";
import type { ISelectOption } from "@/models/global.model";

const pinia = createTestingPinia({ createSpy: vi.fn });
setActivePinia(pinia);

interface IServingEndpointAccessProps {
  modelValue: IServingEndpointAccessData;
  title: string;
  clusterId: string;
  titleTooltip?: string;
  protocolDisabled?: boolean;
  portDisabled?: boolean;
  showProtocolAndPort?: boolean;
  disabled?: boolean;
  disabledInfo?: string;
  policyRules?: SupersetServingPortRules | null;
  isDistributedInference?: boolean;
}

describe("serving-endpoint-access.vue", () => {
  const requiredProps: IServingEndpointAccessProps = {
    modelValue: {
      access: { accessOption: EAccessOptions.PUBLIC, authorizedUsers: null, authorizedGroups: null },
      clusterLocalAccessOnly: false,
    },
    title: "Test Title",
    clusterId: "cluster-v2.24",
  };

  function getMountedComponent(
    props: IServingEndpointAccessProps,
  ): VueWrapper<InstanceType<typeof ServingEndpointAccess>> {
    const authStore = useAuthStore(pinia);
    authStore.$patch({ currentUser: { username: "admin@mail.com" } });

    // Mock cluster store
    const clusterStore = useClusterStore(pinia);
    clusterStore.$patch({
      clusters: [
        {
          uuid: "cluster-with-config-domain",
          dependenciesStatus: {
            optional: {
              knative: {
                available: true,
                components: {
                  kourier: {
                    components: {
                      "external-config-domain": {
                        available: true,
                      },
                    },
                  },
                },
              },
            },
          },
        },
      ],
    });

    clusterStore.isClusterVersionSufficient = vi.fn().mockImplementation((clusterId: string, version: string) => {
      // Simulate version 2.24+ for specific cluster IDs
      if (clusterId === "cluster-v2.24" || clusterId === "cluster-with-config-domain") {
        return true;
      }
      // Simulate version < 2.24 for old clusters
      if (clusterId === "cluster-v2.23" || clusterId === "cluster-without-config-domain") {
        return version !== "2.24"; // Returns false for 2.24 check
      }
      return false;
    });

    return mount(ServingEndpointAccess, {
      global: {
        plugins: [Quasar, pinia],
      },
      props,
    });
  }

  describe("Props", () => {
    it("renders with required props", () => {
      const wrapper = getMountedComponent(requiredProps);
      expect(wrapper.text()).toContain("Test Title");
      expect(wrapper.find("[aid='connection-protocol']").exists()).toBe(false);
      expect(wrapper.find("[aid='container-port']").exists()).toBe(false);
      expect(wrapper.findComponent({ name: "connection-access" }).exists()).toBe(true);
    });

    it("renders protocol", () => {
      // editable
      const wrapper = getMountedComponent({ ...requiredProps, showProtocolAndPort: true });
      expect(wrapper.find("[aid='connection-protocol']").exists()).toBe(true);

      // disabled
      const wrapper2 = getMountedComponent({ ...requiredProps, showProtocolAndPort: true, protocolDisabled: true });
      expect(wrapper2.find("[aid='connection-protocol']").exists()).toBe(true);
    });

    it("renders protocol select when protocolEditable is true", () => {
      const wrapper = getMountedComponent({ ...requiredProps, showProtocolAndPort: true });
      expect(wrapper.findComponent({ name: "RunaiSelect" }).exists()).toBe(true);
    });

    it("renders port", () => {
      // editable
      const wrapper = getMountedComponent({ ...requiredProps, showProtocolAndPort: true });
      expect(wrapper.find("[aid='container-port']").exists()).toBe(true);

      // disabled
      const wrapper2 = getMountedComponent({ ...requiredProps, showProtocolAndPort: true, portDisabled: true });
      expect(wrapper2.find("[aid='container-port']").exists()).toBe(true);
    });
  });

  describe("Emits", () => {
    it("emits an event when protocol is updated", async () => {
      const wrapper = getMountedComponent({ ...requiredProps, showProtocolAndPort: true });
      const option = ProtocolOptions[1];
      await wrapper.findComponent({ name: "RunaiSelect" }).vm.$emit("update:model-value", option);
      expect(wrapper.emitted()[EServingEndpointAccessEvents.UpdateModelValue]).toBeTruthy();
      const emitted: TTestEmit<IServingEndpointAccessData> = wrapper.emitted(
        EServingEndpointAccessEvents.UpdateModelValue,
      ) || [[]];
      // Should be index 0 since we explicitly set clusterLocalAccessOnly in props
      expect(emitted[0][0].protocol).toBe(option.value);
      expect(emitted[0][0].clusterLocalAccessOnly).toBe(false);
    });

    it("emits an event when port is updated", async () => {
      const wrapper = getMountedComponent({ ...requiredProps, showProtocolAndPort: true });
      const portElm = wrapper.find("[aid='container-port']");
      await portElm.setValue("9090");
      expect(wrapper.emitted()[EServingEndpointAccessEvents.UpdateModelValue]).toBeTruthy();
      const emitted: TTestEmit<IServingEndpointAccessData> = wrapper.emitted(
        EServingEndpointAccessEvents.UpdateModelValue,
      ) || [[]];
      // Should be index 0 since we explicitly set clusterLocalAccessOnly in props
      expect(emitted[0][0].port).toBe(9090);
      expect(emitted[0][0].clusterLocalAccessOnly).toBe(false);
    });

    it("emits access-changed event when onAccessChanged is called", async () => {
      const defaultProps: IServingEndpointAccessProps = {
        modelValue: {
          access: { accessOption: EAccessOptions.PUBLIC, authorizedUsers: null, authorizedGroups: null },
          clusterLocalAccessOnly: false,
        },
        title: "Test Title",
        titleTooltip: "Tooltip text",
        clusterId: "cluster-v2.24",
      };
      const wrapper = getMountedComponent(defaultProps);

      const newAccessData = { accessOption: EAccessOptions.EVERYONE, authorizedUsers: null, authorizedGroups: null };
      await wrapper.findComponent({ name: "ConnectionAccess" }).vm.$emit("save", newAccessData);

      expect(wrapper.emitted()[EServingEndpointAccessEvents.UpdateModelValue]).toBeTruthy();
      const emitted: TTestEmit<IServingEndpointAccessData> = wrapper.emitted(
        EServingEndpointAccessEvents.UpdateModelValue,
      ) || [[]];
      // Should be index 0 since we explicitly set clusterLocalAccessOnly in props
      expect(emitted[0][0].access).toEqual(newAccessData);
      expect(emitted[0][0].clusterLocalAccessOnly).toBe(false);
    });
  });

  describe("Endpoint Type (External/Internal)", () => {
    it("initializes with internal-only when clusterLocalAccessOnly is undefined and cluster supports v2.24+", () => {
      const props: IServingEndpointAccessProps = {
        modelValue: {
          access: { accessOption: EAccessOptions.PUBLIC, authorizedUsers: null, authorizedGroups: null },
          // clusterLocalAccessOnly is undefined
        },
        title: "Test Title",
        clusterId: "cluster-v2.24", // Use cluster that supports v2.24+
      };
      const wrapper = getMountedComponent(props);

      // Should emit initialization with clusterLocalAccessOnly=true (internal) since config-domain not available
      const emitted: TTestEmit<IServingEndpointAccessData> = wrapper.emitted(
        EServingEndpointAccessEvents.UpdateModelValue,
      ) || [[]];
      expect(emitted[0][0].clusterLocalAccessOnly).toBe(true);
    });

    it("does not initialize when cluster does not support v2.24", () => {
      const props: IServingEndpointAccessProps = {
        modelValue: {
          access: { accessOption: EAccessOptions.PUBLIC, authorizedUsers: null, authorizedGroups: null },
          // clusterLocalAccessOnly is undefined
        },
        title: "Test Title",
        clusterId: "cluster-v2.23", // Use old cluster version
      };
      const wrapper = getMountedComponent(props);

      // Should NOT emit initialization for old cluster
      const emitted: TTestEmit<IServingEndpointAccessData> | undefined = wrapper.emitted(
        EServingEndpointAccessEvents.UpdateModelValue,
      );
      expect(emitted).toBeUndefined();
    });

    it("shows external access option as disabled when config-domain is not available", () => {
      const props: IServingEndpointAccessProps = {
        ...requiredProps,
        clusterId: "cluster-without-config-domain",
      };
      const wrapper = getMountedComponent(props);

      expect(wrapper.vm.isExternalAccessAvailable).toBe(false);
      expect(wrapper.vm.endpointTypeOptions[0].disable).toBe(true); // External option disabled
      expect(wrapper.vm.endpointTypeOptions[1].disable).toBeFalsy(); // Internal option enabled
    });

    it("shows external access option as enabled when config-domain is available", () => {
      // Mock cluster store to return config-domain as available
      const props: IServingEndpointAccessProps = {
        ...requiredProps,
        clusterId: "cluster-with-config-domain",
      };
      const wrapper = getMountedComponent(props);

      // The store is already mocked in getMountedComponent to return true for "cluster-with-config-domain"
      expect(wrapper.vm.isExternalAccessAvailable).toBe(true);
      expect(wrapper.vm.endpointTypeOptions[0].disable).toBe(false); // External option enabled
      expect(wrapper.vm.endpointTypeOptions[1].disable).toBeFalsy(); // Internal option enabled
    });

    it("shows external access option as enabled for distributed inference without config-domain", () => {
      const props: IServingEndpointAccessProps = {
        ...requiredProps,
        clusterId: "cluster-without-config-domain",
        isDistributedInference: true,
      };
      const wrapper = getMountedComponent(props);

      expect(wrapper.vm.isExternalAccessAvailable).toBe(true);
      expect(wrapper.vm.endpointTypeOptions[0].disable).toBe(false);
      expect(wrapper.vm.endpointTypeOptions[0].disabledTooltip).toBeUndefined();
    });

    it("initializes distributed inference with external access without config-domain", () => {
      const props: IServingEndpointAccessProps = {
        modelValue: {
          access: { accessOption: EAccessOptions.PUBLIC, authorizedUsers: null, authorizedGroups: null },
        },
        title: "Test Title",
        clusterId: "cluster-v2.24",
        isDistributedInference: true,
      };
      const wrapper = getMountedComponent(props);

      const emitted: TTestEmit<IServingEndpointAccessData> = wrapper.emitted(
        EServingEndpointAccessEvents.UpdateModelValue,
      ) || [[]];
      expect(emitted[0][0].clusterLocalAccessOnly).toBe(false);
    });

    it("displays 'external' as selected when clusterLocalAccessOnly is false and external access is available", () => {
      const props: IServingEndpointAccessProps = {
        ...requiredProps,
        modelValue: {
          ...requiredProps.modelValue,
          clusterLocalAccessOnly: false,
        },
        clusterId: "cluster-with-config-domain",
      };
      const wrapper = getMountedComponent(props);

      expect(wrapper.vm.endpointTypeValue).toBe(ENDPOINT_TYPE_EXTERNAL);
    });

    it("displays 'internal' as selected when clusterLocalAccessOnly is true", () => {
      const props: IServingEndpointAccessProps = {
        ...requiredProps,
        modelValue: {
          ...requiredProps.modelValue,
          clusterLocalAccessOnly: true,
        },
      };
      const wrapper = getMountedComponent(props);

      expect(wrapper.vm.endpointTypeValue).toBe(ENDPOINT_TYPE_INTERNAL);
    });

    it("emits update when endpoint type is changed to internal", async () => {
      const props: IServingEndpointAccessProps = {
        ...requiredProps,
        modelValue: {
          ...requiredProps.modelValue,
          clusterLocalAccessOnly: false,
        },
        clusterId: "cluster-v2.24",
      };
      const wrapper = getMountedComponent(props);

      const radioOptions = wrapper.findComponent({ name: "runai-radio-options" });
      await radioOptions.vm.$emit("update:model-value", ENDPOINT_TYPE_INTERNAL);

      const emitted: TTestEmit<IServingEndpointAccessData> = wrapper.emitted(
        EServingEndpointAccessEvents.UpdateModelValue,
      ) || [[]];
      const lastEmit = emitted[emitted.length - 1][0];
      expect(lastEmit.clusterLocalAccessOnly).toBe(true);
    });

    it("emits update when endpoint type is changed to external", async () => {
      const props: IServingEndpointAccessProps = {
        ...requiredProps,
        modelValue: {
          ...requiredProps.modelValue,
          clusterLocalAccessOnly: true,
        },
        clusterId: "cluster-with-config-domain",
      };
      const wrapper = getMountedComponent(props);

      const radioOptions = wrapper.findComponent({ name: "runai-radio-options" });
      await radioOptions.vm.$emit("update:model-value", ENDPOINT_TYPE_EXTERNAL);

      const emitted: TTestEmit<IServingEndpointAccessData> = wrapper.emitted(
        EServingEndpointAccessEvents.UpdateModelValue,
      ) || [[]];
      const lastEmit = emitted[emitted.length - 1][0];
      expect(lastEmit.clusterLocalAccessOnly).toBe(false);
    });

    it("renders runai-radio-options component for endpoint type selection", () => {
      const props: IServingEndpointAccessProps = {
        ...requiredProps,
        clusterId: "cluster-v2.24",
      };
      const wrapper = getMountedComponent(props);

      const radioOptions = wrapper.findComponent({ name: "runai-radio-options" });
      expect(radioOptions.exists()).toBe(true);
      expect(radioOptions.props("options")).toHaveLength(2);
    });

    it("displays disabled tooltip when external access is unavailable", () => {
      const props: IServingEndpointAccessProps = {
        ...requiredProps,
        clusterId: "cluster-v2.24",
      };
      const wrapper = getMountedComponent(props);

      expect(wrapper.vm.isExternalAccessAvailable).toBe(false);

      const radioOptions = wrapper.findComponent({ name: "runai-radio-options" });
      const options = radioOptions.props("options") as ISelectOption[];
      const externalOption = options.find((opt) => opt.value === ENDPOINT_TYPE_EXTERNAL);

      expect(externalOption?.disable).toBe(true);
      expect(externalOption?.disabledTooltip).toContain("To enable external access, contact your administrator about");
      expect(externalOption?.svg).toBe("ban");
    });

    it("does not have disabled tooltip when external access is available", () => {
      const props: IServingEndpointAccessProps = {
        ...requiredProps,
        clusterId: "cluster-with-config-domain",
      };
      const wrapper = getMountedComponent(props);

      expect(wrapper.vm.isExternalAccessAvailable).toBe(true);

      const radioOptions = wrapper.findComponent({ name: "runai-radio-options" });
      const options = radioOptions.props("options") as ISelectOption[];
      const externalOption = options.find((opt) => opt.value === ENDPOINT_TYPE_EXTERNAL);

      expect(externalOption?.disable).toBeFalsy();
      expect(externalOption?.disabledTooltip).toBeUndefined();
      expect(externalOption?.svg).toBeUndefined();
    });

    it("always displays 'internal' when external access is not available, regardless of clusterLocalAccessOnly value", () => {
      const props1: IServingEndpointAccessProps = {
        ...requiredProps,
        modelValue: {
          ...requiredProps.modelValue,
          clusterLocalAccessOnly: false,
        },
        clusterId: "cluster-without-config-domain",
      };
      const wrapper1 = getMountedComponent(props1);
      expect(wrapper1.vm.endpointTypeValue).toBe(ENDPOINT_TYPE_INTERNAL);

      const props2: IServingEndpointAccessProps = {
        modelValue: {
          access: { accessOption: EAccessOptions.PUBLIC, authorizedUsers: null, authorizedGroups: null },
        },
        title: "Test Title",
        clusterId: "cluster-without-config-domain",
      };
      const wrapper2 = getMountedComponent(props2);
      expect(wrapper2.vm.endpointTypeValue).toBe(ENDPOINT_TYPE_INTERNAL);
    });
  });

  describe("Policy logic", () => {
    const basePolicyRules = {
      protocol: { canEdit: true },
      container: { canEdit: true },
    };

    it("protocolShouldBeDisabled: true when protocolDisabled is true", () => {
      const wrapper = getMountedComponent({ ...requiredProps, protocolDisabled: true, policyRules: basePolicyRules });
      expect(wrapper.vm.protocolShouldBeDisabled).toBe(true);
    });

    it("protocolShouldBeDisabled: true when disabled is true", () => {
      const wrapper = getMountedComponent({ ...requiredProps, disabled: true, policyRules: basePolicyRules });
      expect(wrapper.vm.protocolShouldBeDisabled).toBe(true);
    });

    it("protocolShouldBeDisabled: true when policyRules.protocol.canEdit is false", () => {
      const policyRules = { ...basePolicyRules, protocol: { canEdit: false } };
      const wrapper = getMountedComponent({ ...requiredProps, policyRules });
      expect(wrapper.vm.protocolShouldBeDisabled).toBe(true);
    });

    it("protocolShouldBeDisabled: false when all are false", () => {
      const wrapper = getMountedComponent({ ...requiredProps, policyRules: basePolicyRules });
      expect(wrapper.vm.protocolShouldBeDisabled).toBe(false);
    });

    it("protocolDisabledInfo: returns disabledInfo if disabled and disabledInfo", () => {
      const wrapper = getMountedComponent({
        ...requiredProps,
        disabled: true,
        disabledInfo: "Disabled by admin",
        policyRules: basePolicyRules,
      });
      expect(wrapper.vm.protocolDisabledInfo).toBe("Disabled by admin");
    });

    it("protocolDisabledInfo: returns CANT_BE_MODIFIED_POLICY if policyRules.protocol.canEdit is false", () => {
      const policyRules = { ...basePolicyRules, protocol: { canEdit: false } };
      const wrapper = getMountedComponent({ ...requiredProps, policyRules });
      expect(wrapper.vm.protocolDisabledInfo).toBe(errorMessages.CANT_BE_MODIFIED_POLICY);
    });

    it("protocolDisabledInfo: returns container/port defined message if portDisabled", () => {
      const wrapper = getMountedComponent({ ...requiredProps, portDisabled: true, policyRules: basePolicyRules });
      expect(wrapper.vm.protocolDisabledInfo).toBe(
        "The connection protocol and the container port are defined within the environment",
      );
    });

    it("protocolDisabledInfo: returns undefined if none apply", () => {
      const wrapper = getMountedComponent({ ...requiredProps, policyRules: basePolicyRules });
      expect(wrapper.vm.protocolDisabledInfo).toBe(undefined);
    });

    it("portDisabledInfo: returns disabledInfo if disabled and disabledInfo", () => {
      const wrapper = getMountedComponent({
        ...requiredProps,
        disabled: true,
        disabledInfo: "Disabled by admin",
        policyRules: basePolicyRules,
      });
      expect(wrapper.vm.portDisabledInfo).toBe("Disabled by admin");
    });

    it("portDisabledInfo: returns container/port defined message if portDisabled", () => {
      const wrapper = getMountedComponent({ ...requiredProps, portDisabled: true, policyRules: basePolicyRules });
      expect(wrapper.vm.portDisabledInfo).toBe(
        "The connection protocol and the container port are defined within the environment",
      );
    });

    it("portDisabledInfo: returns undefined if none apply", () => {
      const wrapper = getMountedComponent({ ...requiredProps, policyRules: basePolicyRules });
      expect(wrapper.vm.portDisabledInfo).toBe(undefined);
    });

    it("protocolRules: returns protocol rules with isClosedList true if present", () => {
      const policyRules: SupersetServingPortRules = {
        ...basePolicyRules,
        protocol: { canEdit: true, options: [{ value: ServingPortProtocol.Http }] },
      };
      const wrapper = getMountedComponent({ ...requiredProps, policyRules });
      expect(wrapper.vm.protocolRules).toEqual({
        canEdit: true,
        options: [{ value: ServingPortProtocol.Http }],
        isClosedList: true,
      });
    });

    it("protocolRules: returns undefined if no protocol rules", () => {
      const wrapper = getMountedComponent({ ...requiredProps, policyRules: {} });
      expect(wrapper.vm.protocolRules).toBe(undefined);
    });

    it("passes policyRules to child components", () => {
      const wrapper = getMountedComponent({ ...requiredProps, showProtocolAndPort: true, policyRules: basePolicyRules });
      // PolicyNumberInput for port
      const portInput = wrapper.findComponent({ name: "policy-number-input" });
      expect(portInput.props("policyRules")).toStrictEqual(basePolicyRules.container);
      // ConnectionAccess for access
      const connAccess = wrapper.findComponent({ name: "connection-access" });
      expect(connAccess.props("policyRules")).toStrictEqual(basePolicyRules);
    });
  });

  describe("Cluster Version Checking", () => {
    it("hides endpoint type section when cluster version is less than 2.24", () => {
      const props: IServingEndpointAccessProps = {
        ...requiredProps,
        clusterId: "cluster-v2.23",
      };
      const wrapper = getMountedComponent(props);

      expect(wrapper.vm.supportsClusterLocalAccess).toBe(false);
      const runaiSections = wrapper.findAllComponents({ name: "runai-section" });
      const endpointAccessSection = runaiSections.find(
        (section) =>
          section.props("title") ===
          "Set whether the serving endpoint can be accessed externally or only by internal cluster traffic",
      );
      expect(endpointAccessSection).toBeFalsy();
    });

    it("shows endpoint type section when cluster version is 2.24 or higher", () => {
      const props: IServingEndpointAccessProps = {
        ...requiredProps,
        clusterId: "cluster-v2.24",
      };
      const wrapper = getMountedComponent(props);

      expect(wrapper.vm.supportsClusterLocalAccess).toBe(true);
      const runaiSections = wrapper.findAllComponents({ name: "runai-section" });
      const endpointAccessSection = runaiSections.find(
        (section) =>
          section.props("title") ===
          "Set whether the serving endpoint can be accessed externally or only by internal cluster traffic",
      );
      expect(endpointAccessSection).toBeTruthy();

      const radioOptions = wrapper.findComponent({ name: "runai-radio-options" });
      expect(radioOptions.exists()).toBe(true);
    });

    it("does not initialize clusterLocalAccessOnly when cluster version is less than 2.24", () => {
      const props: IServingEndpointAccessProps = {
        ...requiredProps,
        modelValue: {
          access: { accessOption: EAccessOptions.PUBLIC, authorizedUsers: null, authorizedGroups: null },
          // clusterLocalAccessOnly is undefined
        },
        clusterId: "cluster-v2.23",
      };
      const wrapper = getMountedComponent(props);

      // Should not emit initialization
      const emitted = wrapper.emitted(EServingEndpointAccessEvents.UpdateModelValue);
      expect(emitted).toBeUndefined();
    });

    it("initializes clusterLocalAccessOnly when cluster version is 2.24 or higher", () => {
      const props: IServingEndpointAccessProps = {
        ...requiredProps,
        modelValue: {
          access: { accessOption: EAccessOptions.PUBLIC, authorizedUsers: null, authorizedGroups: null },
          // clusterLocalAccessOnly is undefined
        },
        clusterId: "cluster-v2.24",
      };
      const wrapper = getMountedComponent(props);

      // Should emit initialization
      const emitted: TTestEmit<IServingEndpointAccessData> = wrapper.emitted(
        EServingEndpointAccessEvents.UpdateModelValue,
      ) || [[]];
      expect(emitted).toBeDefined();
      expect(emitted[0][0].clusterLocalAccessOnly).toBe(true); // Should default to internal (true)
    });

    it("does not process radio button clicks when cluster version is less than 2.24", async () => {
      const props: IServingEndpointAccessProps = {
        ...requiredProps,
        modelValue: {
          ...requiredProps.modelValue,
          clusterLocalAccessOnly: false,
        },
        clusterId: "cluster-v2.23",
      };
      const wrapper = getMountedComponent(props);

      await wrapper.vm.onEndpointTypeChanged(ENDPOINT_TYPE_INTERNAL);

      const emitted = wrapper.emitted(EServingEndpointAccessEvents.UpdateModelValue);
      expect(emitted).toBeUndefined();
    });

    it("sends undefined clusterLocalAccessOnly value when cluster version is less than 2.24", () => {
      const props: IServingEndpointAccessProps = {
        ...requiredProps,
        modelValue: {
          access: { accessOption: EAccessOptions.PUBLIC, authorizedUsers: null, authorizedGroups: null },
          // clusterLocalAccessOnly is undefined - should remain undefined for old clusters
        },
        clusterId: "cluster-v2.23",
      };
      const wrapper = getMountedComponent(props);

      // Value should remain undefined
      expect(wrapper.vm.modelValue.clusterLocalAccessOnly).toBeUndefined();
    });
  });
});
