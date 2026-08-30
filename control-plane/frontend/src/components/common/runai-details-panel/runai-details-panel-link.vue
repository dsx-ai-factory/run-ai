<template>
  <router-link v-if="link.to" :to="link.to" class="runai-details-panel-link" :aid="aid">
    {{ link.text }}
  </router-link>
  <a
    v-else-if="safeHref"
    class="runai-details-panel-link"
    :href="safeHref"
    target="_blank"
    rel="noopener noreferrer"
    :aid="aid"
  >
    {{ link.text }}
  </a>
  <span v-else>{{ link.text }}</span>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { IDetailsPanelItemLink } from "@/models/details-panel.model";

defineOptions({ name: "runai-details-panel-link" });

interface Props {
  link: IDetailsPanelItemLink;
  aid?: string;
}

const props = withDefaults(defineProps<Props>(), { aid: "details-panel-link" });

const SAFE_PROTOCOLS = new Set(["http:", "https:", "mailto:", "tel:"]);

const safeHref = computed<string | undefined>(() => {
  const raw = props.link.href;
  if (!raw) return undefined;
  try {
    const url = new URL(raw, window.location.origin);
    return SAFE_PROTOCOLS.has(url.protocol) ? raw : undefined;
  } catch {
    return undefined;
  }
});
</script>

<style lang="scss" scoped>
.runai-details-panel-link {
  color: var(--text-link-default);
  text-decoration: none;
  cursor: pointer;
  &:hover {
    text-decoration: underline;
  }
}
</style>
