package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/cntechpower/anywhere/server/api/auth"
	"github.com/cntechpower/anywhere/server/server"
)

var errForbidden = errors.New("forbidden: insufficient permissions")

type authUserContextKey struct{}

func withAuthUser(ctx context.Context, userName string) context.Context {
	return context.WithValue(ctx, authUserContextKey{}, userName)
}

func authUserFromRequest(r *http.Request) (string, bool) {
	if r == nil {
		return "", false
	}
	u, ok := r.Context().Value(authUserContextKey{}).(string)
	return u, ok && u != ""
}

// API is the HTTP console application. All dependencies are injected so the
// handler layer has no package level singletons and can be exercised in tests.
type API struct {
	server *server.Server
	users  *auth.UserValidator
	jwt    *auth.JwtValidator
}

func New(s *server.Server, users *auth.UserValidator, jwt *auth.JwtValidator) *API {
	return &API{server: s, users: users, jwt: jwt}
}

func (a *API) Server() *server.Server {
	return a.server
}

func (a *API) JWT() *auth.JwtValidator {
	return a.jwt
}

// scopedUser returns the username to restrict queries to when the caller is an
// authenticated non-admin user.
func (a *API) scopedUser(r *http.Request) (string, bool) {
	u, ok := authUserFromRequest(r)
	if !ok {
		return "", false
	}
	if a.users != nil && a.users.IsAdmin(u) {
		return "", false
	}
	return u, true
}

func (a *API) authorizeUser(r *http.Request, targetUser string) error {
	scoped, restricted := a.scopedUser(r)
	if !restricted || scoped == targetUser {
		return nil
	}
	return errForbidden
}
