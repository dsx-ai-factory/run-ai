# WorkloadSlim

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** |  | 
**Name** | **string** |  | 
**Type** | Pointer to **string** |  | [optional] 
**ProjectName** | **string** |  | 
**ClusterId** | **string** | The id of the cluster. | 
**DistributedFramework** | Pointer to **string** |  | [optional] 

## Methods

### NewWorkloadSlim

`func NewWorkloadSlim(id string, name string, projectName string, clusterId string, ) *WorkloadSlim`

NewWorkloadSlim instantiates a new WorkloadSlim object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadSlimWithDefaults

`func NewWorkloadSlimWithDefaults() *WorkloadSlim`

NewWorkloadSlimWithDefaults instantiates a new WorkloadSlim object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *WorkloadSlim) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *WorkloadSlim) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *WorkloadSlim) SetId(v string)`

SetId sets Id field to given value.


### GetName

`func (o *WorkloadSlim) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *WorkloadSlim) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *WorkloadSlim) SetName(v string)`

SetName sets Name field to given value.


### GetType

`func (o *WorkloadSlim) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *WorkloadSlim) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *WorkloadSlim) SetType(v string)`

SetType sets Type field to given value.

### HasType

`func (o *WorkloadSlim) HasType() bool`

HasType returns a boolean if a field has been set.

### GetProjectName

`func (o *WorkloadSlim) GetProjectName() string`

GetProjectName returns the ProjectName field if non-nil, zero value otherwise.

### GetProjectNameOk

`func (o *WorkloadSlim) GetProjectNameOk() (*string, bool)`

GetProjectNameOk returns a tuple with the ProjectName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectName

`func (o *WorkloadSlim) SetProjectName(v string)`

SetProjectName sets ProjectName field to given value.


### GetClusterId

`func (o *WorkloadSlim) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *WorkloadSlim) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *WorkloadSlim) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetDistributedFramework

`func (o *WorkloadSlim) GetDistributedFramework() string`

GetDistributedFramework returns the DistributedFramework field if non-nil, zero value otherwise.

### GetDistributedFrameworkOk

`func (o *WorkloadSlim) GetDistributedFrameworkOk() (*string, bool)`

GetDistributedFrameworkOk returns a tuple with the DistributedFramework field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDistributedFramework

`func (o *WorkloadSlim) SetDistributedFramework(v string)`

SetDistributedFramework sets DistributedFramework field to given value.

### HasDistributedFramework

`func (o *WorkloadSlim) HasDistributedFramework() bool`

HasDistributedFramework returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


