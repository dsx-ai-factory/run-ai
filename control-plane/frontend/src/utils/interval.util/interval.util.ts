import type { EIntervalLabels } from "@/models/interval.model";
import { intervalLabelTimeMap } from "@/models/interval.model";

interface IIntervalEntry {
  id: number;
  callback: () => void;
}

const intervalRegistry: { [key: string]: IIntervalEntry } = {};
let intervalsPaused = false;

export const intervalUtil = { startInterval, stopInterval, displayActiveIntervals };

function startInterval(label: EIntervalLabels, callback: () => void): number {
  if (!intervalLabelTimeMap[label]) {
    throw new Error(`Interval label ${label} not found in intervalLabelTimeMap`);
  }
  stopInterval(label); // clear existing if any
  if (intervalsPaused) {
    // store callback so resumeAllIntervals will schedule it when connectivity returns
    intervalRegistry[label] = { id: 0, callback };
    return 0;
  }
  const id = Number(setInterval(callback, intervalLabelTimeMap[label]));
  intervalRegistry[label] = { id, callback };
  return id;
}

function stopInterval(label: EIntervalLabels): void {
  if (!intervalLabelTimeMap[label]) {
    throw new Error(`Interval label ${label} not found in intervalLabelTimeMap`);
  }
  if (intervalRegistry[label]) {
    clearInterval(intervalRegistry[label].id);
    delete intervalRegistry[label];
  }
}

function pauseAllIntervals(): void {
  if (intervalsPaused) return;
  intervalsPaused = true;
  Object.values(intervalRegistry).forEach((entry) => {
    clearInterval(entry.id);
    entry.id = 0;
  });
}

function resumeAllIntervals(): void {
  if (!intervalsPaused) return;
  intervalsPaused = false;
  Object.entries(intervalRegistry).forEach(([label, entry]) => {
    entry.id = Number(setInterval(entry.callback, intervalLabelTimeMap[label as EIntervalLabels]));
  });
}

function displayActiveIntervals(): void {
  const activeLabels = Object.keys(intervalRegistry);
  console.log("Active Interval Labels:", activeLabels);
}

function stopAllIntervals(): void {
  Object.keys(intervalRegistry).forEach((label) => stopInterval(label as EIntervalLabels));
  console.log("intervals stopped");
}

function stopIntervalByLabel(label: EIntervalLabels): void {
  stopInterval(label);
  console.log(`${label} interval stopped`);
}

window.addEventListener("offline", pauseAllIntervals);
window.addEventListener("online", resumeAllIntervals);

/*
   make these available in the browser console
 */
declare global {
  interface Window {
    displayActiveIntervals: () => void;
    stopAllIntervals: () => void;
    stopIntervalByLabel: (label: EIntervalLabels) => void;
  }
}

window.displayActiveIntervals = displayActiveIntervals;
window.stopAllIntervals = stopAllIntervals;
window.stopIntervalByLabel = stopIntervalByLabel;
