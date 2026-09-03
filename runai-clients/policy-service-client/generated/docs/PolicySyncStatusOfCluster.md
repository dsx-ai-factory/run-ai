# PolicySyncStatusOfCluster

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ClusterId** | **string** | The id of the cluster. | 
**Status** | **string** | Policy sync status of the cluster | 
**ErrorMessage** | Pointer to **NullableString** | In case sync failed, holds the error message received from the cluster | [optional] 

## Methods

### NewPolicySyncStatusOfCluster

`func NewPolicySyncStatusOfCluster(clusterId string, status string, ) *PolicySyncStatusOfCluster`

NewPolicySyncStatusOfCluster instantiates a new PolicySyncStatusOfCluster object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPolicySyncStatusOfClusterWithDefaults

`func NewPolicySyncStatusOfClusterWithDefaults() *PolicySyncStatusOfCluster`

NewPolicySyncStatusOfClusterWithDefaults instantiates a new PolicySyncStatusOfCluster object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetClusterId

`func (o *PolicySyncStatusOfCluster) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *PolicySyncStatusOfCluster) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *PolicySyncStatusOfCluster) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetStatus

`func (o *PolicySyncStatusOfCluster) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *PolicySyncStatusOfCluster) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *PolicySyncStatusOfCluster) SetStatus(v string)`

SetStatus sets Status field to given value.


### GetErrorMessage

`func (o *PolicySyncStatusOfCluster) GetErrorMessage() string`

GetErrorMessage returns the ErrorMessage field if non-nil, zero value otherwise.

### GetErrorMessageOk

`func (o *PolicySyncStatusOfCluster) GetErrorMessageOk() (*string, bool)`

GetErrorMessageOk returns a tuple with the ErrorMessage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetErrorMessage

`func (o *PolicySyncStatusOfCluster) SetErrorMessage(v string)`

SetErrorMessage sets ErrorMessage field to given value.

### HasErrorMessage

`func (o *PolicySyncStatusOfCluster) HasErrorMessage() bool`

HasErrorMessage returns a boolean if a field has been set.

### SetErrorMessageNil

`func (o *PolicySyncStatusOfCluster) SetErrorMessageNil(b bool)`

 SetErrorMessageNil sets the value for ErrorMessage to be an explicit nil

### UnsetErrorMessage
`func (o *PolicySyncStatusOfCluster) UnsetErrorMessage()`

UnsetErrorMessage ensures that no value is present for ErrorMessage, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


