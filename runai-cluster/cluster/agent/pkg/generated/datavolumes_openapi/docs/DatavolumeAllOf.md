# DatavolumeAllOf

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SharedScopes** | Pointer to [**[]SharedScope**](SharedScope.md) | Will be returned only if the user has the required permissions to view those scopes | [optional] 
**Status** | [**DatavolumeStatus**](DatavolumeStatus.md) |  | 

## Methods

### NewDatavolumeAllOf

`func NewDatavolumeAllOf(status DatavolumeStatus, ) *DatavolumeAllOf`

NewDatavolumeAllOf instantiates a new DatavolumeAllOf object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDatavolumeAllOfWithDefaults

`func NewDatavolumeAllOfWithDefaults() *DatavolumeAllOf`

NewDatavolumeAllOfWithDefaults instantiates a new DatavolumeAllOf object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSharedScopes

`func (o *DatavolumeAllOf) GetSharedScopes() []SharedScope`

GetSharedScopes returns the SharedScopes field if non-nil, zero value otherwise.

### GetSharedScopesOk

`func (o *DatavolumeAllOf) GetSharedScopesOk() (*[]SharedScope, bool)`

GetSharedScopesOk returns a tuple with the SharedScopes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSharedScopes

`func (o *DatavolumeAllOf) SetSharedScopes(v []SharedScope)`

SetSharedScopes sets SharedScopes field to given value.

### HasSharedScopes

`func (o *DatavolumeAllOf) HasSharedScopes() bool`

HasSharedScopes returns a boolean if a field has been set.

### GetStatus

`func (o *DatavolumeAllOf) GetStatus() DatavolumeStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *DatavolumeAllOf) GetStatusOk() (*DatavolumeStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *DatavolumeAllOf) SetStatus(v DatavolumeStatus)`

SetStatus sets Status field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


