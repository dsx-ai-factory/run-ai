import type { RouteRecordRaw, RouteComponent, RouteParams, NavigationGuard } from "vue-router";
import { type RouteLocationNormalized } from "vue-router";
import type { Action, ResourceType } from "@/swagger-models/authorization-client";
import type { EBadgeLabel } from "@/models/setting.model";
import type { EMenuSubjectNames } from "@/components/infra/runai-layout/routes-model/routes-model";

declare module "vue-router" {
  interface RunaiRouteRecord extends RouteRecordRaw {
    // Unique identifier for the route
    name?: string;
    // Path of the route
    path: string;
    // Component to be rendered when the route is matched
    component: RouteComponent;
    // Additional metadata for the route
    meta?: RouteMeta;
    // Child routes for nested navigation
    children?: RunaiRouteRecord[];
    // Parameters for the route
    params?: RouteParams;
    beforeEnter?: NavigationGuard;
    // Props to be passed to the component
    props?: boolean | RouteParams | ((route: RouteLocationNormalized) => RouteParams);
    // Alias for the route
    alias?: string | string[];
  }

  interface RouteMeta {
    //Page title associated with the route
    pageTitle?: string;
    //Indicates if the route is read-only
    readOnly?: boolean;
    //Indicates if the route supports backward navigation
    supportBack?: boolean;
    //Name of the page to go back to
    backPageName?: string;
    //Indicates whether to hide the cluster version
    hideVersion?: boolean;
    //Indicates whether the page should take up the full width with no header and no footer
    fullPage?: boolean;
    //Indicates whether the page should take up the full width
    fullWidth?: boolean;
    //Indicates if authentication is required
    requiresAuth: boolean;
    //Minimum permitted actions for the user to view the route
    minPermittedActions?: Action[] | null;
    // Type of resource associated with the route
    resourceType?: ResourceType | ResourceType[] | null;
    // Resource types to check with OR logic - user needs the action on ANY of these
    anyOfResourceTypes?: ResourceType[];
    // Action to check for anyOfResourceTypes
    anyOfResourceTypesAction?: Action;
    // Badge label for the route  - legacy, experimental, Beta.
    stageBadge?: EBadgeLabel;
    // Parent subject of the route. Used in the side navigation
    parentSubject?: EMenuSubjectNames;
    // Hide X button in full width view
    hideCloseButton?: boolean;
  }
}
