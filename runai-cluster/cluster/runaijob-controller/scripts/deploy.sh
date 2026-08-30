#!/usr/bin/env bash
set -x
IMAGE_TAG=${IMAGE_TAG:-$USER}

DEBUG=
while getopts d flag
do
    case "${flag}" in
        d) DEBUG=1;;
    esac
done

make push DEBUG=$DEBUG IMAGE_TAG=$IMAGE_TAG
kubectl rollout restart deployment/runai-job-controller -n runai

#patch deployment
# (old. patching deployment. doesn't work since operator overrides. leaving here for reference)
#option 1 - one liner - (update IMAGE_NAME)d
# kubectl patch deployment runai-admission-controller -n runai -p $'spec:\n template:\n  spec:\n   containers:\n   - name: runai-admission-controller\n     image: runai/runai-admission-controller-staging:IMAGE_NAME'
#option 2 - using patch-deployment.yaml (update IMAGE_NAME in patch-deployment.yaml)
# kubectl patch deployment runai-admission-controller -n runai -p "$(cat patch-deployment.yaml)"

#patch runai crd
#option 1 - one liner - (update IMAGE_NAME)
# kubectl patch runaiconfig runai -n runai -p $'spec:\n admission-controller:\n  image:\n   tag: IMAGE_NAME\n   env: lab' --type=merge
#option 2 - using patch-runai-crd.yaml (update IMAGE_NAME in patch-runai-crd.yaml)
# kubectl patch runai runai -n runai -p "$(cat patch-runai-crd.yaml)" --type=merge
