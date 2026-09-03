# AccessKeyAdministration

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** |  | 
**Id** | **string** |  | 
**ClientId** | **string** |  | 
**CreatedBy** | **string** |  | 
**CreatedAt** | **NullableTime** |  | 
**LastLogin** | **NullableTime** |  | 

## Methods

### NewAccessKeyAdministration

`func NewAccessKeyAdministration(name string, id string, clientId string, createdBy string, createdAt NullableTime, lastLogin NullableTime, ) *AccessKeyAdministration`

NewAccessKeyAdministration instantiates a new AccessKeyAdministration object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAccessKeyAdministrationWithDefaults

`func NewAccessKeyAdministrationWithDefaults() *AccessKeyAdministration`

NewAccessKeyAdministrationWithDefaults instantiates a new AccessKeyAdministration object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *AccessKeyAdministration) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *AccessKeyAdministration) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *AccessKeyAdministration) SetName(v string)`

SetName sets Name field to given value.


### GetId

`func (o *AccessKeyAdministration) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *AccessKeyAdministration) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *AccessKeyAdministration) SetId(v string)`

SetId sets Id field to given value.


### GetClientId

`func (o *AccessKeyAdministration) GetClientId() string`

GetClientId returns the ClientId field if non-nil, zero value otherwise.

### GetClientIdOk

`func (o *AccessKeyAdministration) GetClientIdOk() (*string, bool)`

GetClientIdOk returns a tuple with the ClientId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClientId

`func (o *AccessKeyAdministration) SetClientId(v string)`

SetClientId sets ClientId field to given value.


### GetCreatedBy

`func (o *AccessKeyAdministration) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *AccessKeyAdministration) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *AccessKeyAdministration) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.


### GetCreatedAt

`func (o *AccessKeyAdministration) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *AccessKeyAdministration) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *AccessKeyAdministration) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### SetCreatedAtNil

`func (o *AccessKeyAdministration) SetCreatedAtNil(b bool)`

 SetCreatedAtNil sets the value for CreatedAt to be an explicit nil

### UnsetCreatedAt
`func (o *AccessKeyAdministration) UnsetCreatedAt()`

UnsetCreatedAt ensures that no value is present for CreatedAt, not even an explicit nil
### GetLastLogin

`func (o *AccessKeyAdministration) GetLastLogin() time.Time`

GetLastLogin returns the LastLogin field if non-nil, zero value otherwise.

### GetLastLoginOk

`func (o *AccessKeyAdministration) GetLastLoginOk() (*time.Time, bool)`

GetLastLoginOk returns a tuple with the LastLogin field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastLogin

`func (o *AccessKeyAdministration) SetLastLogin(v time.Time)`

SetLastLogin sets LastLogin field to given value.


### SetLastLoginNil

`func (o *AccessKeyAdministration) SetLastLoginNil(b bool)`

 SetLastLoginNil sets the value for LastLogin to be an explicit nil

### UnsetLastLogin
`func (o *AccessKeyAdministration) UnsetLastLogin()`

UnsetLastLogin ensures that no value is present for LastLogin, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


