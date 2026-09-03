# User1AllOf

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Email** | Pointer to **string** | Email address of the user. | [optional] 
**LastLogin** | Pointer to **time.Time** | The last time the user logged into the system. | [optional] 

## Methods

### NewUser1AllOf

`func NewUser1AllOf() *User1AllOf`

NewUser1AllOf instantiates a new User1AllOf object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUser1AllOfWithDefaults

`func NewUser1AllOfWithDefaults() *User1AllOf`

NewUser1AllOfWithDefaults instantiates a new User1AllOf object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEmail

`func (o *User1AllOf) GetEmail() string`

GetEmail returns the Email field if non-nil, zero value otherwise.

### GetEmailOk

`func (o *User1AllOf) GetEmailOk() (*string, bool)`

GetEmailOk returns a tuple with the Email field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmail

`func (o *User1AllOf) SetEmail(v string)`

SetEmail sets Email field to given value.

### HasEmail

`func (o *User1AllOf) HasEmail() bool`

HasEmail returns a boolean if a field has been set.

### GetLastLogin

`func (o *User1AllOf) GetLastLogin() time.Time`

GetLastLogin returns the LastLogin field if non-nil, zero value otherwise.

### GetLastLoginOk

`func (o *User1AllOf) GetLastLoginOk() (*time.Time, bool)`

GetLastLoginOk returns a tuple with the LastLogin field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastLogin

`func (o *User1AllOf) SetLastLogin(v time.Time)`

SetLastLogin sets LastLogin field to given value.

### HasLastLogin

`func (o *User1AllOf) HasLastLogin() bool`

HasLastLogin returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


