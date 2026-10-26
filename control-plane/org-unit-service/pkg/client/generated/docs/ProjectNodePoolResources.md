# ProjectNodePoolResources

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | The unique ID of the resources. | 
**OrgUnitId** | **string** | The project/department ID. | 
**OrgUnitName** | **string** | The project/department name. | 
**ParentId** | **string** |  | 
**ClusterId** | **string** |  | 
**NodePoolId** | **string** |  | 
**NodePoolName** | **string** |  | 
**Rank** | **string** |  | 
**Resources** | [**ResourcesPerNodePool**](ResourcesPerNodePool.md) |  | 
**QuotaStatus** | Pointer to [**QuotaStatus**](QuotaStatus.md) |  | [optional] 

## Methods

### NewProjectNodePoolResources

`func NewProjectNodePoolResources(id string, orgUnitId string, orgUnitName string, parentId string, clusterId string, nodePoolId string, nodePoolName string, rank string, resources ResourcesPerNodePool, ) *ProjectNodePoolResources`

NewProjectNodePoolResources instantiates a new ProjectNodePoolResources object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectNodePoolResourcesWithDefaults

`func NewProjectNodePoolResourcesWithDefaults() *ProjectNodePoolResources`

NewProjectNodePoolResourcesWithDefaults instantiates a new ProjectNodePoolResources object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *ProjectNodePoolResources) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *ProjectNodePoolResources) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *ProjectNodePoolResources) SetId(v string)`

SetId sets Id field to given value.


### GetOrgUnitId

`func (o *ProjectNodePoolResources) GetOrgUnitId() string`

GetOrgUnitId returns the OrgUnitId field if non-nil, zero value otherwise.

### GetOrgUnitIdOk

`func (o *ProjectNodePoolResources) GetOrgUnitIdOk() (*string, bool)`

GetOrgUnitIdOk returns a tuple with the OrgUnitId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrgUnitId

`func (o *ProjectNodePoolResources) SetOrgUnitId(v string)`

SetOrgUnitId sets OrgUnitId field to given value.


### GetOrgUnitName

`func (o *ProjectNodePoolResources) GetOrgUnitName() string`

GetOrgUnitName returns the OrgUnitName field if non-nil, zero value otherwise.

### GetOrgUnitNameOk

`func (o *ProjectNodePoolResources) GetOrgUnitNameOk() (*string, bool)`

GetOrgUnitNameOk returns a tuple with the OrgUnitName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrgUnitName

`func (o *ProjectNodePoolResources) SetOrgUnitName(v string)`

SetOrgUnitName sets OrgUnitName field to given value.


### GetParentId

`func (o *ProjectNodePoolResources) GetParentId() string`

GetParentId returns the ParentId field if non-nil, zero value otherwise.

### GetParentIdOk

`func (o *ProjectNodePoolResources) GetParentIdOk() (*string, bool)`

GetParentIdOk returns a tuple with the ParentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParentId

`func (o *ProjectNodePoolResources) SetParentId(v string)`

SetParentId sets ParentId field to given value.


### GetClusterId

`func (o *ProjectNodePoolResources) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *ProjectNodePoolResources) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *ProjectNodePoolResources) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetNodePoolId

`func (o *ProjectNodePoolResources) GetNodePoolId() string`

GetNodePoolId returns the NodePoolId field if non-nil, zero value otherwise.

### GetNodePoolIdOk

`func (o *ProjectNodePoolResources) GetNodePoolIdOk() (*string, bool)`

GetNodePoolIdOk returns a tuple with the NodePoolId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePoolId

`func (o *ProjectNodePoolResources) SetNodePoolId(v string)`

SetNodePoolId sets NodePoolId field to given value.


### GetNodePoolName

`func (o *ProjectNodePoolResources) GetNodePoolName() string`

GetNodePoolName returns the NodePoolName field if non-nil, zero value otherwise.

### GetNodePoolNameOk

`func (o *ProjectNodePoolResources) GetNodePoolNameOk() (*string, bool)`

GetNodePoolNameOk returns a tuple with the NodePoolName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePoolName

`func (o *ProjectNodePoolResources) SetNodePoolName(v string)`

SetNodePoolName sets NodePoolName field to given value.


### GetRank

`func (o *ProjectNodePoolResources) GetRank() string`

GetRank returns the Rank field if non-nil, zero value otherwise.

### GetRankOk

`func (o *ProjectNodePoolResources) GetRankOk() (*string, bool)`

GetRankOk returns a tuple with the Rank field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRank

`func (o *ProjectNodePoolResources) SetRank(v string)`

SetRank sets Rank field to given value.


### GetResources

`func (o *ProjectNodePoolResources) GetResources() ResourcesPerNodePool`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *ProjectNodePoolResources) GetResourcesOk() (*ResourcesPerNodePool, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *ProjectNodePoolResources) SetResources(v ResourcesPerNodePool)`

SetResources sets Resources field to given value.


### GetQuotaStatus

`func (o *ProjectNodePoolResources) GetQuotaStatus() QuotaStatus`

GetQuotaStatus returns the QuotaStatus field if non-nil, zero value otherwise.

### GetQuotaStatusOk

`func (o *ProjectNodePoolResources) GetQuotaStatusOk() (*QuotaStatus, bool)`

GetQuotaStatusOk returns a tuple with the QuotaStatus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetQuotaStatus

`func (o *ProjectNodePoolResources) SetQuotaStatus(v QuotaStatus)`

SetQuotaStatus sets QuotaStatus field to given value.

### HasQuotaStatus

`func (o *ProjectNodePoolResources) HasQuotaStatus() bool`

HasQuotaStatus returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


