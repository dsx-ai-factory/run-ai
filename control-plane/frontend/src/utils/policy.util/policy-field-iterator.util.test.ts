import { describe, it, expect } from "vitest";
import {
  getValueByPath,
  deleteValueByPath,
  clearWorkloadBlockedFields,
  isFieldBlocked,
  getBlockedFields,
  getBlockedFieldPaths,
  type IBlockableRule,
  type TPolicyWithRules,
} from "./policy-field-iterator.util";
import type { IUISuperWorkloadSpec } from "@/models/workload.model";

describe("policy-field-iterator.util", () => {
  // =========================================================================
  // Path Utilities Tests
  // =========================================================================
  describe("getValueByPath", () => {
    it("should get a top-level value", () => {
      const obj = { name: "test", value: 42 };
      expect(getValueByPath(obj, "name")).toBe("test");
      expect(getValueByPath(obj, "value")).toBe(42);
    });

    it("should get a nested value", () => {
      const obj = { level1: { level2: { level3: "deep" } } };
      expect(getValueByPath(obj, "level1.level2.level3")).toBe("deep");
      expect(getValueByPath(obj, "level1.level2")).toEqual({ level3: "deep" });
    });

    it("should return undefined for non-existent paths", () => {
      const obj = { name: "test" };
      expect(getValueByPath(obj, "missing")).toBeUndefined();
      expect(getValueByPath(obj, "deep.missing.path")).toBeUndefined();
    });

    it("should handle null and undefined objects", () => {
      expect(getValueByPath(null, "path")).toBeUndefined();
      expect(getValueByPath(undefined, "path")).toBeUndefined();
    });

    it("should handle empty path", () => {
      const obj = { name: "test" };
      expect(getValueByPath(obj, "")).toBeUndefined();
    });
  });

  describe("deleteValueByPath", () => {
    it("should delete a top-level value", () => {
      const obj: Record<string, unknown> = { name: "test", value: 42 };
      deleteValueByPath(obj, "name");
      expect(obj.name).toBeUndefined();
      expect(obj.value).toBe(42);
    });

    it("should delete a nested value", () => {
      const obj: Record<string, unknown> = { level1: { level2: { value: "test" } } };
      deleteValueByPath(obj, "level1.level2.value");
      expect((obj.level1 as Record<string, unknown>).level2).toEqual({});
    });

    it("should handle non-existent paths gracefully", () => {
      const obj: Record<string, unknown> = { name: "test" };
      deleteValueByPath(obj, "missing.deep.path");
      expect(obj).toEqual({ name: "test" });
    });
  });

  // =========================================================================
  // Simple Field Blocking Tests
  // =========================================================================
  describe("clearWorkloadBlockedFields - simple fields", () => {
    it("should clear blocked top-level fields", () => {
      const spec = {
        command: "echo hello",
        args: "--verbose",
        image: "nginx:latest",
      } as IUISuperWorkloadSpec;

      const policy = {
        rules: {
          command: { sourceOfRule: {}, blocked: true } as IBlockableRule,
          args: { sourceOfRule: {}, blocked: false } as IBlockableRule,
          image: { sourceOfRule: {}, blocked: null } as IBlockableRule,
        },
      } as TPolicyWithRules;

      const result = clearWorkloadBlockedFields(spec, policy);

      expect(result.command).toBeUndefined();
      expect(result.args).toBe("--verbose");
      expect(result.image).toBe("nginx:latest");
    });

    it("should clear blocked nested fields in compute", () => {
      const spec = {
        compute: {
          cpuCoreRequest: 2,
          cpuCoreLimit: 4,
          gpuDevicesRequest: 1,
        },
      } as IUISuperWorkloadSpec;

      const policy = {
        rules: {
          compute: {
            cpuCoreRequest: { sourceOfRule: {}, blocked: true } as IBlockableRule,
            cpuCoreLimit: { sourceOfRule: {}, blocked: false } as IBlockableRule,
            gpuDevicesRequest: { sourceOfRule: {}, blocked: true } as IBlockableRule,
          },
        },
      } as TPolicyWithRules;

      const result = clearWorkloadBlockedFields(spec, policy);

      expect(result.compute?.cpuCoreRequest).toBeUndefined();
      expect(result.compute?.cpuCoreLimit).toBe(4);
      expect(result.compute?.gpuDevicesRequest).toBeUndefined();
    });

    it("should clear blocked nested fields in security", () => {
      const spec = {
        security: {
          runAsUid: 1000,
          runAsGid: 1000,
          uidGidSource: "fromIdpToken",
        },
      } as IUISuperWorkloadSpec;

      const policy = {
        rules: {
          security: {
            runAsUid: { sourceOfRule: {}, blocked: true } as IBlockableRule,
            runAsGid: { sourceOfRule: {}, blocked: false } as IBlockableRule,
            uidGidSource: { sourceOfRule: {}, blocked: true } as IBlockableRule,
          },
        },
      } as TPolicyWithRules;

      const result = clearWorkloadBlockedFields(spec, policy);

      expect(result.security?.runAsUid).toBeUndefined();
      expect(result.security?.runAsGid).toBe(1000);
      expect(result.security?.uidGidSource).toBeUndefined();
    });

    it("should handle null policy gracefully", () => {
      const spec = {
        command: "echo hello",
      } as IUISuperWorkloadSpec;

      const result = clearWorkloadBlockedFields(spec, null);

      expect(result.command).toBe("echo hello");
    });

    it("should handle policy with no rules", () => {
      const spec = {
        command: "echo hello",
      } as IUISuperWorkloadSpec;

      const policy = {} as TPolicyWithRules;

      const result = clearWorkloadBlockedFields(spec, policy);

      expect(result.command).toBe("echo hello");
    });
  });

  // =========================================================================
  // Array Item Attribute Blocking Tests (the main new functionality)
  // =========================================================================
  describe("clearWorkloadBlockedFields - array item attributes", () => {
    it("should clear blocked attributes from tolerations array items", () => {
      const spec = {
        tolerations: [
          { key: "gpu", value: "true", operator: "Equal", effect: "NoSchedule" },
          { key: "memory", value: "high", operator: "Equal", effect: "NoExecute" },
        ],
      } as IUISuperWorkloadSpec;

      const policy = {
        rules: {
          tolerations: {
            attributes: {
              key: { sourceOfRule: {}, blocked: true } as IBlockableRule,
              value: { sourceOfRule: {}, blocked: false } as IBlockableRule,
            },
          },
        },
      } as TPolicyWithRules;

      const result = clearWorkloadBlockedFields(spec, policy);

      // key should be cleared from all items
      expect(result.tolerations?.[0]?.key).toBeUndefined();
      expect(result.tolerations?.[1]?.key).toBeUndefined();
      // value should remain
      expect(result.tolerations?.[0]?.value).toBe("true");
      expect(result.tolerations?.[1]?.value).toBe("high");
      // other fields should remain
      expect(result.tolerations?.[0]?.operator).toBe("Equal");
      expect(result.tolerations?.[1]?.effect).toBe("NoExecute");
    });

    it("should clear blocked attributes from annotations array items", () => {
      const spec = {
        annotations: [
          { name: "app", value: "myapp" },
          { name: "version", value: "1.0" },
        ],
      } as IUISuperWorkloadSpec;

      const policy = {
        rules: {
          annotations: {
            attributes: {
              value: { sourceOfRule: {}, blocked: true } as IBlockableRule,
            },
          },
        },
      } as TPolicyWithRules;

      const result = clearWorkloadBlockedFields(spec, policy);

      expect(result.annotations?.[0]?.name).toBe("app");
      expect(result.annotations?.[0]?.value).toBeUndefined();
      expect(result.annotations?.[1]?.name).toBe("version");
      expect(result.annotations?.[1]?.value).toBeUndefined();
    });

    it("should handle empty arrays", () => {
      const spec = {
        tolerations: [],
      } as IUISuperWorkloadSpec;

      const policy = {
        rules: {
          tolerations: {
            attributes: {
              key: { sourceOfRule: {}, blocked: true } as IBlockableRule,
            },
          },
        },
      } as TPolicyWithRules;

      const result = clearWorkloadBlockedFields(spec, policy);

      expect(result.tolerations).toEqual([]);
    });

    it("should handle missing array field", () => {
      const spec = {} as IUISuperWorkloadSpec;

      const policy = {
        rules: {
          tolerations: {
            attributes: {
              key: { sourceOfRule: {}, blocked: true } as IBlockableRule,
            },
          },
        },
      } as TPolicyWithRules;

      // Should not throw
      const result = clearWorkloadBlockedFields(spec, policy);
      expect(result.tolerations).toBeUndefined();
    });

    it("should clear blocked storage attributes", () => {
      const spec = {
        storage: {
          pvc: [{ claimName: "data-pvc", path: "/data" }],
          nfs: [{ server: "nfs.example.com", mountPath: "/shared" }],
        },
      } as unknown as IUISuperWorkloadSpec;

      const policy = {
        rules: {
          storage: {
            pvc: {
              attributes: {
                path: { sourceOfRule: {}, blocked: true } as IBlockableRule,
              },
            },
            nfs: {
              attributes: {
                mountPath: { sourceOfRule: {}, blocked: true } as IBlockableRule,
              },
            },
          },
        },
      } as TPolicyWithRules;

      const result = clearWorkloadBlockedFields(spec, policy);

      expect(result.storage?.pvc?.[0]?.path).toBeUndefined();
      expect(result.storage?.pvc?.[0]?.claimName).toBe("data-pvc");
      expect(result.storage?.nfs?.[0]?.mountPath).toBeUndefined();
      expect(result.storage?.nfs?.[0]?.server).toBe("nfs.example.com");
    });

    it("should also clear storage.volumes when storage.pvc attributes are blocked (UI path mapping)", () => {
      const spec = {
        storage: {
          pvc: [{ claimName: "existing-claim", path: "/data" }],
          volumes: [{ claimName: "new-volume", path: "/new-data" }],
        },
      } as unknown as IUISuperWorkloadSpec;

      const policy = {
        rules: {
          storage: {
            pvc: {
              attributes: {
                path: { sourceOfRule: {}, blocked: true } as IBlockableRule,
              },
            },
          },
        },
      } as TPolicyWithRules;

      const result = clearWorkloadBlockedFields(spec, policy);

      expect(result.storage?.pvc?.[0]?.path).toBeUndefined();
      expect(result.storage?.pvc?.[0]?.claimName).toBe("existing-claim");
      expect(result.storage?.volumes?.[0]?.path).toBeUndefined();
      expect(result.storage?.volumes?.[0]?.claimName).toBe("new-volume");
    });
  });

  // =========================================================================
  // Mixed Blocking Tests
  // =========================================================================
  describe("clearWorkloadBlockedFields - mixed fields", () => {
    it("should handle combination of simple and array attribute blocked fields", () => {
      const spec = {
        command: "echo hello",
        image: "nginx",
        tolerations: [{ key: "gpu", value: "true" }],
        compute: {
          cpuCoreRequest: 2,
        },
      } as IUISuperWorkloadSpec;

      const policy = {
        rules: {
          command: { sourceOfRule: {}, blocked: true } as IBlockableRule,
          image: { sourceOfRule: {}, blocked: false } as IBlockableRule,
          tolerations: {
            attributes: {
              key: { sourceOfRule: {}, blocked: true } as IBlockableRule,
            },
          },
          compute: {
            cpuCoreRequest: { sourceOfRule: {}, blocked: true } as IBlockableRule,
          },
        },
      } as TPolicyWithRules;

      const result = clearWorkloadBlockedFields(spec, policy);

      expect(result.command).toBeUndefined();
      expect(result.image).toBe("nginx");
      expect(result.tolerations?.[0]?.key).toBeUndefined();
      expect(result.tolerations?.[0]?.value).toBe("true");
      expect(result.compute?.cpuCoreRequest).toBeUndefined();
    });
  });

  // =========================================================================
  // Utility Function Tests
  // =========================================================================
  describe("isFieldBlocked", () => {
    it("should return true for blocked fields", () => {
      const policy = {
        rules: {
          command: { sourceOfRule: {}, blocked: true },
        },
      } as TPolicyWithRules;

      expect(isFieldBlocked(policy, "command")).toBe(true);
    });

    it("should return false for non-blocked fields", () => {
      const policy = {
        rules: {
          command: { sourceOfRule: {}, blocked: false },
        },
      } as TPolicyWithRules;

      expect(isFieldBlocked(policy, "command")).toBe(false);
    });

    it("should return false for missing fields", () => {
      const policy = {
        rules: {},
      } as TPolicyWithRules;

      expect(isFieldBlocked(policy, "command")).toBe(false);
    });

    it("should return false for null policy", () => {
      expect(isFieldBlocked(null, "command")).toBe(false);
    });

    it("should work with nested paths", () => {
      const policy = {
        rules: {
          compute: {
            cpuCoreRequest: { sourceOfRule: {}, blocked: true },
          },
        },
      } as TPolicyWithRules;

      expect(isFieldBlocked(policy, "compute.cpuCoreRequest")).toBe(true);
    });
  });

  describe("getBlockedFields", () => {
    it("should return all blocked fields including nested and attributes", () => {
      const policy = {
        rules: {
          command: { sourceOfRule: {}, blocked: true } as IBlockableRule,
          args: { sourceOfRule: {}, blocked: false } as IBlockableRule,
          compute: {
            cpuCoreRequest: { sourceOfRule: {}, blocked: true } as IBlockableRule,
          },
          tolerations: {
            attributes: {
              key: { sourceOfRule: {}, blocked: true } as IBlockableRule,
            },
          },
        },
      } as TPolicyWithRules;

      const blockedFields = getBlockedFields(policy);

      expect(blockedFields.length).toBe(3);

      const paths = blockedFields.map((f) => f.policyPath);
      expect(paths).toContain("command");
      expect(paths).toContain("compute.cpuCoreRequest");
      expect(paths).toContain("tolerations.key");
    });

    it("should correctly identify array item attributes", () => {
      const policy = {
        rules: {
          tolerations: {
            attributes: {
              key: { sourceOfRule: {}, blocked: true } as IBlockableRule,
            },
          },
        },
      } as TPolicyWithRules;

      const blockedFields = getBlockedFields(policy);

      expect(blockedFields.length).toBe(1);
      expect(blockedFields[0].isArrayItemAttribute).toBe(true);
      expect(blockedFields[0].arrayFieldName).toBe("tolerations");
      expect(blockedFields[0].attributePath).toBe("key");
      expect(blockedFields[0].modelPath).toBe("tolerations[*].key");
    });
  });

  describe("getBlockedFieldPaths", () => {
    it("should return just the policy paths", () => {
      const policy = {
        rules: {
          command: { sourceOfRule: {}, blocked: true } as IBlockableRule,
          compute: {
            cpuCoreRequest: { sourceOfRule: {}, blocked: true } as IBlockableRule,
          },
        },
      } as TPolicyWithRules;

      const paths = getBlockedFieldPaths(policy);

      expect(paths).toContain("command");
      expect(paths).toContain("compute.cpuCoreRequest");
      expect(paths.length).toBe(2);
    });

    it("should return empty array for null policy", () => {
      expect(getBlockedFieldPaths(null)).toEqual([]);
    });
  });
});
