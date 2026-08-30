import { computed, isRef, ref, watch, type ComputedRef, type Ref } from "vue";
import type { IDetailsPanelSectionDef } from "@/models/details-panel.model";

export interface IUseDetailsPanelOptions {
  sections: Ref<IDetailsPanelSectionDef[]> | IDetailsPanelSectionDef[];
}

export interface IUseDetailsPanelReturn {
  searchTerm: Ref<string>;
  /** Trimmed search term — what callers should hand to match/filter logic. */
  normalizedSearchTerm: ComputedRef<string>;
  expandedById: Ref<Record<string, boolean>>;
  isSearching: ComputedRef<boolean>;
  visibleSections: ComputedRef<IDetailsPanelSectionDef[]>;
  allExpanded: ComputedRef<boolean>;
  onSearchUpdate: (term: string) => void;
  onToggle: (sectionId: string, value: boolean) => void;
  expandAll: () => void;
  collapseAll: () => void;
  toggleAll: () => void;
}

export function useDetailsPanel(options: IUseDetailsPanelOptions): IUseDetailsPanelReturn {
  const sectionsRef: Ref<IDetailsPanelSectionDef[]> = isRef(options.sections) ? options.sections : ref(options.sections);

  const searchTerm = ref("");
  const expandedById = ref<Record<string, boolean>>({});

  const normalizedSearchTerm = computed<string>(() => searchTerm.value.trim());
  const isSearching = computed<boolean>(() => normalizedSearchTerm.value.length > 0);

  const visibleSections = computed<IDetailsPanelSectionDef[]>(() =>
    sectionsRef.value.filter((section) => !isSearching.value || section.customMatch(normalizedSearchTerm.value)),
  );

  const allExpanded = computed<boolean>(() => {
    const visible = visibleSections.value;
    if (visible.length === 0) return false;
    return visible.every((section) => expandedById.value[section.id] === true);
  });

  function applyDefaultExpansion(): void {
    const next: Record<string, boolean> = {};
    for (const section of sectionsRef.value) {
      next[section.id] = true;
    }
    expandedById.value = next;
  }

  function applySearchExpansion(): void {
    const next: Record<string, boolean> = {};
    for (const section of visibleSections.value) {
      next[section.id] = true;
    }
    expandedById.value = next;
  }

  function onSearchUpdate(term: string): void {
    searchTerm.value = term;
    if (isSearching.value) {
      applySearchExpansion();
    } else {
      applyDefaultExpansion();
    }
  }

  function onToggle(sectionId: string, value: boolean): void {
    expandedById.value = { ...expandedById.value, [sectionId]: value };
  }

  function expandAll(): void {
    const next: Record<string, boolean> = { ...expandedById.value };
    for (const section of visibleSections.value) {
      next[section.id] = true;
    }
    expandedById.value = next;
  }

  function collapseAll(): void {
    const next: Record<string, boolean> = {};
    for (const section of sectionsRef.value) {
      next[section.id] = false;
    }
    expandedById.value = next;
  }

  function toggleAll(): void {
    if (allExpanded.value) {
      collapseAll();
    } else {
      expandAll();
    }
  }

  applyDefaultExpansion();

  watch(
    () => sectionsRef.value.map((s) => s.id).join("|"),
    () => {
      const next: Record<string, boolean> = {};
      for (const section of sectionsRef.value) {
        next[section.id] = expandedById.value[section.id] ?? true;
      }
      expandedById.value = next;
    },
  );

  return {
    searchTerm,
    normalizedSearchTerm,
    expandedById,
    isSearching,
    visibleSections,
    allExpanded,
    onSearchUpdate,
    onToggle,
    expandAll,
    collapseAll,
    toggleAll,
  };
}
