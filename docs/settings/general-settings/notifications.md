# Notifications

The NVIDIA Run:ai platform provides multiple notification options to keep administrators and users informed of important events and updates. Administrators can enable any combination of the following methods:

* **Email notifications** - Delivered through a configured SMTP server
* **Slack notifications** - Integrated with your Slack workspace for real-time notifications
* **System notifications** - Displayed directly in the NVIDIA Run:ai interface for all users

## Email Notifications

You can configure an SMTP server to send email notifications from the system. This allow users to receive email notifications and set their own preferences via the [User settings](../user-settings/notifications.md).

1. Go to **General settings** and navigate to Notifications -> Email notifications
2. Click **+EMAIL SERVER**
3. Enter SMTP server details
   * **SMTP host** - Enter your mail server hostname (e.g., `smtp.mail.com`)
   * **Port** - Enter the port number
4. Select **Authentication type**
   * **Plain** - Use a plain username/password authentication
   * **Login** - Use login authentication
   * **None** - No authentication required (not recommended)
5. Provide authentication details (if required)
   * **Username** - Enter the account username
   * **Password** - Enter the account password
6. Define email information
   * **From email address** - Enter the address to appear in the “From” field of notifications (e.g., `runai@company.com`)
   * **From display name** - Enter the sender display name (e.g., `Run:ai Notification`)
7. (Optional) Set a recipient email address to verify connectivity
   * **Email address** - Enter a recipient address to confirm the SMTP server accepts mail for it. No message is sent to this address; it is used only for verification.
8. Click **VERIFY**
   * The system checks connectivity and credentials, and, if a recipient address was provided, confirms the server accepts that recipient
   * If verification fails, correct the values and retry
9. Click **SAVE** to apply

## Slack Notifications

You can connect your Slack workspace to NVIDIA Run:ai to receive real-time notifications. To allow users to receive Slack notifications and set their own preferences via the user settings, first connect your Slack workspace.

1. Go to **General settings** and navigate to Notifications -> Slack notifications
2. Enter your Slack **Access token** in the token field. To get an access token:
   * Visit [https://api.slack.com/apps](https://api.slack.com/apps) and generate an [App Configuration token](https://docs.slack.dev/app-manifests/configuring-apps-with-app-manifests/#config-tokens)
   * Copy the access token (it should start with `xoxe.xoxp`)
3. Click **CONNECT SLACK WORKSPACE**. A new browser window will open, prompting you to sign in to Slack.
4. Select the Slack workspace you want to connect
5. Slack will request permission to connect to your workspace. Authorize NVIDIA Run:ai to post messages on behalf of your workspace.

Once authorization is complete, the UI will show the connected workspace name and a **Status: Connected** indicator. You can also see a **DISCONNECT** button if you need to remove.

## System Notifications

System notifications allow administrators to broadcast custom messages that are displayed to all users. Notifications appear immediately to currently signed-in users and upon login for new sessions.

1. Go to **General settings** and navigate to Notifications -> System notification
2. Click **+ MESSAGE**
3. In the text editor, type the message to display
4. (Optional) Select the checkbox **Display "Don't show this again"** to allow users to dismiss the notification permanently
5. Click **SAVE & PUBLISH**

## Using APIs

To view the available actions, see the [Notification Channels](api:main-api:GET/api/v1/notification-channels) API reference.
