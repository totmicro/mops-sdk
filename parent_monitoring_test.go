package sdk

import (
	"testing"
	"time"
)

func TestSimpleParentMonitoring(t *testing.T) {
	// Reset the health check time to current for testing
	lastHealthCheck = time.Now()
	
	// Test initial health check time is recent
	initialTime := lastHealthCheck
	if time.Since(initialTime) > 1*time.Second {
		t.Errorf("Initial lastHealthCheck should be recent, but was %v ago", time.Since(initialTime))
	}

	// Test that we can update the health check time
	time.Sleep(10 * time.Millisecond) // Small delay to ensure time difference
	
	// Simulate a GetInfo call by creating a server and calling GetInfo
	server := &PluginRPCServer{Impl: &MockPluginImpl{}}
	var resp PluginInfoRPC
	
	err := server.GetInfo(nil, &resp)
	if err != nil {
		t.Fatalf("GetInfo should not return error: %v", err)
	}

	// Verify the health check was updated
	if !lastHealthCheck.After(initialTime) {
		t.Errorf("lastHealthCheck should be updated after GetInfo call")
	}

	// Verify the timeout constant is reasonable
	if parentTimeoutDuration != 30*time.Second {
		t.Errorf("Expected parentTimeoutDuration to be 30s, got %v", parentTimeoutDuration)
	}
}

func TestParentMonitoringTimeout(t *testing.T) {
	// Test that the timeout logic would work
	oldTime := lastHealthCheck
	lastHealthCheck = time.Now().Add(-31 * time.Second) // Set to 31 seconds ago
	
	// Check if timeout condition would be triggered
	if time.Since(lastHealthCheck) <= parentTimeoutDuration {
		t.Errorf("Expected timeout condition to be true for old timestamp")
	}
	
	// Restore original time
	lastHealthCheck = oldTime
}

func TestGetParentMonitoringStatus(t *testing.T) {
	// Reset timestamp for clean test
	lastHealthCheck = time.Now()
	
	// Test that GetParentMonitoringStatus returns expected structure
	status := GetParentMonitoringStatus()
	
	// Verify required fields exist
	expectedFields := []string{
		"monitoring_enabled",
		"last_health_check", 
		"time_since_last_check",
		"timeout_threshold",
		"parent_alive",
	}
	
	for _, field := range expectedFields {
		if _, exists := status[field]; !exists {
			t.Errorf("Expected field '%s' not found in monitoring status", field)
		}
	}
	
	// Test that parent is considered alive initially
	if !status["parent_alive"].(bool) {
		t.Error("Parent should be considered alive initially")
	}
	
	// Test with old timestamp - parent should be considered dead
	oldTime := time.Now().Add(-35 * time.Second)
	lastHealthCheck = oldTime
	
	status = GetParentMonitoringStatus()
	if status["parent_alive"].(bool) {
		t.Error("Parent should be considered dead after timeout")
	}
	
	// Verify monitoring is always enabled
	if !status["monitoring_enabled"].(bool) {
		t.Error("Monitoring should be enabled")
	}
}