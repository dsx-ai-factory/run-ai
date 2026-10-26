# ImagePullSecretsRulesAttributes

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**UserCredentials** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 

## Methods

### NewImagePullSecretsRulesAttributes

`func NewImagePullSecretsRulesAttributes() *ImagePullSecretsRulesAttributes`

NewImagePullSecretsRulesAttributes instantiates a new ImagePullSecretsRulesAttributes object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewImagePullSecretsRulesAttributesWithDefaults

`func NewImagePullSecretsRulesAttributesWithDefaults() *ImagePullSecretsRulesAttributes`

NewImagePullSecretsRulesAttributesWithDefaults instantiates a new ImagePullSecretsRulesAttributes object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *ImagePullSecretsRulesAttributes) GetName() StringRules`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ImagePullSecretsRulesAttributes) GetNameOk() (*StringRules, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ImagePullSecretsRulesAttributes) SetName(v StringRules)`

SetName sets Name field to given value.

### HasName

`func (o *ImagePullSecretsRulesAttributes) HasName() bool`

HasName returns a boolean if a field has been set.

### SetNameNil

`func (o *ImagePullSecretsRulesAttributes) SetNameNil(b bool)`

 SetNameNil sets the value for Name to be an explicit nil

### UnsetName
`func (o *ImagePullSecretsRulesAttributes) UnsetName()`

UnsetName ensures that no value is present for Name, not even an explicit nil
### GetUserCredentials

`func (o *ImagePullSecretsRulesAttributes) GetUserCredentials() BooleanRules`

GetUserCredentials returns the UserCredentials field if non-nil, zero value otherwise.

### GetUserCredentialsOk

`func (o *ImagePullSecretsRulesAttributes) GetUserCredentialsOk() (*BooleanRules, bool)`

GetUserCredentialsOk returns a tuple with the UserCredentials field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserCredentials

`func (o *ImagePullSecretsRulesAttributes) SetUserCredentials(v BooleanRules)`

SetUserCredentials sets UserCredentials field to given value.

### HasUserCredentials

`func (o *ImagePullSecretsRulesAttributes) HasUserCredentials() bool`

HasUserCredentials returns a boolean if a field has been set.

### SetUserCredentialsNil

`func (o *ImagePullSecretsRulesAttributes) SetUserCredentialsNil(b bool)`

 SetUserCredentialsNil sets the value for UserCredentials to be an explicit nil

### UnsetUserCredentials
`func (o *ImagePullSecretsRulesAttributes) UnsetUserCredentials()`

UnsetUserCredentials ensures that no value is present for UserCredentials, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


