export const prodDomains = /^(https:\/\/)?([a-zA-Z0-9-]+\.)?run\.ai$/;
export const stagingDomains = /^(https:\/\/)?([a-zA-Z0-9-]+\.)?staging\.run\.ai$/;
export const testDomains = /^(https:\/\/)?([a-zA-Z0-9-]+\.)?test\.run\.ai$/;
export const nvDomains = /^(https:\/\/)?([a-zA-Z0-9-]+\.)+[a-zA-Z0-9-]*nv[a-zA-Z0-9-]*\.run\.ai$/;

export enum EnvType {
  TEST,
  STAGING,
  PROD,
  DEV,
}

export function isSaas(): boolean {
  return getEnvType(window.location.hostname) !== EnvType.DEV;
}

export function getEnvType(hostName: string = window.location.hostname): EnvType {
  const isProd = prodDomains.test(hostName) || nvDomains.test(hostName);
  const isStaging = stagingDomains.test(hostName);
  const isTest = testDomains.test(hostName);

  if (isTest) {
    return EnvType.TEST;
  } else if (isStaging) {
    return EnvType.STAGING;
  } else if (isProd) {
    return EnvType.PROD;
  }
  return EnvType.DEV;
}

export enum EControlPlanEnv {
  ADOBE = "adobe",
  PREPROD = "nvidia",
  NVIDIA = "preprod",
  RUNAI = "runai-saas",
}

export function getControlPlanEnv(hostName: string = window.location.hostname): EControlPlanEnv {
  const adobe = /^.*\.adobe\.run\.ai/;
  const nvidia = /^.*\.nv\.run\.ai/;
  const preprod = /^.*\.preprod\.run\.ai/;

  if (adobe.test(hostName)) {
    return EControlPlanEnv.ADOBE;
  } else if (nvidia.test(hostName)) {
    return EControlPlanEnv.NVIDIA;
  } else if (preprod.test(hostName)) {
    return EControlPlanEnv.PREPROD;
  }
  return EControlPlanEnv.RUNAI;
}
