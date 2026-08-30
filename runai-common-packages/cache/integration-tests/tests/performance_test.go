//go:build integration
// +build integration

package tests

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/run-ai/runai/runai-common-packages/cache/v2"
	"github.com/stretchr/testify/assert"
)

type performanceTestType struct {
	ID string `json:"id"`
}

func setUpBenchmarkTest(b *testing.B) {
	b.StopTimer()
	url := getEnvOrDefault("CACHE_URL", "localhost:6379")
	readUrl := getEnvOrDefault("CACHE_READ_URL", "localhost:6379")
	cacheType := getEnvOrDefault("CACHE_TYPE", "redis")

	b.Setenv("SERVICE_NAME", "test-service")
	b.Setenv("CACHE_TYPE", cacheType)
	b.Setenv("CACHE_URL", url)
	b.Setenv("CACHE_READ_URL", readUrl)
	b.Setenv("CACHE_ENABLED", "true")
	b.Setenv("SECRETS_MOUNT_PATH", ".")

	if err := cache.Init(nil); err != nil {
		fmt.Printf("Error when initializing cache: %v", err)
		os.Exit(1)
	}
	b.StartTimer()

}

func tearDownBenchmarkTest(b *testing.B) {
	if err := cache.Close(); err != nil {
		fmt.Printf("Error when closing cache: %v", err)
		os.Exit(1)
	}
	b.StartTimer()
}

func BenchmarkRegisterStore(b *testing.B) {
	setUpBenchmarkTest(b)
	defer tearDownBenchmarkTest(b)
	for i := 0; i < b.N; i++ {
		cache.RegisterStore[performanceTestType]()
	}
}

func BenchmarkGet(b *testing.B) {
	setUpBenchmarkTest(b)
	defer tearDownBenchmarkTest(b)
	b.StopTimer()
	err := cache.RegisterStore[performanceTestType]()
	assert.NoError(b, err)
	cache.Set(context.Background(), "test-key", performanceTestType{ID: "test-id"})
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cache.Get[performanceTestType](context.Background(), "test-key")
	}
}

func BenchmarkSet(b *testing.B) {
	setUpBenchmarkTest(b)
	defer tearDownBenchmarkTest(b)
	b.StopTimer()
	err := cache.RegisterStore[performanceTestType]()
	assert.NoError(b, err)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(context.Background(), "test-key", performanceTestType{ID: "test-id"})
	}
}

func BenchmarkDelete(b *testing.B) {
	setUpBenchmarkTest(b)
	defer tearDownBenchmarkTest(b)
	b.StopTimer()
	err := cache.RegisterStore[performanceTestType]()
	assert.NoError(b, err)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		cache.Set(context.Background(), "test-key", performanceTestType{ID: "test-id"})
		b.StartTimer()
		cache.Del[performanceTestType](context.Background(), "test-key")
	}
}

func BenchmarkClearStore(b *testing.B) {
	setUpBenchmarkTest(b)
	defer tearDownBenchmarkTest(b)
	b.StopTimer()
	err := cache.RegisterStore[performanceTestType]()
	assert.NoError(b, err)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		cache.Set(context.Background(), "test-key", performanceTestType{ID: "test-id"})
		b.StartTimer()
		cache.ClearStore[performanceTestType]()
	}
}
