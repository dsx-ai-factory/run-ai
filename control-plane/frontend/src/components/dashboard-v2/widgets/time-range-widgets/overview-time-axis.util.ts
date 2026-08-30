import { PREDEFINED_MINUTES } from "@/models/date.model";

export const ONE_MINUTE_IN_MS = 60 * 1000;
export const TEN_MINUTES_IN_MS = 10 * ONE_MINUTE_IN_MS;

export function getOverviewShortRangeAxisInterval(rangeLabel: string): number {
  return rangeLabel === PREDEFINED_MINUTES.LAST_15_MINUTES ? ONE_MINUTE_IN_MS : TEN_MINUTES_IN_MS;
}
