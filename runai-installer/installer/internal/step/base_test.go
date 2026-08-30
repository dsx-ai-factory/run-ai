package step

import (
	"context"
	"fmt"
	"testing"

	"github.com/run-ai/runai/runai-installer/installer/internal/model"
)

func TestNewRunAIInstallStepImpl(t *testing.T) {
	step := NewRunAIInstallStepImpl("Test Step")

	if step.GetName() != "Test Step" {
		t.Errorf("expected name 'Test Step', got %s", step.GetName())
	}

	if step.GetCachedStatus() != model.NotInstalled {
		t.Errorf("expected status NotInstalled, got %v", step.GetCachedStatus())
	}

	if step.GetError() != nil {
		t.Errorf("expected no error, got %v", step.GetError())
	}
}

func TestRunAIInstallStepImpl_SetStatus(t *testing.T) {
	step := NewRunAIInstallStepImpl("Test Step")

	// Test setting status without error
	status := step.SetStatus(model.Installing, nil)
	if status != model.Installing {
		t.Errorf("expected Installing status, got %v", status)
	}
	if step.GetCachedStatus() != model.Installing {
		t.Errorf("expected Installing status from GetCachedStatus(), got %v", step.GetCachedStatus())
	}
	if step.GetError() != nil {
		t.Errorf("expected no error, got %v", step.GetError())
	}

	// Test setting status with error
	testErr := fmt.Errorf("test error")
	status = step.SetStatus(model.Failed, testErr)
	if status != model.Failed {
		t.Errorf("expected Failed status, got %v", status)
	}
	if step.GetCachedStatus() != model.Failed {
		t.Errorf("expected Failed status from GetCachedStatus(), got %v", step.GetCachedStatus())
	}
	if step.GetError() == nil || step.GetError().Error() != "test error" {
		t.Errorf("expected 'test error', got %v", step.GetError())
	}
}

func TestRunAIInstallStepImpl_IsTerminalStatus(t *testing.T) {
	step := NewRunAIInstallStepImpl("Test Step")

	// Test non-terminal statuses
	step.SetStatus(model.NotInstalled, nil)
	if step.IsTerminalStatus() {
		t.Error("NotInstalled should not be terminal")
	}

	step.SetStatus(model.Installing, nil)
	if step.IsTerminalStatus() {
		t.Error("Installing should not be terminal")
	}

	// Test terminal statuses
	step.SetStatus(model.Failed, nil)
	if !step.IsTerminalStatus() {
		t.Error("Failed should be terminal")
	}

	step.SetStatus(model.Installed, nil)
	if !step.IsTerminalStatus() {
		t.Error("Installed should be terminal")
	}

	step.SetStatus(model.Timeout, nil)
	if !step.IsTerminalStatus() {
		t.Error("Timeout should be terminal")
	}
}

func TestRunAIInstallStepImpl_CheckContextCancellation(t *testing.T) {
	step := NewRunAIInstallStepImpl("Test Step")

	// Test with non-cancelled context
	ctx := context.Background()
	err := step.CheckContextCancellation(ctx)
	if err != nil {
		t.Errorf("expected no error for non-cancelled context, got %v", err)
	}

	// Test with cancelled context
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel the context

	err = step.CheckContextCancellation(cancelledCtx)
	if err == nil {
		t.Error("expected error for cancelled context")
	}

	if step.GetCachedStatus() != model.Failed {
		t.Errorf("expected Failed status after context cancellation, got %v", step.GetCachedStatus())
	}

	if step.GetError() == nil {
		t.Error("expected error to be set after context cancellation")
	}
}

func TestRunAIInstallStepImpl_ConcurrentAccess(t *testing.T) {
	step := NewRunAIInstallStepImpl("Test Step")

	// Test concurrent access to status
	done := make(chan bool, 2)

	// Goroutine 1: Set status repeatedly
	go func() {
		for i := 0; i < 100; i++ {
			step.SetStatus(model.Installing, nil)
		}
		done <- true
	}()

	// Goroutine 2: Get status repeatedly
	go func() {
		for i := 0; i < 100; i++ {
			_ = step.GetCachedStatus()
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// If we get here without a race condition, the test passes
}
