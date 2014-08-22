package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/shipyard/shipyard/controller/manager"
	"github.com/sirupsen/logrus"
)

var (
	logger = logrus.New()
)

func defaultDeniedHostHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "unauthorized", http.StatusUnauthorized)
}

type AuthRequired struct {
	deniedHostHandler http.Handler
	manager           *manager.Manager
}

func NewAuthRequired(m *manager.Manager) *AuthRequired {
	return &AuthRequired{
		deniedHostHandler: http.HandlerFunc(defaultDeniedHostHandler),
		manager:           m,
	}
}

func (a *AuthRequired) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := a.handleRequest(w, r)
		if err != nil {
			logger.Warnf("unauthorized request for %s from %s", r.URL.Path, r.RemoteAddr)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func (a *AuthRequired) handleRequest(w http.ResponseWriter, r *http.Request) error {
	valid := false
	authHeader := r.Header.Get("TOKEN")
	parts := strings.Split(authHeader, ":")
	if len(parts) == 2 {
		// validate
		if err := a.manager.VerifyAuthToken(parts[0], parts[1]); err == nil {
			valid = true
		}
	}

	if !valid {
		a.deniedHostHandler.ServeHTTP(w, r)
		return fmt.Errorf("unauthorized %s", r.RemoteAddr)
	}

	return nil
}

func (a *AuthRequired) HandlerFuncWithNext(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	err := a.handleRequest(w, r)

	if err != nil {
		logger.Warnf("unauthorized request for %s from %s", r.URL.Path, r.RemoteAddr)
		return
	}

	if next != nil {
		next(w, r)
	}
}
