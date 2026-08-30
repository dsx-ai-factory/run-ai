import Axios from "axios";
import { BootConfigApi, Configuration, type BootConfig } from "@/swagger-models/bff-service-client";

const BOOT_CONFIG_TIMEOUT_MS = 5_000;

export const FALLBACK_BOOT_CONFIG: BootConfig = {
  distributionAssets: {
    profile: "cc",
  },
  version: { display: "" },
  docs: {
    baseUrl: "https://run-ai-docs.nvidia.com/self-hosted",
    whatsNew: {
      label: "What's New",
      url: "https://run-ai-docs.nvidia.com/self-hosted/getting-started/whats-new/whats-new",
    },
  },
  support: {
    url: "",
    label: "NVIDIA Run:ai support",
  },
  capabilities: {
    showsClusterLocationStep: true,
  },
  unleashConfig: null,
};

let bootConfig: BootConfig = FALLBACK_BOOT_CONFIG;

const plainAxios = Axios.create({
  timeout: BOOT_CONFIG_TIMEOUT_MS,
  withCredentials: false,
});

const api = new BootConfigApi(new Configuration(), window.location.origin, plainAxios);

export async function loadBootConfig(): Promise<BootConfig> {
  try {
    const response = await api.getBootConfig();
    if (!isBootConfig(response.data)) throw new Error("invalid boot configuration response");
    bootConfig = response.data;
  } catch (error: unknown) {
    console.warn("Could not load frontend boot configuration; using fallback defaults.", error);
    bootConfig = FALLBACK_BOOT_CONFIG;
  }
  return bootConfig;
}

export function getBootConfig(): BootConfig {
  return bootConfig;
}

export function showsClusterLocationStep(): boolean {
  return bootConfig.capabilities.showsClusterLocationStep;
}

function isBootConfig(value: unknown): value is BootConfig {
  if (!value || typeof value !== "object") return false;
  const candidate = value as BootConfig;
  const unleash = candidate.unleashConfig;

  return (
    (candidate.distributionAssets?.profile === "cc" || candidate.distributionAssets?.profile === "ce") &&
    typeof candidate.version?.display === "string" &&
    typeof candidate.docs?.baseUrl === "string" &&
    isSafeBrowserUrl(candidate.docs.baseUrl, false) &&
    typeof candidate.docs?.whatsNew?.label === "string" &&
    typeof candidate.docs?.whatsNew?.url === "string" &&
    isSafeBrowserUrl(candidate.docs.whatsNew.url, false) &&
    (candidate.support?.url === undefined ||
      (typeof candidate.support.url === "string" && isSafeBrowserUrl(candidate.support.url, false))) &&
    typeof candidate.support?.label === "string" &&
    typeof candidate.capabilities?.showsClusterLocationStep === "boolean" &&
    (unleash === null ||
      (typeof unleash?.url === "string" &&
        isSafeBrowserUrl(unleash.url, false) &&
        typeof unleash.clientKey === "string" &&
        typeof unleash.appName === "string" &&
        typeof unleash.environment === "string" &&
        (unleash.project === undefined || typeof unleash.project === "string")))
  );
}

function isSafeBrowserUrl(value: string, allowEmpty: boolean): boolean {
  if (value === "") return allowEmpty;
  if (value.trim() !== value) return false;
  if (value.startsWith("/")) return !/^\/[\\/]/.test(value);

  try {
    const url = new URL(value);
    return url.protocol === "http:" || url.protocol === "https:";
  } catch {
    return false;
  }
}
