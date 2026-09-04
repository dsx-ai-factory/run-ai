# BulkUserCreationResponseResultsInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Success** | **bool** | Whether the user creation was successful | 
**User** | Pointer to [**UserPostResponse**](UserPostResponse.md) |  | [optional] 
**Error** | Pointer to [**NullableBulkUserCreationResponseResultsInnerError**](BulkUserCreationResponseResultsInnerError.md) |  | [optional] 

## Methods

### NewBulkUserCreationResponseResultsInner

`func NewBulkUserCreationResponseResultsInner(success bool, ) *BulkUserCreationResponseResultsInner`

NewBulkUserCreationResponseResultsInner instantiates a new BulkUserCreationResponseResultsInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewBulkUserCreationResponseResultsInnerWithDefaults

`func NewBulkUserCreationResponseResultsInnerWithDefaults() *BulkUserCreationResponseResultsInner`

NewBulkUserCreationResponseResultsInnerWithDefaults instantiates a new BulkUserCreationResponseResultsInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSuccess

`func (o *BulkUserCreationResponseResultsInner) GetSuccess() bool`

GetSuccess returns the Success field if non-nil, zero value otherwise.

### GetSuccessOk

`func (o *BulkUserCreationResponseResultsInner) GetSuccessOk() (*bool, bool)`

GetSuccessOk returns a tuple with the Success field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSuccess

`func (o *BulkUserCreationResponseResultsInner) SetSuccess(v bool)`

SetSuccess sets Success field to given value.


### GetUser

`func (o *BulkUserCreationResponseResultsInner) GetUser() UserPostResponse`

GetUser returns the User field if non-nil, zero value otherwise.

### GetUserOk

`func (o *BulkUserCreationResponseResultsInner) GetUserOk() (*UserPostResponse, bool)`

GetUserOk returns a tuple with the User field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUser

`func (o *BulkUserCreationResponseResultsInner) SetUser(v UserPostResponse)`

SetUser sets User field to given value.

### HasUser

`func (o *BulkUserCreationResponseResultsInner) HasUser() bool`

HasUser returns a boolean if a field has been set.

### GetError

`func (o *BulkUserCreationResponseResultsInner) GetError() BulkUserCreationResponseResultsInnerError`

GetError returns the Error field if non-nil, zero value otherwise.

### GetErrorOk

`func (o *BulkUserCreationResponseResultsInner) GetErrorOk() (*BulkUserCreationResponseResultsInnerError, bool)`

GetErrorOk returns a tuple with the Error field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetError

`func (o *BulkUserCreationResponseResultsInner) SetError(v BulkUserCreationResponseResultsInnerError)`

SetError sets Error field to given value.

### HasError

`func (o *BulkUserCreationResponseResultsInner) HasError() bool`

HasError returns a boolean if a field has been set.

### SetErrorNil

`func (o *BulkUserCreationResponseResultsInner) SetErrorNil(b bool)`

 SetErrorNil sets the value for Error to be an explicit nil

### UnsetError
`func (o *BulkUserCreationResponseResultsInner) UnsetError()`

UnsetError ensures that no value is present for Error, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


