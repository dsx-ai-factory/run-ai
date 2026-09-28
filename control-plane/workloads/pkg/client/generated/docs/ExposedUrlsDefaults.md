# ExposedUrlsDefaults

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Attributes** | Pointer to [**NullableExposedUrl**](ExposedUrl.md) |  | [optional] 
**Instances** | Pointer to [**[]ExposedUrl**](ExposedUrl.md) | Set of container ports that the workload exposes via URLs. | [optional] 

## Methods

### NewExposedUrlsDefaults

`func NewExposedUrlsDefaults() *ExposedUrlsDefaults`

NewExposedUrlsDefaults instantiates a new ExposedUrlsDefaults object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewExposedUrlsDefaultsWithDefaults

`func NewExposedUrlsDefaultsWithDefaults() *ExposedUrlsDefaults`

NewExposedUrlsDefaultsWithDefaults instantiates a new ExposedUrlsDefaults object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAttributes

`func (o *ExposedUrlsDefaults) GetAttributes() ExposedUrl`

GetAttributes returns the Attributes field if non-nil, zero value otherwise.

### GetAttributesOk

`func (o *ExposedUrlsDefaults) GetAttributesOk() (*ExposedUrl, bool)`

GetAttributesOk returns a tuple with the Attributes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAttributes

`func (o *ExposedUrlsDefaults) SetAttributes(v ExposedUrl)`

SetAttributes sets Attributes field to given value.

### HasAttributes

`func (o *ExposedUrlsDefaults) HasAttributes() bool`

HasAttributes returns a boolean if a field has been set.

### SetAttributesNil

`func (o *ExposedUrlsDefaults) SetAttributesNil(b bool)`

 SetAttributesNil sets the value for Attributes to be an explicit nil

### UnsetAttributes
`func (o *ExposedUrlsDefaults) UnsetAttributes()`

UnsetAttributes ensures that no value is present for Attributes, not even an explicit nil
### GetInstances

`func (o *ExposedUrlsDefaults) GetInstances() []ExposedUrl`

GetInstances returns the Instances field if non-nil, zero value otherwise.

### GetInstancesOk

`func (o *ExposedUrlsDefaults) GetInstancesOk() (*[]ExposedUrl, bool)`

GetInstancesOk returns a tuple with the Instances field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstances

`func (o *ExposedUrlsDefaults) SetInstances(v []ExposedUrl)`

SetInstances sets Instances field to given value.

### HasInstances

`func (o *ExposedUrlsDefaults) HasInstances() bool`

HasInstances returns a boolean if a field has been set.

### SetInstancesNil

`func (o *ExposedUrlsDefaults) SetInstancesNil(b bool)`

 SetInstancesNil sets the value for Instances to be an explicit nil

### UnsetInstances
`func (o *ExposedUrlsDefaults) UnsetInstances()`

UnsetInstances ensures that no value is present for Instances, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


