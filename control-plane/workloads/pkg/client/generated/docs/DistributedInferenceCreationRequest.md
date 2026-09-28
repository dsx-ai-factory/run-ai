# DistributedInferenceCreationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The name of the workload. | 
**UseGivenNameAsPrefix** | Pointer to **bool** | When true, the requested name will be treated as a prefix. The final name of the workload will be composed of the name followed by a random set of characters. | [optional] [default to false]
**ProjectId** | **string** | The id of the project. | 
**ClusterId** | **string** | The id of the cluster. | 
**TemplateId** | Pointer to **NullableString** | The unique identifier of the template to use for submitting this workload. The combined values provided in the spec, template and assets will be used to create the workload. | [optional] 
**Assets** | Pointer to [**NullableDistributedInferenceAssetsAssets**](DistributedInferenceAssetsAssets.md) |  | [optional] 
**Spec** | Pointer to [**DistributedInferenceSpecSpec**](DistributedInferenceSpecSpec.md) |  | [optional] 

## Methods

### NewDistributedInferenceCreationRequest

`func NewDistributedInferenceCreationRequest(name string, projectId string, clusterId string, ) *DistributedInferenceCreationRequest`

NewDistributedInferenceCreationRequest instantiates a new DistributedInferenceCreationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedInferenceCreationRequestWithDefaults

`func NewDistributedInferenceCreationRequestWithDefaults() *DistributedInferenceCreationRequest`

NewDistributedInferenceCreationRequestWithDefaults instantiates a new DistributedInferenceCreationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *DistributedInferenceCreationRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *DistributedInferenceCreationRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *DistributedInferenceCreationRequest) SetName(v string)`

SetName sets Name field to given value.


### GetUseGivenNameAsPrefix

`func (o *DistributedInferenceCreationRequest) GetUseGivenNameAsPrefix() bool`

GetUseGivenNameAsPrefix returns the UseGivenNameAsPrefix field if non-nil, zero value otherwise.

### GetUseGivenNameAsPrefixOk

`func (o *DistributedInferenceCreationRequest) GetUseGivenNameAsPrefixOk() (*bool, bool)`

GetUseGivenNameAsPrefixOk returns a tuple with the UseGivenNameAsPrefix field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUseGivenNameAsPrefix

`func (o *DistributedInferenceCreationRequest) SetUseGivenNameAsPrefix(v bool)`

SetUseGivenNameAsPrefix sets UseGivenNameAsPrefix field to given value.

### HasUseGivenNameAsPrefix

`func (o *DistributedInferenceCreationRequest) HasUseGivenNameAsPrefix() bool`

HasUseGivenNameAsPrefix returns a boolean if a field has been set.

### GetProjectId

`func (o *DistributedInferenceCreationRequest) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *DistributedInferenceCreationRequest) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *DistributedInferenceCreationRequest) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.


### GetClusterId

`func (o *DistributedInferenceCreationRequest) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *DistributedInferenceCreationRequest) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *DistributedInferenceCreationRequest) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetTemplateId

`func (o *DistributedInferenceCreationRequest) GetTemplateId() string`

GetTemplateId returns the TemplateId field if non-nil, zero value otherwise.

### GetTemplateIdOk

`func (o *DistributedInferenceCreationRequest) GetTemplateIdOk() (*string, bool)`

GetTemplateIdOk returns a tuple with the TemplateId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTemplateId

`func (o *DistributedInferenceCreationRequest) SetTemplateId(v string)`

SetTemplateId sets TemplateId field to given value.

### HasTemplateId

`func (o *DistributedInferenceCreationRequest) HasTemplateId() bool`

HasTemplateId returns a boolean if a field has been set.

### SetTemplateIdNil

`func (o *DistributedInferenceCreationRequest) SetTemplateIdNil(b bool)`

 SetTemplateIdNil sets the value for TemplateId to be an explicit nil

### UnsetTemplateId
`func (o *DistributedInferenceCreationRequest) UnsetTemplateId()`

UnsetTemplateId ensures that no value is present for TemplateId, not even an explicit nil
### GetAssets

`func (o *DistributedInferenceCreationRequest) GetAssets() DistributedInferenceAssetsAssets`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *DistributedInferenceCreationRequest) GetAssetsOk() (*DistributedInferenceAssetsAssets, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *DistributedInferenceCreationRequest) SetAssets(v DistributedInferenceAssetsAssets)`

SetAssets sets Assets field to given value.

### HasAssets

`func (o *DistributedInferenceCreationRequest) HasAssets() bool`

HasAssets returns a boolean if a field has been set.

### SetAssetsNil

`func (o *DistributedInferenceCreationRequest) SetAssetsNil(b bool)`

 SetAssetsNil sets the value for Assets to be an explicit nil

### UnsetAssets
`func (o *DistributedInferenceCreationRequest) UnsetAssets()`

UnsetAssets ensures that no value is present for Assets, not even an explicit nil
### GetSpec

`func (o *DistributedInferenceCreationRequest) GetSpec() DistributedInferenceSpecSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *DistributedInferenceCreationRequest) GetSpecOk() (*DistributedInferenceSpecSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *DistributedInferenceCreationRequest) SetSpec(v DistributedInferenceSpecSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *DistributedInferenceCreationRequest) HasSpec() bool`

HasSpec returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


