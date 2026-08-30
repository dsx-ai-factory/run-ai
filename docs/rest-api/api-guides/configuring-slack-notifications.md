# Configuring Slack Notifications

This section demonstrates how to install the Slack App in your Slack workspace and receive notifications to your Slack channels and direct messages.

## Prerequisites <a href="#prerequisites" id="prerequisites"></a>

* Make sure you have sufficient permissions for your Slack workspace to install the Slack App.
* Make sure you have an NVIDIA Run:ai API token. See [How to authenticate to the API](../how-to-authenticate-to-the-api.md) for more details.

## Install a New Slack App <a href="#install-a-new-slack-app" id="install-a-new-slack-app"></a>

1. Visit [https://api.slack.com/apps](https://api.slack.com/apps) and generate an [App Configuration token](https://api.slack.com/authentication/config-tokens)
2. Copy the access token (should start with `xoxe.xoxp)`
3. Send a POST request using the Create Slack app API and add the access token:

```bash
{
  "accessToken": "THE_GENERATED_ACCESS_TOKEN_FROM_PREVIOUS_STEP"
}
```

4. The request creates your Slack App and returns the newly created Slack App’s URL. For example:

```bash
{
  "slackAppUrl": "https://api.slack.com/apps/A089PHD5S0Q"
}
```

5. Open the URL to view and manage your Slack App configuration.
