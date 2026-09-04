/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { IConnectionAccess } from "@/components/environment/connection-access/connection-access-modal/connection-access-modal.model";
import type { ISelectOption } from "@/models/global.model";
import { ServingPortInfoProtocolEnum } from "@/swagger-models/assets-service-client";

export const ProtocolOptions: Array<ISelectOption> = [
  {
    label: "HTTP",
    value: ServingPortInfoProtocolEnum.Http,
  },
  {
    label: "gRPC",
    value: ServingPortInfoProtocolEnum.Grpc,
  },
];

export interface IServingEndpointAccessData {
  access?: IConnectionAccess;
  protocol?: string;
  port?: number;
  clusterLocalAccessOnly?: boolean | null;
}

export enum EServingEndpointAccessEvents {
  UpdateModelValue = "update:model-value",
}

export const maxValidPort = 65535;
export const minValidPort = 1;

export const ENDPOINT_TYPE_EXTERNAL = "external";
export const ENDPOINT_TYPE_INTERNAL = "internal";
