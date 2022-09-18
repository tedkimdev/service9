package testgrp

import (
	"context"
	"net/http"

	"github.com/tedkimdev/service9/foundation/web"
	"go.uber.org/zap"
)

// Handlers manages the set of check endpoints.
type Handlers struct {
	Log *zap.SugaredLogger
}

// Test handler is for development.
func (h *Handlers) Test(ctx context.Context, rw http.ResponseWriter, r *http.Request) error {
	status := struct {
		Status string
	}{
		Status: "OK",
	}

	return web.Respond(ctx, rw, status, http.StatusOK)
}
