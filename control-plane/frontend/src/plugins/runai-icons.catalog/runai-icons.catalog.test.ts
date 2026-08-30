import { describe, expect, it } from "vitest";
import { runaiIcons, toQuasarIconName, type IIconEntry } from "./runai-icons.catalog";

describe("toQuasarIconName", () => {
  it("keeps existing flat runai glyph names backward-compatible", () => {
    const entry: IIconEntry = { phosphor: null, weight: "bold", custom: { glyph: "access" } };

    expect(toQuasarIconName(entry)).toBe("runai-access");
  });

  it("includes the runai glyph variant when one is provided", () => {
    const outlined: IIconEntry = {
      phosphor: null,
      weight: "bold",
      custom: { glyph: "access", variant: "outlined" },
    };
    const filled: IIconEntry = {
      phosphor: null,
      weight: "bold",
      custom: { glyph: "access", variant: "filled" },
    };

    expect(toQuasarIconName(outlined)).toBe("runai-outlined-access");
    expect(toQuasarIconName(filled)).toBe("runai-filled-access");
  });

  it("exposes Figma glyph variants from the catalog", () => {
    expect(toQuasarIconName(runaiIcons.nav.workloadsManager.outlined)).toBe("runai-outlined-workloads-manager");
    expect(toQuasarIconName(runaiIcons.nav.workloadsManager.filled)).toBe("runai-filled-workloads-manager");
    expect(toQuasarIconName(runaiIcons.entity.accessRule.outlined)).toBe("runai-outlined-access-rule");
    expect(toQuasarIconName(runaiIcons.entity.accessRule.filled)).toBe("runai-filled-access-rule");
  });

  it("keeps phosphor icon names unchanged", () => {
    expect(toQuasarIconName({ phosphor: "chart-line", weight: "bold" })).toBe("ph-bold ph-chart-line");
    expect(toQuasarIconName({ phosphor: "info", weight: "regular" })).toBe("ph ph-info");
  });

  it("keeps nav policy bold without changing the regular entity policy icon", () => {
    expect(toQuasarIconName(runaiIcons.nav.policy)).toBe("ph-bold ph-clipboard-text");
    expect(toQuasarIconName(runaiIcons.entity.policy)).toBe("ph ph-clipboard-text");
  });

  it("catalogues chart help and zoom controls", () => {
    expect(toQuasarIconName(runaiIcons.object.help)).toBe("ph-bold ph-question");
    expect(toQuasarIconName(runaiIcons.action.zoomIn)).toBe("ph-bold ph-magnifying-glass-plus");
    expect(toQuasarIconName(runaiIcons.action.zoomOut)).toBe("ph-bold ph-magnifying-glass-minus");
    expect(toQuasarIconName(runaiIcons.action.reset)).toBe("ph-bold ph-arrow-counter-clockwise");
  });
});
