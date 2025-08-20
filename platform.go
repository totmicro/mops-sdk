package sdk

import (
	"runtime"
)

// GetCurrentPlatform returns the current platform information  
func GetCurrentPlatform() PlatformInfo {
	return PlatformInfo{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}
}
