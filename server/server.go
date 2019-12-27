package server

import (
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/tellusxdp/tellus-market-sdk-gateway/config"
	"github.com/tellusxdp/tellus-market-sdk-gateway/token"
)

type Server struct {
	PrivateKeyURL string
	Upstream      string
	ProductID     string
}

func New(cfg *config.Config) *Server {
	s := &Server{
		PrivateKeyURL: cfg.PrivateKeyURL,
		Upstream:      cfg.Upstream,
		ProductID:     cfg.ProductID,
	}
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	authenticationHeader := strings.SplitN(r.Header.Get("Authorization"), " ", 2)

	if authenticationHeader[0] != "Bearer" {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
	}

	jwtToken := authenticationHeader[1]
	claim, err := token.ValidateToken(jwtToken, s.PrivateKeyURL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
	}

	log.Info(claim)

	w.Write([]byte(r.URL.String()))
}
