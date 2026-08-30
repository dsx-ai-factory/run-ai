import { describe, expect, test } from "vitest";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import RoleStageBadge from "./role-stage-badge.vue";
import { ERoleStage } from "@/models/roles.model";

const mountComponent = (props: { roleName: string; hiddenStages?: string[] }) => {
  return mount(RoleStageBadge, {
    global: { plugins: [Quasar] },
    props,
  });
};

describe("RoleStageBadge", () => {
  test("renders Legacy badge for a legacy role", () => {
    const wrapper = mountComponent({ roleName: "L1 researcher" });
    expect(wrapper.find(".q-btn").exists()).toBe(true);
    expect(wrapper.find(".q-btn").text()).toBe("Legacy");
    expect(wrapper.find(".bg-grey-4").exists()).toBe(true);
  });

  test("renders New badge for a new role", () => {
    const wrapper = mountComponent({ roleName: "AI practitioner" });
    expect(wrapper.find(".q-btn").exists()).toBe(true);
    expect(wrapper.find(".q-btn").text()).toBe("New");
    expect(wrapper.find(".bg-green-12").exists()).toBe(true);
  });

  test("renders nothing for an unknown role", () => {
    const wrapper = mountComponent({ roleName: "some-custom-role" });
    expect(wrapper.find(".q-btn").exists()).toBe(false);
  });

  test("hides badge when stage is in hiddenStages", () => {
    const wrapper = mountComponent({ roleName: "AI practitioner", hiddenStages: [ERoleStage.New] });
    expect(wrapper.find(".q-btn").exists()).toBe(false);
  });

  test("shows Legacy badge when only New is hidden", () => {
    const wrapper = mountComponent({ roleName: "L1 researcher", hiddenStages: [ERoleStage.New] });
    expect(wrapper.find(".q-btn").exists()).toBe(true);
    expect(wrapper.find(".q-btn").text()).toBe("Legacy");
  });

  test("hides Legacy badge when Legacy is in hiddenStages", () => {
    const wrapper = mountComponent({ roleName: "L1 researcher", hiddenStages: [ERoleStage.Legacy] });
    expect(wrapper.find(".q-btn").exists()).toBe(false);
  });

  test("hides all badges when both stages are hidden", () => {
    const wrapperNew = mountComponent({
      roleName: "AI practitioner",
      hiddenStages: [ERoleStage.New, ERoleStage.Legacy],
    });
    const wrapperLegacy = mountComponent({
      roleName: "L1 researcher",
      hiddenStages: [ERoleStage.New, ERoleStage.Legacy],
    });
    expect(wrapperNew.find(".q-btn").exists()).toBe(false);
    expect(wrapperLegacy.find(".q-btn").exists()).toBe(false);
  });
});
