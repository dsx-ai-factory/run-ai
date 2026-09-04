/* eslint-disable */

import { fileURLToPath, URL } from "url";
import { resolve } from "path";

import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import { quasar, transformAssetUrls } from "@quasar/vite-plugin";
import svgLoader from "vite-svg-loader";
import { createSvgIconsPlugin } from "vite-plugin-svg-icons";
import monacoEditorPlugin from "vite-plugin-monaco-editor";

const __dirname = fileURLToPath(new URL(".", import.meta.url));

const environment = process.env.ENVIRONMENT;
const env_port = process.env.RUN_AI_FRONTEND_PORT;
const env_host = process.env.RUN_AI_FRONTEND_HOST;

const target = process.env.RUN_AI_BACKEND_URL
  ? process.env.RUN_AI_BACKEND_URL
  : "http://localhost:" + (env_port || 8080);

const backendProxy = {
  target,
  headers: env_host ? { Host: env_host } : undefined,
};

console.log("Build environment: ", environment);

const monacoPrefix = `monaco-editor/esm/vs`;

// https://vitejs.dev/config/
export default defineConfig({
  build: {
    outDir: "./dist/frontend",
    sourcemap: false,
    rollupOptions: {
      output: {
        manualChunks: {
          cssWorker: [`${monacoPrefix}/language/css/css.worker`],
          htmlWorker: [`${monacoPrefix}/language/html/html.worker`],
          editorWorker: [`${monacoPrefix}/editor/editor.worker`],
        },
      },
    },
  },
  plugins: [
    vue({
      template: { transformAssetUrls },
    }),
    quasar({
      sassVariables: resolve(__dirname, "src/assets/scss/vendor/_quasar-variables.scss"),
    }),
    svgLoader({ svgo: false }),
    // Bundles every SVG in src/assets/icons/runai-glyph/ into a single sprite registered as
    // `virtual:svg-icons-register`. Flat files keep symbol IDs like `runai-access`; nested
    // variant files such as `filled/access.svg` become `runai-filled-access`.
    // Glyph files author their main paths with `currentColor`, and filled icons use
    // `--runai-glyph-cutout` for internal white details. Keep those colors intact.
    createSvgIconsPlugin({
      iconDirs: [resolve(__dirname, "src/assets/icons/runai-glyph")],
      symbolId: "runai-[dir]-[name]",
      svgoOptions: {
        plugins: [{ name: "preset-default" }],
      },
    }),
    monacoEditorPlugin({ customWorkers: [{ label: "yaml", entry: "monaco-yaml/yaml.worker.js" }] }),
  ],
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
    },
  },
  server: {
    port: env_port || 9000,
    host: env_host || "127.0.0.1",

    proxy: {
      "^/v1": backendProxy,
      "^/_healthy": backendProxy,
      "^/api/v1": backendProxy,
      "^/api/v2": backendProxy,
      "^/api/docs": backendProxy,
      "^/_version": backendProxy,
      "^/cluster-api": backendProxy,
    },
  },
  css: {
    preprocessorOptions: {
      scss: {
        api: "modern-compiler",
        additionalData: (content, loaderContext) => {
          // Use relative paths from the project root
          const scssPath = resolve(__dirname, "src/assets/scss");
          return `
            @import "${resolve(scssPath, "base/_typography.scss")}";
            @import "${resolve(scssPath, "abstracts/_mixin.scss")}";
            @import "${resolve(scssPath, "vendor/_branding.scss")}";
            ${content}
          `;
        },
        silenceDeprecations: ["legacy-js-api"],
        // Add loadPaths so 'src/' imports can be resolved
        loadPaths: [resolve(__dirname, ".")],
      },
      sass: {
        api: "modern-compiler",
        silenceDeprecations: ["legacy-js-api"],
        loadPaths: [resolve(__dirname, ".")],
      },
    },
  },
  test: {
    globals: true,
    environment: "jsdom",
    setupFiles: ["./unit-test/setup.ts"],
    // Vitest still prints console output to stderr even when unit-test/setup.ts mocks console.error.
    // Suppress expected axios network failures logged by ErrorAlert during component tests.
    onConsoleLog(log) {
      if (
        log.includes("Network Error") ||
        log.includes("[status-code: -1]") ||
        log.includes("Error: connect ECONNREFUSED")
      ) {
        return false;
      }
    },
  },
});
