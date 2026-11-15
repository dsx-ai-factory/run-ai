/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect } from "vitest";
import { workloadDetailsUtil, type WorkloadSpec } from "@/utils/workload-details.util/workload-details.util";
import type { WorkloadDetailed } from "@/swagger-models/workloads-client";
import { formatK8sResourceMemory } from "@/utils/format.util/format.util";
import { LINE_BREAK } from "@/utils/workload-hybrid.util/workload-hybrid.util";

describe("workloadDetailsUtil.parseComputeDetails", () => {
  it("includes allocation rows when allocatedResources is present", () => {
    const workloadDetailed = {
      allocatedResources: {
        gpu: 4,
        gpuMemory: "536870912",
        cpu: 1,
        cpuMemory: "104857600",
      },
    } as unknown as WorkloadDetailed;

    const content = workloadDetailsUtil.parseComputeDetails(workloadDetailed)[0]?.content ?? [];

    expect(content).toEqual(
      expect.arrayContaining([
        "GPU compute allocation: 4",
        `GPU memory allocation: ${formatK8sResourceMemory("536870912")}`,
        `CPU memory allocation: ${formatK8sResourceMemory("104857600")}`,
        expect.stringMatching(/^CPU compute allocation: /),
      ]),
    );
  });

  it("omits allocation rows when allocatedResources is absent", () => {
    const workloadDetailed = {} as unknown as WorkloadDetailed;

    const content = workloadDetailsUtil.parseComputeDetails(workloadDetailed)[0]?.content ?? [];

    expect(content.some((line) => line.includes("allocation"))).toBe(false);
  });

  it("omits individual allocation rows whose value is missing", () => {
    const workloadDetailed = {
      allocatedResources: { gpu: 0, cpu: 2 },
    } as unknown as WorkloadDetailed;

    const content = workloadDetailsUtil.parseComputeDetails(workloadDetailed)[0]?.content ?? [];

    expect(content).toContain("GPU compute allocation: 0");
    expect(content.some((line) => line.startsWith("CPU compute allocation"))).toBe(true);
    expect(content.some((line) => line.startsWith("GPU memory allocation"))).toBe(false);
    expect(content.some((line) => line.startsWith("CPU memory allocation"))).toBe(false);
  });
});

describe("workloadDetailsUtil.parseNetworkDetails", () => {
  const servingUrl = "https://i100-runai-p100.inference.example.com";
  const servingPortSpec = { servingPort: { container: 8000, authorizationType: "public" } } as unknown as WorkloadSpec;

  it("falls back to the workload URL when the ServingPort connection has no URL", () => {
    const workloadDetailed = {
      urls: [servingUrl],
      environments: [
        {
          connections: [{ name: "ServingPort", toolType: "custom", connectionType: "ServingPort", containerPort: 8000 }],
        },
      ],
    } as unknown as WorkloadDetailed;

    const content = workloadDetailsUtil.parseNetworkDetails(workloadDetailed, servingPortSpec)[0]?.content ?? [];

    expect(content).toContain(`URL: ${servingUrl}`);
  });

  it("falls back to the workload URL when no spec networking is available", () => {
    const workloadDetailed = {
      urls: [servingUrl],
      environments: [
        {
          connections: [{ name: "ServingPort", toolType: "custom", connectionType: "ServingPort", containerPort: 8000 }],
        },
      ],
    } as unknown as WorkloadDetailed;

    const content = workloadDetailsUtil.parseNetworkDetails(workloadDetailed)[0]?.content ?? [];

    expect(content).toContain(`URL: ${servingUrl}`);
  });

  it("keeps the URL reported on the ServingPort connection", () => {
    const reportedUrl = "https://reported.example.com";
    const workloadDetailed = {
      urls: [servingUrl],
      environments: [
        {
          connections: [
            {
              name: "ServingPort",
              toolType: "custom",
              connectionType: "ServingPort",
              containerPort: 8000,
              url: reportedUrl,
            },
          ],
        },
      ],
    } as unknown as WorkloadDetailed;

    const content = workloadDetailsUtil.parseNetworkDetails(workloadDetailed, servingPortSpec)[0]?.content ?? [];

    expect(content).toContain(`URL: ${reportedUrl}`);
    expect(content).not.toContain(`URL: ${servingUrl}`);
  });

  describe("with a tool URL next to the serving URL", () => {
    const toolUrl = "https://x/tool";
    const serveUrl = "https://x/serve";
    const toolSpec = { container: 3000, toolType: "chatbot-ui", toolName: "ChatbotUI" };

    const getRows = (content: string[]): string[][] =>
      content.reduce<string[][]>(
        (rows, line) => {
          if (line === LINE_BREAK) rows.push([]);
          else rows[rows.length - 1]?.push(line);
          return rows;
        },
        [[]],
      );
    const getRow = (content: string[], name: string): string[] =>
      getRows(content).find((row) => row.includes(`Name: ${name}`)) ?? [];

    const buildWorkload = (connections: unknown[], urls?: string[]): WorkloadDetailed =>
      ({ urls, environments: [{ connections }] }) as unknown as WorkloadDetailed;
    const toolConnection = (url?: string) => ({
      name: "ChatbotUI",
      toolType: "chatbot-ui",
      connectionType: "ExternalUrl",
      containerPort: 3000,
      url,
    });
    const servingConnection = {
      name: "ServingPort",
      toolType: "custom",
      connectionType: "ServingPort",
      containerPort: 8000,
    };

    it("uses the URL the tool connection does not claim", () => {
      const spec = {
        exposedUrls: [toolSpec],
        servingPort: { container: 8000, authorizationType: "public" },
      } as unknown as WorkloadSpec;
      const workloadDetailed = buildWorkload([toolConnection(toolUrl), servingConnection], [toolUrl, serveUrl]);

      const content = workloadDetailsUtil.parseNetworkDetails(workloadDetailed, spec)[0]?.content ?? [];

      expect(getRow(content, "ServingPort")).toContain(`URL: ${serveUrl}`);
      expect(getRow(content, "ChatbotUI")).toContain(`URL: ${toolUrl}`);
    });

    it("treats the spec exposed URL as claimed when the tool URL is not reported yet", () => {
      const spec = {
        exposedUrls: [{ ...toolSpec, url: toolUrl }],
        servingPort: { container: 8000, authorizationType: "public" },
      } as unknown as WorkloadSpec;
      const workloadDetailed = buildWorkload([toolConnection(), servingConnection], [toolUrl, serveUrl]);

      const content = workloadDetailsUtil.parseNetworkDetails(workloadDetailed, spec)[0]?.content ?? [];

      expect(getRow(content, "ServingPort")).toContain(`URL: ${serveUrl}`);
    });

    it("does not guess when several URLs are unclaimed", () => {
      const spec = {
        exposedUrls: [toolSpec],
        servingPort: { container: 8000, authorizationType: "public" },
      } as unknown as WorkloadSpec;
      const workloadDetailed = buildWorkload([toolConnection(), servingConnection], [toolUrl, serveUrl]);

      const content = workloadDetailsUtil.parseNetworkDetails(workloadDetailed, spec)[0]?.content ?? [];

      expect(getRow(content, "ServingPort").some((line) => line.startsWith("URL: "))).toBe(false);
      expect(content).not.toContain(`URL: ${toolUrl}`);
    });

    it("gives the serving URL to the ServingPort connection, not a NodePort on the same port", () => {
      const spec = {
        ports: [{ container: 8000, toolType: "custom", toolName: "NodePort" }],
        servingPort: { container: 8000, authorizationType: "public" },
      } as unknown as WorkloadSpec;
      const nodePortConnection = {
        name: "NodePort",
        toolType: "custom",
        connectionType: "NodePort",
        containerPort: 8000,
        url: "http://10.0.0.1:30080",
      };
      const workloadDetailed = buildWorkload([nodePortConnection, servingConnection], [serveUrl]);

      const content = workloadDetailsUtil.parseNetworkDetails(workloadDetailed, spec)[0]?.content ?? [];

      expect(getRow(content, "ServingPort")).toContain(`URL: ${serveUrl}`);
      expect(getRow(content, "NodePort")).toContain("URL: http://10.0.0.1:30080");
    });
  });

  it("shows no URL when the workload has no URLs", () => {
    const workloadDetailed = {
      environments: [
        {
          connections: [{ name: "ServingPort", toolType: "custom", connectionType: "ServingPort", containerPort: 8000 }],
        },
      ],
    } as unknown as WorkloadDetailed;

    const content = workloadDetailsUtil.parseNetworkDetails(workloadDetailed, servingPortSpec)[0]?.content ?? [];

    expect(content.some((line) => line.startsWith("URL: "))).toBe(false);
  });

  it("shows no URL when the serving port is not exposed externally", () => {
    const spec = { servingPort: { container: 8000, exposeExternally: false } } as unknown as WorkloadSpec;
    const workloadDetailed = {
      urls: [servingUrl],
      environments: [
        {
          connections: [{ name: "ServingPort", toolType: "custom", connectionType: "ServingPort", containerPort: 8000 }],
        },
      ],
    } as unknown as WorkloadDetailed;

    const content = workloadDetailsUtil.parseNetworkDetails(workloadDetailed, spec)[0]?.content ?? [];

    expect(content.some((line) => line.startsWith("URL: "))).toBe(false);
  });
});
