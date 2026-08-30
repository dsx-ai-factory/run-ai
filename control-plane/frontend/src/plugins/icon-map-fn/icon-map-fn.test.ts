import { describe, expect, it } from "vitest";

import { iconMapFn } from "./icon-map-fn";
import { runaiIcons, toQuasarIconName } from "@/plugins/runai-icons.catalog";

describe("iconMapFn — custom runai-glyph + Phosphor integration", () => {
  describe("custom runai-glyphs render from the SVG sprite (svguse)", () => {
    it("maps a flat custom glyph name to its sprite symbol", () => {
      expect(iconMapFn("runai-cluster")).toEqual({ icon: "svguse:#runai-cluster" });
      expect(iconMapFn("runai-workload")).toEqual({ icon: "svguse:#runai-workload" });
    });

    it("maps variant custom glyph names (outlined/filled) to the matching sprite symbol", () => {
      expect(iconMapFn("runai-outlined-access-rule")).toEqual({ icon: "svguse:#runai-outlined-access-rule" });
      expect(iconMapFn("runai-filled-access-rule")).toEqual({ icon: "svguse:#runai-filled-access-rule" });
    });
  });

  describe("Phosphor icons render as a CSS class (no sprite)", () => {
    it("maps regular-weight Phosphor names", () => {
      expect(iconMapFn("ph ph-funnel")).toEqual({ cls: "ph ph-funnel" });
    });

    it("maps weighted Phosphor names", () => {
      expect(iconMapFn("ph-bold ph-cube")).toEqual({ cls: "ph-bold ph-cube" });
      expect(iconMapFn("ph-fill ph-x-circle")).toEqual({ cls: "ph-fill ph-x-circle" });
    });
  });

  describe("everything else falls back to Quasar's default handling", () => {
    it.each(["chevron_right", "img:/foo.png", "mdi-account", ""])("returns undefined for %p", (name) => {
      expect(iconMapFn(name)).toBeUndefined();
    });
  });

  describe("end-to-end: catalog entry -> toQuasarIconName -> iconMapFn", () => {
    it("custom-glyph catalog entries resolve to a sprite reference", () => {
      const cluster = toQuasarIconName(runaiIcons.entity.cluster);
      expect(cluster).toBe("runai-cluster");
      expect(iconMapFn(cluster)).toEqual({ icon: "svguse:#runai-cluster" });

      const accessRuleOutlined = toQuasarIconName(runaiIcons.entity.accessRule.outlined);
      expect(accessRuleOutlined).toBe("runai-outlined-access-rule");
      expect(iconMapFn(accessRuleOutlined)).toEqual({ icon: "svguse:#runai-outlined-access-rule" });

      const workload = toQuasarIconName(runaiIcons.entity.workload);
      expect(workload).toBe("runai-workload");
      expect(iconMapFn(workload)).toEqual({ icon: "svguse:#runai-workload" });
    });

    it("Phosphor catalog entries resolve to a CSS class", () => {
      const node = toQuasarIconName(runaiIcons.entity.node); // ph-bold ph-cube
      expect(node).toBe("ph-bold ph-cube");
      expect(iconMapFn(node)).toEqual({ cls: "ph-bold ph-cube" });

      const policy = toQuasarIconName(runaiIcons.entity.policy); // regular weight -> "ph ph-clipboard-text"
      expect(policy).toBe("ph ph-clipboard-text");
      expect(iconMapFn(policy)).toEqual({ cls: "ph ph-clipboard-text" });
    });
  });
});
