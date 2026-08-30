# NVIDIA Run:ai Native Workload Policies

This guide explains the structure and procedures for managing policies for [NVIDIA Run:ai native workloads](../../workloads-in-nvidia-run-ai/workload-types/native-workloads.md).

<Info>

**Note**

* Workload policies are enabled by default. If you do not see it in the menu, contact your administrator to enable it under **General settings** → Workloads → Policies
* Starting from version 2.23, the creation and editing of control plane policies are no longer synchronized to the cluster. This ensures that new policy features function correctly and prevents conflicts with outdated cluster policies.
  * For cluster versions 2.21 and above, changes to control plane policies do not affect existing cluster policies, because workload policies are stored and enforced only in the control plane. Existing cluster policies impact only workloads submitted directly to the cluster, and administrators may choose to keep or manually delete these older cluster policies.
  * For cluster versions 2.18 to 2.20, policies in the cluster affect both control plane and cluster-submitted workloads. When a policy is created or edited in the control plane, the corresponding or effective policy in the cluster is automatically removed to avoid conflicts. Administrator acknowledgment is required before this change takes effect.
  * To enforce policies directly on the cluster in order to prevent direct workload submission, or for more information, please contact [NVIDIA Run:ai Support](https://www.nvidia.com/en-eu/support/enterprise/#contact-us).

</Info>

## Workload Policies Table

The Workload policies table can be found under **Policies** in the NVIDIA Run:ai platform.

The Workload policies table provides a list of all the policies defined in the platform, and allows you to manage them.

<figure><img src="../../assets/policies-new.png" alt=""/><figcaption></figcaption></figure>

The Workload policies table consists of the following columns:

| Column               | Description                                                                                                                                                                                      |
| -------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Policy               | The policy name which is a combination of the policy scope and the policy type                                                                                                                   |
| Type                 | The policy type is per NVIDIA Run:ai workload type. This allows administrators to set different policies for each [workload type](../../workloads-in-nvidia-run-ai/workload-types/supported-workload-types.md).           |
| Policy architecture  | Indicates whether the policy targets a **Standard** workload (Training, Workspace, or Inference) or a **Distributed** workload.                                                                  |
| Status               | The current state of the policy. Indicates whether the policy is Ready to enforce workloads.                                                                                                     |
| Scope         | The scope the policy affects. Click the name of the scope to view the organizational tree diagram. You can only view the parts of the organizational tree for which you have permission to view. |
| Created by    | The user who created the policy                                                                                                                                                                  |
| Creation time | The timestamp for when the policy was created                                                                                                                                                    |
| Last updated  | The last time the policy was updated                                                                                                                                                             |

### Customizing the Table View

* Filter - Click ADD FILTER, select the column to filter by, and enter the filter values
* Search - Click SEARCH and type the value to search by
* Sort - Click each column header to sort by
* Column selection - Click COLUMNS and select the columns to display in the table
* Refresh - Click REFRESH to update the table with the latest data

## Policy Structure

A workload policy consists of three constructs:

* **Rules** - Constraints that limit or control the values of workload fields. Rules are enforced during submission and cannot be overridden by users.
* **Defaults** - Suggested values for workload fields. Defaults are pre-filled during submission but can be modified by users.
* **Imposed assets** - Assets (such as a PVC data source) that are automatically applied to any workload submitted under the policy's scope.

For more information, see [rules](policy-yaml-reference.md#rules), [defaults](policy-yaml-reference.md#defaults), and [imposed assets](policies-and-rules.md).

## Adding a Policy

To create a new policy:

1. Click **+NEW POLICY**
2. Select a **scope**
3. Select the **workload type**
4. Click **+POLICY YAML**
5. In the **YAML editor**, type or paste a YAML policy with defaults and rules.\
   You can use the following references and examples:
   * [Policy YAML reference](policy-yaml-reference.md)
   * [Policy YAML examples](policy-yaml-examples.md)
6. Click **SAVE POLICY**

## Editing a Policy

1. Select the policy you want to edit
2. Click **EDIT**
3. Update the policy and click **APPLY**
4. Click **SAVE POLICY**

## Troubleshooting

Listed below are issues that might occur when creating or editing a policy via the YAML Editor:

| Issue                                                                                                              | Message                                                                                                                              | Mitigation                                                                                                                                                     |
| ------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Cluster connectivity issues                                                                                        | There's no communication from cluster “cluster\_name“. Actions may be affected, and the data may be stale.                           | Verify that you are on a network that has been allowed access to the cluster. Reach out to your cluster administrator for instructions on verifying the issue. |
| Policy can’t be applied due to a rule that is occupied by a different policy                                       | Field “field\_name” already has rules in cluster: “cluster\_id”                                                                      | Remove the rule from the new policy or adjust the old policy for the specific rule.                                                                            |
| Policy is not visible in the UI                                                                                    | -                                                                                                                                    | Check that the policy hasn’t been deleted.                                                                                                                     |
| Policy syntax is no valid                                                                                          | Add a valid policy YAML;json: unknown field "field\_name"                                                                            | For correct syntax check the [Policy YAML reference](policy-yaml-reference.md) or the [Policy YAML examples](policy-yaml-examples.md).                         |
| Policy can’t be saved for some reason                                                                              | The policy couldn't be saved due to a network or other unknown issue. Download your draft and try pasting and saving it again later. | Possible cluster connectivity issues. Try updating the policy once again at a different time.                                                                  |

## Viewing a Policy

To view a policy:

1. Select the policy for which you want to view its [policies](policies-and-rules.md).
2. Click **VIEW POLICY**
3. In the Policy form per workload section, view the workload rules and defaults:
   * **Parameter**\
     The workload submission parameter that Rules and Defaults are applied to
   * **Type** (applicable for data sources only)\
     The data source type (Git, S3, nfs, pvc etc.)
   * **Default**\
     The default value of the Parameter
   * **Rule**\
     Set up constraint on workload policy field
   * **Source**\
     The origin of the applied policy (cluster, department or project)

<Info>

**Note**

Some of the rules and defaults may be derived from policies of a parent cluster and/or department. You can see the source of each rule in the policy form. For more information, check the [Scope of effectiveness documentation](policies-and-rules.md#scope-of-effectiveness).

</Info>

## Deleting a Policy

1. Select the policy you want to delete
2. Click **DELETE**
3. On the dialog, click **DELETE** to confirm the deletion

## Using API

Go to the [Policies](https://run-ai-docs.nvidia.com/api/2.26/policies/policy) API reference to view the available actions.
