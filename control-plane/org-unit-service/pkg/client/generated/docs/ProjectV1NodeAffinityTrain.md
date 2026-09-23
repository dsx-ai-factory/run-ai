# ProjectV1NodeAffinityTrain

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AffinityType** | Pointer to **string** | The type of affinity of the jobs on the nodes. | [optional] 
**SelectedTypes** | Pointer to [**[]ProjectV1NodeAffinityResponseTrainSelectedTypesInner**](ProjectV1NodeAffinityResponseTrainSelectedTypesInner.md) |  | [optional] 

## Methods

### NewProjectV1NodeAffinityTrain

`func NewProjectV1NodeAffinityTrain() *ProjectV1NodeAffinityTrain`

NewProjectV1NodeAffinityTrain instantiates a new ProjectV1NodeAffinityTrain object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectV1NodeAffinityTrainWithDefaults

`func NewProjectV1NodeAffinityTrainWithDefaults() *ProjectV1NodeAffinityTrain`

NewProjectV1NodeAffinityTrainWithDefaults instantiates a new ProjectV1NodeAffinityTrain object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAffinityType

`func (o *ProjectV1NodeAffinityTrain) GetAffinityType() string`

GetAffinityType returns the AffinityType field if non-nil, zero value otherwise.

### GetAffinityTypeOk

`func (o *ProjectV1NodeAffinityTrain) GetAffinityTypeOk() (*string, bool)`

GetAffinityTypeOk returns a tuple with the AffinityType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAffinityType

`func (o *ProjectV1NodeAffinityTrain) SetAffinityType(v string)`

SetAffinityType sets AffinityType field to given value.

### HasAffinityType

`func (o *ProjectV1NodeAffinityTrain) HasAffinityType() bool`

HasAffinityType returns a boolean if a field has been set.

### GetSelectedTypes

`func (o *ProjectV1NodeAffinityTrain) GetSelectedTypes() []ProjectV1NodeAffinityResponseTrainSelectedTypesInner`

GetSelectedTypes returns the SelectedTypes field if non-nil, zero value otherwise.

### GetSelectedTypesOk

`func (o *ProjectV1NodeAffinityTrain) GetSelectedTypesOk() (*[]ProjectV1NodeAffinityResponseTrainSelectedTypesInner, bool)`

GetSelectedTypesOk returns a tuple with the SelectedTypes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSelectedTypes

`func (o *ProjectV1NodeAffinityTrain) SetSelectedTypes(v []ProjectV1NodeAffinityResponseTrainSelectedTypesInner)`

SetSelectedTypes sets SelectedTypes field to given value.

### HasSelectedTypes

`func (o *ProjectV1NodeAffinityTrain) HasSelectedTypes() bool`

HasSelectedTypes returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


