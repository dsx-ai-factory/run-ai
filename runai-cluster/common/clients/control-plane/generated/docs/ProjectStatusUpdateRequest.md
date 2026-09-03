# ProjectStatusUpdateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** |  | 
**ClusterId** | **string** | The id of the cluster. | 
**Status** | **string** |  | 
**QuotaStatuses** | Pointer to [**[]ProjectStatusUpdateRequestQuotaStatusesInner**](ProjectStatusUpdateRequestQuotaStatusesInner.md) |  | [optional] 

## Methods

### NewProjectStatusUpdateRequest

`func NewProjectStatusUpdateRequest(name string, clusterId string, status string, ) *ProjectStatusUpdateRequest`

NewProjectStatusUpdateRequest instantiates a new ProjectStatusUpdateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectStatusUpdateRequestWithDefaults

`func NewProjectStatusUpdateRequestWithDefaults() *ProjectStatusUpdateRequest`

NewProjectStatusUpdateRequestWithDefaults instantiates a new ProjectStatusUpdateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *ProjectStatusUpdateRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ProjectStatusUpdateRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ProjectStatusUpdateRequest) SetName(v string)`

SetName sets Name field to given value.


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

`func (o *ProjectStatusUpdateRequest) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *ProjectStatusUpdateRequest) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *ProjectStatusUpdateRequest) SetStatus(v string)`

SetStatus sets Status field to given value.


### GetQuotaStatuses

`func (o *ProjectStatusUpdateRequest) GetQuotaStatuses() []ProjectStatusUpdateRequestQuotaStatusesInner`

GetQuotaStatuses returns the QuotaStatuses field if non-nil, zero value otherwise.

### GetQuotaStatusesOk

`func (o *ProjectStatusUpdateRequest) GetQuotaStatusesOk() (*[]ProjectStatusUpdateRequestQuotaStatusesInner, bool)`

GetQuotaStatusesOk returns a tuple with the QuotaStatuses field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetQuotaStatuses

`func (o *ProjectStatusUpdateRequest) SetQuotaStatuses(v []ProjectStatusUpdateRequestQuotaStatusesInner)`

SetQuotaStatuses sets QuotaStatuses field to given value.

### HasQuotaStatuses

`func (o *ProjectStatusUpdateRequest) HasQuotaStatuses() bool`

HasQuotaStatuses returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


