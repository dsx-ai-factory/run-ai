# ProjectsResourcesResetRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ClusterId** | **string** | The id of the cluster. | 
**NodePoolIds** | **[]string** | The node pool ids to which the request applies | 
**Action** | [**ResourcesResetAction**](ResourcesResetAction.md) |  | 
**ParentId** | Pointer to **string** | apply only to projects under the specified department id | [optional] 

## Methods

### NewProjectsResourcesResetRequest

`func NewProjectsResourcesResetRequest(clusterId string, nodePoolIds []string, action ResourcesResetAction, ) *ProjectsResourcesResetRequest`

NewProjectsResourcesResetRequest instantiates a new ProjectsResourcesResetRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectsResourcesResetRequestWithDefaults

`func NewProjectsResourcesResetRequestWithDefaults() *ProjectsResourcesResetRequest`

NewProjectsResourcesResetRequestWithDefaults instantiates a new ProjectsResourcesResetRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetClusterId

`func (o *ProjectsResourcesResetRequest) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *ProjectsResourcesResetRequest) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *ProjectsResourcesResetRequest) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetNodePoolIds

`func (o *ProjectsResourcesResetRequest) GetNodePoolIds() []string`

GetNodePoolIds returns the NodePoolIds field if non-nil, zero value otherwise.

### GetNodePoolIdsOk

`func (o *ProjectsResourcesResetRequest) GetNodePoolIdsOk() (*[]string, bool)`

GetNodePoolIdsOk returns a tuple with the NodePoolIds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePoolIds

`func (o *ProjectsResourcesResetRequest) SetNodePoolIds(v []string)`

SetNodePoolIds sets NodePoolIds field to given value.


### GetAction

`func (o *ProjectsResourcesResetRequest) GetAction() ResourcesResetAction`

GetAction returns the Action field if non-nil, zero value otherwise.

### GetActionOk

`func (o *ProjectsResourcesResetRequest) GetActionOk() (*ResourcesResetAction, bool)`

GetActionOk returns a tuple with the Action field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAction

`func (o *ProjectsResourcesResetRequest) SetAction(v ResourcesResetAction)`

SetAction sets Action field to given value.


### GetParentId

`func (o *ProjectsResourcesResetRequest) GetParentId() string`

GetParentId returns the ParentId field if non-nil, zero value otherwise.

### GetParentIdOk

`func (o *ProjectsResourcesResetRequest) GetParentIdOk() (*string, bool)`

GetParentIdOk returns a tuple with the ParentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParentId

`func (o *ProjectsResourcesResetRequest) SetParentId(v string)`

SetParentId sets ParentId field to given value.

### HasParentId

`func (o *ProjectsResourcesResetRequest) HasParentId() bool`

HasParentId returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


