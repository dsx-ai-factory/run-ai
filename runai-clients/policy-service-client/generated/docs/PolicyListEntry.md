# PolicyListEntry

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Type** | Pointer to [**PolicyType**](PolicyType.md) |  | [optional] 
**Meta** | Pointer to [**PolicyMeta**](PolicyMeta.md) |  | [optional] 
**Status** | Pointer to [**PolicyListEntryStatus**](PolicyListEntryStatus.md) |  | [optional] 
**Availability** | Pointer to [**NullablePolicyStatus**](PolicyStatus.md) |  | [optional] 

## Methods

### NewPolicyListEntry

`func NewPolicyListEntry() *PolicyListEntry`

NewPolicyListEntry instantiates a new PolicyListEntry object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPolicyListEntryWithDefaults

`func NewPolicyListEntryWithDefaults() *PolicyListEntry`

NewPolicyListEntryWithDefaults instantiates a new PolicyListEntry object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetType

`func (o *PolicyListEntry) GetType() PolicyType`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *PolicyListEntry) GetTypeOk() (*PolicyType, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *PolicyListEntry) SetType(v PolicyType)`

SetType sets Type field to given value.

### HasType

`func (o *PolicyListEntry) HasType() bool`

HasType returns a boolean if a field has been set.

### GetMeta

`func (o *PolicyListEntry) GetMeta() PolicyMeta`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *PolicyListEntry) GetMetaOk() (*PolicyMeta, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *PolicyListEntry) SetMeta(v PolicyMeta)`

SetMeta sets Meta field to given value.

### HasMeta

`func (o *PolicyListEntry) HasMeta() bool`

HasMeta returns a boolean if a field has been set.

### GetStatus

`func (o *PolicyListEntry) GetStatus() PolicyListEntryStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *PolicyListEntry) GetStatusOk() (*PolicyListEntryStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *PolicyListEntry) SetStatus(v PolicyListEntryStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *PolicyListEntry) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetAvailability

`func (o *PolicyListEntry) GetAvailability() PolicyStatus`

GetAvailability returns the Availability field if non-nil, zero value otherwise.

### GetAvailabilityOk

`func (o *PolicyListEntry) GetAvailabilityOk() (*PolicyStatus, bool)`

GetAvailabilityOk returns a tuple with the Availability field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAvailability

`func (o *PolicyListEntry) SetAvailability(v PolicyStatus)`

SetAvailability sets Availability field to given value.

### HasAvailability

`func (o *PolicyListEntry) HasAvailability() bool`

HasAvailability returns a boolean if a field has been set.

### SetAvailabilityNil

`func (o *PolicyListEntry) SetAvailabilityNil(b bool)`

 SetAvailabilityNil sets the value for Availability to be an explicit nil

### UnsetAvailability
`func (o *PolicyListEntry) UnsetAvailability()`

UnsetAvailability ensures that no value is present for Availability, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


