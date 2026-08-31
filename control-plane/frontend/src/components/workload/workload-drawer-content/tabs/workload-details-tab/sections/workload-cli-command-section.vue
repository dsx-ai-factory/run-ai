<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="workload-cli-command-section column">
    <div class="intro text-b-body1">
      To submit a new workload based on the selected workload, copy the following command line to the terminal:
    </div>
    <runai-copyable-code-block :code="data.command" tooltip-anchor="top middle" />
    <div class="note text-b-body1">CLI command is available only for workloads submitted via the CLI</div>
    <div class="docs-link text-b-body1">
      To learn how to submit a workload via CLI, see the
      <a aid="workload-cli-command-docs-link" target="_blank" rel="noopener noreferrer" :href="DOCS.CLI.SUBMIT">
        CLI Commands Reference
      </a>
    </div>
  </div>
</template>

<script setup lang="ts">
import { DOCS } from "@/constants/docs";
import { RunaiCopyableCodeBlock } from "@/components/common/runai-copyable-code-block";
import type { ICliCommandSectionData } from "@/models/workload-details.model";

defineOptions({ name: "workload-cli-command-section" });

export interface Props {
  data: ICliCommandSectionData;
}

defineProps<Props>();
</script>

<style lang="scss" scoped>
.workload-cli-command-section {
  gap: var(--space-s);

  .note,
  .docs-link {
    color: var(--b-text-sub, #{$black-54});
  }

  .docs-link {
    font-style: italic;

    a {
      color: var(--text-link-default);
      text-decoration: none;

      &:hover {
        text-decoration: underline;
      }
    }
  }
}
</style>
