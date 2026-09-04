<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="runai-editor-wrapper" :style="{ height }">
    <q-linear-progress v-if="!initComplete" animation-speed="100" indeterminate />
    <div
      ref="runai-editor-container"
      v-show="initComplete"
      class="runai-editor-container"
      :class="{ readOnly }"
      :style="{ height }"
      @keydown.enter.stop
    />

    <!-- Placeholder overlay when editor is empty -->
    <div v-if="showPlaceholder && initComplete && !currentCode" class="editor-placeholder text-b-body1">
      {{ placeholder }}
    </div>

    <q-btn
      v-if="showExpandButton"
      flat
      dense
      round
      :icon="isExpanded ? 'ph-bold ph-arrows-in' : 'ph-bold ph-arrows-out'"
      size="sm"
      aid="toggle-expand-btn"
      @click="$emit('toggle-expand')"
      class="expand-btn runai-icon-btn--default"
    >
      <q-tooltip anchor="center right" self="center left">{{
        isExpanded ? "Minimize window" : "Maximize window"
      }}</q-tooltip>
    </q-btn>
  </div>
</template>
<script lang="ts">
import { defineComponent, type PropType, markRaw } from "vue";
import type { editor } from "monaco-editor";
// utils
import { yamlUtil } from "@/utils/yaml.util";
// config
import {
  initMonacoEnvironment,
  runaiDarkTheme,
  createEditorOptions,
  getContentChangeOptions,
  RUNAI_DARK_THEME,
} from "./runai-editor.config";

initMonacoEnvironment();

export default defineComponent({
  emits: ["changed", "validation-status", "editor-did-mount", "toggle-expand", "pasted"],
  props: {
    code: {
      type: String as PropType<string>,
      default: "",
    },
    language: {
      type: String as PropType<"yaml">,
      default: "yaml",
    },
    height: {
      type: String as PropType<string>,
      default: "350px",
    },
    readOnly: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    showExpandButton: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    isExpanded: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    showPlaceholder: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    placeholder: {
      type: String as PropType<string>,
      default: "Or paste your YAML content here",
    },
  },
  data() {
    return {
      container: "container",
      editor: null,
      editorInstance: null as editor.IStandaloneCodeEditor | null,
      initComplete: false,
      currentCode: this.code,
    };
  },
  async mounted() {
    this.$nextTick(async () => {
      await this.initEditor();
    });
  },
  methods: {
    async initEditor(): Promise<void> {
      // @ts-ignore
      await import("monaco-editor/esm/vs/editor/editor.all.js");
      await import("monaco-editor/esm/vs/basic-languages/yaml/yaml.contribution");
      const monaco = await import("monaco-editor/esm/vs/editor/editor.api");
      const configureMonacoYaml = (await import("monaco-yaml")).configureMonacoYaml;

      configureMonacoYaml(monaco);
      monaco.editor.defineTheme(RUNAI_DARK_THEME, runaiDarkTheme);

      const options = createEditorOptions({
        code: this.code,
        language: this.language,
        readOnly: this.readOnly,
        showPlaceholder: this.showPlaceholder,
      });

      const el: HTMLElement = this.$refs["runai-editor-container"] as HTMLElement;
      const editor = monaco.editor.create(el, options);

      // Use markRaw to prevent Vue from wrapping Monaco in a reactive Proxy
      this.editorInstance = markRaw(editor);

      editor.onDidChangeModelContent((event) => {
        const value: string = editor.getValue();
        this.currentCode = value;

        // Toggle line numbers and highlight based on content
        if (this.showPlaceholder) {
          editor.updateOptions(getContentChangeOptions(!!value));
        }

        this.emitValidState(value);
        if (this.code !== value) {
          this.$emit("changed", value, event);
        }
      });

      // Detect paste events
      editor.onDidPaste(() => {
        const value: string = editor.getValue();
        this.$emit("pasted", value);
      });

      this.initComplete = true;
      this.$emit("editor-did-mount");
    },
    emitValidState(code: string): void {
      const error = yamlUtil.validateYaml(code);
      if (!error) {
        this.$emit("validation-status", { isValid: true });
      } else {
        this.$emit("validation-status", { isValid: false, error });
      }
    },
    setContent(content: string): void {
      // If editor isn't ready yet, just update local state; editor will initialize with `code`.
      if (!this.editorInstance) {
        this.currentCode = content;
        return;
      }
      this.editorInstance.setValue(content);
    },
  },
});
</script>
<style scoped lang="scss">
.runai-editor-wrapper {
  position: relative;
  border-radius: var(--b-border-radius-radius-xs);
  overflow: hidden;
}

.readOnly {
  opacity: 0.8;
}

.expand-btn {
  position: absolute;
  bottom: var(--space-s);
  right: var(--space-s);
  z-index: 5;
  width: 40px;
  height: 40px;
  color: var(--text-inverse);
}

.editor-placeholder {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  color: var(--text-inverse-edit);
  pointer-events: none;
  z-index: 1;
}
</style>
