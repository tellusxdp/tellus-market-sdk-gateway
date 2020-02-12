package server

import (
	"crypto/tls"
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
		Logger:   log.WithField("tool_id", cfg.ToolID),
	}
	s.CounterChan = s.StartCountRequestLoop()
	return s, nil
}

func (s *Server) ListenAndServe() error {

	if s.Config.HTTP.TLS == nil {
		s.Logger.Infof("Listen on %s", s.Config.HTTP.ListenAddress)
		err := http.ListenAndServe(s.Config.HTTP.ListenAddress, s)
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

	if s.Config.HTTP.TLS.Autocert {
		autocertManager := &autocert.Manager{
			Prompt: autocert.AcceptTOS,
		}
		tlsConf.GetCertificate = autocertManager.GetCertificate
		go func() {
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

	server := &http.Server{Addr: ":8000", Handler: s}
	s.Logger.Infof("listening on %v with TLS", ln.Addr())
	return server.Serve(ln)
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
		s.Logger.Warn(err.Error())
		writeError(w, http.StatusUnauthorized, "Unauthorized (invalid)")
		return
	}

	if claim.ToolID != s.Config.ToolID {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}

	requestID := r.Header.Get(HEADER_REQUESTID)
	if requestID == "" {
		u, err := uuid.NewRandom()
		if err != nil {
			log.Errorf("Cannot generate UUID: %s", err.Error())
		}
		requestID = u.String()
	}

	director := func(req *http.Request) {
		req.URL.Scheme = s.Upstream.Scheme
		req.URL.Host = s.Upstream.Host
		req.Header.Set(HEADER_USER, claim.Subject)
		req.Header.Set(HEADER_REQUESTID, requestID)
		for k, v := range s.Config.Upstream.Headers {
			req.Header.Set(k, v)
		}
	}

	rp := &httputil.ReverseProxy{Director: director}
	lw := NewLoggingResponseWriter(w)
	rp.ServeHTTP(lw, r)

	if 200 <= lw.StatusCode && lw.StatusCode <= 299 {
		// 有効なレスポンス
		go func() {
			c := CountRequest{
				ToolID:    s.Config.ToolID,
				UserID:    claim.Subject,
				Token:     jwtToken,
				RequestID: requestID,
			}
			s.CounterChan <- c
		}()
	}
}
