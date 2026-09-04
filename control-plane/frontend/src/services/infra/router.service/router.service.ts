/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { NAVIGATION_EVENTS } from "@/common/event.constant";
import { eventEmitter } from "@/services/infra/event-emitter.service/event-emitter.service";

export async function requestToLeave(): Promise<boolean> {
  return new Promise((resolve) => {
    const listenerId = eventEmitter.addListener(
      NAVIGATION_EVENTS.LEAVE_PAGE_STATUS,
      (data: Record<string, unknown>): void => {
        eventEmitter.removeListener(NAVIGATION_EVENTS.LEAVE_PAGE_STATUS, listenerId);
        resolve(!!data.allowToLeave);
      },
    );
    eventEmitter.emit(NAVIGATION_EVENTS.LEAVE_PAGE_REQUEST);
  });
}
