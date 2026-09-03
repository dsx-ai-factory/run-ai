# WorkloadCreationMeta

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The name of the workload. | 
**UseGivenNameAsPrefix** | Pointer to **bool** | When true, the requested name will be treated as a prefix. The final name of the workload will be composed of the name followed by a random set of characters. | [optional] [default to false]
**ProjectId** | **string** | The id of the project. | 
**ClusterId** | **string** | The id of the cluster. | 

## Methods

### NewWorkloadCreationMeta

`func NewWorkloadCreationMeta(name string, projectId string, clusterId string, ) *WorkloadCreationMeta`

NewWorkloadCreationMeta instantiates a new WorkloadCreationMeta object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadCreationMetaWithDefaults

`func NewWorkloadCreationMetaWithDefaults() *WorkloadCreationMeta`

NewWorkloadCreationMetaWithDefaults instantiates a new WorkloadCreationMeta object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *WorkloadCreationMeta) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *WorkloadCreationMeta) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *WorkloadCreationMeta) SetName(v string)`

SetName sets Name field to given value.


### GetUseGivenNameAsPrefix

`func (o *WorkloadCreationMeta) GetUseGivenNameAsPrefix() bool`

GetUseGivenNameAsPrefix returns the UseGivenNameAsPrefix field if non-nil, zero value otherwise.

### GetUseGivenNameAsPrefixOk

`func (o *WorkloadCreationMeta) GetUseGivenNameAsPrefixOk() (*bool, bool)`

GetUseGivenNameAsPrefixOk returns a tuple with the UseGivenNameAsPrefix field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUseGivenNameAsPrefix

`func (o *WorkloadCreationMeta) SetUseGivenNameAsPrefix(v bool)`

SetUseGivenNameAsPrefix sets UseGivenNameAsPrefix field to given value.

### HasUseGivenNameAsPrefix

`func (o *WorkloadCreationMeta) HasUseGivenNameAsPrefix() bool`

HasUseGivenNameAsPrefix returns a boolean if a field has been set.

### GetProjectId

`func (o *WorkloadCreationMeta) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *WorkloadCreationMeta) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *WorkloadCreationMeta) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.


### GetClusterId

`func (o *WorkloadCreationMeta) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *WorkloadCreationMeta) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *WorkloadCreationMeta) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


