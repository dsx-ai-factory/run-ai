/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { debounce } from "quasar";
import { useAuthStore } from "@/stores/auth.store";

// utils
import { storageUtil } from "@/utils/storage.util";
import { intervalUtil } from "@/utils/interval.util";

// consts
import { EIntervalLabels } from "@/models/interval.model";
import { DEFAULT_BROWSER_SESSION_TIMEOUT_MINUTES } from "@/models/setting.model";

interface IEventListener {
  event: string;
  callback: EventListenerOrEventListenerObject;
}

interface IFunction {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  (...argArray: any[]): any;
}

const second = 1000;
const minute = 60 * second;

export class IdleTimer {
  private static instance: IdleTimer | null = null;

  // if you change the events list here, please make sure you udpate the
  // list in v1 as well
  events: Array<string> = ["mousemove", "scroll", "keydown", "click"];
  callbacks: Record<string, Array<IEventListener>> = {};
  eventHandler: IFunction;

  timeout: number = 60 * minute;
  timeoutTracker: number | undefined;
  expiredTimeName = "_expiredTime";

  private constructor() {
    // binding the handler because we use this inside the fanction to get the timeout value
    this.eventHandler = this.updateExpiredTime.bind(this);
    this.cleanUp();
  }

  public static getInstance(): IdleTimer {
    if (!IdleTimer.instance) {
      IdleTimer.instance = new IdleTimer();
    }
    return IdleTimer.instance;
  }

  async track(minutesToLogout: number): Promise<void> {
    if (minutesToLogout === DEFAULT_BROWSER_SESSION_TIMEOUT_MINUTES) {
      this.cleanUp();
      return;
    }

    this.timeout = minutesToLogout * minute;

    // add default listeners
    this.addListener("default", window.document);

    this.updateExpiredTime();
    this.startInterval();
  }

  async updateTimeout(minutesToLogout: number): Promise<void> {
    this.cleanUp();
    await this.track(minutesToLogout);
  }

  onTimeout(): void {
    this.cleanUp();
    useAuthStore().logout();
  }

  private cleanUp(): void {
    // remove default listeners
    this.removeListener("default", window.document);

    // window.clearInterval(this.interval);
    intervalUtil.stopInterval(EIntervalLabels.IdleTimer);
    storageUtil.remove(this.expiredTimeName);
  }

  private addListener(name: string, doc: Document) {
    if (this.callbacks[name]) return;

    this.callbacks[name] = this.events.map((event: string) => {
      const callback: EventListenerOrEventListenerObject = this.getHandler();
      doc.addEventListener(event, callback);
      return { event, callback };
    });
  }

  private removeListener(name: string, doc: Document) {
    if (!this.callbacks[name]) return;

    const eventListeners: Array<IEventListener> = this.callbacks[name];
    eventListeners.forEach((listener: { event: string; callback: EventListenerOrEventListenerObject }) => {
      doc.removeEventListener(listener.event, listener.callback);
    });

    delete this.callbacks[name];
  }

  private getHandler(): EventListenerOrEventListenerObject {
    return debounce(this.eventHandler, second); // the bounce here is an optimization only
  }

  private startInterval(): void {
    intervalUtil.stopInterval(EIntervalLabels.IdleTimer);
    intervalUtil.startInterval(EIntervalLabels.IdleTimer, () => {
      // the default  `${Date.now() + 60 * minute}` is for the edge case
      // when you clear the value from local storage to set new expired time
      // and the interval cycle is trying to load the expired time before the new one is set
      const expiredTime = parseInt(storageUtil.get(this.expiredTimeName) || `${Date.now() + 60 * minute}`, 10);
      if (expiredTime < Date.now()) {
        this.onTimeout();
      }
    });
  }

  private updateExpiredTime(): void {
    if (this.timeoutTracker) {
      window.clearTimeout(this.timeoutTracker);
    }

    // adding kind of delay to save the new expired time.
    this.timeoutTracker = window.setTimeout(() => {
      const minutesUntilExpiry: number = Date.now() + this.timeout;
      storageUtil.save(this.expiredTimeName, `${minutesUntilExpiry}`);
    }, 300);
  }
}
