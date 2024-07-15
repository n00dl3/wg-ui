package api

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/embarkstudios/wireguard-ui/vpn"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type ServerConfig struct {
	Endpoint   string
	AllowedIps []net.IPNet
	PubKey     wgtypes.Key
}

func (s ServerConfig) MarshalJSON() ([]byte, error) {
	allowedIps := make([]string, len(s.AllowedIps))
	for i, ipNet := range s.AllowedIps {
		allowedIps[i] = ipNet.String()
	}
	data := map[string]interface{}{
		"endpoint":  s.Endpoint,
		"alowedIPs": allowedIps,
		"publicKey": hex.EncodeToString(s.PubKey[:]),
	}
	return json.Marshal(data)
}

type Client struct {
	IP         net.IP
	Server     ServerConfig
	Dns        net.IP
	AllowedIPs []net.IPNet
	PubKey     wgtypes.Key
	Psk        wgtypes.Key
	Name       string
	Mtu        int
	Notes      string
	KeepAlive  int
	Created    time.Time
	Updated    time.Time
}

func newClientResponse(c *vpn.ClientConfig, s *vpn.ServerConfig) Client {
	return Client{
		IP: c.IP,
		Server: ServerConfig{
			Endpoint:   s.Endpoint.String(),
			AllowedIps: s.AllowedIPs,
			PubKey:     s.PrivateKey.PublicKey(),
		},
		Dns:        c.DNS,
		AllowedIPs: c.AllowedIPs,
		PubKey:     c.PublicKey,
		Psk:        c.PresharedKey,
		Name:       c.Name,
		Mtu:        int(c.MTU),
		Notes:      c.Notes,
		KeepAlive:  c.KeepAlive,
		Created:    c.Created,
		Updated:    c.Modified,
	}
}

func parseClientPayload(req *http.Request) (Client, error) {
	var c Client
	if err := json.NewDecoder(req.Body).Decode(&c); err != nil {
		return Client{}, err
	}
	return c, nil

}

var (
	ErrInvalidPublicKey  = errors.New("invalid public key")
	ErrInvalidPSK        = errors.New("invalid preshared key")
	ErrInvalidAllowedIps = errors.New("invalid allowed IPs")
)

func (c *Client) UnmarshalJSON(data []byte) error {
	var (
		psk        wgtypes.Key
		mtu        int
		name       string
		notes      string
		dnsIP      net.IP
		allowedIps []net.IPNet
	)
	m := make(map[string]interface{})
	json.Unmarshal(data, &m)
	if _, ok := m["publicKey"].(string); !ok {
		return ErrInvalidPublicKey
	}
	publicKey, err := parseHexKey(m["publicKey"].(string))
	if err != nil {
		errors.Join(ErrInvalidPublicKey, err)
	}
	if _, ok := m["psk"].(string); ok {
		if psk, err = parseHexKey(m["psk"].(string)); err != nil {
			return ErrInvalidPSK
		}
	}
	if err != nil {
		return errors.Join(ErrInvalidPSK, err)
	}
	if _, ok := m["dns"]; ok {
		dnsIP = net.ParseIP(m["dns"].(string))
	}
	if _, ok := m["allowedIps"]; ok {
		allowedCIDR := m["allowedIps"].([]interface{})
		allowedIps := make([]net.IPNet, len(allowedCIDR))
		for i := range allowedCIDR {
			cidr, ok := allowedCIDR[i].(string)
			if !ok {
				return errors.Join(ErrInvalidAllowedIps, errors.New("allowedIps must be an array of CIDR strings"))
			}
			_, ipNet, err := net.ParseCIDR(cidr)
			if err != nil {
				return errors.Join(ErrInvalidAllowedIps, err)
			}
			allowedIps[i] = *ipNet
		}
	}
	if _, ok := m["mtu"].(float64); ok {
		mtu = int(m["mtu"].(float64))
	}
	if _, ok := m["name"]; ok {
		name = m["name"].(string)
	}
	if _, ok := m["notes"]; ok {
		notes = m["notes"].(string)
	}
	*c = Client{
		PubKey:     publicKey,
		Psk:        psk,
		Name:       name,
		Dns:        dnsIP,
		Mtu:        mtu,
		AllowedIPs: allowedIps,
		Notes:      notes,
	}
	return nil
}

func (c Client) MarshalJSON() ([]byte, error) {
	allowedIps := make([]string, len(c.AllowedIPs))
	for i, ip := range c.AllowedIPs {
		allowedIps[i] = ip.String()
	}
	data := map[string]interface{}{
		"ip":         c.IP.String(),
		"publicKey":  hex.EncodeToString(c.PubKey[:]),
		"name":       c.Name,
		"dns":        c.Dns.String(),
		"mtu":        c.Mtu,
		"allowedIPs": allowedIps,
		"server":     c.Server,
		"notes":      c.Notes,
		"keepalive":  c.KeepAlive,
		"created":    c.Created,
		"updated":    c.Updated,
	}
	if c.Dns != nil {
		data["dns"] = c.Dns.String()
	}
	if c.Psk != (wgtypes.Key{}) {
		data["psk"] = hex.EncodeToString(c.Psk[:])
	}
	return json.Marshal(data)
}
