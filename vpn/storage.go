package vpn

import (
	"errors"
	"fmt"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// ██████╗ ███████╗██████╗  ██████╗ ███████╗
// ██╔══██╗██╔════╝██╔══██╗██╔═══██╗██╔════╝
// ██████╔╝█████╗  ██████╔╝██║   ██║███████╗
// ██╔══██╗██╔══╝  ██╔═══╝ ██║   ██║╚════██║
// ██║  ██║███████╗██║     ╚██████╔╝███████║
// ╚═╝  ╚═╝╚══════╝╚═╝      ╚═════╝ ╚══════╝

// ServerConfigRepository is an interface for storing and retrieving server configurations
type ServerConfigRepository interface {
	Get() (*ServerConfig, error)
	Persist(*ServerConfig) error
}

// PeerConfigRepository is an interface for storing and retrieving client configurations
type PeerConfigRepository interface {
	GetClient(userId string, publicKey wgtypes.Key) (*ClientConfig, error)
	Count(userId string) (int, error)
	AddClient(userId string, client *ClientConfig) error
	RemoveClient(userId string, publicKey wgtypes.Key) error
	GetAllClients() ([]*ClientConfig, error)
	ListClients(userId string) ([]*ClientConfig, error)
}

var (
	ErrStorageError        = errors.New("storage error")
	ErrCannotListClients   = fmt.Errorf("%w : Cannot list clients", ErrStorageError)
	ErrCannotGetClient     = fmt.Errorf("%w : Cannot get client", ErrStorageError)
	ErrCannotCountClients  = fmt.Errorf("%w : Cannot count clients", ErrStorageError)
	ErrCannotAddClient     = fmt.Errorf("%w : Cannot add client", ErrStorageError)
	ErrCannotRemoveClient  = fmt.Errorf("%w : Cannot remove client", ErrStorageError)
	ErrCannotGetAllClients = fmt.Errorf("%w : Cannot get all clients", ErrStorageError)
	ErrCannotPersistConfig = fmt.Errorf("%w : Cannot persist config", ErrStorageError)
)
