import type { AutoScalingRules } from "@/swagger-models/assets-service-client";
import type { AutoScaling } from "@/swagger-models/workloads-client";
import { SCALE_TO_ZERO_RETENTION_IN_MINUTES_DEFAULT_VALUE } from "./scale-down-section.model";

export const scaleDownUtil = {
  getScaleToZeroRetentionSecondsDefaultValue,
};

function toSeconds(minutes: number): number {
  return minutes * 60;
}

function getScaleToZeroRetentionSecondsDefaultValue(
  scaleToZeroRetentionSeconds?: number,
  autoscalingDefaults?: AutoScaling | null,
  autoscalingRules?: AutoScalingRules | null,
): number {
  if (scaleToZeroRetentionSeconds) return scaleToZeroRetentionSeconds;

  const scaleToZeroRetentionSecondsDefault = autoscalingDefaults?.scaleToZeroRetentionSeconds;
  if (scaleToZeroRetentionSecondsDefault) return scaleToZeroRetentionSecondsDefault;

  const minRule = autoscalingRules?.scaleToZeroRetentionSeconds?.min;

  let retScaleToZeroRetentionSeconds = toSeconds(SCALE_TO_ZERO_RETENTION_IN_MINUTES_DEFAULT_VALUE);
  if (minRule) {
    retScaleToZeroRetentionSeconds = minRule > retScaleToZeroRetentionSeconds ? minRule : retScaleToZeroRetentionSeconds;
  }

  return retScaleToZeroRetentionSeconds;
}
