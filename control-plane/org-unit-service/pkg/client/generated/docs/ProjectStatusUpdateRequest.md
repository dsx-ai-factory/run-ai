# ProjectStatusUpdateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** |  | 
**ClusterId** | **string** | The id of the cluster. | 
**Status** | [**ProjectStatus**](ProjectStatus.md) |  | 

## Methods

### NewProjectStatusUpdateRequest

`func NewProjectStatusUpdateRequest(id string, clusterId string, status ProjectStatus, ) *ProjectStatusUpdateRequest`

NewProjectStatusUpdateRequest instantiates a new ProjectStatusUpdateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectStatusUpdateRequestWithDefaults

`func NewProjectStatusUpdateRequestWithDefaults() *ProjectStatusUpdateRequest`

NewProjectStatusUpdateRequestWithDefaults instantiates a new ProjectStatusUpdateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *ProjectStatusUpdateRequest) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *ProjectStatusUpdateRequest) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *ProjectStatusUpdateRequest) SetId(v string)`

SetId sets Id field to given value.


### GetClusterId

`func (o *ProjectStatusUpdateRequest) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *ProjectStatusUpdateRequest) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *ProjectStatusUpdateRequest) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetStatus

`func (o *ProjectStatusUpdateRequest) GetStatus() ProjectStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *ProjectStatusUpdateRequest) GetStatusOk() (*ProjectStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *ProjectStatusUpdateRequest) SetStatus(v ProjectStatus)`

SetStatus sets Status field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


