/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

export const PodsStateColumn = 'PODS-Running-(Pending)';

export const GPUAllocationColumn = 'GPUs-Allocated-(Requested)';

export class CliUtils {
    static parseProjectListTableOutput(cliTubularResponse: string): { [key: string]: string }[] {
        return CliUtils.parseTableOutput(cliTubularResponse, 0);
    }

    static parseTableOutput(cliTubularResponse: string, headerLine: number): { [key: string]: string }[] {
        const lines = cliTubularResponse.split('\n');
        const dataLines = lines
            .slice(headerLine + 1)
            .filter((line) => line.trim() !== '')
            .filter((line) => !/^[\s─━_=-]+$/u.test(line));

        const regex = /\S+\s*/g;
        const columnHeaders = lines[headerLine].match(regex) ?? [];

        const rows = [];

        dataLines.map((line) => {
            const columns: { [key: string]: string } = {};
            let offset = 0;
            columnHeaders.forEach((columnHeader: string) => {
                columns[columnHeader.trim()] = line.slice(offset, offset + columnHeader.length - 1).trim();
                offset += columnHeader.length;
            });
            rows.push(columns);
        });

        return rows;
    }

    parseListJobsTable(cliTubularResponse: string, headerLine: number): { [key: string]: string }[] {
        // make sure each column header consists of one word
        cliTubularResponse = cliTubularResponse.replace('GPUs Allocated (Requested)', GPUAllocationColumn);
        cliTubularResponse = cliTubularResponse.replace('PODs Running (Pending)', PodsStateColumn);
        cliTubularResponse = cliTubularResponse.replace('SERVICE URL(S)', 'SERVICE-URL(S)');
        // alias new `runai workload list` column headers back to the legacy `runai list jobs` names
        cliTubularResponse = cliTubularResponse.replace('Workload', 'NAME');
        cliTubularResponse = cliTubularResponse.replace('Status', 'STATUS');
        cliTubularResponse = cliTubularResponse.replace('Project', 'PROJECT');

        return CliUtils.parseTableOutput(cliTubularResponse, headerLine);
    }
    /**
     * @summary Get all field values from given cli output.
     * @param cliResponse CLI output. e.g: 'runai describe job' output
     * @param field Field in cliResponse.
     */

    static getFieldValuesOfAllResources(cliResponse: string, field: string): string[] {
        const matches: RegExpMatchArray = cliResponse.match(new RegExp(`^${field}(.*)$`, 'gm'));
        if (matches) {
            return matches.map((item) => {
                return item.split(': ')[1].trim();
            });
        } else {
            return [];
        }
    }

    /**
     * @summary Get field value from given cli output.
     * @param cliResponse CLI output. e.g: 'runai describe job' output
     * @param field Field in cliResponse.
     */
    static getFieldValue(cliResponse: string, field: string): string {
        return CliUtils.getFieldValuesOfAllResources(cliResponse, field)[0];
    }

    /**
     * @summary Get numeric field value from given cli output.
     * @param cliResponse CLI output. e.g: 'runai describe job' output
     * @param field Field in cliResponse.
     */
    static getNumericFieldValue(cliResponse: string, field: string): number {
        return Number(CliUtils.getFieldValue(cliResponse, field));
    }

    /**
     * @summary Get all field values from given cli output.
     * @param cliTable CLI table. e.g: 'runai list node' output
     * @param columnName column in cliResponse.
     * @param isTableContainsSeparator is there a separator on the second row.
     */
    getColumn(cliTable: string, columnName: string, isTableContainsSeparator = true): string[] {
        let parsedTable: Record<string, string>[] = CliUtils.parseTableOutput(cliTable, 1);
        if (isTableContainsSeparator) {
            parsedTable = parsedTable.slice(1);
        }
        return parsedTable.map((element) => {
            return element[columnName];
        });
    }
}
