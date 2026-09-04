/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";
import type { ClaimInfo } from "@/swagger-models/assets-service-client";
import type { ClaimSizeCustomization, StorageClassV2 } from "@/swagger-models/k8s-objects-tracker-client";
import { ref } from "vue";

const DEFAULT_CLAIM_SIZE = "1G";

export function useStorageClassConfiguration(formType: "asset" | "workload") {
  const storageClasses = ref<StorageClassV2[]>([]);
  const loadingStorageClass = ref<boolean>(false);

  async function loadStorageClasses(
    clusterId: string,
    selectedStorageClasses: Array<string | undefined | null> = [],
    includeCustomAttributes = false,
  ): Promise<void> {
    try {
      loadingStorageClass.value = true;
      storageClasses.value = (await dataSourceService.listStorageClassV2(clusterId)).filter((sc: StorageClassV2) => {
        if (formType === "workload") {
          return (
            sc.permissions?.allowedForWorkloads !== false &&
            (sc.permissions?.allowedForPersistentVolumes !== false ||
              sc.permissions?.allowedForEphemeralVolumes !== false)
          );
        } else {
          return sc.permissions?.allowedForAssets !== false;
        }
      });
      if (selectedStorageClasses.length) {
        const storageClasses = Array.from(
          new Set(selectedStorageClasses.filter((v): v is string => typeof v === "string" && !!v)),
        );
        await Promise.all(
          storageClasses.map((sc) => loadStorageClassConfiguration(clusterId, sc, includeCustomAttributes)),
        );
      }
    } catch (e: unknown) {
      throw Error("Failed to load storage classes: " + e);
    } finally {
      loadingStorageClass.value = false;
    }
  }

  async function loadStorageClassConfiguration(
    clusterId: string,
    name: string,
    includeCustomAttributes = false,
  ): Promise<ClaimInfo> {
    try {
      loadingStorageClass.value = true;
      const existingStorageClass = storageClasses.value.find((sc: StorageClassV2) => sc.storageClassName === name);

      if (!existingStorageClass) {
        console.warn("Storage class:", name, "not found");
        return {};
      }

      if (existingStorageClass.customization) {
        return _getStorageClassConfigurationClaimInfo(existingStorageClass, includeCustomAttributes);
      }

      const storageClassConfigurations = await dataSourceService.getStorageClassConfigurationByName(clusterId, name);
      existingStorageClass.customization = storageClassConfigurations[0]?.customization;
      return _getStorageClassConfigurationClaimInfo(existingStorageClass, includeCustomAttributes);
    } catch (e: unknown) {
      throw Error("Failed to load storage class configuration: " + e);
    } finally {
      loadingStorageClass.value = false;
    }
  }

  function _getStorageClassConfigurationClaimInfo(
    storageClass: StorageClassV2,
    includeCustomAttributes = false,
  ): ClaimInfo {
    return {
      storageClass: storageClass.storageClassName,
      accessModes: storageClass.customization?.accessMode?.default || {},
      volumeMode: storageClass.customization?.volumeMode?.default || null,
      size:
        storageClass.customization?.claimSize?.default || _getClaimSizeDefault(storageClass.customization?.claimSize),
      ...(includeCustomAttributes && {
        addedAttrValues: storageClass.customization?.attributes?.map((attr) => ({
          key: attr.key,
          value: attr.defaultValue || undefined,
        })),
      }),
    };
  }

  function _getClaimSizeDefault(claimSize: ClaimSizeCustomization | undefined | null): string {
    if (!claimSize) return DEFAULT_CLAIM_SIZE;

    if (claimSize.default) return claimSize.default;

    if (claimSize.min) return claimSize.min;

    if (claimSize.supportedUnits?.length) return `1${claimSize.supportedUnits[0]}`;

    return DEFAULT_CLAIM_SIZE;
  }

  return {
    storageClasses,
    loadingStorageClass,
    loadStorageClasses,
    loadStorageClassConfiguration,
  };
}
