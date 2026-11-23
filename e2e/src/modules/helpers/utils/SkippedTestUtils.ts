/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import * as fs from 'fs';
import * as path from 'path';
import { fileURLToPath } from 'url';
import { dirname } from 'path';
import { logger } from '../logger/logger';
import { config, settings } from '../config/config';
import { teams } from '../config/tags';
import { TestInfo } from 'playwright/test';
import { fetchTestResults } from '../../../../aggregateResultsPerTeam';

// ESM equivalent of __dirname
const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const projectRoot = path.resolve(__dirname, '..', '..', '..', '..');
const skipReportPath = path.join(projectRoot, 'skips_report.txt');
const skipReportHtml = path.join(projectRoot, 'skips_report.html');
const teamLabels = Object.values(teams);

function capitalize(str: string): string {
    return str.charAt(0).toUpperCase() + str.slice(1);
}

function reportField(value: string): string {
    return value.replace(/[\t\r\n]+/g, ' ').trim();
}

function escapeHtml(value: string): string {
    return value.replace(
        /[&<>"']/g,
        (character) =>
            ({
                '&': '&amp;',
                '<': '&lt;',
                '>': '&gt;',
                '"': '&quot;',
                '\'': '&#39;'
            })[character]!
    );
}

function logSkip(testName: string, reason: string, teamName: string, labels: string[]): void {
    const reportEntry = [teamName, testName, reason, labels.join(', ')].map(reportField).join('\t');
    fs.appendFileSync(skipReportPath, `${reportEntry}\n`, 'utf8');
}

function addUnrecordedSkip(testName: string, teamName: string, labels: string[], reason: string): void {
    if (teamName && testName) {
        logSkip(testName, reason || 'No reason recorded', capitalize(teamName), labels);
    } else {
        logger.error('Team name or test name is missing when adding an unrecorded skip.');
    }
}

//remove tags
function cleanTestTitle(title: string): string {
    return title.replace(/@\w+/g, '').trim();
}

function isTestNameInSkipReport(testName: string): boolean {
    const reportContent = fs.readFileSync(skipReportPath, 'utf8');
    return reportContent.includes(testName);
}

function extractTagsFromPlaywright(testInfo: TestInfo): string[] {
    const labels: string[] = [];
    const tagRegex = /@([\w-]+)/g;

    testInfo.titlePath.forEach((title: string) => {
        let match;
        while ((match = tagRegex.exec(title)) !== null) {
            const foundLabel = match[1];
            if (!labels.includes(foundLabel) && !Object.keys(teams).includes(foundLabel)) {
                labels.push(foundLabel);
            }
        }
    });

    return labels;
}

export function goOverSkippedTestsFromAllureResults(): void {
    if (!settings.isLocalRun) {
        const allTestResults = fetchTestResults();

        // Filter skipped tests
        const skippedTests = allTestResults.filter((test) => test.status === 'skipped');

        logger.info(`Found ${skippedTests.length} skipped tests`);

        for (const test of skippedTests) {
            const { testName, team, labels, skipReason } = test;
            if (!isTestNameInSkipReport(testName)) {
                addUnrecordedSkip(testName, team, labels, skipReason);
            }
        }
    } else {
        logger.debug(`isLocalRun=${settings.isLocalRun}, skipping.`);
    }
}


function findTeamNameFromPlaywrightTestInfo(testInfo: TestInfo): string {
    const teamTags = Object.values(teams);
    const reversedTitlePath = [...testInfo.titlePath].reverse();

    const matchingTag = reversedTitlePath.find((element) => teamTags.some((tag) => element.includes(tag)));

    if (matchingTag) {
        const teamName = Object.keys(teams).find((key) => {
            const tags = matchingTag.match(/@\w+/g);
            return tags?.includes(teams[key]);
        });
        return teamName ? capitalize(teamName) : 'Not specified';
    }
    return 'Not specified';
}

export function skipPlaywrightTest(testContext: any, testInfo: TestInfo, reason: string): void {
    if (!reason.trim()) {
        throw new Error('A skipped test must include a reason.');
    }
    const cleanedTitle = cleanTestTitle(testInfo.title);
    const teamName = findTeamNameFromPlaywrightTestInfo(testInfo);
    const labels = extractTagsFromPlaywright(testInfo);
    if (process.env.CREATE_SKIPS_REPORT) {
        logSkip(cleanedTitle, reason, teamName, labels);
    }
    testContext.skip(true, `Skipping test: ${cleanedTitle}. Reason: ${reason}`);
}

export function generateSkipsHTMLReport(): void {
    if (process.env.CREATE_SKIPS_REPORT) {
        if (!fs.existsSync(skipReportPath)) {
            fs.writeFileSync(skipReportPath, '', 'utf8');
        }

        goOverSkippedTestsFromAllureResults();

        const header = `
        <html>
        <head>
            <style>
                body {
                    font-family: Arial, sans-serif;
                    margin: 0;
                    padding: 0;
                    display: flex;
                    justify-content: center;
                    align-items: flex-start;
                    height: 100vh;
                    background-color: #f5f5f5; 
                }
                
                .main-page {
                    width: 90%;
                    display: flex;
                    flex-direction: column;
                    align-items: center;
                    padding: 10px;
                }

                h2 {
                    font-family: Helvetica;
                    font-size: 2.5rem;
                    text-align: center;
                    color: #e04cd9;
                    text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.3);
                    background: linear-gradient(45deg, #2e44b3, #000000);
                    -webkit-background-clip: text;
                    -webkit-text-fill-color: transparent;
                    padding: 10px;
                    margin: 20px 0;
                }

                table {
                    border-collapse: collapse;
                    margin-top: 10px; /* Reduced space between filters and table */
                    box-shadow: 0px 0px 10px rgba(0, 0, 0, 0.1);
                    background-color: #fff;
                }
                th, td {
                    padding: 12px;
                    text-align: left;
                    border-bottom: 1px solid #ddd;
                }
                th {
                    background-color: rgba(223, 25, 149, 0.06);
                    color: rgb(0,38,62);
                }
                tr:hover {
                    background-color: #f5f5f5;
                }
                a {
                    color: #2d5889;
                    text-decoration: none;
                }
                a:hover {
                    text-decoration: underline;
                }

                .label {
                    display: inline-block;
                    padding: 4px 8px;
                    font-size: 0.9rem;
                    font-weight: bold;
                    border-radius: 12px;
                    color: white;
                    cursor: pointer; /* Added cursor pointer for labels */
                }
                .label-blue { background-color: #82c0ff; } /* Blue */
                .label-green { background-color: #81e29e; } /* Green */
                .label-yellow { background-color: #f5d76e; } /* Yellow */
                .label-orange { background-color: #ffb274; } /* Orange */
                .label-pink { background-color: #ff9ca7; } /* Pink */
                .label-purple { background-color: #c39add; } /* Purple */
                .label-red { background-color: #ff6b6b; } /* Red */
                .label-cyan { background-color: #76d7c4; } /* Cyan */
                .label-burgundy { background-color: #933600; } /* Brown */
                .label-gray { background-color: #b0b0b0; } /* Gray */
                .label-dark-blue { background-color: #252a43; } /* Black */

                .filter-input {
                    padding: 8px;
                    font-size: 0.9rem;
                    border: 1px solid #ccc;
                    border-radius: 5px;
                    width: 180px; /* Adjusted width to make the cells less wide */
                    margin-right: 10px; /* Space between filters */
                }
                .controls-container {
                    display: flex;
                    align-items: center;
                    justify-content: flex-start;
                    gap: 10px; /* Add some spacing between filters */
                    margin: 10px 20px;
                }
                #filters {
                    display: flex;
                    align-items: center;
                }
            </style>
        </head>
        <body>
            <div class="main-page">
                <header>
                    <h2>SKIPPED TESTS | ${config.clusterNamePrefix}</h2>
                </header>                
                <div class="controls-container">
                    <div id="filters">                    
                        <input type="text" id="teamFilter" class="filter-input" placeholder="Filter by team name">
                        <input type="text" id="testNameFilter" class="filter-input" placeholder="Filter by test name">
                        <input type="text" id="reasonFilter" class="filter-input" placeholder="Filter by skip reason">
                        <input type="text" id="tagFilter" class="filter-input" placeholder="Filter by tag">
                    </div>
                </div>
                <table id="skipsTable">
                    <tr>
                        <th>Team</th>
                        <th>Test Name</th>
                        <th>Skip reason</th>
                        <th>Tags</th>
                    </tr>`;

        const footer = `
                    </table>
                    <div id="totalRows" style="margin-top: 10px;">Rows: 0</div>
                </div>
                <script>
                    document.querySelector("h2").innerText = "SKIPPED TESTS | ${config.clusterNamePrefix}";

                    const table = document.getElementById('skipsTable');
                    const teamFilter = document.getElementById('teamFilter');
                    const testNameFilter = document.getElementById('testNameFilter');
                    const reasonFilter = document.getElementById('reasonFilter');
                    const tagFilter = document.getElementById('tagFilter');

                    teamFilter.addEventListener('input', () => filterTable());
                    testNameFilter.addEventListener('input', () => filterTable());
                    reasonFilter.addEventListener('input', () => filterTable());
                    tagFilter.addEventListener('input', () => filterTable());

                    function filterTable() {
                        const teamValue = teamFilter.value.toLowerCase();
                        const testNameValue = testNameFilter.value.toLowerCase();
                        const reasonValue = reasonFilter.value.toLowerCase();
                        const tagValue = tagFilter.value.toLowerCase();
                        let visibleRowCount = 0;
                        
                        for (let i = 1; i < table.rows.length; i++) {
                            const teamCell = table.rows[i].cells[0].textContent.toLowerCase();
                            const testNameCell = table.rows[i].cells[1].textContent.toLowerCase();
                            const reasonCell = table.rows[i].cells[2].textContent.toLowerCase();
                            const tagCell = table.rows[i].cells[3].textContent.toLowerCase();

                            if (teamCell.includes(teamValue) && testNameCell.includes(testNameValue) && reasonCell.includes(reasonValue) && tagCell.includes(tagValue)) {
                                table.rows[i].style.display = '';
                                visibleRowCount++;
                            } else {
                                table.rows[i].style.display = 'none';
                            }
                        }

                        document.getElementById('totalRows').innerText = 'Rows: ' + visibleRowCount;
                    }

                    function filterByLabel(label) {
                        tagFilter.value = label;
                        filterTable();
                    }

                    function filterByTeam(team) {
                        teamFilter.value = team;
                        filterTable();
                    }
                </script>   
            </body>
        </html>`;

        const teamLabels: { [key: string]: string } = {
            Viper: 'label-purple',
            Phoenix: 'label-cyan',
            'Not specified': 'label-gray'
        };

        const rows = fs
            .readFileSync(skipReportPath, 'utf8')
            .split('\n')
            .filter((line) => line.trim())
            .map((line) => {
                const [teamName, testName, reason, labelsStr = ''] = line.split('\t');
                const labelClass = teamLabels[teamName] || 'label-gray';
                const labels = labelsStr
                    .split(', ')
                    .map(
                        (label) =>
                            `<span class="label label-dark-blue" onclick="filterByLabel('${label}')">${label}</span>`
                    )
                    .join(' ');

                const skipReason = escapeHtml(reason || 'No reason recorded');

                return {
                    teamName,
                    row: `<tr><td><span class="label ${labelClass}" onclick="filterByTeam('${teamName.toLowerCase()}')">${teamName}</span></td><td>${escapeHtml(testName)}</td><td>${skipReason}</td><td>${labels}</td></tr>`
                };
            })
            .sort((a, b) => a.teamName.localeCompare(b.teamName)) // Sorting by team name
            .map((entry) => entry.row)
            .join('\n');

        const totalRows = (rows.match(/<tr>/g) || []).length;

        const finalHtml = `${header}${rows}${footer}`.replace('Rows: 0', `Rows: ${totalRows}`);

        fs.writeFileSync(skipReportHtml, finalHtml, 'utf8');
        logger.info('SKIP REPORT IS READY!');
    }
}
