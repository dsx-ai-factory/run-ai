/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package reports

// Uses minIO third party S3 play server

//var (
//	csvS3Saver S3Saver
//	bucketName string
//	reportId   = uuid.NewString()
//)
//
//func TestMain(m *testing.M) {
//	s3_client.Init(s3_client.ConstructTestConfig())
//	bucketName = fmt.Sprintf("%s-%s", "test", uuid.NewString())
//	client, _ := s3_client.Client()
//	client.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{
//		Region: "us-east-1",
//	})
//
//	csvS3Saver = NewCSVS3Saver(bucketName)
//
//	code := m.Run()
//
//	client.RemoveObject(context.Background(), bucketName, reportId, minio.RemoveObjectOptions{
//		ForceDelete: true,
//	})
//	client.RemoveBucket(context.Background(), bucketName)
//	os.Exit(code)
//}
//
//func TestSaveCSVFile_Success(t *testing.T) {
//	err := csvS3Saver.Save(
//		context.Background(),
//		"1",
//		reportId,
//		map[string][]string{
//			"a,b,c": {"1", "1"},
//			"a,b,d": {"2", "2"},
//			"a,e,f": {"4", "4"},
//		},
//		[]string{"poject", "department", "nodepool"}, // N = 3
//		[]string{"GPUs", "CPUs"},                     // M = 2
//	)
//	assert.NoError(t, err)
//
//	client, _ := s3_client.Client()
//	obj, err := client.GetObject(context.Background(), bucketName, reportId, minio.GetObjectOptions{})
//	assert.NoError(t, err)
//	assert.NotNil(t, obj)
//
//	info, _ := obj.Stat()
//	assert.NotEqual(t, int64(0), info.Size)
//}
