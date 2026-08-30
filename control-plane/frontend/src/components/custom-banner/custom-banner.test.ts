import { mount } from "@vue/test-utils";
import { describe, it, expect } from "vitest";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import CustomBanner from "./custom-banner.vue";

const createWrapper = (bannerConfig = {}) => {
  return mount(CustomBanner, {
    global: {
      plugins: [
        Quasar,
        createTestingPinia({
          createSpy: () => () => {},
          initialState: {
            Setting: { bannerConfig },
          },
        }),
      ],
    },
  });
};

describe("CustomBanner", () => {
  it("does not render when enabled is false", () => {
    const wrapper = createWrapper({ enabled: false, text: "PROD", backgroundColor: "#D32F2F", fontColor: "#FFFFFF" });
    expect(wrapper.find(".custom-banner").exists()).toBe(false);
  });

  it("does not render when bannerConfig is undefined", () => {
    const wrapper = createWrapper(undefined);
    expect(wrapper.find(".custom-banner").exists()).toBe(false);
  });

  it("renders when enabled is true", () => {
    const wrapper = createWrapper({
      enabled: true,
      text: "PRODUCTION ENVIRONMENT",
      backgroundColor: "#D32F2F",
      fontColor: "#FFFFFF",
    });
    expect(wrapper.find(".custom-banner").exists()).toBe(true);
  });

  it("displays the banner text", () => {
    const wrapper = createWrapper({
      enabled: true,
      text: "PRODUCTION ENVIRONMENT",
      backgroundColor: "#D32F2F",
      fontColor: "#FFFFFF",
    });
    expect(wrapper.find(".custom-banner-text").text()).toBe("PRODUCTION ENVIRONMENT");
  });

  it("applies backgroundColor style", () => {
    const wrapper = createWrapper({ enabled: true, text: "PROD", backgroundColor: "#D32F2F", fontColor: "#FFFFFF" });
    expect(wrapper.find(".custom-banner").attributes("style")).toContain("background-color: rgb(211, 47, 47)");
  });

  it("applies fontColor style", () => {
    const wrapper = createWrapper({ enabled: true, text: "PROD", backgroundColor: "#D32F2F", fontColor: "#FFFFFF" });
    expect(wrapper.find(".custom-banner").attributes("style")).toContain("color: rgb(255, 255, 255)");
  });

  it("applies fontSize style", () => {
    const wrapper = createWrapper({
      enabled: true,
      text: "PROD",
      backgroundColor: "#D32F2F",
      fontColor: "#FFFFFF",
      fontSize: 14,
    });
    expect(wrapper.find(".custom-banner").attributes("style")).toContain("font-size: 14px");
  });

  it("caps fontSize at 20px", () => {
    const wrapper = createWrapper({
      enabled: true,
      text: "PROD",
      backgroundColor: "#D32F2F",
      fontColor: "#FFFFFF",
      fontSize: 100,
    });
    expect(wrapper.find(".custom-banner").attributes("style")).toContain("font-size: 20px");
  });

  it("applies default fontSize of 14px when not set", () => {
    const wrapper = createWrapper({ enabled: true, text: "PROD", backgroundColor: "#D32F2F", fontColor: "#FFFFFF" });
    expect(wrapper.find(".custom-banner").attributes("style")).toContain("font-size: 14px");
  });

  it("renders text as plain string — HTML tags are not interpreted", () => {
    const wrapper = createWrapper({
      enabled: true,
      text: "<b>PROD</b>",
      backgroundColor: "#D32F2F",
      fontColor: "#FFFFFF",
    });
    expect(wrapper.find(".custom-banner-text").text()).toBe("<b>PROD</b>");
    expect(wrapper.find(".custom-banner-text b").exists()).toBe(false);
  });

  it("renders gracefully when backgroundColor is empty", () => {
    const wrapper = createWrapper({ enabled: true, text: "PROD", backgroundColor: "", fontColor: "#FFFFFF" });
    expect(wrapper.find(".custom-banner").exists()).toBe(true);
  });

  it("renders gracefully when fontColor is empty", () => {
    const wrapper = createWrapper({ enabled: true, text: "PROD", backgroundColor: "#D32F2F", fontColor: "" });
    expect(wrapper.find(".custom-banner").exists()).toBe(true);
  });

  it("renders gracefully when text is empty string", () => {
    const wrapper = createWrapper({ enabled: true, text: "", backgroundColor: "#D32F2F", fontColor: "#FFFFFF" });
    expect(wrapper.find(".custom-banner").exists()).toBe(true);
    expect(wrapper.find(".custom-banner-text").text()).toBe("");
  });
});
