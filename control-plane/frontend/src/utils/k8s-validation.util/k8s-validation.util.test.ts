/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, it } from "vitest";
import { isExtendedResourceName, isValidExtendedResourceQuantity, isValidK8sKey } from "./k8s-validation.util";

describe("k8s-validation.util", () => {
  describe("isExtendedResourceName", () => {
    describe("valid hugepages resources", () => {
      it("should accept hugepages-2Mi", () => {
        expect(isExtendedResourceName("hugepages-2Mi")).toBe(true);
      });

      it("should accept hugepages-1Gi", () => {
        expect(isExtendedResourceName("hugepages-1Gi")).toBe(true);
      });

      it("should accept hugepages with Ki suffix", () => {
        expect(isExtendedResourceName("hugepages-1024Ki")).toBe(true);
      });

      it("should accept hugepages with Ti suffix", () => {
        expect(isExtendedResourceName("hugepages-1Ti")).toBe(true);
      });

      it("should accept hugepages with Pi suffix", () => {
        expect(isExtendedResourceName("hugepages-2Pi")).toBe(true);
      });

      it("should accept hugepages with Ei suffix", () => {
        expect(isExtendedResourceName("hugepages-4Ei")).toBe(true);
      });

      it("should accept hugepages with multi-digit sizes", () => {
        expect(isExtendedResourceName("hugepages-2048Mi")).toBe(true);
        expect(isExtendedResourceName("hugepages-1024Gi")).toBe(true);
      });
    });

    describe("invalid hugepages resources", () => {
      it("should reject hugepages without size", () => {
        expect(isExtendedResourceName("hugepages-")).toBe(false);
      });

      it("should reject hugepages with invalid unit", () => {
        expect(isExtendedResourceName("hugepages-2Mi2")).toBe(false);
        expect(isExtendedResourceName("hugepages-2GB")).toBe(false);
        expect(isExtendedResourceName("hugepages-2Mb")).toBe(false);
      });

      it("should reject hugepages with lowercase units", () => {
        expect(isExtendedResourceName("hugepages-2mi")).toBe(false);
        expect(isExtendedResourceName("hugepages-1gi")).toBe(false);
      });

      it("should reject hugepages with decimal sizes", () => {
        expect(isExtendedResourceName("hugepages-1.5Mi")).toBe(false);
        expect(isExtendedResourceName("hugepages-0.5Gi")).toBe(false);
      });

      it("should reject hugepages with no numeric value", () => {
        expect(isExtendedResourceName("hugepages-Mi")).toBe(false);
        expect(isExtendedResourceName("hugepages-Gi")).toBe(false);
      });
    });

    describe("valid qualified names (subdomain/resource-name)", () => {
      it("should accept standard subdomain/resource-name format", () => {
        expect(isExtendedResourceName("example.com/foo")).toBe(true);
        expect(isExtendedResourceName("nvidia.com/gpu")).toBe(true);
      });

      it("should accept multi-level subdomains", () => {
        expect(isExtendedResourceName("sub.example.com/resource")).toBe(true);
        expect(isExtendedResourceName("a.b.c.example.com/resource")).toBe(true);
      });

      it("should accept resource names with hyphens and underscores", () => {
        expect(isExtendedResourceName("example.com/my-resource")).toBe(true);
        expect(isExtendedResourceName("example.com/my_resource")).toBe(true);
        expect(isExtendedResourceName("example.com/my-resource_name")).toBe(true);
      });

      it("should accept resource names with dots", () => {
        expect(isExtendedResourceName("example.com/my.resource")).toBe(true);
      });

      it("should accept AWS EFA resource", () => {
        expect(isExtendedResourceName("vpc.amazonaws.com/efa")).toBe(true);
      });
    });

    describe("invalid qualified names", () => {
      it("should reject names with uppercase letters in subdomain", () => {
        expect(isExtendedResourceName("Example.com/resource")).toBe(false);
        expect(isExtendedResourceName("EXAMPLE.COM/resource")).toBe(false);
      });

      it("should reject names with invalid characters in subdomain", () => {
        expect(isExtendedResourceName("exam_ple.com/resource")).toBe(false);
        expect(isExtendedResourceName("exam ple.com/resource")).toBe(false);
      });

      it("should reject names starting or ending with hyphens in subdomain", () => {
        expect(isExtendedResourceName("-example.com/resource")).toBe(false);
        expect(isExtendedResourceName("example-.com/resource")).toBe(false);
      });

      it("should reject names with empty subdomain or resource name", () => {
        expect(isExtendedResourceName("/resource")).toBe(false);
        expect(isExtendedResourceName("example.com/")).toBe(false);
      });

      it("should reject names with multiple slashes", () => {
        expect(isExtendedResourceName("example.com/sub/resource")).toBe(false);
      });
    });

    describe("native resources (should be rejected)", () => {
      it("should reject standard Kubernetes native resources without prefix", () => {
        expect(isExtendedResourceName("cpu")).toBe(false);
        expect(isExtendedResourceName("memory")).toBe(false);
        expect(isExtendedResourceName("storage")).toBe(false);
        expect(isExtendedResourceName("ephemeral-storage")).toBe(false);
      });

      it("should reject resources with requests. prefix", () => {
        expect(isExtendedResourceName("requests.cpu")).toBe(false);
        expect(isExtendedResourceName("requests.memory")).toBe(false);
        expect(isExtendedResourceName("requests.storage")).toBe(false);
      });

      it("should reject kubernetes.io prefixed resources", () => {
        expect(isExtendedResourceName("kubernetes.io/hostname")).toBe(false);
        expect(isExtendedResourceName("kubernetes.io/os")).toBe(false);
      });
    });

    describe("edge cases", () => {
      it("should reject empty string", () => {
        expect(isExtendedResourceName("")).toBe(false);
      });

      it("should reject names that are too long (> 253 chars)", () => {
        const longName = "a".repeat(250) + ".com/resource";
        expect(isExtendedResourceName(longName)).toBe(false);
      });

      it("should accept names at reasonable length", () => {
        // Create a valid qualified name that's under the limit
        // "requests." + name must be valid, and the resource name part has a 253 char limit
        const subdomain = "example.com";
        const resourceName = "my-resource-with-a-long-name";
        const name = `${subdomain}/${resourceName}`;
        expect(isExtendedResourceName(name)).toBe(true);
      });
    });
  });

  describe("isValidExtendedResourceQuantity", () => {
    describe("valid quantities", () => {
      it("should accept plain integers", () => {
        expect(isValidExtendedResourceQuantity("1")).toBe(true);
        expect(isValidExtendedResourceQuantity("100")).toBe(true);
        expect(isValidExtendedResourceQuantity("1000")).toBe(true);
      });

      it("should accept decimal numbers", () => {
        expect(isValidExtendedResourceQuantity("1.5")).toBe(true);
        expect(isValidExtendedResourceQuantity("0.5")).toBe(true);
        expect(isValidExtendedResourceQuantity("100.25")).toBe(true);
      });

      it("should accept numbers with standard SI suffixes", () => {
        expect(isValidExtendedResourceQuantity("1k")).toBe(true);
        expect(isValidExtendedResourceQuantity("1K")).toBe(true);
        expect(isValidExtendedResourceQuantity("100M")).toBe(true);
        expect(isValidExtendedResourceQuantity("10G")).toBe(true);
        expect(isValidExtendedResourceQuantity("5T")).toBe(true);
      });

      it("should accept numbers with binary SI suffixes", () => {
        expect(isValidExtendedResourceQuantity("1Ki")).toBe(true);
        expect(isValidExtendedResourceQuantity("100Mi")).toBe(true);
        expect(isValidExtendedResourceQuantity("10Gi")).toBe(true);
        expect(isValidExtendedResourceQuantity("5Ti")).toBe(true);
      });

      it("should accept numbers with milli suffix", () => {
        expect(isValidExtendedResourceQuantity("100m")).toBe(true);
        expect(isValidExtendedResourceQuantity("500m")).toBe(true);
        expect(isValidExtendedResourceQuantity("1000m")).toBe(true);
      });

      it("should accept numbers with nano, micro suffixes", () => {
        expect(isValidExtendedResourceQuantity("1n")).toBe(true);
        expect(isValidExtendedResourceQuantity("100u")).toBe(true);
        expect(isValidExtendedResourceQuantity("100μ")).toBe(true); // Greek mu
      });

      it("should accept decimal numbers with suffixes", () => {
        expect(isValidExtendedResourceQuantity("1.5Mi")).toBe(true);
        expect(isValidExtendedResourceQuantity("0.5Gi")).toBe(true);
        expect(isValidExtendedResourceQuantity("2.75K")).toBe(true);
      });
    });

    describe("invalid quantities", () => {
      it("should reject empty string", () => {
        expect(isValidExtendedResourceQuantity("")).toBe(false);
      });

      it("should reject non-numeric values", () => {
        expect(isValidExtendedResourceQuantity("abc")).toBe(false);
        expect(isValidExtendedResourceQuantity("Mi")).toBe(false);
      });

      it("should reject invalid suffixes", () => {
        expect(isValidExtendedResourceQuantity("100X")).toBe(false);
        expect(isValidExtendedResourceQuantity("100GB")).toBe(false);
        expect(isValidExtendedResourceQuantity("100Mb")).toBe(false);
      });

      it("should reject numbers with spaces", () => {
        expect(isValidExtendedResourceQuantity("100 Mi")).toBe(false);
        expect(isValidExtendedResourceQuantity("1 00")).toBe(false);
      });

      it("should reject negative numbers", () => {
        expect(isValidExtendedResourceQuantity("-100")).toBe(false);
        expect(isValidExtendedResourceQuantity("-1.5Mi")).toBe(false);
      });
    });
  });

  describe("isValidK8sKey", () => {
    describe("valid keys", () => {
      it("should accept simple lowercase keys", () => {
        expect(isValidK8sKey("key")).toBe(true);
        expect(isValidK8sKey("mykey")).toBe(true);
      });

      it("should accept keys with hyphens", () => {
        expect(isValidK8sKey("my-key")).toBe(true);
        expect(isValidK8sKey("my-long-key")).toBe(true);
      });

      it("should accept keys with underscores", () => {
        expect(isValidK8sKey("my_key")).toBe(true);
        expect(isValidK8sKey("my_long_key")).toBe(true);
      });

      it("should accept keys with dots", () => {
        expect(isValidK8sKey("my.key")).toBe(true);
        expect(isValidK8sKey("my.long.key")).toBe(true);
      });

      it("should accept keys with numbers", () => {
        expect(isValidK8sKey("key1")).toBe(true);
        expect(isValidK8sKey("my-key-123")).toBe(true);
      });

      it("should accept prefixed keys (subdomain/key)", () => {
        expect(isValidK8sKey("example.com/key")).toBe(true);
        expect(isValidK8sKey("sub.example.com/my-key")).toBe(true);
      });

      it("should accept DNS subdomain prefixes with hyphens", () => {
        expect(isValidK8sKey("my-domain.com/key")).toBe(true);
        expect(isValidK8sKey("my-sub.example.com/key")).toBe(true);
      });
    });

    describe("invalid keys", () => {
      it("should reject keys with uppercase letters", () => {
        expect(isValidK8sKey("MyKey")).toBe(false);
        expect(isValidK8sKey("MYKEY")).toBe(false);
      });

      it("should reject keys starting with hyphen", () => {
        expect(isValidK8sKey("-key")).toBe(false);
      });

      it("should reject keys ending with hyphen", () => {
        expect(isValidK8sKey("key-")).toBe(false);
      });

      it("should reject keys with uppercase in prefix", () => {
        expect(isValidK8sKey("Example.com/key")).toBe(false);
        expect(isValidK8sKey("EXAMPLE.COM/key")).toBe(false);
      });

      it("should reject keys with invalid characters", () => {
        expect(isValidK8sKey("my@key")).toBe(false);
        expect(isValidK8sKey("my key")).toBe(false);
        expect(isValidK8sKey("my!key")).toBe(false);
      });

      it("should reject empty keys", () => {
        expect(isValidK8sKey("")).toBe(false);
      });

      it("should reject keys that are too long (> 253 chars)", () => {
        const longKey = "a".repeat(254);
        expect(isValidK8sKey(longKey)).toBe(false);
      });

      it("should reject keys with multiple slashes", () => {
        expect(isValidK8sKey("example.com/sub/key")).toBe(false);
      });

      it("should reject prefixes starting with hyphen", () => {
        expect(isValidK8sKey("-example.com/key")).toBe(false);
      });

      it("should reject prefixes ending with hyphen", () => {
        expect(isValidK8sKey("example-.com/key")).toBe(false);
      });
    });
  });
});
