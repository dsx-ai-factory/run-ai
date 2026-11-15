/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { IResourcePermissions } from "@/models/resource.model";
import { ClusterState } from "@/swagger-models/cluster-service-client";
import { DOCS } from "@/constants/docs";

export const CLUSTER_STATUS_TOOLTIP = {
  WaitingToConnect:
    "Once the Run:ai cluster has been installed, it will appear as connected." +
    "<br> If the cluster was already installed and isn't connected," +
    "<br> try the following troubleshooting options:",
  Connected: "The Run:ai cluster is connected, and all Run:ai services are running.",
  Disconnected:
    "There's no communication from the Run:ai cluster." +
    "<br> First make sure the cluster and its network are working." +
    "<br> If so, try the following Run:ai troubleshooting options:",
  MissingPrerequisites:
    "The following prerequisites are missing from the cluster: {{missing_prerequisit}}." +
    "<br>As a result, some features may be impacted.",
  MissingGpuFractioningPrerequisites: "{{reason}}" + "<br>As a result, GPU fractions are not supported.",
  // Emitted once after the sections above, which may both be present.
  PrerequisitesGuide: DOCS.INSTALLATION.CLUSTER_PREREQUISITES
    ? `For more information see the <a href="${DOCS.INSTALLATION.CLUSTER_PREREQUISITES}" style="text-decoration: underline;" target="_blank">Prerequisites guide</a>.`
    : "",
  ServiceIssues:
    "Some Run:ai service(s) in the cluster are not functioning properly. <br>" +
    "As a result, some features may be impacted." +
    `<br> See the list of <a href="javascript:void(0)" style="text-decoration: underline;" id="status-col-emit-services">nonfunctioning services</a> and try the following troubleshooting options:`,
  Unknown: "The cluster status couldn't be determined. If this issue persists, <br>",
};

export enum EClusterVersions {
  v2_9 = "2.9",
}

export enum EClusterDistIds {
  VANILLA = "vanilla",
  OPEN_SHIFT = "openshift",
  EKS = "eks",
  AKS = "aks",
  GKE = "gke",
  RKE = "rke",
}
export enum EClusterDistVals {
  VANILLA = "op",
  OPEN_SHIFT = "openshift",
  EKS = "aws",
  AKS = "aks",
  GKE = "gcp",
  // eslint-disable-next-line @typescript-eslint/no-duplicate-enum-values -- RKE intentionally shares value with VANILLA
  RKE = "op",
}
export enum EClusterDistLabels {
  VANILLA = "Vanilla Kubernetes",
  OPEN_SHIFT = "OpenShift",
  EKS = "EKS",
  AKS = "AKS",
  GKE = "GKE",
  RKE = "RKE",
}

export enum EClusterLocations {
  CONTROL_PLANE = "controlPlane",
  REMOTE = "remote",
}

export enum ENodeTelemetryGroupBy {
  ClusterId = "ClusterId",
  Nodepool = "Nodepool",
  Node = "Node",
}

export enum EClusterErrorMessage {
  NotConnected = "Your cluster is not connected",
}

export const clusterNewInstallerThreshold = "2.15";
export const TEST_ENV_VERSION = "master";
export const CLUSTER_PR_TO_MASTER_VERSION_PREFIX = "0.0.0";
export const TRAINING_DISTRIBUTED_MIN_VERSION = "2.13";

export const adminWarningStatuses: ClusterState[] = [
  ClusterState.Disconnected,
  ClusterState.MissingPrerequisites,
  ClusterState.ServiceIssues,
  ClusterState.WaitingToConnect,
];

export const nonAdminWarningStatuses: ClusterState[] = [
  ClusterState.Disconnected,
  ClusterState.MissingPrerequisites,
  ClusterState.WaitingToConnect,
];

export const clusterK8SDistributionOptions: IK8sDistributionCardData[] = [
  {
    id: EClusterDistIds.VANILLA,
    title: EClusterDistLabels.VANILLA,
    description: "k8s native installation",
    icon: "vanillak8s",
    value: EClusterDistVals.VANILLA,
  },
  {
    id: EClusterDistIds.OPEN_SHIFT,
    title: EClusterDistLabels.OPEN_SHIFT,
    description: "OpenShift Container Platform",
    icon: "openshift",
    value: EClusterDistVals.OPEN_SHIFT,
  },
  {
    id: EClusterDistIds.EKS,
    title: EClusterDistLabels.EKS,
    description: "Amazon Elastic Kubernetes Service",
    icon: "eks",
    value: EClusterDistVals.EKS,
  },
  {
    id: EClusterDistIds.AKS,
    title: EClusterDistLabels.AKS,
    description: "Azure Kubernetes Services",
    icon: "aks",
    value: EClusterDistVals.AKS,
  },
  {
    id: EClusterDistIds.GKE,
    title: EClusterDistLabels.GKE,
    description: "Google Kubernetes Engine",
    icon: "gke",
    value: EClusterDistVals.GKE,
  },
  {
    id: EClusterDistIds.RKE,
    title: EClusterDistLabels.RKE,
    description: "Rancher Kubernetes Engine",
    icon: "rke",
    value: EClusterDistVals.RKE,
  },
];

export interface IK8sDistributionCardData {
  id: string;
  title: string;
  description: string;
  icon: string;
  value: string;
}
export interface IClusterProject {
  id: string;
  name: string;
  managedNamespace: string;
  createdAt: Date;
  deservedGpus?: number;
  clusterUuid: string;
  permissions?: IResourcePermissions;
  interactiveJobTimeLimitSecs?: number;
  interactiveJobMaxIdleDurationSecs?: number;
  interactivePreemptibleJobMaxIdleDurationSecs?: number;
  trainingJobMaxIdleDurationSecs?: number;
  trainNodeAffinity?: string | string[];
  interactiveNodeAffinity?: string | string[];
  departmentName?: string;
}

export interface IPrerequisites {
  location: string;
  clusterUrl: string;
  k8sDistribution: string;
}
export interface IClusterInstalltionModel {
  name: string;
  version?: string;
  prerequisites: IPrerequisites;
}

export const REMOTE_CLUSTER_LOCATION = "Remote control plane";
export const SAME_CLUSTER_LOCATION = "Same as control plane";
export const clusterLocationText = (isRemote: boolean) =>
  `Cluster location: ${isRemote ? REMOTE_CLUSTER_LOCATION : SAME_CLUSTER_LOCATION}`;
