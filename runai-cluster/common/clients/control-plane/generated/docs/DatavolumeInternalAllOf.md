# DatavolumeInternalAllOf

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SharedScopes** | Pointer to [**[]SharedScope**](SharedScope.md) | Will be returned only if the user has the required permissions to view those scopes | [optional] 
**Phase** | Pointer to [**DatavolumePhase**](DatavolumePhase.md) |  | [optional] 

## Methods

### NewDatavolumeInternalAllOf

`func NewDatavolumeInternalAllOf() *DatavolumeInternalAllOf`

NewDatavolumeInternalAllOf instantiates a new DatavolumeInternalAllOf object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDatavolumeInternalAllOfWithDefaults

`func NewDatavolumeInternalAllOfWithDefaults() *DatavolumeInternalAllOf`

NewDatavolumeInternalAllOfWithDefaults instantiates a new DatavolumeInternalAllOf object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSharedScopes

`func (o *DatavolumeInternalAllOf) GetSharedScopes() []SharedScope`

GetSharedScopes returns the SharedScopes field if non-nil, zero value otherwise.

### GetSharedScopesOk

`func (o *DatavolumeInternalAllOf) GetSharedScopesOk() (*[]SharedScope, bool)`

GetSharedScopesOk returns a tuple with the SharedScopes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSharedScopes

`func (o *DatavolumeInternalAllOf) SetSharedScopes(v []SharedScope)`

SetSharedScopes sets SharedScopes field to given value.

### HasSharedScopes

`func (o *DatavolumeInternalAllOf) HasSharedScopes() bool`

HasSharedScopes returns a boolean if a field has been set.

### GetPhase

`func (o *DatavolumeInternalAllOf) GetPhase() DatavolumePhase`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *DatavolumeInternalAllOf) GetPhaseOk() (*DatavolumePhase, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *DatavolumeInternalAllOf) SetPhase(v DatavolumePhase)`

SetPhase sets Phase field to given value.

### HasPhase

`func (o *DatavolumeInternalAllOf) HasPhase() bool`

HasPhase returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


