<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-addable-multi-select
    class="subject-addable-multi-select"
    dense-chip
    server-side-filter
    :label="subjectSelectLabel"
    :new-option-label="newSubjectOptionLabel"
    @options-select="updateSelectedSubjectsOptions"
    @input="onSearchInput"
    @add-option="addEmail"
    :options="combinedEmailOptions"
    :selected-options="selectedSubjectsOptions"
    :loading="loadingEmails"
    :max-values="subjectMaxAmount"
    :hint="hint"
    :can-search="canSearchSubjects"
    required
    counter
  >
    <template v-slot:selected-item="scope">
      <q-chip :removable="!scope.isDisabled" @remove="scope.removeAtIndex(scope.index)" :tabindex="scope.tabindex">
        <span v-if="scope.opt.primaryLabel" class="text-b-body1-md">{{ scope.opt.primaryLabel }}</span>
        <span v-if="scope.opt.secondaryLabel" class="text-b-body1">&nbsp;{{ scope.opt.secondaryLabel }}</span>
        <span v-if="!scope.opt.primaryLabel" class="text-b-body1">{{ scope.opt.label }}</span>
      </q-chip>
    </template>
    <template v-slot:option="scope">
      <q-item v-bind="scope.itemProps">
        <q-item-section>
          <q-item-label v-if="scope.opt.primaryLabel">
            <span class="text-b-body1">{{ scope.opt.primaryLabel }}</span>
            <span v-if="scope.opt.secondaryLabel" class="text-b-body1 text-b-sub"
              >&nbsp;{{ scope.opt.secondaryLabel }}</span
            >
          </q-item-label>
          <q-item-label v-else class="text-b-body1">{{ scope.opt.label }}</q-item-label>
        </q-item-section>
      </q-item>
    </template>
  </runai-addable-multi-select>
</template>
<script lang="ts">
import { defineComponent, ref, computed, watch, onMounted, type PropType } from "vue";
import { usePermissionStore } from "@/stores/permissions.store";
import { RunaiAddableMultiSelect } from "@/components/common/runai-addable-multi-select";
import type { ISelectOption } from "@/models/global.model";
import userService from "@/services/control-plane/user.service/user.service";
import { SubjectType, ResourceType, Action } from "@/swagger-models/authorization-client";
import type { ServiceAccount } from "@/swagger-models/identity-manager-client";
import { UsersFilterSortFields } from "@/swagger-models/identity-manager-client";
import { debounce, useQuasar } from "quasar";
import { accessRuleService } from "@/services/control-plane/rbac/access-rule.service/access-rule.service";
import { serviceAccountsService } from "@/services/control-plane/service-accounts.service/service-accounts.service";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { stringUtil } from "@/utils/string.util";
import { userUtils } from "@/utils/user.util";
import { ErrorAlert } from "@/utils/error-alert.util";

const SUBJECT_MAX_AMOUNT = 10;
const INITIAL_FETCH_LIMIT = 10;
const SEARCH_DEBOUNCE_TIME = 300;
const MIN_SEARCH_LENGTH = 2;

interface ISubjectOption extends ISelectOption {
  primaryLabel?: string;
  secondaryLabel?: string;
}

export default defineComponent({
  name: "subject-addable-multi-select",
  components: { RunaiAddableMultiSelect },
  emits: ["update:selected-subjects", "add-subject"],
  props: {
    subjectType: {
      type: Object as PropType<ISelectOption | null>,
      required: true,
    },
    selectedSubjectsOptions: {
      type: Array as PropType<ISelectOption[]>,
      default: () => [],
    },
  },
  setup(props, { emit }) {
    const $q = useQuasar();
    const searchQuery = ref("");
    const loadingEmails = ref(false);
    const manualOptions = ref<ISubjectOption[]>([]);
    const fetchedEmailOptions = ref<ISubjectOption[]>([]);
    const cachedServiceAccounts = ref<ServiceAccount[]>([]);
    const permissionStore = usePermissionStore();

    const canReadUsers = permissionStore.hasPermission(ResourceType.Users, Action.Read);
    const canReadServiceAccounts = permissionStore.hasPermission(ResourceType.ServiceAccount, Action.Read);

    const canSearchSubjects = computed(() => {
      const subjectType = props.subjectType?.value as SubjectType;
      if (subjectType === SubjectType.User && !canReadUsers) {
        return false;
      }
      if (subjectType === SubjectType.ServiceAccount && !canReadServiceAccounts) {
        return false;
      }
      return true;
    });

    const combinedEmailOptions = computed(() => [...manualOptions.value, ...fetchedEmailOptions.value]);

    const newSubjectOptionLabel = computed(() => {
      if (!props.subjectType?.label) {
        return "";
      }
      return props.subjectType.label;
    });

    const subjectSelectLabel = computed(() => {
      switch (props.subjectType?.value) {
        case SubjectType.User:
          return "User(s) email";
        case SubjectType.ServiceAccount:
          return "Service account(s) name";
        case SubjectType.Group:
          return "Group(s) name";
        default:
          return "";
      }
    });

    const hint = computed(() => {
      const suffix = stringUtil.toLowerCase(props.subjectType?.label || "") + "s";
      return `Select up to ${SUBJECT_MAX_AMOUNT} ${suffix}`;
    });

    const updateSelectedSubjectsOptions = (selectedSubjectsOptions: ISelectOption[]): void => {
      emit("update:selected-subjects", selectedSubjectsOptions);
    };

    const addEmail = (name: string, doneCallback: () => void): void => {
      const newEmailOption: ISelectOption = { label: name, value: name };
      manualOptions.value.push(newEmailOption);

      setTimeout(() => {
        emit("add-subject", newEmailOption);
        searchQuery.value = "";
      }, 0);
      doneCallback();
    };

    const isSSO = computed(() => props.subjectType?.value === SubjectType.Group);
    const isServiceAccount = computed(() => props.subjectType?.value === SubjectType.ServiceAccount);

    const onSearchInput = (event: Event): void => {
      if (!canSearchSubjects.value) {
        return;
      }
      const target = event.target as HTMLInputElement;
      const query = target.value;
      searchQuery.value = query;
      const trimmed = query.trim();
      const isUser = !isSSO.value && !isServiceAccount.value;
      const willFetch = !(isUser && trimmed.length > 0 && trimmed.length < MIN_SEARCH_LENGTH);
      if (willFetch) {
        loadingEmails.value = true;
      } else {
        loadingEmails.value = false;
      }
      debouncedFetchUsers(query);
    };

    const fetchUsers = async (query = ""): Promise<void> => {
      const trimmedQuery = query.trim();
      // For users: allow empty query (initial fetch) but skip 1-char partial queries.
      // Service accounts use client-side filtering so no minimum length applies.
      const isUser = !isSSO.value && !isServiceAccount.value;
      if (isUser && trimmedQuery.length > 0 && trimmedQuery.length < MIN_SEARCH_LENGTH) {
        return;
      }
      try {
        loadingEmails.value = true;

        let newOptions: ISubjectOption[];
        if (isSSO.value) {
          const subjectIds = await accessRuleService.getSsoGroupSubjectIds(trimmedQuery);
          newOptions = subjectIds.map((id) => ({ label: id, value: id }));
        } else if (isServiceAccount.value) {
          // Fetch once and cache; subsequent keystrokes filter the cached list client-side
          if (!cachedServiceAccounts.value.length) {
            cachedServiceAccounts.value = await serviceAccountsService.list();
          }
          newOptions = cachedServiceAccounts.value
            .filter((sa) => !trimmedQuery || sa.name.toLowerCase().includes(trimmedQuery.toLowerCase()))
            .map((sa) => ({ label: sa.name, value: sa.name }));
        } else {
          const users = trimmedQuery
            ? await userService.getUsersMinimal(trimmedQuery)
            : await userService.getUsersMinimal(
                undefined,
                // username=@@ (contains "@") matches all email-format usernames.
                // This assumes the identity provider uses email addresses as usernames (e.g. OIDC/Keycloak).
                // Users with non-email usernames (e.g. LDAP cn=...) will not appear in the initial list.
                [filterUtil.getContainsFilterString(UsersFilterSortFields.Username, "@")],
                undefined,
                INITIAL_FETCH_LIMIT,
              );
          newOptions = users.map((user) => {
            const display = userUtils.getUserDisplayLabel(user);
            return {
              label: display.label,
              value: user.username,
              primaryLabel: display.primaryLabel,
              secondaryLabel: display.secondaryLabel,
            };
          });
        }

        fetchedEmailOptions.value = trimmedQuery ? newOptions : newOptions.slice(0, INITIAL_FETCH_LIMIT);
      } catch (error) {
        const entity = isSSO.value ? "groups" : isServiceAccount.value ? "service accounts" : "users";
        const errorAlert = new ErrorAlert({ generalMessage: ErrorAlert.failedLoadingMessage(entity) });
        $q.notify(errorAlert.getNotification(error));
      } finally {
        loadingEmails.value = false;
      }
    };
    const debouncedFetchUsers = debounce(fetchUsers, SEARCH_DEBOUNCE_TIME);

    const resetAndPrefetch = (): void => {
      manualOptions.value = [];
      fetchedEmailOptions.value = [];
      cachedServiceAccounts.value = [];
      searchQuery.value = "";
      emit("update:selected-subjects", []);
      fetchUsers("");
    };

    watch(() => props.subjectType, resetAndPrefetch);

    onMounted(() => {
      if (props.subjectType) {
        fetchUsers("");
      }
    });

    return {
      searchQuery,
      loadingEmails,
      subjectMaxAmount: SUBJECT_MAX_AMOUNT,
      combinedEmailOptions,
      newSubjectOptionLabel,
      subjectSelectLabel,
      hint,
      canSearchSubjects,
      updateSelectedSubjectsOptions,
      addEmail,
      onSearchInput,
      fetchUsers,
    };
  },
});
</script>

<style lang="scss">
.subject-addable-multi-select.dense-chip .q-select .q-chip {
  padding-right: 2px;

  span {
    position: relative;
    top: 1px;
  }

  .q-chip__icon--remove {
    font-size: 14px;
    margin: 0 0 0 4px;
  }
}
</style>
