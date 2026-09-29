# RelatedUrlsDefaults

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Attributes** | Pointer to [**NullableRelatedUrl**](RelatedUrl.md) |  | [optional] 
**Instances** | Pointer to [**[]RelatedUrl**](RelatedUrl.md) | Set of URLs that are related to the workload. | [optional] 

## Methods

### NewRelatedUrlsDefaults

`func NewRelatedUrlsDefaults() *RelatedUrlsDefaults`

NewRelatedUrlsDefaults instantiates a new RelatedUrlsDefaults object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRelatedUrlsDefaultsWithDefaults

`func NewRelatedUrlsDefaultsWithDefaults() *RelatedUrlsDefaults`

NewRelatedUrlsDefaultsWithDefaults instantiates a new RelatedUrlsDefaults object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAttributes

`func (o *RelatedUrlsDefaults) GetAttributes() RelatedUrl`

GetAttributes returns the Attributes field if non-nil, zero value otherwise.

### GetAttributesOk

`func (o *RelatedUrlsDefaults) GetAttributesOk() (*RelatedUrl, bool)`

GetAttributesOk returns a tuple with the Attributes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAttributes

`func (o *RelatedUrlsDefaults) SetAttributes(v RelatedUrl)`

SetAttributes sets Attributes field to given value.

### HasAttributes

`func (o *RelatedUrlsDefaults) HasAttributes() bool`

HasAttributes returns a boolean if a field has been set.

### SetAttributesNil

`func (o *RelatedUrlsDefaults) SetAttributesNil(b bool)`

 SetAttributesNil sets the value for Attributes to be an explicit nil

### UnsetAttributes
`func (o *RelatedUrlsDefaults) UnsetAttributes()`

UnsetAttributes ensures that no value is present for Attributes, not even an explicit nil
### GetInstances

`func (o *RelatedUrlsDefaults) GetInstances() []RelatedUrl`

GetInstances returns the Instances field if non-nil, zero value otherwise.

### GetInstancesOk

`func (o *RelatedUrlsDefaults) GetInstancesOk() (*[]RelatedUrl, bool)`

GetInstancesOk returns a tuple with the Instances field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstances

`func (o *RelatedUrlsDefaults) SetInstances(v []RelatedUrl)`

SetInstances sets Instances field to given value.

### HasInstances

`func (o *RelatedUrlsDefaults) HasInstances() bool`

HasInstances returns a boolean if a field has been set.

### SetInstancesNil

`func (o *RelatedUrlsDefaults) SetInstancesNil(b bool)`

 SetInstancesNil sets the value for Instances to be an explicit nil

### UnsetInstances
`func (o *RelatedUrlsDefaults) UnsetInstances()`

UnsetInstances ensures that no value is present for Instances, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


