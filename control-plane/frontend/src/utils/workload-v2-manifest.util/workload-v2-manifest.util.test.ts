/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, it } from "vitest";
import { EWorkloadType } from "@/models/workload.model";
import { workloadV2ManifestUtil } from "./workload-v2-manifest.util";

describe("workloadV2ManifestUtil", () => {
  it("extracts command and arguments from a pod template", () => {
    const spec = workloadV2ManifestUtil.getSpecFromManifest({
      spec: {
        template: {
          spec: {
            containers: [
              {
                command: ["python", "-m", "trainer"],
                args: ["--epochs", "10"],
              },
            ],
          },
        },
      },
    });

    expect(spec).toEqual({ command: "python -m trainer", args: "--epochs 10" });
  });

  it("returns command and arguments without the full spec shape", () => {
    const commandAndArguments = workloadV2ManifestUtil.getCommandAndArgumentsFromManifest({
      spec: {
        template: podTemplateWithCommand(["python"], ["main.py"]),
      },
    });

    expect(commandAndArguments).toEqual({ command: "python", arguments: "main.py" });
  });

  it("extracts command and arguments from the first supported distributed replica type", () => {
    const spec = workloadV2ManifestUtil.getSpecFromManifest({
      spec: {
        pytorchReplicaSpecs: {
          Master: {
            template: podTemplateWithCommand(["python", "-m", "master"], ["--rank", "0"]),
          },
          Worker: {
            template: podTemplateWithCommand(["python", "-m", "worker"], ["--rank", "1"]),
          },
        },
      },
    });

    expect(spec).toEqual({ command: "python -m worker", args: "--rank 1" });
  });

  it("falls back to another supported distributed replica type when Worker is missing", () => {
    const spec = workloadV2ManifestUtil.getSpecFromManifest({
      spec: {
        pytorchReplicaSpecs: {
          Master: {
            template: podTemplateWithCommand(["python", "-m", "master"], ["--rank", "0"]),
          },
        },
      },
    });

    expect(spec).toEqual({ command: "python -m master", args: "--rank 0" });
  });

  it("builds master and worker specs for v2 distributed manifests", () => {
    const crudWorkload = workloadV2ManifestUtil.getCrudWorkloadFromManifest(
      {
        spec: {
          pytorchReplicaSpecs: {
            Master: {
              template: podTemplateWithCommand(["python"], ["--mode", "master"], "master:latest"),
            },
            Worker: {
              template: podTemplateWithCommand(["python"], ["--mode", "worker"], "worker:latest"),
            },
          },
        },
      },
      EWorkloadType.Distributed,
    );

    expect(crudWorkload).toEqual({
      masterSpec: { image: "master:latest", command: "python", args: "--mode master" },
      spec: { image: "worker:latest", command: "python", args: "--mode worker" },
    });
  });

  it("builds a single spec for non-distributed v2 manifests", () => {
    const crudWorkload = workloadV2ManifestUtil.getCrudWorkloadFromManifest({
      spec: {
        template: podTemplateWithCommand(["python"], ["main.py"], "trainer:latest"),
      },
    });

    expect(crudWorkload).toEqual({
      spec: { image: "trainer:latest", command: "python", args: "main.py" },
    });
  });

  it("returns undefined when no command or arguments are present", () => {
    const spec = workloadV2ManifestUtil.getSpecFromManifest({
      spec: {},
    });

    expect(spec).toBeUndefined();
  });
});

function podTemplateWithCommand(command: string[], args: string[], image?: string) {
  return {
    spec: {
      containers: [
        {
          image,
          command,
          args,
        },
      ],
    },
  };
}
