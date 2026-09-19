#!/usr/bin/env node
/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
/* eslint-env node */

/**
 * Manual Accessibility Audit Script (Snapshot Mode with axe-core)
 * 1. Opens Chrome with remote debugging
 * 2. You login and set up the form as needed
 * 3. Press Enter → audits CURRENT page state (NO refresh!)
 *
 * Uses axe-core which audits the DOM as-is, preserving your form state.
 */

import { spawn } from "child_process";
import { createInterface } from "readline";
import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";
import puppeteer from "puppeteer-core";
import { AxePuppeteer } from "@axe-core/puppeteer";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const OUTPUT_DIR = path.join(__dirname, "../a11y-reports");
const DEBUG_PORT = 9222;
const APP_URL = "http://localhost:9000";

function ensureOutputDir() {
  if (!fs.existsSync(OUTPUT_DIR)) {
    fs.mkdirSync(OUTPUT_DIR, { recursive: true });
  }
}

function launchChrome() {
  console.log("🚀 Launching Chrome with remote debugging...\n");

  const chromePath =
    process.platform === "darwin"
      ? "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
      : process.platform === "win32"
        ? "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe"
        : "google-chrome";

  try {
    const chrome = spawn(
      chromePath,
      [
        `--remote-debugging-port=${DEBUG_PORT}`,
        "--no-first-run",
        "--no-default-browser-check",
        "--user-data-dir=/tmp/chrome-a11y-profile",
        `${APP_URL}/projects/new-v2`,
      ],
      { detached: true, stdio: "ignore" },
    );

    chrome.on("error", (err) => {
      console.log(`   Chrome spawn error: ${err.message}`);
    });

    chrome.unref();
    return chrome;
  } catch (err) {
    console.log(`   Failed to launch Chrome: ${err.message}`);
    return null;
  }
}

async function getCurrentTabUrl() {
  try {
    const response = await fetch(`http://localhost:${DEBUG_PORT}/json`);
    const pages = await response.json();
    const page = pages.find((p) => p.type === "page" && !p.url.startsWith("chrome://"));
    return page?.url || null;
  } catch {
    return null;
  }
}

async function waitForEnter(prompt) {
  const rl = createInterface({ input: process.stdin, output: process.stdout });
  return new Promise((resolve) => {
    rl.question(prompt, () => {
      rl.close();
      resolve();
    });
  });
}

async function runAxeAudit(outputName) {
  console.log(`\n🔍 Running axe-core accessibility audit (NO page refresh)...`);

  const outputPath = path.join(OUTPUT_DIR, outputName);

  try {
    // Connect to existing Chrome
    const browser = await puppeteer.connect({
      browserURL: `http://localhost:${DEBUG_PORT}`,
    });

    const pages = await browser.pages();
    const page = pages.find((p) => !p.url().startsWith("chrome://")) || pages[0];
    const currentUrl = page.url();

    console.log(`   Page: ${currentUrl}`);
    console.log(`   Analyzing current DOM state...`);

    // Run axe-core on the current page state (NO navigation!)
    const results = await new AxePuppeteer(page)
      .withTags(["wcag2a", "wcag2aa", "wcag21a", "wcag21aa", "best-practice"])
      .analyze();

    // Disconnect (don't close browser)
    browser.disconnect();

    // Save JSON report only
    fs.writeFileSync(`${outputPath}.json`, JSON.stringify(results, null, 2));

    return { success: true, results, url: currentUrl };
  } catch (err) {
    console.error(`   Error: ${err.message}`);
    return { success: false, error: err.message };
  }
}

function printResults(results) {
  const violations = results.violations || [];
  const passes = results.passes || [];

  const totalIssues = violations.reduce((sum, v) => sum + v.nodes.length, 0);
  const score =
    passes.length > 0
      ? Math.round((passes.length / (passes.length + violations.length)) * 100)
      : violations.length === 0
        ? 100
        : 0;

  console.log("\n" + "═".repeat(60));
  console.log("📊 ACCESSIBILITY AUDIT RESULTS (axe-core)");
  console.log("═".repeat(60));
  console.log(`\nScore: ${score}/100 ${score >= 90 ? "✅" : score >= 70 ? "⚠️" : "❌"}`);
  console.log(`Violations: ${violations.length} types, ${totalIssues} total elements`);
  console.log(`Passed: ${passes.length} checks\n`);

  if (violations.length > 0) {
    console.log("Issues found:\n");

    // Sort by impact
    const impactOrder = { critical: 0, serious: 1, moderate: 2, minor: 3 };
    violations.sort((a, b) => (impactOrder[a.impact] || 4) - (impactOrder[b.impact] || 4));

    for (const v of violations) {
      const impactEmoji = { critical: "🔴", serious: "🟠", moderate: "🟡", minor: "⚪" };
      console.log(`${impactEmoji[v.impact] || "⚪"} [${v.impact}] ${v.help} (${v.nodes.length} elements)`);
      console.log(`   Rule: ${v.id}`);

      // Show first 2 affected elements
      for (const node of v.nodes.slice(0, 2)) {
        const selector = node.target.join(" → ");
        console.log(`   └─ ${selector.substring(0, 70)}`);
      }
      if (v.nodes.length > 2) {
        console.log(`   └─ ... and ${v.nodes.length - 2} more`);
      }
      console.log("");
    }
  } else {
    console.log("✅ No accessibility violations found!");
  }
}

async function main() {
  console.log("═".repeat(60));
  console.log("  ACCESSIBILITY AUDIT - Snapshot Mode (axe-core)");
  console.log("═".repeat(60));
  console.log("\nThis audits the CURRENT page state - no refresh!");
  console.log("");

  ensureOutputDir();
  launchChrome();

  console.log("   Waiting for Chrome to start...");
  await new Promise((r) => setTimeout(r, 3000));

  let testUrl = await getCurrentTabUrl();

  if (!testUrl) {
    console.log("\n⚠️  Chrome didn't start automatically.");
    console.log("   Please start Chrome manually in another terminal:\n");
    console.log(
      `   /Applications/Google\\ Chrome.app/Contents/MacOS/Google\\ Chrome --remote-debugging-port=${DEBUG_PORT} --user-data-dir=/tmp/chrome-a11y-profile "${APP_URL}/projects/new-v2"\n`,
    );

    await waitForEnter("Press ENTER once Chrome is running... ");

    testUrl = await getCurrentTabUrl();
    if (!testUrl) {
      console.log("\n❌ Still can't connect to Chrome. Exiting.");
      return;
    }
  }

  console.log("✅ Chrome connected\n");
  console.log("📝 Now in Chrome:");
  console.log("   1. Login if needed");
  console.log("   2. Navigate to the form you want to audit");
  console.log("   3. Open inputs, expand sections, interact with the form");
  console.log("   4. Come back here and press Enter\n");
  console.log("   ✨ Your form state will be PRESERVED during the audit!\n");

  await waitForEnter("Press ENTER when ready to audit... ");

  // Get current URL for naming
  testUrl = await getCurrentTabUrl();
  const urlPath = new URL(testUrl).pathname.replace(/\//g, "-").replace(/^-|-$/g, "") || "page";
  const timestamp = new Date().toISOString().slice(0, 16).replace(/[:.]/g, "-");
  const outputName = `${urlPath}-${timestamp}`;

  const result = await runAxeAudit(outputName);

  if (result.success) {
    printResults(result.results);
    const outputPath = path.join(OUTPUT_DIR, outputName);
    console.log(`\n📁 Report saved: ${outputPath}.json`);
  } else {
    console.log("\n❌ Audit failed:", result.error);
  }

  console.log("\n💡 Chrome stays open - close it manually when done");
}

main().catch(console.error);
