# ProjectV1NodeAffinityResponseTrain

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AffinityType** | Pointer to **string** | The type of affinity of the jobs on the nodes. | [optional] 
**SelectedTypes** | [**[]ProjectV1NodeAffinityResponseTrainSelectedTypesInner**](ProjectV1NodeAffinityResponseTrainSelectedTypesInner.md) |  | 

## Methods

### NewProjectV1NodeAffinityResponseTrain

`func NewProjectV1NodeAffinityResponseTrain(selectedTypes []ProjectV1NodeAffinityResponseTrainSelectedTypesInner, ) *ProjectV1NodeAffinityResponseTrain`

NewProjectV1NodeAffinityResponseTrain instantiates a new ProjectV1NodeAffinityResponseTrain object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectV1NodeAffinityResponseTrainWithDefaults

`func NewProjectV1NodeAffinityResponseTrainWithDefaults() *ProjectV1NodeAffinityResponseTrain`

NewProjectV1NodeAffinityResponseTrainWithDefaults instantiates a new ProjectV1NodeAffinityResponseTrain object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAffinityType

`func (o *ProjectV1NodeAffinityResponseTrain) GetAffinityType() string`

GetAffinityType returns the AffinityType field if non-nil, zero value otherwise.

### GetAffinityTypeOk

`func (o *ProjectV1NodeAffinityResponseTrain) GetAffinityTypeOk() (*string, bool)`

GetAffinityTypeOk returns a tuple with the AffinityType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAffinityType

`func (o *ProjectV1NodeAffinityResponseTrain) SetAffinityType(v string)`

SetAffinityType sets AffinityType field to given value.

### HasAffinityType

`func (o *ProjectV1NodeAffinityResponseTrain) HasAffinityType() bool`

HasAffinityType returns a boolean if a field has been set.

### GetSelectedTypes

`func (o *ProjectV1NodeAffinityResponseTrain) GetSelectedTypes() []ProjectV1NodeAffinityResponseTrainSelectedTypesInner`

GetSelectedTypes returns the SelectedTypes field if non-nil, zero value otherwise.

### GetSelectedTypesOk

`func (o *ProjectV1NodeAffinityResponseTrain) GetSelectedTypesOk() (*[]ProjectV1NodeAffinityResponseTrainSelectedTypesInner, bool)`

GetSelectedTypesOk returns a tuple with the SelectedTypes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSelectedTypes

`func (o *ProjectV1NodeAffinityResponseTrain) SetSelectedTypes(v []ProjectV1NodeAffinityResponseTrainSelectedTypesInner)`

SetSelectedTypes sets SelectedTypes field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


