# DepartmentNodePoolResources

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | The unique ID of the resources. | 
**OrgUnitId** | **string** | The project/department ID. | 
**OrgUnitName** | **string** | The project/department name. | 
**ClusterId** | **string** |  | 
**NodePoolId** | **string** |  | 
**NodePoolName** | **string** |  | 
**Rank** | **string** |  | 
**Resources** | [**ResourcesPerNodePool**](ResourcesPerNodePool.md) |  | 

## Methods

### NewDepartmentNodePoolResources

`func NewDepartmentNodePoolResources(id string, orgUnitId string, orgUnitName string, clusterId string, nodePoolId string, nodePoolName string, rank string, resources ResourcesPerNodePool, ) *DepartmentNodePoolResources`

NewDepartmentNodePoolResources instantiates a new DepartmentNodePoolResources object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDepartmentNodePoolResourcesWithDefaults

`func NewDepartmentNodePoolResourcesWithDefaults() *DepartmentNodePoolResources`

NewDepartmentNodePoolResourcesWithDefaults instantiates a new DepartmentNodePoolResources object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *DepartmentNodePoolResources) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *DepartmentNodePoolResources) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *DepartmentNodePoolResources) SetId(v string)`

SetId sets Id field to given value.


### GetOrgUnitId

`func (o *DepartmentNodePoolResources) GetOrgUnitId() string`

GetOrgUnitId returns the OrgUnitId field if non-nil, zero value otherwise.

### GetOrgUnitIdOk

`func (o *DepartmentNodePoolResources) GetOrgUnitIdOk() (*string, bool)`

GetOrgUnitIdOk returns a tuple with the OrgUnitId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrgUnitId

`func (o *DepartmentNodePoolResources) SetOrgUnitId(v string)`

SetOrgUnitId sets OrgUnitId field to given value.


### GetOrgUnitName

`func (o *DepartmentNodePoolResources) GetOrgUnitName() string`

GetOrgUnitName returns the OrgUnitName field if non-nil, zero value otherwise.

### GetOrgUnitNameOk

`func (o *DepartmentNodePoolResources) GetOrgUnitNameOk() (*string, bool)`

GetOrgUnitNameOk returns a tuple with the OrgUnitName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrgUnitName

`func (o *DepartmentNodePoolResources) SetOrgUnitName(v string)`

SetOrgUnitName sets OrgUnitName field to given value.


### GetClusterId

`func (o *DepartmentNodePoolResources) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *DepartmentNodePoolResources) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *DepartmentNodePoolResources) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetNodePoolId

`func (o *DepartmentNodePoolResources) GetNodePoolId() string`

GetNodePoolId returns the NodePoolId field if non-nil, zero value otherwise.

### GetNodePoolIdOk

`func (o *DepartmentNodePoolResources) GetNodePoolIdOk() (*string, bool)`

GetNodePoolIdOk returns a tuple with the NodePoolId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePoolId

`func (o *DepartmentNodePoolResources) SetNodePoolId(v string)`

SetNodePoolId sets NodePoolId field to given value.


### GetNodePoolName

`func (o *DepartmentNodePoolResources) GetNodePoolName() string`

GetNodePoolName returns the NodePoolName field if non-nil, zero value otherwise.

### GetNodePoolNameOk

`func (o *DepartmentNodePoolResources) GetNodePoolNameOk() (*string, bool)`

GetNodePoolNameOk returns a tuple with the NodePoolName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePoolName

`func (o *DepartmentNodePoolResources) SetNodePoolName(v string)`

SetNodePoolName sets NodePoolName field to given value.


### GetRank

`func (o *DepartmentNodePoolResources) GetRank() string`

GetRank returns the Rank field if non-nil, zero value otherwise.

### GetRankOk

`func (o *DepartmentNodePoolResources) GetRankOk() (*string, bool)`

GetRankOk returns a tuple with the Rank field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRank

`func (o *DepartmentNodePoolResources) SetRank(v string)`

SetRank sets Rank field to given value.


### GetResources

`func (o *DepartmentNodePoolResources) GetResources() ResourcesPerNodePool`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *DepartmentNodePoolResources) GetResourcesOk() (*ResourcesPerNodePool, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *DepartmentNodePoolResources) SetResources(v ResourcesPerNodePool)`

SetResources sets Resources field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


