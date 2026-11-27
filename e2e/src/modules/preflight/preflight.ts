/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { existsSync, mkdirSync, writeFileSync } from 'fs';
import { join } from 'path';
import { ClusterProbe } from './clusterProbe';
import { defaultRunaiVersion, getSupportMatrix } from './supportMatrix';
import type {
    CheckResult,
    CheckStatus,
    ComponentSpec,
    Platform,
    PreflightReport,
    ProbeResult,
    SupportMatrix,
    VersionRange
} from './supportMatrix';

/**
 * Kit preflight gate. Reads the versions of the prerequisites NVIDIA Run:ai
 * documents as supported, compares them with the published support matrix, and
 * fails the run before any test starts.
 *
 * Environment:
 *   PREFLIGHT_MODE            enforce (default) | warn | off
 *   PREFLIGHT_RUNAI_VERSION   support matrix to check against (default 2.26)
 *   PREFLIGHT_RESULTS_DIR     where the report files are written
 *   PREFLIGHT_DEBUG           also print every lookup the probes made
 *   KUBERNETES_PLATFORM       k8s | ocp, overrides platform auto-detection
 */

// -------------------------------------------------------------------------
// Version comparison
// -------------------------------------------------------------------------

/**
 * Version handling for the preflight gate.
 *
 * The components in the support matrix do not share a versioning scheme:
 * Kubernetes and Knative are SemVer, the NVIDIA operators are CalVer (`25.10`),
 * and cluster objects report versions with assorted prefixes and suffixes
 * (`v25.10.0`, `1.34.2-eks-1234567`). All of them compare correctly as a list of
 * numeric components, so that is what the gate uses instead of a SemVer parser.
 */

const NUMERIC_PREFIX = /^\D*(\d+(?:\.\d+)*)/;

/** Extracts the leading numeric components, dropping any prefix or suffix. */
export const parseVersion = (raw: string | undefined): number[] | undefined => {
    if (!raw) {
        return undefined;
    }
    const match = NUMERIC_PREFIX.exec(raw.trim());
    if (!match) {
        return undefined;
    }
    return match[1].split('.').map(Number);
};

/** Negative, zero or positive, comparing only as deep as the shorter version. */
export const compareVersions = (left: number[], right: number[]): number => {
    const depth = Math.min(left.length, right.length);
    for (let index = 0; index < depth; index++) {
        if (left[index] !== right[index]) {
            return left[index] - right[index];
        }
    }
    return 0;
};

/** `indeterminate` means the detected value carries too little to be placed. */
export type RangeVerdict = 'inside' | 'outside' | 'indeterminate';

/**
 * A version shorter than the bound ties on its shared prefix only, so `1` could
 * be any 1.x and cannot be placed against `1.9.2`. The tie is decidable when
 * the rest of the bound is zeroes, as `76` is against `76.0`.
 */
const tieIsAmbiguous = (version: number[], bound: number[]): boolean =>
    version.length < bound.length && bound.slice(version.length).some((part) => part !== 0);

/**
 * Where `raw` sits against the inclusive range. Bounds are compared at the
 * precision they are written with, so a `26.3` upper bound admits `26.3.5`.
 */
export const checkRange = (raw: string | undefined, range: VersionRange): RangeVerdict => {
    const version = parseVersion(raw);
    if (!version) {
        return 'indeterminate';
    }
    for (const [bound, outside] of [
        [parseVersion(range.min), -1],
        [parseVersion(range.max), 1]
    ] as const) {
        if (!bound) {
            continue;
        }
        const order = compareVersions(version, bound);
        if (Math.sign(order) === outside) {
            return 'outside';
        }
        if (order === 0 && tieIsAmbiguous(version, bound)) {
            return 'indeterminate';
        }
    }
    return 'inside';
};

/** True only when `raw` can be placed and falls inside the range. */
export const satisfiesRange = (raw: string | undefined, range: VersionRange): boolean =>
    checkRange(raw, range) === 'inside';

export const formatRange = (range: VersionRange): string =>
    range.max ? `${range.min} - ${range.max}` : `${range.min} or above`;

// -------------------------------------------------------------------------
// Evaluation
// -------------------------------------------------------------------------

export interface EvaluateInput {
    matrix: SupportMatrix;
    platform: Platform;
    probeResults: ProbeResult[];
}

const describeActual = (probe: ProbeResult | undefined): string => {
    if (probe?.detected) {
        return probe.detectedVia ? `${probe.detected} (via ${probe.detectedVia})` : probe.detected;
    }
    return 'not detected';
};

const evaluateComponent = (
    component: ComponentSpec,
    probe: ProbeResult | undefined,
    platform: Platform
): CheckResult => {
    const supportedRange = formatRange(component.supported);
    const base = {
        componentId: component.id,
        displayName: component.displayName,
        expected: `supported ${supportedRange}`,
        actual: describeActual(probe),
        attempts: probe?.attempts
    };

    if (component.skipOnPlatform === platform) {
        return {
            ...base,
            status: 'skip',
            actual: 'not checked',
            detail: `Not applicable on ${platform === 'ocp' ? 'OpenShift' : 'Kubernetes'}.`
        };
    }

    if (!probe?.detected) {
        const unreadable = probe?.error ? ` Version could not be read from the cluster: ${probe.error}` : '';
        const seen = (probe?.attempts ?? []).find((attempt) => attempt.matched > 0);
        if (seen) {
            // Calling this missing would send the reader looking for an install
            // that is already there, so say where it was seen instead.
            return {
                ...base,
                actual: 'deployed, no version readable',
                status: component.requirement === 'required' ? 'fail' : 'warn',
                detail:
                    `${component.displayName} is deployed — ${seen.via} matched ${seen.matched} ` +
                    'object(s) — but none of them carried a version, so the supported range ' +
                    `${supportedRange} could not be checked.${unreadable}`
            };
        }
        if (component.requirement === 'required') {
            return {
                ...base,
                status: 'fail',
                detail: `${component.displayName} is required but was not found on the cluster.${unreadable}`
            };
        }
        const conditional = component.appliesWhen
            ? `${component.displayName} was not found. It is required when ${component.appliesWhen}.`
            : `${component.displayName} was not found.`;
        return { ...base, status: 'warn', detail: `${conditional}${unreadable}` };
    }

    const verdict = checkRange(probe.detected, component.supported);
    if (verdict === 'indeterminate') {
        // The component is installed, so reporting it as missing would be wrong,
        // but an image tag such as `v1-5170a36` says nothing about the release.
        return {
            ...base,
            status: component.requirement === 'required' ? 'fail' : 'warn',
            detail:
                `${component.displayName} is deployed, but '${probe.detected}' does not give a version ` +
                `that can be compared with the supported range ${supportedRange}.`
        };
    }
    if (verdict === 'outside') {
        // A component that is only needed for some workload types never blocks
        // the run: the tests that depend on it are the ones that will suffer,
        // and the reader can decide whether that matters to them.
        const applies = component.appliesWhen ? ` It is required when ${component.appliesWhen}.` : '';
        return {
            ...base,
            status: component.requirement === 'required' ? 'fail' : 'warn',
            detail:
                `Deployed ${probe.detected} is outside the supported range ${supportedRange}.` +
                (component.requirement === 'required' ? '' : applies)
        };
    }
    return { ...base, status: 'pass', detail: `Deployed ${probe.detected} is supported.` };
};

export const evaluatePreflight = (input: EvaluateInput): PreflightReport => {
    const { matrix, platform, probeResults } = input;
    const probesById = new Map(probeResults.map((result) => [result.componentId, result]));

    const checks = matrix.components.map((component) =>
        evaluateComponent(component, probesById.get(component.id), platform)
    );

    const summary: Record<CheckStatus, number> = { pass: 0, fail: 0, warn: 0, skip: 0 };
    for (const check of checks) {
        summary[check.status]++;
    }

    return {
        runaiVersion: matrix.runaiVersion,
        source: matrix.source,
        platform,
        generatedAt: new Date().toISOString(),
        checks,
        summary,
        passed: summary.fail === 0
    };
};

// -------------------------------------------------------------------------
// Rendering
// -------------------------------------------------------------------------

const STATUS_LABEL: Record<CheckStatus, string> = {
    pass: 'PASS',
    fail: 'FAIL',
    warn: 'WARN',
    skip: 'SKIP'
};

const padded = (value: string, width: number): string => value.padEnd(width, ' ');

/** Fixed-width table for the kit run log, which has no colour or markdown. */
export const renderText = (report: PreflightReport): string => {
    const rows = report.checks.map((check) => ({
        status: STATUS_LABEL[check.status],
        name: check.displayName,
        expected: check.expected,
        actual: check.actual
    }));

    const nameWidth = Math.max(9, ...rows.map((row) => row.name.length));
    const expectedWidth = Math.max(8, ...rows.map((row) => row.expected.length));

    const lines: string[] = [];
    lines.push(`NVIDIA Run:ai ${report.runaiVersion} prerequisite preflight (${report.platform})`);
    lines.push(`Support matrix: ${report.source}`);
    lines.push('');
    lines.push(
        `${padded('STATUS', 6)}  ${padded('COMPONENT', nameWidth)}  ${padded('EXPECTED', expectedWidth)}  ACTUAL`
    );
    for (const row of rows) {
        lines.push(
            `${padded(row.status, 6)}  ${padded(row.name, nameWidth)}  ${padded(row.expected, expectedWidth)}  ${row.actual}`
        );
    }

    const problems = report.checks.filter((check) => check.status === 'fail' || check.status === 'warn');
    if (problems.length > 0) {
        lines.push('');
        lines.push('Details:');
        for (const check of problems) {
            lines.push(`  [${STATUS_LABEL[check.status]}] ${check.displayName}: ${check.detail}`);
        }
    }

    lines.push('');
    lines.push(
        `Summary: ${report.summary.pass} passed, ${report.summary.fail} failed, ` +
            `${report.summary.warn} warnings, ${report.summary.skip} skipped`
    );
    return lines.join('\n');
};

/**
 * Every lookup the probes made, so a component reported as not found can be
 * diagnosed from the report instead of by repeating kubectl against the cluster.
 */
export const renderProbeTrace = (report: PreflightReport): string => {
    const lines: string[] = ['Probe trace:'];
    for (const check of report.checks) {
        lines.push(`  ${check.displayName}`);
        if (!check.attempts || check.attempts.length === 0) {
            lines.push('    no lookup was made');
            continue;
        }
        for (const attempt of check.attempts) {
            const outcome = attempt.extracted
                ? `read ${attempt.extracted}`
                : `no version in ${attempt.matched} object(s)`;
            lines.push(`    ${attempt.via}: ${outcome}${attempt.note ? ` (${attempt.note})` : ''}`);
        }
    }
    return lines.join('\n');
};

const escapePipes = (value: string): string => value.replace(/\|/g, '\\|');

/** Markdown table for the kit results package. */
export const renderMarkdown = (report: PreflightReport): string => {
    const lines: string[] = [];
    lines.push(`# NVIDIA Run:ai ${report.runaiVersion} prerequisite preflight`);
    lines.push('');
    lines.push(`- Generated: ${report.generatedAt}`);
    lines.push(`- Platform: ${report.platform}`);
    lines.push(`- Support matrix: ${report.source}`);
    lines.push(`- Result: ${report.passed ? 'PASSED' : 'FAILED'}`);
    lines.push('');
    lines.push('| Status | Component | Expected | Actual | Detail |');
    lines.push('| --- | --- | --- | --- | --- |');
    for (const check of report.checks) {
        lines.push(
            `| ${STATUS_LABEL[check.status]} | ${escapePipes(check.displayName)} | ${escapePipes(check.expected)} | ` +
                `${escapePipes(check.actual)} | ${escapePipes(check.detail)} |`
        );
    }
    lines.push('');
    lines.push('## Probe trace');
    lines.push('');
    lines.push('```');
    lines.push(renderProbeTrace(report));
    lines.push('```');
    lines.push('');
    return lines.join('\n');
};

// -------------------------------------------------------------------------
// Entry point
// -------------------------------------------------------------------------

export type PreflightMode = 'enforce' | 'warn' | 'off';

const isTruthy = (raw: string | undefined): boolean =>
    ['1', 'true', 'yes', 'on'].includes((raw ?? '').trim().toLowerCase());

const readMode = (raw: string | undefined): PreflightMode => {
    const mode = (raw ?? 'enforce').toLowerCase();
    if (mode === 'enforce' || mode === 'warn' || mode === 'off') {
        return mode;
    }
    throw new Error(`Invalid PREFLIGHT_MODE '${raw}'. Use enforce, warn or off.`);
};

const readPlatform = (raw: string | undefined, probe: ClusterProbe): Platform => {
    const platform = raw?.toLowerCase();
    if (platform === 'ocp' || platform === 'k8s') {
        return platform;
    }
    return probe.detectPlatform();
};

const writeReportFiles = (report: PreflightReport, directory: string): string[] => {
    try {
        mkdirSync(directory, { recursive: true });
    } catch {
        return [];
    }
    const written: string[] = [];
    for (const [name, contents] of [
        ['preflight-report.json', `${JSON.stringify(report, null, 2)}\n`],
        ['preflight-report.md', renderMarkdown(report)]
    ]) {
        const path = join(directory, name);
        try {
            writeFileSync(path, contents);
            written.push(path);
        } catch {
            // The results volume is optional; the log always carries the report.
        }
    }
    return written;
};

export const buildReport = (probe: ClusterProbe = new ClusterProbe()): PreflightReport => {
    const matrix = getSupportMatrix(process.env.PREFLIGHT_RUNAI_VERSION || defaultRunaiVersion);
    const platform = readPlatform(process.env.KUBERNETES_PLATFORM, probe);

    const probeResults = matrix.components
        .filter((component) => component.skipOnPlatform !== platform)
        .map((component) => probe.resolve(component));

    return evaluatePreflight({ matrix, platform, probeResults });
};

export const main = (): number => {
    let mode: PreflightMode;
    try {
        mode = readMode(process.env.PREFLIGHT_MODE);
    } catch (error) {
        console.error(
            `Preflight configuration error: ${error instanceof Error ? error.message : String(error)}`
        );
        return 1;
    }

    if (mode === 'off') {
        console.log('Prerequisite preflight is disabled (PREFLIGHT_MODE=off).');
        return 0;
    }

    let report: PreflightReport;
    try {
        report = buildReport();
    } catch (error) {
        console.error(`Preflight could not run: ${error instanceof Error ? error.message : String(error)}`);
        return mode === 'enforce' ? 1 : 0;
    }

    console.log('');
    console.log(renderText(report));
    console.log('');

    if (isTruthy(process.env.PREFLIGHT_DEBUG)) {
        console.log(renderProbeTrace(report));
        console.log('');
    }

    const resultsDir = process.env.PREFLIGHT_RESULTS_DIR || '/app/e2e/results';
    if (existsSync(resultsDir)) {
        for (const path of writeReportFiles(report, resultsDir)) {
            console.log(`Preflight report written to ${path}`);
        }
    }

    if (report.passed) {
        return 0;
    }
    if (mode === 'warn') {
        console.warn('Preflight found unsupported prerequisites. Continuing because PREFLIGHT_MODE=warn.');
        return 0;
    }
    console.error('Preflight failed: the cluster does not meet the documented NVIDIA Run:ai prerequisites.');
    console.error('Re-run with PREFLIGHT_MODE=warn to continue anyway.');
    return 1;
};
