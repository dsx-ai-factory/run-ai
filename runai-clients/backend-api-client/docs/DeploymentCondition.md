# DeploymentCondition

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Type** | **string** |  | 
**Status** | **string** |  | 
**Reason** | Pointer to **string** |  | [optional] 
**Message** | Pointer to **string** |  | [optional] 

## Methods

### NewDeploymentCondition

`func NewDeploymentCondition(type_ string, status string, ) *DeploymentCondition`

NewDeploymentCondition instantiates a new DeploymentCondition object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDeploymentConditionWithDefaults

`func NewDeploymentConditionWithDefaults() *DeploymentCondition`

NewDeploymentConditionWithDefaults instantiates a new DeploymentCondition object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetType

`func (o *DeploymentCondition) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *DeploymentCondition) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *DeploymentCondition) SetType(v string)`

SetType sets Type field to given value.


### GetStatus

`func (o *DeploymentCondition) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *DeploymentCondition) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *DeploymentCondition) SetStatus(v string)`

SetStatus sets Status field to given value.


### GetReason

`func (o *DeploymentCondition) GetReason() string`

GetReason returns the Reason field if non-nil, zero value otherwise.

### GetReasonOk

`func (o *DeploymentCondition) GetReasonOk() (*string, bool)`

GetReasonOk returns a tuple with the Reason field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReason

`func (o *DeploymentCondition) SetReason(v string)`

SetReason sets Reason field to given value.

### HasReason

`func (o *DeploymentCondition) HasReason() bool`

HasReason returns a boolean if a field has been set.

### GetMessage

`func (o *DeploymentCondition) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *DeploymentCondition) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *DeploymentCondition) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *DeploymentCondition) HasMessage() bool`

HasMessage returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


