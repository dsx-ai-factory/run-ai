#! /bin/bash

# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0


# GuyContinue: Move the cluster CRD to the runai-chart/crds directory and the runai-crd-upgrader directory

crdUpgraderPath="../runai-crd-upgrader/crds"
runaiChartCRDsSubchartPath="../runai-chart/charts/crds/crds"
runaiChartCRDsPath="../runai-chart/crds"

crdsNeededForChartRenderingDestinations=(
    $crdUpgraderPath
    $runaiChartCRDsPath
)

crdsNotNeededForChartRenderingDestinations=(
    $crdUpgraderPath
    $runaiChartCRDsSubchartPath
)

for destination in "${crdsNotNeededForChartRenderingDestinations[@]}"; do
    if [ ! -d "$destination" ]; then
        echo "Destination directory $destination does not exist"
        exit 1
    fi

    cp config/crd/bases/run.ai_runaiconfigs.yaml "$destination"/run_ai_v1_runaiconfig.yaml
    cp config/crd/bases/run.ai_remoteconfigs.yaml "$destination"/run.ai_remoteconfigs.yaml
    cp config/crd/bases/run.ai_remoteconfigsyncs.yaml "$destination"/run.ai_remoteconfigsyncs.yaml
    cp config/crd/bases/run.ai_managednodesconfigs.yaml "$destination"/run.ai_managednodesconfigs.yaml

    # The kai.resources CRDs are shipped by the KRM chart, not from here.
    cp config/crd/bases/run.ai_nodepools.yaml "$destination"/run_ai_nodepool.yaml
    cp config/crd/bases/run.ai_applications.yaml "$destination"/run_ai_application.yaml
    cp config/crd/bases/run.ai_accessrules.yaml "$destination"/run_ai_accessrule.yaml
    cp config/crd/bases/run.ai_externalworkloads.yaml "$destination"/run_ai_externalworkload.yaml
    cp config/crd/bases/run.ai_datavolumes.yaml "$destination"/run.ai_datavolumes.yaml
    # WorkloadRunner CRD moved to dromos - see dromos/Makefile copy-crds target.
    cp config/crd/bases/run.ai_restartrequests.yaml "$destination"/run.ai_restartrequest.yaml
done

for destination in "${crdsNeededForChartRenderingDestinations[@]}"; do
    if [ ! -d "$destination" ]; then
        echo "Destination directory $destination does not exist"
        exit 1
    fi

    cp config/crd/bases/run.ai_clusters.yaml "$destination"/run.ai_clusters.yaml
done
