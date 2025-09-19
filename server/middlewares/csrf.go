// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package middlewares

import (
	"net/http"

	"github.com/justinas/nosurf"

	"github.com/joshuar/go-hass-agent/models"
)

// SaveCSRFToken will save a new CSRF token for this request.
func SaveCSRFToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		next.ServeHTTP(res, req.WithContext(models.CSRFTokenToCtx(req.Context(), nosurf.Token(req))))
	})
}
