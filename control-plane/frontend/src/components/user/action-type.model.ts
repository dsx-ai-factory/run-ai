/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

export enum ActionTypeEnum {
  CREATE = "create",
  RESET = "reset",
  REGENERATE = "regenerate",
  DONE = "done",
}

export type TActionTypes = ActionTypeEnum.CREATE | ActionTypeEnum.RESET;
