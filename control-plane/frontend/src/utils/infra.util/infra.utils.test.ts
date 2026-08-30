import { describe, expect, it } from "vitest";
import { getEnvType, EnvType } from "./infra.util";

describe("infra.util", () => {
  describe("infra.util::getEnvConfig", () => {
    it("should return test env type", () => {
      expect(getEnvType("https://envinaclicktest.test.run.ai")).toBe(EnvType.TEST);

      expect(getEnvType("envinaclicktest.test.run.ai")).toBe(EnvType.TEST);
    });
    it("should return staging env type", () => {
      expect(getEnvType("https://envinaclickstaging.staging.run.ai")).toBe(EnvType.STAGING);
      expect(getEnvType("envinaclickstaging.staging.run.ai")).toBe(EnvType.STAGING);
    });
    it("should return prod env type", () => {
      expect(getEnvType("https://envinaclick.run.ai")).toBe(EnvType.PROD);
      expect(getEnvType("envinaclick.run.ai")).toBe(EnvType.PROD);
      expect(getEnvType("https://dgxc.nv.run.ai")).toBe(EnvType.PROD);
      expect(getEnvType("https://dgxc.envidia.run.ai")).toBe(EnvType.PROD);
    });
    it("should return dev env type", () => {
      expect(getEnvType("https://test.runailabs.com")).toBe(EnvType.DEV);
      expect(getEnvType("http://localhost:9000")).toBe(EnvType.DEV);
      expect(getEnvType("http://visa.platform.run.ai")).toBe(EnvType.DEV);
    });
  });
});
