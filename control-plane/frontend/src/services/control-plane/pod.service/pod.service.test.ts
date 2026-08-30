// @ts-nocheck
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import { podService } from "@/services/control-plane/pod.service/pod.service";
import { controlPlaneService } from "../control-plane.service/control-plane.service";

describe("pod.service", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it("Get pods and without deployment", async () => {
    controlPlaneService.get = vi.fn();
    await podService.getPods("jobId1", "1234");
    expect(controlPlaneService.get).toBeCalledWith("v1/k8s/clusters/1234/jobs/jobId1/pods");
  });

  it("Get pods and with deployment", async () => {
    controlPlaneService.get = vi.fn();
    await podService.getPods("jobId1", "1234", "5678");
    expect(controlPlaneService.get).toBeCalledWith("v1/k8s/clusters/1234/deployments/5678/pods");
  });
});
