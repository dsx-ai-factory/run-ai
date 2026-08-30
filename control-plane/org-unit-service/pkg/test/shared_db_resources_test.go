package testing

import (
	"testing"
)

func TestSharedCleanup(t *testing.T) {
	// Reset state for testing
	activeTests = 0

	// Simulate multiple tests starting
	RegisterTest()
	RegisterTest()
	RegisterTest()

	// Verify that we have 3 active tests
	if activeTests != 3 {
		t.Errorf("Expected 3 active tests, got %d", activeTests)
	}

	// Simulate first test finishing
	UnregisterTest()

	// Verify that we still have 2 active tests
	if activeTests != 2 {
		t.Errorf("Expected 2 active tests, got %d", activeTests)
	}

	// Simulate second test finishing
	UnregisterTest()

	// Verify that we still have 1 active test
	if activeTests != 1 {
		t.Errorf("Expected 1 active test, got %d", activeTests)
	}

	// Simulate last test finishing
	UnregisterTest()

	// Verify that we have 0 active tests
	if activeTests != 0 {
		t.Errorf("Expected 0 active tests, got %d", activeTests)
	}

	// Verify that calling UnregisterTest again doesn't cause issues
	UnregisterTest()

	// Verify that we still have 0 active tests (shouldn't go negative)
	if activeTests != 0 {
		t.Errorf("Expected 0 active tests, got %d", activeTests)
	}
}
