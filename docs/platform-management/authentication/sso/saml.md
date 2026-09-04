# Set Up SSO with SAML

Single Sign-On (SSO) is an authentication scheme, allowing users to log in with a single pair of credentials to multiple, independent software systems.

This guide explains the procedure to [configure SSO to NVIDIA Run:ai](../overview.md#single-sign-on-sso) using the SAML 2.0 protocol.

## Prerequisites

Before you start, make sure you have the **IDP Metadata** **XML** available from your identity provider.

## Setup

### Adding the Identity Provider

1. Go to **General settings**
2. Open the Security section and click **+IDENTITY PROVIDER**
3. Select **Custom SAML 2.0**
4. Select either **From computer** or **From URL** to upload your identity provider metadata file
   * **From computer** - Click the Metadata XML file field, then select your file for upload
   * **From URL** - In the Metadata XML field, enter the URL to the IDP Metadata XML file
5. You can either copy the **Redirect URL** and **Entity ID** displayed on the screen and enter them in your identity provider, or use the **service provider metadata XML**, which contains the same information in XML format. This file becomes available after you click **SAVE** in step 7.
6. Optional: Enter the user attributes and their value in the identity provider as shown in the below table
7. Click **SAVE.** After save, click **Open service provider metadata XML** to access the metadata file. This file can be used to configure your identity provider.
8. Optional: Enable **Auto-Redirect to SSO** to automatically redirect users to your configured identity provider’s login page when accessing the platform.

| Attribute            | Default value in NVIDIA Run:ai | Description                                                                                                                                                                                                  |
| -------------------- | ------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| User role groups     | GROUPS                         | If it exists in the IDP, it allows you to assign NVIDIA Run:ai role groups via the IDP. The IDP attribute must be a list of strings.                                                                         |
| Linux User ID        | UID                            | If it exists in the IDP, it allows Researcher containers to start with the Linux User UID. Used to map access to network resources such as file systems to users. The IDP attribute must be of type integer. |
| Linux Group ID       | GID                            | If it exists in the IDP, it allows Researcher containers to start with the Linux Group GID. The IDP attribute must be of type integer.                                                                       |
| Supplementary Groups | SUPPLEMENTARYGROUPS            | If it exists in the IDP, it allows Researcher containers to start with the relevant Linux supplementary groups. The IDP attribute must be a list of integers.                                                |
| Email                | email                          | Defines the user attribute in the IDP holding the user's email address, which is the user identifier in NVIDIA Run:ai.                                                                                       |
| User first name      | firstName                      | Used as the user’s first name appearing in the NVIDIA Run:ai platform.                                                                                                                                       |
| User last name       | lastName                       | Used as the user’s last name appearing in the NVIDIA Run:ai platform.                                                                                                                                        |

### Testing the Setup

1. Open the NVIDIA Run:ai platform as an admin
2. Add [access rules](../accessrules.md) to an SSO user defined in the IDP
3. Open the NVIDIA Run:ai platform in an incognito browser tab
4. On the sign-in page click **CONTINUE WITH SSO.**\
   You are redirected to the identity provider sign in page
5. In the identity provider sign-in page, log in with the SSO user who you granted with access rules
6. If you are unsuccessful signing-in to the identity provider, follow the [Troubleshooting](saml.md#troubleshooting) section below

### Editing the Identity Provider

You can view the identity provider details and edit its configuration:

1. Go to **General settings**
2. Open the Security section
3. On the identity provider box, click **Edit identity provider**
4. You can edit either the metadata file or the user attributes
5. You can view the identity provider URL, identity provider entity ID, and the certificate expiration date

### Removing the Identity Provider

You can remove the identity provider configuration:

1. Go to **General settings**
2. Open the Security section
3. On the identity provider card, click **Remove identity provider**
4. In the dialog, click **REMOVE** to confirm the action

<Info>

**Note**

To avoid losing access, removing the identity provider must be carried out by a local user.

</Info>

## Downloading the IDP Metadata XML File

You can download the XML file to view the identity provider settings:

1. Go to **General settings**
2. Open the Security section
3. On the identity provider card, click **Edit identity provider**
4. In the dialog, click **DOWNLOAD IDP METADATA XML FILE**

## Troubleshooting

If testing the setup was unsuccessful, try the different troubleshooting scenarios according to the error you received. If an error still occurs, check the [advanced troubleshooting section](saml.md#advanced-troubleshooting).

### Troubleshooting Scenarios

<Accordion title="Error: 'Invalid signature in response from identity provider'">

**Error: 'Invalid signature in response from identity provider'**

**Description**: After trying to log in, the following message is received in the NVIDIA Run:ai login page.

**Mitigation:**

1. Go to the **General settings** menu
2. Open the Security section
3. In the identity provider box, check for a "Certificate expired” error
4. If it is expired, update the SAML metadata file to include a valid certificate

</Accordion>

<Accordion title="Error: '401 - We’re having trouble identifying your account because your email is incorrect or can’t be found.'">

**Error: '401 - We’re having trouble identifying your account because your email is incorrect or can’t be found.'**

**Description:** Authentication failed because email attribute was not found.

**Mitigation**: Validate the user’s email attribute is mapped correctly

</Accordion>

<Accordion title="Error: '403 - Sorry, we can’t let you see this page. Something about permissions…'">

**Error: '403 - Sorry, we can’t let you see this page. Something about permissions…'**

**Description:** The authenticated user is missing permissions

**Mitigation**:

1. Validate either the user or its related group/s are assigned with [access rules](../accessrules.md)
2. Validate the user’s groups attribute is mapped correctly

**Advanced:**

1. Open the Chrome DevTools: Right-click on page → Inspect → Network tab
2. Navigate to the Clusters page to trigger an API request
3. In the Network tab, find the request to the Clusters API
4. In the request headers, copy the value of the Authorization header (the token starts with `Bearer`)
5. Paste the token value (without the `Bearer` prefix) in [https://jwt.io](https://jwt.io)
6. Under the Payload section validate the values of the user's attributes

</Accordion>

### Advanced Troubleshooting

<Accordion title="Validating the SAML request">

**Validating the SAML request**

The SAML login flow can be separated into two parts:

* NVIDIA Run:ai redirects to the IDP for log-ins using a SAML Request
* On successful log-in, the IDP redirects back to NVIDIA Run:ai with a SAML Response

Validate the SAML Request to ensure the SAML flow works as expected:

1. Go to the NVIDIA Run:ai login screen
2. Open the Chrome Network inspector: Right-click → Inspect on the page → Network tab
3. On the sign-in page click CONTINUE WITH SSO.
4. Once redirected to the Identity Provider, search in the Chrome network inspector for an HTTP request showing the SAML Request. Depending on the IDP url, this would be a request to the IDP domain name. For example, `accounts.google.com/idp?1234`.
5. When found, go to the Payload tab and copy the value of the SAML Request
6. Paste the value into a SAML decoder (e.g. [https://www.samltool.com/decode.php](https://www.samltool.com/decode.php))
7. Validate the request:
   * The content of the `<saml:Issuer>` tag is the same as `Entity ID` given when [adding the identity provider](saml.md#adding-the-identity-provider)
   * The content of the `AssertionConsumerServiceURL` is the same as the `Redirect URI` given when [adding the identity provider](saml.md#adding-the-identity-provider)
8. Validate the response:
   * The user email under the `<saml2:Subject>` tag is the same as the logged-in user
   * Make sure that under the `<saml2:AttributeStatement>` tag, there is an Attribute named `email` (lowercase). This attribute is mandatory.
   * If other, optional user attributes (`groups`, `firstName`, `lastName`, `uid`, `gid`) are mapped make sure they also exist under `<saml2:AttributeStatement>` along with their respective values.

</Accordion>
