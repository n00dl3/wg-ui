package vpn

import (
	"errors"
	"fmt"
	"net"
	"time"

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
	ErrInvalidPrivateKey = fmt.Errorf("%e: invalid private key", ErrValidationError)
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

type CipheredKey [60]byte

// ClientConfigRepository is an interface for the client configuration repository
type ClientConfig struct {
	IP           net.IP
	AllowedIPs   []net.IPNet
	PublicKey    wgtypes.Key
	PresharedKey wgtypes.Key
	PrivateKey   CipheredKey
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
	if c.PrivateKey == (CipheredKey{}) {
		return ErrInvalidPrivateKey
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
	if config.PrivateKey != (CipheredKey{}) {
		c.PrivateKey = config.PrivateKey
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
