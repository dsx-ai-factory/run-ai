# Configure Priority and Preemption

Set the priority and preemptibility of a workload when you submit it. For how priority and preemptibility work, the defaults for each workload type, and how administrators change the defaults, refer to [Workload priority and preemption](../../platform-management/runai-scheduler/scheduling/workload-priority-control.md).

## Setting Priority and Preemptibility During Workload Submission

<Note>

**Note**

* If preemptibility is not explicitly configured, the system uses the default preemptibility behavior associated with the selected workload priority.
* Changing a workload’s priority and preemptibility may impact its ability to be scheduled. For example, switching a workload from a low priority, preemptible value (which allows over-quota usage) to high priority, non-preemptible value (which requires in-quota resources) may reduce its chances of being scheduled in cases where the required quota is unavailable.

</Note>

### NVIDIA Run:ai Native Workloads

For native NVIDIA Run:ai workloads, priority and preemptibility can be set during workload submission using one of the following methods:

* **UI** - Set workload priority and preemptibility under **General** settings
* **API** - Set using the `priorityClass` and `preemptibility` field
* **CLI** - Set using the `--priority` and `--preemptibility` flag

### Supported Workload Types

<Note>

**Note**

If priority or preemptibility is set through the UI, API, or CLI, those values override any values defined in the YAML manifest.

</Note>

For [supported workload types](../workload-types/supported-workload-types.md) submitted with a YAML manifest, priority and preemptibility can be set as follows:

* **UI** - Set workload priority and preemptibility under **General** settings
* **API** - Set using the `priority` and `preemptibility` fields
* **CLI** - Set using the `--priority` and `--preemptibility` flags
* **via YAML manifest** - Set by adding the following labels to your YAML manifest under the `metadata.labels` section of your workload definition.
  * Use the following values for priority - `very-low`, `low`, `medium-low`, `medium`, `medium-high`, `high`, `very-high` :
  *   Use the following values for preemptibility - `preemptible`, `semi-preemptible`, or `non-preemptible`

      ```yaml
      metadata:
        labels:
          priorityClassName: <priority>
          kai.scheduler/preemptibility: <preemptibility_value>
      ```
