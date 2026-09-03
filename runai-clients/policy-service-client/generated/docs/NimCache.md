# NimCache

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **NullableString** | The NIMCache resource name (mandatory). | [optional] 
**Profile** | Pointer to **NullableString** | The NIM profile to use (optional). | [optional] 

## Methods

### NewNimCache

`func NewNimCache() *NimCache`

NewNimCache instantiates a new NimCache object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNimCacheWithDefaults

`func NewNimCacheWithDefaults() *NimCache`

NewNimCacheWithDefaults instantiates a new NimCache object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *NimCache) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *NimCache) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *NimCache) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *NimCache) HasName() bool`

HasName returns a boolean if a field has been set.

### SetNameNil

`func (o *NimCache) SetNameNil(b bool)`

 SetNameNil sets the value for Name to be an explicit nil

### UnsetName
`func (o *NimCache) UnsetName()`

UnsetName ensures that no value is present for Name, not even an explicit nil
### GetProfile

`func (o *NimCache) GetProfile() string`

GetProfile returns the Profile field if non-nil, zero value otherwise.

### GetProfileOk

`func (o *NimCache) GetProfileOk() (*string, bool)`

GetProfileOk returns a tuple with the Profile field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProfile

`func (o *NimCache) SetProfile(v string)`

SetProfile sets Profile field to given value.

### HasProfile

`func (o *NimCache) HasProfile() bool`

HasProfile returns a boolean if a field has been set.

### SetProfileNil

`func (o *NimCache) SetProfileNil(b bool)`

 SetProfileNil sets the value for Profile to be an explicit nil

### UnsetProfile
`func (o *NimCache) UnsetProfile()`

UnsetProfile ensures that no value is present for Profile, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


