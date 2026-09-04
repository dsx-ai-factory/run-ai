# WorkspaceCreationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The name of the workload. | 
**UseGivenNameAsPrefix** | Pointer to **bool** | When true, the requested name will be treated as a prefix. The final name of the workload will be composed of the name followed by a random set of characters. | [optional] [default to false]
**ProjectId** | **string** | The id of the project. | 
**ClusterId** | **string** | The id of the cluster. | 
**Spec** | Pointer to [**WorkspaceSpecSpec**](WorkspaceSpecSpec.md) |  | [optional] 

## Methods

### NewWorkspaceCreationRequest

`func NewWorkspaceCreationRequest(name string, projectId string, clusterId string, ) *WorkspaceCreationRequest`

NewWorkspaceCreationRequest instantiates a new WorkspaceCreationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkspaceCreationRequestWithDefaults

`func NewWorkspaceCreationRequestWithDefaults() *WorkspaceCreationRequest`

NewWorkspaceCreationRequestWithDefaults instantiates a new WorkspaceCreationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *WorkspaceCreationRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *WorkspaceCreationRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *WorkspaceCreationRequest) SetName(v string)`

SetName sets Name field to given value.


### GetUseGivenNameAsPrefix

`func (o *WorkspaceCreationRequest) GetUseGivenNameAsPrefix() bool`

GetUseGivenNameAsPrefix returns the UseGivenNameAsPrefix field if non-nil, zero value otherwise.

### GetUseGivenNameAsPrefixOk

`func (o *WorkspaceCreationRequest) GetUseGivenNameAsPrefixOk() (*bool, bool)`

GetUseGivenNameAsPrefixOk returns a tuple with the UseGivenNameAsPrefix field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUseGivenNameAsPrefix

`func (o *WorkspaceCreationRequest) SetUseGivenNameAsPrefix(v bool)`

SetUseGivenNameAsPrefix sets UseGivenNameAsPrefix field to given value.

### HasUseGivenNameAsPrefix

`func (o *WorkspaceCreationRequest) HasUseGivenNameAsPrefix() bool`

HasUseGivenNameAsPrefix returns a boolean if a field has been set.

### GetProjectId

`func (o *WorkspaceCreationRequest) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *WorkspaceCreationRequest) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *WorkspaceCreationRequest) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.


### GetClusterId

`func (o *WorkspaceCreationRequest) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *WorkspaceCreationRequest) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *WorkspaceCreationRequest) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetSpec

`func (o *WorkspaceCreationRequest) GetSpec() WorkspaceSpecSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *WorkspaceCreationRequest) GetSpecOk() (*WorkspaceSpecSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *WorkspaceCreationRequest) SetSpec(v WorkspaceSpecSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *WorkspaceCreationRequest) HasSpec() bool`

HasSpec returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


