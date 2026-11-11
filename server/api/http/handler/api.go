package handler

import (
	"github.com/cntechpower/anywhere/server/api/auth"
	"github.com/cntechpower/anywhere/server/server"
)

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
