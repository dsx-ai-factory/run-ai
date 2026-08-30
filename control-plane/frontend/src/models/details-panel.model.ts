import type { RouteLocationRaw } from "vue-router";

export interface IDetailsPanelItemLink {
  /** Internal route — rendered via <router-link>. */
  to?: RouteLocationRaw;
  /** External URL — rendered via <a target="_blank">. Mutually exclusive with `to`. */
  href?: string;
  /** Display text. */
  text?: string;
}

export interface IDetailsPanelItem {
  /** Rendered in the labels column AND used for key-side matching. */
  label: string;
  /** Used for value-side matching (case-insensitive substring). Array values match against the joined string. */
  value: string | string[];
  /** Optional link payload — used for value-side matching against `link.text` and consumed by the consumer when rendering the row's slot. */
  link?: IDetailsPanelItemLink;
  /** When true, the value is rendered with code styling (mono background) instead of plain text. */
  isCode?: boolean;
}

/** A visually-grouped batch of items. Within a group: 7 px gap. Between groups: 16 px gap. */
export type TDetailsPanelItemGroup = IDetailsPanelItem[];

export interface IDetailsPanelTopologyPlacement {
  level?: string;
  value?: string;
}

export interface IDetailsPanelTopology {
  name?: string;
  preferredPlacement?: string;
  requiredPlacement?: string;
  actualPlacement?: IDetailsPanelTopologyPlacement | null;
}

export interface IDetailsPanelNetworkConnection {
  name: string;
  tool?: string;
  connectionType?: string;
  access: string;
  port?: number;
  url?: string;
}

export interface IDetailsPanelStorageVolumeField {
  label: string;
  value: string;
  /** When true, the value is rendered with code styling (mono background). */
  isCode?: boolean;
}

export interface IDetailsPanelStorageVolume {
  name: string;
  fields: IDetailsPanelStorageVolumeField[];
}

/**
 * Generic section definition consumed by `<runai-details-panel>`.
 * Each section carries its own typed `data` payload (read by the section component via the
 * named slot) and a `customMatch` predicate that decides whether the card is visible during
 * a search.
 */
export interface IDetailsPanelSectionDef<TData = unknown> {
  /** Stable id used for the section slot name (`section-${id}`) and expansion-state tracking. */
  id: string;
  /** Card header text. */
  title: string;
  /** Returns true when the section should remain visible for the given search term. */
  customMatch: (term: string) => boolean;
  /** Section-specific render payload. Read by the section component via the slot. */
  data: TData;
}
