# DockerRegistryToAddOrUpdate

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The name of the asset. | 
**Spec** | [**DockerRegistrySpec**](DockerRegistrySpec.md) |  | 
**Checksum** | **string** | The checksum of the asset&#39;s spec. | 

## Methods

### NewDockerRegistryToAddOrUpdate

`func NewDockerRegistryToAddOrUpdate(name string, spec DockerRegistrySpec, checksum string, ) *DockerRegistryToAddOrUpdate`

NewDockerRegistryToAddOrUpdate instantiates a new DockerRegistryToAddOrUpdate object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDockerRegistryToAddOrUpdateWithDefaults

`func NewDockerRegistryToAddOrUpdateWithDefaults() *DockerRegistryToAddOrUpdate`

NewDockerRegistryToAddOrUpdateWithDefaults instantiates a new DockerRegistryToAddOrUpdate object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *DockerRegistryToAddOrUpdate) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *DockerRegistryToAddOrUpdate) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *DockerRegistryToAddOrUpdate) SetName(v string)`

SetName sets Name field to given value.


### GetSpec

`func (o *DockerRegistryToAddOrUpdate) GetSpec() DockerRegistrySpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *DockerRegistryToAddOrUpdate) GetSpecOk() (*DockerRegistrySpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *DockerRegistryToAddOrUpdate) SetSpec(v DockerRegistrySpec)`

SetSpec sets Spec field to given value.


### GetChecksum

`func (o *DockerRegistryToAddOrUpdate) GetChecksum() string`

GetChecksum returns the Checksum field if non-nil, zero value otherwise.

### GetChecksumOk

`func (o *DockerRegistryToAddOrUpdate) GetChecksumOk() (*string, bool)`

GetChecksumOk returns a tuple with the Checksum field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChecksum

`func (o *DockerRegistryToAddOrUpdate) SetChecksum(v string)`

SetChecksum sets Checksum field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


