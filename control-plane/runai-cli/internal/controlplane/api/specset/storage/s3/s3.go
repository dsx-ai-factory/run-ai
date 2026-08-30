package s3

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/specset/parser"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

func NewS3List(hostList []string) ([]openapi.S3Instance, error) {
	var s3ObjList = make([]openapi.S3Instance, 0)
	for _, s3Att := range hostList {
		s3Obj, err := NewS3(s3Att)
		if err != nil {
			return nil, err
		}
		s3ObjList = append(s3ObjList, *s3Obj)
	}
	return s3ObjList, nil
}

func NewS3(s3Bucket string) (*openapi.S3Instance, error) {
	s := NewParser(s3Bucket)
	err := s.Parse()
	if err != nil {
		return nil, err
	}

	s3, err := createS3(s)
	if err != nil {
		return nil, err
	}

	if s.Get(AccessKey) != nil {
		s3.SetAccessKeySecret(*s.Get(AccessKey))
	} else if s.Get(Key) != nil && s.Get(Secret) != nil {
		s3.SetSecretKeyOfAccessKeyId(*s.Get(Key))
		s3.SetSecretKeyOfSecretKey(*s.Get(Secret))
	}

	url := s.Get(Url)
	if url != nil {
		s3.SetUrl(*url)
	}
	return s3, nil
}

func createS3(s *parser.AttributeMapper) (*openapi.S3Instance, error) {
	var (
		err    error
		name   *string
		bucket *string
		path   *string
	)
	err = parser.CheckMandatoryFields(s, Bucket, Path)
	if err != nil {
		return nil, err
	}

	name = s.Get(Name)
	bucket = s.Get(Bucket)
	path = s.Get(Path)
	return &openapi.S3Instance{
		Name:   *openapi.NewNullableString(name),
		Bucket: *openapi.NewNullableString(bucket),
		Path:   *openapi.NewNullableString(path),
	}, nil
}
