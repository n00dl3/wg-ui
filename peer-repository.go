package main

import (
	"encoding/json"
	"errors"

	"github.com/embarkstudios/wireguard-ui/vpn"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

var (
	ErrClientNotFound = errors.New("client not found")
	ErrUserNotFound   = errors.New("user not found")
)

// ██╗   ██╗███████╗███████╗██████╗
// ██║   ██║██╔════╝██╔════╝██╔══██╗
// ██║   ██║███████╗█████╗  ██████╔╝
// ██║   ██║╚════██║██╔══╝  ██╔══██╗
// ╚██████╔╝███████║███████╗██║  ██║
// ╚═════╝ ╚══════╝╚══════╝╚═╝  ╚═╝

// UserConfig represents a user and its clients
type UserConfig map[wgtypes.Key]*vpn.ClientConfig

// Get returns a client configuration by its public key
// @return *vpn.ClientConfig
// @return error
func (u UserConfig) Get(key wgtypes.Key) (*vpn.ClientConfig, error) {
	if _, ok := u[key]; !ok {
		return nil, ErrClientNotFound
	}
	return u[key], nil
}

// Add a client configuration to the current user
func (u UserConfig) Add(client *vpn.ClientConfig) {
	u[client.PublicKey] = client
}

// Remove a client configuration by its public key
func (u UserConfig) Remove(key wgtypes.Key) error {
	if _, ok := u[key]; !ok {
		return ErrClientNotFound
	}
	delete(u, key)
	return nil
}

// Count returns the number of clients for the current user
func (u UserConfig) Count() int {
	return len(u)
}

// List all clients for the current user
func (u UserConfig) List() []*vpn.ClientConfig {
	clients := make([]*vpn.ClientConfig, 0, len(u))
	for _, client := range u {
		clients = append(clients, client)
	}
	return clients
}

// MergeWith merges the current UserConfig with another UserConfig
func (u UserConfig) MergeWith(config UserConfig) {
	for key, client := range config {
		if _, ok := u[key]; !ok {
			u[key] = client
		} else {
			u[key].MergeWith(*client)
		}
	}
}

func (u UserConfig) MarshalJSON() ([]byte, error) {
	clients := make(map[string]*vpn.ClientConfig)
	for key, client := range u {
		clients[key.String()] = client
	}
	return json.Marshal(clients)
}

func (u *UserConfig) UnmarshalJSON(data []byte) error {
	clients := make(map[string]*vpn.ClientConfig)
	if err := json.Unmarshal(data, &clients); err != nil {
		return err
	}
	user := make(UserConfig)
	for key, client := range clients {
		k, err := wgtypes.ParseKey(key)
		if err != nil {
			return err
		}
		user[k] = client
	}
	*u = user
	return nil
}

type UserConfigRepository map[string]UserConfig

func (r UserConfigRepository) Get(userId string) (UserConfig, error) {
	user, ok := r[userId]
	if !ok {
		r[userId] = make(UserConfig)
		user = r[userId]
	}
	return user, nil
}

func (r UserConfigRepository) GetClient(userId string, publicKey wgtypes.Key) (*vpn.ClientConfig, error) {
	var err error
	var user UserConfig
	if user, err = r.Get(userId); err != nil {
		return nil, err
	}
	return user.Get(publicKey)
}
func (r UserConfigRepository) Count(userId string) (int, error) {
	var err error
	var user UserConfig
	if user, err = r.Get(userId); err != nil {
		return 0, err
	}
	return user.Count(), nil
}

func (r UserConfigRepository) AddClient(userId string, client *vpn.ClientConfig) error {
	var err error
	var user UserConfig
	if user, err = r.Get(userId); err != nil {
		return err
	}
	user.Add(client)
	return nil
}

func (r UserConfigRepository) RemoveClient(userId string, publicKey wgtypes.Key) error {
	var err error
	var user UserConfig
	if user, err = r.Get(userId); err != nil {
		return err
	}
	return user.Remove(publicKey)
}
func (r UserConfigRepository) GetAllClients() ([]*vpn.ClientConfig, error) {
	var clients []*vpn.ClientConfig
	for _, user := range r {
		for _, client := range user {
			clients = append(clients, client)
		}
	}
	return clients, nil
}
func (r UserConfigRepository) ListClients(userId string) ([]*vpn.ClientConfig, error) {
	var err error
	var user UserConfig
	if user, err = r.Get(userId); err != nil {
		return nil, err
	}
	return user.List(), nil
}

func newPeerRepository() UserConfigRepository {
	return make(UserConfigRepository)
}
