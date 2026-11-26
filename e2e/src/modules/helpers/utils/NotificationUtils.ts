/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import axios from 'axios';
import { StatusCodes } from 'http-status-codes';
import { notificationsProxyApi } from '../../api';
import { logger } from '../logger/logger';
import { authUtils, Utils } from '..';
import { EmailChannel, EmailChannelForPatch, NotificationState } from '../../api/apis/notificationsProxyApi';
import { saasConsoleUrl, tenantClientSecret } from '../config/config';
import { TokenResponse } from '../../api/apis/identityManagerApi/generated/api';
import { settings } from '../config/config';


// Mailtrap credentials must be supplied by the test environment.
export const SMTP_CONFIG = {
    host: process.env.MAILTRAP_HOST || 'sandbox.smtp.mailtrap.io',
    port: process.env.MAILTRAP_PORT || '587',
    user: process.env.MAILTRAP_USER || '',
    password: process.env.MAILTRAP_PASSWORD || '',
    inboxId: process.env.MAILTRAP_INBOX_ID || '',
    apiToken: process.env.MAILTRAP_API_TOKEN || ''
};

export class Notifications {
    async createEmailChannel(accessToken: string, channel: EmailChannel): Promise<void> {
        notificationsProxyApi.emailApi.accessToken = accessToken;
        try {
            const result = await notificationsProxyApi.emailApi.getEmail();
            if (result.body) {
                await notificationsProxyApi.emailApi.patchEmail(channel as EmailChannelForPatch);
            } else {
                await notificationsProxyApi.emailApi.createEmail(channel);
            }
        } catch (error: any) {
            if (error.statusCode === StatusCodes.NOT_FOUND) {
                await notificationsProxyApi.emailApi.createEmail(channel);
            } else {
                throw error;
            }
        }
    }

    async deleteEmailChannel(accessToken: string): Promise<void> {
        notificationsProxyApi.emailApi.accessToken = accessToken;
        await notificationsProxyApi.emailApi.deleteEmail();
    }

    async getEmailChannel(accessToken: string): Promise<EmailChannel> {
        notificationsProxyApi.emailApi.accessToken = accessToken;
        const result = await notificationsProxyApi.emailApi.getEmail();
        return result.body;
    }

    async patchEmailChannel(accessToken: string, channel: EmailChannelForPatch): Promise<void> {
        notificationsProxyApi.emailApi.accessToken = accessToken;
        await notificationsProxyApi.emailApi.patchEmail(channel);
    }

    async createOrPatchEmailChannel(accessToken: string, channel: EmailChannel): Promise<void> {
        try {
            await this.patchEmailChannel(accessToken, channel as EmailChannelForPatch);
        } catch (error: any) {
            if (error.statusCode === StatusCodes.NOT_FOUND) {
                logger.info('Email channel not found, creating it');
                await this.createEmailChannel(accessToken, channel);
                return;
            }
            logger.error('Error patching email channel:', error);
            throw error;
        }
    }

    async checkEmailReceived(
        emailAddress: string,
        validSubjectHandler: (a: string) => boolean
    ): Promise<boolean> {
        const { data: messages } = await axios.get(
            `https://mailtrap.io/api/v1/inboxes/${SMTP_CONFIG.inboxId}/messages`,
            {
                headers: { 'Api-Token': SMTP_CONFIG.apiToken }
            }
        );

        const emailMessage = messages.find(
            (msg: any) => msg.to_email === emailAddress && validSubjectHandler(msg.subject)
        );

        if (!emailMessage) {
            logger.info(`Email message not received for ${emailAddress}`);
            return false;
        }
        logger.info(`Email message received for ${emailAddress}: ${emailMessage.subject}`);
        return true;
    }

    async checkEmailReceivedWithDelay(
        validSubjectHandler: (str: string) => boolean,
        emailAddress: string,
        attempts = 25,
        waitIntervalMs = 10000
    ): Promise<boolean> {
        for (let attempt = 0; attempt < attempts; attempt++) {
            await Utils.delay(waitIntervalMs);
            const emailReceived = await this.checkEmailReceived(emailAddress, validSubjectHandler);
            if (emailReceived) {
                return true;
            }
        }
        return false;
    }

    async fetchEmailBodyContent(mailtrapApiToken: string, htmlPath: string): Promise<string> {
        const { data: htmlContent } = await axios.get(`https://mailtrap.io${htmlPath}`, {
            headers: { 'Api-Token': mailtrapApiToken }
        });
        return htmlContent;
    }

    validateClickableLinksInEmail(htmlContent: string): boolean {
        const linkRegex = /<a[^>]+href\s*=\s*["']([^"']+)["'][^>]*>/gi;
        const matches = htmlContent.match(linkRegex);
        return !!(matches && matches.length > 0);
    }

    extractLinksFromEmail(htmlContent: string): string[] {
        const linkRegex = /<a[^>]+href\s*=\s*["']([^"']+)["'][^>]*>/gi;
        const matches = htmlContent.match(linkRegex) || [];

        return matches
            .map((match) => {
                const hrefMatch = match.match(/href\s*=\s*["']([^"']+)["']/i);
                return hrefMatch ? hrefMatch[1] : '';
            })
            .filter((link) => link);
    }

    extractSignInLink(htmlContent: string): string | null {
        const linkRegex = /<a[^>]+href\s*=\s*["']([^"']+)["'][^>]*>/gi;
        const matches = htmlContent.match(linkRegex) || [];

        // First priority: authentication token links
        for (const match of matches) {
            const hrefMatch = match.match(/href\s*=\s*["']([^"']+)["']/i);
            if (hrefMatch && hrefMatch[1].includes('login-actions/action-token')) {
                return hrefMatch[1];
            }
        }

        // Second priority: auth realm links (non-nvidia)
        for (const match of matches) {
            const hrefMatch = match.match(/href\s*=\s*["']([^"']+)["']/i);
            if (hrefMatch) {
                const url = hrefMatch[1];
                if (url.includes('auth/realms') && !url.includes('nvidia.com')) {
                    return url;
                }
            }
        }

        return null;
    }

    extractSignInLinkByText(htmlContent: string): string | null {
        // Try direct authentication token link first
        const authTokenLink = this.extractSignInLink(htmlContent);
        if (authTokenLink) {
            return authTokenLink;
        }

        // Look for links with sign-in related text
        const signInLinkRegex =
            /<a[^>]+href\s*=\s*["']([^"']+)["'][^>]*>[\s\S]*?(?:SIGN IN|Sign In|sign in|Set.*Password|Reset.*Password)[\s\S]*?<\/a>/gi;
        const matches = htmlContent.match(signInLinkRegex);

        if (matches?.length) {
            // Check for auth tokens first
            for (const match of matches) {
                const hrefMatch = match.match(/href\s*=\s*["']([^"']+)["']/i);
                if (hrefMatch?.[1]?.includes('login-actions/action-token')) {
                    return hrefMatch[1];
                }
            }

            // Return first match if no auth token
            const hrefMatch = matches[0].match(/href\s*=\s*["']([^"']+)["']/i);
            return hrefMatch?.[1] || null;
        }

        return null;
    }

    isValidUrl(url: string): boolean {
        try {
            new URL(url);
            return true;
        } catch {
            return false;
        }
    }

    async validateLinks(
        links: string[],
        testAccessibility: boolean = false
    ): Promise<{
        validLinks: string[];
        invalidLinks: string[];
        accessibleLinks: string[];
        inaccessibleLinks: string[];
    }> {
        const validLinks = links.filter((link) => this.isValidUrl(link));
        const invalidLinks = links.filter((link) => !this.isValidUrl(link));
        const accessibleLinks: string[] = [];
        const inaccessibleLinks: string[] = [];

        if (testAccessibility) {
            for (const link of validLinks) {
                try {
                    await axios.head(link, {
                        timeout: 5000,
                        validateStatus: (status) => status < 400
                    });
                    accessibleLinks.push(link);
                } catch {
                    inaccessibleLinks.push(link);
                }
            }
        }

        return { validLinks, invalidLinks, accessibleLinks, inaccessibleLinks };
    }

    async getEmailLinksAndValidate(
        mailtrapInboxId: string,
        mailtrapApiToken: string,
        emailAddress: string,
        validSubjectHandler: (a: string) => boolean,
        testAccessibility: boolean = false
    ): Promise<{
        emailReceived: boolean;
        links: string[];
        linkValidation: {
            validLinks: string[];
            invalidLinks: string[];
            accessibleLinks: string[];
            inaccessibleLinks: string[];
        };
    } | null> {
        try {
            const { data: messages } = await axios.get(
                `https://mailtrap.io/api/v1/inboxes/${mailtrapInboxId}/messages`,
                { headers: { 'Api-Token': mailtrapApiToken } }
            );

            const emailMessage = messages.find(
                (msg: any) => msg.to_email === emailAddress && validSubjectHandler(msg.subject)
            );

            if (!emailMessage) {
                return {
                    emailReceived: false,
                    links: [],
                    linkValidation: {
                        validLinks: [],
                        invalidLinks: [],
                        accessibleLinks: [],
                        inaccessibleLinks: []
                    }
                };
            }

            const htmlContent = await this.fetchEmailBodyContent(mailtrapApiToken, emailMessage.html_path);
            const links = this.extractLinksFromEmail(htmlContent);
            const linkValidation = await this.validateLinks(links, testAccessibility);

            return { emailReceived: true, links, linkValidation };
        } catch (error) {
            logger.error(`Error getting email links for ${emailAddress}:`, error);
            return null;
        }
    }

    async getSignInLinkFromEmail(
        mailtrapInboxId: string,
        mailtrapApiToken: string,
        emailAddress: string,
        validSubjectHandler: (a: string) => boolean
    ): Promise<string | null> {
        try {
            const { data: messages } = await axios.get(
                `https://mailtrap.io/api/v1/inboxes/${mailtrapInboxId}/messages`,
                { headers: { 'Api-Token': mailtrapApiToken } }
            );

            const emailMessage = messages.find(
                (msg: any) => msg.to_email === emailAddress && validSubjectHandler(msg.subject)
            );

            if (!emailMessage) return null;

            const htmlContent = await this.fetchEmailBodyContent(mailtrapApiToken, emailMessage.html_path);
            const link = this.extractSignInLinkByText(htmlContent);
            return link ? link.replaceAll('..', '.') : null;
        } catch (error) {
            logger.error(`Error getting sign-in link for ${emailAddress}:`, error);
            return null;
        }
    }

    async getSignInLinkFromEmailWithDelay(
        mailtrapInboxId: string,
        mailtrapApiToken: string,
        emailAddress: string,
        validSubjectHandler: (a: string) => boolean,
        attempts = 25
    ): Promise<string | null> {
        for (let attempt = 0; attempt < attempts; attempt++) {
            await Utils.delay(10000);
            const signInLink = await this.getSignInLinkFromEmail(
                mailtrapInboxId,
                mailtrapApiToken,
                emailAddress,
                validSubjectHandler
            );
            if (signInLink) return signInLink;
        }
        return null;
    }

    async checkEmailReceivedWithLinks(
        mailtrapInboxId: string,
        mailtrapApiToken: string,
        emailAddress: string,
        validSubjectHandler: (a: string) => boolean
    ): Promise<boolean> {
        try {
            const { data: messages } = await axios.get(
                `https://mailtrap.io/api/v1/inboxes/${mailtrapInboxId}/messages`,
                { headers: { 'Api-Token': mailtrapApiToken } }
            );

            const emailMessage = messages.find(
                (msg: any) => msg.to_email === emailAddress && validSubjectHandler(msg.subject)
            );

            if (!emailMessage) return false;

            const htmlContent = await this.fetchEmailBodyContent(mailtrapApiToken, emailMessage.html_path);
            return this.validateClickableLinksInEmail(htmlContent);
        } catch (error) {
            logger.error(`Error checking email with links for ${emailAddress}:`, error);
            return false;
        }
    }

    async checkEmailReceivedWithLinksAndDelay(
        // eslint-disable-next-line @typescript-eslint/no-unused-vars
        validSubjectHandler = (str: string): boolean => true,
        emailAddress: string,
        attempts = 25,
        mailtrapInboxId: string,
        mailtrapApiToken: string
    ): Promise<boolean> {
        for (let attempt = 0; attempt < attempts; attempt++) {
            await Utils.delay(10000);
            const emailReceivedWithLinks = await this.checkEmailReceivedWithLinks(
                mailtrapInboxId,
                mailtrapApiToken,
                emailAddress,
                validSubjectHandler
            );
            if (emailReceivedWithLinks) return true;
        }
        return false;
    }

    async getSystemScopedSuperAdminToken(): Promise<TokenResponse> {
        return await authUtils.authenticateSuperAdminAppAndGetToken(
            'default_e2e',
            tenantClientSecret,
            saasConsoleUrl
        );
    }

    getSmtpChannel(recipients: string[]): EmailChannel {
        return {
            config: {
                smtpHost: SMTP_CONFIG.host,
                smtpPort: +SMTP_CONFIG.port,
                username: SMTP_CONFIG.user,
                password: SMTP_CONFIG.password,
                recipients: recipients,
                directNotifications: true,
                smtpTlsEnabled: true,
                subject: '{{ .title }}'
            }
        };
    }

    async enableNotifications(accessToken: string): Promise<void> {
        notificationsProxyApi.notificationStateApi.accessToken = accessToken;
        const state = new NotificationState();
        state.email = { enabled: true };
        await notificationsProxyApi.notificationStateApi.updateNotificationState(state);
        logger.info('Notifications enabled successfully');
    }

    async disableNotifications(accessToken: string): Promise<void> {
        notificationsProxyApi.notificationStateApi.accessToken = accessToken;
        const state = new NotificationState();
        state.email = { enabled: false };
        await notificationsProxyApi.notificationStateApi.updateNotificationState(state);
        logger.info('Notifications disabled successfully');
    }

    async restoreSaasEmailChannel(): Promise<void> {
        logger.info(`going to patch email channel with settings: ${settings.smtpHost}, ${settings.smtpPort}, ${settings.smtpUsername}`);
        if (settings.isSelfHosted) {
            logger.info('skipping email channel patching for self hosted environment');
            return;
        }

        const systemScopedToken = await this.getSystemScopedSuperAdminToken();
        await this.patchEmailChannel(systemScopedToken.accessToken, {
            config: {
                smtpHost: settings.smtpHost,
                smtpPort: +settings.smtpPort,
                username: settings.smtpUsername,
                password: settings.smtpPassword,
                subject: 'NVIDIA Run:ai notification - {{ .category }} {{ .reason }}',
                recipients: [],
                directNotifications: true,
                smtpTlsEnabled: true,
            }
        });
    }
}
