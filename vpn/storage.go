package vpn

import "errors"

// ██████╗ ███████╗██████╗  ██████╗ ███████╗
// ██╔══██╗██╔════╝██╔══██╗██╔═══██╗██╔════╝
// ██████╔╝█████╗  ██████╔╝██║   ██║███████╗
// ██╔══██╗██╔══╝  ██╔═══╝ ██║   ██║╚════██║
// ██║  ██║███████╗██║     ╚██████╔╝███████║
// ╚═╝  ╚═╝╚══════╝╚═╝      ╚═════╝ ╚══════╝

// ConfigPersister is an interface for storing and retrieving server configurations
type ConfigPersister interface {
	Persist(*ServerConfig) error
}

// ███████╗██████╗ ██████╗  ██████╗ ██████╗ ███████╗
// ██╔════╝██╔══██╗██╔══██╗██╔═══██╗██╔══██╗██╔════╝
// █████╗  ██████╔╝██████╔╝██║   ██║██████╔╝███████╗
// ██╔══╝  ██╔══██╗██╔══██╗██║   ██║██╔══██╗╚════██║
// ███████╗██║  ██║██║  ██║╚██████╔╝██║  ██║███████║
// ╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═╝╚══════╝
var (
	ErrConfigStorageLoad    = errors.New("cannot load server configuration")
	ErrConfigStoragePersist = errors.New("cannot persist server configuration")
)

type StorageError struct {
	Err error
	Aux string
}

func (e *StorageError) Error() string {
	return e.Aux
}

func (e *StorageError) Unwrap() error {
	return e.Err
}

func newStorageError(aux string, err error) *StorageError {
	return &StorageError{
		Err: err,
		Aux: aux,
	}
}
