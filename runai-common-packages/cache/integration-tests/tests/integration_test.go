//go:build integration
// +build integration

/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/run-ai/runai/runai-common-packages/cache/v2"
	"github.com/run-ai/runai/runai-common-packages/cache/v2/options"
	"github.com/stretchr/testify/assert"
)

func setUpIntegrationTest(t *testing.T) {
	setEnvs(t)

	if err := cache.Init(nil); err != nil {
		fmt.Printf("Error when initializing cache: %v", err)
		os.Exit(1)
	}
}

func tearDownIntegrationTest() {
	if err := cache.Close(); err != nil {
		fmt.Printf("Error when closing cache: %v", err)
		os.Exit(1)
	}
}

func setEnvs(t *testing.T) {
	url := getEnvOrDefault("CACHE_URL", "localhost:4222")
	readUrl := getEnvOrDefault("CACHE_READ_URL", "localhost:4223")
	cacheType := getEnvOrDefault("CACHE_TYPE", "nats")

	t.Setenv("SERVICE_NAME", "test-service")
	t.Setenv("CACHE_TYPE", cacheType)
	t.Setenv("CACHE_URL", url)
	t.Setenv("CACHE_READ_URL", readUrl)
	t.Setenv("CACHE_ENABLED", "true")
	t.Setenv("SECRETS_MOUNT_PATH", ".")
}

type testType struct {
}
type unhashableType struct {
	Slice []int
}

type testStruct struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type nameToIdKey string
type BiDirectionalMapping struct {
	NameToId map[nameToIdKey]*int32
	IdToName map[int32]*string
}

func newNameToIdKey(clusterId string, name string) nameToIdKey {
	return nameToIdKey(fmt.Sprintf("%s-%s", clusterId, name))
}

func TestRegisterStore(t *testing.T) {
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()

	err := cache.RegisterStore[testType]()
	assert.NoError(t, err)

	err = cache.Set[testType](context.Background(), "test-key", testType{})
	assert.NoError(t, err)

	err = cache.RegisterStore[testStruct]()
	assert.NoError(t, err)

	err = cache.Set[testStruct](context.Background(), "test-key", testStruct{
		Name: "test",
	})
	assert.NoError(t, err)

	res1, err := cache.Get[testType](context.Background(), "test-key")
	assert.NoError(t, err)
	assert.Equal(t, *res1, testType{})

	res2, err := cache.Get[testStruct](context.Background(), "test-key")
	assert.NoError(t, err)
	assert.Equal(t, *res2, testStruct{Name: "test"})
}

func TestRegisterStore_Failed(t *testing.T) {
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()

	err := cache.RegisterStore[*testType]()
	assert.EqualError(t, err, "only structs are accepted as store types")

	err = cache.RegisterStore[map[string]string]()
	assert.EqualError(t, err, "only structs are accepted as store types")

	err = cache.RegisterStore[string]()
	assert.EqualError(t, err, "only structs are accepted as store types")
}

func TestRegisterStore_InvalidName(t *testing.T) {
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()
	err := cache.RegisterSharedStore[testType]("name.with.dots", options.ReadOnly())
	assert.EqualError(t, err, "store name contains invalid characters")
}

func TestRegisterSharedStore_InvalidName(t *testing.T) {
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()
	err := cache.RegisterSharedStore[testType]("name_with#", options.ReadOnly())
	assert.EqualError(t, err, "store name contains invalid characters")
}

func TestSharedRegisterStore_SameStoreDifferentNamesCreation(t *testing.T) {
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()

	err := cache.RegisterSharedStore[testType]("users")
	assert.NoError(t, err)

	err = cache.Set[testType](context.Background(), "test-key", testType{})
	assert.NoError(t, err)

	err = cache.RegisterSharedStore[testType]("tenants")
	assert.EqualError(t, err, "store of type tests.testType already registered, can not be registered again with other name, current shared-users-tests_testType new shared-tenants-tests_testType")

	res, err := cache.Get[testType](context.Background(), "test-key")
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestSharedRegisterStore_MultipleStoresWithSameName(t *testing.T) {
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()

	err := cache.RegisterSharedStore[testType]("test")
	assert.NoError(t, err)

	err = cache.Set[testType](context.Background(), "test-key", testType{})
	assert.NoError(t, err)

	err = cache.RegisterSharedStore[testStruct]("test")
	assert.NoError(t, err)

	err = cache.Set[testStruct](context.Background(), "test-key", testStruct{
		Name: "test",
	})
	assert.NoError(t, err)

	res1, err := cache.Get[testType](context.Background(), "test-key")
	assert.NoError(t, err)
	assert.NotNil(t, res1)

	res2, err := cache.Get[testStruct](context.Background(), "test-key")
	assert.NoError(t, err)
	assert.Equal(t, *res2, testStruct{Name: "test"})
}

func TestRegisterStore_CantRegisterReadOnly(t *testing.T) {
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()

	err := cache.RegisterStore[testType](options.ReadOnly())
	assert.EqualError(t, err, "read only is only allowed for shared stores")
}

func TestSetAndThenGet_UnhashableType(t *testing.T) {
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()

	key := "test-unhashable"
	val := unhashableType{Slice: []int{1, 2, 3}}
	err := cache.Set[unhashableType](context.Background(), key, val)
	assert.NoError(t, err)

	res, err := cache.Get[unhashableType](context.Background(), key)
	assert.NoError(t, err)
	assert.Equal(t, val, *res)
}

func TestSetAndThenGet_Struct_BadDecode_DeleteKey(t *testing.T) {
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()

	key := "test-struct"
	user := testStruct{Id: 1001, Name: "name"}

	err := cache.Set[testStruct](context.Background(), key, user)
	assert.NoError(t, err)

	_, err = cache.Get[string](context.Background(), key)
	assert.Error(t, err)

	res, err := cache.Get[string](context.Background(), key)
	assert.NoError(t, err)
	assert.Nil(t, res)
}

func TestSetAndThenGet_String(t *testing.T) {
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()

	key := "test-string"
	val := "hello world"

	err := cache.Set[string](context.Background(), key, val)
	assert.NoError(t, err)

	res, err := cache.Get[string](context.Background(), key)
	assert.NoError(t, err)
	assert.Equal(t, val, *res)
}

func TestSetAndThenGet_BiDirectionalMapping(t *testing.T) {
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()
	val1 := int32(1)
	name := "test"
	key := "test-bi-directional-mapping"
	val := BiDirectionalMapping{
		NameToId: map[nameToIdKey]*int32{
			newNameToIdKey("1", name): &val1,
		},
		IdToName: map[int32]*string{
			1: &name,
		},
	}

	err := cache.Set[BiDirectionalMapping](context.Background(), key, val)
	assert.NoError(t, err)

	res, err := cache.Get[BiDirectionalMapping](context.Background(), key)
	assert.NoError(t, err)
	assert.Equal(t, val, *res)
}

func TestSetAndThenGet_Int(t *testing.T) {
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()

	ctx := context.Background()

	key := "test-int"
	val := 42

	err := cache.Set[int](ctx, key, val)
	assert.NoError(t, err)

	res, err := cache.Get[int](ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, val, *res)
}

func TestSetAndThenGet_Bool(t *testing.T) {
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()

	key := "test-bool"
	expected := true

	cache.Set(context.Background(), key, &expected)
	val, err := cache.Get[bool](context.Background(), key)

	assert.NoError(t, err)
	assert.Equal(t, expected, *val)
}

func TestSetAndThenGet_Struct(t *testing.T) {
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()

	key := "test-struct"
	user := testStruct{Id: 1001, Name: "name"}

	err := cache.Set[testStruct](context.Background(), key, user)
	assert.NoError(t, err)

	res, err := cache.Get[testStruct](context.Background(), key)

	assert.NoError(t, err)
	assert.Equal(t, user.Id, res.Id)
	assert.Equal(t, user.Name, res.Name)
}

func TestGetMissingKey(t *testing.T) {
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()

	key := "test-missing"

	val, err := cache.Get[string](context.Background(), key)

	assert.Nil(t, err)
	assert.Nil(t, val)
}

func TestDel(t *testing.T) {
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()

	key := "test-key"
	val := "hello world"

	err := cache.Set[string](context.Background(), key, val)
	assert.NoError(t, err)

	err = cache.Del[string](context.Background(), key)
	assert.NoError(t, err)

	res1, err := cache.Get[string](context.Background(), key)
	assert.Nil(t, res1)
	assert.NoError(t, err)
}

func TestSet_ReadOnly(t *testing.T) {
	setUpIntegrationTest(t)
	err := cache.RegisterSharedStore[testType]("name", options.ReadOnly())
	assert.NoError(t, err)
	defer tearDownIntegrationTest()

	err = cache.Set[testType](context.Background(), "test-key", testType{})
	assert.EqualError(t, err, "store is read only")
	res1, err := cache.Get[testType](context.Background(), "test-key")
	assert.Nil(t, res1)
	assert.NoError(t, err)
}

func TestDel_ReadOnly(t *testing.T) {
	setUpIntegrationTest(t)
	err := cache.RegisterSharedStore[testType]("name", options.ReadOnly())
	assert.NoError(t, err)
	defer tearDownIntegrationTest()

	err = cache.Del[testType](context.Background(), "test-key")
	assert.EqualError(t, err, "store is read only")
}

func TestSet_ExceedMaxRecordSize(t *testing.T) {
	t.Setenv("CACHE_MAX_RECORD_SIZE_KB", "1")
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()

	str := strings.Repeat("a", 10*1000)
	err := cache.Set[string](context.Background(), "test-key", str)
	assert.EqualError(t, err, "record size 10.010 KB exceeds max allowed size of 1 KB")
	res1, err := cache.Get[string](context.Background(), "test-key")
	assert.Nil(t, res1)
	assert.NoError(t, err)
}

func TestSet_ExceedMaxTotalSize(t *testing.T) {
	if os.Getenv("CACHE_TYPE") == "inmemory" {
		t.Skip("inmemory cache has eviction policy")
	}
	t.Setenv("CACHE_MAX_STORE_SIZE_MB", "20")
	t.Setenv("CACHE_MAX_RECORD_SIZE_KB", "200")
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()

	cache.ClearCache()
	str := strings.Repeat("a", 120*1000) // 150kb

	var i = 0
	var err error
	for err == nil {
		i++
		err = cache.Set[string](context.Background(), fmt.Sprintf("test-key-%d", i), str)
	}
	err = cache.Set[string](context.Background(), fmt.Sprintf("test-key-%d", i), str)
	assert.ErrorContains(t, err, "exceed")

	res, err := cache.Get[string](context.Background(), fmt.Sprintf("test-key-%d", i-1))
	assert.NoError(t, err)
	assert.NotNil(t, res)
	res, err = cache.Get[string](context.Background(), fmt.Sprintf("test-key-%d", i))
	assert.NoError(t, err)
	assert.Nil(t, res)
}

func TestTTL(t *testing.T) {
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()

	err := cache.RegisterStore[testType](options.Ttl(time.Second))
	assert.NoError(t, err)

	err = cache.Set[testType](context.Background(), "test-key", testType{})
	assert.NoError(t, err)
	res, err := cache.Get[testType](context.Background(), "test-key")
	assert.NoError(t, err)
	assert.NotNil(t, res)

	time.Sleep(100*time.Millisecond + time.Second)
	res, err = cache.Get[testType](context.Background(), "test-key")
	assert.NoError(t, err)
	assert.Nil(t, res)
}

func TestClearCache(t *testing.T) {
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()

	err := cache.RegisterStore[testType]()
	assert.NoError(t, err)

	err = cache.Set[testType](context.Background(), "test-key", testType{})
	assert.NoError(t, err)
	res, err := cache.Get[testType](context.Background(), "test-key")
	assert.NoError(t, err)
	assert.NotNil(t, res)

	err = cache.ClearCache()
	assert.NoError(t, err)

	res, err = cache.Get[testType](context.Background(), "test-key")
	assert.NoError(t, err)
	assert.Nil(t, res)
}

func getEnvOrDefault(key, defaultVal string) string {
	if value, ok := os.LookupEnv(key); ok {
		return strings.Trim(value, " ")
	}
	return defaultVal
}
