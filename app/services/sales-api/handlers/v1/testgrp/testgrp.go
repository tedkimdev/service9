package testgrp

import (
	"context"
	"errors"
	"math/rand"
	"net/http"

	"github.com/tedkimdev/service9/business/sys/validate"
	"github.com/tedkimdev/service9/foundation/web"
	"go.uber.org/zap"
)

// Handlers manages the set of check endpoints.
type Handlers struct {
	Log *zap.SugaredLogger
}

// Test handler is for development.
func (h *Handlers) Test(ctx context.Context, rw http.ResponseWriter, r *http.Request) error {
	if n := rand.Intn(100); n%2 == 0 {
		// return errors.New("untrusted error")
		return validate.NewRequestError(errors.New("trusted error"), http.StatusBadRequest)
	}

	status := struct {
		Status string
	}{
		Status: "OK",
	}

	return web.Respond(ctx, rw, status, http.StatusOK)
}
