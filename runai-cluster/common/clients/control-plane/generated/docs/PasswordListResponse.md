# PasswordListResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Entries** | [**[]Password**](Password.md) |  | 

## Methods

### NewPasswordListResponse

`func NewPasswordListResponse(entries []Password, ) *PasswordListResponse`

NewPasswordListResponse instantiates a new PasswordListResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPasswordListResponseWithDefaults

`func NewPasswordListResponseWithDefaults() *PasswordListResponse`

NewPasswordListResponseWithDefaults instantiates a new PasswordListResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEntries

`func (o *PasswordListResponse) GetEntries() []Password`

GetEntries returns the Entries field if non-nil, zero value otherwise.

### GetEntriesOk

`func (o *PasswordListResponse) GetEntriesOk() (*[]Password, bool)`

GetEntriesOk returns a tuple with the Entries field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEntries

`func (o *PasswordListResponse) SetEntries(v []Password)`

SetEntries sets Entries field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


