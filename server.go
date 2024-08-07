package main

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"path"
	"regexp"
	"strconv"

	"github.com/embarkstudios/wireguard-ui/api"
	"github.com/embarkstudios/wireguard-ui/vpn"
	log "github.com/sirupsen/logrus"
)

const (
	wgDefaultMtu      = 1420
	defaultDataDir    = "/var/lib/wireguard-ui"
	defaultKeyStorage = defaultDataDir + "/jwt-secret.bin"
	defaultConfigFile = defaultDataDir + "/config.json"
)

var (
	dataDir               = startServerCMD.Flag("data-dir", "Directory used for storage").Default(defaultDataDir).String()
	listenAddr            = startServerCMD.Flag("listen-address", "Address to listen to").Default(":8080").String()
	natEnabled            = startServerCMD.Flag("nat", "Whether NAT is enabled or not").Default("true").Bool()
	natLink               = startServerCMD.Flag("nat-device", "Network interface to masquerade").Default("wlp2s0").String()
	serverCIDR            = startServerCMD.Flag("server-ip", "Server IP address").Default("172.31.255.1/24").String()
	clientIPRange         = startServerCMD.Flag("client-ip-range", "Client IpNet CIDR").Default("172.31.255.0/24").String()
	authUserHeader        = startServerCMD.Flag("auth-user-header", "Header containing username").Default("X-Forwarded-User").String()
	authBasicUser         = startServerCMD.Flag("auth-basic-user", "Basic auth static username").Default("").String()
	authBasicPass         = startServerCMD.Flag("auth-basic-pass", "Basic auth static password").Default("").String()
	maxNumberClientConfig = startServerCMD.Flag("max-number-client-config", "Max number of configs an client can use. 0 is unlimited").Default("0").Int()
	wgLinkName            = startServerCMD.Flag("wg-device-name", "WireGuard network device name").Default("wg0").String()
	wgEndpoint            = startServerCMD.Flag("wg-endpoint", "WireGuard endpoint address").Default("127.0.0.1:51820").String()
	wgAllowedIPs          = startServerCMD.Flag("wg-allowed-ips", "WireGuard client allowed ips").Default("0.0.0.0/0").Strings()
	wgDNS                 = startServerCMD.Flag("wg-dns", "WireGuard client DNS server (optional)").Default("").String()
	wgKeepAlive           = startServerCMD.Flag("wg-keepalive", "WireGuard Keepalive for peers, defined in seconds (optional)").Default("").String()
	wgServerMtu           = startServerCMD.Flag("wg-server-mtu", "WireGuard server MTU").Default("1420").Int()
	wgPeerMtu             = startServerCMD.Flag("wg-peer-mtu", "WireGuard default peer MTU").Default(strconv.Itoa(wgDefaultMtu)).Int()
	devUIServer           = startServerCMD.Flag("dev-ui-server", "Developer mode: If specified, proxy all static assets to this endpoint").String()
	filenameRe            = regexp.MustCompile("[^a-zA-Z0-9]+")
)

//  ██████╗ ██████╗ ███╗   ██╗███████╗██╗ ██████╗     ██████╗ ███████╗██████╗  ██████╗
// ██╔════╝██╔═══██╗████╗  ██║██╔════╝██║██╔════╝     ██╔══██╗██╔════╝██╔══██╗██╔═══██╗
// ██║     ██║   ██║██╔██╗ ██║█████╗  ██║██║  ███╗    ██████╔╝█████╗  ██████╔╝██║   ██║
// ██║     ██║   ██║██║╚██╗██║██╔══╝  ██║██║   ██║    ██╔══██╗██╔══╝  ██╔═══╝ ██║   ██║
// ╚██████╗╚██████╔╝██║ ╚████║██║     ██║╚██████╔╝    ██║  ██║███████╗██║     ╚██████╔╝
//  ╚═════╝ ╚═════╝ ╚═╝  ╚═══╝╚═╝     ╚═╝ ╚═════╝     ╚═╝  ╚═╝╚══════╝╚═╝      ╚═════╝

type configRepository struct {
	configPath string
}

func (c configRepository) Persist(config *vpn.ServerConfig) error {
	var err error
	var f *os.File
	if err := ensureDir(path.Dir(c.configPath)); err != nil {
		return err
	}
	if f, err = os.Create(c.configPath); err != nil {
		return err
	}
	defer f.Close()
	err = json.NewEncoder(f).Encode(config)
	return err
}

func (c configRepository) Get() (*vpn.ServerConfig, error) {
	var err error
	var f *os.File
	cfg, err := configFromArgs()
	if err != nil {
		return cfg, err
	}
	if err := ensureDir(path.Dir(c.configPath)); err != nil {
		return nil, err
	}
	if f, err = os.Open(c.configPath); err == nil {
		defer f.Close()
		err = json.NewDecoder(f).Decode(cfg)
		return cfg, err
	}
	if errors.Is(err, fs.ErrNotExist) {
		return cfg, nil
	}
	return cfg, err
}

var (
	ErrInvalidServerIp   = errors.New("Invalid server IP")
	ErrInvalidAllowedIps = errors.New("Invalid allowed IPs")
	ErrInvalidEndpoint   = errors.New("Invalid endpoint")
)

func configFromArgs() (*vpn.ServerConfig, error) {
	ip, ipNet, err := net.ParseCIDR(*serverCIDR)
	if err != nil {
		return nil, errors.Join(ErrInvalidServerIp, err)
	}
	endpoint, err := net.ResolveUDPAddr("udp", *wgEndpoint)
	if err != nil {
		return nil, errors.Join(ErrInvalidEndpoint, err)
	}
	allowedIps := make([]net.IPNet, len(*wgAllowedIPs))
	for i, cidr := range *wgAllowedIPs {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, errors.Join(ErrInvalidAllowedIps, err)
		}
		allowedIps[i] = *ipNet
	}
	return vpn.NewServerConfig(
		*endpoint,
		allowedIps,
		*maxNumberClientConfig,
		vpn.MTU(*wgPeerMtu),
		vpn.NewLinkConfig(
			*wgLinkName,
			ip,
			*ipNet,
			vpn.MTU(*wgServerMtu),
			*natLink,
		),
	)
}

//      ██╗██╗    ██╗████████╗    ██╗  ██╗███████╗██╗   ██╗
//      ██║██║    ██║╚══██╔══╝    ██║ ██╔╝██╔════╝╚██╗ ██╔╝
//      ██║██║ █╗ ██║   ██║       █████╔╝ █████╗   ╚████╔╝
// ██   ██║██║███╗██║   ██║       ██╔═██╗ ██╔══╝    ╚██╔╝
// ╚█████╔╝╚███╔███╔╝   ██║       ██║  ██╗███████╗   ██║
//  ╚════╝  ╚══╝╚══╝    ╚═╝       ╚═╝  ╚═╝╚══════╝   ╚═╝

// readKey reads a key from the provided file path
func readKey(filePath string) ([]byte, error) {
	if err := ensureDir(path.Dir(filePath)); err != nil {
		return nil, err
	}
	key, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read key from file: %w", err)
	}
	if len(key) > 64 {
		log.Warn("Key is longer than 64 bytes, a 32 bytes key will be hashed from it")
	}
	return key, nil
}

func generateKey() ([]byte, error) {
	key := make([]byte, 64)
	_, err := rand.Read(key)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}
	return key, nil
}

// saveKey saves a key to the provided file path
func saveKey(filePath string, key []byte) error {
	if err := ensureDir(path.Dir(filePath)); err != nil {
		return err
	}
	if err := os.WriteFile(filePath, key, 0600); err != nil {
		return fmt.Errorf("failed to save key to file: %w", err)
	}
	return nil
}

func loadKey(filepath string) ([]byte, error) {
	// Load the HMAC key
	k, err := readKey(filepath)
	if errors.Is(err, os.ErrNotExist) {
		log.Warn("No HMAC key found, generating a new one")
		k, err = generateKey()
		if err != nil {
			log.WithError(err).Error("Failed to generate HMAC key")
			return nil, err
		}
		err = saveKey(filepath, k)
		if err != nil {
			log.WithError(err).Error("Failed to save HMAC key")
			return nil, err
		}

	}
	return k, err
}

func ensureDir(dir string) error {
	if stat, err := os.Stat(dir); errors.Is(err, fs.ErrNotExist) {
		return os.MkdirAll(dir, 0700)
	} else {
		if !stat.IsDir() {
			return fmt.Errorf("path %s is not a directory", dir)
		}
	}
	return nil
}

// ██████╗ ██╗   ██╗███╗   ██╗    ███████╗███████╗██████╗ ██╗   ██╗███████╗██████╗
// ██╔══██╗██║   ██║████╗  ██║    ██╔════╝██╔════╝██╔══██╗██║   ██║██╔════╝██╔══██╗
// ██████╔╝██║   ██║██╔██╗ ██║    ███████╗█████╗  ██████╔╝██║   ██║█████╗  ██████╔╝
// ██╔══██╗██║   ██║██║╚██╗██║    ╚════██║██╔══╝  ██╔══██╗╚██╗ ██╔╝██╔══╝  ██╔══██╗
// ██║  ██║╚██████╔╝██║ ╚████║    ███████║███████╗██║  ██║ ╚████╔╝ ███████╗██║  ██║
// ╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═══╝    ╚══════╝╚══════╝╚═╝  ╚═╝  ╚═══╝  ╚══════╝╚═╝  ╚═╝

// runServer starts the vpn server
func runServer() {
	cfgStorage := configRepository{
		configPath: *dataDir + "/config.json",
	}
	cfg, err := cfgStorage.Get()
	if err != nil {
		log.WithError(err).Error("Failed to load server config")
		os.Exit(1)
	}
	if err != nil {
		log.WithError(err).Error("Failed to parse client IP range")
		os.Exit(1)
	}
	c, _ := configFromArgs()
	cfg.MergeWith(*c)
	vpnServer := vpn.NewServer(cfg, cfgStorage, newPeerRepository())
	if err = vpnServer.Start(); err != nil {
		log.WithError(err).Error("Failed to start vpn server")
		os.Exit(1)
	}
	key, err := loadKey(*dataDir + "/jwt-secret.bin")
	if err != nil {
		log.WithError(err).Error("Failed to load JWT secret")
		os.Exit(1)
	}
	server := api.NewServer(vpnServer, *authUserHeader, *listenAddr, *authBasicUser, []byte(*authBasicPass), key, *devUIServer)
	if err := server.Start(); err != nil {
		log.WithError(err).Error("Failed to start server")
		os.Exit(1)
	}
}
