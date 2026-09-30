# RelatedUrl

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Url** | Pointer to **NullableString** | The URL for connecting an external service related to the workload. (mandatory) | [optional] 
**Type** | Pointer to **NullableString** | The type of service that the url provides. For example, wandb (Weights &amp; Biases). (mandatory) | [optional] 
**Name** | Pointer to **NullableString** | Unique name to identify the instance. primarily used for policy locked rules. | [optional] 
**Exclude** | Pointer to **NullableBool** | Use &#39;true&#39; in case the item is defined in defaults of the policy, and you wish to exclude it from the workload. | [optional] [default to false]

## Methods

### NewRelatedUrl

`func NewRelatedUrl() *RelatedUrl`

NewRelatedUrl instantiates a new RelatedUrl object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRelatedUrlWithDefaults

`func NewRelatedUrlWithDefaults() *RelatedUrl`

NewRelatedUrlWithDefaults instantiates a new RelatedUrl object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUrl

`func (o *RelatedUrl) GetUrl() string`

GetUrl returns the Url field if non-nil, zero value otherwise.

### GetUrlOk

`func (o *RelatedUrl) GetUrlOk() (*string, bool)`

GetUrlOk returns a tuple with the Url field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUrl

`func (o *RelatedUrl) SetUrl(v string)`

SetUrl sets Url field to given value.

### HasUrl

`func (o *RelatedUrl) HasUrl() bool`

HasUrl returns a boolean if a field has been set.

### SetUrlNil

`func (o *RelatedUrl) SetUrlNil(b bool)`

 SetUrlNil sets the value for Url to be an explicit nil

### UnsetUrl
`func (o *RelatedUrl) UnsetUrl()`

UnsetUrl ensures that no value is present for Url, not even an explicit nil
### GetType

`func (o *RelatedUrl) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *RelatedUrl) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *RelatedUrl) SetType(v string)`

SetType sets Type field to given value.

### HasType

`func (o *RelatedUrl) HasType() bool`

HasType returns a boolean if a field has been set.

### SetTypeNil

`func (o *RelatedUrl) SetTypeNil(b bool)`

 SetTypeNil sets the value for Type to be an explicit nil

### UnsetType
`func (o *RelatedUrl) UnsetType()`

UnsetType ensures that no value is present for Type, not even an explicit nil
### GetName

`func (o *RelatedUrl) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *RelatedUrl) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *RelatedUrl) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *RelatedUrl) HasName() bool`

HasName returns a boolean if a field has been set.

### SetNameNil

`func (o *RelatedUrl) SetNameNil(b bool)`

 SetNameNil sets the value for Name to be an explicit nil

### UnsetName
`func (o *RelatedUrl) UnsetName()`

UnsetName ensures that no value is present for Name, not even an explicit nil
### GetExclude

`func (o *RelatedUrl) GetExclude() bool`

GetExclude returns the Exclude field if non-nil, zero value otherwise.

### GetExcludeOk

`func (o *RelatedUrl) GetExcludeOk() (*bool, bool)`

GetExcludeOk returns a tuple with the Exclude field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExclude

`func (o *RelatedUrl) SetExclude(v bool)`

SetExclude sets Exclude field to given value.

### HasExclude

`func (o *RelatedUrl) HasExclude() bool`

HasExclude returns a boolean if a field has been set.

### SetExcludeNil

`func (o *RelatedUrl) SetExcludeNil(b bool)`

 SetExcludeNil sets the value for Exclude to be an explicit nil

### UnsetExclude
`func (o *RelatedUrl) UnsetExclude()`

UnsetExclude ensures that no value is present for Exclude, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


