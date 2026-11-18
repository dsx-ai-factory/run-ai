# Hotfixes for Version 3.0

This section provides details on all hotfixes available for version 3.0. Hotfixes are critical updates released between our major and minor versions to address specific issues or vulnerabilities. These updates ensure the system remains secure, stable, and optimized without requiring a full version upgrade.

<Note>

**Note**

The versions listed on this page are the versions published in [NGC](https://catalog.ngc.nvidia.com/orgs/nvidia/runai-oss/collections/runai/-). For the full list of changes in every release, see the [changelog](https://github.com/dsx-ai-factory/run-ai/releases).

</Note>

<table><thead><tr><th width="113">Version</th><th width="130">Date (MM/DD/YYYY)</th><th width="100">Issue #</th><th>Description</th></tr></thead><tbody>
<tr><td>3.0.24</td><td>11/15/2026</td><td>#3239</td><td>Fixed an issue where the Prerequisites guide link did not appear in the Missing prerequisites and GPU fractions tooltips.</td></tr>
<tr><td>3.0.24</td><td>11/15/2026</td><td>#3444</td><td>Fixed an issue where distributed workload submission failed with a Bad Gateway error when the policy set defaults for only the worker or only the master.</td></tr>
<tr><td>3.0.24</td><td>11/15/2026</td><td>#3230</td><td>Fixed an issue where custom labels were not applied to the pod templates of Workspace and Training workloads.</td></tr>
<tr><td>3.0.24</td><td>11/15/2026</td><td>#3743</td><td>Fixed an issue where <code>runai training exec</code> closed after about five seconds without output and returned exit code 0, instead of staying connected until the configured idle timeout.</td></tr>
<tr><td>3.0.24</td><td>11/15/2026</td><td>#3753</td><td>Fixed an issue where the workloads helper crashed when pod metrics were missing labels.</td></tr>
<tr><td>3.0.24</td><td>11/15/2026</td><td>#3769</td><td>Fixed an issue where the unused <code>admin-cli</code> authentication client was enabled in tenant realms.</td></tr>
<tr><td>3.0.24</td><td>11/15/2026</td><td>#3221</td><td>Fixed an issue where a workload with a GPU memory request equal to its limit was never admitted and no pod was created.</td></tr>
<tr><td>3.0.24</td><td>11/15/2026</td><td>#4556</td><td>Fixed an issue where requests to the Keycloak Admin API were rejected with a 401 Unauthorized error when the admin client did not have the <code>kc-admin-api</code> scope.</td></tr>
<tr><td>3.0.24</td><td>11/15/2026</td><td>#4372</td><td>Fixed several issues on the overview dashboard, including widget links that returned no rows, fractional-GPU idle workloads that were hidden, and CSV exports that returned cluster-wide data when a node pool was selected.</td></tr>
<tr><td>3.0.24</td><td>11/15/2026</td><td>#4539</td><td>Fixed an issue where filtering workloads by node pool matched only the first node pool of a workload running in several node pools.</td></tr>
<tr><td>3.0.24</td><td>11/15/2026</td><td>#4747</td><td>Fixed an issue where the External (public access) option was not offered for inference workloads on clusters where Knative Serving uses the Gateway API networking layer without Kourier.</td></tr>
<tr><td>3.0.24</td><td>11/15/2026</td><td>#4738</td><td>Fixed an issue where Dynamo workloads got stuck when MNNVL did not detect their GPU resources.</td></tr>
<tr><td>3.0.24</td><td>11/15/2026</td><td>N/A</td><td>Fixed an issue where NIM workloads submitted with autoscaling failed admission on NIM operator 3.1.x.</td></tr>
<tr><td>3.0.24</td><td>11/15/2026</td><td>N/A</td><td>Fixed an issue where GPU fractions reported <code>GPUOperatorVersionUnsupported</code> on clusters with OLM and no NVIDIA GPU Operator installed.</td></tr>
<tr><td>3.0.24</td><td>11/15/2026</td><td>N/A</td><td>Fixed an issue where NUMA node details stayed outdated after the NUMA-aware scheduling setting of a node pool was changed.</td></tr>
<tr><td>3.0.24</td><td>11/15/2026</td><td>N/A</td><td>Fixed an issue where the in-quota and over-quota GPU hours metrics returned incorrect values for a cluster.</td></tr>
<tr><td>3.0.24</td><td>11/15/2026</td><td>N/A</td><td>Fixed an issue where the preempted pods count metric returned fractional values instead of whole numbers.</td></tr>
<tr><td>3.0.24</td><td>11/15/2026</td><td>N/A</td><td>Fixed an issue where the authentication client allowed direct password login.</td></tr>
<tr><td>3.0.24</td><td>11/15/2026</td><td>N/A</td><td>Fixed an issue where a project policy could not override a rule defined by the system policy when a tenant, cluster, or department policy existed between them.</td></tr>
<tr><td>3.0.24</td><td>11/15/2026</td><td>N/A</td><td>Fixed an issue where MPI v2beta1 launcher and worker pods were not associated with the correct workload components.</td></tr>
<tr><td>3.0.24</td><td>11/15/2026</td><td>N/A</td><td>Fixed an issue where the CLI install dialog showed no command for a connected cluster in the Missing prerequisites state.</td></tr>
<tr><td>3.0.24</td><td>11/15/2026</td><td>N/A</td><td>Fixed a security vulnerability related to CVE-2026-111943 with severity HIGH.</td></tr>
<tr><td>3.0.24</td><td>11/15/2026</td><td>N/A</td><td>Fixed an issue where a deleted external workload could reappear after a late sync update.</td></tr>
<tr><td>3.0.24</td><td>11/15/2026</td><td>N/A</td><td>Fixed an issue where applications using S3 sidecars started before the S3 mount was ready.</td></tr>
<tr><td>3.0.24</td><td>11/15/2026</td><td>N/A</td><td>Fixed a security vulnerability related to CVE-2026-111755 with severity HIGH.</td></tr>
<tr><td>3.0.24</td><td>11/15/2026</td><td>N/A</td><td>Fixed a security vulnerability related to CVE-2026-63209 with severity HIGH.</td></tr>
<tr><td>3.0.24</td><td>11/15/2026</td><td>N/A</td><td>Fixed an issue where access rules synchronization to the cluster stopped working for a role that was renamed during the upgrade.</td></tr>
<tr><td>3.0.24</td><td>11/15/2026</td><td>N/A</td><td>Fixed an issue where KServe workloads in Standard (RawDeployment) mode remained in the Creating status.</td></tr>
<tr><td>3.0.24</td><td>11/15/2026</td><td>N/A</td><td>Fixed an issue where the department list did not return the default node pools of each department.</td></tr>
</tbody></table>
