/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { UserPostResponse } from "@/swagger-models/identity-manager-client";
import { tableUtil } from "@/utils/table.util/table.util";
import { ErrorAlert } from "@/utils/error-alert.util";
import { Notify } from "quasar";
import { alertUtil } from "@/utils/alert.util";
import type { ITableColumn } from "@/models/table.model";

export interface DownloadCredentialsOptions {
  createdUsers?: UserPostResponse[];
  teamName?: string;
}

export function useDownloadCredentials() {
  const downloadCredentials = async (options: DownloadCredentialsOptions) => {
    try {
      let dataToExport: UserPostResponse[];

      if (!options.createdUsers || options.createdUsers.length === 0) {
        throw new Error("No user credentials available to export. Please complete the team setup process first.");
      } else {
        dataToExport = options.createdUsers;
      }

      const columns: ITableColumn[] = [
        { name: "username", label: "Username", field: "username", align: "left" },
        { name: "tempPassword", label: "Temporary Password", field: "tempPassword", align: "left" },
      ];

      tableUtil.exportTableAsCsv(`${options.teamName || "team"}-credentials`, dataToExport, columns);

      Notify.create(alertUtil.getSuccess("Team credentials downloaded successfully"));
    } catch (error: unknown) {
      console.error("Failed to download credentials:", error);
      const errorAlert = new ErrorAlert({ generalMessage: "Failed to export team credentials" });
      Notify.create(errorAlert.getNotification(error));
    }
  };

  return {
    downloadCredentials,
  };
}
