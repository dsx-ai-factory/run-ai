# ClusterReportedCustomCA

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Enabled** | Pointer to **NullableBool** | indicates if a custom CA bundle is configured for this consumer | [optional] 
**SecretName** | Pointer to **NullableString** | the name of the Secret holding the custom CA bundle | [optional] 
**SecretKey** | Pointer to **NullableString** | the key within the Secret that holds the custom CA bundle | [optional] 
**Optional** | Pointer to **NullableBool** | indicates if the Secret is optional (the consumer should tolerate a missing Secret) | [optional] 

## Methods

### NewClusterReportedCustomCA

`func NewClusterReportedCustomCA() *ClusterReportedCustomCA`

NewClusterReportedCustomCA instantiates a new ClusterReportedCustomCA object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClusterReportedCustomCAWithDefaults

`func NewClusterReportedCustomCAWithDefaults() *ClusterReportedCustomCA`

NewClusterReportedCustomCAWithDefaults instantiates a new ClusterReportedCustomCA object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEnabled

`func (o *ClusterReportedCustomCA) GetEnabled() bool`

GetEnabled returns the Enabled field if non-nil, zero value otherwise.

### GetEnabledOk

`func (o *ClusterReportedCustomCA) GetEnabledOk() (*bool, bool)`

GetEnabledOk returns a tuple with the Enabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabled

`func (o *ClusterReportedCustomCA) SetEnabled(v bool)`

SetEnabled sets Enabled field to given value.

### HasEnabled

`func (o *ClusterReportedCustomCA) HasEnabled() bool`

HasEnabled returns a boolean if a field has been set.

### SetEnabledNil

`func (o *ClusterReportedCustomCA) SetEnabledNil(b bool)`

 SetEnabledNil sets the value for Enabled to be an explicit nil

### UnsetEnabled
`func (o *ClusterReportedCustomCA) UnsetEnabled()`

UnsetEnabled ensures that no value is present for Enabled, not even an explicit nil
### GetSecretName

`func (o *ClusterReportedCustomCA) GetSecretName() string`

GetSecretName returns the SecretName field if non-nil, zero value otherwise.

### GetSecretNameOk

`func (o *ClusterReportedCustomCA) GetSecretNameOk() (*string, bool)`

GetSecretNameOk returns a tuple with the SecretName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecretName

`func (o *ClusterReportedCustomCA) SetSecretName(v string)`

SetSecretName sets SecretName field to given value.

### HasSecretName

`func (o *ClusterReportedCustomCA) HasSecretName() bool`

HasSecretName returns a boolean if a field has been set.

### SetSecretNameNil

`func (o *ClusterReportedCustomCA) SetSecretNameNil(b bool)`

 SetSecretNameNil sets the value for SecretName to be an explicit nil

### UnsetSecretName
`func (o *ClusterReportedCustomCA) UnsetSecretName()`

UnsetSecretName ensures that no value is present for SecretName, not even an explicit nil
### GetSecretKey

`func (o *ClusterReportedCustomCA) GetSecretKey() string`

GetSecretKey returns the SecretKey field if non-nil, zero value otherwise.

### GetSecretKeyOk

`func (o *ClusterReportedCustomCA) GetSecretKeyOk() (*string, bool)`

GetSecretKeyOk returns a tuple with the SecretKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecretKey

`func (o *ClusterReportedCustomCA) SetSecretKey(v string)`

SetSecretKey sets SecretKey field to given value.

### HasSecretKey

`func (o *ClusterReportedCustomCA) HasSecretKey() bool`

HasSecretKey returns a boolean if a field has been set.

### SetSecretKeyNil

`func (o *ClusterReportedCustomCA) SetSecretKeyNil(b bool)`

 SetSecretKeyNil sets the value for SecretKey to be an explicit nil

### UnsetSecretKey
`func (o *ClusterReportedCustomCA) UnsetSecretKey()`

UnsetSecretKey ensures that no value is present for SecretKey, not even an explicit nil
### GetOptional

`func (o *ClusterReportedCustomCA) GetOptional() bool`

GetOptional returns the Optional field if non-nil, zero value otherwise.

### GetOptionalOk

`func (o *ClusterReportedCustomCA) GetOptionalOk() (*bool, bool)`

GetOptionalOk returns a tuple with the Optional field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOptional

`func (o *ClusterReportedCustomCA) SetOptional(v bool)`

SetOptional sets Optional field to given value.

### HasOptional

`func (o *ClusterReportedCustomCA) HasOptional() bool`

HasOptional returns a boolean if a field has been set.

### SetOptionalNil

`func (o *ClusterReportedCustomCA) SetOptionalNil(b bool)`

 SetOptionalNil sets the value for Optional to be an explicit nil

### UnsetOptional
`func (o *ClusterReportedCustomCA) UnsetOptional()`

UnsetOptional ensures that no value is present for Optional, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


