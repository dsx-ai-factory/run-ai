/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package itemized

import (
	"strings"

	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/datastructs"
	"github.com/run-ai/runai/runai-application-packages/jsonutils"
	policyApi "github.com/run-ai/runai/runai-clients/policy-service-client/generated"
)

const (
	InstancesSuffix  = ".instances"
	AttributesSuffix = ".attributes"

	PolicyInstances  = "instances"
	PolicyAttributes = "attributes"

	ServingPortPrefix          = "serving-port"
	TolerationPrefix           = "toleration"
	PortPrefix                 = "port"
	ExposedUrlPrefix           = "url"
	ExtendedResourceNickPrefix = "ext-resource"
	HostPathPrefix             = "host-path"
	PvcPrefix                  = "pvc"
	DataVolumePrefix           = "data-volume"
	NfsPrefix                  = "nfs"
	GitPrefix                  = "git"
	S3Prefix                   = "s3"
	ConfigMapPrefix            = "config-map"
	SecretStoragePrefix        = "secret"
	EmptyDirPrefix             = "empty-dir"
)

var distributedJsonCracker = jsonutils.NewJsonCracker(policyApi.DistributedPolicyDefaultsV2{})
var workerPrefix = distributedJsonCracker.NameOf(&distributedJsonCracker.Struct().Worker) + "."
var masterPrefix = distributedJsonCracker.NameOf(&distributedJsonCracker.Struct().Master) + "."

var policyJsonCracker = jsonutils.NewJsonCracker(policyApi.SupersetRules{})
var computeSection = policyJsonCracker.NameOf(&policyJsonCracker.Struct().Compute)
var storageSection = policyJsonCracker.NameOf(&policyJsonCracker.Struct().Storage)

var storageJsonCracker = jsonutils.NewJsonCracker(policyApi.SupersetDefaultsAllOfStorage{})
var computeJsonCracker = jsonutils.NewJsonCracker(policyApi.SupersetDefaultsAllOfCompute{})
var distributedInferenceJsonCracker = jsonutils.NewJsonCracker(policyApi.DistributedInferenceDefaultsV2{})
var leaderPrefix = distributedInferenceJsonCracker.NameOf(&distributedInferenceJsonCracker.Struct().Leader) + "."
var workerInferencePrefix = distributedInferenceJsonCracker.NameOf(&distributedInferenceJsonCracker.Struct().Worker) + "."

// JSON crackers for distributed inference leader/worker structs
var distributedInferenceLeaderWorkerJsonCracker = jsonutils.NewJsonCracker(policyApi.DistributedInferenceLeaderWorkerDefaultsV2{})
var distributedInferenceStorageJsonCracker = jsonutils.NewJsonCracker(policyApi.DistributedInferenceLeaderWorkerDefaultsV2Storage{})

// keep a set of all the sections of policy payload that has inner fields
var NestedPolicySections = datastructs.NewSet(
	policyJsonCracker.NameOf(&policyJsonCracker.Struct().Security),
	policyJsonCracker.NameOf(&policyJsonCracker.Struct().Compute),
	policyJsonCracker.NameOf(&policyJsonCracker.Struct().Storage),
	policyJsonCracker.NameOf(&policyJsonCracker.Struct().Autoscaling),
	distributedInferenceJsonCracker.NameOf(&distributedInferenceJsonCracker.Struct().Leader),
	distributedInferenceJsonCracker.NameOf(&distributedInferenceJsonCracker.Struct().Worker),
	// Add nested sections for distributed inference leader/worker
	convutils.DotUs(leaderPrefix, distributedInferenceLeaderWorkerJsonCracker.NameOf(&distributedInferenceLeaderWorkerJsonCracker.Struct().Storage)),
	convutils.DotUs(workerInferencePrefix, distributedInferenceLeaderWorkerJsonCracker.NameOf(&distributedInferenceLeaderWorkerJsonCracker.Struct().Storage)),
)

type ItemizedFieldsInfo struct {
	KeyAttribute string

	// key is always mandatory in policy, but for datasource, for example, it is just a nickname
	// without any impact on cluster, just used for locked rules, so it can be omitted.
	KeyIsOptional bool

	// the nickname prefix
	NicknamePrefix string
}

func (ifi *ItemizedFieldsInfo) WithPrefix(prefix string) *ItemizedFieldsInfo {
	ifi.NicknamePrefix = prefix
	return ifi
}

// name is the key attribute of most variables
var keyAttributeName = ItemizedFieldsInfo{KeyAttribute: "name"}
var optionalName = ItemizedFieldsInfo{KeyAttribute: "name", KeyIsOptional: true}
var optionalId = ItemizedFieldsInfo{KeyAttribute: "id", KeyIsOptional: false}

// resource is the key attribute of extendedResources
var extendedResourcesCracker = jsonutils.NewJsonCracker(policyApi.ExtendedResource{})
var keyAttributeResource = ItemizedFieldsInfo{KeyAttribute: extendedResourcesCracker.NameOf(&extendedResourcesCracker.Struct().Resource)}

// for itemized fields, we need to know the attribute which serves as the key of the item.
var itemizedFields = map[string]ItemizedFieldsInfo{
	policyJsonCracker.NameOf(&policyJsonCracker.Struct().EnvironmentVariables): keyAttributeName,
	policyJsonCracker.NameOf(&policyJsonCracker.Struct().Annotations):          keyAttributeName,
	policyJsonCracker.NameOf(&policyJsonCracker.Struct().Labels):               keyAttributeName,
	policyJsonCracker.NameOf(&policyJsonCracker.Struct().ImagePullSecrets):     keyAttributeName,

	policyJsonCracker.NameOf(&policyJsonCracker.Struct().Tolerations): *optionalName.WithPrefix(TolerationPrefix),
	policyJsonCracker.NameOf(&policyJsonCracker.Struct().Ports):       *optionalName.WithPrefix(PortPrefix),
	policyJsonCracker.NameOf(&policyJsonCracker.Struct().ExposedUrls): *optionalName.WithPrefix(ExposedUrlPrefix),
	policyJsonCracker.NameOf(&policyJsonCracker.Struct().RelatedUrls): optionalName, // no prefix because it is not generated to the cluster

	convutils.DotUs(computeSection, computeJsonCracker.NameOf(&computeJsonCracker.Struct().ExtendedResources)): *keyAttributeResource.WithPrefix(ExtendedResourceNickPrefix),

	convutils.DotUs(storageSection, storageJsonCracker.NameOf(&storageJsonCracker.Struct().HostPath)):        *optionalName.WithPrefix(HostPathPrefix),
	convutils.DotUs(storageSection, storageJsonCracker.NameOf(&storageJsonCracker.Struct().Nfs)):             *optionalName.WithPrefix(NfsPrefix),
	convutils.DotUs(storageSection, storageJsonCracker.NameOf(&storageJsonCracker.Struct().Pvc)):             *optionalName.WithPrefix(PvcPrefix),
	convutils.DotUs(storageSection, storageJsonCracker.NameOf(&storageJsonCracker.Struct().Git)):             *optionalName.WithPrefix(GitPrefix),
	convutils.DotUs(storageSection, storageJsonCracker.NameOf(&storageJsonCracker.Struct().S3)):              *optionalName.WithPrefix(S3Prefix),
	convutils.DotUs(storageSection, storageJsonCracker.NameOf(&storageJsonCracker.Struct().ConfigMapVolume)): *optionalName.WithPrefix(ConfigMapPrefix),
	convutils.DotUs(storageSection, storageJsonCracker.NameOf(&storageJsonCracker.Struct().SecretVolume)):    *optionalName.WithPrefix(SecretStoragePrefix),
	convutils.DotUs(storageSection, storageJsonCracker.NameOf(&storageJsonCracker.Struct().DataVolume)):      *optionalId.WithPrefix(DataVolumePrefix),
	convutils.DotUs(storageSection, storageJsonCracker.NameOf(&storageJsonCracker.Struct().EmptyDirVolume)):  *optionalName.WithPrefix(EmptyDirPrefix),

	// Distributed inference leader storage fields
	convutils.DotUs(leaderPrefix, convutils.DotUs(distributedInferenceLeaderWorkerJsonCracker.NameOf(&distributedInferenceLeaderWorkerJsonCracker.Struct().Storage),
		distributedInferenceStorageJsonCracker.NameOf(&distributedInferenceStorageJsonCracker.Struct().Pvc))): *optionalName.WithPrefix(PvcPrefix),
	convutils.DotUs(leaderPrefix, convutils.DotUs(distributedInferenceLeaderWorkerJsonCracker.NameOf(&distributedInferenceLeaderWorkerJsonCracker.Struct().Storage),
		distributedInferenceStorageJsonCracker.NameOf(&distributedInferenceStorageJsonCracker.Struct().ConfigMapVolume))): *optionalName.WithPrefix(ConfigMapPrefix),
	convutils.DotUs(leaderPrefix, convutils.DotUs(distributedInferenceLeaderWorkerJsonCracker.NameOf(&distributedInferenceLeaderWorkerJsonCracker.Struct().Storage),
		distributedInferenceStorageJsonCracker.NameOf(&distributedInferenceStorageJsonCracker.Struct().SecretVolume))): *optionalName.WithPrefix(SecretStoragePrefix),
	convutils.DotUs(leaderPrefix, convutils.DotUs(distributedInferenceLeaderWorkerJsonCracker.NameOf(&distributedInferenceLeaderWorkerJsonCracker.Struct().Storage),
		distributedInferenceStorageJsonCracker.NameOf(&distributedInferenceStorageJsonCracker.Struct().EmptyDirVolume))): *optionalName.WithPrefix(EmptyDirPrefix),

	// Distributed inference worker storage fields
	convutils.DotUs(workerInferencePrefix, convutils.DotUs(distributedInferenceLeaderWorkerJsonCracker.NameOf(&distributedInferenceLeaderWorkerJsonCracker.Struct().Storage),
		distributedInferenceStorageJsonCracker.NameOf(&distributedInferenceStorageJsonCracker.Struct().Pvc))): *optionalName.WithPrefix(PvcPrefix),
	convutils.DotUs(workerInferencePrefix, convutils.DotUs(distributedInferenceLeaderWorkerJsonCracker.NameOf(&distributedInferenceLeaderWorkerJsonCracker.Struct().Storage),
		distributedInferenceStorageJsonCracker.NameOf(&distributedInferenceStorageJsonCracker.Struct().ConfigMapVolume))): *optionalName.WithPrefix(ConfigMapPrefix),
	convutils.DotUs(workerInferencePrefix, convutils.DotUs(distributedInferenceLeaderWorkerJsonCracker.NameOf(&distributedInferenceLeaderWorkerJsonCracker.Struct().Storage),
		distributedInferenceStorageJsonCracker.NameOf(&distributedInferenceStorageJsonCracker.Struct().SecretVolume))): *optionalName.WithPrefix(SecretStoragePrefix),
	convutils.DotUs(workerInferencePrefix, convutils.DotUs(distributedInferenceLeaderWorkerJsonCracker.NameOf(&distributedInferenceLeaderWorkerJsonCracker.Struct().Storage),
		distributedInferenceStorageJsonCracker.NameOf(&distributedInferenceStorageJsonCracker.Struct().EmptyDirVolume))): *optionalName.WithPrefix(EmptyDirPrefix),

	// Distributed inference leader itemized fields
	convutils.DotUs(leaderPrefix, distributedInferenceLeaderWorkerJsonCracker.NameOf(&distributedInferenceLeaderWorkerJsonCracker.Struct().EnvironmentVariables)): keyAttributeName,
	convutils.DotUs(leaderPrefix, distributedInferenceLeaderWorkerJsonCracker.NameOf(&distributedInferenceLeaderWorkerJsonCracker.Struct().Labels)):               keyAttributeName,
	convutils.DotUs(leaderPrefix, distributedInferenceLeaderWorkerJsonCracker.NameOf(&distributedInferenceLeaderWorkerJsonCracker.Struct().Annotations)):          keyAttributeName,

	// Distributed inference worker itemized fields
	convutils.DotUs(workerInferencePrefix, distributedInferenceLeaderWorkerJsonCracker.NameOf(&distributedInferenceLeaderWorkerJsonCracker.Struct().EnvironmentVariables)): keyAttributeName,
	convutils.DotUs(workerInferencePrefix, distributedInferenceLeaderWorkerJsonCracker.NameOf(&distributedInferenceLeaderWorkerJsonCracker.Struct().Labels)):               keyAttributeName,
	convutils.DotUs(workerInferencePrefix, distributedInferenceLeaderWorkerJsonCracker.NameOf(&distributedInferenceLeaderWorkerJsonCracker.Struct().Annotations)):          keyAttributeName,
}

func GetItemizedFieldInfo(fieldName string) (ItemizedFieldsInfo, bool) {
	if strings.HasPrefix(fieldName, workerPrefix) {
		fieldName = fieldName[len(workerPrefix):]
	} else if strings.HasPrefix(fieldName, masterPrefix) {
		fieldName = fieldName[len(masterPrefix):]
	} else if strings.HasPrefix(fieldName, leaderPrefix) {
		fieldName = fieldName[len(leaderPrefix):]
	} else if strings.HasPrefix(fieldName, workerInferencePrefix) {
		fieldName = fieldName[len(workerInferencePrefix):]
	}

	result, exist := itemizedFields[fieldName]
	return result, exist
}
