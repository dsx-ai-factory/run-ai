/* eslint-env node */
require("@rushstack/eslint-patch/modern-module-resolution");

module.exports = {
  root: true,
  plugins: ["vuejs-accessibility", "security", "no-secrets"],
  extends: [
    "plugin:vue/vue3-essential",
    "eslint:recommended",
    "@vue/eslint-config-typescript/recommended",
    "@vue/eslint-config-prettier",
    "plugin:security/recommended-legacy",
  ],
  env: {
    "vue/setup-compiler-macros": true,
  },
  overrides: [
    {
      // These components render `autofocus` on their own single root element, so an inline
      // eslint-disable template comment would add a Comment vnode as a root sibling and turn
      // the component into a multi-root Fragment (breaking $el-based test lookups). Disabled
      // here instead. Every current consumer only sets autofocus behind a user action (opening
      // a dialog, or -- for runai-name-validation's fromCopyId case -- the explicit "duplicate"
      // navigation flow), never a bare page load.
      files: [
        "src/components/common/runai-name-validation/runai-name-validation.vue",
        "src/components/common/runai-email-validation/runai-email-validation.vue",
      ],
      rules: {
        "vuejs-accessibility/no-autofocus": "off",
      },
    },
  ],
  rules: {
    "@typescript-eslint/consistent-type-imports": "error",
    "@typescript-eslint/no-non-null-assertion": "error",
    "@typescript-eslint/no-explicit-any": "error",
    "@typescript-eslint/ban-ts-comment": "off",
    "vue/attribute-hyphenation": ["error", "always"],
    "vue/no-unused-vars": [
      "error",
      {
        ignorePattern: "^_",
      },
    ],
    "vue/v-on-event-hyphenation": [
      "error",
      "always",
      {
        ignore: [],
      },
    ],
    "vue/order-in-components": [
      "error",
      {
        order: [
          "name",
          "components",
          "emits",
          "inject",
          "props",
          "provide",
          "data",
          "setup",
          "created",
          "mounted",
          "computed",
          "methods",
          "watch",
          "unmounted",
        ],
      },
    ],
    "vue/component-tags-order": [
      "error",
      {
        order: [["script", "template"], "style"],
      },
    ],
    "vue/component-name-in-template-casing": [
      "error",
      "kebab-case",
      {
        registeredComponentsOnly: true,
        ignores: [],
      },
    ],
    "vue/custom-event-name-casing": ["error", "kebab-case"],
    "vue/this-in-template": ["error", "never"],
    "no-unused-vars": "off",
    "@typescript-eslint/no-unused-vars": ["error", { argsIgnorePattern: "^_" }],
    // Some duplicate enum values are intentional (e.g., EWorkloadCategory, errorMessages)
    "@typescript-eslint/no-duplicate-enum-values": "warn",
    "prefer-const": "error",
    "no-secrets/no-secrets": ["error", { tolerance: 4.5 }],
    // A11y rules — set to warn for ratchet-based threshold enforcement
    "vuejs-accessibility/alt-text": "warn",
    "vuejs-accessibility/anchor-has-content": "warn",
    "vuejs-accessibility/aria-props": "warn",
    "vuejs-accessibility/aria-role": "warn",
    "vuejs-accessibility/aria-unsupported-elements": "warn",
    "vuejs-accessibility/click-events-have-key-events": "warn",
    "vuejs-accessibility/form-control-has-label": "warn",
    // Treat the sanitized v-dompurify-html directive as accessible heading content (same as the built-in v-html).
    "vuejs-accessibility/heading-has-content": ["warn", { accessibleDirectives: ["dompurify-html"] }],
    "vuejs-accessibility/iframe-has-title": "warn",
    "vuejs-accessibility/interactive-supports-focus": "warn",
    "vuejs-accessibility/label-has-for": "warn",
    "vuejs-accessibility/media-has-caption": "warn",
    "vuejs-accessibility/mouse-events-have-key-events": "warn",
    "vuejs-accessibility/no-access-key": "warn",
    "vuejs-accessibility/no-aria-hidden-on-focusable": "warn",
    "vuejs-accessibility/no-autofocus": "warn",
    "vuejs-accessibility/no-distracting-elements": "warn",
    "vuejs-accessibility/no-redundant-roles": "warn",
    "vuejs-accessibility/no-role-presentation-on-focusable": "warn",
    "vuejs-accessibility/no-static-element-interactions": "warn",
    "vuejs-accessibility/role-has-required-aria-props": "warn",
    "vuejs-accessibility/tabindex-no-positive": "warn",
  },
  globals: {
    globals: true,
  },
  ignorePatterns: ["*/setup.js"],
};
