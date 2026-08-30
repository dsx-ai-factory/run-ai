# BootConfigDocs

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**BaseUrl** | **string** | Base URL used to resolve documentation links. | 
**WhatsNew** | [**BootConfigWhatsNew**](BootConfigWhatsNew.md) |  | 

## Methods

### NewBootConfigDocs

`func NewBootConfigDocs(baseUrl string, whatsNew BootConfigWhatsNew, ) *BootConfigDocs`

NewBootConfigDocs instantiates a new BootConfigDocs object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewBootConfigDocsWithDefaults

`func NewBootConfigDocsWithDefaults() *BootConfigDocs`

NewBootConfigDocsWithDefaults instantiates a new BootConfigDocs object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBaseUrl

`func (o *BootConfigDocs) GetBaseUrl() string`

GetBaseUrl returns the BaseUrl field if non-nil, zero value otherwise.

### GetBaseUrlOk

`func (o *BootConfigDocs) GetBaseUrlOk() (*string, bool)`

GetBaseUrlOk returns a tuple with the BaseUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBaseUrl

`func (o *BootConfigDocs) SetBaseUrl(v string)`

SetBaseUrl sets BaseUrl field to given value.


### GetWhatsNew

`func (o *BootConfigDocs) GetWhatsNew() BootConfigWhatsNew`

GetWhatsNew returns the WhatsNew field if non-nil, zero value otherwise.

### GetWhatsNewOk

`func (o *BootConfigDocs) GetWhatsNewOk() (*BootConfigWhatsNew, bool)`

GetWhatsNewOk returns a tuple with the WhatsNew field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWhatsNew

`func (o *BootConfigDocs) SetWhatsNew(v BootConfigWhatsNew)`

SetWhatsNew sets WhatsNew field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


