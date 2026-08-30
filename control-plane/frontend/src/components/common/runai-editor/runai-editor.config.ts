import type { editor } from "monaco-editor";

// Workers
import EditorWorker from "monaco-editor/esm/vs/editor/editor.worker?worker";
import CssWorker from "monaco-editor/esm/vs/language/css/css.worker?worker";
import HtmlWorker from "monaco-editor/esm/vs/language/html/html.worker?worker";
import YamlWorker from "./yaml.worker.js?worker";

export const RUNAI_DARK_THEME = "runai-dark";

/**
 * Initialize Monaco environment with workers
 */
export function initMonacoEnvironment(): void {
  window.MonacoEnvironment = {
    getWorker(_, label) {
      switch (label) {
        case "css":
          return new CssWorker();
        case "html":
          return new HtmlWorker();
        case "yaml":
          return new YamlWorker();
        default:
          return new EditorWorker();
      }
    },
  };
}

/**
 * Custom Run:ai dark theme definition
 */
export const runaiDarkTheme: editor.IStandaloneThemeData = {
  base: "vs-dark",
  inherit: true,
  rules: [],
  colors: {
    // Monaco takes literal colors here, so this must stay in sync with the
    // --bg-code-surface CSS var (src/assets/scss/vendor/_quasar-variables.scss).
    "editor.background": "#333333",
    "editor.lineHighlightBackground": "#3c3c3c",
    "editor.lineHighlightBorder": "#3c3c3c",
  },
};

export interface IEditorOptionsParams {
  code: string;
  language: string;
  readOnly: boolean;
  showPlaceholder: boolean;
}

/**
 * Create Monaco editor options
 */
export function createEditorOptions(params: IEditorOptionsParams): editor.IStandaloneEditorConstructionOptions {
  const { code, language, readOnly, showPlaceholder } = params;

  // Determine if editor should show line numbers initially
  const hasInitialContent = !!code;
  const shouldHideLineNumbers = showPlaceholder && !hasInitialContent;

  return {
    value: code,
    theme: RUNAI_DARK_THEME,
    fixedOverflowWidgets: true,
    readOnly,
    automaticLayout: true,
    language,
    fontFamily: '"Roboto Mono", SFMono-Regular, Consolas, Menlo, monospace',
    fontSize: 12,
    lineNumbers: shouldHideLineNumbers ? "off" : "on",
    renderLineHighlight: shouldHideLineNumbers ? "none" : "line",
    glyphMargin: false,
    folding: !shouldHideLineNumbers,
    padding: { top: 8 },
    lineDecorationsWidth: shouldHideLineNumbers ? 30 : 0,
    scrollbar: {
      vertical: shouldHideLineNumbers ? "hidden" : "auto",
      horizontal: shouldHideLineNumbers ? "hidden" : "auto",
      verticalScrollbarSize: shouldHideLineNumbers ? 0 : 14,
      horizontalScrollbarSize: shouldHideLineNumbers ? 0 : 10,
    },
  };
}

/**
 * Get editor options for content change (toggle line numbers based on content)
 */
export function getContentChangeOptions(hasContent: boolean): editor.IEditorOptions {
  return {
    lineNumbers: hasContent ? "on" : "off",
    renderLineHighlight: hasContent ? "line" : "none",
    glyphMargin: false,
    folding: hasContent,
    lineDecorationsWidth: hasContent ? 0 : 30,
    scrollbar: {
      vertical: hasContent ? "auto" : "hidden",
      horizontal: hasContent ? "auto" : "hidden",
      verticalScrollbarSize: hasContent ? 14 : 0,
      horizontalScrollbarSize: hasContent ? 10 : 0,
    },
  };
}
