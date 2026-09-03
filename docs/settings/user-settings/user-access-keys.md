# User Access Keys (formerly applications)

Access keys are used for API integrations with NVIDIA Run:ai. An access key contains a client ID and a client secret. With the client credentials, you can obtain a token as detailed in [API authentication](https://run-ai-docs.nvidia.com/api/2.26/getting-started/how-to-authenticate-to-the-api) and use it within subsequent API calls.

<Info>

**Note**

* All clusters in the tenant must be version 2.20 and onward.
* The token obtained through access keys assumes the [roles and permissions](../../platform-management/authentication/roles.md) of the user.

</Info>

## Creating Access Key

To create an access key:

1. Click the user avatar at the top right corner, then select **Settings**
2. Click **+ACCESS KEY**
3. Enter the access key's **name**
4. Click **CREATE**
5. Copy the **Client ID** and **Client secret** and store securely
6. Click **DONE**

You can create up to 20 access keys.

<Info>

**Note**

The client secret is visible only at the time of creation. It cannot be recovered but can be regenerated.

</Info>

## Regenerating a Client Secret

To regenerate a client secret:

1. Locate the access key you want to regenerate its client secret
2. Click **Regenerate client secret**
3. Click **REGENERATE**
4. Copy the **New client secret** and store it securely
5. Click **DONE**

<Warning>

**Important**

Regenerating a client secret revokes the previous one.

</Warning>

## Deleting an Access Key

1. Locate the access key you want to delete
2. Click on the trash icon
3. On the dialog, click **DELETE** to confirm

## Using API

Go to the [Access Keys](https://run-ai-docs.nvidia.com/api/2.26/authentication-and-authorization/access-keys) API reference to view the available actions.
