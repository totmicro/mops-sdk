package sdk

import (
	"fmt"
	"sync"
)

// hostSharedStorageClient is a singleton that provides access to host shared storage
type hostSharedStorageClient struct {
	setVarFunc    func(key, value string) error
	getVarFunc    func(key string) (string, error)
	listVarsFunc  func() ([]string, error)
}

var (
	hostStorageClient *hostSharedStorageClient
	hostStorageMutex  sync.RWMutex
)

// SetHostSharedStorageClient sets the host shared storage client
// This is called by the host during plugin initialization
func SetHostSharedStorageClient(setVar func(string, string) error, getVar func(string) (string, error), listVars func() ([]string, error)) {
	hostStorageMutex.Lock()
	defer hostStorageMutex.Unlock()
	
	hostStorageClient = &hostSharedStorageClient{
		setVarFunc:   setVar,
		getVarFunc:   getVar,
		listVarsFunc: listVars,
	}
	
	GetPluginLogger().Debug("Host shared storage client initialized")
}

// SetSharedVar stores a key-value pair in the host shared storage
func SetSharedVar(key, value string) error {
	hostStorageMutex.RLock()
	defer hostStorageMutex.RUnlock()
	
	if hostStorageClient == nil {
		return fmt.Errorf("host shared storage not available")
	}
	
	GetPluginLogger().Debugf("Setting shared variable: %s = %s", key, value)
	return hostStorageClient.setVarFunc(key, value)
}

// GetSharedVar retrieves a value from the host shared storage
func GetSharedVar(key string) (string, error) {
	hostStorageMutex.RLock()
	defer hostStorageMutex.RUnlock()
	
	if hostStorageClient == nil {
		return "", fmt.Errorf("host shared storage not available")
	}
	
	GetPluginLogger().Debugf("Getting shared variable: %s", key)
	value, err := hostStorageClient.getVarFunc(key)
	if err != nil {
		GetPluginLogger().Debugf("Failed to get shared variable %s: %v", key, err)
		return "", err
	}
	
	GetPluginLogger().Debugf("Retrieved shared variable: %s = %s", key, value)
	return value, nil
}

// ListSharedVars returns all keys in the host shared storage
func ListSharedVars() ([]string, error) {
	hostStorageMutex.RLock()
	defer hostStorageMutex.RUnlock()
	
	if hostStorageClient == nil {
		return nil, fmt.Errorf("host shared storage not available")
	}
	
	GetPluginLogger().Debug("Listing shared variables")
	keys, err := hostStorageClient.listVarsFunc()
	if err != nil {
		GetPluginLogger().Debugf("Failed to list shared variables: %v", err)
		return nil, err
	}
	
	GetPluginLogger().Debugf("Listed %d shared variables", len(keys))
	return keys, nil
}