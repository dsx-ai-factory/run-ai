/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { CallbackFunction } from "@/models/global.model";
import { makeId } from "@/utils/common.util";

type ListenerObj = {
  cb: CallbackFunction;
  id: string;
};

type EventMap = Map<string, Array<ListenerObj>>;
const listeners: EventMap = new Map<string, Array<ListenerObj>>();

export const eventEmitter = {
  addListener,
  removeListener,
  emit,
};

function addListener(event: string, cb: CallbackFunction): string {
  const id: string = makeId();
  listeners.set(event, _getListeners(event).concat({ cb, id }));
  return id;
}

function removeListener(event: string, listenerId: string): void {
  listeners.set(
    event,
    _getListeners(event).filter((listener: ListenerObj) => listener.id !== listenerId),
  );
}

function emit(event: string, data?: unknown): void {
  _getListeners(event).forEach((listener: ListenerObj) => listener.cb(data));
}

function _getListeners(event: string): Array<ListenerObj> {
  return listeners.get(event) || [];
}
