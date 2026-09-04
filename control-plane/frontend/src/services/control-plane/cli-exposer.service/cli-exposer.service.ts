/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { cliApi } from "@/services/infra/client-apis/cli-exposer-api/cli-exposer-api";
import { httpService } from "@/services/infra/https.service/http.service";
import type { Command } from "@/swagger-models/cli-exposer-client";
import fileDownload from "js-file-download";

export const cliService = {
  getInstallerUnix,
  getInstallerWindowsCommands,
  getUIWindowsCommand,
  getUILinuxCommand,
  getUIMacCommand,
  getUIUnixInstallerDownload,
  getBinary,
  getBinaryDownload,
  getBinaryWithProgress,
};

async function getInstallerUnix(): Promise<string> {
  try {
    const response = await cliApi.cliApi.getInstallerUnix();
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}
async function getInstallerLinuxCommands(): Promise<Command> {
  try {
    const response = await cliApi.cliApi.getInstallerLinuxCommands();
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getInstallerMacCommands(): Promise<Command> {
  try {
    const response = await cliApi.cliApi.getInstallerMacCommands();
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getInstallerWindowsCommands(): Promise<Command> {
  try {
    const response = await cliApi.cliApi.getInstallerWindowsCommands();
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getUIWindowsCommand(): Promise<string> {
  return (await getInstallerWindowsCommands()).command;
}
async function getUILinuxCommand(): Promise<string> {
  return (await getInstallerLinuxCommands()).command;
}
async function getUIMacCommand(): Promise<string> {
  return (await getInstallerMacCommands()).command;
}

async function getUIUnixInstallerDownload(): Promise<void> {
  const content = await getInstallerUnix();
  const blob = new Blob([content], { type: "application/octet-stream" });

  fileDownload(blob, `runai-cli-unix`);
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
async function getBinary(os: EOperatingSystems, architecture: EArchitectureType): Promise<any> {
  try {
    const response = await cliApi.cliApi.getBinary(os, architecture);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getBinaryWithProgress(
  os: EOperatingSystems,
  architecture: EArchitectureType,
  progress: Record<"value", number>,
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
): Promise<any> {
  try {
    const response = await cliApi.cliApi.getBinary(os, architecture, {
      responseType: "blob",
      onDownloadProgress: (event) => {
        if (event.total && event.total > 0) {
          progress.value = Math.round((event.loaded * 100) / event.total);
        }
      },
    });
    const blob = new Blob([response.data], { type: "application/octet-stream" });

    fileDownload(blob, `runai-cli-${architecture}`);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

export enum EArchitectureType {
  AMD = "amd64",
  ARM = "arm64",
}

export enum EOperatingSystems {
  WINDOWS = "windows",
  MAC = "darwin",
  LINUX = "linux",
}

async function getBinaryDownload(os: EOperatingSystems, architecture: EArchitectureType): Promise<void> {
  const content = await getBinary(os, architecture);
  const blob = new Blob([content], { type: "application/octet-stream" });

  fileDownload(blob, `runai-cli-unix`);
}
