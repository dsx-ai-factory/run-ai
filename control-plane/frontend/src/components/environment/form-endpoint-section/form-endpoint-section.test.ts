import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import { Quasar } from "quasar";
import FormEndpointSection from "./form-endpoint-section.vue";
import { ServingPortInfoProtocolEnum } from "@/swagger-models/assets-service-client";
import { InternalConnectionType } from "@/swagger-models/assets-service-client";
import { ToolType } from "@/swagger-models/assets-service-client";
import type { IUIConnection } from "@/models/environment.model";

describe("form-endpoint-section.vue", () => {
  const defaultConnection: IUIConnection = {
    id: InternalConnectionType.ServingPort,
    name: InternalConnectionType.ServingPort,
    toolType: ToolType.Custom,
    connectionType: InternalConnectionType.ServingPort,
    protocol: ServingPortInfoProtocolEnum.Http,
    containerPort: 8000,
  };

  function mountComponent(props: {
    connection: IUIConnection;
    protocolLockedToHttp?: boolean;
  }): VueWrapper<InstanceType<typeof FormEndpointSection>> {
    return mount(FormEndpointSection, {
      global: { plugins: [Quasar] },
      props: {
        connection: props.connection,
        protocolLockedToHttp: props.protocolLockedToHttp ?? false,
      },
    });
  }

  it("renders endpoint section with protocol and container port", () => {
    const wrapper = mountComponent({ connection: defaultConnection });
    expect(wrapper.text()).toContain("Set an inference serving endpoint");
    expect(wrapper.find("[aid='connection-protocol']").exists()).toBe(true);
    expect(wrapper.find("[aid='container-port']").exists()).toBe(true);
  });

  it("emits update when protocol is changed", async () => {
    const wrapper = mountComponent({ connection: defaultConnection });
    const select = wrapper.findComponent({ name: "runai-select" });
    await select.vm.$emit("update:model-value", { value: ServingPortInfoProtocolEnum.Grpc });
    const emitted = wrapper.emitted("update");
    expect(emitted).toHaveLength(1);
    expect(emitted?.[0]?.[0]).toMatchObject({
      ...defaultConnection,
      protocol: ServingPortInfoProtocolEnum.Grpc,
    });
  });

  describe("protocolLockedToHttp", () => {
    it("does not disable protocol select when protocolLockedToHttp is false", () => {
      const wrapper = mountComponent({
        connection: defaultConnection,
        protocolLockedToHttp: false,
      });
      const protocolSelect = wrapper.find("[aid='connection-protocol']");
      expect(protocolSelect.attributes("disable")).toBeUndefined();
      const runaiSelect = wrapper.findComponent({ name: "runai-select" });
      expect(runaiSelect.props("disable")).toBe(false);
    });

    it("disables protocol select when protocolLockedToHttp is true", () => {
      const wrapper = mountComponent({
        connection: defaultConnection,
        protocolLockedToHttp: true,
      });
      const runaiSelect = wrapper.findComponent({ name: "runai-select" });
      expect(runaiSelect.props("disable")).toBe(true);
    });

    it("renders tooltip for distributed inference when protocolLockedToHttp is true", () => {
      const wrapper = mountComponent({
        connection: defaultConnection,
        protocolLockedToHttp: true,
      });
      const tooltip = wrapper.findComponent({ name: "q-tooltip" });
      expect(tooltip.exists()).toBe(true);
    });

    it("does not show distributed inference tooltip when protocolLockedToHttp is false", () => {
      const wrapper = mountComponent({
        connection: defaultConnection,
        protocolLockedToHttp: false,
      });
      expect(wrapper.text()).not.toContain("For distributed inference only HTTP is supported");
    });
  });
});
