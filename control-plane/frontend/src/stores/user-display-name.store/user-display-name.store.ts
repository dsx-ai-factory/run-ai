/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { defineStore } from "pinia";
import { userService } from "@/services/control-plane/user.service/user.service";
import { userUtils, type IUserDisplayLabel } from "@/utils/user.util/user.util";

// Matches backend UsersMinimalMaxFilterBy.
const BATCH_SIZE = 25;
const REQUEST_TIMEOUT_MS = 3000;

function isPermanentFailure(err: unknown): boolean {
  const status =
    (err as { response?: { status?: number }; status?: number })?.response?.status ??
    (err as { status?: number })?.status;
  return status === 403 || status === 404;
}

export interface IUserDisplayInfo {
  firstName?: string | null;
  lastName?: string | null;
}

// Cache is session-lifetime — full page navigation clears it.
interface IUserDisplayNameState {
  cacheByUsername: Record<string, IUserDisplayInfo | null>;
  pendingUsernames: Record<string, true>;
  // Subset of pendingUsernames not yet sent, so in-flight names aren't refetched.
  queuedUsernames: Record<string, true>;
  isFlushScheduled: boolean;
}

export const useUserDisplayNameStore = defineStore("UserDisplayName", {
  state: (): IUserDisplayNameState => ({
    cacheByUsername: Object.create(null) as Record<string, IUserDisplayInfo | null>,
    pendingUsernames: Object.create(null) as Record<string, true>,
    queuedUsernames: Object.create(null) as Record<string, true>,
    isFlushScheduled: false,
  }),
  getters: {
    isResolved(state) {
      return (username: string): boolean => Object.prototype.hasOwnProperty.call(state.cacheByUsername, username);
    },
    isPending(state) {
      return (username: string): boolean => state.pendingUsernames[username] === true;
    },
    getDisplayInfo(state) {
      return (username: string): IUserDisplayInfo | null | undefined => state.cacheByUsername[username];
    },
    getDisplayLabel(state) {
      return (username: string): IUserDisplayLabel | null => {
        const info = state.cacheByUsername[username];
        if (!info) return null;
        return userUtils.getUserDisplayLabel({ username, firstName: info.firstName, lastName: info.lastName });
      };
    },
  },
  actions: {
    resolveUsernames(usernames: Array<string>): void {
      const toFetch = usernames.filter((u) => u && !this.isResolved(u) && !this.isPending(u));
      if (toFetch.length === 0) return;

      toFetch.forEach((u) => {
        this.pendingUsernames[u] = true;
        this.queuedUsernames[u] = true;
      });

      if (!this.isFlushScheduled) {
        this.isFlushScheduled = true;
        void Promise.resolve().then(() => this._flush());
      }
    },
    async _flush(): Promise<void> {
      const usernames = Object.keys(this.queuedUsernames);
      this.queuedUsernames = Object.create(null) as Record<string, true>;
      this.isFlushScheduled = false;
      if (usernames.length === 0) return;

      const batches: Array<Array<string>> = [];
      for (let i = 0; i < usernames.length; i += BATCH_SIZE) {
        batches.push(usernames.slice(i, i + BATCH_SIZE));
      }
      await Promise.all(batches.map((batch) => this._fetchBatch(batch)));
    },
    async _fetchBatch(usernames: Array<string>): Promise<void> {
      try {
        const users = await userService.resolveUsersByUsernames(usernames, REQUEST_TIMEOUT_MS);
        const byUsername = new Map(users.map((u) => [u.username, u]));
        const patch: Record<string, IUserDisplayInfo | null> = Object.create(null);
        usernames.forEach((u) => {
          const hit = byUsername.get(u);
          patch[u] = hit ? { firstName: hit.firstName, lastName: hit.lastName } : null;
        });
        Object.assign(this.cacheByUsername, patch);
      } catch (err) {
        // 403/404 won't change this session — cache as null so we don't retry-storm.
        // Transient errors (timeout, 5xx, network) intentionally NOT cached so the
        // next render of the same cell retries automatically.
        if (isPermanentFailure(err)) {
          const patch: Record<string, IUserDisplayInfo | null> = Object.create(null);
          usernames.forEach((u) => {
            if (!this.isResolved(u)) patch[u] = null;
          });
          Object.assign(this.cacheByUsername, patch);
        }
      } finally {
        usernames.forEach((u) => delete this.pendingUsernames[u]);
      }
    },
  },
});
