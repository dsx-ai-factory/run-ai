# CLI Commands Examples

This section provides examples of popular use cases illustrating how to use the Command Line Interface (CLI).

## Logging In

### Logging in via NVIDIA Run:ai Sign in Page (Web)

You can log in from the UI, if you are using SSO or credentials

```shell
runai login
```

### Logging in via Terminal (Credentials)

```shell
runai login user -u john@acme.com -p "password"
```

## Configuration

### Setting a Default Project

```shell
runai project set "project-name"
```

## Submitting a Workload

### Naming a Workload

Use the commands below to provide a name for a workload.

#### **Setting the Workload Name ( my_workload_name)**

```shell
runai workspace submit my-workload-name -p test -i ubuntu 
```

#### **Setting a Random Name with Prefix (prefix=workload type)**

```shell
runai workspace submit -p test -i ubuntu 
```

#### **Setting a Random Name with Specific Prefix (prefix determined by flag)**

```shell
runai workspace submit --prefix-name my-prefix-workload-name -p test -i ubuntu 
```

### Labels and Annotations

#### **Labels**

```shell
runai workspace submit -p test -i ubuntu --label name=value --label name2=value2
```

#### **Annotations**

```shell
runai workspace submit -p test -i ubuntu --annotation name=value --annotation name2=value2
```

### Container's Environment Variables

```shell
runai workspace submit -p test -i ubuntu -e name=value -e name2=value2
```

### Requests and Limits

```shell
runai workspace submit -p "project-name" -i nvcr.io/nvidia/pytorch:25.06-py3 --cpu-core-request 0.3 --cpu-core-limit 1 --cpu-memory-request 50M --cpu-memory-limit 1G --gpu-devices-request 1 --gpu-memory-request 1G
```

### Submitting and Attaching to Process

```shell
runai workspace submit  -p "project-name" -i python  --attach -- python3
```

### Submitting a Jupyter Notebook

```shell
runai workspace submit --image jupyter/scipy-notebook -p "project-name" --gpu-devices-request 1 --external-url container=8888 --name-prefix jupyter --command -- start-notebook.sh --NotebookApp.base_url='/${RUNAI_PROJECT}/${RUNAI_JOB_NAME}' --NotebookApp.token=''
```

### Submitting Distributed Training Workload with TensorFlow

```shell
runai tensorflow submit --workers=5 --no-master -g 1 -i kubeflow/tf-mnist-with-summaries:latest -p "project-name" --command -- python /var/tf_mnist/mnist_with_summaries.py --max_steps 1000000
```

### Submitting a Multi-Pod Workload

```shell
runai training submit  -i alpine -p test --parallelism 2 --completions 2  -- sleep 100000
```

### Submit and Bash

#### **Submitting a Workload with Bash Command**

```shell

runai training pytorch submit  -p "project-name" -i nvidia/cuda:11.8.0-cudnn8-runtime-ubuntu20.04 -g 1 --workers 3 --command -- bash -c 'trap : TERM INT; sleep infinity & wait'
```

#### **Bashing Into the Workload**

```shell

runai training pytorch bash pytorch-06027b585626 -p "project-name"
```

### Submitting Distributed Training Workload with MPI

```shell

runai mpi submit dist1 --workers=2 -g 1 \
    -i ubuntu -p "project-name" --command -- sleep 60
```

### Submitting with PVC

#### **New PVC Bounded to the Workspace**

New PVCs will be deleted when the workload is deleted

```shell

runai workspace submit -i ubuntu --new-pvc claimname=yuval-3,size=10M,path=/tmp/test
```

#### **New Ephemeral PVC**

New ephemeral PVCs will be deleted when the workload is deleted or paused

```shell

runai workspace submit -i ubuntu --new-pvc claimname=yuval2,size=10M,path=/tmp/test,ephemeral
```

#### **Existing PVC**

Existing PVCs will not be deleted when the workload is deleted

```shell
runai workspace submit -i ubuntu --existing-pvc claimname=test-pvc-2-project-mn2xs,path=/home/test
```

### Master/Worker Configuration

\--command flag and -- are set both leader (master) and workers command/arguments

\--master-args flag sets the master arguments

\--master-command flag sets the master commands with arguments

\--master-args and --master-command flags can be set together

#### **Overriding Both the Leader (Master) and Worker Image's Arguments**

```shell
runai pytorch submit -i ubuntu -- -a argument_a -b argument_b -c
```

#### **Overriding Both the Leader (Master) and Worker Image's Commands with Arguments**

```shell
runai pytorch submit -i ubuntu --command -- python -m pip install
```

#### **Overriding Arguments of the Leader (Master) and Worker Image's Arguments with Different Values**

```shell
runai pytorch submit -i ubuntu --master-args "-a master_arg_a -b master-arg_b'" -- '-a worker_arg_a'
```

#### **Overriding Command with Arguments of the Leader (Master) and Worker Image's Arguments**

```shell
runai pytorch submit -i ubuntu --master-command "python_master -m pip install'" --command -- 'python_worker -m pip install'
```

### Listing Objects

#### Submitting with Clean Pod Policy

Submitting a distributed workload with a policy to clean:

```
runai mpi submit --clean-pod-policy All -i nvidia/cuda:11.8.0-cudnn8-runtime-ubuntu20.04 -g 1
```

#### Listing All Workloads in the User's Scope

```shell
runai workload list -A
```

#### Listing Projects in a YAML Format

```shell

runai project list --yaml
```

#### Listing Nodes in a JSON Format

```shell
runai node list --json
```

## CLI Reference

For the full guide of the CLI syntax, see the [CLI reference](runai/).
