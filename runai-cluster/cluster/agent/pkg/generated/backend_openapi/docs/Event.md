# Event

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Uid** | Pointer to **string** |  | [optional] 
**Type** | Pointer to **string** |  | [optional] 
**Message** | Pointer to **string** |  | [optional] 
**Reason** | Pointer to **string** |  | [optional] 
**CreationTimestamp** | Pointer to **float32** |  | [optional] 
**SourceComponent** | Pointer to **string** |  | [optional] 
**InvolvedObjUID** | Pointer to **string** |  | [optional] 
**InvolvedObjKind** | Pointer to **string** |  | [optional] 
**InvolvedObjName** | Pointer to **string** |  | [optional] 
**InvolvedObjNamespace** | Pointer to **string** |  | [optional] 

## Methods

### NewEvent

`func NewEvent() *Event`

NewEvent instantiates a new Event object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewEventWithDefaults

`func NewEventWithDefaults() *Event`

NewEventWithDefaults instantiates a new Event object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUid

`func (o *Event) GetUid() string`

GetUid returns the Uid field if non-nil, zero value otherwise.

### GetUidOk

`func (o *Event) GetUidOk() (*string, bool)`

GetUidOk returns a tuple with the Uid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUid

`func (o *Event) SetUid(v string)`

SetUid sets Uid field to given value.

### HasUid

`func (o *Event) HasUid() bool`

HasUid returns a boolean if a field has been set.

### GetType

`func (o *Event) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *Event) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *Event) SetType(v string)`

SetType sets Type field to given value.

### HasType

`func (o *Event) HasType() bool`

HasType returns a boolean if a field has been set.

### GetMessage

`func (o *Event) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *Event) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *Event) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *Event) HasMessage() bool`

HasMessage returns a boolean if a field has been set.

### GetReason

`func (o *Event) GetReason() string`

GetReason returns the Reason field if non-nil, zero value otherwise.

### GetReasonOk

`func (o *Event) GetReasonOk() (*string, bool)`

GetReasonOk returns a tuple with the Reason field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReason

`func (o *Event) SetReason(v string)`

SetReason sets Reason field to given value.

### HasReason

`func (o *Event) HasReason() bool`

HasReason returns a boolean if a field has been set.

### GetCreationTimestamp

`func (o *Event) GetCreationTimestamp() float32`

GetCreationTimestamp returns the CreationTimestamp field if non-nil, zero value otherwise.

### GetCreationTimestampOk

`func (o *Event) GetCreationTimestampOk() (*float32, bool)`

GetCreationTimestampOk returns a tuple with the CreationTimestamp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreationTimestamp

`func (o *Event) SetCreationTimestamp(v float32)`

SetCreationTimestamp sets CreationTimestamp field to given value.

### HasCreationTimestamp

`func (o *Event) HasCreationTimestamp() bool`

HasCreationTimestamp returns a boolean if a field has been set.

### GetSourceComponent

`func (o *Event) GetSourceComponent() string`

GetSourceComponent returns the SourceComponent field if non-nil, zero value otherwise.

### GetSourceComponentOk

`func (o *Event) GetSourceComponentOk() (*string, bool)`

GetSourceComponentOk returns a tuple with the SourceComponent field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceComponent

`func (o *Event) SetSourceComponent(v string)`

SetSourceComponent sets SourceComponent field to given value.

### HasSourceComponent

`func (o *Event) HasSourceComponent() bool`

HasSourceComponent returns a boolean if a field has been set.

### GetInvolvedObjUID

`func (o *Event) GetInvolvedObjUID() string`

GetInvolvedObjUID returns the InvolvedObjUID field if non-nil, zero value otherwise.

### GetInvolvedObjUIDOk

`func (o *Event) GetInvolvedObjUIDOk() (*string, bool)`

GetInvolvedObjUIDOk returns a tuple with the InvolvedObjUID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInvolvedObjUID

`func (o *Event) SetInvolvedObjUID(v string)`

SetInvolvedObjUID sets InvolvedObjUID field to given value.

### HasInvolvedObjUID

`func (o *Event) HasInvolvedObjUID() bool`

HasInvolvedObjUID returns a boolean if a field has been set.

### GetInvolvedObjKind

`func (o *Event) GetInvolvedObjKind() string`

GetInvolvedObjKind returns the InvolvedObjKind field if non-nil, zero value otherwise.

### GetInvolvedObjKindOk

`func (o *Event) GetInvolvedObjKindOk() (*string, bool)`

GetInvolvedObjKindOk returns a tuple with the InvolvedObjKind field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInvolvedObjKind

`func (o *Event) SetInvolvedObjKind(v string)`

SetInvolvedObjKind sets InvolvedObjKind field to given value.

### HasInvolvedObjKind

`func (o *Event) HasInvolvedObjKind() bool`

HasInvolvedObjKind returns a boolean if a field has been set.

### GetInvolvedObjName

`func (o *Event) GetInvolvedObjName() string`

GetInvolvedObjName returns the InvolvedObjName field if non-nil, zero value otherwise.

### GetInvolvedObjNameOk

`func (o *Event) GetInvolvedObjNameOk() (*string, bool)`

GetInvolvedObjNameOk returns a tuple with the InvolvedObjName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInvolvedObjName

`func (o *Event) SetInvolvedObjName(v string)`

SetInvolvedObjName sets InvolvedObjName field to given value.

### HasInvolvedObjName

`func (o *Event) HasInvolvedObjName() bool`

HasInvolvedObjName returns a boolean if a field has been set.

### GetInvolvedObjNamespace

`func (o *Event) GetInvolvedObjNamespace() string`

GetInvolvedObjNamespace returns the InvolvedObjNamespace field if non-nil, zero value otherwise.

### GetInvolvedObjNamespaceOk

`func (o *Event) GetInvolvedObjNamespaceOk() (*string, bool)`

GetInvolvedObjNamespaceOk returns a tuple with the InvolvedObjNamespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInvolvedObjNamespace

`func (o *Event) SetInvolvedObjNamespace(v string)`

SetInvolvedObjNamespace sets InvolvedObjNamespace field to given value.

### HasInvolvedObjNamespace

`func (o *Event) HasInvolvedObjNamespace() bool`

HasInvolvedObjNamespace returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


