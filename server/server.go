package server

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/tellusxdp/tellus-market-sdk-gateway/config"
	"github.com/tellusxdp/tellus-market-sdk-gateway/token"
)

type Server struct {
	PrivateKeyURL string
	Upstream      *url.URL
	ToolID        string
}

func New(cfg *config.Config) (*Server, error) {
	u, err := url.Parse(cfg.Upstream)
	if err != nil {
		return nil, err
	}

	s := &Server{
		PrivateKeyURL: cfg.PrivateKeyURL,
		Upstream:      u,
		ToolID:        cfg.ToolID,
	}
	return s, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	authenticationHeader := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(authenticationHeader) != 2 {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if authenticationHeader[0] != "Bearer" {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	jwtToken := authenticationHeader[1]
	claim, err := token.ValidateToken(jwtToken, s.PrivateKeyURL)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if claim.ToolID != s.ToolID {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}

	director := func(req *http.Request) {
		req.URL.Scheme = s.Upstream.Scheme
		req.URL.Host = s.Upstream.Host
		req.Header.Set("X-Tellus-Market-User", claim.Subject)
	}

	rp := &httputil.ReverseProxy{Director: director}
	rp.ServeHTTP(w, r)
}
