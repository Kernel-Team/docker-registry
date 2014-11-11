package factory

import (
	"fmt"

	"github.com/docker/docker-registry/storagedriver"
	"github.com/docker/docker-registry/storagedriver/ipc"
)

// driverFactories stores an internal mapping between storage driver names and their respective
// factories
var driverFactories = make(map[string]StorageDriverFactory)

// StorageDriverFactory is a factory interface for creating storagedriver.StorageDriver interfaces
// Storage drivers should call Register() with a factory to make the driver available by name
type StorageDriverFactory interface {
	// Create returns a new storagedriver.StorageDriver with the given parameters
	// Parameters will vary by driver and may be ignored
	// Each parameter key must only consist of lowercase letters and numbers
	Create(parameters map[string]string) (storagedriver.StorageDriver, error)
}

// Register makes a storage driver available by the provided name.
// If Register is called twice with the same name or if driver factory is nil, it panics.
func Register(name string, factory StorageDriverFactory) {
	if factory == nil {
		panic("Must not provide nil StorageDriverFactory")
	}
	_, registered := driverFactories[name]
	if registered {
		panic(fmt.Sprintf("StorageDriverFactory named %s already registered", name))
	}

	driverFactories[name] = factory
}

// Create a new storagedriver.StorageDriver with the given name and parameters
// To run in-process, the StorageDriverFactory must first be registered with the given name
// If no in-process drivers are found with the given name, this attempts to create an IPC driver
// If no in-process or external drivers are found, an InvalidStorageDriverError is returned
func Create(name string, parameters map[string]string) (storagedriver.StorageDriver, error) {
	driverFactory, ok := driverFactories[name]
	if !ok {
		// No registered StorageDriverFactory found, try ipc
		driverClient, err := ipc.NewDriverClient(name, parameters)
		if err != nil {
			return nil, InvalidStorageDriverError{name}
		}
		err = driverClient.Start()
		if err != nil {
			return nil, err
		}
		return driverClient, nil
	}
	return driverFactory.Create(parameters)
}

// InvalidStorageDriverError records an attempt to construct an unregistered storage driver
type InvalidStorageDriverError struct {
	Name string
}

func (err InvalidStorageDriverError) Error() string {
	return fmt.Sprintf("StorageDriver not registered: %s", err.Name)
}