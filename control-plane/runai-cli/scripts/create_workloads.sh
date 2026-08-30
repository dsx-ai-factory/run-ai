#!/bin/bash

# Define colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

logInfo() {
  echo -e "${GREEN}==> ${NC}$1"
}

logError() {
  echo -e "${RED}==> Error: ${NC}$1" >&2
}

runai login || {
  logError "Failed to login"
  exit 1
}

# Define arrays for configurations
types=("training" "workspace" "pytorch" "tensorflow" "mpi" "xgboost")
gpus=(1 0)
commands=("sleep 10" "sleep infinity" "failed command")
cp_projects=$(runai project list | tail -n +3 | awk '{print $1}' | xargs -I {} echo "\"{}\"" | tr '\n' ' ')
# shellcheck disable=SC2206
projects=($cp_projects)

if [ -z "$projects" ]; then
    logError "No projects found"
    exit 1
fi

# Loop through configurations
echo "Creating general workloads"
for type in "${types[@]}"; do
    for gpu in "${gpus[@]}"; do
        for project in "${projects[@]}"; do

            if [ $gpu -eq 0 ]; then
                gpu_text="with 0 GPU"
                gpu_param="-g 0"
            else
                gpu_text="with 1 GPU"
                gpu_param="-g 1"
            fi

            echo ""
            logInfo "Creating $type workloads $gpu_text on project $project"
            echo "---------------------------------------------"

            for cmd in "${commands[@]}"; do
                runai_cmd="runai $type submit -i ubuntu -p $project $gpu_param --command -- $cmd"
                echo "$runai_cmd"
                eval "$runai_cmd"
                echo ""
            done
        done
    done
done

# Create Jupyter workloads
logInfo "Creating Jupyter workloads"
for project in "${projects[@]}"; do
    runai_cmd="runai workspace submit --image jupyter/scipy-notebook --gpu-devices-request 1 -p $project \
        --external-url container=8888 --name-prefix jupyter --command \
        -- start-notebook.sh --NotebookApp.base_url='\${RUNAI_PROJECT}\${RUNAI_JOB_NAME}' --NotebookApp.token=''"
    echo "$runai_cmd"
    eval "$runai_cmd"
    echo ""
done