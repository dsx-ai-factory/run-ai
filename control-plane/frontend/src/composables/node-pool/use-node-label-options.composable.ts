/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { onScopeDispose, reactive, watch, type Ref } from "vue";
import { nodeLabelsService } from "@/services/control-plane/node-labels.service/node-labels.service";
import { errorMessages } from "@/common/error-message.constant";

const PAGE_SIZE = 50;
function initialPage() {
  return { options: [] as string[], search: "", offset: 0, more: false, loading: false, error: "", version: 0 };
}
export function useNodeLabelOptions(clusterId: Readonly<Ref<string>>, key: Readonly<Ref<string>>) {
  const keys = reactive(initialPage());
  const values = reactive(initialPage());
  let disposed = false;
  function reset(page: typeof keys) {
    Object.assign(page, { ...initialPage(), version: page.version + 1 });
  }
  watch(
    clusterId,
    () => {
      reset(keys);
      reset(values);
    },
    { flush: "sync" },
  );
  watch(key, () => reset(values), { flush: "sync" });
  onScopeDispose(() => {
    disposed = true;
    reset(keys);
    reset(values);
  });

  async function load(kind: "keys" | "values", search?: string) {
    const page = kind === "keys" ? keys : values;
    if (search !== undefined) {
      reset(page);
      page.search = search;
    } else if (!page.more || page.loading) return;
    if (disposed || !clusterId.value || (kind === "values" && !key.value)) return;
    const version = page.version;
    page.loading = true;
    page.error = "";
    const query = { contains: page.search, offset: page.offset, limit: PAGE_SIZE };
    try {
      const result = await (kind === "keys"
        ? nodeLabelsService.keys(clusterId.value, query)
        : nodeLabelsService.values(clusterId.value, key.value, query));
      if (disposed || version !== page.version) return;
      page.options = [...new Set([...page.options, ...result])];
      page.offset += result.length;
      page.more = result.length === PAGE_SIZE;
    } catch {
      if (version === page.version && !disposed) {
        page.error = errorMessages.NODE_LABEL_SUGGESTIONS_UNAVAILABLE;
        page.more = false;
      }
    } finally {
      if (version === page.version) page.loading = false;
    }
  }
  return {
    keys,
    values,
    searchKeys: (text: string) => load("keys", text),
    searchValues: (text: string) => load("values", text),
    moreKeys: () => load("keys"),
    moreValues: () => load("values"),
  };
}
