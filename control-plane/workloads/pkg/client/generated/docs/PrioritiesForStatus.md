# PrioritiesForStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**FullSync** | Pointer to **bool** | Upsert all items and delete all missing items | [optional] 
**Priorities** | Pointer to [**[]PrioritiesForStatusPrioritiesInner**](PrioritiesForStatusPrioritiesInner.md) |  | [optional] 

## Methods

### NewPrioritiesForStatus

`func NewPrioritiesForStatus() *PrioritiesForStatus`

NewPrioritiesForStatus instantiates a new PrioritiesForStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPrioritiesForStatusWithDefaults

`func NewPrioritiesForStatusWithDefaults() *PrioritiesForStatus`

NewPrioritiesForStatusWithDefaults instantiates a new PrioritiesForStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetFullSync

`func (o *PrioritiesForStatus) GetFullSync() bool`

GetFullSync returns the FullSync field if non-nil, zero value otherwise.

### GetFullSyncOk

`func (o *PrioritiesForStatus) GetFullSyncOk() (*bool, bool)`

GetFullSyncOk returns a tuple with the FullSync field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFullSync

`func (o *PrioritiesForStatus) SetFullSync(v bool)`

SetFullSync sets FullSync field to given value.

### HasFullSync

`func (o *PrioritiesForStatus) HasFullSync() bool`

HasFullSync returns a boolean if a field has been set.

### GetPriorities

`func (o *PrioritiesForStatus) GetPriorities() []PrioritiesForStatusPrioritiesInner`

GetPriorities returns the Priorities field if non-nil, zero value otherwise.

### GetPrioritiesOk

`func (o *PrioritiesForStatus) GetPrioritiesOk() (*[]PrioritiesForStatusPrioritiesInner, bool)`

GetPrioritiesOk returns a tuple with the Priorities field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriorities

`func (o *PrioritiesForStatus) SetPriorities(v []PrioritiesForStatusPrioritiesInner)`

SetPriorities sets Priorities field to given value.

### HasPriorities

`func (o *PrioritiesForStatus) HasPriorities() bool`

HasPriorities returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


