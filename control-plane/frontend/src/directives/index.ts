import type { App } from "vue";
import { clickOutside } from "./ClickOutside.directive";
import { hasPermission } from "@/directives/permissions.directive";

export const useDirectives = (app: App) => {
  app.directive("click-outside", clickOutside);
  app.directive("permission", hasPermission);
};
