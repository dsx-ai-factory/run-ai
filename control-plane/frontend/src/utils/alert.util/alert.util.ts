import type { QNotifyCreateOptions } from "quasar";

export const alertUtil = {
  getSuccess,
  getError,
  getWarning,
  getInfo,
  getOngoing,
  getLoading,
};

function getSuccess(message: string, options = {}): QNotifyCreateOptions {
  return {
    position: "bottom",
    type: "positive",
    iconSize: "18px",
    icon: "",
    message,
    ...options,
  };
}

function getError(message: string, options = {}): QNotifyCreateOptions {
  return {
    position: "bottom",
    type: "negative",
    icon: "",
    actions: [{ icon: "ph-bold ph-x", color: "white", flat: true, round: true, size: "sm" }],
    message,
    ...options,
  };
}

function getWarning(message: string, options = {}): QNotifyCreateOptions {
  return {
    position: "bottom",
    type: "warning",
    textColor: "white",
    iconSize: "18px",
    message,
    ...options,
  };
}

function getInfo(message: string, options = {}): QNotifyCreateOptions {
  return {
    position: "bottom",
    type: "info",
    iconSize: "18px",
    message,
    ...options,
  };
}

function getOngoing(message: string, options = {}): QNotifyCreateOptions {
  return {
    icon: "ph-bold ph-warning-circle",
    position: "bottom",
    type: "ongoing",
    iconSize: "18px",
    message,
    ...options,
  };
}

function getLoading(message: string, options = {}): QNotifyCreateOptions {
  return {
    spinner: true,
    type: "positive",
    message,
    ...options,
  };
}
