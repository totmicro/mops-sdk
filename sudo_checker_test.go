package sdk

import (
	"runtime"
	"testing"
)

func TestSudoChecker(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Sudo checker tests are not applicable on Windows")
	}

	checker := NewSudoChecker()

	t.Run("CheckSudoPrivileges", func(t *testing.T) {
		info, err := checker.CheckSudoPrivileges()
		if err != nil {
			t.Fatalf("CheckSudoPrivileges failed: %v", err)
		}

		if info == nil {
			t.Fatal("Expected SudoInfo, got nil")
		}

		// Test that we get some basic information
		if info.CurrentUser == "" {
			t.Error("Expected current user to be populated")
		}

		t.Logf("Current user: %s", info.CurrentUser)
		t.Logf("Is root: %v", info.IsRoot)
		t.Logf("Has sudo: %v", info.HasSudo)
		t.Logf("Sudo path: %s", info.SudoPath)
		t.Logf("Can sudo without password: %v", info.CanSudoWithoutPwd)
		t.Logf("Sudo time left: %d minutes", info.SudoTimeLeft)
		
		if info.ErrorMessage != "" {
			t.Logf("Error message: %s", info.ErrorMessage)
		}
	})

	t.Run("HasSudo", func(t *testing.T) {
		hasSudo := checker.HasSudo()
		t.Logf("Has sudo privileges: %v", hasSudo)
		
		// Just verify the function runs without error
		// Result will vary based on system configuration
	})

	t.Run("GetSudoStatus", func(t *testing.T) {
		status := checker.GetSudoStatus()
		if status == "" {
			t.Error("Expected non-empty status string")
		}
		t.Logf("Sudo status: %s", status)
	})

	t.Run("RequireSudo", func(t *testing.T) {
		err := checker.RequireSudo()
		if err != nil {
			// This is expected in many cases where sudo isn't available
			// or user doesn't have privileges
			t.Logf("RequireSudo returned error (expected in many cases): %v", err)
		} else {
			t.Log("RequireSudo succeeded - user has sudo privileges")
		}
	})
}

func TestSudoCheckerWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific tests")
	}

	checker := NewSudoChecker()

	t.Run("WindowsNotSupported", func(t *testing.T) {
		info, err := checker.CheckSudoPrivileges()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if info.ErrorMessage == "" {
			t.Error("Expected error message for Windows")
		}

		if info.HasSudo {
			t.Error("Expected HasSudo to be false on Windows")
		}

		// Test that HasSudo returns false
		if checker.HasSudo() {
			t.Error("Expected HasSudo() to return false on Windows")
		}

		// Test that RequireSudo returns error
		if err := checker.RequireSudo(); err == nil {
			t.Error("Expected RequireSudo to return error on Windows")
		}
	})
}

func TestSudoCheckerInternalMethods(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix-specific internal method tests")
	}

	checker := NewSudoChecker()

	t.Run("getCurrentUser", func(t *testing.T) {
		user := checker.getCurrentUser()
		if user == "" {
			t.Error("Expected non-empty username")
		}
		t.Logf("Current user: %s", user)
	})

	t.Run("isRunningAsRoot", func(t *testing.T) {
		isRoot := checker.isRunningAsRoot()
		t.Logf("Running as root: %v", isRoot)
		// Just verify the function runs
	})

	t.Run("findSudo", func(t *testing.T) {
		path, found := checker.findSudo()
		t.Logf("Sudo found: %v, path: %s", found, path)
		
		if found && path == "" {
			t.Error("If sudo is found, path should not be empty")
		}
	})

	t.Run("canSudoWithoutPassword", func(t *testing.T) {
		canSudo := checker.canSudoWithoutPassword()
		t.Logf("Can sudo without password: %v", canSudo)
		// Just verify the function runs
	})

	t.Run("getSudoCacheTimeLeft", func(t *testing.T) {
		timeLeft := checker.getSudoCacheTimeLeft()
		t.Logf("Sudo cache time left: %d minutes", timeLeft)
		
		if timeLeft < 0 {
			t.Error("Time left should not be negative")
		}
	})
}