import { describe, it, expect } from "vitest";
import { aiApplicationUtil } from "@/utils/ai-application.util/ai-application.util";
import type { IApplicationTable } from "@/models/ai-application.model";

const makeApp = (overrides: Partial<IApplicationTable> = {}): IApplicationTable => ({
  id: "app-1",
  name: "test-app",
  ...overrides,
});

describe("aiApplicationUtil.getStatusColOptions", () => {
  it("returns success color for Deployed phase", () => {
    const result = aiApplicationUtil.getStatusColOptions(makeApp({ phase: "Deployed" }));
    expect(result.status).toBe("Deployed");
    expect(result.color).toBe("success");
  });

  it("returns negative color for Failed phase", () => {
    const result = aiApplicationUtil.getStatusColOptions(makeApp({ phase: "Failed" }));
    expect(result.status).toBe("Failed");
    expect(result.color).toBe("negative");
  });

  it("returns no color for unknown phase", () => {
    const result = aiApplicationUtil.getStatusColOptions(makeApp({ phase: "Pending" }));
    expect(result.color).toBeUndefined();
  });

  it("normalizes lowercase phase from API before color lookup", () => {
    const result = aiApplicationUtil.getStatusColOptions(makeApp({ phase: "deployed" }));
    expect(result.status).toBe("Deployed");
    expect(result.color).toBe("success");
  });

  it("normalizes lowercase failed phase from API", () => {
    const result = aiApplicationUtil.getStatusColOptions(makeApp({ phase: "failed" }));
    expect(result.status).toBe("Failed");
    expect(result.color).toBe("negative");
  });

  it("falls back to actionStatus when phase is absent", () => {
    const result = aiApplicationUtil.getStatusColOptions(makeApp({ phase: undefined, actionStatus: "Failed" }));
    expect(result.status).toBe("Failed");
    expect(result.color).toBe("negative");
  });

  it("falls back to actionStatus when phase is null", () => {
    const result = aiApplicationUtil.getStatusColOptions(
      makeApp({ phase: null as unknown as undefined, actionStatus: "creating" }),
    );
    expect(result.status).toBe("Creating");
    expect(result.color).toBeUndefined();
  });

  it("falls back to Unknown when both phase and actionStatus are absent", () => {
    const result = aiApplicationUtil.getStatusColOptions(makeApp({ phase: undefined, actionStatus: undefined }));
    expect(result.status).toBe("Unknown");
    expect(result.color).toBeUndefined();
  });

  it("sets statusUpdatedTimeInMs from updatedAt", () => {
    const updatedAt = "2026-06-10T12:00:00Z";
    const result = aiApplicationUtil.getStatusColOptions(makeApp({ phase: "Deployed", updatedAt }));
    expect(result.statusUpdatedTimeInMs).toEqual(new Date(updatedAt));
  });

  it("leaves statusUpdatedTimeInMs undefined when updatedAt is absent", () => {
    const result = aiApplicationUtil.getStatusColOptions(makeApp({ phase: "Deployed", updatedAt: undefined }));
    expect(result.statusUpdatedTimeInMs).toBeUndefined();
  });

  it("sets tooltipText from actionMessage", () => {
    const result = aiApplicationUtil.getStatusColOptions(makeApp({ phase: "Failed", actionMessage: "Chart not found" }));
    expect(result.tooltipText).toBe("Chart not found");
  });

  it("leaves tooltipText undefined when actionMessage is absent", () => {
    const result = aiApplicationUtil.getStatusColOptions(makeApp({ phase: "Deployed", actionMessage: undefined }));
    expect(result.tooltipText).toBeUndefined();
  });

  it("always sets displayAnimation to false", () => {
    const result = aiApplicationUtil.getStatusColOptions(makeApp({ phase: "Deployed" }));
    expect(result.displayAnimation).toBe(false);
  });
});
