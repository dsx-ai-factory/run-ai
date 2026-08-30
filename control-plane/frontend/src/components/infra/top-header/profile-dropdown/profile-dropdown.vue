<template>
  <span>
    <q-btn-dropdown
      class="profile-dropdown"
      :class="{ dark: darkMode, 'light-dark': lightDarkMode }"
      :content-class="contentClass"
      flat
      dropdown-icon="none"
      no-icon-animation
      fab-mini
      aid="profile-btn"
    >
      <template #label>
        <runai-avatar :initial="userInitials" />
      </template>
      <div class="profile-dropdown-menu-content">
        <q-list>
          <q-item>
            <q-item-section>
              <q-item-label class="user-details">
                <runai-avatar size="large" :initial="userInitials" class="q-mb-sm" />
                <div class="user-email">{{ userEmail }}</div>
                <div v-if="userFullName" class="text-b-subtitle1 text-b-sub q-mt-xs">{{ userFullName }}</div>
              </q-item-label>
            </q-item-section>
          </q-item>

          <q-separator />
          <q-item clickable v-close-popup @click="openUserSettings">
            <q-item-section>
              <q-item-label class="settings-row">
                <div>Settings</div>
              </q-item-label>
            </q-item-section>
          </q-item>

          <template v-if="!isSSOUser">
            <q-item clickable v-close-popup @click="onChangePassword">
              <q-item-section>
                <q-item-label class="change-password-row">
                  <div>Change password</div>
                </q-item-label>
              </q-item-section>
            </q-item>
          </template>

          <q-item clickable v-close-popup @click="onSignOut">
            <q-item-section>
              <q-item-label class="logout-row" aid="sign-out-btn">
                <div>Sign out</div>
              </q-item-label>
            </q-item-section>
          </q-item>
        </q-list>
      </div>
    </q-btn-dropdown>
    <q-tooltip class="q-pa-sm" :class="{ 'tooltip-dark': !lightDarkMode }">
      <div>{{ userEmail }}</div>
      <div v-if="userFullName">{{ userFullName }}</div>
    </q-tooltip>

    <runai-change-password-modal v-if="showChangePassModal" @close="showChangePassModal = false" />
  </span>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmps
import { RunaiChangePasswordModal } from "@/components/common/runai-change-password-modal";
import { RunaiAvatar } from "@/components/common/runai-avatar";

// stores
import { useAuthStore } from "@/stores/auth.store";

export default defineComponent({
  components: {
    RunaiChangePasswordModal,
    RunaiAvatar,
  },
  props: {
    darkMode: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    lightDarkMode: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      authStore: useAuthStore(),
      showChangePassModal: false as boolean,
    };
  },
  computed: {
    userEmail(): string {
      return this.authStore.currentUser.username || "";
    },
    userInitials(): string {
      return this.authStore.userInitials;
    },
    userFullName(): string {
      return this.authStore.userFullName;
    },
    isSSOUser(): boolean {
      return this.authStore.isCurrentUserSSO;
    },
    contentClass(): string {
      return this.lightDarkMode ? "profile-dropdown-menu-container light-dark" : "profile-dropdown-menu-container";
    },
  },
  methods: {
    onSignOut(): void {
      this.authStore.logout();
    },
    onChangePassword(): void {
      this.showChangePassModal = true;
    },
    openUserSettings(): void {
      this.$router.push("/user-settings");
    },
  },
});
</script>

<style lang="scss">
.profile-dropdown {
  .q-icon {
    display: none;
  }

  &.q-btn {
    background: transparent;
    min-height: 36px;
    min-width: 36px;
    padding: 0;
  }
}

.profile-dropdown-menu-container {
  background: $dark-gray !important;
  color: $white;

  &.light-dark {
    background: $white !important;
    color: $black-87;
  }

  .profile-dropdown-menu-content {
    width: 300px;

    .user-details {
      padding: 15px 0;
      display: flex;
      justify-content: center;
      flex-direction: column;
      align-items: center;
    }
  }
}
</style>
