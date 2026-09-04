# BootConfigSupport

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Url** | Pointer to **string** | Support destination. Omitted when the deployment does not provide a support link. | [optional] 
**Label** | **string** |  | 

## Methods

### NewBootConfigSupport

`func NewBootConfigSupport(label string, ) *BootConfigSupport`

NewBootConfigSupport instantiates a new BootConfigSupport object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewBootConfigSupportWithDefaults

`func NewBootConfigSupportWithDefaults() *BootConfigSupport`

NewBootConfigSupportWithDefaults instantiates a new BootConfigSupport object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUrl

`func (o *BootConfigSupport) GetUrl() string`

GetUrl returns the Url field if non-nil, zero value otherwise.

### GetUrlOk

`func (o *BootConfigSupport) GetUrlOk() (*string, bool)`

GetUrlOk returns a tuple with the Url field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUrl

`func (o *BootConfigSupport) SetUrl(v string)`

SetUrl sets Url field to given value.

### HasUrl

`func (o *BootConfigSupport) HasUrl() bool`

HasUrl returns a boolean if a field has been set.

### GetLabel

`func (o *BootConfigSupport) GetLabel() string`

GetLabel returns the Label field if non-nil, zero value otherwise.

### GetLabelOk

`func (o *BootConfigSupport) GetLabelOk() (*string, bool)`

GetLabelOk returns a tuple with the Label field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabel

`func (o *BootConfigSupport) SetLabel(v string)`

SetLabel sets Label field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


