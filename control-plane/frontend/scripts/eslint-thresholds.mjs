#!/usr/bin/env node
/* eslint-env node */
/**
 * Runs ESLint once on all source files and enforces three checks from the
 * same JSON output:
 *   - a11y:         vuejs-accessibility/* warning count (exact match — fails high AND low)
 *   - script-setup: file count using Options API (asymmetric — fails only on increase)
 *   - security:     any error-level finding from security/* or no-secrets/* rules
 *
 * The vue/component-api-style rule is intentionally NOT in .eslintrc.cjs (would
 * spam normal lint output with thousands of warnings). It is injected here via
 * --rule so the same JSON output covers both checks.
 *
 * Usage:
 *   node scripts/eslint-thresholds.mjs [--a11y-threshold N] [--script-setup-threshold N]
 *
 * Env vars: A11Y_WARNING_THRESHOLD, SCRIPT_SETUP_THRESHOLD
 */

import { spawnSync } from "child_process";
import { relative } from "path";

// Update these whenever you fix a11y warnings or migrate a component to <script setup>.
const A11Y_DEFAULT_THRESHOLD = 0;
const SCRIPT_SETUP_DEFAULT_THRESHOLD = 654;

const SCRIPT_PATH = "control-plane/frontend/scripts/eslint-thresholds.mjs";

const args = process.argv.slice(2);

function getThreshold({ argName, envName, defaultValue }) {
  const flag = `--${argName}`;
  const idx = args.indexOf(flag);
  let raw;
  if (idx !== -1) {
    const next = args[idx + 1];
    if (next === undefined || next.startsWith("-")) {
      console.error(`Missing value for ${flag}.`);
      process.exit(1);
    }
    raw = next;
  } else {
    raw = process.env[envName] ?? String(defaultValue);
  }
  if (!/^\d+$/.test(raw)) {
    console.error(`Invalid threshold for ${flag}: "${raw}". Must be a non-negative integer.`);
    process.exit(1);
  }
  return Number(raw);
}

const a11yThreshold = getThreshold({
  argName: "a11y-threshold",
  envName: "A11Y_WARNING_THRESHOLD",
  defaultValue: A11Y_DEFAULT_THRESHOLD,
});
const scriptSetupThreshold = getThreshold({
  argName: "script-setup-threshold",
  envName: "SCRIPT_SETUP_THRESHOLD",
  defaultValue: SCRIPT_SETUP_DEFAULT_THRESHOLD,
});

console.log("Running consolidated ESLint threshold check...");
console.log(`  a11y warning threshold:     ${a11yThreshold}`);
console.log(`  Options API file threshold: ${scriptSetupThreshold}\n`);

const result = spawnSync(
  "npx",
  [
    "eslint",
    ".",
    "--ext",
    ".vue,.js,.jsx,.cjs,.mjs,.ts,.tsx,.cts,.mts",
    "--no-error-on-unmatched-pattern",
    "--rule",
    '{"vue/component-api-style": ["warn", ["script-setup"]]}',
    "--format",
    "json",
  ],
  { encoding: "utf-8", stdio: ["ignore", "pipe", "pipe"], maxBuffer: 100 * 1024 * 1024 },
);

const stdout = result.stdout ?? "";
const stderr = result.stderr ?? "";

function logOutputExcerpt(label, output) {
  if (!output) return;
  const max = 5000;
  const excerpt = output.length > max ? `${output.slice(0, max)}\n... output truncated ...` : output;
  console.error(`${label}:\n${excerpt}`);
}

// ESLint exits 1 when warnings/errors exist (expected here). status 2 = config/usage error.
if (result.error || result.status === 2 || result.status === null) {
  console.error("Failed to execute ESLint.");
  if (result.error) console.error(result.error.message);
  logOutputExcerpt("stderr", stderr);
  logOutputExcerpt("stdout", stdout);
  process.exit(1);
}

let eslintResults;
try {
  eslintResults = JSON.parse(stdout.trim() || "[]");
} catch (err) {
  console.error("Failed to parse ESLint JSON output.");
  console.error(err instanceof Error ? err.message : String(err));
  logOutputExcerpt("stderr", stderr);
  logOutputExcerpt("stdout", stdout);
  process.exit(1);
}

function formatPath(p) {
  const rel = relative(process.cwd(), p);
  return rel.startsWith("..") ? p : rel;
}

function increment(map, key) {
  map.set(key, (map.get(key) ?? 0) + 1);
}

const SECURITY_RULE_PREFIXES = ["security/", "no-secrets/"];

// === Aggregate ===
const a11yRuleCounts = new Map();
const a11yFileCounts = new Map();
const optionsApiFiles = new Set();
const securityErrors = [];

for (const fileResult of eslintResults) {
  const filePath = fileResult.filePath ?? "";
  for (const msg of fileResult.messages ?? []) {
    if (msg.ruleId?.startsWith("vuejs-accessibility/")) {
      increment(a11yRuleCounts, msg.ruleId);
      if (filePath) increment(a11yFileCounts, filePath);
    }
    if (msg.ruleId === "vue/component-api-style" && filePath) {
      optionsApiFiles.add(filePath);
    }
    if (msg.severity === 2 && SECURITY_RULE_PREFIXES.some((p) => msg.ruleId?.startsWith(p))) {
      securityErrors.push({ file: filePath, rule: msg.ruleId, line: msg.line, message: msg.message });
    }
  }
}

const a11yCount = [...a11yRuleCounts.values()].reduce((a, b) => a + b, 0);
const scriptSetupCount = optionsApiFiles.size;

// === A11y summary ===
console.log("=== A11y Summary ===");
console.log(`Total warnings: ${a11yCount}`);
console.log(`Files affected: ${a11yFileCounts.size}`);

const topRules = [...a11yRuleCounts.entries()].sort((a, b) => b[1] - a[1]);
if (topRules.length > 0) {
  console.log("\nWarnings by rule:");
  for (const [rule, count] of topRules) {
    console.log(`  ${count.toString().padStart(4)}x  ${rule}`);
  }
}

const topFiles = [...a11yFileCounts.entries()].sort((a, b) => b[1] - a[1]).slice(0, 10);
if (topFiles.length > 0) {
  console.log("\nTop files by warning count:");
  for (const [file, count] of topFiles) {
    console.log(`  ${count.toString().padStart(4)}x  ${formatPath(file)}`);
  }
}
console.log("====================\n");

// === Script-setup summary ===
console.log("=== Script Setup Summary ===");
console.log(`Files still using Options API: ${scriptSetupCount}`);
console.log(`Threshold: ${scriptSetupThreshold}`);
console.log("============================\n");

// === Verdicts (print both, then exit) ===
let failed = false;

// a11y: exact-match ratchet
if (a11yCount > a11yThreshold) {
  console.error(
    `❌ A11y warning count (${a11yCount}) exceeds threshold (${a11yThreshold}).` +
      ` Fix at least ${a11yCount - a11yThreshold} warning(s) to get back under the limit.`,
  );
  failed = true;
} else if (a11yCount < a11yThreshold) {
  console.error(
    `❌ A11y warning count (${a11yCount}) is ${a11yThreshold - a11yCount} below the threshold (${a11yThreshold}).\n` +
      `   You must lower A11Y_DEFAULT_THRESHOLD to ${a11yCount} to lock in the improvement.\n` +
      `   File: ${SCRIPT_PATH}`,
  );
  failed = true;
} else {
  console.log(`✅ A11y warning count (${a11yCount}) matches threshold.`);
}

// script-setup: asymmetric ratchet (warns when below, never blocks)
if (scriptSetupCount > scriptSetupThreshold) {
  console.error(
    `❌ Options API file count (${scriptSetupCount}) exceeds threshold (${scriptSetupThreshold}).` +
      ` New components must use <script setup>.` +
      ` Convert ${scriptSetupCount - scriptSetupThreshold} file(s) or fix the new one(s).`,
  );
  failed = true;
} else {
  if (scriptSetupCount < scriptSetupThreshold) {
    console.log(
      `⚠️  Options API file count (${scriptSetupCount}) is ${scriptSetupThreshold - scriptSetupCount} below threshold (${scriptSetupThreshold}).\n` +
        `   Please lower SCRIPT_SETUP_DEFAULT_THRESHOLD to ${scriptSetupCount} to lock in the improvement.\n` +
        `   File: ${SCRIPT_PATH}`,
    );
  }
  console.log(`✅ Script-setup threshold check passed (${scriptSetupCount} ≤ ${scriptSetupThreshold}).`);
}

// security: fail on any error-level finding from security/* or no-secrets/*
console.log("\n=== Security Summary ===");
if (securityErrors.length > 0) {
  console.error(`Security errors: ${securityErrors.length}`);
  const shown = securityErrors.slice(0, 20);
  for (const e of shown) {
    console.error(`  ${formatPath(e.file)}:${e.line}  [${e.rule}]  ${e.message}`);
  }
  if (securityErrors.length > 20) {
    console.error(`  ... and ${securityErrors.length - 20} more`);
  }
  console.error(`❌ Security errors found. Fix the ${securityErrors.length} finding(s) above.`);
  failed = true;
} else {
  console.log(`✅ No security errors found.`);
}
console.log("========================\n");

process.exit(failed ? 1 : 0);
