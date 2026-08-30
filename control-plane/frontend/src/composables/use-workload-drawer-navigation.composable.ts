import { useRouter, useRoute } from "vue-router";
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";
import type { TWorkloadDrawerTab, IWorkloadDrawerNavigationPayload } from "@/models/workload.model";

export interface IWorkloadDrawerUrlParams extends IWorkloadDrawerNavigationPayload {
  workloadId: string;
}

export function useWorkloadDrawerNavigation() {
  const router = useRouter();
  const route = useRoute();

  /**
   * Navigate to workload page with drawer params (for cross-page navigation)
   */
  function navigateToWorkloadDrawer(workloadId: string, tab: TWorkloadDrawerTab, podName?: string): void {
    router.push({
      name: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX,
      query: {
        openDrawer: "true",
        workloadId,
        drawerTab: tab,
        podName,
      },
    });
  }

  /**
   * Read drawer params from URL (for workload-index on mount)
   */
  function getDrawerParamsFromUrl(): IWorkloadDrawerUrlParams | null {
    if (route.query.openDrawer !== "true") return null;

    const workloadId = route.query.workloadId;
    const tab = route.query.drawerTab;

    if (typeof workloadId !== "string" || typeof tab !== "string") return null;

    const podName = route.query.podName;

    return {
      workloadId,
      tab: tab as TWorkloadDrawerTab,
      podName: typeof podName === "string" ? podName : "",
    };
  }

  /**
   * Clear drawer params from URL after handling
   */
  function clearDrawerParams(): void {
    const query = { ...route.query };
    delete query.openDrawer;
    delete query.workloadId;
    delete query.drawerTab;
    delete query.podName;
    // Clean up structure-page params carried over by back-or-close-btn
    delete query.tab;
    delete query.name;
    delete query.type;
    delete query.phase;
    delete query.phaseUpdatedAt;
    delete query.elementId;
    router.replace({ query });
  }

  return {
    navigateToWorkloadDrawer,
    getDrawerParamsFromUrl,
    clearDrawerParams,
  };
}
