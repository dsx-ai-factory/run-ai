# AIApplicationCredentialBinding

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Variable** | **string** | Destination in Helm values. Supports dotted keys, escaped literal characters, array indexes, and double-quoted bracket keys. Destinations must not overlap. global.imagePullSecret.name and global.ngcApiSecret.name also set the sibling create value to false.  | 
**AssetId** | **string** | Selected Run:ai credential asset ID. Never a credential value or a Kubernetes Secret name. | 
**CredentialType** | **string** | ngc-api-key selects an NGC asset&#39;s generic Secret. image-pull-secret selects a registry asset or the registry representation of an NGC asset. generic selects a generic Secret asset or an NGC asset&#39;s generic representation.  | 
**UserCredential** | Pointer to **bool** | Set by the server from the selected asset&#39;s ownership and visibility. | [optional] [readonly] 
**ExistingSecretName** | Pointer to **string** | Set by the server when the authorized credential asset references an existing Kubernetes Secret. | [optional] [readonly] 

## Methods

### NewAIApplicationCredentialBinding

`func NewAIApplicationCredentialBinding(variable string, assetId string, credentialType string, ) *AIApplicationCredentialBinding`

NewAIApplicationCredentialBinding instantiates a new AIApplicationCredentialBinding object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAIApplicationCredentialBindingWithDefaults

`func NewAIApplicationCredentialBindingWithDefaults() *AIApplicationCredentialBinding`

NewAIApplicationCredentialBindingWithDefaults instantiates a new AIApplicationCredentialBinding object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetVariable

`func (o *AIApplicationCredentialBinding) GetVariable() string`

GetVariable returns the Variable field if non-nil, zero value otherwise.

### GetVariableOk

`func (o *AIApplicationCredentialBinding) GetVariableOk() (*string, bool)`

GetVariableOk returns a tuple with the Variable field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVariable

`func (o *AIApplicationCredentialBinding) SetVariable(v string)`

SetVariable sets Variable field to given value.


### GetAssetId

`func (o *AIApplicationCredentialBinding) GetAssetId() string`

GetAssetId returns the AssetId field if non-nil, zero value otherwise.

### GetAssetIdOk

`func (o *AIApplicationCredentialBinding) GetAssetIdOk() (*string, bool)`

GetAssetIdOk returns a tuple with the AssetId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssetId

`func (o *AIApplicationCredentialBinding) SetAssetId(v string)`

SetAssetId sets AssetId field to given value.


### GetCredentialType

`func (o *AIApplicationCredentialBinding) GetCredentialType() string`

GetCredentialType returns the CredentialType field if non-nil, zero value otherwise.

### GetCredentialTypeOk

`func (o *AIApplicationCredentialBinding) GetCredentialTypeOk() (*string, bool)`

GetCredentialTypeOk returns a tuple with the CredentialType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCredentialType

`func (o *AIApplicationCredentialBinding) SetCredentialType(v string)`

SetCredentialType sets CredentialType field to given value.


### GetUserCredential

`func (o *AIApplicationCredentialBinding) GetUserCredential() bool`

GetUserCredential returns the UserCredential field if non-nil, zero value otherwise.

### GetUserCredentialOk

`func (o *AIApplicationCredentialBinding) GetUserCredentialOk() (*bool, bool)`

GetUserCredentialOk returns a tuple with the UserCredential field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserCredential

`func (o *AIApplicationCredentialBinding) SetUserCredential(v bool)`

SetUserCredential sets UserCredential field to given value.

### HasUserCredential

`func (o *AIApplicationCredentialBinding) HasUserCredential() bool`

HasUserCredential returns a boolean if a field has been set.

### GetExistingSecretName

`func (o *AIApplicationCredentialBinding) GetExistingSecretName() string`

GetExistingSecretName returns the ExistingSecretName field if non-nil, zero value otherwise.

### GetExistingSecretNameOk

`func (o *AIApplicationCredentialBinding) GetExistingSecretNameOk() (*string, bool)`

GetExistingSecretNameOk returns a tuple with the ExistingSecretName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExistingSecretName

`func (o *AIApplicationCredentialBinding) SetExistingSecretName(v string)`

SetExistingSecretName sets ExistingSecretName field to given value.

### HasExistingSecretName

`func (o *AIApplicationCredentialBinding) HasExistingSecretName() bool`

HasExistingSecretName returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


