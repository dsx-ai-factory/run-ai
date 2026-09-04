/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

//@ts-nocheck
import { describe, it, expect, vi, beforeEach } from "vitest";
import { identityManagerApi } from "@/services/infra/client-apis/identity-manager-api/identity-manager-api";
import { httpService } from "@/services/infra/https.service/http.service";
import { accessKeysService } from "@/services/control-plane/access-keys.service/access-keys.service";

// Mock identityManagerApi and httpService
vi.mock("@/services/infra/client-apis/identity-manager-api/identity-manager-api", () => ({
  identityManagerApi: {
    accessKeysApi: {
      getAccessKeys: vi.fn(),
      getAccessKeyById: vi.fn(),
      createAccessKey: vi.fn(),
      deleteAccessKeyById: vi.fn(),
      regenerateAccessKeySecret: vi.fn(),
    },
  },
}));
vi.mock("@/services/infra/https.service/http.service", () => ({
  httpService: {
    handleHttpError: vi.fn(),
  },
}));

describe("accessKeysService", () => {
  const mockError = new Error("Test Error");

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("getAccessKeys", () => {
    it("should return the list of access keys", async () => {
      identityManagerApi.accessKeysApi.getAccessKeys.mockResolvedValue({
        data: {
          accessKeys: [
            { name: "key1", id: "1" },
            { name: "key2", id: "2" },
            { name: "key3", id: "3" },
          ],
        },
      });

      const result = await accessKeysService.getAccessKeys();
      expect(result).toStrictEqual([
        { name: "key1", id: "1" },
        { name: "key2", id: "2" },
        { name: "key3", id: "3" },
      ]);
    });

    it("should handle errors", async () => {
      identityManagerApi.accessKeysApi.getAccessKeys.mockRejectedValue(mockError);
      httpService.handleHttpError.mockImplementation((err) => {
        throw err;
      });

      await expect(accessKeysService.getAccessKeys()).rejects.toThrow(mockError);
    });
  });

  describe("getAccessKey", () => {
    it("should return the access key data", async () => {
      identityManagerApi.accessKeysApi.getAccessKeyById.mockResolvedValue({
        data: { id: "123", name: "Test Key" },
      });

      const result = await accessKeysService.getAccessKey("123");
      expect(result).toStrictEqual({ id: "123", name: "Test Key" });
    });

    it("should handle errors", async () => {
      identityManagerApi.accessKeysApi.getAccessKeyById.mockRejectedValue(mockError);
      httpService.handleHttpError.mockImplementation((err) => {
        throw err;
      });

      await expect(accessKeysService.getAccessKey("123")).rejects.toThrow(mockError);
    });
  });

  describe("createAccessKey", () => {
    it("should create a new access key and return the response", async () => {
      const keyCreationRequest = { name: "New Key" };
      const keyPostResponse = { id: "123", name: "New Key" };
      identityManagerApi.accessKeysApi.createAccessKey.mockResolvedValue({ data: keyPostResponse });

      const result = await accessKeysService.createAccessKey(keyCreationRequest);
      expect(result).toEqual(keyPostResponse);
    });

    it("should handle errors", async () => {
      identityManagerApi.accessKeysApi.createAccessKey.mockRejectedValue(mockError);
      httpService.handleHttpError.mockImplementation((err) => {
        throw err;
      });

      await expect(accessKeysService.createAccessKey({ name: "New Key" })).rejects.toThrow(mockError);
    });
  });

  describe("removeAccessKey", () => {
    it("should remove an access key", async () => {
      identityManagerApi.accessKeysApi.deleteAccessKeyById.mockResolvedValue({});

      await expect(accessKeysService.removeAccessKey("123")).resolves.toBeUndefined();
    });

    it("should handle errors", async () => {
      identityManagerApi.accessKeysApi.deleteAccessKeyById.mockRejectedValue(mockError);
      httpService.handleHttpError.mockImplementation((err) => {
        throw err;
      });

      await expect(accessKeysService.removeAccessKey("123")).rejects.toThrow(mockError);
    });
  });

  describe("generateAccessKeySecret", () => {
    it("should regenerate and return a new access key secret", async () => {
      const secretResponse = { secret: "new-secret" };
      identityManagerApi.accessKeysApi.regenerateAccessKeySecret.mockResolvedValue({
        data: secretResponse,
      });

      const result = await accessKeysService.generateAccessKeySecret("123");
      expect(result).toEqual(secretResponse);
    });

    it("should handle errors", async () => {
      identityManagerApi.accessKeysApi.regenerateAccessKeySecret.mockRejectedValue(mockError);
      httpService.handleHttpError.mockImplementation((err) => {
        throw err;
      });

      await expect(accessKeysService.generateAccessKeySecret("123")).rejects.toThrow(mockError);
    });
  });
});
