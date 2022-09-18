// Package web contains a small web framework extension.
package web

import (
	"context"
	"net/http"
	"os"

	"github.com/dimfeld/httptreemux/v5"
)

// A Handler is a type that handles a http request within our own little mini
// framework.
type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

// App is the entrypoint into our application and what configures our context
// object for each of our http handlers. Feel free to add any configuration
// data/logic on this App struct.
type App struct {
	*httptreemux.ContextMux
	shutdown chan os.Signal
}

// NewApp creates an App value that handle a set of routes for the application.
func NewApp(shutdown chan os.Signal) *App {
	mux := httptreemux.NewContextMux()

	return &App{
		ContextMux: mux,
		shutdown:   shutdown,
	}
}

// Handle sets a handler function for a given HTTP method and path pair
// to the application server mux.
func (a *App) Handle(method string, group string, path string, handler Handler) {

	// PRE CODE PROCESSING

	h := func(rw http.ResponseWriter, r *http.Request) {

		if err := handler(r.Context(), rw, r); err != nil {

			// ERROR HANDLING
			return
		}
	}

	// POST CODE PROCESSING

	finalPath := path
	if group != "" {
		finalPath = "/" + group + path
	}

	a.ContextMux.Handle(method, finalPath, h)
}
