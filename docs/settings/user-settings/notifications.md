# Notifications

Managing numerous data science workloads requires monitoring various stages, including submission, scheduling, initialization, execution, and completion. Handling suspensions and failures is also critical for ensuring workloads complete on time. Email and slack notifications help address this need by alerting users to important changes in the workload life cycle, allowing timely response and intervention.

Once email and/or slack notifications are configured by the system administrator, a user can set preferences to receive alerts about:

* A workload transitions from one status to another
* A workload is approaching termination due to a project-defined timeout

Each email/slack notification includes:

* Workload type
* Project and cluster information
* Event timestamp

## Prerequisites

* Your organization's email server must be configured. If not, contact your administrator to enable email notifications.
* Your Slack workspace must be connected to NVIDIA Run:ai to enable Slack notifications. If not, contact your administrator.

## Configuring Notifications

To configure the types of email/slack notifications you receive:

1. Click the user avatar at the top right corner, then select **Settings**
2. In the **Notifications** section, under **Send me system notifications via**, select **Email** and/or **Slack**
3. Under **Notify me about my workloads when:**
   * Select the relevant workload statuses you want to monitor
   * Select workload types to receive warnings before timeout - **Workspace workload** or **Training workload**. An email or slack message will be sent when half of the defined timeout period has passed, as specified in the project's [scheduling rules](../../platform-management/policies/scheduling-rules.md).
4. Click **SAVE**
