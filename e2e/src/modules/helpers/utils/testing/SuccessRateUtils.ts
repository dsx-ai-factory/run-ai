/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';
import { dirname } from 'path';
import { config } from '../../index';
import { teams } from '../../config/tags';

// ESM equivalent of __dirname
const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const teamLabels: { [key: string]: string } = {
    Viper: 'label-purple',
    Phoenix: 'label-cyan',
    'Not specified': 'label-gray'
};

const projectRoot = path.resolve(__dirname, '..', '..', '..', '..', '..');
const txtFilePath = path.join(projectRoot, 'success_rates_ordered.txt');
const htmlFilePath = path.join(projectRoot, 'success_rates_ordered.html');
const dirPath = './allure-report/data/test-cases';

export function calculateSuccessRates(): void {
    // Clear the output file
    fs.writeFileSync(txtFilePath, '');

    // Read all JSON files in the directory
    const files = fs.readdirSync(dirPath).filter((file) => file.endsWith('.json'));

    const successRates: { successRate: number; teamName: string; cleanTestName: string; tags: string[] }[] =
        [];

    for (const file of files) {
        const filePath = path.join(dirPath, file);

        try {
            const fileData = JSON.parse(fs.readFileSync(filePath, 'utf-8'));
            const passed = fileData?.extra?.history?.statistic?.passed;
            const skipped = fileData?.extra?.history?.statistic?.skipped;
            const total = fileData?.extra?.history?.statistic?.total;
            const isPlaywrightTest = fileData['labels'].some(
                (label) => label.name === 'framework' && label.value === 'Playwright'
            );
            if (isPlaywrightTest) {
                fileData.fullName = fileData['labels']?.find((label) => label.name === 'titlePath')?.value;
            }
            const fullName = fileData?.fullName;

            // Check if passed, skipped, and total are valid numbers
            if (
                passed === null ||
                skipped === null ||
                total === null ||
                typeof passed !== 'number' ||
                typeof total !== 'number'
            ) {
                console.error(`Error in file ${filePath}: passed or total is null or not a number`);
                continue;
            }

            // Extract the team name and tags
            let teamName = 'Not specified';
            const tags = [];

            // Extract tags and determine team name

            const tagMatches = fullName.match(/@\w[\w-]*/g);
            if (tagMatches) {
                tagMatches.forEach((tag) => {
                    const tagName = tag.substring(1);
                    if (teams[tagName]) {
                        teamName = tagName.charAt(0).toUpperCase() + tagName.slice(1);
                    } else {
                        tags.push(tagName);
                    }
                });
            }

            const lastIndex = fullName.lastIndexOf('>');
            const testName = isPlaywrightTest ? fileData.name : fullName.slice(lastIndex + 1).trim(); // Get text after the last '>'

            // Clean test name by removing tags
            const cleanTestName = testName.replace(/@\w[\w-]*/g, '').trim();

            // Calculate success rate if applicable
            if (total - skipped > 0) {
                const passSuccessRate = ((passed / (total - skipped)) * 100).toFixed(2);

                successRates.push({
                    successRate: parseFloat(passSuccessRate),
                    teamName,
                    cleanTestName,
                    tags
                });
            }
        } catch (error) {
            console.error(`Failed to process file ${filePath}:`, error);
        }
    }

    // Sort by success rate in ascending order
    successRates.sort((a, b) => a.successRate - b.successRate);

    // Write sorted results to file
    for (const rate of successRates) {
        fs.appendFileSync(
            txtFilePath,
            `${rate.successRate}\t${rate.teamName}\t${rate.cleanTestName}\t${rate.tags.join(' ')}\n`
        );
    }
}

export function generateSuccessRatesReport(): void {
    calculateSuccessRates();

    try {
        if (!fs.existsSync(txtFilePath)) {
            console.error(`Input file does not exist at path: ${txtFilePath}`);
            return;
        }

        const fileContent = fs.readFileSync(txtFilePath, 'utf8').trim();
        if (!fileContent) {
            console.error('Input file is empty.');
            return;
        }

        const rows = fileContent
            .split('\n')
            .filter((line) => line.trim()) // Skip empty lines
            .map((line) => {
                const [successRateStr, teamName, testName, tags] = line.split('\t');

                const successRate = parseFloat(successRateStr);
                if (isNaN(successRate)) {
                    console.warn(`Skipping line with invalid success rate: ${line}`);
                    return null;
                }

                const colorClass = teamLabels[teamName] || 'label-gray';

                return `<tr>
                    <td>${successRate}%</td>
                    <td><span class="label ${colorClass}" onclick="filterByTeam('${teamName}')">${teamName}</span></td>
                    <td>${testName}</td>
                    <td>${
                        tags
                            ? tags
                                  .split(' ')
                                  .map(
                                      (tag) =>
                                          `<span class="label label-dark-blue" onclick="filterByTag('${tag}')">${tag}</span>`
                                  )
                                  .join(' ')
                            : ''
                    }</td>
                </tr>`;
            })
            .filter((row) => row !== null) // Filter out null values
            .join('\n');

        if (!rows) {
            console.error('No rows were processed. Check the input data and format.');
            return;
        }

        const totalRows = (rows.match(/<tr>/g) || []).length;

        const htmlContent = `
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
                        margin-top: 10px;
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

                    .label {
                        display: inline-block;
                        padding: 4px 8px;
                        font-size: 0.9rem;
                        font-weight: bold;
                        border-radius: 12px;
                        color: white;
                        cursor: pointer;
                        user-select: none;
                    }
                    .label-blue { background-color: #82c0ff; }
                    .label-green { background-color: #81e29e; }
                    .label-yellow { background-color: #f5d76e; }
                    .label-orange { background-color: #ffb274; }
                    .label-pink { background-color: #ff9ca7; }
                    .label-purple { background-color: #c39add; }
                    .label-red { background-color: #ff6b6b; }
                    .label-cyan { background-color: #76d7c4; }
                    .label-burgundy { background-color: #933600; }
                    .label-gray { background-color: #b0b0b0; }
                    .label-dark-blue { background-color: #252a43; }

                    .filter-container {
                        margin-bottom: 20px;
                    }
                    .filter-input {
                        padding: 8px;
                        font-size: 0.9rem;
                        border: 1px solid #ccc;
                        border-radius: 5px;
                        margin-right: 10px;
                        width: 180px;
                    }
                    #totalRows {
                        margin-top: 3mm;
                        font-size: 1rem;
                    }
                </style>
                <script>
                    function filterByTeam(teamName) {
                        document.getElementById('teamFilter').value = teamName;
                        filterTable();
                    }

                    function filterByTag(tag) {
                        document.getElementById('labelsFilter').value = tag;
                        filterTable();
                    }

                    function filterTable() {
                        const teamValue = document.getElementById('teamFilter').value.toLowerCase();
                        const testNameValue = document.getElementById('testNameFilter').value.toLowerCase();
                        const labelsValue = document.getElementById('labelsFilter').value.toLowerCase();
                        const filterSuccessRate = document.getElementById('successRateCheckbox').checked;
                        const rows = document.querySelectorAll('#data-table tbody tr');
                        let visibleRows = 0;

                        rows.forEach(row => {
                            const successRate = parseFloat(row.cells[0].textContent);
                            const team = row.cells[1].textContent.toLowerCase();
                            const testName = row.cells[2].textContent.toLowerCase();
                            const labels = row.cells[3].textContent.toLowerCase();

                            const matchesTeam = team.includes(teamValue);
                            const matchesTestName = testName.includes(testNameValue);
                            const matchesLabels = labels.includes(labelsValue);
                            const matchesSuccessRate = filterSuccessRate ? successRate < 100 : true;

                            if (matchesTeam && matchesTestName && matchesLabels && matchesSuccessRate) {
                                row.style.display = '';
                                visibleRows++;
                            } else {
                                row.style.display = 'none';
                            }
                        });

                        document.getElementById('totalRows').textContent = 'Rows: ' + visibleRows;
                    }

                    document.addEventListener('DOMContentLoaded', () => {
                        const header = document.querySelector('h2');
                        header.innerText = \`SUCCESS RATE | ${config.clusterNamePrefix}\`;
                        filterTable();
                    });
                </script>
            </head>
            <body>
                <div class="main-page">
                    <h2></h2>
                    <div class="filter-container">
                        <input type="text" id="teamFilter" class="filter-input" placeholder="Filter by Team" oninput="filterTable()">
                        <input type="text" id="testNameFilter" class="filter-input" placeholder="Filter by Test Name" oninput="filterTable()">
                        <input type="text" id="labelsFilter" class="filter-input" placeholder="Filter by Labels" oninput="filterTable()">
                        <input type="checkbox" id="successRateCheckbox" checked onchange="filterTable()"> Hide tests with a success rate of 100%
                    </div>
                    <table id="data-table">
                        <thead>
                            <tr>
                                <th>Success Rate</th>
                                <th>Team</th>
                                <th>Test Name</th>
                                <th>Tags</th>
                            </tr>
                        </thead>
                        <tbody>
                            ${rows}
                        </tbody>
                    </table>
                    <div id="totalRows">Rows: ${totalRows}</div>
                </div>
            </body>
            </html>
        `;

        fs.writeFileSync(htmlFilePath, htmlContent, 'utf8');
        console.log(`HTML report generated at ${htmlFilePath}`);
    } catch (error) {
        console.error('Failed to generate HTML report:', error);
    }
}
