/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { logger } from '../helpers';
import { Client } from 'pg';
import { kubeCli } from '../cli';
import { namespaceCP, helmRepoName } from '../helpers/config/const';
import { HelmReleaseUtils, resolveHelmSecretRef } from '../helpers/utils/HelmReleaseUtils';

const helmReleaseUtils = new HelmReleaseUtils();


export interface DatabaseConfig {
    host: string;
    port: number;
    database: string;
    user: string;
    password: string;
    schema?: string;
}

export interface QueryResult {
    rows: any[];
    rowCount: number;
}

export class DatabaseConnection {
    private config: DatabaseConfig;

    private client: any = null;

    constructor(config?: Partial<DatabaseConfig>) {
        this.config = {
            host: config?.host || process.env.DB_HOST || '127.0.0.1',
            port: config?.port || parseInt(process.env.DB_PORT || '5432'),
            database: config?.database || process.env.DB_NAME || 'backend',
            user: config?.user || process.env.DB_USER || 'user',
            password: config?.password || process.env.DB_PASSWORD || '',
            schema: config?.schema || 'tenants_manager'
        };
    }

    private async resolvePassword(): Promise<string> {
        if (this.config.password) {
            return this.config.password;
        }

        try {
            const values = helmReleaseUtils.helmGetValues(helmRepoName, namespaceCP);
            const { secretName, key } = resolveHelmSecretRef(
                values?.global?.postgresql?.auth,
                'runai-backend-gen-postgresql',
                'password',
                'userPasswordKey'
            );
            const secret = await kubeCli.getSecret(secretName, namespaceCP);
            const passwordB64 = secret.data?.[key];
            if (passwordB64) {
                logger.debug(`Fetched DB password from cluster secret ${secretName}`);
                return Buffer.from(passwordB64, 'base64').toString('utf-8');
            }
        } catch (error) {
            logger.warn(`Failed to resolve DB password from helm values: ${error.message}`);
        }

        throw new Error(
            `No DB password available. Set DB_PASSWORD env var or ensure kubectl access to runai-backend-gen-postgresql in namespace ${namespaceCP}`
        );
    }

    /**
     * Connect to the PostgreSQL database
     */
    async connect(): Promise<void> {
        try {
            this.config.password = await this.resolvePassword();

            this.client = new Client({
                host: this.config.host,
                port: this.config.port,
                database: this.config.database,
                user: this.config.user,
                password: this.config.password
            });

            await this.client.connect();

            if (this.config.schema) {
                await this.client.query(`SET search_path TO ${this.config.schema}`);
            }

            logger.info(
                `Connected to PostgreSQL database: ${this.config.host}:${this.config.port}/${this.config.database}`
            );
        } catch (error) {
            logger.error('Failed to connect to database:', error);
            throw error;
        }
    }

    /**
     * Disconnect from the database
     */
    async disconnect(): Promise<void> {
        if (this.client) {
            await this.client.end();
            this.client = null;
            logger.info('Disconnected from database');
        }
    }

    /**
     * Execute a query
     */
    async query(sql: string, params?: any[]): Promise<QueryResult> {
        if (!this.client) {
            throw new Error('Database not connected. Call connect() first.');
        }

        try {
            logger.debug(`Executing query: ${sql}`, params);
            const result = await this.client.query(sql, params);
            return {
                rows: result.rows,
                rowCount: result.rowCount
            };
        } catch (error) {
            logger.error('Query execution failed:', error);
            throw error;
        }
    }

    /**
     * Check if tenant exists in the database
     */
    async checkTenantExists(tenantId: number): Promise<boolean> {
        const result = await this.query('SELECT COUNT(*) as count FROM tenants WHERE id = $1', [tenantId]);
        return parseInt(result.rows[0].count) > 0;
    }

    /**
     * Check if tenant exists by name
     */
    async checkTenantExistsByName(tenantName: string): Promise<boolean> {
        const result = await this.query('SELECT COUNT(*) as count FROM tenants WHERE name = $1', [
            tenantName
        ]);
        return parseInt(result.rows[0].count) > 0;
    }

    /**
     * Get tenant data by ID
     */
    async getTenantById(tenantId: number): Promise<any> {
        const result = await this.query('SELECT * FROM tenants WHERE id = $1', [tenantId]);
        return result.rows[0] || null;
    }

    /**
     * Get tenant data by name
     */
    async getTenantByName(tenantName: string): Promise<any> {
        const result = await this.query('SELECT * FROM tenants WHERE name = $1', [tenantName]);
        return result.rows[0] || null;
    }

    /**
     * Check if tenant was properly deleted (soft delete check)
     */
    async checkTenantDeletionStatus(tenantId: number): Promise<{
        exists: boolean;
        deletedAt: string | null;
        status: string | null;
    }> {
        const result = await this.query('SELECT deleted_at, status FROM tenants WHERE id = $1', [tenantId]);

        if (result.rows.length === 0) {
            return { exists: false, deletedAt: null, status: null };
        }

        const tenant = result.rows[0];
        return {
            exists: true,
            deletedAt: tenant.deleted_at,
            status: tenant.status
        };
    }

    /**
     * Get all tenant-related data that should be cleaned up after deletion
     */
    async getTenantCleanupData(tenantId: number): Promise<{
        projects: any[];
        workloads: any[];
        users: any[];
        settings: any[];
    }> {
        const projects = await this.query('SELECT * FROM projects WHERE tenant_id = $1', [tenantId]);

        const workloads = await this.query('SELECT * FROM workloads WHERE tenant_id = $1', [tenantId]);

        const users = await this.query('SELECT * FROM users WHERE tenant_id = $1', [tenantId]);

        const settings = await this.query('SELECT * FROM tenant_settings WHERE tenant_id = $1', [tenantId]);

        return {
            projects: projects.rows,
            workloads: workloads.rows,
            users: users.rows,
            settings: settings.rows
        };
    }

    /**
     * Get database connection info
     */
    getConnectionInfo(): Omit<DatabaseConfig, 'password'> {
        return {
            host: this.config.host,
            port: this.config.port,
            database: this.config.database,
            user: this.config.user,
            schema: this.config.schema
        };
    }

    /**
     * Get comprehensive tenant data count across all schemas and tables
     * This method dynamically finds all tables with tenant_id column and counts records
     */
    async getTenantDataCountAcrossAllTables(tenantId: number): Promise<{
        totalCount: number;
        tableBreakdown: Array<{
            schema: string;
            table: string;
            count: number;
            fullTableName: string;
        }>;
        summary: string[];
    }> {
        try {
            // Get all tables with tenant_id column from all relevant schemas
            const tablesQuery = `
                SELECT 
                    table_schema,
                    table_name,
                    table_schema || '.' || table_name AS full_table_name
                FROM information_schema.columns
                WHERE column_name = 'tenant_id'
                  AND table_schema NOT IN ('pg_catalog', 'information_schema', 'public')
                ORDER BY table_schema, table_name
            `;

            const tablesResult = await this.query(tablesQuery);
            const tableBreakdown: Array<{
                schema: string;
                table: string;
                count: number;
                fullTableName: string;
            }> = [];
            let totalCount = 0;
            const summary: string[] = [];

            logger.info(`Found ${tablesResult.rows.length} tables with tenant_id column`);

            // Query each table for the specific tenant
            for (const tableRow of tablesResult.rows) {
                const { table_schema, table_name, full_table_name } = tableRow;

                try {
                    const countQuery = `
                        SELECT COUNT(*) as count 
                        FROM "${table_schema}"."${table_name}" 
                        WHERE tenant_id = $1
                    `;

                    const countResult = await this.query(countQuery, [tenantId]);
                    const count = parseInt(countResult.rows[0].count);

                    if (count > 0) {
                        tableBreakdown.push({
                            schema: table_schema,
                            table: table_name,
                            count: count,
                            fullTableName: full_table_name
                        });

                        totalCount += count;
                        summary.push(`${full_table_name}: ${count} records`);
                        logger.info(`Found ${count} records in ${full_table_name} for tenant ${tenantId}`);
                    } else {
                        logger.debug(`No records found in ${full_table_name} for tenant ${tenantId}`);
                    }
                } catch (error) {
                    logger.warn(
                        `Could not query table ${full_table_name} for tenant ${tenantId}:`,
                        error.message
                    );
                    summary.push(`${full_table_name}: ERROR - ${error.message}`);
                }
            }

            return {
                totalCount,
                tableBreakdown: tableBreakdown.sort((a, b) => b.count - a.count), // Sort by count descending
                summary
            };
        } catch (error) {
            logger.error('Failed to get tenant data count across all tables:', error);
            throw error;
        }
    }

    /**
     * Generate SQL queries to check all tables with tenant_id (for manual execution)
     */
    async generateTenantDataQueries(tenantId: number): Promise<string[]> {
        try {
            const tablesQuery = `
                SELECT 
                    'SELECT ''' || table_schema || '.' || table_name || ''' AS table_full_name, COUNT(*) AS count FROM ' 
                    || quote_ident(table_schema) || '.' || quote_ident(table_name) 
                    || ' WHERE tenant_id = ' || $1 || ';' AS query_sql
                FROM information_schema.columns
                WHERE column_name = 'tenant_id'
                  AND table_schema NOT IN ('pg_catalog', 'information_schema', 'public')
                ORDER BY table_schema, table_name
            `;

            const result = await this.query(tablesQuery, [tenantId]);
            return result.rows.map((row) => row.query_sql);
        } catch (error) {
            logger.error('Failed to generate tenant data queries:', error);
            throw error;
        }
    }

    /**
     * Get services that should be checked for tenant data (based on codebase analysis)
     */
    getRelevantServicesForTenantCheck(): string[] {
        return [
            'assets',
            'audit',
            'authorization',
            'backend',
            'cluster_service',
            'datavolumes',
            'identity_manager',
            'k8s_objects_tracker',
            'metrics',
            'notifications_service',
            'org_unit_service',
            'policies',
            'tenants_manager',
            'workloads'
        ];
    }
}
