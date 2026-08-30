import { beforeEach, describe, expect, it, vi } from "vitest";

const bootConfigState = vi.hoisted(() => ({ supportUrl: "" }));

vi.mock("@/services/infra/boot-config.service/boot-config.service", () => ({
  getBootConfig: () => ({
    docs: {
      baseUrl: "https://docs.example.test/product",
      whatsNew: { label: "Updates", url: "https://docs.example.test/product/updates" },
    },
    support: { url: bootConfigState.supportUrl, label: "Example support" },
  }),
}));

describe("email settings support copy", () => {
  beforeEach(() => {
    vi.resetModules();
    bootConfigState.supportUrl = "";
  });

  it("removes the support clause when the support URL is missing", async () => {
    const { EmailErrorMessages } = await import("./setting.model");

    expect(EmailErrorMessages.EMAIL_SAVED_ERROR).toBe("Email server couldn't be saved. Try again.");
    expect(EmailErrorMessages.EMAIL_NOTIFICATION_ENABLE_ERROR).toBe(
      "Email notifications couldn't be enabled. Try again.",
    );
  });

  it("retains the support clause when the support URL is configured", async () => {
    bootConfigState.supportUrl = "https://support.example.test/contact";
    const { EmailErrorMessages } = await import("./setting.model");

    expect(EmailErrorMessages.EMAIL_SAVED_ERROR).toBe(
      "Email server couldn't be saved. Try again. If the issue persists, contact Run:ai support.",
    );
    expect(EmailErrorMessages.EMAIL_NOTIFICATION_ENABLE_ERROR).toBe(
      "Email notifications couldn't be enabled. Try again. If the issue persists, contact Run:ai support.",
    );
  });
});
