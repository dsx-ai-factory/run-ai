import type { EIntervalLabels } from "@/models/interval.model";
import { onMounted, onUnmounted } from "vue";
import { intervalUtil } from "@/utils/interval.util";

export function useRefresh(label: EIntervalLabels, callback: () => void): void {
  onMounted(() => {
    callback();
    intervalUtil.startInterval(label, callback);
  });

  onUnmounted(() => {
    intervalUtil.stopInterval(label);
  });
}
