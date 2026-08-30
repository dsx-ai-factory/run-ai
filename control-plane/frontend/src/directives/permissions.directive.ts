import type { ResourceType, Action } from "@/swagger-models/authorization-client";
import { usePermissionStore } from "@/stores/permissions.store";

export interface IResourceAccessAttempt {
  action: Action;
  resourceType: ResourceType;
}

export const hasPermission = {
  beforeMount: (el: HTMLInputElement, { value }: { value: IResourceAccessAttempt }) => {
    const shouldDisable = !usePermissionStore().hasPermission(value.resourceType, value.action);
    if (!shouldDisable) return;
    el.disabled = true;
  },
};
