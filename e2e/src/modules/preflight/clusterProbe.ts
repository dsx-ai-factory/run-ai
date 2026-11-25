/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { spawnSync } from 'child_process';
import type {
    ComponentSpec,
    HelmProbe,
    KubectlProbe,
    Platform,
    Probe,
    ProbeAttempt,
    ProbeResult
} from './supportMatrix';

export interface CommandResult {
    code: number;
    stdout: string;
    stderr: string;
}

/** Injected so the probe logic can be exercised without a cluster. */
export type CommandRunner = (command: string, args: string[]) => CommandResult;

const REQUEST_TIMEOUT = '20s';

export const spawnCommandRunner: CommandRunner = (command, args) => {
    const result = spawnSync(command, args, { encoding: 'utf-8', maxBuffer: 32 * 1024 * 1024 });
    if (result.error) {
        return { code: 127, stdout: '', stderr: result.error.message };
    }
    return { code: result.status ?? 1, stdout: result.stdout ?? '', stderr: result.stderr ?? '' };
};

interface HelmRelease {
    chart?: string;
    app_version?: string;
}

const parseJson = (raw: string): any | undefined => {
    try {
        return JSON.parse(raw);
    } catch {
        return undefined;
    }
};

/** Kubernetes objects come back either as a single object or as a List. */
const asObjects = (parsed: any): Record<string, any>[] => {
    if (!parsed) {
        return [];
    }
    return Array.isArray(parsed.items) ? parsed.items : [parsed];
};

export class ClusterProbe {
    private helmReleases: HelmRelease[] | undefined;

    constructor(private readonly run: CommandRunner = spawnCommandRunner) {}

    /** OpenShift is identified by the presence of a `ClusterVersion` object. */
    public detectPlatform(): Platform {
        const result = this.run('kubectl', [
            'get',
            'clusterversions.config.openshift.io',
            '-o',
            'json',
            `--request-timeout=${REQUEST_TIMEOUT}`
        ]);
        if (result.code !== 0) {
            return 'k8s';
        }
        return asObjects(parseJson(result.stdout)).length > 0 ? 'ocp' : 'k8s';
    }

    public resolve(component: ComponentSpec): ProbeResult {
        const errors: string[] = [];
        const attempts: ProbeAttempt[] = [];
        for (const probe of component.probes) {
            try {
                const found = this.runProbe(probe, attempts);
                if (found) {
                    return {
                        componentId: component.id,
                        detected: found.version,
                        detectedVia: found.via,
                        attempts
                    };
                }
            } catch (error) {
                errors.push(error instanceof Error ? error.message : String(error));
            }
        }
        return errors.length === component.probes.length
            ? { componentId: component.id, error: errors.join('; '), attempts }
            : { componentId: component.id, attempts };
    }

    private runProbe(probe: Probe, attempts: ProbeAttempt[]): { version: string; via: string } | undefined {
        switch (probe.kind) {
            case 'k8sServerVersion':
                return this.probeServerVersion(attempts);
            case 'helm':
                return this.probeHelm(probe, attempts);
            case 'kubectl':
                return this.probeKubectl(probe, attempts);
        }
    }

    private probeServerVersion(attempts: ProbeAttempt[]): { version: string; via: string } | undefined {
        const via = 'kubectl version';
        const result = this.run('kubectl', ['version', '-o', 'json', `--request-timeout=${REQUEST_TIMEOUT}`]);
        // kubectl exits non-zero when the client cannot reach the API server, but
        // also when only the client version could be printed, so parse regardless.
        const version = parseJson(result.stdout)?.serverVersion?.gitVersion;
        if (!version) {
            const note = result.stderr.trim() || 'no serverVersion in output';
            attempts.push({ via, matched: 0, note });
            throw new Error(`kubectl version failed: ${note}`);
        }
        attempts.push({ via, matched: 1, extracted: version });
        return { version, via };
    }

    private probeHelm(probe: HelmProbe, attempts: ProbeAttempt[]): { version: string; via: string } | undefined {
        const via = `helm release of chart ${probe.chart}`;
        const release = this.listHelmReleases().find((candidate) =>
            (candidate.chart ?? '').startsWith(`${probe.chart}-`)
        );
        if (!release) {
            // A chart installed as a dependency of another release is not listed.
            attempts.push({ via, matched: 0, note: 'no release of that chart in helm list -A' });
            return undefined;
        }
        const chartVersion = (release.chart ?? '').slice(probe.chart.length + 1);
        const preferred =
            probe.preferAppVersion === false ? chartVersion : release.app_version || chartVersion;
        attempts.push({ via, matched: 1, ...(preferred ? { extracted: preferred } : {}) });
        return preferred ? { version: preferred, via: `helm release chart ${release.chart}` } : undefined;
    }

    private listHelmReleases(): HelmRelease[] {
        if (this.helmReleases) {
            return this.helmReleases;
        }
        const result = this.run('helm', ['list', '-A', '-o', 'json']);
        const parsed = result.code === 0 ? parseJson(result.stdout) : undefined;
        this.helmReleases = Array.isArray(parsed) ? parsed : [];
        return this.helmReleases;
    }

    private probeKubectl(
        probe: KubectlProbe,
        attempts: ProbeAttempt[]
    ): { version: string; via: string } | undefined {
        for (const lookup of this.plannedLookups(probe)) {
            const { objects, note } = this.kubectlGet(lookup.args);
            let extracted: string | undefined;
            for (const object of objects) {
                const version = probe.extract(object);
                if (typeof version === 'string' && version.trim() !== '') {
                    extracted = version.trim();
                    break;
                }
            }
            attempts.push({
                via: lookup.via,
                matched: objects.length,
                ...(extracted ? { extracted } : {}),
                ...(note ? { note } : {})
            });
            if (extracted) {
                return { version: extracted, via: lookup.via };
            }
        }
        return undefined;
    }

    /**
     * Named lookups in the documented namespaces first, then a cluster-wide
     * fallback for installations that moved the component elsewhere. Lazy, so a
     * hit in the first namespace costs a single kubectl call.
     */
    private *plannedLookups(probe: KubectlProbe): Generator<{ args: string[]; via: string }> {
        // kubectl rejects a selector together with an object name, so the name
        // takes precedence where a probe carries both.
        const selector = probe.name ? [] : probe.labelSelector ? ['-l', probe.labelSelector] : [];
        const describe = probe.name
            ? `${probe.resource}/${probe.name}`
            : probe.labelSelector
              ? `${probe.resource} matching ${probe.labelSelector}`
              : probe.resource;

        const namespaces = probe.namespaces ?? [];
        for (const namespace of namespaces) {
            const nameArgs = probe.name ? [probe.name] : [];
            yield {
                args: ['get', probe.resource, ...nameArgs, ...selector, '-n', namespace, '-o', 'json'],
                via: `${describe} -n ${namespace}`
            };
        }
        if (probe.name && namespaces.length === 0) {
            yield {
                args: ['get', probe.resource, probe.name, '-o', 'json'],
                via: describe
            };
        }
        // Without a discriminator the listing would return every object of the
        // resource type and the first one with a version would be reported as
        // this component, so fall back only on something that identifies it.
        const discriminator = probe.labelSelector
            ? ['-l', probe.labelSelector]
            : probe.name
              ? [`--field-selector=metadata.name=${probe.name}`]
              : [];
        yield {
            args: ['get', probe.resource, '-A', ...discriminator, '-o', 'json'],
            via: `${describe} across all namespaces`
        };
    }

    /**
     * A missing object or a missing CRD is a normal outcome here, so failures
     * return no objects rather than throwing. The stderr is handed back so the
     * recorded attempt can say why the lookup came back empty.
     */
    private kubectlGet(args: string[]): { objects: Record<string, any>[]; note?: string } {
        const result = this.run('kubectl', [...args, `--request-timeout=${REQUEST_TIMEOUT}`]);
        if (result.code !== 0) {
            return { objects: [], note: result.stderr.trim() || `kubectl exited ${result.code}` };
        }
        return { objects: asObjects(parseJson(result.stdout)) };
    }
}
