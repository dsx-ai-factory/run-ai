# DepartmentsResourcesResetRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ClusterId** | **string** | The id of the cluster. | 
**NodePoolIds** | **[]string** | The node pool ids to which the request applies | 
**Action** | [**ResourcesResetAction**](ResourcesResetAction.md) |  | 

## Methods

### NewDepartmentsResourcesResetRequest

`func NewDepartmentsResourcesResetRequest(clusterId string, nodePoolIds []string, action ResourcesResetAction, ) *DepartmentsResourcesResetRequest`

NewDepartmentsResourcesResetRequest instantiates a new DepartmentsResourcesResetRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDepartmentsResourcesResetRequestWithDefaults

`func NewDepartmentsResourcesResetRequestWithDefaults() *DepartmentsResourcesResetRequest`

NewDepartmentsResourcesResetRequestWithDefaults instantiates a new DepartmentsResourcesResetRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetClusterId

`func (o *DepartmentsResourcesResetRequest) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *DepartmentsResourcesResetRequest) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *DepartmentsResourcesResetRequest) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetNodePoolIds

`func (o *DepartmentsResourcesResetRequest) GetNodePoolIds() []string`

GetNodePoolIds returns the NodePoolIds field if non-nil, zero value otherwise.

### GetNodePoolIdsOk

`func (o *DepartmentsResourcesResetRequest) GetNodePoolIdsOk() (*[]string, bool)`

GetNodePoolIdsOk returns a tuple with the NodePoolIds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePoolIds

`func (o *DepartmentsResourcesResetRequest) SetNodePoolIds(v []string)`

SetNodePoolIds sets NodePoolIds field to given value.


### GetAction

`func (o *DepartmentsResourcesResetRequest) GetAction() ResourcesResetAction`

GetAction returns the Action field if non-nil, zero value otherwise.

### GetActionOk

`func (o *DepartmentsResourcesResetRequest) GetActionOk() (*ResourcesResetAction, bool)`

GetActionOk returns a tuple with the Action field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAction

`func (o *DepartmentsResourcesResetRequest) SetAction(v ResourcesResetAction)`

SetAction sets Action field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


