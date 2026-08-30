import type { TTestEmit } from "@/models/unit-test.model";
import { DOMWrapper, type VueWrapper } from "@vue/test-utils";
import { useSettingStore } from "@/stores/setting.store";
import type { SettingKeys } from "@/models/setting.model";
import { vi } from "vitest";
import { useRoute } from "vue-router";
import type { DisplayedCluster } from "@/swagger-models/cluster-service-client";
import { ClusterState } from "@/swagger-models/cluster-service-client";
import { useClusterStore } from "@/stores/cluster.store";
import type { Pinia } from "pinia";

export const unitTestUtil = {
  getEmittedValues,
  getModalElement,
  getMenuElement,
  getListElement,
  mockSettingStore,
  getInputValue,
  getInputValueByIndex,
  waitForElement,
  isElementDisabled,
  getMultiSelectValues,
  mockRoute,
  closeExpansionItem,
  mockClusterStore,
  getMenuElementBySelector,
};

const MOCK_CLUSTERS: DisplayedCluster[] = [
  {
    createdAt: "2023-06-14T11:19:48.755Z",
    domain: "https://dev-some-sh-no-cluster.runailabs.com",
    name: "some-test",
    tenantId: 1,
    uuid: "cfd445af-bf35-483b-a742-7603165205b1",
    version: "0.0.0-10000817.master-2a38a34",
    status: {
      state: ClusterState.Connected,
    },
    updatedAt: "2023-06-14T11:19:48.755Z",
    lastLiveness: "2023-06-14T11:19:48.755Z",
  },
];
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function getEmittedValues<T>(cmp: any, eventName: string): TTestEmit<T> {
  return cmp.emitted(eventName) || [[]];
}

function isElementDisabled(wrapper: DOMWrapper<Element>, selector: string): boolean {
  return wrapper.find(selector).attributes("disabled") !== undefined;
}

async function closeExpansionItem(wrapper: VueWrapper): Promise<void> {
  await wrapper.find(".q-expansion-item").trigger("click");
  await wrapper.vm.$nextTick();
}

/********** DOM **********/
export function getModalElement(): DOMWrapper<Element> {
  const documentWrapper = new DOMWrapper(document.body);
  return documentWrapper.find(".q-dialog");
}
export function getMenuElement(): DOMWrapper<Element> {
  const documentWrapper = new DOMWrapper(document.body);
  return documentWrapper.find(".q-menu");
}
function getMenuElementBySelector(selector: string): DOMWrapper<Element> {
  const documentWrapper = new DOMWrapper(document.body);
  const menuElements = documentWrapper.findAll(".q-menu");
  for (const menuElement of menuElements) {
    if (menuElement.find(selector).exists()) {
      return menuElement;
    }
  }
  throw new Error(`Element with selector ${selector} not found in any menu`);
}
export function getListElement(): DOMWrapper<Element> {
  const documentWrapper = new DOMWrapper(document.body);
  return documentWrapper.find(".q-list");
}

/****** input ******/
export function getInputValue(wrapper: VueWrapper, selector: string): string {
  return (wrapper.find(selector).element as HTMLInputElement).value;
}
export function getInputValueByIndex(wrapper: VueWrapper, selector: string, index: number): string {
  return (wrapper.findAll(selector)[index].element as HTMLInputElement).value;
}

/********* select *********/
function getMultiSelectValues(wrapper: VueWrapper): string[] {
  return wrapper.findAll(".q-chip__content").map((chip) => chip.text());
}

/****** store ******/
export function mockSettingStore(settings: { key: SettingKeys; value: boolean }[]) {
  const settingStore = useSettingStore();
  const settingMap = new Map();
  settings.forEach((setting) => {
    settingMap.set(setting.key, { value: setting.value });
  });
  settingStore.$patch({ settings: settingMap });
  return settingStore;
}

export function mockClusterStore(pinia: Pinia, clusters = MOCK_CLUSTERS) {
  const clusterStore = useClusterStore(pinia);
  clusterStore.$patch({ selectedCluster: clusters[0], clusters, isClustersLoaded: true });
  return clusterStore;
}

export async function waitForElement(
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  wrapper: VueWrapper<any>,
  selector: string,
  timeout = 3000,
  interval = 50,
): Promise<DOMWrapper<Element>> {
  const element = await vi.waitUntil(() => (wrapper.find(selector)?.exists() ? wrapper.find(selector) : undefined), {
    timeout,
    interval,
  });
  if (element) {
    return element as DOMWrapper<Element>;
  }
  throw new Error(`Element ${selector} not found within ${timeout}ms`);
}

export const modalStub = {
  RunaiBaseDialog: {
    template: '<div><slot name="header"></slot><slot name="body"></slot><slot name="footer"></slot></div>',
  },
};

/****** mocking ******/

export function mockRoute(query = {}): void {
  vi.mocked(useRoute).mockImplementationOnce(() => ({
    name: "",
    query,
    matched: [],
    fullPath: "",
    hash: "",
    redirectedFrom: undefined,
    params: {},
    path: "",
    meta: { requiresAuth: false, minPermittedActions: [], resourceType: null },
  }));
}

export function mockConsoleError() {
  const errorLog = console.error;
  console.error = vi.fn(); // I want to prevent the console.error from being printed
  return errorLog;
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function restoreConsoleError(errorLog: any) {
  console.error = errorLog;
}

/****** quasar ******/

const mockQuasarObj = {
  screen: { gt: { xs: true } },
  dark: { isActive: false },
  iconSet: { arrow: { dropdown: "ph-fill ph-caret-down" } },
  lang: {
    label: {
      collapse: vi.fn(),
      expand: vi.fn(),
      clear: "Clear",
      ok: "OK",
      cancel: "Cancel",
      close: "Close",
      set: vi.fn(),
      select: "Select",
      reset: "Reset",
      remove: "Remove",
      update: "Update",
      create: "Create",
      search: "Search",
      filter: "Filter",
      refresh: "Refresh",
    },
  },
  notify: vi.fn(),
};

export { mockQuasarObj as mockQuasar };

export function setQuasarMock(wrapper: VueWrapper): void {
  (wrapper.vm as unknown as { $q: typeof mockQuasarObj }).$q = { ...mockQuasarObj };
}
