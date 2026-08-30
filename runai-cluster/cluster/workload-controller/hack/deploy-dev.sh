#!/bin/bash
dockerTagName=${IMAGE_NAME:-"test"} # You can also put it in the git ignored .env
dockerRepo="gcr.io/run-ai-lab/workload-controller"
DEBUG=

if [ -f .env ]; then
  export $(cat .env | xargs)
fi

IMG=${dockerRepo}:${dockerTagName} make docker-dev &&
docker push ${dockerRepo}:${dockerTagName} &&
kubectl rollout restart deployment/workload-controller -n runai
