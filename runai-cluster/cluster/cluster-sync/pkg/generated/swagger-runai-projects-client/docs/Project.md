# Project

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | Project Name | [optional] [default to null]
**DepartmentId** | **int32** | ID of the department owning the project | [optional] [default to null]
**DeservedGpus** | **float64** | The project GPU allocation that it is deserved to get in case the cluster has those resources | [optional] [default to null]
**MaxAllowedGpus** | **float64** | An upper limit for the amount of GPUs the project can get (Even if over quota is allowed and resources are available) | [optional] [default to null]
**GpuOverQuotaWeight** | **float64** | The priority the project gets for over quota resources | [optional] [default to null]
**SwapEnabled** | **bool** | Flag for enabling thin provisioning for jobs of this project | [optional] [default to null]
**DefaultNodePools** | **[]string** | Default node pools for workload submission under this project (If a workload doesnt specify node pools list, this list will be used). | [optional] [default to null]
**InteractiveJobTimeLimitSecs** | **float64** | A limit for the duration of interactive jobs from this project in seconds | [optional] [default to null]
**InteractiveJobMaxIdleDurationSecs** | **float64** | Maximum time duration in seconds that interactive job can be idle before it will be terminated | [optional] [default to null]
**InteractivePreemptibleJobMaxIdleDurationSecs** | **float64** | Maximum time duration in seconds that interactive preemptible job can be idle before it will be terminated | [optional] [default to null]
**TrainingJobTimeLimitSecs** | **float64** | A limit for the duration of training jobs from this project in seconds. Available only from cluster version 2.12 | [optional] [default to null]
**TrainingJobMaxIdleDurationSecs** | **float64** | Maximum time duration in seconds that training job can be idle before it will be terminated | [optional] [default to null]
**NodeAffinity** | [***JobsNodeAffinity**](JobsNodeAffinity.md) |  | [optional] [default to null]
**Permissions** | [***ResourcePermissions**](ResourcePermissions.md) |  | [optional] [default to null]
**Resources** | [***AssignedResources**](AssignedResources.md) |  | [optional] [default to null]
**NodePoolsResources** | [**[]NodePoolAssignedResources**](NodePoolAssignedResources.md) | Resources assigned to this Project per Node Pool | [optional] [default to null]
**Namespace** | **string** | The name of an existing namespace to use for the project in the cluster. | [optional] [default to null]
**Id** | **int32** | The project id | [optional] [default to null]
**TenantId** | **int32** | ID of the tenant owning the project | [optional] [default to null]
**ClusterUuid** | **string** | ID of the cluster owning the project | [optional] [default to null]
**DepartmentName** | **string** | Name of the department owning the project | [optional] [default to null]
**InteractiveNodeAffinity** | **string** |  | [optional] [default to null]
**TrainNodeAffinity** | **string** |  | [optional] [default to null]
**CreatedAt** | [**time.Time**](time.Time.md) | The creation date of the project | [optional] [default to null]
**Status** | [***ProjectStatus**](ProjectStatus.md) |  | [optional] [default to null]
**Phase** | **string** | project&#x27;s phase | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

