<#-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved. -->
<#-- SPDX-License-Identifier: Apache-2.0 -->
<#assign selectedIdp = "">
<#import "template.ftl" as layout>
<@layout.registrationLayout displayMessage=!messagesPerField.existsError('username','password') displayInfo=realm.password && realm.registrationAllowed && !registrationDisabled??; section>
    <#if section = "header">
        ${msg("loginAccountTitle")}
    <#elseif section = "form">
        <div id="kc-content-wrapper">
            <div class="alert-error pf-c-alert pf-m-inline pf-m-danger">
                <span id="error-container" class="pf-c-alert__title kc-feedback-text"></span>
            </div>
            <div id="kc-form">
                <#if realm.password>
                    <div id="kc-form-wrapper">
                        <form id="kc-form-login" onsubmit="login.disabled = true; return true;" action="${url.loginAction}" method="post">
                            <div class="${properties.kcFormGroupClass!}">
                                <script lang="javascript">
                                    function getQueryParams() {
                                        const params = new URLSearchParams(window.location.search);
                                        return {
                                            error: params.get('error'),
                                            errorDescription: params.get('error_description')
                                        };
                                    }

                                    function displayErrorMessagesIfNeed() {
                                        const { errorDescription } = getQueryParams();
                                        const errorContainer = document.getElementById('error-container');

                                        if (errorDescription) {
                                            errorContainer.textContent = errorDescription;
                                        } else {
                                            errorContainer.style.display = 'none';
                                        }
                                    }
                                    displayErrorMessagesIfNeed();
                                </script>
                                <label for="username" class="${properties.kcLabelClass!} label-color"><#if !realm.loginWithEmailAllowed>${msg("username")}<#elseif !realm.registrationEmailAsUsername>${msg("usernameOrEmail")}<#else>${msg("email")}</#if></label>
                                <#if usernameEditDisabled??>
                                    <input tabindex="1" id="username" class="${properties.kcInputClass!} field-wrapper" name="username" value="${(login.username!'')}" type="text" disabled />
                                <#else>
                                    <input tabindex="1" id="username" class="${properties.kcInputClass!} field-wrapper" name="username" value="${(login.username!'')}"  type="text" autofocus autocomplete="off"
                                           aria-invalid="<#if messagesPerField.existsError('username','password')>true</#if>"
                                    />
                                    <#if messagesPerField.existsError('username','password')>
                                        <span id="input-error" class="${properties.kcInputErrorMessageClass!}" aria-live="polite">
                                            ${kcSanitize(messagesPerField.getFirstError('username','password'))?no_esc}
                                    </span>
                                    </#if>
                                </#if>
                            </div>
                            <div class="${properties.kcFormGroupClass!}">
                                <label for="password" class="${properties.kcLabelClass!} label-color">${msg("password")}</label>
                                <input tabindex="2" id="password" class="${properties.kcInputClass!} field-wrapper" name="password" type="password" autocomplete="off"
                                       aria-invalid="<#if messagesPerField.existsError('username','password')>true</#if>"
                                />
                            </div>
                            <div class="${properties.kcFormGroupClass!} ${properties.kcFormSettingClass!}">
                                <div id="kc-form-options">
                                    <#if realm.rememberMe && !usernameEditDisabled??>
                                        <div class="checkbox">
                                            <label>
                                                <#if login.rememberMe??>
                                                    <input tabindex="3" id="rememberMe" name="rememberMe" type="checkbox" checked> ${msg("rememberMe")}
                                                <#else>
                                                    <input tabindex="3" id="rememberMe" name="rememberMe" type="checkbox"> ${msg("rememberMe")}
                                                </#if>
                                            </label>
                                        </div>
                                    </#if>
                                </div>
                                <div class="${properties.kcFormOptionsWrapperClass!}">
                                    <#if realm.resetPasswordAllowed>
                                        <span><a tabindex="5" href="${url.loginResetCredentialsUrl}">${msg("doForgotPassword")}</a></span>
                                    </#if>
                                </div>
                            </div>
                            <div id="kc-form-buttons" class="${properties.kcFormGroupClass!}">
                                <input type="hidden" id="id-hidden-input" name="credentialId" <#if auth.selectedCredential?has_content>value="${auth.selectedCredential}"</#if>/>
                                <input tabindex="4" class="${properties.kcButtonClass!} ${properties.kcButtonPrimaryClass!} ${properties.kcButtonBlockClass!} ${properties.kcButtonLargeClass!}" name="login" id="kc-login" type="submit" value="${msg("doLogIn")}"/>
                            </div>
                        </form>
                        <#if social?? && social.providers?has_content>
                            <div class="line">OR</div>
                            <a id="sso-button-id" class="${properties.kcButtonClass!} ${properties.kcButtonPrimaryClass!} ${properties.kcButtonBlockClass!} ${properties.kcButtonLargeClass!} sso-button field-wrapper" type="button" href="${social.providers[0].loginUrl}">
                                <span class="sso-button-text">
                                        <svg width="16" height="16" viewBox="0 0 256 256" fill="#76B900" xmlns="http://www.w3.org/2000/svg" aria-hidden="true" focusable="false">
                                            <path d="M216.57,39.43A80,80,0,0,0,83.91,120.78L28.69,176A15.86,15.86,0,0,0,24,187.31V216a16,16,0,0,0,16,16H72a8,8,0,0,0,8-8V208H96a8,8,0,0,0,8-8V184h16a8,8,0,0,0,5.66-2.34l9.56-9.57A79.73,79.73,0,0,0,160,176h.1A80,80,0,0,0,216.57,39.43ZM180,92a16,16,0,1,1,16-16A16,16,0,0,1,180,92Z"/>
                                        </svg>
                                        ${msg("loginWithSSO")}
                                </span>
                            </a>
                        </#if>
                    </div>
                </#if>
            </div>
        </div>
    <#elseif section = "info" >
        <#if realm.password && realm.registrationAllowed && !registrationDisabled??>
            <div id="kc-registration-container">
                <div id="kc-registration">
                    <span>${msg("noAccount")} <a tabindex="6" href="${url.registrationUrl}">${msg("doRegister")}</a></span>
                </div>
            </div>
        </#if>
    </#if>
</@layout.registrationLayout>
