package server

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/tellusxdp/tellus-market-sdk-gateway/config"
	"github.com/tellusxdp/tellus-market-sdk-gateway/token"
)

const (
	HEADER_USER      = "X-Tellus-Market-User"
	HEADER_REQUESTID = "X-Tellus-Market-RequestID"
)

type Server struct {
	PrivateKeyURL string
	CounterURL    string
	APIKey        string
	Upstream      *url.URL
	ToolID        string
	Logger        *log.Entry
	CounterChan   chan<- CountRequest
}

func New(cfg *config.Config) (*Server, error) {
	u, err := url.Parse(cfg.Upstream)
	if err != nil {
		return nil, err
	}

	s := &Server{
		PrivateKeyURL: cfg.PrivateKeyURL,
		CounterURL:    cfg.CounterURL,
		APIKey:        cfg.APIKey,
		Upstream:      u,
		ToolID:        cfg.ToolID,
		Logger:        log.WithField("tool_id", cfg.ToolID),
	}
	s.CounterChan = s.StartCountRequestLoop()
	return s, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// Allow CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "false")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

	if r.Method == "OPTIONS" {
		// preflight
		// Access-Control-Request-Headersが含まれていた場合はpreflight成功
		if r.Header.Get("Access-Control-Request-Headers") != "" {
			w.WriteHeader(204)
			return
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
	claim, err := token.ValidateToken(jwtToken, s.PrivateKeyURL)
	if err != nil {
		s.Logger.Warn(err.Error())
		writeError(w, http.StatusUnauthorized, "Unauthorized (invalid)")
		return
	}

	if claim.ToolID != s.ToolID {
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
	}

	rp := &httputil.ReverseProxy{Director: director}
	lw := NewLoggingResponseWriter(w)
	rp.ServeHTTP(lw, r)

	if 200 <= lw.StatusCode && lw.StatusCode <= 299 {
		// 有効なレスポンス
		go func() {
			c := CountRequest{
				ToolID:    s.ToolID,
				UserID:    claim.Subject,
				Token:     jwtToken,
				RequestID: requestID,
			}
			s.CounterChan <- c
		}()
	}
}
