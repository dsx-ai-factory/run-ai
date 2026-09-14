<#-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved. -->
<#-- SPDX-License-Identifier: Apache-2.0 -->
<#import "template.ftl" as layout>
<#assign isExpiredCodeMessage = message.summary?? && message.summary == msg("expiredActionTokenNoSessionMessage")>
<@layout.registrationLayout displayMessage=false; section>
    <#if section = "header">
        <#if isExpiredCodeMessage>
            ${kcSanitize(msg("actionExpiredTitle"))?no_esc}
        <#else>
            ${kcSanitize(msg("errorTitle"))?no_esc}
        </#if>
    <#elseif section = "form">
        <div id="kc-content-wrapper">
            <#if isExpiredCodeMessage>
                <#assign tenantUriScheme = "https">
                <#if url.loginUrl?? && url.loginUrl?contains("://")>
                    <#assign tenantUriScheme = url.loginUrl?substring(0, url.loginUrl?index_of("://"))>
                </#if>
                <div id="kc-expired-content">
                <h2 id="kc-expired-link-title">${msg("actionExpiredTitle")}</h2>
                <p id="kc-expired-link-subtitle">${msg("actionExpiredMessage")}</p>
                <div id="kc-form-buttons" class="${properties.kcFormButtonsClass!}">
                    <a id="resend-reset-password" class="${properties.kcButtonClass!} ${properties.kcButtonPrimaryClass!} ${properties.kcButtonBlockClass!} ${properties.kcButtonLargeClass!}" href="#" data-realm="${realm.name}" data-reset-url="${url.loginResetCredentialsUrl}" data-tenant-scheme="${tenantUriScheme}">
                        ${msg("resendResetPassword")}
                    </a>
                </div>
                </div>
                <script>
                    (function() {
                        var link = document.getElementById('resend-reset-password');
                        if (!link) return;

                        link.addEventListener('click', function(e) {
                            e.preventDefault();

                            var resetUrl = (this.getAttribute('data-reset-url') || '').trim();
                            if (!resetUrl) {
                                window.location.href = window.location.origin + '/';
                                return;
                            }

                            var realm = this.getAttribute('data-realm') || '';
                            var hostname = window.location.hostname;
                            var tenantHost = (!realm || realm === 'runai' || hostname === realm || hostname.startsWith(realm + '.'))
                                ? hostname
                                : realm + '.' + hostname;

                            var schemeAttr = (this.getAttribute('data-tenant-scheme') || window.location.protocol).replace(/:$/, '');
                            var scheme = (schemeAttr === 'http' || schemeAttr === 'https') ? schemeAttr : 'https';

                            var tenantRedirectUri = (scheme + '://' + tenantHost).replace(/\/+$/, '');
                            var url = resetUrl.startsWith('http') ? resetUrl : window.location.origin + resetUrl;
                            url += (url.includes('?') ? '&' : '?') + 'client_id=runai-admin-ui&redirect_uri=' + encodeURIComponent(tenantRedirectUri);
                            window.location.href = url;
                        });
                    })();
                </script>
            <#else>
                ${kcSanitize(message.summary)?no_esc}
                <#if !skipLink??>
                    <#if client?? && client.baseUrl?has_content>
                        <p><a href="${client.baseUrl}">${kcSanitize(msg("backToApplication"))?no_esc}</a></p>
                    <#else>
                        <p><a href="${url.loginUrl}">${kcSanitize(msg("backToLogin"))?no_esc}</a></p>
                    </#if>
                </#if>
            </#if>
        </div>
    </#if>
</@layout.registrationLayout>
