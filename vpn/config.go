package vpn

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// ███████╗██████╗ ██████╗  ██████╗ ██████╗ ███████╗
// ██╔════╝██╔══██╗██╔══██╗██╔═══██╗██╔══██╗██╔════╝
// █████╗  ██████╔╝██████╔╝██║   ██║██████╔╝███████╗
// ██╔══╝  ██╔══██╗██╔══██╗██║   ██║██╔══██╗╚════██║
// ███████╗██║  ██║██║  ██║╚██████╔╝██║  ██║███████║
// ╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═╝╚══════╝

var (
	ErrCannotGenerateKey = errors.New("cannot generate key")
	ErrClientNotFound    = errors.New("client not found")
	ErrTooManyClients    = errors.New("too many clients")
	ErrRangeExhausted    = errors.New("IP range exhausted")
	ErrValidationError   = errors.New("validation error")
	ErrInvalidMTU        = fmt.Errorf("%e: MTU must be between 1280 and 1500", ErrValidationError)
	ErrInvalidPublicKey  = fmt.Errorf("%e: invalid public key", ErrValidationError)
	ErrInvalidIP         = fmt.Errorf("%e: invalid IP", ErrValidationError)
	ErrInvalidKeepalive  = fmt.Errorf("%e: invalid KeepAlive", ErrValidationError)
)

// ███████╗███████╗██████╗ ██╗   ██╗███████╗██████╗
// ██╔════╝██╔════╝██╔══██╗██║   ██║██╔════╝██╔══██╗
// ███████╗█████╗  ██████╔╝██║   ██║█████╗  ██████╔╝
// ╚════██║██╔══╝  ██╔══██╗╚██╗ ██╔╝██╔══╝  ██╔══██╗
// ███████║███████╗██║  ██║ ╚████╔╝ ███████╗██║  ██║
// ╚══════╝╚══════╝╚═╝  ╚═╝  ╚═══╝  ╚══════╝╚═╝  ╚═╝

// ServerConfig holds the VPN server configuration
type ServerConfig struct {
	wgtypes.Config
	AllowedIPs        []net.IPNet
	Endpoint          net.UDPAddr
	PrivateKey        wgtypes.Key
	LinkConfig        *LinkConfig
	MaxClientsPerUser int
	DefaultPeerMTU    MTU
	Users             map[string]UserConfig
}

func (s *ServerConfig) getUser(userId string) UserConfig {
	if _, ok := s.Users[userId]; !ok {
		s.Users[userId] = make(UserConfig)
	}
	return s.Users[userId]
}

// GetClient returns a user's client configuration by its public key
func (s *ServerConfig) GetClient(userId string, key wgtypes.Key) (*ClientConfig, error) {
	user := s.getUser(userId)
	return user.Get(key)
}

// AddClient adds a client configuration to the provided user
func (s *ServerConfig) AddClient(userId string, allowedIPs []net.IPNet, pubKey wgtypes.Key,
	psk wgtypes.Key, name string, mtu MTU, dns net.IP, keepalive int) (*ClientConfig, error) {
	user := s.getUser(userId)
	if s.MaxClientsPerUser > 0 {
		count := user.Count()
		if count >= s.MaxClientsPerUser {
			log.Error(fmt.Errorf("user %q have too many configs %d", user, s.MaxClientsPerUser))
			return nil, ErrTooManyClients
		}
	}
	if name == "" {
		log.Debugf("No clientName:using default: \"Unnamed Client\"")
		name = "Unnamed Client"
	}
	ip, err := s.allocateIp()
	if err != nil {
		return nil, err
	}
	if err := mtu.Validate(); err != nil {
		mtu = s.DefaultPeerMTU
	}
	client := &ClientConfig{
		IP:           ip,
		AllowedIPs:   allowedIPs,
		PublicKey:    pubKey,
		PresharedKey: psk,
		Name:         name,
		MTU:          mtu,
		DNS:          dns,
		KeepAlive:    keepalive,
		Created:      time.Now(),
		Modified:     time.Now(),
	}
	user.Add(client)
	return client, nil
}

func (s *ServerConfig) allocateIp() (net.IP, error) {
	allocated := make(map[string]bool)
	allocated[s.LinkConfig.IP.String()] = true
	clients := s.GetAllClients()
	for _, dev := range clients {
		allocated[dev.IP.String()] = true

	}
	serverIp := s.LinkConfig.IP
	serverNet := s.LinkConfig.IPNet
	for ip := serverIp.Mask(serverNet.Mask); serverNet.Contains(ip); {
		for i := len(ip) - 1; i >= 0; i-- {
			ip[i]++
			if ip[i] > 0 {
				break
			}
		}
		if !allocated[ip.String()] {
			log.Debug("Allocated IpNet: ", ip)
			return ip, nil
		}
	}
	return nil, ErrRangeExhausted
}

func (s *ServerConfig) RemoveClient(userId string, key wgtypes.Key) error {
	user := s.getUser(userId)
	return user.Remove(key)
}

func (s *ServerConfig) CountClients(userId string) int {
	user := s.getUser(userId)
	return user.Count()
}

// NewServerConfig creates a new server configuration
func NewServerConfig(
	endpoint net.UDPAddr,
	allowedIPs []net.IPNet,
	maxClientsPerUser int,
	defaultPeerMTU MTU,
	linkConfig *LinkConfig,
) (*ServerConfig, error) {
	privateKey, err := wgtypes.GenerateKey()
	if err != nil {
		return nil, errors.Join(ErrCannotGenerateKey, err)
	}
	return &ServerConfig{
		PrivateKey:        privateKey,
		Endpoint:          endpoint,
		MaxClientsPerUser: maxClientsPerUser,
		DefaultPeerMTU:    defaultPeerMTU,
		AllowedIPs:        allowedIPs,
		LinkConfig: &LinkConfig{
			Name:    linkConfig.Name,
			IP:      linkConfig.IP,
			IPNet:   linkConfig.IPNet,
			MTU:     linkConfig.MTU,
			NATLink: linkConfig.NATLink,
		},
		Users: make(map[string]UserConfig),
	}, nil
}

// MergeWith merges the current ServerConfig with another ServerConfig
func (s *ServerConfig) MergeWith(config ServerConfig) {
	if config.PrivateKey != (wgtypes.Key{}) {
		s.PrivateKey = config.PrivateKey
	}
	if config.LinkConfig != nil {
		if s.LinkConfig == nil {
			s.LinkConfig = &LinkConfig{}
		}
		s.LinkConfig.MergeWith(*config.LinkConfig)
	}
	if !config.Endpoint.IP.IsUnspecified() {
		s.Endpoint.IP = config.Endpoint.IP
	}
	if config.Endpoint.Port != 0 {
		s.Endpoint.Port = config.Endpoint.Port
	}
	if config.Endpoint.Zone != "" {
		s.Endpoint.Zone = config.Endpoint.Zone
	}
	if config.MaxClientsPerUser != 0 {
		s.MaxClientsPerUser = config.MaxClientsPerUser
	}
	if config.DefaultPeerMTU != 0 {
		s.DefaultPeerMTU = config.DefaultPeerMTU
	}

}

func (s *ServerConfig) GetAllClients() []*ClientConfig {
	clients := make([]*ClientConfig, 0)
	for _, user := range s.Users {
		for _, client := range user {
			clients = append(clients, client)
		}
	}
	return clients
}

// ██╗   ██╗███████╗███████╗██████╗
// ██║   ██║██╔════╝██╔════╝██╔══██╗
// ██║   ██║███████╗█████╗  ██████╔╝
// ██║   ██║╚════██║██╔══╝  ██╔══██╗
// ╚██████╔╝███████║███████╗██║  ██║
// ╚═════╝ ╚══════╝╚══════╝╚═╝  ╚═╝

// UserConfig represents a user and its clients
type UserConfig map[wgtypes.Key]*ClientConfig

// Get returns a client configuration by its public key
// @return *ClientConfig
// @return error
func (u UserConfig) Get(key wgtypes.Key) (*ClientConfig, error) {
	if _, ok := u[key]; !ok {
		return nil, ErrClientNotFound
	}
	return u[key], nil
}

// Add a client configuration to the current user
func (u UserConfig) Add(client *ClientConfig) {
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
func (u UserConfig) List() []*ClientConfig {
	clients := make([]*ClientConfig, 0, len(u))
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
	clients := make(map[string]*ClientConfig)
	for key, client := range u {
		clients[key.String()] = client
	}
	return json.Marshal(clients)
}

func (u *UserConfig) UnmarshalJSON(data []byte) error {
	clients := make(map[string]*ClientConfig)
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

// ██╗     ██╗███╗   ██╗██╗  ██╗
// ██║     ██║████╗  ██║██║ ██╔╝
// ██║     ██║██╔██╗ ██║█████╔╝
// ██║     ██║██║╚██╗██║██╔═██╗
// ███████╗██║██║ ╚████║██║  ██╗
// ╚══════╝╚═╝╚═╝  ╚═══╝╚═╝  ╚═╝

// LinkConfig holds the VPN link configuration
type LinkConfig struct {
	Name string
	IP   net.IP
	// IPNet is the IP address and subnet mask for the VPN network
	IPNet   net.IPNet
	MTU     MTU
	NATLink string
}

func (c *LinkConfig) MergeWith(config LinkConfig) {
	if config.Name != "" {
		c.Name = config.Name
	}
	if err := config.MTU.Validate(); err != nil {
		c.MTU = config.MTU
	}
	if config.NATLink != "" {
		c.NATLink = config.NATLink
	}
	if config.IP != nil {
		c.IP = config.IP
	}
	if config.IPNet.IP != nil && config.IPNet.Mask != nil {
		c.IPNet = config.IPNet
	}
}

func NewLinkConfig(name string, ip net.IP, ipNet net.IPNet, mtu MTU, natLink string) *LinkConfig {
	return &LinkConfig{
		Name:    name,
		IP:      ip,
		IPNet:   ipNet,
		MTU:     mtu,
		NATLink: natLink,
	}
}

// ███╗   ███╗████████╗██╗   ██╗
// ████╗ ████║╚══██╔══╝██║   ██║
// ██╔████╔██║   ██║   ██║   ██║
// ██║╚██╔╝██║   ██║   ██║   ██║
// ██║ ╚═╝ ██║   ██║   ╚██████╔╝
// ╚═╝     ╚═╝   ╚═╝    ╚═════╝

// MTU represents the Maximum Transmission Unit
type MTU int

func (mtu MTU) Validate() error {
	if mtu < 1280 || mtu > 1500 {
		return ErrInvalidMTU
	}
	return nil
}

// ClientConfigRepository is an interface for the client configuration repository
type ClientConfig struct {
	IP           net.IP
	AllowedIPs   []net.IPNet
	PublicKey    wgtypes.Key
	PresharedKey wgtypes.Key
	Name         string
	Notes        string
	MTU          MTU
	Created      time.Time
	Modified     time.Time
	DNS          net.IP
	KeepAlive    int
}

// ██████╗██╗     ██╗███████╗███╗   ██╗████████╗
// ██╔════╝██║     ██║██╔════╝████╗  ██║╚══██╔══╝
// ██║     ██║     ██║█████╗  ██╔██╗ ██║   ██║
// ██║     ██║     ██║██╔══╝  ██║╚██╗██║   ██║
// ╚██████╗███████╗██║███████╗██║ ╚████║   ██║
// ╚═════╝╚══════╝╚═╝╚══════╝╚═╝  ╚═══╝   ╚═╝

func (c *ClientConfig) Validate() error {
	if err := c.MTU.Validate(); err != nil {
		return err
	}
	if c.PublicKey == (wgtypes.Key{}) {
		return ErrInvalidPublicKey
	}
	if c.IP == nil || c.IP.IsUnspecified() {
		return ErrInvalidIP
	}
	if c.KeepAlive < 0 {
		return ErrInvalidKeepalive
	}

	return nil
}

func (c *ClientConfig) Update(allowedIPs []net.IPNet, psk wgtypes.Key, name, notes string, mtu MTU, dns net.IP, keepalive int) error {
	c.AllowedIPs = allowedIPs
	c.PresharedKey = psk
	c.Name = name
	c.Notes = notes
	c.MTU = mtu
	c.DNS = dns
	c.KeepAlive = keepalive
	c.Modified = time.Now()
	if err := c.Validate(); err != nil {
		return err
	}
	return nil
}

func (c *ClientConfig) MergeWith(config ClientConfig) {
	if config.IP != nil {
		c.IP = config.IP
	}
	if config.AllowedIPs != nil {
		c.AllowedIPs = config.AllowedIPs
	}
	if config.PublicKey != (wgtypes.Key{}) {
		c.PublicKey = config.PublicKey
	}
	if config.PresharedKey != (wgtypes.Key{}) {
		c.PresharedKey = config.PresharedKey
	}
	if config.Name != "" {
		c.Name = config.Name
	}
	if config.Notes != "" {
		c.Notes = config.Notes
	}
	if config.MTU != 0 {
		c.MTU = config.MTU
	}
	if config.Created != (time.Time{}) {
		c.Created = config.Created
	}
	if config.Modified != (time.Time{}) {
		c.Modified = config.Modified
	}
	if config.DNS != nil {
		c.DNS = config.DNS
	}
	if config.KeepAlive != 0 {
		c.KeepAlive = config.KeepAlive
	}
}
