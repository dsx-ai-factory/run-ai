#!/usr/bin/env node
/* eslint-env node */
/**
 * Runs vue-tsc type checking and fails if the number of TS errors exceeds
 * a configured threshold. Use this as a ratchet to prevent error count growth.
 *
 * Usage:
 *   node scripts/typecheck-threshold.mjs [--threshold <n>]
 *
 * The threshold can be overridden via the TS_ERROR_THRESHOLD env var or the
 * --threshold CLI argument. Otherwise it defaults to the hardcoded value below.
 */

import { spawnSync } from "child_process";

// Update this value whenever you fix errors to lock in the improvement.
const DEFAULT_THRESHOLD = 0;

const args = process.argv.slice(2);
const thresholdArgIndex = args.indexOf("--threshold");
const threshold =
  thresholdArgIndex !== -1
    ? parseInt(args[thresholdArgIndex + 1], 10)
    : parseInt(process.env.TS_ERROR_THRESHOLD ?? String(DEFAULT_THRESHOLD), 10);

if (isNaN(threshold)) {
  console.error("Invalid threshold value.");
  process.exit(1);
}

console.log(`Running TypeScript type check (error threshold: ${threshold})...`);

const result = spawnSync("npx", ["vue-tsc", "--noEmit", "-p", "tsconfig.vitest.json", "--composite", "false"], {
  encoding: "utf-8",
  stdio: ["ignore", "pipe", "pipe"],
});

const output = (result.stdout ?? "") + (result.stderr ?? "");
const errorLines = output.split("\n").filter((line) => /error TS\d+/.test(line));
const errorCount = errorLines.length;

console.log(`TypeScript errors found: ${errorCount}`);

// Summary
const fileErrorMap = {};
const codeErrorMap = {};

for (const line of errorLines) {
  const fileMatch = line.match(/^(.+?)\(\d+,\d+\)/);
  const codeMatch = line.match(/error (TS\d+)/);
  if (fileMatch) fileErrorMap[fileMatch[1]] = (fileErrorMap[fileMatch[1]] ?? 0) + 1;
  if (codeMatch) codeErrorMap[codeMatch[1]] = (codeErrorMap[codeMatch[1]] ?? 0) + 1;
}

const filesAffected = Object.keys(fileErrorMap).length;
const topFiles = Object.entries(fileErrorMap)
  .sort((a, b) => b[1] - a[1])
  .slice(0, 10);
const topCodes = Object.entries(codeErrorMap)
  .sort((a, b) => b[1] - a[1])
  .slice(0, 10);

console.log("=== Summary ===");
console.log(`Total errors : ${errorCount}`);
console.log(`Files affected: ${filesAffected}`);

if (topFiles.length > 0) {
  console.log("\nTop files by error count:");
  for (const [file, count] of topFiles) {
    console.log(`  ${count.toString().padStart(4)}x  ${file}`);
  }
}

if (topCodes.length > 0) {
  console.log("\nTop error codes:");
  for (const [code, count] of topCodes) {
    console.log(`  ${count.toString().padStart(4)}x  ${code}`);
  }
}
console.log("===============\n");

if (errorCount > threshold) {
  console.log("\n--- TS Errors ---");
  console.log(errorLines.join("\n"));
  console.log("-----------------\n");
  console.error(
    `❌ TS error count (${errorCount}) exceeds threshold (${threshold}).` +
      ` Fix at least ${errorCount - threshold} error(s) to get back under the limit.`,
  );
  process.exit(1);
}

if (errorCount < threshold) {
  console.error(
    `❌ TS error count (${errorCount}) is ${threshold - errorCount} below the threshold (${threshold}).\n` +
      `   You must lower DEFAULT_THRESHOLD to ${errorCount} to lock in the improvement.\n` +
      `   File: control-plane/frontend/scripts/typecheck-threshold.mjs (line 17)`,
  );
  process.exit(1);
}

console.log(`✅ TS error count (${errorCount}) matches threshold (${threshold}). All good!`);
