// Copyright 2019 Ka-Hing Cheung
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package internal

import (
	. "github.com/kahing/goofys/api/common"

	"context"
	"crypto/fips140"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/jacobsa/fuse"
)

// GCS variant of S3
type GCS3 struct {
	*S3Backend
}

type GCS3MultipartBlobCommitInput struct {
	Size uint64
	ETag *string
	Prev *MultipartBlobAddInput
}

func NewGCS3(bucket string, flags *FlagStorage, config *S3Config) (*GCS3, error) {
	s3Backend, err := NewS3(bucket, flags, config)
	if err != nil {
		return nil, err
	}
	s3Backend.Capabilities().Name = "gcs3"
	s := &GCS3{S3Backend: s3Backend}
	s.S3Backend.gcs = true
	s.S3Backend.cap.NoParallelMultipart = true
	return s, nil
}

func (s *GCS3) Delegate() interface{} {
	return s
}

func (s *GCS3) DeleteBlobs(param *DeleteBlobsInput) (*DeleteBlobsOutput, error) {
	// GCS does not have multi-delete
	var wg sync.WaitGroup
	var overallErr error

	for _, key := range param.Items {
		wg.Add(1)
		go func(key string) {
			_, err := s.DeleteBlob(&DeleteBlobInput{
				Key: key,
			})
			if err != nil && err != fuse.ENOENT {
				overallErr = err
			}
			wg.Done()
		}(key)
	}
	wg.Wait()
	if overallErr != nil {
		return nil, mapAwsError(overallErr)
	}

	return &DeleteBlobsOutput{}, nil
}

func (s *GCS3) MultipartBlobBegin(param *MultipartBlobBeginInput) (*MultipartBlobCommitInput, error) {
	mpu := s3.CreateMultipartUploadInput{
		Bucket:       &s.bucket,
		Key:          &param.Key,
		StorageClass: &s.config.StorageClass,
		ContentType:  param.ContentType,
	}

	if s.config.UseSSE {
		mpu.ServerSideEncryption = &s.sseType
		if s.config.UseKMS && s.config.KMSKeyID != "" {
			mpu.SSEKMSKeyId = &s.config.KMSKeyID
		}
	}

	if s.config.ACL != "" {
		mpu.ACL = &s.config.ACL
	}

	req, _ := s.CreateMultipartUploadRequest(&mpu)

	var location string
	if fips140.Enabled() {
		// SigV2 (HMAC-SHA1) is banned under FIPS, and AWS SigV4 makes GCS reject
		// the resumable initiate: SigV4 stamps x-amz-* onto the request and GCS
		// forbids mixing x-amz and x-goog headers (400 ExcessHeaderValues), which
		// the mandatory x-goog-resumable header triggers. Sign it with GCS's own
		// GOOG4-HMAC-SHA256 instead: x-goog-* only, HMAC-SHA256 (FIPS-approved).
		//
		// Build() resolves the {Bucket}/{Key} URL template (normally done inside
		// Send, which we bypass); then drop the ?uploads= query for the initiate.
		if err := req.Build(); err != nil {
			s3Log.Errorf("CreateMultipartUpload %v = %v", param.Key, err)
			return nil, mapAwsError(err)
		}
		req.HTTPRequest.URL.RawQuery = ""
		loc, err := s.goog4ResumableStart(req.HTTPRequest.URL)
		if err != nil {
			s3Log.Errorf("CreateMultipartUpload %v = %v", param.Key, err)
			return nil, err
		}
		location = loc
	} else {
		// get rid of ?uploads=
		req.HTTPRequest.URL.RawQuery = ""
		req.HTTPRequest.Header.Set("x-goog-resumable", "start")
		// v4 signing of this fails
		s.setV2Signer(&req.Handlers)
		err := req.Send()
		if err != nil {
			s3Log.Errorf("CreateMultipartUpload %v = %v", param.Key, err)
			return nil, mapAwsError(err)
		}
		location = req.HTTPResponse.Header.Get("Location")
	}

	if _, err := url.Parse(location); err != nil {
		s3Log.Errorf("CreateMultipartUpload %v %v = %v", param.Key, location, err)
		return nil, mapAwsError(err)
	}

	return &MultipartBlobCommitInput{
		Key:         &param.Key,
		Metadata:    param.Metadata,
		UploadId:    &location,
		Parts:       make([]*string, 10000), // at most 10K parts
		backendData: &GCS3MultipartBlobCommitInput{},
	}, nil
}

const (
	goog4Algorithm   = "GOOG4-HMAC-SHA256"
	goog4Region      = "auto"
	goog4Service     = "storage"
	goog4Request     = "goog4_request"
	goog4TimeFormat  = "20060102T150405Z"
	goog4DateFormat  = "20060102"
	emptyPayloadHash = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
)

// goog4ResumableStart initiates a GCS resumable upload signed with GCS's native
// GOOG4-HMAC-SHA256 scheme (x-goog-* headers only, HMAC-SHA256). Unlike AWS
// SigV4 it emits no x-amz-* headers, so GCS does not reject it for mixing header
// families. Returns the resumable session URI from the Location header.
func (s *GCS3) goog4ResumableStart(u *url.URL) (string, error) {
	// Use the client's resolved credentials (session/env chain), not
	// s.config.Credentials, which is nil when keys come from the environment.
	creds, err := s.Config.Credentials.Get()
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, u.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("x-goog-resumable", "start")
	if s.config.StorageClass != "" {
		req.Header.Set("x-goog-storage-class", s.config.StorageClass)
	}

	signGoog4(req, creds.AccessKeyID, creds.SecretAccessKey, goog4Region, time.Now())

	// Reuse the SDK's configured client (flags.HTTPTimeout + shared transport)
	// rather than http.DefaultClient, which has no timeout.
	client := s.Config.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("resumable upload start failed: %s: %s", resp.Status, string(body))
	}

	return resp.Header.Get("Location"), nil
}

// signGoog4 adds a GOOG4-HMAC-SHA256 Authorization header to req, following
// GCS's V4 signing process (the Google-native equivalent of AWS SigV4).
func signGoog4(req *http.Request, accessKey, secretKey, region string, t time.Time) {
	t = t.UTC()
	timestamp := t.Format(goog4TimeFormat)
	datestamp := t.Format(goog4DateFormat)

	req.Header.Set("x-goog-date", timestamp)
	req.Header.Set("x-goog-content-sha256", emptyPayloadHash)

	// Canonical/signed headers: host plus every x-goog-* header, lowercased and sorted.
	canonHeaders := map[string]string{"host": req.URL.Host}
	for k, v := range req.Header {
		lk := strings.ToLower(k)
		if strings.HasPrefix(lk, "x-goog-") {
			canonHeaders[lk] = strings.TrimSpace(strings.Join(v, ","))
		}
	}
	names := make([]string, 0, len(canonHeaders))
	for k := range canonHeaders {
		names = append(names, k)
	}
	sort.Strings(names)

	var headerBlock strings.Builder
	for _, n := range names {
		headerBlock.WriteString(n)
		headerBlock.WriteString(":")
		headerBlock.WriteString(canonHeaders[n])
		headerBlock.WriteString("\n")
	}
	signedHeaders := strings.Join(names, ";")

	canonicalRequest := strings.Join([]string{
		req.Method,
		req.URL.EscapedPath(),
		req.URL.RawQuery,
		headerBlock.String(),
		signedHeaders,
		emptyPayloadHash,
	}, "\n")

	scope := datestamp + "/" + region + "/" + goog4Service + "/" + goog4Request
	stringToSign := strings.Join([]string{
		goog4Algorithm,
		timestamp,
		scope,
		hexSHA256(canonicalRequest),
	}, "\n")

	kDate := hmacSHA256([]byte("GOOG4"+secretKey), datestamp)
	kRegion := hmacSHA256(kDate, region)
	kService := hmacSHA256(kRegion, goog4Service)
	kSigning := hmacSHA256(kService, goog4Request)
	signature := hex.EncodeToString(hmacSHA256(kSigning, stringToSign))

	req.Header.Set("Authorization", fmt.Sprintf("%s Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		goog4Algorithm, accessKey, scope, signedHeaders, signature))
}

func hmacSHA256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return h.Sum(nil)
}

func hexSHA256(data string) string {
	sum := sha256.Sum256([]byte(data))
	return hex.EncodeToString(sum[:])
}

func (s *GCS3) uploadPart(param *MultipartBlobAddInput, totalSize uint64, last bool) (etag *string, err error) {
	atomic.AddUint32(&param.Commit.NumParts, 1)

	if closer, ok := param.Body.(io.Closer); ok {
		defer closer.Close()
	}

	// the mpuId serves as authentication token so
	// technically we don't need to sign this anymore and
	// can just use a plain HTTP request, but going
	// through aws-sdk-go anyway to get retry handling
	params := &s3.PutObjectInput{
		Bucket: &s.bucket,
		Key:    param.Commit.Key,
		Body:   param.Body,
	}

	s3Log.Debug(params)

	req, resp := s.PutObjectRequest(params)
	req.Handlers.Sign.Clear()
	req.HTTPRequest.URL, _ = url.Parse(*param.Commit.UploadId)

	start := totalSize - param.Size
	end := totalSize - 1
	var size string
	if last {
		size = strconv.FormatUint(totalSize, 10)
	} else {
		size = "*"
	}

	contentRange := fmt.Sprintf("bytes %v-%v/%v", start, end, size)

	req.HTTPRequest.Header.Set("Content-Length", strconv.FormatUint(param.Size, 10))
	req.HTTPRequest.Header.Set("Content-Range", contentRange)

	err = req.Send()
	if err != nil {
		// status indicating that we need more parts to finish this
		if req.HTTPResponse.StatusCode == 308 {
			err = nil
		} else {
			err = mapAwsError(err)
			return
		}
	}

	etag = resp.ETag

	return
}

func (s *GCS3) MultipartBlobAdd(param *MultipartBlobAddInput) (*MultipartBlobAddOutput, error) {
	var commitData *GCS3MultipartBlobCommitInput
	var ok bool
	if commitData, ok = param.Commit.backendData.(*GCS3MultipartBlobCommitInput); !ok {
		panic("Incorrect commit data type")
	}

	if commitData.Prev != nil {
		if commitData.Prev.Size == 0 || commitData.Prev.Size%(256*1024) != 0 {
			s3Log.Errorf("size of each block must be multiple of 256KB: %v", param.Size)
			return nil, fuse.EINVAL
		}

		_, err := s.uploadPart(commitData.Prev, commitData.Size, false)
		if err != nil {
			return nil, err
		}
	}
	commitData.Size += param.Size

	copy := *param
	commitData.Prev = &copy
	param.Body = nil

	return &MultipartBlobAddOutput{}, nil
}

func (s *GCS3) MultipartBlobCommit(param *MultipartBlobCommitInput) (*MultipartBlobCommitOutput, error) {
	var commitData *GCS3MultipartBlobCommitInput
	var ok bool
	if commitData, ok = param.backendData.(*GCS3MultipartBlobCommitInput); !ok {
		panic("Incorrect commit data type")
	}

	if commitData.Prev == nil {
		panic("commit should include last part")
	}

	etag, err := s.uploadPart(commitData.Prev, commitData.Size, true)
	if err != nil {
		return nil, err
	}

	return &MultipartBlobCommitOutput{
		ETag: etag,
	}, nil
}
