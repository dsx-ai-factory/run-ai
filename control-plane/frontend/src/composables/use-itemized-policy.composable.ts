/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { ItemRules } from "@/swagger-models/policy-service-client";
import { makeId } from "@/utils/common.util";

import type { ItemizedMapperEntities } from "./helpers/use-itemized-policy.mapper";
import { useItemizedPolicyMapper } from "./helpers/use-itemized-policy.mapper";
import { errorMessages } from "@/common/error-message.constant";
import { computed, type Ref, ref } from "vue";

export interface IBaseItem {
  exclude?: boolean | null | undefined;
}

export function useItemizedPolicy<T extends IBaseItem>(
  items: T[],
  entity: ItemizedMapperEntities,
  policyRules?: ItemRules | null,
  defaultItems?: T[] | null,
) {
  type IPolicyItemized = {
    id: string;
    origin: T;
    isLocked: boolean;
    isDefault: boolean;
  };
  // Helpers
  const isItemLocked = (item: T): boolean => {
    const identifierValue = useItemizedPolicyMapper[entity].getEntityIdentifierValue(item);
    return policyRules?.locked?.some((lockedItem) => lockedItem === identifierValue) || false;
  };

  const isItemDefault = (item: T): boolean => {
    const identifierValue = useItemizedPolicyMapper[entity].getEntityIdentifierValue(item);

    return (
      defaultItems
        ?.map((e) => useItemizedPolicyMapper[entity].getEntityIdentifierValue(e))
        .some((defaultItem) => defaultItem === identifierValue) || false
    );
  };

  // this is a managed list for the composable, it should not be accessed from the outside
  const dirtyList: Ref<IPolicyItemized[] | null> = ref(null);

  const init = (items: T[]) => {
    dirtyList.value = items.map((item: T) => ({
      origin: item,
      id: makeId(6),
      isLocked: isItemLocked(item),
      isDefault: isItemDefault(item),
    }));
  };

  init(items);

  const fullList = computed(() => {
    if (!dirtyList.value) return [];
    return dirtyList.value.map((item) => item.origin);
  });

  const displayedItems = computed(() => {
    if (!dirtyList.value) return [];
    return (
      dirtyList.value
        .filter((item) => !item.origin.exclude)
        .sort((a, b) => Number(b.isLocked || 0) - Number(a.isLocked || 0)) || []
    );
  });

  const cannotAddDueToPolicy = computed(() => {
    return policyRules?.canAdd === false;
  });

  const disabledByPolicyText = errorMessages.CANT_BE_MODIFIED_POLICY;

  // Actions
  const addItem = (item: T): void => {
    dirtyList.value &&
      dirtyList.value.push({
        origin: item,
        id: makeId(6),
        isLocked: isItemLocked(item),
        isDefault: isItemDefault(item),
      });
  };

  const updateItem = (id: string, item: T): void => {
    if (!dirtyList.value) return;
    const idx = dirtyList.value.findIndex((i: IPolicyItemized) => i.id === id);
    if (idx === -1) return;
    dirtyList.value[idx].origin = item;
  };

  const removeItem = (id: string): void => {
    if (!dirtyList.value) return;

    const itemToRemove = dirtyList.value.find((i: IPolicyItemized) => i.id === id);
    if (!itemToRemove) return;
    if (itemToRemove.isLocked) return;

    if (itemToRemove.isDefault) {
      itemToRemove.origin.exclude = true;
    } else {
      dirtyList.value = dirtyList.value?.filter((i) => i.id !== id);
    }
  };

  return {
    disabledByPolicyText,
    cannotAddDueToPolicy,
    addItem,
    removeItem,
    updateItem,
    displayedItems,
    fullList,
  };
}
