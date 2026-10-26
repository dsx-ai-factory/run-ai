# Best Practices: Checkpointing Preemptible Training Workloads

NVIDIA Run:ai allows you to define whether a workload is [preemptible](../../platform-management/runai-scheduler/scheduling/workload-priority-control.md), meaning the [NVIDIA Run:ai Scheduler](../../platform-management/runai-scheduler/scheduling/concepts-and-principles.md) may pause a running workload and temporarily reassign its GPU resources to higher priority workloads. When resources become available, NVIDIA Run:ai automatically resumes the preempted workload.

While any workload can be preemptible, checkpointing is primarily relevant for training workloads that run for long durations and need to maintain progress between interruptions. To prevent data loss and ensure continuity, it's a best practice to periodically save checkpoints and configure your workload to resume from the latest checkpoint, typically at the end of each epoch.

## Where to Save Checkpoints

Always use **shared network storage** (e.g., NFS). When a preempted workload is resumed, it may be scheduled on a different node than before. Saving checkpoints to local disk risks data loss. You can mount a preconfigured shared [data source](../workload-assets/overview.md) or specify one during [workload submission](../workloads.md).

Example using CLI:

```sh
runai tensorflow submit train-with-checkpoints -i 
tensorflow/tensorflow:1.14.0-gpu-py3 --host-path 
mount=/mnt/nfs_share/john, path=/mydir -g 1 --working-dir /mydir 
--command -- ./startup.sh
```

The command saves the checkpoints in an NFS checkpoints folder `/mnt/nfs_share/john`.

## When to Save Checkpoints

### Save Periodically

The most common strategy is to save checkpoints at regular intervals, such as at the end of each epoch. For example:

```python
checkpoints_file = "weights.best.hdf5"
checkpoint = ModelCheckpoint(checkpoints_file, monitor='val_acc', verbose=1, 
    save_best_only=True, mode='max')
```

### Save on Exit Signal

If periodic checkpoints are not enough, you can use a signal-hook provided by NVIDIA Run:ai (via Kubernetes). The hook is Python code that is called before your workload is suspended and allows you to save your checkpoints as well as other state data you may wish to store. By default, you will have 30 seconds to save your checkpoints. You can configure this time window to be up to 5 minutes:

```python
import signal
import time

def graceful_exit_handler(signum, frame):
    # save your checkpoints to shared storage

    # exit with status "1" is important for the Job to return later.  
    exit(1)

signal.signal(signal.SIGTERM, graceful_exit_handler)
```

<Note>

**Note**

For the signal to be captured, it must be propagated from the startup script to the Python child process. See code [here](https://github.com/run-ai/docs/blob/master/quickstart/unattended-execution/startup.sh).

</Note>

## Grace Period for Preemption

NVIDIA Run:ai includes a grace period mechanism for [standard](train-models.md) and [distributed training](distributed-training-models.md) workloads. This configurable delay allows workloads time to finish a checkpoint before being forcibly stopped. Use grace period together with signal hooks to reduce the risk of data loss.

## Resuming with Saved Checkpoints

At NVIDIA Run:ai a workload that is resumed will run the same startup script as on the first run. It is the responsibility of the script developer to add code that:

* Checks if saved checkpoints exist (see above)
* If saved checkpoints exist, load them and start the run using these checkpoints

```python
import os

checkpoints_file = "weights.best.hdf5"
if os.path.isfile(checkpoints_file):
    print("loading checkpoint file: " + checkpoints_file)
    model.load_weights(checkpoints_file)
```

## Sample Code

Most ML frameworks, including [TensorFlow](https://www.tensorflow.org/guide/checkpoint) and [PyTorch](https://docs.pytorch.org/tutorials/beginner/saving_loading_models.html), offer built-in checkpointing mechanisms. The [sample code](https://github.com/run-ai/docs/tree/master/quickstart/unattended-execution) provided in the accompanying GitHub repository uses Keras to demonstrate how to implement checkpointing.
