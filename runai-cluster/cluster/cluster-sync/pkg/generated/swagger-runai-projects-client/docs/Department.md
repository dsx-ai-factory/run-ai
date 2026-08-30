# Department

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The name of the department. | [optional] [default to null]
**DeservedGpus** | **float64** | Deserved GPUs for the department. | [optional] [default to null]
**AllowOverQuota** | **bool** | Is over quota allowed for the department. | [optional] [default to null]
**MaxAllowedGpus** | **float64** | Max allowed GPUs for the department. | [optional] [default to null]
**Resources** | [***AssignedResources**](AssignedResources.md) |  | [optional] [default to null]
**NodePoolsResources** | [**[]NodePoolAssignedResources**](NodePoolAssignedResources.md) | Resources assigned to the Department per node pool | [optional] [default to null]
**TenantId** | **int32** | The tenant id this cluster belongs to. | [optional] [default to null]
**ClusterUuid** | **string** | The cluster UUID this department belongs to. | [optional] [default to null]
**CreatedAt** | [**time.Time**](time.Time.md) | The creation date of the department. | [optional] [default to null]
**Id** | **int32** | The unique id identifying the department. | [optional] [default to null]
**Projects** | [**[]Project**](Project.md) | Projects under this department. | [optional] [default to null]
**ProjectsDeservedGpus** | **string** | Total deserved GPUs of the projects under this department - as string. | [optional] [default to null]
**DepartmentAdmins** | **[]string** | Id&#x27;s of users with department admin role that are assigned to managed the department | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

