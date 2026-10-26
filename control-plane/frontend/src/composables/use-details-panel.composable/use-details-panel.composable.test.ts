/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect } from "vitest";
import { ref } from "vue";
import { useDetailsPanel } from "./use-details-panel.composable";
import type { IDetailsPanelItem, IDetailsPanelSectionDef } from "@/models/details-panel.model";
import { matchAnyItem } from "@/utils/details-panel.utils";

function makeSection(id: string, title: string, items: IDetailsPanelItem[]): IDetailsPanelSectionDef {
  return {
    id,
    title,
    customMatch: (term) => matchAnyItem(items, term),
    data: { items },
  };
}

const generalSection = makeSection("general", "General information", [
  { label: "Type", value: "Workspace" },
  { label: "Category", value: "Build" },
  { label: "Project", value: "p-1", link: { text: "p-1", href: "/projects/p-1" } },
  { label: "Cluster", value: "test-cluster" },
]);

const computeSection = makeSection("compute", "Compute", [
  { label: "Node pool", value: "default" },
  { label: "GPU", value: ["1 device", "10GB memory"] },
]);

const cliSection: IDetailsPanelSectionDef = {
  id: "cli",
  title: "CLI Command",
  customMatch: (term) => "runai workspace submit".includes(term.toLowerCase()),
  data: { command: "runai workspace submit" },
};

describe("useDetailsPanel", () => {
  describe("initial state", () => {
    it("opens all sections by default", () => {
      const { expandedById } = useDetailsPanel({
        sections: [generalSection, computeSection],
      });

      expect(expandedById.value).toEqual({ general: true, compute: true });
    });

    it("starts with empty search term and isSearching false", () => {
      const { searchTerm, isSearching } = useDetailsPanel({ sections: [generalSection] });

      expect(searchTerm.value).toBe("");
      expect(isSearching.value).toBe(false);
    });

    it("returns all sections when not searching", () => {
      const { visibleSections } = useDetailsPanel({
        sections: [generalSection, computeSection, cliSection],
      });

      expect(visibleSections.value).toHaveLength(3);
    });
  });

  describe("search filtering", () => {
    it("filters visible sections by key match", () => {
      const { onSearchUpdate, visibleSections } = useDetailsPanel({
        sections: [generalSection, computeSection],
      });

      onSearchUpdate("project");

      expect(visibleSections.value.map((s) => s.id)).toEqual(["general"]);
    });

    it("filters visible sections by value match", () => {
      const { onSearchUpdate, visibleSections } = useDetailsPanel({
        sections: [generalSection, computeSection],
      });

      onSearchUpdate("default");

      expect(visibleSections.value.map((s) => s.id)).toEqual(["compute"]);
    });

    it("matches case-insensitively", () => {
      const { onSearchUpdate, visibleSections } = useDetailsPanel({ sections: [generalSection] });

      onSearchUpdate("WORKSPACE");

      expect(visibleSections.value).toHaveLength(1);
    });

    it("matches against array values (joined)", () => {
      const { onSearchUpdate, visibleSections } = useDetailsPanel({ sections: [computeSection] });

      onSearchUpdate("10GB");

      expect(visibleSections.value).toHaveLength(1);
    });

    it("matches link text", () => {
      const { onSearchUpdate, visibleSections } = useDetailsPanel({ sections: [generalSection] });

      onSearchUpdate("p-1");

      expect(visibleSections.value).toHaveLength(1);
    });

    it("hides a section whose customMatch always returns false", () => {
      const neverMatchSection: IDetailsPanelSectionDef = {
        id: "topology",
        title: "Topology",
        customMatch: () => false,
        data: {},
      };
      const { onSearchUpdate, visibleSections } = useDetailsPanel({
        sections: [generalSection, neverMatchSection],
      });

      onSearchUpdate("topology");

      expect(visibleSections.value.map((s) => s.id)).toEqual([]);
    });

    it("uses customMatch for custom-rendered sections", () => {
      const { onSearchUpdate, visibleSections } = useDetailsPanel({
        sections: [generalSection, cliSection],
      });

      onSearchUpdate("submit");

      expect(visibleSections.value.map((s) => s.id)).toEqual(["cli"]);
    });

    it("treats a whitespace-only term as an empty search (all sections visible, default expansion)", () => {
      const { onSearchUpdate, visibleSections, isSearching, expandedById, normalizedSearchTerm } = useDetailsPanel({
        sections: [generalSection, computeSection],
      });

      onSearchUpdate("   ");

      expect(isSearching.value).toBe(false);
      expect(normalizedSearchTerm.value).toBe("");
      expect(visibleSections.value.map((s) => s.id)).toEqual(["general", "compute"]);
      expect(expandedById.value).toEqual({ general: true, compute: true });
    });

    it("exposes normalizedSearchTerm trimmed of surrounding whitespace", () => {
      const { onSearchUpdate, normalizedSearchTerm } = useDetailsPanel({ sections: [generalSection, computeSection] });

      onSearchUpdate("  project  ");

      expect(normalizedSearchTerm.value).toBe("project");
    });
  });

  describe("expansion behaviour", () => {
    it("expands every visible section when search becomes active", () => {
      const { onSearchUpdate, expandedById } = useDetailsPanel({
        sections: [generalSection, computeSection],
      });

      onSearchUpdate("default");

      expect(expandedById.value).toEqual({ compute: true });
    });

    it("returns to default expansion (all expanded) when search term is cleared", () => {
      const { onSearchUpdate, expandedById } = useDetailsPanel({
        sections: [generalSection, computeSection],
      });
      onSearchUpdate("default");

      onSearchUpdate("");

      expect(expandedById.value).toEqual({ general: true, compute: true });
    });

    it("allows manual toggle of a section", () => {
      const { onToggle, expandedById } = useDetailsPanel({
        sections: [generalSection, computeSection],
      });

      onToggle("general", false);

      expect(expandedById.value.general).toBe(false);
    });
  });

  describe("expand/collapse all", () => {
    it("allExpanded is true by default (all sections are open)", () => {
      const { allExpanded } = useDetailsPanel({
        sections: [generalSection, computeSection],
      });

      expect(allExpanded.value).toBe(true);
    });

    it("toggleAll collapses when all are open, expands when all are closed", () => {
      const { toggleAll, expandedById, allExpanded } = useDetailsPanel({
        sections: [generalSection, computeSection],
      });

      toggleAll();
      expect(expandedById.value).toEqual({ general: false, compute: false });
      expect(allExpanded.value).toBe(false);

      toggleAll();
      expect(expandedById.value).toEqual({ general: true, compute: true });
      expect(allExpanded.value).toBe(true);
    });
  });

  describe("reactive sections", () => {
    it("preserves manual collapse when the sections list changes (async load), new sections expand by default", async () => {
      const sections = ref<IDetailsPanelSectionDef[]>([generalSection]);
      const { onToggle, expandedById } = useDetailsPanel({ sections });
      onToggle("general", false);

      sections.value = [generalSection, computeSection];
      await Promise.resolve();

      expect(expandedById.value.general).toBe(false);
      expect(expandedById.value.compute).toBe(true);
    });

    it("opens all sections when sections appear after an empty start", async () => {
      const sections = ref<IDetailsPanelSectionDef[]>([]);
      const { expandedById } = useDetailsPanel({ sections });

      sections.value = [generalSection, computeSection];
      await Promise.resolve();

      expect(expandedById.value.general).toBe(true);
      expect(expandedById.value.compute).toBe(true);
    });
  });
});
