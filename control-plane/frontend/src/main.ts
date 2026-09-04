/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Registers the runai-glyph SVG sprite injected by vite-plugin-svg-icons.
// Side-effect import only — exposes <symbol id="runai-<name>"> elements consumed by iconMapFn.
import "virtual:svg-icons-register";

import "@quasar/extras/roboto-font/roboto-font.css";
import "quasar/src/css/index.sass";

import "@/assets/scss/main.scss";
import { loadBootConfig } from "@/services/infra/boot-config.service/boot-config.service";

async function bootstrap(): Promise<void> {
  const bootConfig = await loadBootConfig();
  await import("./plugins");

  const [
    { createApp },
    { createPinia },
    { Quasar, Notify },
    { useDirectives },
    { VueDOMPurifyHTML, VueDOMPurifyHTMLOptions },
    { default: App },
    { default: router },
    { initUnleashClient },
    { default: iconSet },
    { iconMapFn },
    { getActiveTheme, theme },
    { docsLinkMixin },
    { INIT_OPTIONS_KEY, THEME_KEY },
    { RUNAI_ECHARTS_THEME_NAME },
  ] = await Promise.all([
    import("vue"),
    import("pinia"),
    import("quasar"),
    import("./directives"),
    import("./plugins/vue-dom-purify-html"),
    import("./App.vue"),
    import("./router"),
    import("./plugins/unleash"),
    import("./plugins/phosphor-icon-set"),
    import("./plugins/icon-map-fn"),
    import("./plugins/theme"),
    import("./mixins/docs-link.mixin"),
    import("vue-echarts"),
    import("./plugins/echarts"),
  ]);

  const app = createApp(App);
  // Select the final theme during bootstrap because it depends on asynchronously loaded configuration.
  const finalTheme = getActiveTheme(bootConfig.distributionAssets.profile);
  useDirectives(app);
  initUnleashClient(app);

  app.use(createPinia());
  app.use(theme, finalTheme);
  app.use(router);
  app.use(Quasar, {
    plugins: { Notify },
    iconSet,
    config: { iconMapFn },
  });
  app.use(VueDOMPurifyHTML, VueDOMPurifyHTMLOptions);
  app.mixin(docsLinkMixin);
  app.provide(INIT_OPTIONS_KEY, { renderer: "svg" });
  app.provide(THEME_KEY, RUNAI_ECHARTS_THEME_NAME);
  app.mount("#app");
}

void bootstrap();
