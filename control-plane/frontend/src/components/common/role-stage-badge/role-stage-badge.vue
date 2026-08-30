<template>
  <q-btn v-if="roleStage" @click.stop :class="badgeClass" :label="roleStage" size="xs" class="q-pa-xs">
    <q-menu class="q-pa-sm" anchor="bottom middle" self="top middle" max-width="400px">
      <div class="role-stage-tooltip-text">
        {{ tooltipMessage.text }}
        <template v-for="(link, index) in tooltipMessage.links" :key="index">
          <a
            v-if="link.url"
            :href="link.url"
            target="_blank"
            rel="noopener noreferrer"
            class="role-stage-link"
            @click.stop
          >
            {{ link.text }}
          </a>
          <span v-else>{{ link.text }}</span>
        </template>
      </div>
    </q-menu>
  </q-btn>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { ROLE_STAGE_CONFIG, ERoleStage } from "@/models/roles.model";
import DOCS from "@/constants/docs";

export interface Props {
  roleName: string;
  hiddenStages?: string[];
}

export interface ITooltipLink {
  text: string;
  url?: string;
}

export interface ITooltipMessage {
  text: string;
  links: ITooltipLink[];
}

const props = defineProps<Props>();

const roleConfig = computed(() => ROLE_STAGE_CONFIG[props.roleName]);

const roleStage = computed(() => {
  const stage = roleConfig.value?.stage;
  if (stage && props.hiddenStages?.includes(stage)) return undefined;
  return stage;
});

const badgeClass = computed(() => {
  if (roleStage.value === ERoleStage.New) {
    return "bg-green-12";
  }
  if (roleStage.value === ERoleStage.Legacy) {
    return "bg-grey-4";
  }
  return "";
});

const tooltipMessage = computed((): ITooltipMessage => {
  const config = ROLE_STAGE_CONFIG[props.roleName];
  if (!config) return { text: "", links: [] };

  if (config.stage === ERoleStage.New) {
    return {
      text: "Explore the ",
      links: [
        { text: "new roles", url: DOCS.AUTHENTICATION.ROLES },
        { text: " and their permissions to see how they can fit your needs" },
      ],
    };
  }

  // Legacy role
  if (config.replacement?.replacementRole) {
    return {
      text: "This role will be removed in a future release. Consider using the ",
      links: [
        { text: config.replacement.replacementRole, url: DOCS.AUTHENTICATION.ROLES },
        { text: " instead or " },
        { text: "create a custom role via API", url: DOCS.AUTHENTICATION.CUSTOM_ROLES_API },
        { text: "." },
      ],
    };
  }

  return {
    text: "This role will be removed in a future release. Consider using a ",
    links: [
      { text: "supported role", url: DOCS.AUTHENTICATION.ROLES },
      { text: " instead or " },
      { text: "create a custom role via API", url: DOCS.AUTHENTICATION.CUSTOM_ROLES_API },
      { text: "." },
    ],
  };
});
</script>

<style lang="scss" scoped>
.role-stage-link {
  color: $primary;
  text-decoration: none;
  cursor: pointer;

  &:hover {
    text-decoration: underline;
  }
}
</style>
