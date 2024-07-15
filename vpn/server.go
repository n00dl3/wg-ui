package vpn

import (
	"net"
	"os"
	"sync"

	"github.com/google/nftables"
	"github.com/google/nftables/expr"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// Server is the running server
type Server struct {
	configStorage ConfigPersister
	mutex         sync.RWMutex
	Config        *ServerConfig
}

type wgLink struct {
	attrs *netlink.LinkAttrs
}

func (w *wgLink) Attrs() *netlink.LinkAttrs {
	return w.attrs
}

func (w *wgLink) Type() string {
	return "wireguard"
}

func ifname(n string) []byte {
	b := make([]byte, 16)
	copy(b, []byte(n+"\x00"))
	return b
}

// NewServer returns an instance of Server which contains both the webserver and the reference to Wireguard
func NewServer(cfg *ServerConfig, configStorage ConfigPersister) *Server {
	s := Server{
		configStorage: configStorage,
		Config:        cfg,
		mutex:         sync.RWMutex{},
	}
	return &s
}

func (s *Server) enableIPForward() error {
	p := "/proc/sys/net/ipv4/ip_forward"

	content, err := os.ReadFile(p)
	if err != nil {
		return err
	}

	if string(content) == "0\n" {
		log.Info("Enabling sys.net.ipv4.ip_forward")
		return os.WriteFile(p, []byte("1"), 0600)
	}

	return nil
}

// initInterface initializes the wireguard interface
func (s *Server) initInterface() error {
	attrs := netlink.NewLinkAttrs()
	attrs.Name = s.Config.LinkConfig.Name

	link := wgLink{
		attrs: &attrs,
	}

	log.Debug("Adding wireguard device: ", s.Config.LinkConfig.Name)
	err := netlink.LinkAdd(&link)
	if os.IsExist(err) {
		log.Infof("WireGuard interface %s already exists. Reusing.", s.Config.LinkConfig.Name)
	} else if err != nil {
		return err
	}

	ipNet := s.Config.LinkConfig.IPNet
	ipNet.IP = s.Config.LinkConfig.IP
	log.Debug("Adding ip address to wireguard device: ", ipNet.String())
	addr, _ := netlink.ParseAddr(ipNet.String())
	err = netlink.AddrAdd(&link, addr)
	if os.IsExist(err) {
		log.Infof("WireGuard interface %s already has the requested address: ", ipNet.String())
	} else if err != nil {
		return err
	}

	log.Debug("Setting link MTU: ", s.Config.LinkConfig.MTU)
	err = netlink.LinkSetMTU(&link, int(s.Config.LinkConfig.MTU))
	if err != nil {
		log.Error("Error setting link MTU: ", s.Config.LinkConfig.Name)
		return err
	}

	log.Debug("Bringing up wireguard device: ", s.Config.LinkConfig.Name)
	err = netlink.LinkSetUp(&link)
	if err != nil {
		log.Error("Error bringing up device: ", s.Config.LinkConfig.Name)
		return err
	}

	if s.Config.LinkConfig.NATLink != "" {
		log.Debug("Adding NAT / IpNet masquerading using nftables")
		ns, err := netns.Get()
		if err != nil {
			return err
		}

		conn := nftables.Conn{NetNS: int(ns)}

		log.Debug("Flushing nftable rulesets")
		conn.FlushRuleset()

		log.Debug("Setting up nftable rules for ip masquerading")

		nat := conn.AddTable(&nftables.Table{
			Family: nftables.TableFamilyIPv4,
			Name:   "nat",
		})

		conn.AddChain(&nftables.Chain{
			Name:     "prerouting",
			Table:    nat,
			Type:     nftables.ChainTypeNAT,
			Hooknum:  nftables.ChainHookPrerouting,
			Priority: nftables.ChainPriorityFilter,
		})

		post := conn.AddChain(&nftables.Chain{
			Name:     "postrouting",
			Table:    nat,
			Type:     nftables.ChainTypeNAT,
			Hooknum:  nftables.ChainHookPostrouting,
			Priority: nftables.ChainPriorityNATSource,
		})

		conn.AddRule(&nftables.Rule{
			Table: nat,
			Chain: post,
			Exprs: []expr.Any{
				&expr.Meta{Key: expr.MetaKeyOIFNAME, Register: 1},
				&expr.Cmp{
					Op:       expr.CmpOpEq,
					Register: 1,
					Data:     ifname(s.Config.LinkConfig.NATLink),
				},
				&expr.Masq{},
			},
		})

		if err := conn.Flush(); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) reconfigure() {
	log.Debug("Reconfiguring")
	err := s.configStorage.Persist(s.Config)
	if err != nil {
		// TODO: handle error
		log.Fatal(err)
	}

	err = s.configureWireGuard()
	if err != nil {
		// TODO: handle error
		log.Fatal(err)
	}
}

func (s *Server) configureWireGuard() error {
	log.Debugf("Reconfiguring wireguard interface %s", s.Config.LinkConfig.Name)

	wg, err := wgctrl.New()
	if err != nil {
		return err
	}

	log.Debugf("Getting current Wireguard config")
	currentDev, err := wg.Device(s.Config.LinkConfig.Name)
	if err != nil {
		return err
	}
	currentPeers := currentDev.Peers
	diffPeers := make([]wgtypes.PeerConfig, 0)

	peers := make([]wgtypes.PeerConfig, 0)
	clients := s.Config.GetAllClients()
	for id, dev := range clients {
		allowedIPs := make([]net.IPNet, 1+len(dev.AllowedIPs))
		allowedIPs[0] = *netlink.NewIPNet(dev.IP)
		for i, cidr := range dev.AllowedIPs {
			allowedIPs[1+i] = cidr
		}
		peer := wgtypes.PeerConfig{
			PublicKey:         dev.PublicKey,
			ReplaceAllowedIPs: true,
			AllowedIPs:        allowedIPs,
			PresharedKey:      &dev.PresharedKey,
		}

		log.WithFields(log.Fields{"client": id, "key": dev.PublicKey, "allowedIPs": peer.AllowedIPs}).Debug("Adding wireguard peer")

		peers = append(peers, peer)
	}

	// Determine peers updated and to be removed from WireGuard
	for _, i := range currentPeers {
		found := false
		for _, j := range peers {
			if i.PublicKey == j.PublicKey {
				found = true
				j.UpdateOnly = true
				diffPeers = append(diffPeers, j)
				break
			}
		}
		if !found {
			peertoremove := wgtypes.PeerConfig{
				PublicKey: i.PublicKey,
				Remove:    true,
			}
			diffPeers = append(diffPeers, peertoremove)
		}
	}

	// Determine peers to be added to WireGuard
	for _, i := range peers {
		found := false
		for _, j := range currentPeers {
			if i.PublicKey == j.PublicKey {
				found = true
				break
			}
		}
		if !found {
			diffPeers = append(diffPeers, i)
		}
	}

	cfg := wgtypes.Config{
		PrivateKey:   &s.Config.PrivateKey,
		ListenPort:   &s.Config.Endpoint.Port,
		ReplacePeers: false,
		Peers:        diffPeers,
	}
	err = wg.ConfigureDevice(s.Config.LinkConfig.Name, cfg)
	if err != nil {
		return err
	}

	return nil
}

// Start configures wiregard and initiates the interfaces as well as starts the webserver to accept clients
func (s *Server) Start() error {
	err := s.enableIPForward()
	if err != nil {
		return err
	}

	err = s.initInterface()
	if err != nil {
		return err
	}

	err = s.configureWireGuard()
	if err != nil {
		return err
	}
	return nil
}

// GetClients returns a list of all clients for the current user
func (s *Server) GetClients(user string) ([]*ClientConfig, error) {
	s.mutex.RLock()

	defer s.mutex.RUnlock()
	return s.Config.getUser(user).List(), nil
}

// GetClient returns a specific client for the current user
func (s *Server) GetClient(user string, publicKey wgtypes.Key) (*ClientConfig, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	client, err := s.Config.GetClient(user, publicKey)
	if err != nil {
		return nil, err
	}
	return client, err
}

// EditClient edits the specific client passed by the current user
func (s *Server) EditClient(user string, pubKey wgtypes.Key, allowedIPs []net.IPNet,
	psk wgtypes.Key, name, notes string, mtu MTU, dns net.IP, keepalive int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	var client *ClientConfig
	var err error
	if client, err = s.Config.GetClient(user, pubKey); err != nil {
		return err
	}
	if err = client.Update(allowedIPs, psk, name, notes, mtu, dns, keepalive); err != nil {
		return err
	}
	s.reconfigure()
	return nil
}

// DeleteClient deletes the specified client for the current user
func (s *Server) DeleteClient(user string, publicKey wgtypes.Key) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if err := s.Config.RemoveClient(user, publicKey); err != nil {
		return err
	}
	s.reconfigure()
	log.WithField("user", user).Debug("Deleted client: ", publicKey)
	return nil
}

// CreateClient creates a new client for the current user
func (s *Server) CreateClient(user string, allowedIPs []net.IPNet, pubKey wgtypes.Key,
	psk wgtypes.Key, name string, mtu MTU, dns net.IP, keepalive int) (*ClientConfig, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	log.WithField("user", user).Debug("CreateClient")
	client, err := s.Config.AddClient(user, allowedIPs, pubKey, psk, name, mtu, dns, keepalive)
	if err != nil {
		return nil, err
	}
	s.reconfigure()
	return client, nil
}
