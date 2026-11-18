// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

import { spawnSync } from 'node:child_process';
import { copyFileSync, existsSync, mkdirSync, readdirSync, readFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const packageRoot = path.dirname(fileURLToPath(import.meta.url));
const repositoryRoot = path.resolve(packageRoot, process.env.RUNAI_REPO_PATH || '..');
const clientsRoot = path.join(packageRoot, 'src/modules/api/apis');
const serviceSpec = (service) => `control-plane/${service}/swagger/_build/swagger.yaml`;
const specs = {
    assetsApi: serviceSpec('assets-service'),
    auditApi: serviceSpec('audit-service'),
    authorizationApi: serviceSpec('authorization'),
    backendApi: null,
    backendClusterApi: null,
    backendUndocumentedApi: 'e2e/src/modules/api/apis/backendUndocumentedApi/swagger.yaml',
    bffServiceApi: serviceSpec('bff-service'),
    catalogApi: serviceSpec('catalog-service'),
    cliExposer: serviceSpec('cli-exposer'),
    clusterApi: serviceSpec('cluster-service'),
    datavolumesApi: serviceSpec('datavolumes-service'),
    identityManagerApi: serviceSpec('identity-manager'),
    k8sObjectsTrackerApi: serviceSpec('k8s-objects-tracker'),
    notificationsProxyApi: serviceSpec('notifications-proxy'),
    orgUnitApi: serviceSpec('org-unit-service'),
    policyApi: serviceSpec('policy-service'),
    projectsDepartmentsApi: serviceSpec('org-unit-service'),
    tempBackendSettingsApi: serviceSpec('temp-backend-settings'),
    tenantsManagerApi: serviceSpec('tenants-manager'),
    workloadsApi: serviceSpec('workloads'),
    workloadsManagerApi: serviceSpec('workloads-manager')
};

function main() {
    const args = process.argv.slice(2);
    const checkOnly = args.includes('--check');
    const outputIndex = args.indexOf('--output-dir');
    let outputRoot;
    if (outputIndex !== -1) {
        if (!args[outputIndex + 1] || args[outputIndex + 1].startsWith('--')) {
            throw new Error('--output-dir requires a directory.');
        }
        outputRoot = path.resolve(args[outputIndex + 1]);
        args.splice(outputIndex, 2);
    }
    const names = args.filter((arg) => arg !== '--check' && arg !== '--all');
    const clients = args.includes('--all') ? Object.keys(specs) : names;
    if (!clients.length) {
        throw new Error('Select a client or --all. Use --check to validate inputs without generation.');
    }
    const directories = readdirSync(clientsRoot, { withFileTypes: true })
        .filter((entry) => entry.isDirectory() && existsSync(path.join(clientsRoot, entry.name, 'index.ts')))
        .map((entry) => entry.name);
    const unlisted = directories.filter((name) => !Object.hasOwn(specs, name));
    if (unlisted.length) {
        throw new Error(`Clients missing from generation coverage: ${unlisted.join(', ')}`);
    }
    // Check every requested input before writing any generated output.
    const errors = clients.flatMap((name) => {
        if (!Object.hasOwn(specs, name)) return [`Unknown client: ${name}`];
        if (!specs[name]) return [`${name}: no verified public source spec; resolve legacy contracts before generation.`];
        if (!existsSync(path.join(repositoryRoot, specs[name]))) return [`${name}: missing source spec ${specs[name]}`];
        return [];
    });
    if (errors.length) throw new Error(errors.join('\n'));
    if (checkOnly) {
        console.log(`Validated source specs for ${clients.length} clients.`);
        return;
    }
    const outputs = new Map(clients.map((name) => [name, outputRoot
        ? path.join(outputRoot, name)
        : path.join(clientsRoot, name, 'generated')]));
    for (const [name, output] of outputs) {
        if (existsSync(output) && readdirSync(output).length) {
            throw new Error(`${name}: output must be empty. Use --output-dir for a preview; review old-file removals before replacement.`);
        }
    }
    const version = JSON.parse(readFileSync(path.join(packageRoot, 'openapitools.json'), 'utf8'))['generator-cli'].version;
    const jar = process.env.OPENAPI_GENERATOR_JAR;
    const java = process.env.JAVA_HOME ? path.join(process.env.JAVA_HOME, 'bin/java') : 'java';
    if (jar) {
        const result = spawnSync(java, ['-jar', jar, 'version'], { encoding: 'utf8' });
        if (result.status !== 0 || result.stdout.trim() !== version) {
            throw new Error(`OPENAPI_GENERATOR_JAR must run generator ${version}.`);
        }
    }
    for (const name of clients) {
        // Generate into an empty directory so removed operations cannot survive as stale files.
        const output = outputs.get(name);
        mkdirSync(output, { recursive: true });
        const input = path.join(repositoryRoot, specs[name]);
        const templates = path.join(clientsRoot, 'templates');
        const options = ['-g', 'typescript-node', '--additional-properties=supportsES6=true,npmVersion=6.9.0,typescriptThreePlus=true'];
        const command = jar ? java : 'docker';
        const commandArgs = jar
            ? ['-jar', jar, 'generate', '-i', input, '-t', templates, '-o', output, ...options]
            : ['run', '--rm', '-v', `${repositoryRoot}:/specs:ro`, '-v', `${templates}:/templates:ro`, '-v', `${output}:/output`,
                `openapitools/openapi-generator-cli:v${version}`, 'generate', '-i', `/specs/${specs[name]}`, '-t', '/templates', '-o', '/output', ...options];
        const result = spawnSync(command, commandArgs, { stdio: 'inherit' });
        if (result.error) throw result.error;
        if (result.status !== 0) throw new Error(`Generation failed for ${name}.`);
        copyFileSync(path.join(templates, 'httpError.ts'), path.join(output, 'api/httpError.ts'));
    }
}

try {
    main();
} catch (error) {
    console.error(error.message);
    process.exitCode = 1;
}
