<#import "template.ftl" as layout>
<@layout.registrationLayout; section>
<#if section = "header">
${msg("logoutConfirmTitle")}
<#elseif section = "form">
<style>
    .login-pf{
        display: none; /* Hides the button and question */
    }
</style>
<div id="kc-logout-confirm" class="content-area">
    <p class="instruction">${msg("logoutConfirmHeader")}</p>

    <form class="form-actions" action="${url.logoutConfirmAction}" method="POST" id="kc-logout-form">
        <input type="hidden" name="session_code" value="${logoutConfirm.code}">
        <div class="${properties.kcFormGroupClass!}">
            <div id="kc-form-options">
                <div class="${properties.kcFormOptionsWrapperClass!}">
                </div>
            </div>

            <div id="kc-form-buttons" class="${properties.kcFormGroupClass!}">
                <input tabindex="4"
                       class="${properties.kcButtonClass!} ${properties.kcButtonPrimaryClass!} ${properties.kcButtonBlockClass!} ${properties.kcButtonLargeClass!}"
                       name="confirmLogout" id="kc-logout" type="submit" value="${msg("doLogout")}"/>
            </div>

        </div>
    </form>
    <script>
        window.onload = function() {
            document.getElementById('kc-logout').click();
        };
    </script>
    <div id="kc-info-message">
        <#if logoutConfirm.skipLink>
        <#else>
        <#if (client.baseUrl)?has_content>
        <p><a href="${client.baseUrl}">${kcSanitize(msg("backToApplication"))?no_esc}</a></p>
        </#if>
        </#if>
    </div>

    <div class="clearfix"></div>
</div>
</#if>
</@layout.registrationLayout>
