package field_types

// +k8s:deepcopy-gen=true
type S3bucketField struct {
	Rules     *MapRules                        `json:"rules,omitempty"`
	ItemRules *S3bucketItemRules               `json:"itemRules,omitempty"`
	Value     *map[string]S3bucketItemAndRules `json:"items,omitempty"`
}

// +k8s:deepcopy-gen=true
type S3bucketItemAndRules struct {
	Rules   *MapItemRules      `json:"rules,omitempty"`
	Value   *S3bucketItemValue `json:"value,omitempty"`
	Deleted *bool              `json:"deleted,omitempty"`
}

// +k8s:deepcopy-gen=true
type S3bucketItemRules struct {
	Bucket *StringRules `json:"bucket"`
	Path   *StringRules `json:"path"`
	Url    *StringRules `json:"url"`

	SecretName *StringRules `json:"secretName"`
	Key        *StringRules `json:"key"`
	Secret     *StringRules `json:"secret"`

	Image *StringRules `json:"image,omitempty"`
}

// +k8s:deepcopy-gen=true
type S3bucketItemValue struct {
	//+kubebuilder:validation:Required
	Bucket *string `json:"bucket"`
	// +kubebuilder:validation:Required
	Path *string `json:"path"`
	//  url of the S3 server. default: https://s3.amazon.com
	Url *string `json:"url,omitempty"`

	// optional name of kubernetes secret that holds your S3 access key ID and S3 secret access.
	SecretName *string `json:"secretName,omitempty"`
	// if secretName is provided, this field should contain the key, within the provided kubernetes secret, which holds the
	// value of your S3 access key ID. otherwise, this field should specify your S3 access key ID in plain text
	// (example: AKIAIOSFODNN7EXAMPLE).
	Key *string `json:"key,omitempty"`
	// if secretName is provided, this field should contain the key, within the provided kubernetes secret, which holds the
	// value of your S3 secret access key. otherwise, this field should specify your S3 secret access key in plain text
	// (example: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY).
	Secret *string `json:"secret,omitempty"`

	// An AWS session token. It should be provided (in addition to the access key ID and
	// secret access key) to authenticate to a bucket that is configured to use AWS
	// Security Token Service (STS).
	SessionToken *string `json:"sessionToken,omitempty"`

	// this field is reserved for internal purposes.
	Image *string `json:"image,omitempty"`
}

func (self *S3bucketField) Deref() map[string]S3bucketItemAndRules {
	if self != nil && self.Value != nil {
		return *self.Value
	} else {
		return map[string]S3bucketItemAndRules{}
	}
}

func NewS3bucketField(s3BucketMap map[string]S3bucketItemAndRules) *S3bucketField {
	if len(s3BucketMap) == 0 {
		return nil
	}
	return &S3bucketField{
		Value: &s3BucketMap,
	}
}
