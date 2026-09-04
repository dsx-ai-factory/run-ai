# GenericSecretListResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Entries** | [**[]GenericSecret**](GenericSecret.md) |  | 

## Methods

### NewGenericSecretListResponse

`func NewGenericSecretListResponse(entries []GenericSecret, ) *GenericSecretListResponse`

NewGenericSecretListResponse instantiates a new GenericSecretListResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGenericSecretListResponseWithDefaults

`func NewGenericSecretListResponseWithDefaults() *GenericSecretListResponse`

NewGenericSecretListResponseWithDefaults instantiates a new GenericSecretListResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEntries

`func (o *GenericSecretListResponse) GetEntries() []GenericSecret`

GetEntries returns the Entries field if non-nil, zero value otherwise.

### GetEntriesOk

`func (o *GenericSecretListResponse) GetEntriesOk() (*[]GenericSecret, bool)`

GetEntriesOk returns a tuple with the Entries field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEntries

`func (o *GenericSecretListResponse) SetEntries(v []GenericSecret)`

SetEntries sets Entries field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


