// This file is safe to edit. Once it exists it will not be overwritten.

package restapi

import (
	"crypto/tls"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"

	"github.com/cntechpower/anywhere/server/api/http/api/restapi/operations"
)

func configureFlags(api *operations.AnywhereServerAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

// configureAPI is only referenced by the generated restapi.Server bootstrap,
// which this application does not use. The console builds its own handler tree
// in server/api/http/handler/router.go so operations can be dependency
// injected. The function is kept so the generated package keeps compiling after
// `make api`; calling the generated bootstrap instead would return 501.
func configureAPI(api *operations.AnywhereServerAPI) http.Handler {
	api.ServeError = errors.ServeError
	api.JSONProducer = runtime.JSONProducer()
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "restapi bootstrap is disabled; handlers are wired in server/api/http/handler/router.go", http.StatusNotImplemented)
	})
}

// The TLS configuration before HTTPS server starts.
func configureTLS(_ *tls.Config) {}

// As soon as server is initialized but not run yet, this function will be called.
func configureServer(_ *http.Server, _, _ string) {}

// The middleware configuration is for the handler executors.
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything.
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return handler
}
