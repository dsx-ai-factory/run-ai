/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

declare module "vue-draggable-plus" {
  import type { DefineComponent, ComponentOptionsMixin, PublicProps } from "vue";
  import type { Options, SortableEvent } from "sortablejs";

  export { Options, SortableEvent } from "sortablejs";

  export interface DraggableEvent<T = unknown> extends SortableEvent {
    data: T;
    clonedData: T;
  }

  export interface UseDraggableOptions<T = unknown> extends Options {
    clone?: (element: T) => T;
    immediate?: boolean;
    customUpdate?: (event: SortableEvent) => void;
  }

  export const VueDraggable: DefineComponent<
    IProps,
    Record<string, never>,
    Record<string, never>,
    Record<string, never>,
    Record<string, never>,
    ComponentOptionsMixin,
    ComponentOptionsMixin,
    Record<string, never>,
    string,
    PublicProps,
    Readonly<IProps>,
    Record<string, never>,
    Record<string, never>
  >;
}
