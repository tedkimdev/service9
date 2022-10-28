package usergrp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	userCore "github.com/tedkimdev/service9/business/core/user"
	"github.com/tedkimdev/service9/business/data/store/user"
	"github.com/tedkimdev/service9/business/sys/auth"
	"github.com/tedkimdev/service9/business/sys/database"
	"github.com/tedkimdev/service9/business/sys/validate"
	"github.com/tedkimdev/service9/foundation/web"
)

type Handlers struct {
	User userCore.Core
	Auth *auth.Auth
}

// Query returns a list of users with paging.
func (h Handlers) Query(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	page := web.Param(r, "page")
	pageNumber, err := strconv.Atoi(page)
	if err != nil {
		return validate.NewRequestError(fmt.Errorf("invalid page format [%s]", page), http.StatusBadRequest)
	}
	rows := web.Param(r, "rows")
	rowsPerPage, err := strconv.Atoi(rows)
	if err != nil {
		return validate.NewRequestError(fmt.Errorf("invalid rows format [%s]", rows), http.StatusBadRequest)
	}

	users, err := h.User.Query(ctx, pageNumber, rowsPerPage)
	if err != nil {
		return fmt.Errorf("unable to query for users: %w", err)
	}

	return web.Respond(ctx, w, users, http.StatusOK)
}

// QueryByID returns a user by its ID.
func (h Handlers) QueryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	claims, err := auth.GetClaims(ctx)
	if err != nil {
		return validate.NewRequestError(auth.ErrForbidden, http.StatusForbidden)
	}

	userID := web.Param(r, "id")
	usr, err := h.User.QueryByID(ctx, claims, userID)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrInvalidID):
			return validate.NewRequestError(err, http.StatusBadRequest)
		case errors.Is(err, database.ErrNotFound):
			return validate.NewRequestError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("ID[%s]: %w", userID, err)
		}
	}

	return web.Respond(ctx, w, usr, http.StatusOK)
}

// Create adds a new user to the system.
func (h Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web value missing from context")
	}

	var nu user.NewUser
	if err := web.Decode(r, &nu); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	usr, err := h.User.Create(ctx, nu, v.Now)
	if err != nil {
		return fmt.Errorf("user[%+v]: %w", &usr, err)
	}

	return web.Respond(ctx, w, usr, http.StatusCreated)
}

// Update updates a user in the system.
func (h Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web value missing from context")
	}

	claims, err := auth.GetClaims(ctx)
	if err != nil {
		return validate.NewRequestError(auth.ErrForbidden, http.StatusForbidden)
	}

	var upd user.UpdateUser
	if err := web.Decode(r, &upd); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	userID := web.Param(r, "id")
	if err := h.User.Update(ctx, claims, userID, upd, v.Now); err != nil {
		switch {
		case errors.Is(err, database.ErrInvalidID):
			return validate.NewRequestError(err, http.StatusBadRequest)
		case errors.Is(err, database.ErrNotFound):
			return validate.NewRequestError(err, http.StatusNotFound)
		case errors.Is(err, database.ErrForbidden):
			return validate.NewRequestError(err, http.StatusForbidden)
		default:
			return fmt.Errorf("ID[%s] User[%+v]: %w", userID, &upd, err)
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// Delete removes a user from the system.
func (h Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	claims, err := auth.GetClaims(ctx)
	if err != nil {
		return validate.NewRequestError(auth.ErrForbidden, http.StatusForbidden)
	}

	userID := web.Param(r, "id")
	if err := h.User.Delete(ctx, claims, userID); err != nil {
		switch {
		case errors.Is(err, database.ErrInvalidID):
			return validate.NewRequestError(err, http.StatusBadRequest)
		default:
			return fmt.Errorf("ID[%s]: %w", userID, err)
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// Token provides an API token for the authenticated user.
func (h Handlers) Token(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web value missing from context")
	}

	email, pass, ok := r.BasicAuth()
	if !ok {
		err := errors.New("must provide email and password in Basic auth")
		return validate.NewRequestError(err, http.StatusUnauthorized)
	}

	claims, err := h.User.Authenticate(ctx, v.Now, email, pass)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrNotFound):
			return validate.NewRequestError(err, http.StatusNotFound)
		case errors.Is(err, database.ErrAuthenticationFailure):
			return validate.NewRequestError(err, http.StatusUnauthorized)
		default:
			return fmt.Errorf("authenticating: %w", err)
		}
	}

	var tkn struct {
		Token string `json:"token"`
	}
	tkn.Token, err = h.Auth.GenerateToken(claims)
	if err != nil {
		return fmt.Errorf("generating token: %w", err)
	}

	return web.Respond(ctx, w, tkn, http.StatusOK)
}
