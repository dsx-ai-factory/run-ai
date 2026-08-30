<template>
  <section class="access-rule-management" :class="noInnerPadding ? '' : 'q-px-md q-pt-md'">
    <div class="sub-header">{{ subheader }}</div>
    <template v-if="!hideSearch && (!noAccessRules || !!searchTerm)">
      <div class="row">
        <div class="col-4">
          <q-input
            :model-value="searchTerm"
            @update:model-value="updateSearchTerm"
            clearable
            outlined
            dense
            clear-icon="ph-bold ph-x"
            debounce="300"
            placeholder="Search"
          >
            <template v-slot:prepend>
              <q-icon name="ph-bold ph-magnifying-glass" size="20px" color="icon-default" />
            </template>
          </q-input>
        </div>
      </div>

      <q-separator class="q-mt-md" />
    </template>
    <q-card-section class="scroll-target q-px-none" :class="noAccessRules ? 'q-pa-none' : 'q-py-none'">
      <!-- The 'overflow-auto' class is required for proper functionality of this component -->
      <div v-if="initialLoader" class="row justify-center q-py-md">
        <q-spinner-dots size="40px" color="primary" />
      </div>
      <div
        v-else
        :style="containerStyle"
        class="overflow-auto q-px-none"
        :class="noAccessRules ? 'q-py-none' : 'q-py-md'"
      >
        <q-infinite-scroll
          ref="infiniteScrollRef"
          :offset="250"
          @load="loadMoreAccessRules"
          scroll-target=".scroll-target"
        >
          <access-rule-item
            v-for="(accessRule, i) in groupAccessRules"
            :key="`group-${accessRule.id ?? i}`"
            :access-rule="accessRule"
            :display-config="displayConfig"
            :subject-options="subjects"
            :role-options="rolesOptions"
            :page="formOptions.page"
            :is-inherited-from-group="true"
          />
          <access-rule-item
            v-for="(accessRule, i) in allAuthAccessRules"
            :key="`all-auth-${accessRule.id ?? i}`"
            :access-rule="accessRule"
            :display-config="displayConfig"
            :subject-options="subjects"
            :role-options="rolesOptions"
            :page="formOptions.page"
            :is-inherited-all-auth="true"
          />
          <access-rule-item
            v-for="(accessRule, i) in accessRules"
            :key="accessRule.subjectId"
            :access-rule="accessRule"
            :display-config="displayConfig"
            :subject-options="subjects"
            :role-options="rolesOptions"
            :page="formOptions.page"
            :is-inherited-from-group="false"
            @delete-rule="deleteRule(accessRule.id, i)"
          />
          <div v-if="displayEmptyState" class="empty-state">No access rules match your filters</div>
          <template v-slot:loading>
            <div class="row justify-center">
              <q-spinner-dots size="40px" color="primary" />
            </div>
          </template>
        </q-infinite-scroll>
      </div>
    </q-card-section>
    <q-separator />
    <q-card-actions class="q-px-none" v-if="!isNewAccessRuleFormOpen">
      <q-btn
        :disable="disableAccessRuleBtn"
        v-permission="{ resourceType: ResourceType.AccessRules, action: Action.Create }"
        @click="toggleNewAccessRuleForm"
        flat
        color="primary"
        class="btn-icon-sm"
        icon="ph-bold ph-plus"
        label="ACCESS RULE"
      />
    </q-card-actions>

    <div v-if="isNewAccessRuleFormOpen">
      <q-card-section class="q-px-none" :class="noInnerPadding ? 'q-pb-none' : ''">
        <new-management-access-rule-form
          ref="newAccessRuleFormRef"
          :display-config="displayConfig"
          :options="formOptions"
          :role-options="rolesOptions"
          :is-sso="isSso"
          @close="toggleNewAccessRuleForm"
          @save="onAccessRulesSaved"
        />
      </q-card-section>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from "vue";
// Components
import { AccessRuleItem } from "@/components/rbac/access-rule/access-rule-management-modal/access-rule-item/";
import { NewManagementAccessRuleForm } from "@/components/rbac/access-rule/access-rule-management-modal/new-management-access-rule-form";
import { QInfiniteScroll } from "quasar";
import { useQuasar } from "quasar";

// Models
import type { IAccessRuleDisplayConfig, IAccessRuleManagementFormOptions } from "@/models/access-rule.model";
import { subjectTypeOptions, subjectTypeOptionsWithSSO } from "@/models/access-rule.model";
import {
  Action,
  ResourceType,
  SubjectType,
  SubjectTypeExtended,
  type RoleV1,
  type AccessRule,
  AccessRulesSortFilterFields,
  type AccessRulesBatchCreateResponse,
} from "@/swagger-models/authorization-client";
import type { ISelectOption } from "@/models/global.model";

// Services
import { accessRuleService } from "@/services/control-plane/rbac/access-rule.service/access-rule.service";
import { rolesService } from "@/services/control-plane/rbac/roles.service/roles.service";

// Stores
import { usePermissionStore } from "@/stores/permissions.store";
import { useRoleStore } from "@/stores/role.store";
import { useAuthStore } from "@/stores/auth.store";

// Utils
import { accessRuleUtil } from "@/utils/rbac.util/access-rule.util/access-rule.util";
import { alertUtil } from "@/utils/alert.util";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { ErrorAlert } from "@/utils/error-alert.util";

const BATCH_SIZE = 10;

// Props
const props = defineProps<{
  formOptions: IAccessRuleManagementFormOptions;
  noInnerPadding?: boolean;
  autoOpenForm?: boolean;
  hideSearch?: boolean;
}>();

// Emits
const emit = defineEmits<{
  "access-rules-created": [createdAccessRules: AccessRulesBatchCreateResponse];
  "access-rule-deleted": [index: number];
}>();

// Composables
const $q = useQuasar();
const roleStore = useRoleStore();
const permissionStore = usePermissionStore();
const authStore = useAuthStore();

// Refs
const displayConfig = ref<IAccessRuleDisplayConfig>({} as IAccessRuleDisplayConfig);
const loadingMore = ref(false);
const initialLoader = ref(true);
const isNewAccessRuleFormOpen = ref(false);
const searchTerm = ref("");
const accessRules = ref<AccessRule[]>([]);
const allGroupAccessRules = ref<AccessRule[]>([]);
const allAuthAccessRulesRaw = ref<AccessRule[]>([]);

const infiniteScrollRef = ref<QInfiniteScroll>();
const totalAccessRulesCount = ref(100);
const newAccessRuleFormRef = ref<InstanceType<typeof NewManagementAccessRuleForm> | null>(null);

// Computed
const subheader = computed(() => accessRuleUtil.getAccessRuleModalSubHeader(props.formOptions.page));

const groupAccessRules = computed<AccessRule[]>(() => {
  if (!searchTerm.value) return allGroupAccessRules.value;
  const term = searchTerm.value.toLowerCase();
  return allGroupAccessRules.value.filter(
    (r) =>
      r.roleName?.toLowerCase().includes(term) ||
      r.scopeType?.toLowerCase().includes(term) ||
      r.scopeName?.toLowerCase().includes(term) ||
      r.subjectId?.toLowerCase().includes(term),
  );
});

const allAuthAccessRules = computed<AccessRule[]>(() => {
  if (!searchTerm.value) return allAuthAccessRulesRaw.value;
  const term = searchTerm.value.toLowerCase();
  return allAuthAccessRulesRaw.value.filter(
    (r) =>
      r.roleName?.toLowerCase().includes(term) ||
      r.scopeType?.toLowerCase().includes(term) ||
      r.scopeName?.toLowerCase().includes(term),
  );
});

const noAccessRules = computed(
  () => accessRules.value.length === 0 && groupAccessRules.value.length === 0 && allAuthAccessRules.value.length === 0,
);

const disableAccessRuleBtn = computed(() => !hasReadRolesPermission.value);

const hasReadRolesPermission = computed(() => permissionStore.hasPermission(ResourceType.Roles, Action.Read));

const subjects = computed<ISelectOption[]>(() => (isSso.value ? subjectTypeOptionsWithSSO : subjectTypeOptions));

const roles = computed<RoleV1[]>(() =>
  roleStore.roleList.filter((role: RoleV1) => rolesService.isPermittedRole(role, permissionStore.permissionsList)),
);

const rolesOptions = computed<ISelectOption[]>(() =>
  roles.value.map((role: RoleV1) => ({
    label: role.name,
    value: role.id,
    roleStageBadge: role.name, // Pass role name for badge lookup
  })),
);

const isSso = computed(() => authStore.isSSO);

const getFilterBy = computed<string[]>(() => {
  if (props.formOptions.scopeId) {
    return [filterUtil.getEqualsFilterString(AccessRulesSortFilterFields.ScopeId, props.formOptions.scopeId)];
  }
  if (props.formOptions.subjectId) {
    const filters = [
      filterUtil.getEqualsFilterString(AccessRulesSortFilterFields.SubjectId, props.formOptions.subjectId),
    ];
    if (props.formOptions.subjectType) {
      filters.push(
        filterUtil.getEqualsFilterString(AccessRulesSortFilterFields.SubjectType, props.formOptions.subjectType),
      );
    }
    return filters;
  }

  return [];
});

const displayEmptyState = computed(
  () =>
    accessRules.value.length === 0 &&
    groupAccessRules.value.length === 0 &&
    allAuthAccessRules.value.length === 0 &&
    !loadingMore.value &&
    !!searchTerm.value,
);

const containerStyle = computed(() => ({
  maxHeight:
    accessRules.value.length > 0 || groupAccessRules.value.length > 0 || allAuthAccessRules.value.length > 0
      ? "47vh"
      : "none",
}));

// Computed for unsaved changes guard
const isFormDirty = computed(() => {
  if (!isNewAccessRuleFormOpen.value) return false;
  return newAccessRuleFormRef.value?.isDirty ?? false;
});

// Methods
async function deleteRule(accessRuleId: number, index: number): Promise<void> {
  try {
    await accessRuleService.deleteAccessRule(accessRuleId);
    $q.notify(alertUtil.getSuccess("Access rule deleted"));

    emit("access-rule-deleted", index);
    accessRules.value.splice(index, 1);
    // Update total count after deletion
    totalAccessRulesCount.value = Math.max(0, totalAccessRulesCount.value - 1);
  } catch (error: unknown) {
    handleError(error, "Failed to delete access rule");
  }
}

function handleError(error: unknown, customMessage: string): void {
  const errorAlert = new ErrorAlert({
    generalMessage: customMessage,
  });
  $q.notify(errorAlert.getNotification(error));
}

function updateSearchTerm(value: string | number | null) {
  const term = typeof value === "string" ? value : "";
  searchTerm.value = term || "";
  resetAndTriggerInfiniteScroll();
}

function resetAndTriggerInfiniteScroll(): void {
  const infiniteScroll = infiniteScrollRef.value;
  if (!infiniteScroll) return;

  accessRules.value = [];
  setTimeout(() => {
    infiniteScroll.reset();
    infiniteScroll.resume();
  }, 0);
}

async function loadMoreAccessRules(index: number, done: (stop?: boolean) => void) {
  if (props.hideSearch && totalAccessRulesCount.value === 0) {
    done(true);
    return;
  }
  try {
    loadingMore.value = true;
    const start = (index - 1) * BATCH_SIZE;
    const fetchedAccessRules = await accessRuleService.getAccessRulesV2(
      getFilterBy.value,
      AccessRulesSortFilterFields.CreatedAt,
      "desc",
      start,
      BATCH_SIZE,
      searchTerm.value,
    );
    loadingMore.value = false;
    accessRules.value.push(...fetchedAccessRules);
    const allLoaded = accessRules.value.length >= totalAccessRulesCount.value || fetchedAccessRules.length < BATCH_SIZE;
    done(allLoaded);
  } catch (error: unknown) {
    handleError(error, "Failed to load access rules");
    done(true);
  } finally {
    loadingMore.value = false;
  }
}

function onAccessRulesSaved(createdAccessRules: AccessRulesBatchCreateResponse): void {
  isNewAccessRuleFormOpen.value = false;
  const succeededCount = createdAccessRules?.succeeded?.length || 0;
  const totalCount = (createdAccessRules?.succeeded?.length || 0) + (createdAccessRules?.failed?.length || 0);
  notifyCreatedAccessRules(succeededCount, totalCount);
  // Update total count with newly created rules
  totalAccessRulesCount.value += succeededCount;
  resetAndTriggerInfiniteScroll();
  emit("access-rules-created", createdAccessRules);
}

function notifyCreatedAccessRules(succeededCount: number, totalCount: number): void {
  const message =
    succeededCount === totalCount
      ? `${succeededCount} access rule(s) created`
      : `${succeededCount}/${totalCount} access rule(s) created`;

  $q.notify(alertUtil.getSuccess(message));
}

function toggleNewAccessRuleForm() {
  isNewAccessRuleFormOpen.value = !isNewAccessRuleFormOpen.value;
}

function closeForm(): void {
  isNewAccessRuleFormOpen.value = false;
}

async function submitForm(): Promise<void> {
  await newAccessRuleFormRef.value?.submit();
}

// Lifecycle hooks
async function initialize() {
  try {
    initialLoader.value = true;
    if (hasReadRolesPermission.value && roleStore.roleList.length === 0) {
      await roleStore.loadRoles();
    }
    if (props.hideSearch) {
      totalAccessRulesCount.value = 0;
    } else {
      totalAccessRulesCount.value = await accessRuleService.getAccessRulesCount(getFilterBy.value);

      const userGroups = Array.isArray(props.formOptions.userGroups) ? props.formOptions.userGroups : [];
      if (userGroups.length > 0) {
        const groupRulesPromises = userGroups.map((groupName) => {
          const groupFilterBy = [
            filterUtil.getEqualsFilterString(AccessRulesSortFilterFields.SubjectId, groupName),
            filterUtil.getEqualsFilterString(AccessRulesSortFilterFields.SubjectType, SubjectType.Group),
          ];
          return accessRuleService.getAccessRulesV2(groupFilterBy, undefined, undefined, 0, 500);
        });
        const groupRulesArrays = await Promise.all(groupRulesPromises);
        allGroupAccessRules.value = groupRulesArrays.flat();
      }

      // all-authenticated-users rules apply to every subject; surface them read-only on the
      // subject-oriented pages (User/Application), where the list is otherwise filtered by subjectId.
      // Isolated so this enrichment failing never aborts the primary rule load.
      if (props.formOptions.subjectId && !props.formOptions.scopeId) {
        try {
          allAuthAccessRulesRaw.value = await accessRuleService.getAccessRulesV2(
            [
              filterUtil.getEqualsFilterString(
                AccessRulesSortFilterFields.SubjectType,
                SubjectTypeExtended.AllAuthenticatedUsers,
              ),
            ],
            undefined,
            undefined,
            0,
            500,
          );
        } catch (error: unknown) {
          handleError(error, ErrorAlert.failedLoadingMessage("access rules"));
        }
      }
    }
    displayConfig.value = accessRuleUtil.getAccessRuleModalDisplayConfig(props.formOptions.page);
    if (props.autoOpenForm) {
      isNewAccessRuleFormOpen.value = true;
    }
  } catch (error: unknown) {
    handleError(error, "Failed to load roles");
  } finally {
    initialLoader.value = false;
  }
}

onMounted(() => {
  initialize();
});

defineExpose({
  accessRulesCount: totalAccessRulesCount,
  isFormOpen: isNewAccessRuleFormOpen,
  isFormDirty,
  closeForm,
  submitForm,
});
</script>

<style scoped lang="scss">
.sub-header {
  margin-top: 5px;
  margin-bottom: 21px;
}
.empty-state {
  min-height: 124px;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
