package api

import (
	"context"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/embarkstudios/wireguard-ui/vpn"
	"github.com/fujiwara/go-amzn-oidc/validator"
	"github.com/golang-jwt/jwt"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// go:embed ui/dist
var assetsFS embed.FS

type contextKey string

func parseHexKey(s string) (wgtypes.Key, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return wgtypes.Key{}, err
	}
	return wgtypes.NewKey(b)
}

const key = contextKey("user")

type DefaultClientOptions struct {
	AllowedIPs []net.IPNet
	DNS        net.IP
	KeepAlive  int
}

type Server struct {
	vpnServer            *vpn.Server
	assets               http.Handler
	authUserHeader       string
	listenAddr           string
	authBasicUser        string
	authBasicPass        []byte
	devUIServer          string
	hmacKey              []byte
	defaultClientOptions DefaultClientOptions
}

// Start configures wiregard and initiates the interfaces as well as starts the webserver to accept clients
func (s *Server) Start() error {
	router := httprouter.New()
	router.GET("/api/v1/whoami", s.WhoAmI)
	router.GET("/api/v1/users/:user/clients/:client", s.withAuth(s.GetClient))
	router.PUT("/api/v1/users/:user/clients/:client", s.withAuth(s.EditClient))
	router.DELETE("/api/v1/users/:user/clients/:client", s.withAuth(s.DeleteClient))
	router.GET("/api/v1/users/:user/clients", s.withAuth(s.GetClients))
	router.POST("/api/v1/users/:user/clients", s.withAuth(s.CreateClient))
	log.Debug("Serving static assets embedded in binary")
	router.GET("/about", s.Index)
	router.GET("/client/:client", s.Index)
	router.GET("/newclient", s.Index)
	router.NotFound = s.assets
	if s.devUIServer != "" {
		handler := cors.New(cors.Options{
			AllowedOrigins:   []string{s.devUIServer},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
			AllowCredentials: true,
			AllowedHeaders:   []string{"*"},
		}).Handler(router)
		return http.ListenAndServe(s.listenAddr, s.basicAuth(s.userFromHeader(handler)))
	}
	log.WithField("listenAddr", s.listenAddr).Info("Starting server")
	return http.ListenAndServe(s.listenAddr, s.basicAuth(s.userFromHeader(router)))
}

func (s *Server) basicAuth(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// If we specified a user, require auth
		if s.authBasicUser != "" {
			u, p, ok := r.BasicAuth()
			if !ok || u != s.authBasicUser || bcrypt.CompareHashAndPassword([]byte(s.authBasicPass), []byte(p)) != nil {
				w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}

		handler.ServeHTTP(w, r)

	})
}

func (s *Server) userFromHeader(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Header.Get(s.authUserHeader)
		if user == "" {
			log.Debug("Unauthenticated request")
			user = "anonymous"
		}

		if s.authUserHeader == "X-Goog-Authenticated-User-Email" {
			user = strings.TrimPrefix(user, "accounts.google.com:")
		}

		// AWS ALB-specific JWT header (https://docs.aws.amazon.com/elasticloadbalancing/latest/application/listener-authenticate-users.html)
		if s.authUserHeader == "x-amzn-oidc-data" {
			claims, err := validator.Validate(user)
			if err != nil {
				log.Debug("Unauthenticated request")
				user = "anonymous"
			} else {
				user = claims.Email()
			}
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
			"user": user,
			"exp":  time.Now().Add(time.Hour * 24).Unix(),
			"iss":  "wireguard-ui",
		})
		if tokenString, err := token.SignedString(s.hmacKey); err == nil {
			cookie := http.Cookie{
				Name:  "wguser",
				Value: tokenString,
				Path:  "/",
			}
			http.SetCookie(w, &cookie)

			ctx := context.WithValue(r.Context(), key, user)
			handler.ServeHTTP(w, r.WithContext(ctx))
		} else {
			log.Error("Error signing token: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return

		}
	})
}

func (s *Server) withAuth(handler httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		log.Debug("Auth required")

		user := r.Context().Value(key)
		if user == nil {
			log.Error("Error getting username from request context")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if user != ps.ByName("user") {
			log.WithField("user", user).WithField("path", r.URL.Path).Warn("Unauthorized access")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		handler(w, r, ps)
	}
}

// WhoAmI returns the identity of the current user
func (s *Server) WhoAmI(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user := r.Context().Value(key).(string)
	log.Debug(user)
	err := json.NewEncoder(w).Encode(struct{ User string }{user})
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// GetClients returns a list of all clients for the current user
func (s *Server) GetClients(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user := r.Context().Value(key).(string)
	log.Debug(user)
	clients, err := s.vpnServer.GetClients(user)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	list := make([]Client, len(clients))
	for i, c := range clients {
		list[i] = newClientResponse(c, s.vpnServer.Config)
	}
	err = json.NewEncoder(w).Encode(list)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// Index returns the single-page app
func (s *Server) Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	log.Debug("Serving single-page app from URL: ", r.URL)
	r.URL.Path = "/"
	s.assets.ServeHTTP(w, r)
}

// GetClient returns a specific client for the current user
func (s *Server) GetClient(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user := r.Context().Value(key).(string)
	k, err := parseHexKey(ps.ByName("client"))
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	client, err := s.vpnServer.GetClient(user, k)
	if errors.Is(err, vpn.ErrClientNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	err = json.NewEncoder(w).Encode(newClientResponse(client, s.vpnServer.Config))
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// EditClient edits the specific client passed by the current user
func (s *Server) EditClient(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	payload, err := parseClientPayload(r)
	if err != nil {
		log.Warn("Error parsing request: ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = s.vpnServer.EditClient(
		r.Context().Value(key).(string),
		payload.PubKey,
		payload.AllowedIPs,
		payload.Psk,
		payload.Name,
		payload.Notes,
		vpn.MTU(payload.Mtu),
		s.defaultClientOptions.DNS,
		s.defaultClientOptions.KeepAlive,
	)
	if err != nil {
		if errors.Is(err, vpn.ErrClientNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else if errors.Is(err, vpn.ErrInvalidPublicKey) {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	client, err := s.vpnServer.GetClient(r.Context().Value(key).(string), payload.PubKey)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	c := newClientResponse(client, s.vpnServer.Config)
	if err := json.NewEncoder(w).Encode(c); err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// DeleteClient deletes the specified client for the current user
func (s *Server) DeleteClient(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user := r.Context().Value(key).(string)
	k, err := parseHexKey(ps.ByName("client"))
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err = s.vpnServer.DeleteClient(user, k); err != nil {
		log.Error(err)
		if errors.Is(err, vpn.ErrClientNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.WithField("user", user).Debug("Deleted client: ", k)
	w.WriteHeader(http.StatusOK)
}

// CreateClient creates a new client for the current user
func (s *Server) CreateClient(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var c *vpn.ClientConfig
	var err error
	user := r.Context().Value(key).(string)
	log.WithField("user", user).Debug("CreateClient")
	client, err := parseClientPayload(r)
	if err != nil {
		log.Warn("Error parsing request: ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if c, err = s.vpnServer.CreateClient(
		user,
		client.AllowedIPs,
		client.PrivateKey,
		client.PubKey,
		client.Psk,
		client.Name,
		vpn.MTU(client.Mtu),
		s.defaultClientOptions.DNS,
		s.defaultClientOptions.KeepAlive,
	); err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	client = newClientResponse(c, s.vpnServer.Config)
	if err := json.NewEncoder(w).Encode(client); err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func NewServer(
	vpnServer *vpn.Server,
	authUserHeader,
	listenAddr,
	authBasicUser string,
	authBasicPass,
	hmacKey []byte,
	devUIServer string,
) *Server {

	var fsys fs.FS = assetsFS
	if f, err := fs.Sub(fsys, "ui/dist"); err != nil {
		log.Error(fmt.Errorf("ui/dist does not exist in fs :%w", err))
	} else {
		fsys = f
	}
	assets := http.FileServer(http.FS(fsys))
	return &Server{
		vpnServer:      vpnServer,
		assets:         assets,
		authUserHeader: authUserHeader,
		listenAddr:     listenAddr,
		authBasicUser:  authBasicUser,
		authBasicPass:  authBasicPass,
		devUIServer:    devUIServer,
		hmacKey:        hmacKey,
	}
}
