package server

import (
	"crypto/tls"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/acme/autocert"

	"github.com/tellusxdp/tellus-market-sdk-gateway/config"
	"github.com/tellusxdp/tellus-market-sdk-gateway/token"
)

const (
	HEADER_USER      = "X-Tellus-Market-User"
	HEADER_REQUESTID = "X-Tellus-Market-RequestID"
)

type Server struct {
	Config      *config.Config
	Upstream    *url.URL
	Logger      *log.Entry
	CounterChan chan<- CountRequest
}

func New(cfg *config.Config) (*Server, error) {
	u, err := url.Parse(cfg.Upstream.URL)
	if err != nil {
		return nil, err
	}

	s := &Server{
		Config:   cfg,
		Upstream: u,
		Logger:   log.WithField("product_id", cfg.ProductID),
	}
	s.CounterChan = s.StartCountRequestLoop()
	return s, nil
}

func (s *Server) ListenAndServe() error {

	mux := http.NewServeMux()
	mux.HandleFunc("/.tellus/config", s.configHandler)
	mux.Handle("/", s)

	if s.Config.HTTP.TLS == nil {
		s.Logger.Infof("Listen on %s", s.Config.HTTP.ListenAddress)
		err := http.ListenAndServe(s.Config.HTTP.ListenAddress, mux)
		if err != nil {
			return err
		}
	}

	tlsConf := &tls.Config{
		ClientAuth:               tls.NoClientCert,
		NextProtos:               []string{"h2", "http/1.1"},
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
		},
	}

	if s.Config.HTTP.TLS.Autocert.Enabled {
		autocertCacheDir := "/tmp/autocert"

		if s.Config.HTTP.TLS.Autocert.CacheDir != "" {
			autocertCacheDir = s.Config.HTTP.TLS.Autocert.CacheDir
		}

		autocertManager := &autocert.Manager{
			Prompt: autocert.AcceptTOS,
			Cache:  autocert.DirCache(autocertCacheDir),
		}

		tlsConf.GetCertificate = autocertManager.GetCertificate
		go func() {
			s.Logger.Info("listening autocert on 0.0.0.0:80")
			err := http.ListenAndServe("0.0.0.0:80", autocertManager.HTTPHandler(nil))
			if err != nil {
				s.Logger.Warn(err)
			}
		}()
	}

	if s.Config.HTTP.TLS.Certificate != "" && s.Config.HTTP.TLS.Key != "" {
		var err error
		tlsConf.Certificates = make([]tls.Certificate, 1)
		tlsConf.Certificates[0], err = tls.LoadX509KeyPair(s.Config.HTTP.TLS.Certificate, s.Config.HTTP.TLS.Key)
		if err != nil {
			return err
		}
	}

	conn, err := net.Listen("tcp", s.Config.HTTP.ListenAddress)
	if err != nil {
		return err
	}
	ln := tls.NewListener(conn, tlsConf)

	server := &http.Server{Addr: ln.Addr().String(), Handler: mux}
	s.Logger.Infof("listening on %v with TLS", ln.Addr())
	return server.Serve(ln)
}

func (s *Server) configHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)

	type configResponse struct {
		ProductID        string   `yaml:"product_id"`
		AllowedAuthTypes []string `yaml:"allowed_auth_types"`
	}
	resp := &configResponse{
		ProductID:        s.Config.ProductID,
		AllowedAuthTypes: s.Config.AllowedAuthTypes,
	}

	body, _ := json.Marshal(resp)
	w.Write(body)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	originHeader := r.Header.Get("Origin")

	if originHeader != "" {
		// Allow Origin
		w.Header().Set("Access-Control-Allow-Origin", originHeader)

		if r.Method == "OPTIONS" {
			// preflight
			// Access-Control-Request-Headersが含まれていた場合はpreflight成功
			if r.Header.Get("Access-Control-Request-Headers") != "" {
				w.Header().Set("Access-Control-Max-Age", "600")
				w.Header().Set("Access-Control-Allow-Credentials", "false")
				w.Header().Set("Access-Control-Allow-Headers", r.Header.Get("Access-Control-Request-Headers"))
				w.Header().Set("Access-Control-Allow-Methods", r.Header.Get("Access-Control-Request-Method"))
				w.WriteHeader(204)
				return
			}
		}
	}

	authenticationHeader := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(authenticationHeader) != 2 {
		writeError(w, http.StatusUnauthorized, "Unauthorized (missing)")
		return
	}

	if authenticationHeader[0] != "Bearer" {
		writeError(w, http.StatusUnauthorized, "Unauthorized (unknown type)")
		return
	}

	jwtToken := authenticationHeader[1]
	claim, err := token.ValidateToken(jwtToken, s.Config.PrivateKeyURL)
	if err != nil {
		s.Logger.Debugf("Invalid token: %s", err.Error())
		writeError(w, http.StatusUnauthorized, "Unauthorized (invalid signature)")
		return
	}

	if claim.ProductID != s.Config.ProductID {
		s.Logger.Debugf("Invalid product id %s", claim.ProductID)
		writeError(w, http.StatusUnauthorized, "Unauthorized (invalid product)")
		return
	}

	isAllowedAuthType := false
	for _, t := range s.Config.AllowedAuthTypes {
		if claim.AuthType == t {
			isAllowedAuthType = true
		}
	}
	if !isAllowedAuthType {
		s.Logger.Debugf("Not allowed auth type %s", claim.AuthType)
		writeError(w, http.StatusUnauthorized, "Unauthorized (not allowed auth type)")
		return
	}

	requestID := r.Header.Get(HEADER_REQUESTID)
	if requestID == "" {
		u, err := uuid.NewRandom()
		if err != nil {
			s.Logger.Errorf("Cannot generate UUID: %s", err.Error())
		}
		requestID = u.String()
	}

	director := func(req *http.Request) {
		req.URL.Scheme = s.Upstream.Scheme
		req.URL.Host = s.Upstream.Host
		req.Host = s.Upstream.Host
		req.Header.Set(HEADER_USER, claim.Subject)
		req.Header.Set(HEADER_REQUESTID, requestID)
		req.Header.Set("X-Forwarded-Host", req.Host)
		for k, v := range s.Config.Upstream.Headers {
			req.Header.Set(k, v)
		}
	}

	reverseProxyErrorHandler := func(rw http.ResponseWriter, req *http.Request, err error) {
		logger := s.Logger.WithField("request_id", requestID)
		logger.Errorf("http: proxy error: %v", err)
		rw.WriteHeader(http.StatusBadGateway)
	}

	rp := &httputil.ReverseProxy{Director: director, ErrorHandler: reverseProxyErrorHandler}
	lw := NewLoggingResponseWriter(w)
	rp.ServeHTTP(lw, r)

	if 200 <= lw.StatusCode && lw.StatusCode <= 299 {
		// 有効なレスポンス
		go func() {
			c := CountRequest{
				ProductID: s.Config.ProductID,
				UserID:    claim.Subject,
				Token:     jwtToken,
				RequestID: requestID,
			}
			s.CounterChan <- c
		}()
	}
}
